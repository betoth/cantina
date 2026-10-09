package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunReport(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		gh       *fakeGitHub
		wantCode int
		wantErr  string
		comments int
	}{
		{"no issue", []string{"report"}, &fakeGitHub{}, 2, "-issue is required", 0},
		{"report without comment", []string{"report", "-issue", "11"}, &fakeGitHub{}, 0, "", 0},
		{"comment", []string{"report", "-issue", "11", "-comment"}, &fakeGitHub{}, 0, "comment created", 1},
		{"sub-issues fail without comment", []string{"report", "-issue", "11"}, &fakeGitHub{err: errors.New("offline")}, 0, "warning:", 0},
		{"sub-issues fail with comment", []string{"report", "-issue", "11", "-comment"}, &fakeGitHub{err: errors.New("offline")}, 1, "comment not posted", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newEnv(t)
			e.register(e.transcript("s1.jsonl", lines(line{MsgID: "m1", Branch: "feat/11-a", Out: 10})))
			e.collect()
			orig := newGitHub
			newGitHub = func(string) github { return tt.gh }
			t.Cleanup(func() { newGitHub = orig })

			var stdout, stderr bytes.Buffer
			args := append(tt.args, "-costs", e.cfg.OutPath)
			if code := run(args, nil, &stdout, &stderr); code != tt.wantCode {
				t.Errorf("exit code = %d, want %d; stderr: %s", code, tt.wantCode, stderr.String())
			}
			if !strings.Contains(stderr.String(), tt.wantErr) {
				t.Errorf("stderr = %q, want %q", stderr.String(), tt.wantErr)
			}
			if tt.wantCode == 0 && !strings.Contains(stdout.String(), "## Custo de IA da #11") {
				t.Errorf("stdout = %q, want the report", stdout.String())
			}
			if got := len(tt.gh.comments[11]); got != tt.comments {
				t.Errorf("comments = %d, want %d", got, tt.comments)
			}
		})
	}
}

func TestRunChartDefaultPaths(t *testing.T) {
	root := t.TempDir()
	e := newEnv(t)
	e.cfg.OutPath = filepath.Join(root, "docs", "ai-costs.csv")
	e.register(e.transcript("s1.jsonl", lines(line{MsgID: "m1", Branch: "feat/11-a", Out: 10})))
	e.collect()
	var stdout, stderr bytes.Buffer
	if code := run([]string{"chart", "-root", root}, nil, &stdout, &stderr); code != 0 {
		t.Fatalf("exit code = %d; stderr: %s", code, stderr.String())
	}
	b, err := os.ReadFile(filepath.Join(root, "docs", "ai-costs.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `x-axis ["#11"]`) {
		t.Errorf("ai-costs.md:\n%s\nwant a bar for #11", b)
	}
}
