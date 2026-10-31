package e2e

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/BertPratya/sliceflow/internal/gittest"
)

// TestSpecTemplateS14: 0002 S14. "spec template" prints the spec template
// with {{slice}} and {{date}} left as they are, also outside a git repo
// (0002 R28).
func TestSpecTemplateS14(t *testing.T) {
	r := run(t, t.TempDir(), "spec", "template")
	if r.code != 0 || r.stderr != "" {
		t.Fatalf("code = %d, stderr = %q, want 0, empty", r.code, r.stderr)
	}
	want := "---\nslice: {{slice}}\ncreated: {{date}}\n---\n\n<!-- HOW TO FILL THIS TEMPLATE\n"
	if !strings.HasPrefix(r.stdout, want) {
		t.Errorf("stdout starts with %q, want %q", r.stdout[:min(len(r.stdout), len(want))], want)
	}
	if strings.Contains(r.stdout, "title:") || strings.Contains(r.stdout, "{{title}}") {
		t.Error("stdout has a title line, want none")
	}
}

// TestSpecTemplateSameAsCreate: the template is the text "slice create"
// writes, before {{slice}} and {{date}} are filled (0002 R28).
func TestSpecTemplateSameAsCreate(t *testing.T) {
	repo := gittest.NewRepo(t)
	worktree := createSlice(t, repo, "spec-submit")
	written := readFile(t, filepath.ToSlash(filepath.Join(worktree, "docs", "slices", "spec-submit", "spec.md")))

	r := run(t, repo, "spec", "template")
	if r.code != 0 {
		t.Fatalf("code = %d, stderr = %q", r.code, r.stderr)
	}
	date := strings.TrimPrefix(strings.Split(written, "\n")[2], "created: ")
	filled := strings.Replace(strings.Replace(r.stdout, "{{slice}}", "spec-submit", 1), "{{date}}", date, 1)
	if filled != written {
		t.Errorf("the template with slice and date filled is not the spec of slice create")
	}
}

// TestSpecTemplateWrongUse: "spec template" takes no names and no flags
// other than -h.
func TestSpecTemplateWrongUse(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"a name", []string{"spec", "template", "x"}, "too many arguments"},
		{"a flag", []string{"spec", "template", "--json"}, "unknown flag --json"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := run(t, t.TempDir(), tt.args...)
			want := "Error: " + tt.want + ". Run sliceflow --help.\n"
			if r.code != 2 || r.stdout != "" || r.stderr != want {
				t.Errorf("code = %d, stdout = %q, stderr = %q, want 2, empty, %q", r.code, r.stdout, r.stderr, want)
			}
		})
	}
}

// TestSpecTemplateHelp: -h prints the usage of "spec template".
func TestSpecTemplateHelp(t *testing.T) {
	r := run(t, t.TempDir(), "spec", "template", "-h")
	if r.code != 0 || r.stderr != "" || !strings.Contains(r.stdout, "Usage:\n  sliceflow spec template [flags]") {
		t.Errorf("code = %d, stderr = %q, stdout = %q", r.code, r.stderr, r.stdout)
	}
}
