// Package linkcheck finds relative Markdown links whose target file does not exist.
//
// It checks inline links and images, and reference definitions ([a]: path),
// skipping fenced code (``` or ~~~) and single-backtick inline code. A fence
// closes only on a line with the same character, at least as long, and nothing
// after it. Footnote definitions ([^1]: text) are not links. Not covered: link targets containing ')', angle-bracket targets (<path>) and
// inline code with more than one backtick.
package linkcheck

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Targets are the repo paths whose Markdown files are checked: directories are
// walked recursively, files are checked when present.
var Targets = []string{"CLAUDE.md", "README.md", "README.pt-BR.md", "docs", ".claude"}

// Broken is a relative link whose target does not exist.
type Broken struct {
	File string // repo-relative path of the Markdown file
	Line int
	Link string
}

func (b Broken) String() string { return fmt.Sprintf("%s:%d: %s", b.File, b.Line, b.Link) }

var (
	linkRe       = regexp.MustCompile(`\]\(([^)\s]+)(?:\s+"[^"]*")?\)`)
	refDefRe     = regexp.MustCompile(`^ {0,3}\[[^\]^][^\]]*\]:\s*(\S+)`)
	inlineCodeRe = regexp.MustCompile("`[^`]*`")
	schemeRe     = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.-]*:`)
)

// Check walks targets under root and returns every broken relative link.
func Check(root string, targets []string) ([]Broken, error) {
	var broken []Broken
	for _, target := range targets {
		err := filepath.WalkDir(filepath.Join(root, target), func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				if errors.Is(err, fs.ErrNotExist) && path == filepath.Join(root, target) {
					return nil
				}
				return err
			}
			if d.IsDir() || filepath.Ext(path) != ".md" {
				return nil
			}
			found, err := checkFile(root, path)
			broken = append(broken, found...)
			return err
		})
		if err != nil {
			return nil, fmt.Errorf("walk %s: %w", target, err)
		}
	}
	return broken, nil
}

func checkFile(root, path string) ([]Broken, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	rel, err := filepath.Rel(root, path)
	if err != nil {
		return nil, fmt.Errorf("relative path of %s: %w", path, err)
	}
	var broken []Broken
	fence := "" // opening run of the code fence (``` or ~~~, 3 or more), empty outside one
	scanner := bufio.NewScanner(f)
	for n := 1; scanner.Scan(); n++ {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if fence == "" {
			if fence = fenceRun(trimmed); fence != "" {
				continue
			}
		} else {
			// A closing fence repeats the opening character at least as many
			// times, with nothing after it.
			if run := fenceRun(trimmed); run != "" && run == trimmed && run[0] == fence[0] && len(run) >= len(fence) {
				fence = ""
			}
			continue
		}
		line = inlineCodeRe.ReplaceAllString(line, "")
		matches := linkRe.FindAllStringSubmatch(line, -1)
		if m := refDefRe.FindStringSubmatch(line); m != nil {
			matches = append(matches, m)
		}
		for _, m := range matches {
			target, ok := localTarget(root, filepath.Dir(path), m[1])
			if !ok {
				continue
			}
			if _, err := os.Stat(target); err != nil {
				broken = append(broken, Broken{File: filepath.ToSlash(rel), Line: n, Link: m[1]})
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return broken, nil
}

// fenceRun returns the leading run of ` or ~ when it is long enough to open a
// code fence, or "" otherwise.
func fenceRun(line string) string {
	if line == "" || (line[0] != '`' && line[0] != '~') {
		return ""
	}
	n := 1
	for n < len(line) && line[n] == line[0] {
		n++
	}
	if n < 3 {
		return ""
	}
	return line[:n]
}

// localTarget resolves a link to a file path, or reports false for links that
// are not local files (URLs, mailto, same-page anchors).
func localTarget(root, dir, link string) (string, bool) {
	if schemeRe.MatchString(link) || strings.HasPrefix(link, "#") {
		return "", false
	}
	link, _, _ = strings.Cut(link, "#")
	if unescaped, err := url.PathUnescape(link); err == nil {
		link = unescaped
	}
	if strings.HasPrefix(link, "/") {
		return filepath.Join(root, link), true
	}
	return filepath.Join(dir, link), true
}
