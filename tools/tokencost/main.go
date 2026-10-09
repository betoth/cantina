// Command tokencost measures the Claude Code token cost of each GitHub issue
// from local transcripts and stores the numbers in the repository.
package main

import (
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
