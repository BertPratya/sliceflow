// Package gittest makes temporary git repositories for tests. It turns off
// the user's own git settings, so a test works the same on every computer.
package gittest

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Init makes a new repository on the branch main, with no commits, and
// returns its folder. It is removed when the test ends.
func Init(t *testing.T) string {
	t.Helper()
	isolate(t)

	dir := t.TempDir()
	Run(t, dir, "init", "-b", "main")
	Run(t, dir, "config", "user.name", "sliceflow test")
	Run(t, dir, "config", "user.email", "test@example.com")
	Run(t, dir, "config", "commit.gpgsign", "false")
	return dir
}

// NewRepo makes a new repository on the branch main, with one empty commit,
// and returns its folder.
func NewRepo(t *testing.T) string {
	t.Helper()
	dir := Init(t)
	Commit(t, dir, "first")
	return dir
}

// Commit makes a commit with no files in the repository at dir.
func Commit(t *testing.T, dir, msg string) {
	t.Helper()
	Run(t, dir, "commit", "--allow-empty", "-m", msg)
}

// Run runs git with args in dir, and returns what git printed, without the
// new line at the end. It stops the test if git fails.
func Run(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "LC_ALL=C")

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, stderr.String())
	}
	return strings.TrimSpace(string(out))
}

// isolate makes git ignore the user's global and system settings until the
// test ends, e.g. a global sliceflow.base, hooks, or commit signing.
func isolate(t *testing.T) {
	t.Helper()
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(t.TempDir(), "gitconfig"))
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
}
