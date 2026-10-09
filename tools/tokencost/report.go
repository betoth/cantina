package main

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
)

// commentMarker identifies the cost comment of an issue, so later reports
// edit it instead of adding another.
const commentMarker = "<!-- tokencost -->"

const overheadLabel = "overhead (chamadas fora do transcript)"

// issueNode is an issue and its sub-issues.
type issueNode struct {
	Issue    int
	Children []*issueNode
}

// issueTree fetches the sub-issues of issue recursively. On error it returns
// the issue alone and the error, so the report still shows its own cost.
func issueTree(ctx context.Context, gh github, issue int) (*issueNode, error) {
	seen := map[int]bool{}
	var walk func(n int) (*issueNode, error)
	walk = func(n int) (*issueNode, error) {
		seen[n] = true
		node := &issueNode{Issue: n}
		subs, err := gh.SubIssues(ctx, n)
		if err != nil {
			return nil, err
		}
		for _, s := range subs {
			if seen[s] {
				continue
			}
			child, err := walk(s)
			if err != nil {
				return nil, err
			}
			node.Children = append(node.Children, child)
		}
		return node, nil
	}
	tree, err := walk(issue)
	if err != nil {
		return &issueNode{Issue: issue}, err
	}
	return tree, nil
}

// report renders the cost of issue and of its sub-issues as Markdown. When
// the sub-issues cannot be fetched, it renders the issue's own cost and
// returns the error with it.
func report(ctx context.Context, rows []row, issue int, gh github) (string, error) {
	tree, treeErr := issueTree(ctx, gh, issue)
	if treeErr != nil {
		treeErr = fmt.Errorf("sub-issues of #%d not fetched, showing its own cost only: %w", issue, treeErr)
	}
	byIssue := map[int][]row{}
	for _, r := range rows {
		byIssue[r.Issue] = append(byIssue[r.Issue], r)
	}

	var b strings.Builder
	b.WriteString(commentMarker + "\n")
	fmt.Fprintf(&b, "## Custo de IA da #%d\n\n", issue)
	if len(tree.Children) == 0 {
		writeModelTable(&b, byIssue[issue])
		return b.String(), treeErr
	}

	b.WriteString("### Por issue\n\n")
	b.WriteString("| Issue | Custo próprio (US$) |\n|---|---:|\n")
	var all []row
	var walk func(n *issueNode, depth int)
	walk = func(n *issueNode, depth int) {
		rs := byIssue[n.Issue]
		all = append(all, rs...)
		fmt.Fprintf(&b, "| %s#%d | %s |\n", strings.Repeat("↳ ", depth), n.Issue, formatCost(totalCost(rs)))
		for _, c := range n.Children {
			walk(c, depth+1)
		}
	}
	walk(tree, 0)
	fmt.Fprintf(&b, "| **Total da árvore** | **%s** |\n\n", formatCost(totalCost(all)))

	fmt.Fprintf(&b, "### Árvore por modelo (#%d e sub-issues)\n\n", issue)
	writeModelTable(&b, all)
	fmt.Fprintf(&b, "\n### Só a #%d\n\n", issue)
	writeModelTable(&b, byIssue[issue])
	return b.String(), treeErr
}

// writeModelTable writes the tokens and cost of rows by model, with the total,
// the number of sessions and the period.
func writeModelTable(b *strings.Builder, rows []row) {
	if len(rows) == 0 {
		b.WriteString("Sem custo registrado.\n")
		return
	}
	type modelTotal struct {
		requests int64
		usage    usage
		cost     int64
	}
	models := map[string]*modelTotal{}
	sessions := map[string]bool{}
	var start, end time.Time
	var total modelTotal
	for _, r := range rows {
		m := models[r.Model]
		if m == nil {
			m = &modelTotal{}
			models[r.Model] = m
		}
		m.requests += r.Requests
		m.usage = m.usage.add(r.Usage)
		m.cost += r.Cost
		total.requests += r.Requests
		total.usage = total.usage.add(r.Usage)
		total.cost += r.Cost
		sessions[r.SessionID] = true
		if start.IsZero() || r.Start.Before(start) {
			start = r.Start
		}
		if r.End.After(end) {
			end = r.End
		}
	}
	names := make([]string, 0, len(models))
	for name := range models {
		names = append(names, name)
	}
	// Overhead goes last, after the models the transcripts record.
	sort.Slice(names, func(i, j int) bool {
		if (names[i] == overheadModel) != (names[j] == overheadModel) {
			return names[j] == overheadModel
		}
		return names[i] < names[j]
	})

	b.WriteString("| Modelo | Requisições | Entrada | Saída | Escrita no cache (5 min) | Escrita no cache (1 h) | Leitura do cache | Custo (US$) |\n")
	b.WriteString("|---|---:|---:|---:|---:|---:|---:|---:|\n")
	for _, name := range names {
		m := models[name]
		if name == overheadModel {
			fmt.Fprintf(b, "| %s | | | | | | | %s |\n", overheadLabel, formatCost(m.cost))
			continue
		}
		u := m.usage
		fmt.Fprintf(b, "| %s | %d | %d | %d | %d | %d | %d | %s |\n",
			name, m.requests, u.Input, u.Output, u.CacheWrite5m, u.CacheWrite1h, u.CacheRead, formatCost(m.cost))
	}
	u := total.usage
	fmt.Fprintf(b, "| **Total** | %d | %d | %d | %d | %d | %d | **%s** |\n\n",
		total.requests, u.Input, u.Output, u.CacheWrite5m, u.CacheWrite1h, u.CacheRead, formatCost(total.cost))
	fmt.Fprintf(b, "Sessões: %d · Período: %s a %s (UTC)\n",
		len(sessions), start.UTC().Format("2006-01-02 15:04"), end.UTC().Format("2006-01-02 15:04"))
}

func totalCost(rows []row) int64 {
	var c int64
	for _, r := range rows {
		c += r.Cost
	}
	return c
}

// upsertComment edits the comment of issue that has the marker, or creates
// one. It reports whether the comment was created.
func upsertComment(ctx context.Context, gh github, issue int, body string) (bool, error) {
	cs, err := gh.Comments(ctx, issue)
	if err != nil {
		return false, err
	}
	for _, c := range cs {
		if strings.Contains(c.Body, commentMarker) {
			return false, gh.UpdateComment(ctx, c.ID, body)
		}
	}
	return true, gh.CreateComment(ctx, issue, body)
}
