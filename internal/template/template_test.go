package template

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"
)

func testTime() time.Time {
	return time.Date(2026, 10, 24, 23, 30, 0, 0, time.FixedZone("UTC-5", -5*3600))
}

// splitFront splits a document at the end of its front matter.
func splitFront(t *testing.T, doc []byte) (front, body []byte) {
	t.Helper()
	s := string(doc)
	if !strings.HasPrefix(s, "---\n") {
		t.Fatalf("no front matter start: %q", s)
	}
	i := strings.Index(s[4:], "\n---\n")
	if i < 0 {
		t.Fatalf("no front matter end")
	}
	end := 4 + i + len("\n---\n")
	return []byte(s[:end]), []byte(s[end:])
}

func TestSpecFrontMatter(t *testing.T) {
	got := string(Spec("spec-submit", testTime()))
	want := "---\nslice: spec-submit\ncreated: 2026-10-25\nideas: []\n---\n"
	if !strings.HasPrefix(got, want) {
		t.Fatalf("front matter mismatch:\n%q", got)
	}
}

func TestSpecNoPlaceholdersInFrontMatter(t *testing.T) {
	front, _ := splitFront(t, Spec("spec-submit", testTime()))
	for _, line := range strings.Split(string(front), "\n") {
		if strings.HasPrefix(line, "title:") {
			t.Errorf("title line present: %q", line)
		}
	}
	for _, p := range []string{"{{slice}}", "{{title}}", "{{date}}"} {
		if strings.Contains(string(front), p) {
			t.Errorf("placeholder %s left", p)
		}
	}
}

func TestSpecBody(t *testing.T) {
	raw, err := os.ReadFile("spec.md")
	if err != nil {
		t.Fatal(err)
	}
	_, want := splitFront(t, raw)
	_, got := splitFront(t, Spec("spec-submit", testTime()))
	if !bytes.Equal(got, want) {
		t.Fatalf("body differs from embedded template")
	}
}

func TestSpecLineEndings(t *testing.T) {
	if bytes.Contains(Spec("spec-submit", testTime()), []byte("\r")) {
		t.Fatal("output has \\r")
	}
}

func TestRaw(t *testing.T) {
	want, err := os.ReadFile("spec.md")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(Raw(), want) {
		t.Fatal("Raw differs from spec.md")
	}
}
