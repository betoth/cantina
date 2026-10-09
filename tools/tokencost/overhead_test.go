package main

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"
)

func TestSplit(t *testing.T) {
	tests := []struct {
		name    string
		total   int64
		weights []int64
		want    []int64
	}{
		{"proportional", 40, []int64{60, 20}, []int64{30, 10}},
		{"remainder to largest fraction", 10, []int64{1, 1, 1}, []int64{4, 3, 3}},
		{"remainder by fraction, not order", 5, []int64{1, 3}, []int64{1, 4}},
		{"no weight splits equally", 3, []int64{0, 0}, []int64{2, 1}},
		{"single issue takes all", 7, []int64{5}, []int64{7}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := split(tt.total, tt.weights); !slices.Equal(got, tt.want) {
				t.Errorf("split(%d, %v) = %v, want %v", tt.total, tt.weights, got, tt.want)
			}
		})
	}
}

func costState(session string, usd float64) string {
	return fmt.Sprintf(`{"type":"cost-state","sessionId":%q,"totalCostUSD":%v}`+"\n", session, usd)
}

// overheadCosts returns the overhead cost of each issue.
func overheadCosts(rows []row) map[int]int64 {
	got := map[int]int64{}
	for _, r := range rows {
		if r.Model == overheadModel {
			got[r.Issue] = r.Cost
		}
	}
	return got
}

