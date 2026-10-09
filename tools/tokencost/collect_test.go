package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const fixtureSession = "16bb9a95-140c-4927-b514-c910fe073009"

var opusPrice = price{Input: 4, Output: 20, CacheWrite5m: 5, CacheWrite1h: 8, CacheRead: 0.2}

// env is a temporary repository with its own Claude Code projects dir.
type env struct {
	t   *testing.T
	cfg collectConfig
}

func newEnv(t *testing.T) *env {
	t.Helper()
	dir := t.TempDir()
	e := &env{t: t, cfg: collectConfig{
		Root:        "/repo",
		LogPath:     filepath.Join(dir, "sessions.log"),
		PricesPath:  filepath.Join(dir, "prices.json"),
		OutPath:     filepath.Join(dir, "docs", "ai-costs.csv"),
		ProjectsDir: filepath.Join(dir, "projects"),
	}}
	e.setPrices(map[string]price{"claude-opus-5-5": opusPrice})
	return e
}

func (e *env) setPrices(models map[string]price) {
	e.t.Helper()
	b, err := json.Marshal(priceTable{Models: models})
	if err != nil {
		e.t.Fatal(err)
	}
	e.write(e.cfg.PricesPath, string(b))
}

func (e *env) write(path, content string) {
	e.t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		e.t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		e.t.Fatal(err)
	}
}

func (e *env) appendTo(path, content string) {
	e.t.Helper()
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		e.t.Fatal(err)
	}
	defer f.Close()
	if _, err := f.WriteString(content); err != nil {
		e.t.Fatal(err)
	}
}

// transcript writes a transcript under the projects dir and returns its path.
func (e *env) transcript(name, content string) string {
	e.t.Helper()
	p := filepath.Join(e.cfg.ProjectsDir, "-repo", name)
	e.write(p, content)
	return p
}

func (e *env) register(paths ...string) {
	e.t.Helper()
	e.write(e.cfg.LogPath, strings.Join(paths, "\n")+"\n")
}

func (e *env) collect() collectResult {
	e.t.Helper()
	res, err := collect(e.cfg)
	if err != nil {
		e.t.Fatal(err)
	}
	return res
}

func (e *env) rows() []row {
	e.t.Helper()
	rows, err := readRows(e.cfg.OutPath)
	if err != nil {
		e.t.Fatal(err)
	}
	return rows
}

func (e *env) csv() string {
	e.t.Helper()
	b, err := os.ReadFile(e.cfg.OutPath)
	if err != nil {
		e.t.Fatal(err)
	}
	return string(b)
}

