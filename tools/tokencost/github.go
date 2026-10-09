package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// github is what report needs from the GitHub repository of the project.
type github interface {
	SubIssues(ctx context.Context, issue int) ([]int, error)
	Comments(ctx context.Context, issue int) ([]comment, error)
	CreateComment(ctx context.Context, issue int, body string) error
	UpdateComment(ctx context.Context, id int64, body string) error
}

type comment struct {
	ID   int64  `json:"id"`
	Body string `json:"body"`
}

// ghCLI talks to GitHub through the gh CLI, run in the repository so that
// {owner}/{repo} resolves to it.
type ghCLI struct {
	dir string
}

func (g ghCLI) run(ctx context.Context, stdin []byte, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "gh", append([]string{"api"}, args...)...)
	cmd.Dir = g.dir
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("gh api %s: %w: %s", args[len(args)-1], err, strings.TrimSpace(stderr.String()))
	}
	return out, nil
}

func (g ghCLI) SubIssues(ctx context.Context, issue int) ([]int, error) {
	out, err := g.run(ctx, nil, "--paginate", "--jq", ".[].number",
		fmt.Sprintf("repos/{owner}/{repo}/issues/%d/sub_issues", issue))
	if err != nil {
		return nil, err
	}
	var numbers []int
	for _, f := range strings.Fields(string(out)) {
		n, err := strconv.Atoi(f)
		if err != nil {
			return nil, fmt.Errorf("sub-issue number %q: %w", f, err)
		}
		numbers = append(numbers, n)
	}
	return numbers, nil
}

func (g ghCLI) Comments(ctx context.Context, issue int) ([]comment, error) {
	out, err := g.run(ctx, nil, "--paginate", "--jq", ".[] | {id, body}",
		fmt.Sprintf("repos/{owner}/{repo}/issues/%d/comments", issue))
	if err != nil {
		return nil, err
	}
	var cs []comment
	s := bufio.NewScanner(bytes.NewReader(out))
	s.Buffer(nil, 1<<24)
	for s.Scan() {
		var c comment
		if err := json.Unmarshal(s.Bytes(), &c); err != nil {
			return nil, fmt.Errorf("decode comment: %w", err)
		}
		cs = append(cs, c)
	}
	return cs, s.Err()
}

func (g ghCLI) CreateComment(ctx context.Context, issue int, body string) error {
	return g.send(ctx, "POST", fmt.Sprintf("repos/{owner}/{repo}/issues/%d/comments", issue), body)
}

func (g ghCLI) UpdateComment(ctx context.Context, id int64, body string) error {
	return g.send(ctx, "PATCH", fmt.Sprintf("repos/{owner}/{repo}/issues/comments/%d", id), body)
}

func (g ghCLI) send(ctx context.Context, method, path, body string) error {
	b, err := json.Marshal(map[string]string{"body": body})
	if err != nil {
		return err
	}
	_, err = g.run(ctx, b, "--method", method, "--input", "-", path)
	return err
}
