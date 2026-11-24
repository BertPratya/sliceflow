package e2e

import (
	"strings"
	"testing"
)

// The full help texts are checked by the tests of internal/cli. These
// tests check that the real program passes on the right words, prints
// to the right stream, and exits with the right code.

// sliceflow --help prints the guide to stdout, nothing to stderr, and
// exits with 0.
func TestHelpGoesToStdout(t *testing.T) {
	got := runSliceflow(t, t.TempDir(), newHome(t), "--help")

	wantStart := "sliceflow runs a spec-driven workflow with AI coding agents.\n"
	if !strings.HasPrefix(got.stdout, wantStart) {
		t.Errorf("stdout is %q, want it to start with %q", got.stdout, wantStart)
	}
	if got.stderr != "" {
		t.Errorf("stderr is %q, want nothing", got.stderr)
	}
	if got.code != 0 {
		t.Errorf("exit code is %d, want 0", got.code)
	}
}

// A wrong command line prints its two lines to stderr, nothing to
// stdout, and exits with 2.
func TestWrongUseGoesToStderr(t *testing.T) {
	got := runSliceflow(t, t.TempDir(), newHome(t), "nope")

	wantStderr := "Error: unknown command \"nope\". Nothing was changed.\n" +
		"To see the commands, run: sliceflow --help\n"
	if got.stderr != wantStderr {
		t.Errorf("stderr is %q, want %q", got.stderr, wantStderr)
	}
	if got.stdout != "" {
		t.Errorf("stdout is %q, want nothing", got.stdout)
	}
	if got.code != 2 {
		t.Errorf("exit code is %d, want 2", got.code)
	}
}

// The help of slice create and slice status says what they do while a
// slice is planned: create makes it on the base branch, and --from
// chooses that branch.
func TestSliceHelp(t *testing.T) {
	home := newHome(t)

	create := runSliceflow(t, t.TempDir(), home, "slice", "create", "--help")
	wantCode(t, create, 0)
	wantFirst := "Make a new slice on the base branch: a spec and an event log, in one commit.\n"
	if !strings.HasPrefix(create.stdout, wantFirst) {
		t.Errorf("stdout is:\n%s\nwant it to start with:\n%s", create.stdout, wantFirst)
	}
	wantFrom := "The base branch. Default: the sliceflow.base setting, else main."
	found := false
	for _, line := range strings.Split(create.stdout, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "--from ") && strings.HasSuffix(line, wantFrom) {
			found = true
		}
	}
	if !found {
		t.Errorf("stdout is:\n%s\nwant a line for --from that ends with %q", create.stdout, wantFrom)
	}

	status := runSliceflow(t, t.TempDir(), home, "slice", "status", "--help")
	wantCode(t, status, 0)
	wantFirst = "Show one slice: its state, the next step, its title, description, creation time, base, and events.\n"
	if !strings.HasPrefix(status.stdout, wantFirst) {
		t.Errorf("stdout is:\n%s\nwant it to start with:\n%s", status.stdout, wantFirst)
	}
}
