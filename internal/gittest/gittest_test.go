package gittest_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BertPratya/sliceflow/internal/gittest"
)

// Test 1: after Isolate, git reads the test's global config and not the
// user's: the name, the email, and gc.auto are the test's own, and HOME
// and USERPROFILE are the returned home folder.
func TestIsolateUsesTheTestsOwnConfig(t *testing.T) {
	home := gittest.Isolate(t)
	dir := t.TempDir()

	if got := gittest.Git(t, dir, "config", "--global", "--get", "user.email"); got != gittest.UserEmail {
		t.Errorf("user.email = %q, want %q", got, gittest.UserEmail)
	}
	if got := gittest.Git(t, dir, "config", "--global", "--get", "user.name"); got != gittest.UserName {
		t.Errorf("user.name = %q, want %q", got, gittest.UserName)
	}
	if got := gittest.Git(t, dir, "config", "--get", "gc.auto"); got != "0" {
		t.Errorf("gc.auto = %q, want 0: no git work in the background", got)
	}
	for _, key := range []string{"HOME", "USERPROFILE"} {
		if got := os.Getenv(key); got != home {
			t.Errorf("%s = %q, want %q", key, got, home)
		}
	}
	if got := os.Getenv("GIT_CONFIG_NOSYSTEM"); got != "1" {
		t.Errorf("GIT_CONFIG_NOSYSTEM = %q, want %q", got, "1")
	}
}

// Test 2: Isolate removes GIT_DIR, which a git hook sets. Left in place,
// git would use that repo instead of the test's.
func TestIsolateRemovesGitDir(t *testing.T) {
	t.Setenv("GIT_DIR", filepath.Join(t.TempDir(), "elsewhere"))

	gittest.Isolate(t)

	if _, ok := os.LookupEnv("GIT_DIR"); ok {
		t.Errorf("GIT_DIR is still set, want it removed")
	}
}

// Test 3: New makes a repo on the branch main with one commit, and Commit
// adds a commit with the file in it.
func TestNewAndCommit(t *testing.T) {
	r := gittest.New(t)

	if got := r.Git("branch", "--show-current"); got != "main" {
		t.Errorf("branch = %q, want main", got)
	}
	if got := r.Git("rev-list", "--count", "HEAD"); got != "1" {
		t.Errorf("commits = %s, want 1", got)
	}

	hash := r.Commit("docs/a/b.txt", "hello\n")

	if got := r.Git("rev-parse", "HEAD"); got != hash {
		t.Errorf("Commit returned %q, want HEAD %q", hash, got)
	}
	if got := r.Git("show", "HEAD:docs/a/b.txt"); got != "hello" {
		t.Errorf("file in HEAD = %q, want %q", got, "hello")
	}
}

// Test 4: NewEmpty makes a repo with no commits.
func TestNewEmptyHasNoCommits(t *testing.T) {
	r := gittest.NewEmpty(t)

	if got := r.Git("rev-list", "-n", "1", "--all"); got != "" {
		t.Errorf("rev-list --all = %q, want no commits", got)
	}
}
