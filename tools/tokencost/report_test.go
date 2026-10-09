package main

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// fakeGitHub keeps sub-issues and comments in memory.
type fakeGitHub struct {
	subs     map[int][]int
	err      error
	comments map[int][]comment
	nextID   int64
	created  int
	updated  int
}

func (f *fakeGitHub) SubIssues(_ context.Context, issue int) ([]int, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.subs[issue], nil
}

func (f *fakeGitHub) Comments(_ context.Context, issue int) ([]comment, error) {
	return f.comments[issue], nil
}

func (f *fakeGitHub) CreateComment(_ context.Context, issue int, body string) error {
	if f.comments == nil {
		f.comments = map[int][]comment{}
	}
	f.nextID++
	f.comments[issue] = append(f.comments[issue], comment{ID: f.nextID, Body: body})
	f.created++
	return nil
}

func (f *fakeGitHub) UpdateComment(_ context.Context, id int64, body string) error {
	for issue, cs := range f.comments {
		for i := range cs {
			if cs[i].ID == id {
				f.comments[issue][i].Body = body
				f.updated++
				return nil
			}
		}
	}
	return errors.New("comment not found")
}

func at(t testing.TB, s string) time.Time {
	t.Helper()
	tm, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatal(err)
	}
	return tm
}

// reportRows has issue 11 in two sessions, with two models and overhead,
// sub-issues 20 and 21, sub-sub-issue 30 and an unrelated issue 40.
func reportRows(t testing.TB) []row {
	return []row{
		{SessionID: "s1", Issue: 11, Model: "claude-opus-5-5", Start: at(t, "2026-10-09T08:00:00Z"), End: at(t, "2026-10-09T09:00:00Z"),
			Requests: 3, Usage: usage{Input: 10, Output: 200, CacheWrite5m: 30, CacheWrite1h: 40, CacheRead: 500}, Cost: 1000},
		{SessionID: "s1", Issue: 11, Model: overheadModel, Start: at(t, "2026-10-09T08:00:00Z"), End: at(t, "2026-10-09T09:00:00Z"), Cost: 100},
		{SessionID: "s2", Issue: 11, Model: "claude-haiku-5-5", Start: at(t, "2026-10-10T10:00:00Z"), End: at(t, "2026-10-10T10:30:00Z"),
			Requests: 1, Usage: usage{Input: 5, Output: 7}, Cost: 1},
		{SessionID: "s3", Issue: 20, Model: "claude-opus-5-5", Start: at(t, "2026-10-11T10:00:00Z"), End: at(t, "2026-10-11T11:00:00Z"), Requests: 1, Cost: 200},
		{SessionID: "s3", Issue: 21, Model: "claude-opus-5-5", Start: at(t, "2026-10-11T11:00:00Z"), End: at(t, "2026-10-11T12:00:00Z"), Requests: 1, Cost: 300},
		{SessionID: "s4", Issue: 30, Model: "claude-opus-5-5", Start: at(t, "2026-10-12T10:00:00Z"), End: at(t, "2026-10-12T11:00:00Z"), Requests: 1, Cost: 400},
		{SessionID: "s4", Issue: 40, Model: "claude-opus-5-5", Start: at(t, "2026-10-12T11:00:00Z"), End: at(t, "2026-10-12T12:00:00Z"), Requests: 1, Cost: 9999},
	}
}

func TestReportWithoutSubIssues(t *testing.T) {
	got, err := report(context.Background(), reportRows(t), 11, &fakeGitHub{})
	want := `<!-- tokencost -->
## Custo de IA da #11

| Modelo | Requisições | Entrada | Saída | Escrita no cache (5 min) | Escrita no cache (1 h) | Leitura do cache | Custo (US$) |
|---|---:|---:|---:|---:|---:|---:|---:|
| claude-haiku-5-5 | 1 | 5 | 7 | 0 | 0 | 0 | 0.0001 |
| claude-opus-5-5 | 3 | 10 | 200 | 30 | 40 | 500 | 0.1000 |
| overhead (chamadas fora do transcript) | | | | | | | 0.0100 |
| **Total** | 4 | 15 | 207 | 30 | 40 | 500 | **0.1101** |

Sessões: 2 · Período: 2026-10-09 08:00 a 2026-10-10 10:30 (UTC)
`
	if got != want {
		t.Errorf("report:\n%s\nwant:\n%s", got, want)
	}
	if err != nil {
		t.Errorf("err = %v, want nil", err)
	}
}

func TestReportWithSubIssues(t *testing.T) {
	gh := &fakeGitHub{subs: map[int][]int{11: {20, 21}, 21: {30}}}
	got, err := report(context.Background(), reportRows(t), 11, gh)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"| #11 | 0.1101 |\n| ↳ #20 | 0.0200 |\n| ↳ #21 | 0.0300 |\n| ↳ ↳ #30 | 0.0400 |\n| **Total da árvore** | **0.2001** |",
		"### Árvore por modelo (#11 e sub-issues)",
		"| claude-opus-5-5 | 6 | 10 | 200 | 30 | 40 | 500 | 0.1900 |",
		"| **Total** | 7 | 15 | 207 | 30 | 40 | 500 | **0.2001** |",
		"Sessões: 4 · Período: 2026-10-09 08:00 a 2026-10-12 11:00 (UTC)",
		"### Só a #11",
		"| **Total** | 4 | 15 | 207 | 30 | 40 | 500 | **0.1101** |",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("report misses %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "#40") {
		t.Errorf("report includes unrelated issue 40:\n%s", got)
	}
}

func TestReportWhenGitHubFails(t *testing.T) {
	gh := &fakeGitHub{err: errors.New("gh: not found")}
	got, err := report(context.Background(), reportRows(t), 11, gh)
	if !strings.Contains(got, "| **Total** | 4 | 15 | 207 | 30 | 40 | 500 | **0.1101** |") || strings.Contains(got, "Por issue") {
		t.Errorf("report should show only the own cost:\n%s", got)
	}
	if err == nil || !strings.Contains(err.Error(), "gh: not found") {
		t.Errorf("err = %v, want one about gh", err)
	}
}

func TestReportWithoutCost(t *testing.T) {
	got, _ := report(context.Background(), nil, 99, &fakeGitHub{})
	if !strings.Contains(got, "Sem custo registrado.") {
		t.Errorf("report:\n%s\nwant a no cost note", got)
	}
}

func TestUpsertComment(t *testing.T) {
	gh := &fakeGitHub{comments: map[int][]comment{11: {{ID: 7, Body: "LGTM"}}}}
	ctx := context.Background()
	created, err := upsertComment(ctx, gh, 11, commentMarker+"\nfirst")
	if err != nil || !created {
		t.Fatalf("first upsert = %v, %v, want created", created, err)
	}
	created, err = upsertComment(ctx, gh, 11, commentMarker+"\nsecond")
	if err != nil || created {
		t.Fatalf("second upsert = %v, %v, want updated", created, err)
	}
	cs := gh.comments[11]
	if len(cs) != 2 || cs[0].Body != "LGTM" || cs[1].Body != commentMarker+"\nsecond" {
		t.Errorf("comments = %+v, want the other comment untouched and one cost comment updated", cs)
	}
	if gh.created != 1 || gh.updated != 1 {
		t.Errorf("created %d, updated %d, want 1 and 1", gh.created, gh.updated)
	}
}
