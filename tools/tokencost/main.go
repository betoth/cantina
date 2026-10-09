// Command tokencost measures the Claude Code token cost of each GitHub issue
// from local transcripts and stores the numbers in the repository.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const usageText = `usage: tokencost <command> [flags]

commands:
  record    register the transcript of a session (called by Claude Code hooks)
  collect   read registered transcripts and update the costs file
  report    print the cost of an issue and its sub-issues, optionally as an issue comment
  chart     write the cost charts to docs/ai-costs.md
`

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, usageText)
		return 2
	}
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintln(stderr, "tokencost:", err)
		return 1
	}
	switch args[0] {
	case "record":
		if err := record(stdin, os.Getenv("CLAUDE_PROJECT_DIR"), home); err != nil {
			fmt.Fprintln(stderr, "tokencost record:", err)
			return 1
		}
		return 0
	case "collect":
		return runCollect(args[1:], home, stdout, stderr)
	case "report":
		return runReport(args[1:], stdout, stderr)
	case "chart":
		return runChart(args[1:], stdout, stderr)
	default:
		fmt.Fprint(stderr, usageText)
		return 2
	}
}

func runCollect(args []string, home string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("collect", flag.ContinueOnError)
	fs.SetOutput(stderr)
	root := fs.String("root", ".", "repository root")
	importOld := fs.Bool("import-old", false, "also read every transcript of this repository under -projects")
	prices := fs.String("prices", "", "price table (default <root>/.claude/prices.json)")
	logPath := fs.String("log", "", "sessions log (default <root>/.claude/sessions.log)")
	out := fs.String("out", "", "costs file (default <root>/docs/ai-costs.csv)")
	projects := fs.String("projects", filepath.Join(home, ".claude", "projects"), "Claude Code projects dir")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	absRoot, err := filepath.Abs(*root)
	if err != nil {
		fmt.Fprintln(stderr, "tokencost collect:", err)
		return 1
	}
	orDefault := func(v string, parts ...string) string {
		if v != "" {
			return v
		}
		return filepath.Join(append([]string{absRoot}, parts...)...)
	}
	res, err := collect(collectConfig{
		Root:        absRoot,
		LogPath:     orDefault(*logPath, ".claude", "sessions.log"),
		PricesPath:  orDefault(*prices, ".claude", "prices.json"),
		OutPath:     orDefault(*out, "docs", "ai-costs.csv"),
		ProjectsDir: *projects,
		ImportOld:   *importOld,
	})
	if err != nil {
		fmt.Fprintln(stderr, "tokencost collect:", err)
		return 1
	}
	for _, c := range res.Changes {
		r := c.Row
		fmt.Fprintf(stdout, "%-7s %s #%d %s US$ %s\n", c.Kind, r.SessionID, r.Issue, r.Model, formatCost(r.Cost))
	}
	for _, w := range res.Warnings {
		fmt.Fprintln(stderr, "warning:", w)
	}
	fmt.Fprintf(stdout, "%d new rows, %d updated rows, %d removed rows, %d warnings\n", res.New, res.Updated, res.Removed, len(res.Warnings))
	return 0
}

// newGitHub is replaced in tests.
var newGitHub = func(dir string) github { return ghCLI{dir: dir} }

func runReport(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("report", flag.ContinueOnError)
	fs.SetOutput(stderr)
	root := fs.String("root", ".", "repository root")
	costs := fs.String("costs", "", "costs file (default <root>/docs/ai-costs.csv)")
	issue := fs.Int("issue", 0, "issue number")
	post := fs.Bool("comment", false, "also create or update the cost comment on the issue")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *issue <= 0 {
		fmt.Fprintln(stderr, "tokencost report: -issue is required")
		return 2
	}
	if *costs == "" {
		*costs = filepath.Join(*root, "docs", "ai-costs.csv")
	}
	rows, err := readRows(*costs)
	if err != nil {
		fmt.Fprintln(stderr, "tokencost report:", err)
		return 1
	}
	ctx := context.Background()
	gh := newGitHub(*root)
	body, treeErr := report(ctx, rows, *issue, gh)
	fmt.Fprint(stdout, body)
	if !*post {
		if treeErr != nil {
			fmt.Fprintln(stderr, "warning:", treeErr)
		}
		return 0
	}
	// A comment without the sub-issues would replace a complete one.
	if treeErr != nil {
		fmt.Fprintln(stderr, "tokencost report: comment not posted:", treeErr)
		return 1
	}
	created, err := upsertComment(ctx, gh, *issue, body)
	if err != nil {
		fmt.Fprintln(stderr, "tokencost report:", err)
		return 1
	}
	if created {
		fmt.Fprintf(stderr, "comment created on #%d\n", *issue)
	} else {
		fmt.Fprintf(stderr, "comment updated on #%d\n", *issue)
	}
	return 0
}

func runChart(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("chart", flag.ContinueOnError)
	fs.SetOutput(stderr)
	root := fs.String("root", ".", "repository root")
	costs := fs.String("costs", "", "costs file (default <root>/docs/ai-costs.csv)")
	out := fs.String("out", "", "charts file (default <root>/docs/ai-costs.md)")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *costs == "" {
		*costs = filepath.Join(*root, "docs", "ai-costs.csv")
	}
	if *out == "" {
		*out = filepath.Join(*root, "docs", "ai-costs.md")
	}
	rows, err := readRows(*costs)
	if err != nil {
		fmt.Fprintln(stderr, "tokencost chart:", err)
		return 1
	}
	if err := os.WriteFile(*out, []byte(chart(rows)), 0o644); err != nil {
		fmt.Fprintln(stderr, "tokencost chart:", err)
		return 1
	}
	fmt.Fprintln(stdout, "wrote", *out)
	return 0
}
