package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// goodSpec is a spec with no problems.
const goodSpec = `---
slice: login
created: 2026-11-18
---

# Spec

## Story 1: Log in

A user logs in.

Uses: none

### Rule ` + "`log-in`" + `: a user logs in

When a user logs in, the shop shows their name.
Check: e2e

- **Example:** **When** I log in,
  **Then** the shop shows my name.

## Open questions

None
`

// writeSpec writes text as the spec of the slice name in the checkout at
// repo, without committing it.
func writeSpec(t *testing.T, repo, name, text string) {
	t.Helper()
	writeText(t, filepath.Join(repo, "docs", "slices", name, "spec.md"), text)
}

// spec validate reads the spec as it is now, changes not committed
// included, and says it has no problems.
func TestSpecValidateNoProblems(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)
	createSlice(t, repo, home, "login")
	writeSpec(t, repo, "login", goodSpec)

	got := runSliceflow(t, repo, home, "spec", "validate", "login")

	wantOutput(t, got, specFile(t, repo, home, "login")+": no problems found.\n", "")
}

// spec validate lists each problem with its path, line, and hint, in
// line order, then counts them, and exits with 1.
func TestSpecValidateListsProblems(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)
	createSlice(t, repo, home, "login")
	// Line 12 is the shared rule unknown-event, which no story uses, and
	// line 40 is the rule edit-empty-values, which has no Check: line.
	lines := []string{"---", "slice: login", "created: 2026-11-18", "---", "", "# Spec", "", "## Shared rules", "", "Text.", ""}
	lines = append(lines, "### Rule `unknown-event`: an unknown event", "", "Text.", "Check: unit", "", "## Story 1: Edit", "", "Uses: none", "")
	for len(lines) < 39 {
		lines = append(lines, "Text.")
	}
	lines = append(lines, "### Rule `edit-empty-values`: empty values", "", "Text.", "",
		"- **Example:** **When** I edit,", "  **Then** it is empty.")
	writeSpec(t, repo, "login", strings.Join(lines, "\n")+"\n")

	got := runSliceflow(t, repo, home, "spec", "validate", "login")

	path := specFile(t, repo, home, "login")
	want := path + ":12: error: shared rule unknown-event is not used by any story.\n" +
		"  Name it in a story's Uses: line, or delete it.\n" +
		path + ":40: error: rule edit-empty-values has no Check: line.\n" +
		"  Add a line like \"Check: e2e\" under the rule's heading.\n" +
		"2 problems found.\n"
	wantCode(t, got, 1)
	if got.stdout != want {
		t.Errorf("stdout is:\n%s\nwant:\n%s", got.stdout, want)
	}
	if got.stderr != "" {
		t.Errorf("stderr is %q, want nothing", got.stderr)
	}

	t.Run("one problem", func(t *testing.T) {
		writeSpec(t, repo, "login", strings.Replace(strings.Join(lines, "\n")+"\n", "Uses: none", "Uses: unknown-event", 1))
		got := runSliceflow(t, repo, home, "spec", "validate", "login")
		wantCode(t, got, 1)
		if !strings.HasSuffix(got.stdout, "\n1 problem found.\n") {
			t.Errorf("stdout is:\n%s\nwant it to end with 1 problem found.", got.stdout)
		}
	})
}

// spec validate reports a wrong name, a missing name, and a slice with
// no log exactly as slice status does.
func TestSpecValidateErrorsAsBefore(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)

	got := runSliceflow(t, repo, home, "spec", "validate", "a.b")
	wantError(t, got, 2, "Error: invalid slice name \"a.b\". Nothing was changed.\n"+
		"Allowed characters: letters, digits, hyphens, underscores. Start with a letter or a digit.\n")

	got = runSliceflow(t, repo, home, "spec", "validate")
	wantError(t, got, 2, "Error: missing slice name. Nothing was changed.\n"+
		"Usage: sliceflow spec validate <name> [options]\n")

	status := runSliceflow(t, repo, home, "slice", "status", "login")
	got = runSliceflow(t, repo, home, "spec", "validate", "login")
	wantError(t, got, 1, status.stderr)
	if status.code != 1 || !strings.HasPrefix(status.stderr, "Error: no slice named login.") {
		t.Errorf("slice status gave %d and %q, want the no-slice error", status.code, status.stderr)
	}
}

// With a log and no spec, spec validate says where it looked, and how to
// start from the template.
func TestSpecValidateNeedsTheSpec(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)
	createSlice(t, repo, home, "login")
	if err := os.Remove(filepath.Join(repo, "docs", "slices", "login", "spec.md")); err != nil {
		t.Fatal(err)
	}

	got := runSliceflow(t, repo, home, "spec", "validate", "login")

	wantError(t, got, 1, "Error: the spec of slice login was not found at "+specFile(t, repo, home, "login")+". Nothing was changed.\n"+
		"To start from the template, run: sliceflow spec template\n")
}

// spec validate does not read the log, so a log with an unknown event, or
// a damaged first line, gives no warning.
func TestSpecValidateInAnyState(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)
	createSlice(t, repo, home, "login")
	writeSpec(t, repo, "login", goodSpec)
	appendText(t, sliceLog(repo, "login"), eventLine(2, "spec.submitted", `{}`))

	got := runSliceflow(t, repo, home, "spec", "validate", "login")
	wantOutput(t, got, specFile(t, repo, home, "login")+": no problems found.\n", "")

	writeText(t, sliceLog(repo, "login"), "cut off\n")
	got = runSliceflow(t, repo, home, "spec", "validate", "login")
	wantOutput(t, got, specFile(t, repo, home, "login")+": no problems found.\n", "")
}

// spec validate --help and spec --help print their help.
func TestSpecValidateHelp(t *testing.T) {
	home := newHome(t)

	got := runSliceflow(t, t.TempDir(), home, "spec", "validate", "--help")
	wantOutput(t, got, "Find the problems in a slice's spec, each with its line and a hint on how to fix it.\n"+
		"\n"+
		"Usage:\n"+
		"  sliceflow spec validate <name> [options]\n"+
		"\n"+
		"Arguments:\n"+
		"  <name>  The name of the slice whose spec to read.\n"+
		"\n"+
		"Options:\n"+
		"  -h, --help  Show this help.\n", "")

	got = runSliceflow(t, t.TempDir(), home, "spec", "--help")
	wantOutput(t, got, "Print the spec template, and find problems in specs.\n"+
		"\n"+
		"Usage:\n"+
		"  sliceflow spec <command> [options]\n"+
		"\n"+
		"Commands:\n"+
		"  template         Print the spec template\n"+
		"  validate <name>  Find the problems in a spec\n"+
		"\n"+
		"Options:\n"+
		"  -h, --help  Show this help.\n", "")

	got = runSliceflow(t, t.TempDir(), home, "--help")
	if !strings.Contains(got.stdout, "\n  spec   Print the spec template, and find problems in specs\n") {
		t.Errorf("the guide is:\n%s\nwant a line for spec", got.stdout)
	}
}
