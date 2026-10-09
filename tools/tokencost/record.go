package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type hookInput struct {
	TranscriptPath string `json:"transcript_path"`
	Cwd            string `json:"cwd"`
}

// record appends the transcript path received from a Claude Code hook to the
// sessions log of the project. It reads no transcript and makes no network
// call: SessionEnd hooks may be cut short.
func record(stdin io.Reader, projectDir, home string) error {
	var in hookInput
	if err := json.NewDecoder(stdin).Decode(&in); err != nil {
		return fmt.Errorf("decode hook input: %w", err)
	}
	path := in.TranscriptPath
	if path == "" {
		return errors.New("hook input has no transcript_path")
	}
	if !strings.HasSuffix(path, ".jsonl") {
		return fmt.Errorf("transcript_path %q is not a .jsonl file", path)
	}
	if rest, ok := strings.CutPrefix(path, "~/"); ok {
		path = filepath.Join(home, rest)
	}
	root := projectDir
	if root == "" {
		root = in.Cwd
	}
	if root == "" {
		return errors.New("no project dir: CLAUDE_PROJECT_DIR and cwd are empty")
	}
	logPath := filepath.Join(root, ".claude", "sessions.log")
	registered, err := readLog(logPath)
	if err != nil {
		return err
	}
	for _, p := range registered {
		if p == path {
			return nil
		}
	}
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		return fmt.Errorf("create sessions log dir: %w", err)
	}
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open sessions log: %w", err)
	}
	if _, err := fmt.Fprintln(f, path); err != nil {
		f.Close()
		return fmt.Errorf("write sessions log: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("write sessions log: %w", err)
	}
	return nil
}
