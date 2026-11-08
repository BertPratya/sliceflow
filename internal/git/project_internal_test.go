package git

// These tests are in package git, not git_test, because absPaths is not
// exported.

import (
	"path/filepath"
	"runtime"
	"testing"
)

// Test 1: absPaths reads git's answer: exactly n full paths, one per
// line, as clean paths of the system. It accepts "\r\n" line ends.
func TestAbsPathsGoodAnswer(t *testing.T) {
	a, b := "/repo/.git", "/repo/.git/worktrees/x"
	if runtime.GOOS == "windows" {
		a, b = "C:/repo/.git", "C:/repo/.git/worktrees/x"
	}

	for _, out := range []string{a + "\n" + b + "\n", a + "\r\n" + b + "\r\n", a + "\n" + b} {
		got, err := absPaths(out, 2)
		if err != nil {
			t.Fatalf("absPaths(%q) returned the error %v, want none", out, err)
		}
		if got[0] != filepath.FromSlash(a) || got[1] != filepath.FromSlash(b) {
			t.Errorf("absPaths(%q) = %q, want %q and %q", out, got, filepath.FromSlash(a), filepath.FromSlash(b))
		}
	}
}

// Test 2: an answer that is not n full paths is an error. The first case
// is what a git older than 2.31 prints: it does not know
// --path-format=absolute, and prints it back with relative paths.
func TestAbsPathsBadAnswer(t *testing.T) {
	full := "/repo/.git"
	if runtime.GOOS == "windows" {
		full = "C:/repo/.git"
	}

	for _, out := range []string{
		"--path-format=absolute\n.git\n.git\n",
		".git\n.git\n",
		full + "\n",
		full + "\n" + full + "\n" + full + "\n",
		"",
	} {
		if got, err := absPaths(out, 2); err == nil {
			t.Errorf("absPaths(%q) = %q, want an error", out, got)
		}
	}
}
