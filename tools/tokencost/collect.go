package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type collectConfig struct {
	// Root is the absolute path of the repository; only responses whose cwd
	// is inside it are counted.
	Root        string
	LogPath     string
	PricesPath  string
	OutPath     string
	ProjectsDir string
	ImportOld   bool
}

type collectResult struct {
	New     int
	Updated int
	Removed int
	// Changes lists the new, updated and removed rows, in that order of
	// discovery.
	Changes  []change
	Warnings []string
}

type change struct {
	Kind string // "new", "updated" or "removed"
	Row  row
}

func (res *collectResult) record(kind string, r row) {
	switch kind {
	case "new":
		res.New++
	case "updated":
		res.Updated++
	case "removed":
		res.Removed++
	}
	res.Changes = append(res.Changes, change{kind, r})
}

func collect(cfg collectConfig) (collectResult, error) {
	prices, err := loadPrices(cfg.PricesPath)
	if err != nil {
		return collectResult{}, err
	}
	paths, err := transcriptPaths(cfg)
	if err != nil {
		return collectResult{}, err
	}
	ts := newTranscripts()
	for _, p := range paths {
		if err := ts.readFile(p); err != nil {
			return collectResult{}, err
		}
	}
	var res collectResult
	fresh, notes, mixed := aggregate(ts.responses, cfg.Root)
	existing, err := readRows(cfg.OutPath)
	if err != nil {
		return collectResult{}, err
	}
	rows := merge(existing, fresh, notes, prices, &res)
	// The cost Claude Code computed for a session that also worked outside
	// the repository covers that work too, so its gap is not overhead here.
	changed := map[string]bool{}
	for _, c := range res.Changes {
		changed[c.Row.SessionID] = true
	}
	sessionCost := map[string]int64{}
	var warnMixed []string
	for session, c := range ts.sessionCost {
		if !mixed[session] {
			sessionCost[session] = c
		} else if changed[session] {
			warnMixed = append(warnMixed, session)
		}
	}
	sort.Strings(warnMixed)
	for _, session := range warnMixed {
		res.Warnings = append(res.Warnings, fmt.Sprintf("session %s also worked outside the repository: no overhead", session))
	}
	rows = applyOverhead(rows, sessionCost, mixed, &res)
	if err := writeRows(cfg.OutPath, rows); err != nil {
		return collectResult{}, err
	}
	return res, nil
}

// transcriptPaths lists the registered transcripts, the transcripts of their
// subagents and, with ImportOld, every transcript under ProjectsDir.
func transcriptPaths(cfg collectConfig) ([]string, error) {
	registered, err := readLog(cfg.LogPath)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var paths []string
	addAll := func(ps []string) {
		for _, p := range ps {
			if !seen[p] {
				seen[p] = true
				paths = append(paths, p)
			}
		}
	}
	for _, p := range registered {
		addAll([]string{p})
		subs, err := subagentTranscripts(p)
		if err != nil {
			return nil, fmt.Errorf("list subagent transcripts: %w", err)
		}
		addAll(subs)
	}
	if cfg.ImportOld {
		err := filepath.WalkDir(cfg.ProjectsDir, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && strings.HasSuffix(p, ".jsonl") {
				addAll([]string{p})
			}
			return nil
		})
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("list transcripts: %w", err)
		}
	}
	sort.Strings(paths)
	return paths, nil
}

func readLog(path string) ([]string, error) {
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("open sessions log: %w", err)
	}
	defer f.Close()
	var paths []string
	s := bufio.NewScanner(f)
	for s.Scan() {
		if p := strings.TrimSpace(s.Text()); p != "" {
			paths = append(paths, p)
		}
	}
	if err := s.Err(); err != nil {
		return nil, fmt.Errorf("read sessions log: %w", err)
	}
	return paths, nil
}

func sortKeys(keys []rowKey) {
	sort.Slice(keys, func(i, j int) bool {
		a, b := keys[i], keys[j]
		if a.SessionID != b.SessionID {
			return a.SessionID < b.SessionID
		}
		if a.Issue != b.Issue {
			return a.Issue < b.Issue
		}
		return a.Model < b.Model
	})
}

func inside(root, dir string) bool {
	return dir == root || strings.HasPrefix(dir, root+string(filepath.Separator))
}