func TestCollectOverhead(t *testing.T) {
	// Issue 1: 300 × 20 = US$ 0.0060; issue 2: 100 × 20 = US$ 0.0020.
	transcript := lines(
		line{MsgID: "m1", Branch: "feat/1-a", Out: 300, Time: "2026-10-09T10:00:00.000Z"},
		line{MsgID: "m2", Branch: "feat/2-b", Out: 100, Time: "2026-10-09T11:00:00.000Z"},
	)
	tests := []struct {
		name      string
		costState string
		want      map[int]int64
	}{
		{"gap split by issue cost", costState("s1", 0.012), map[int]int64{1: 30, 2: 10}},
		{"no cost state", "", map[int]int64{}},
		{"claude code cost not above transcript cost", costState("s1", 0.008), map[int]int64{}},
		{"cost state of another session", costState("s9", 1), map[int]int64{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newEnv(t)
			e.register(e.transcript("s1.jsonl", transcript+tt.costState))
			e.collect()
			got := overheadCosts(e.rows())
			if fmt.Sprint(got) != fmt.Sprint(tt.want) {
				t.Errorf("overhead = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCollectOverheadUpdates(t *testing.T) {
	e := newEnv(t)
	s := e.transcript("s1.jsonl", lines(
		line{MsgID: "m1", Branch: "feat/1-a", Out: 300, Time: "2026-10-09T10:00:00.000Z"},
		line{MsgID: "m2", Branch: "feat/2-b", Out: 100, Time: "2026-10-09T11:00:00.000Z"},
	)+costState("s1", 0.012))
	e.register(s)
	if res := e.collect(); res.New != 4 {
		t.Errorf("first collect result = %+v, want 4 new rows", res)
	}
	if res := e.collect(); res.New != 0 || res.Updated != 0 {
		t.Errorf("second collect result = %+v, want no changes", res)
	}

	// Resumed session: Claude Code writes a higher cumulative total.
	e.appendTo(s, costState("s1", 0.016))
	res := e.collect()
	if got, want := overheadCosts(e.rows()), map[int]int64{1: 60, 2: 20}; fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("overhead = %v, want %v", got, want)
	}
	if res.New != 0 || res.Updated != 2 {
		t.Errorf("third collect result = %+v, want 2 updated rows", res)
	}
}

func TestCollectOverheadByRequests(t *testing.T) {
	// No price, so no transcript cost: the gap is split by requests.
	e := newEnv(t)
	e.register(e.transcript("s1.jsonl", lines(
		line{MsgID: "m1", Branch: "feat/1-a", Out: 10, Model: "claude-fable-5-1"},
		line{MsgID: "m2", Branch: "feat/1-a", Out: 10, Model: "claude-fable-5-1"},
		line{MsgID: "m3", Branch: "feat/2-b", Out: 10, Model: "claude-fable-5-1"},
	)+costState("s1", 0.003)))
	e.collect()
	if got, want := overheadCosts(e.rows()), map[int]int64{1: 20, 2: 10}; fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("overhead = %v, want %v", got, want)
	}
}

func TestCollectOverheadOfDeletedTranscript(t *testing.T) {
	e := newEnv(t)
	s := e.transcript("s1.jsonl", lines(
		line{MsgID: "m1", Branch: "feat/1-a", Out: 300},
	)+costState("s1", 0.012))
	e.register(s)
	e.collect()
	before := e.csv()
	if err := os.Remove(s); err != nil {
		t.Fatal(err)
	}
	if res := e.collect(); res.New != 0 || res.Updated != 0 {
		t.Errorf("result = %+v, want no changes", res)
	}
	if got := e.csv(); got != before {
		t.Errorf("csv after transcript deletion:\n%s\nwant:\n%s", got, before)
	}
}

func TestCollectOverheadRemovedWithoutGap(t *testing.T) {
	// The transcript grows past the cost Claude Code computed, with no newer
	// cost-state yet: there is no gap left to record.
	e := newEnv(t)
	s := e.transcript("s1.jsonl", lines(
		line{MsgID: "m1", Branch: "feat/1-a", Out: 300},
	)+costState("s1", 0.012))
	e.register(s)
	e.collect()
	e.appendTo(s, lines(line{MsgID: "m2", Branch: "feat/1-a", Out: 600}))
	res := e.collect()
	if got := overheadCosts(e.rows()); len(got) != 0 {
		t.Errorf("overhead = %v, want none", got)
	}
	if res.New != 0 || res.Updated != 1 || res.Removed != 1 {
		t.Errorf("result = %+v, want the transcript row updated and the overhead row removed", res)
	}
}

func TestCollectOverheadOfSessionOutsideRepository(t *testing.T) {
	// Part of the session ran outside the repository, so the cost Claude Code
	// computed covers work that is not counted here.
	e := newEnv(t)
	s := e.transcript("s1.jsonl", lines(
		line{MsgID: "m1", Branch: "feat/1-a", Out: 300},
		line{MsgID: "m2", Branch: "feat/1-a", Out: 300, Cwd: "/elsewhere"},
	)+costState("s1", 0.012))
	e.register(s)
	res := e.collect()
	if got := overheadCosts(e.rows()); len(got) != 0 {
		t.Errorf("overhead = %v, want none", got)
	}
	if len(res.Warnings) != 1 || !strings.Contains(res.Warnings[0], "outside the repository") {
		t.Errorf("warnings = %q, want one about work outside the repository", res.Warnings)
	}
	if res := e.collect(); len(res.Warnings) != 0 {
		t.Errorf("second collect warnings = %q, want none", res.Warnings)
	}
}

func TestCollectOverheadRemovedWhenSessionLeavesRepository(t *testing.T) {
	// A resumed session works outside the repository after its overhead was
	// recorded: the stored overhead no longer holds.
	e := newEnv(t)
	s := e.transcript("s1.jsonl", lines(
		line{MsgID: "m1", Branch: "feat/1-a", Out: 300},
	)+costState("s1", 0.012))
	e.register(s)
	e.collect()
	e.appendTo(s, lines(
		line{MsgID: "m2", Branch: "feat/1-a", Out: 300},
		line{MsgID: "m3", Branch: "feat/1-a", Out: 300, Cwd: "/elsewhere"},
	)+costState("s1", 0.020))
	res := e.collect()
	if got := overheadCosts(e.rows()); len(got) != 0 {
		t.Errorf("overhead = %v, want none", got)
	}
	if res.Removed != 1 {
		t.Errorf("result = %+v, want the overhead row removed", res)
	}
}
