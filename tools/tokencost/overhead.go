package main

import (
	"sort"
	"time"
)

// overheadModel names the rows that hold the cost Claude Code computed for a
// session beyond what its transcripts record (background calls such as
// titles and summaries, and calls of the main model left out of the
// transcript).
const overheadModel = "overhead"

// applyOverhead replaces the overhead rows of every session that has a cost
// computed by Claude Code. The gap between that cost and the cost of the
// session's transcript rows is split among the session's issues in
// proportion to their transcript cost. Sessions in noOverhead lose their
// stored overhead rows; other sessions without a known cost keep them.
func applyOverhead(rows []row, sessionCost map[string]int64, noOverhead map[string]bool, res *collectResult) []row {
	type issueTotal struct {
		issue      int
		cost       int64
		requests   int64
		start, end time.Time
	}
	bySession := map[string]map[int]*issueTotal{}
	stored := map[rowKey]row{}
	var kept []row
	for _, r := range rows {
		_, known := sessionCost[r.SessionID]
		if r.Model == overheadModel {
			if known || noOverhead[r.SessionID] {
				stored[r.key()] = r
			} else {
				kept = append(kept, r)
			}
			continue
		}
		kept = append(kept, r)
		if !known {
			continue
		}
		issues := bySession[r.SessionID]
		if issues == nil {
			issues = map[int]*issueTotal{}
			bySession[r.SessionID] = issues
		}
		it := issues[r.Issue]
		if it == nil {
			it = &issueTotal{issue: r.Issue, start: r.Start, end: r.End}
			issues[r.Issue] = it
		}
		it.cost += r.Cost
		it.requests += r.Requests
		if r.Start.Before(it.start) {
			it.start = r.Start
		}
		if r.End.After(it.end) {
			it.end = r.End
		}
	}

	fresh := map[rowKey]row{}
	for session, issues := range bySession {
		totals := make([]*issueTotal, 0, len(issues))
		var transcriptCost int64
		for _, it := range issues {
			totals = append(totals, it)
			transcriptCost += it.cost
		}
		gap := sessionCost[session] - transcriptCost
		if gap <= 0 {
			continue
		}
		sort.Slice(totals, func(i, j int) bool { return totals[i].issue < totals[j].issue })
		weights := make([]int64, len(totals))
		for i, it := range totals {
			weights[i] = it.cost
		}
		if transcriptCost == 0 {
			for i, it := range totals {
				weights[i] = it.requests
			}
		}
		for i, share := range split(gap, weights) {
			if share == 0 {
				continue
			}
			it := totals[i]
			r := row{SessionID: session, Issue: it.issue, Model: overheadModel, Start: it.start, End: it.end, Cost: share}
			fresh[r.key()] = r
		}
	}

	keys := make([]rowKey, 0, len(fresh))
	for k := range fresh {
		keys = append(keys, k)
	}
	sortKeys(keys)
	for _, k := range keys {
		r := fresh[k]
		old, ok := stored[k]
		switch {
		case !ok:
			res.record("new", r)
		case old.Cost != r.Cost || !old.Start.Equal(r.Start) || !old.End.Equal(r.End):
			res.record("updated", r)
		}
		kept = append(kept, r)
	}
	keys = keys[:0]
	for k := range stored {
		if _, ok := fresh[k]; !ok {
			keys = append(keys, k)
		}
	}
	sortKeys(keys)
	for _, k := range keys {
		res.record("removed", stored[k])
	}
	return kept
}

// split divides total in proportion to weights, by largest remainder; ties go
// to the lowest index. With no weight, it is split equally.
func split(total int64, weights []int64) []int64 {
	shares := make([]int64, len(weights))
	if len(weights) == 0 {
		return shares
	}
	var sum int64
	for _, w := range weights {
		sum += w
	}
	if sum == 0 {
		weights = make([]int64, len(weights))
		for i := range weights {
			weights[i] = 1
		}
		sum = int64(len(weights))
	}
	rest := make([]int64, len(weights))
	left := total
	for i, w := range weights {
		shares[i] = total * w / sum
		rest[i] = total * w % sum
		left -= shares[i]
	}
	order := make([]int, len(weights))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(a, b int) bool { return rest[order[a]] > rest[order[b]] })
	for _, i := range order[:left] {
		shares[i]++
	}
	return shares
}
