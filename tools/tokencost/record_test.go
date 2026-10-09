package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRecordRejects(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{"invalid json", `{"transcript_path":`},
		{"no transcript_path", `{"cwd":"/repo"}`},
		{"not a jsonl file", `{"transcript_path":"/home/u/.env"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			if err := record(strings.NewReader(tt.in), root, "/home/u"); err == nil {
				t.Error("record returned no error")
			}
			if _, err := os.Stat(filepath.Join(root, ".claude")); !os.IsNotExist(err) {
				t.Errorf("record wrote to the project: stat err = %v", err)
			}
		})
	}
}

func TestRecord(t *testing.T) {
	tests := []struct {
		name       string
		in         string
		projectDir bool
		want       string
	}{
		{"project dir wins over cwd", `{"transcript_path":"/home/u/.claude/projects/p/s.jsonl","cwd":"/nonexistent/sub"}`, true, "/home/u/.claude/projects/p/s.jsonl"},
		{"expands home", `{"transcript_path":"~/.claude/projects/p/s.jsonl","cwd":"%ROOT%"}`, true, "/home/u/.claude/projects/p/s.jsonl"},
		{"root from cwd when no project dir", `{"transcript_path":"/t/s.jsonl","cwd":"%ROOT%"}`, false, "/t/s.jsonl"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			in := strings.ReplaceAll(tt.in, "%ROOT%", root)
			projectDir := ""
			if tt.projectDir {
				projectDir = root
			}
			// Twice: start and end hooks of the same session register it once.
			for range 2 {
				if err := record(strings.NewReader(in), projectDir, "/home/u"); err != nil {
					t.Fatal(err)
				}
			}
			b, err := os.ReadFile(filepath.Join(root, ".claude", "sessions.log"))
			if err != nil {
				t.Fatal(err)
			}
			if got := string(b); got != tt.want+"\n" {
				t.Errorf("sessions.log = %q, want %q", got, tt.want+"\n")
			}
		})
	}
}
