package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestIssueFromBranch(t *testing.T) {
	tests := []struct {
		branch string
		want   int
	}{
		{"feat/42-x", 42},
		{"42-x", 42},
		{"fix/7-y", 7},
		{"main", 0},
		{"feature/abc", 0},
		{"", 0},
		{"HEAD", 0},
		{"tmp/6-rebase", 6},
		{"v2-release", 0},
	}
	for _, tt := range tests {
		t.Run(tt.branch, func(t *testing.T) {
			if got := issueFromBranch(tt.branch); got != tt.want {
				t.Errorf("issueFromBranch(%q) = %d, want %d", tt.branch, got, tt.want)
			}
		})
	}
}

// line describes a transcript line for tests.
type line struct {
	Session, Branch, Cwd, Time, MsgID, ReqID, Model string
	In, Out, W5m, W1h, Read                         int64
	// Legacy omits the per-TTL cache breakdown and puts W5m in the total.
	Legacy bool
	Speed  string
}

func (l line) String() string {
	if l.Session == "" {
		l.Session = "s1"
	}
	if l.Cwd == "" {
		l.Cwd = "/repo"
	}
	if l.Time == "" {
		l.Time = "2026-10-09T10:00:00.000Z"
	}
	if l.Model == "" {
		l.Model = "claude-opus-5-5"
	}
	if l.ReqID == "" {
		l.ReqID = "req_" + l.MsgID
	}
	u := map[string]any{
		"input_tokens":                l.In,
		"output_tokens":               l.Out,
		"cache_creation_input_tokens": l.W5m + l.W1h,
		"cache_read_input_tokens":     l.Read,
		"speed":                       l.Speed,
	}
	if !l.Legacy {
		u["cache_creation"] = map[string]int64{"ephemeral_5m_input_tokens": l.W5m, "ephemeral_1h_input_tokens": l.W1h}
	}
	// Maps of strings and numbers always marshal.
	b, _ := json.Marshal(map[string]any{
		"type": "assistant", "sessionId": l.Session, "requestId": l.ReqID, "gitBranch": l.Branch,
		"cwd": l.Cwd, "timestamp": l.Time,
		"message": map[string]any{"id": l.MsgID, "model": l.Model, "content": []any{}, "usage": u},
	})
	return string(b)
}

func lines(ls ...line) string {
	var b strings.Builder
	for _, l := range ls {
		b.WriteString(l.String())
		b.WriteByte('\n')
	}
	return b.String()
}

func TestParseLineSkips(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{"invalid json", `{"type":"assistant", not json`},
		{"user line", `{"type":"user","message":{"role":"user","content":"..."}}`},
		{"assistant without usage", `{"type":"assistant","message":{"id":"m1","model":"claude-opus-5-5"}}`},
		{"synthetic model", line{MsgID: "m1", Model: "<synthetic>"}.String()},
		{"no message id", line{}.String()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if r, ok := parseLine([]byte(tt.in)); ok {
				t.Errorf("parseLine(%s) = %+v, want skipped", tt.in, r)
			}
		})
	}
}

func TestParseLineCacheTTL(t *testing.T) {
	tests := []struct {
		name       string
		in         line
		want       usage
		wantLegacy bool
	}{
		{
			name: "per TTL breakdown",
			in:   line{MsgID: "m1", In: 1, Out: 2, W5m: 3, W1h: 4, Read: 5},
			want: usage{Input: 1, Output: 2, CacheWrite5m: 3, CacheWrite1h: 4, CacheRead: 5},
		},
		{
			name:       "legacy total counts as 5 minutes",
			in:         line{MsgID: "m1", In: 1, Out: 2, W5m: 7, Read: 5, Legacy: true},
			want:       usage{Input: 1, Output: 2, CacheWrite5m: 7, CacheRead: 5},
			wantLegacy: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, ok := parseLine([]byte(tt.in.String()))
			if !ok {
				t.Fatal("parseLine skipped a valid line")
			}
			if r.Usage != tt.want || r.LegacyCache != tt.wantLegacy {
				t.Errorf("got usage %+v legacy %v, want %+v legacy %v", r.Usage, r.LegacyCache, tt.want, tt.wantLegacy)
			}
		})
	}
}

func TestResponsesDedup(t *testing.T) {
	partial := line{MsgID: "m1", In: 2, Out: 6, W1h: 100, Read: 50}
	final := partial
	final.Out = 523
	other := line{MsgID: "m2", In: 1, Out: 10}
	tests := []struct {
		name  string
		files []string
	}{
		{"three partial lines and the final one", []string{lines(partial, partial, partial, final, other)}},
		{"final line before partial ones", []string{lines(final, partial, partial, other)}},
		{"same response in two transcripts", []string{lines(partial, final), lines(final, other)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := newTranscripts()
			for _, f := range tt.files {
				if err := ts.read(strings.NewReader(f)); err != nil {
					t.Fatal(err)
				}
			}
			if len(ts.responses) != 2 {
				t.Fatalf("got %d responses, want 2", len(ts.responses))
			}
			got := ts.responses["m1|req_m1"].Usage
			want := usage{Input: 2, Output: 523, CacheWrite1h: 100, CacheRead: 50}
			if got != want {
				t.Errorf("m1 usage = %+v, want %+v", got, want)
			}
		})
	}
}

func TestResponsesReadLongLine(t *testing.T) {
	long := `{"type":"user","message":{"content":"` + strings.Repeat("x", 1<<20) + `"}}` + "\n"
	ts := newTranscripts()
	if err := ts.read(strings.NewReader(long + lines(line{MsgID: "m1", Out: 1}))); err != nil {
		t.Fatal(err)
	}
	if len(ts.responses) != 1 {
		t.Errorf("got %d responses, want 1", len(ts.responses))
	}
}
