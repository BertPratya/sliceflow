package e2e

import (
	"strings"
	"testing"

	"github.com/BertPratya/sliceflow/internal/template"
)

// TestTemplateShow proves R35 (S34): the template, outside a git repo.
func TestTemplateShow(t *testing.T) {
	dir := t.TempDir()
	out, errOut, code := run(t, dir, "template", "show", "spec")
	if code != 0 || errOut != "" {
		t.Fatalf("code %d stderr %q", code, errOut)
	}
	if out != string(template.Raw()) {
		t.Fatalf("stdout is not the raw template: %q", out)
	}
	start := "---\nslice: {{slice}}\ncreated: {{date}}\nideas:\n  - {{an idea file this slice builds on, e.g. docs/ideas/2026-10-23-due-dates.md; none: ideas: []}}\n---\n"
	if !strings.HasPrefix(out, start) {
		t.Fatalf("stdout starts with %q", out[:min(len(out), 200)])
	}
}

// TestTemplateShowUnknown proves R35 (S35).
func TestTemplateShowUnknown(t *testing.T) {
	out, errOut, code := run(t, t.TempDir(), "template", "show", "plan")
	want := "Error: unknown template \"plan\". Known templates: spec.\n"
	if code != 1 || out != "" || errOut != want {
		t.Fatalf("code %d out %q err %q", code, out, errOut)
	}
}

// TestGroupsAlone proves R36.
func TestGroupsAlone(t *testing.T) {
	dir := t.TempDir()
	for _, g := range []string{"spec", "template"} {
		out, errOut, code := run(t, dir, g)
		if code != 0 || errOut != "" || !strings.Contains(out, "template show <template>") {
			t.Fatalf("%s: code %d out %q err %q", g, code, out, errOut)
		}
	}
	cases := []struct {
		args []string
		want string
	}{
		{[]string{"template", "show"}, "Error: missing template name. Run sliceflow --help.\n"},
		{[]string{"template", "show", "spec", "--json"}, "Error: unknown flag --json. Run sliceflow --help.\n"},
		{[]string{"template", "show", "spec", "x"}, "Error: too many arguments. Run sliceflow --help.\n"},
	}
	for _, c := range cases {
		out, errOut, code := run(t, dir, c.args...)
		if code != 2 || out != "" || errOut != c.want {
			t.Fatalf("%v: code %d out %q err %q", c.args, code, out, errOut)
		}
	}
}
