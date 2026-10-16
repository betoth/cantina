package linkcheck

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestCheck(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string
	}{
		{"existing file", "[a](other.md)", nil},
		{"existing directory", "[a](sub/)", nil},
		{"existing file with anchor", "[a](other.md#title)", nil},
		{"escaped path", "[a](sub/with%20space.md)", nil},
		{"root-relative path", "[a](/docs/other.md)", nil},
		{"missing file", "[a](missing.md)", []string{"missing.md"}},
		{"missing file with anchor", "[a](missing.md#title)", []string{"missing.md#title"}},
		{"missing parent", "[a](../nope/x.md)", []string{"../nope/x.md"}},
		{"image", "![a](missing.png)", []string{"missing.png"}},
		{"link with title", `[a](missing.md "title")`, []string{"missing.md"}},
		{"several on a line", "[a](other.md) and [b](missing.md)", []string{"missing.md"}},
		{"URL", "[a](https://example.com/missing.md)", nil},
		{"mailto", "[a](mailto:x@example.com)", nil},
		{"same-page anchor", "[a](#title)", nil},
		{"inline code", "`[a](missing.md)`", nil},
		{"fenced code", "```\n[a](missing.md)\n```", nil},
		{"tilde fenced code", "~~~\n[a](missing.md)\n~~~", nil},
		{"backticks inside tilde fence", "~~~\n```\n~~~\n[a](missing.md)", []string{"missing.md"}},
		{"inner shorter fence", "````\n```\n[a](missing.md)\n````", nil},
		{"inner fence with info string", "```\n```go\n[a](missing.md)\n```", nil},
		{"link after closed long fence", "````\n````\n[a](missing.md)", []string{"missing.md"}},
		{"blank line inside fence", "```\n\n[a](missing.md)\n```", nil},
		{"footnote", "[^1]: missing.md", nil},
		{"existing reference definition", "[a]: other.md", nil},
		{"missing reference definition", "[a]: missing.md", []string{"missing.md"}},
		{"reference definition with URL", "[a]: https://example.com", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			write(t, filepath.Join(root, "docs", "other.md"), "")
			write(t, filepath.Join(root, "docs", "sub", "with space.md"), "")
			write(t, filepath.Join(root, "docs", "page.md"), tt.content)

			got, err := Check(root, []string{"docs", "README.md"})
			if err != nil {
				t.Fatalf("Check: %v", err)
			}
			var links []string
			for _, b := range got {
				if b.File != "docs/page.md" {
					t.Errorf("broken link reported in %s, want docs/page.md", b.File)
				}
				links = append(links, b.Link)
			}
			if !slices.Equal(links, tt.want) {
				t.Errorf("broken links = %q, want %q", links, tt.want)
			}
		})
	}
}

func TestCheckReportsLine(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "CLAUDE.md"), "# Title\n\n[a](missing.md)\n")

	got, err := Check(root, []string{"CLAUDE.md"})
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	want := []Broken{{File: "CLAUDE.md", Line: 3, Link: "missing.md"}}
	if !slices.Equal(got, want) {
		t.Errorf("Check = %v, want %v", got, want)
	}
}

// TestRelativeLinks fails when a relative link in the repo docs points to a
// file that does not exist.
func TestRelativeLinks(t *testing.T) {
	broken, err := Check(filepath.Join("..", ".."), Targets)
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	for _, b := range broken {
		t.Errorf("broken link: %s", b)
	}
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
