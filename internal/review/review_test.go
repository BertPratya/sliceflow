package review

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

const wantFront = "---\n" +
	"review: spec-1\n" +
	"slice: login-google\n" +
	"kind: spec\n" +
	"target: spec.md\n" +
	"round: 1\n" +
	"reviewer: placeholder\n" +
	"---\n"

const wantBody = "# Review\n\n" +
	"## Summary\n\nNo review was done: placeholder reviewer.\n\n" +
	"## Checked\n\nNone.\n\n" +
	"## Tests\n\nNone.\n\n" +
	"## Previous findings\n\nNone.\n\n" +
	"## Findings\n\nNone.\n\n" +
	"## Requirements check\n\nNone.\n\n" +
	"## Needs human decision\n\nNone.\n"

func TestFrontMatter(t *testing.T) {
	got := string(FrontMatter("login-google", 1, "placeholder"))
	if got != wantFront {
		t.Fatalf("got %q, want %q", got, wantFront)
	}
}

func TestPlaceholder(t *testing.T) {
	r := Placeholder()
	if r.Name() != "placeholder" {
		t.Fatalf("name %q", r.Name())
	}
	file := filepath.Join(t.TempDir(), "spec-1.md")
	if err := os.WriteFile(file, FrontMatter("login-google", 1, r.Name()), 0o644); err != nil {
		t.Fatal(err)
	}
	req := Request{Slice: "login-google", Doc: "spec.md", File: file, Round: 1}
	if err := r.Review(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != wantFront+wantBody {
		t.Fatalf("got %q", got)
	}
}

func TestReadResult(t *testing.T) {
	cases := []struct{ name, in, want string }{
		{"placeholder", wantFront + wantBody, "passed"},
		{"crlf", "# Review\r\n\r\n## Needs human decision\r\n\r\nNone.\r\n", "passed"},
		{"question", "# Review\n\n## Needs human decision\n\n- **Q1.** Which one?\n", "needs_human"},
		{"question then section", "## Needs human decision\n\n- **Q1.** x\n\n## Other\n\nNone.\n", "needs_human"},
		{"none then section", "## Needs human decision\n\nNone.\n\n## Other\n\nx\n", "passed"},
		{"missing", "# Review\n\n## Summary\n\nx\n", "needs_human"},
		{"empty section", "## Needs human decision\n", "needs_human"},
	}
	for _, c := range cases {
		if got := ReadResult([]byte(c.in)); got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}
