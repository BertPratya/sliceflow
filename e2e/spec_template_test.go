package e2e

import (
	"strings"
	"testing"

	"github.com/BertPratya/sliceflow/templates"
)

// spec template prints the built-in template exactly as it is stored,
// with its placeholders, in a folder that is not inside a git repository.
func TestSpecTemplate(t *testing.T) {
	got := runSliceflow(t, t.TempDir(), newHome(t), "spec", "template")

	wantOutput(t, got, templates.Spec, "")
	wantStart := "---\nslice: {{slice}}\ncreated: {{date}}\n---\n"
	if !strings.HasPrefix(got.stdout, wantStart) {
		t.Errorf("stdout starts with %q, want %q", got.stdout[:min(len(got.stdout), len(wantStart))], wantStart)
	}
}

// spec template --help prints its help, and an argument is refused as an
// unexpected argument, with exit code 2.
func TestSpecTemplateHelp(t *testing.T) {
	home := newHome(t)

	got := runSliceflow(t, t.TempDir(), home, "spec", "template", "--help")
	want := "Print the spec template built into sliceflow, with its guidance comments.\n" +
		"\n" +
		"Usage:\n" +
		"  sliceflow spec template [options]\n" +
		"\n" +
		"Options:\n" +
		"  -h, --help  Show this help.\n"
	wantOutput(t, got, want, "")

	got = runSliceflow(t, t.TempDir(), home, "spec", "template", "login")
	wantError(t, got, 2, "Error: unexpected argument \"login\". Nothing was changed.\n"+
		"Usage: sliceflow spec template [options]\n")
}