func TestCollectFixture(t *testing.T) {
	e := newEnv(t)
	abs, err := filepath.Abs(filepath.Join("testdata", fixtureSession+".jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	e.register(abs)
	res := e.collect()
	// Expected values computed independently with jq over the fixture.
	want := `session_id,issue,model,start,end,requests,input_tokens,output_tokens,cache_write_5m_tokens,cache_write_1h_tokens,cache_read_tokens,cost_usd
16bb9a95-140c-4927-b514-c910fe073009,0,claude-opus-5-5,2026-10-05T18:54:47.036Z,2026-10-06T00:52:48.734Z,13,26,6585,0,46159,1501875,0.8015
16bb9a95-140c-4927-b514-c910fe073009,3,claude-opus-5-5,2026-10-06T00:25:47.932Z,2026-10-06T00:47:12.774Z,31,62,16676,0,23341,6587132,1.8379
16bb9a95-140c-4927-b514-c910fe073009,2,claude-opus-5-5,2026-10-06T00:48:00.883Z,2026-10-06T01:15:33.443Z,51,102,27674,34841,53821,11407096,3.4401
`
	if got := e.csv(); got != want {
		t.Errorf("csv:\n%s\nwant:\n%s", got, want)
	}
	if res.New != 3 || res.Updated != 0 || len(res.Warnings) != 0 {
		t.Errorf("result = %+v, want 3 new rows, no warnings", res)
	}
}

func TestCollectIdempotent(t *testing.T) {
	e := newEnv(t)
	e.register(e.transcript("s1.jsonl", lines(
		line{MsgID: "m1", Branch: "feat/1-a", Out: 10},
		line{MsgID: "m2", Branch: "feat/2-b", Out: 20, Time: "2026-10-09T09:00:00.000Z"},
	)))
	e.collect()
	first := e.csv()
	res := e.collect()
	if got := e.csv(); got != first {
		t.Errorf("second collect changed the file:\n%s\nwant:\n%s", got, first)
	}
	if res.New != 0 || res.Updated != 0 {
		t.Errorf("second collect result = %+v, want no changes", res)
	}
}

func TestCollectRows(t *testing.T) {
	tests := []struct {
		name  string
		setup func(e *env)
		want  []row
	}{
		{
			name: "session in two branches gives one row per issue",
			setup: func(e *env) {
				e.register(e.transcript("s1.jsonl", lines(
					line{MsgID: "m1", Branch: "feat/1-a", Out: 10, Time: "2026-10-09T10:00:00.000Z"},
					line{MsgID: "m2", Branch: "feat/1-a", Out: 10, Time: "2026-10-09T10:05:00.000Z"},
					line{MsgID: "m3", Branch: "fix/2-b", Out: 30, Time: "2026-10-09T11:00:00.000Z"},
				)))
			},
			want: []row{
				{SessionID: "s1", Issue: 1, Model: "claude-opus-5-5", Requests: 2, Usage: usage{Output: 20}, Cost: 4},
				{SessionID: "s1", Issue: 2, Model: "claude-opus-5-5", Requests: 1, Usage: usage{Output: 30}, Cost: 6},
			},
		},
		{
			name: "subagent transcripts of a registered session count",
			setup: func(e *env) {
				s := e.transcript("s1.jsonl", lines(line{MsgID: "m1", Branch: "feat/1-a", Out: 10}))
				e.transcript("s1/subagents/agent-a.jsonl", lines(line{MsgID: "m2", Branch: "feat/1-a", Out: 40}))
				e.register(s)
			},
			want: []row{
				{SessionID: "s1", Issue: 1, Model: "claude-opus-5-5", Requests: 2, Usage: usage{Output: 50}, Cost: 10},
			},
		},
		{
			name: "use outside an issue branch goes to issue 0",
			setup: func(e *env) {
				e.register(e.transcript("s1.jsonl", lines(line{MsgID: "m1", Branch: "main", Out: 10})))
			},
			want: []row{
				{SessionID: "s1", Issue: 0, Model: "claude-opus-5-5", Requests: 1, Usage: usage{Output: 10}, Cost: 2},
			},
		},
		{
			name: "import-old reads only transcripts of this repository",
			setup: func(e *env) {
				e.cfg.ImportOld = true
				e.transcript("s1.jsonl", lines(line{MsgID: "m1", Branch: "feat/1-a", Out: 10, Cwd: "/repo/sub"}))
				e.transcript("s1/subagents/agent-a.jsonl", lines(line{MsgID: "m2", Branch: "feat/1-a", Out: 40}))
				e.transcript("s2.jsonl", lines(line{Session: "s2", MsgID: "m3", Branch: "feat/1-a", Out: 10, Cwd: "/repository"}))
			},
			want: []row{
				{SessionID: "s1", Issue: 1, Model: "claude-opus-5-5", Requests: 2, Usage: usage{Output: 50}, Cost: 10},
			},
		},
		{
			name: "import-old reads nested subagent transcripts",
			setup: func(e *env) {
				e.cfg.ImportOld = true
				e.transcript("s1.jsonl", lines(line{MsgID: "m1", Branch: "feat/1-a", Out: 10}))
				e.transcript("s1/subagents/agent-a/subagents/agent-b.jsonl", lines(line{MsgID: "m2", Branch: "feat/1-a", Out: 40}))
			},
			want: []row{
				{SessionID: "s1", Issue: 1, Model: "claude-opus-5-5", Requests: 2, Usage: usage{Output: 50}, Cost: 10},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newEnv(t)
			tt.setup(e)
			e.collect()
			got := e.rows()
			if len(got) != len(tt.want) {
				t.Fatalf("got %d rows, want %d:\n%s", len(got), len(tt.want), e.csv())
			}
			for i, w := range tt.want {
				g := got[i]
				if g.key() != w.key() || g.Requests != w.Requests || g.Usage != w.Usage || g.Cost != w.Cost {
					t.Errorf("row %d = %+v, want %+v", i, g, w)
				}
			}
		})
	}
}

func TestCollectUnpricedModel(t *testing.T) {
	e := newEnv(t)
	e.register(e.transcript("s1.jsonl", lines(line{MsgID: "m1", Branch: "feat/1-a", Out: 10, Model: "claude-fable-5-1"})))
	res := e.collect()
	rows := e.rows()
	if len(rows) != 1 || rows[0].Cost != 0 {
		t.Errorf("rows = %+v, want one row with cost 0", rows)
	}
	if len(res.Warnings) != 1 || !strings.Contains(res.Warnings[0], "claude-fable-5-1") {
		t.Errorf("warnings = %q, want one about claude-fable-5-1", res.Warnings)
	}
}

func TestCollectPriceChange(t *testing.T) {
	e := newEnv(t)
	s := e.transcript("s1.jsonl", lines(
		line{MsgID: "m1", Branch: "feat/1-a", Out: 1000},
		line{MsgID: "m2", Branch: "feat/2-b", Out: 1000},
	))
	e.register(s)
	e.collect() // 1000 × 20 = US$ 0.02 per row

	e.setPrices(map[string]price{"claude-opus-5-5": {Output: 50}})
	e.appendTo(s, lines(line{MsgID: "m3", Branch: "feat/2-b", Out: 1000}))
	res := e.collect()

	rows := e.rows()
	costs := map[int]int64{}
	for _, r := range rows {
		costs[r.Issue] = r.Cost
	}
	// Issue 1 had no new tokens and keeps US$ 0.02; issue 2 adds 1000 × 50 = US$ 0.05.
	want := map[int]int64{1: 200, 2: 700}
	for issue, c := range want {
		if costs[issue] != c {
			t.Errorf("issue %d cost = %s, want %s", issue, formatCost(costs[issue]), formatCost(c))
		}
	}
	if res.New != 0 || res.Updated != 1 {
		t.Errorf("result = %+v, want 1 updated row", res)
	}
}

func TestCollectKeepsRowOfDeletedTranscript(t *testing.T) {
	e := newEnv(t)
	s := e.transcript("s1.jsonl", lines(line{MsgID: "m1", Branch: "feat/1-a", Out: 10}))
	e.register(s)
	e.collect()
	before := e.csv()
	if err := os.Remove(s); err != nil {
		t.Fatal(err)
	}
	e.collect()
	if got := e.csv(); got != before {
		t.Errorf("csv after transcript deletion:\n%s\nwant:\n%s", got, before)
	}
}

func TestStoreRoundTrip(t *testing.T) {
	e := newEnv(t)
	e.register(e.transcript("s1.jsonl", lines(line{MsgID: "m1", Branch: "feat/1-a", In: 3, Out: 10, W5m: 4, W1h: 5, Read: 6})))
	e.collect()
	var buf bytes.Buffer
	if err := encodeRows(&buf, e.rows()); err != nil {
		t.Fatal(err)
	}
	if got := buf.String(); got != e.csv() {
		t.Errorf("re-encoded csv:\n%s\nwant:\n%s", got, e.csv())
	}
}

func TestCollectFewerTokensKeepsRow(t *testing.T) {
	e := newEnv(t)
	s := e.transcript("s1.jsonl", lines(line{MsgID: "m1", Branch: "feat/1-a", Out: 100}))
	e.register(s)
	e.collect()
	before := e.csv()
	e.write(s, lines(line{MsgID: "m1", Branch: "feat/1-a", Out: 50}))
	res := e.collect()
	if got := e.csv(); got != before {
		t.Errorf("csv:\n%s\nwant:\n%s", got, before)
	}
	if len(res.Warnings) != 1 || !strings.Contains(res.Warnings[0], "fewer tokens") {
		t.Errorf("warnings = %q, want one about fewer tokens", res.Warnings)
	}
}

func TestCollectNotesOnlyForChangedRows(t *testing.T) {
	e := newEnv(t)
	s := e.transcript("s1.jsonl", lines(
		line{MsgID: "m1", Branch: "feat/1-a", Out: 10, W5m: 100, Legacy: true},
		line{MsgID: "m2", Branch: "feat/2-b", Out: 10, Speed: "fast"},
	))
	e.register(s)
	if res := e.collect(); len(res.Warnings) != 2 {
		t.Errorf("first collect warnings = %q, want legacy cache and fast mode", res.Warnings)
	}
	if res := e.collect(); len(res.Warnings) != 0 {
		t.Errorf("second collect warnings = %q, want none", res.Warnings)
	}
	e.appendTo(s, lines(line{MsgID: "m3", Branch: "feat/2-b", Out: 10}))
	res := e.collect()
	if len(res.Warnings) != 1 || !strings.Contains(res.Warnings[0], "fast mode") {
		t.Errorf("third collect warnings = %q, want only fast mode of the updated row", res.Warnings)
	}
}

func TestCollectChanges(t *testing.T) {
	e := newEnv(t)
	s := e.transcript("s1.jsonl", lines(line{MsgID: "m1", Branch: "feat/1-a", Out: 300})+costState("s1", 0.012))
	e.register(s)
	if got, want := changes(e.collect()), "new s1 1 claude-opus-5-5 0.0060, new s1 1 overhead 0.0060"; got != want {
		t.Errorf("first collect changes = %q, want %q", got, want)
	}
	e.appendTo(s, lines(line{MsgID: "m2", Branch: "feat/1-a", Out: 600}))
	if got, want := changes(e.collect()), "updated s1 1 claude-opus-5-5 0.0180, removed s1 1 overhead 0.0060"; got != want {
		t.Errorf("second collect changes = %q, want %q", got, want)
	}
}

func changes(res collectResult) string {
	var cs []string
	for _, c := range res.Changes {
		cs = append(cs, fmt.Sprintf("%s %s %d %s %s", c.Kind, c.Row.SessionID, c.Row.Issue, c.Row.Model, formatCost(c.Row.Cost)))
	}
	return strings.Join(cs, ", ")
}

func TestReadRowsRejectsUnknownHeader(t *testing.T) {
	e := newEnv(t)
	e.write(e.cfg.OutPath, "issue,session_id\n1,s1\n")
	if _, err := readRows(e.cfg.OutPath); err == nil || !strings.Contains(err.Error(), "header") {
		t.Errorf("readRows error = %v, want one about the header", err)
	}
}