// notes counts what a row holds that is priced differently than it was
// billed.
type notes struct {
	legacy, fast, searches int64
}

func (n notes) add(o notes) notes {
	return notes{n.legacy + o.legacy, n.fast + o.fast, n.searches + o.searches}
}

// aggregate sums the responses of this repository into rows, without cost,
// and counts the notes of each row. It also returns the sessions that have
// responses both inside and outside the repository.
func aggregate(rs map[string]response, root string) (map[rowKey]row, map[rowKey]notes, map[string]bool) {
	rows := map[rowKey]row{}
	ns := map[rowKey]notes{}
	in, out := map[string]bool{}, map[string]bool{}
	for _, r := range rs {
		if !inside(root, r.Cwd) {
			out[r.SessionID] = true
			continue
		}
		in[r.SessionID] = true
		k := rowKey{r.SessionID, r.Issue, r.Model}
		n := ns[k]
		if r.LegacyCache {
			n.legacy++
		}
		if r.Fast {
			n.fast++
		}
		n.searches += r.WebSearches
		ns[k] = n
		rw, ok := rows[k]
		if !ok {
			rw = row{SessionID: r.SessionID, Issue: r.Issue, Model: r.Model, Start: r.Time, End: r.Time}
		}
		if r.Time.Before(rw.Start) {
			rw.Start = r.Time
		}
		if r.Time.After(rw.End) {
			rw.End = r.Time
		}
		rw.Requests++
		rw.Usage = rw.Usage.add(r.Usage)
		rows[k] = rw
	}
	mixed := map[string]bool{}
	for session := range in {
		if out[session] {
			mixed[session] = true
		}
	}
	return rows, ns, mixed
}

// merge upserts fresh rows into existing ones. A stored cost never changes:
// tokens added to a row since the last collection are priced at current
// prices and added to it. Rows absent from fresh are kept. Notes are reported
// only for new and updated rows, so a collection does not repeat the notes of
// previous ones.
func merge(existing []row, fresh map[rowKey]row, ns map[rowKey]notes, prices priceTable, res *collectResult) []row {
	byKey := make(map[rowKey]int, len(existing))
	rows := append([]row(nil), existing...)
	for i, r := range rows {
		byKey[r.key()] = i
	}
	keys := make([]rowKey, 0, len(fresh))
	for k := range fresh {
		keys = append(keys, k)
	}
	sortKeys(keys)
	unpriced := map[string]bool{}
	var changed notes
	for _, k := range keys {
		f := fresh[k]
		p, ok := prices.lookup(f.Model)
		if !ok && !unpriced[f.Model] {
			unpriced[f.Model] = true
			res.Warnings = append(res.Warnings, fmt.Sprintf("no price for model %s: cost 0", f.Model))
		}
		i, exists := byKey[k]
		switch {
		case !exists:
			f.Cost = cost(f.Usage, p)
			rows = append(rows, f)
			res.record("new", f)
			changed = changed.add(ns[k])
		case f.Usage == rows[i].Usage:
		case f.Usage.covers(rows[i].Usage):
			old := rows[i]
			f.Cost = old.Cost + cost(f.Usage.sub(old.Usage), p)
			if old.Start.Before(f.Start) {
				f.Start = old.Start
			}
			if old.End.After(f.End) {
				f.End = old.End
			}
			rows[i] = f
			res.record("updated", f)
			changed = changed.add(ns[k])
		default:
			res.Warnings = append(res.Warnings, fmt.Sprintf(
				"session %s issue %d model %s: fewer tokens than stored, keeping stored row", k.SessionID, k.Issue, k.Model))
		}
	}
	if changed.legacy > 0 {
		res.Warnings = append(res.Warnings, fmt.Sprintf("%d responses without cache TTL breakdown: cache writes priced as 5 minutes", changed.legacy))
	}
	if changed.fast > 0 {
		res.Warnings = append(res.Warnings, fmt.Sprintf("%d responses in fast mode: priced at standard rates", changed.fast))
	}
	if changed.searches > 0 {
		res.Warnings = append(res.Warnings, fmt.Sprintf("%d web searches: not priced", changed.searches))
	}
	return rows
}
