package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var branchIssue = regexp.MustCompile(`(?:^|/)(\d+)-`)

// issueFromBranch returns the issue number encoded in a branch name, or 0.
func issueFromBranch(branch string) int {
	m := branchIssue.FindStringSubmatch(branch)
	if m == nil {
		return 0
	}
	n, err := strconv.Atoi(m[1])
	if err != nil {
		return 0
	}
	return n
}

// usage holds token counts by price category.
type usage struct {
	Input        int64
	Output       int64
	CacheWrite5m int64
	CacheWrite1h int64
	CacheRead    int64
}

func (u usage) add(o usage) usage {
	return usage{
		Input:        u.Input + o.Input,
		Output:       u.Output + o.Output,
		CacheWrite5m: u.CacheWrite5m + o.CacheWrite5m,
		CacheWrite1h: u.CacheWrite1h + o.CacheWrite1h,
		CacheRead:    u.CacheRead + o.CacheRead,
	}
}

func (u usage) sub(o usage) usage {
	return usage{
		Input:        u.Input - o.Input,
		Output:       u.Output - o.Output,
		CacheWrite5m: u.CacheWrite5m - o.CacheWrite5m,
		CacheWrite1h: u.CacheWrite1h - o.CacheWrite1h,
		CacheRead:    u.CacheRead - o.CacheRead,
	}
}

// covers reports whether every count in u is at least the one in o.
func (u usage) covers(o usage) bool {
	return u.Input >= o.Input && u.Output >= o.Output &&
		u.CacheWrite5m >= o.CacheWrite5m && u.CacheWrite1h >= o.CacheWrite1h &&
		u.CacheRead >= o.CacheRead
}

// response is one model response, deduplicated across transcript lines.
type response struct {
	Key       string
	SessionID string
	Issue     int
	Model     string
	Cwd       string
	Time      time.Time
	Usage     usage
	// LegacyCache means the line had no per-TTL cache breakdown.
	LegacyCache bool
	Fast        bool
	WebSearches int64
}

type rawLine struct {
	Type      string    `json:"type"`
	SessionID string    `json:"sessionId"`
	RequestID string    `json:"requestId"`
	GitBranch string    `json:"gitBranch"`
	Cwd       string    `json:"cwd"`
	Timestamp time.Time `json:"timestamp"`
	Message   *struct {
		ID    string    `json:"id"`
		Model string    `json:"model"`
		Usage *rawUsage `json:"usage"`
	} `json:"message"`
}

type rawUsage struct {
	InputTokens              int64 `json:"input_tokens"`
	OutputTokens             int64 `json:"output_tokens"`
	CacheCreationInputTokens int64 `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int64 `json:"cache_read_input_tokens"`
	CacheCreation            *struct {
		Ephemeral5m int64 `json:"ephemeral_5m_input_tokens"`
		Ephemeral1h int64 `json:"ephemeral_1h_input_tokens"`
	} `json:"cache_creation"`
	ServerToolUse *struct {
		WebSearchRequests int64 `json:"web_search_requests"`
	} `json:"server_tool_use"`
	Speed string `json:"speed"`
}

// parseLine converts one transcript line into a response. ok is false for
// lines that are not model responses with usage, including invalid JSON.
func parseLine(b []byte) (response, bool) {
	var l rawLine
	if err := json.Unmarshal(b, &l); err != nil {
		return response{}, false
	}
	if l.Type != "assistant" || l.Message == nil || l.Message.Usage == nil || l.Message.ID == "" {
		return response{}, false
	}
	// Claude Code writes local messages with models like "<synthetic>"; they
	// are not API calls.
	if strings.HasPrefix(l.Message.Model, "<") {
		return response{}, false
	}
	u := l.Message.Usage
	r := response{
		Key:       l.Message.ID + "|" + l.RequestID,
		SessionID: l.SessionID,
		Issue:     issueFromBranch(l.GitBranch),
		Model:     l.Message.Model,
		Cwd:       l.Cwd,
		Time:      l.Timestamp.UTC(),
		Usage: usage{
			Input:     u.InputTokens,
			Output:    u.OutputTokens,
			CacheRead: u.CacheReadInputTokens,
		},
		Fast: u.Speed == "fast",
	}
	if u.CacheCreation != nil {
		r.Usage.CacheWrite5m = u.CacheCreation.Ephemeral5m
		r.Usage.CacheWrite1h = u.CacheCreation.Ephemeral1h
	} else {
		r.Usage.CacheWrite5m = u.CacheCreationInputTokens
		r.LegacyCache = u.CacheCreationInputTokens > 0
	}
	if u.ServerToolUse != nil {
		r.WebSearches = u.ServerToolUse.WebSearchRequests
	}
	return r, true
}

// transcripts holds what was read from all transcripts.
type transcripts struct {
	// responses is deduplicated by key. Streaming writes the same response in
	// several lines with partial usage; the line with the most output tokens
	// is the final one. The response time is the earliest of its lines, so
	// the result does not depend on reading order.
	responses map[string]response
	// sessionCost is the total cost Claude Code computed for each session, in
	// units of costUnit, from its cost-state lines. It also covers API calls
	// that the transcript does not record.
	sessionCost map[string]int64
}

func newTranscripts() *transcripts {
	return &transcripts{responses: map[string]response{}, sessionCost: map[string]int64{}}
}

func (t *transcripts) add(r response) {
	old, ok := t.responses[r.Key]
	if !ok {
		t.responses[r.Key] = r
		return
	}
	first := old.Time
	if r.Time.Before(first) {
		first = r.Time
	}
	if r.Usage.Output > old.Usage.Output {
		old = r
	}
	old.Time = first
	t.responses[r.Key] = old
}

type rawCostState struct {
	Type         string  `json:"type"`
	SessionID    string  `json:"sessionId"`
	TotalCostUSD float64 `json:"totalCostUSD"`
}

// addCostState keeps the highest total of a session: cost-state lines are
// cumulative.
func (t *transcripts) addCostState(b []byte) {
	var c rawCostState
	if err := json.Unmarshal(b, &c); err != nil || c.Type != "cost-state" || c.SessionID == "" {
		return
	}
	units := int64(math.Round(c.TotalCostUSD / costUnit))
	if units > t.sessionCost[c.SessionID] {
		t.sessionCost[c.SessionID] = units
	}
}

// read adds every response and cost state in the transcript. Lines are read
// with ReadBytes because transcript lines can exceed bufio.Scanner's token
// limit.
func (t *transcripts) read(r io.Reader) error {
	br := bufio.NewReader(r)
	for {
		b, err := br.ReadBytes('\n')
		if len(b) > 0 {
			if resp, ok := parseLine(b); ok {
				t.add(resp)
			} else if bytes.Contains(b, []byte(`"cost-state"`)) {
				t.addCostState(b)
			}
		}
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

// readFile adds the contents of a transcript file. A missing file is not an
// error: Claude Code deletes old transcripts.
func (t *transcripts) readFile(path string) error {
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("open transcript: %w", err)
	}
	defer f.Close()
	if err := t.read(f); err != nil {
		return fmt.Errorf("read transcript %s: %w", path, err)
	}
	return nil
}

// subagentTranscripts returns the transcripts of the subagents of a session,
// stored in <session>/subagents/ next to the session transcript.
func subagentTranscripts(sessionPath string) ([]string, error) {
	dir := strings.TrimSuffix(sessionPath, ".jsonl")
	return filepath.Glob(filepath.Join(dir, "subagents", "*.jsonl"))
}
