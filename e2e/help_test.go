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
