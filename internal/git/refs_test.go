package git_test

import (
	"context"
	"errors"
	"testing"

	"github.com/BertPratya/sliceflow/internal/git"
	"github.com/BertPratya/sliceflow/internal/gittest"
)

// Test 1: BranchCommit gives the commit at the tip of the local branch,
// not the commit that is checked out (R10, R11). The branch feature is
// one commit ahead of main, and main is checked out.
func TestBranchCommitGivesTheTip(t *testing.T) {
	r := gittest.New(t)
	mainTip := r.Git("rev-parse", "HEAD")
	r.Git("checkout", "--quiet", "-b", "feature")
	featureTip := r.Commit("b.txt", "b\n")
	r.Git("checkout", "--quiet", "main")

	for branch, want := range map[string]string{"main": mainTip, "feature": featureTip} {
		got, err := git.BranchCommit(context.Background(), r.Dir, branch)
		if err != nil {
			t.Fatalf("BranchCommit(%q) returned the error %v, want none", branch, err)
		}
		if got != want {
			t.Errorf("BranchCommit(%q) = %q, want %q", branch, got, want)
		}
	}
}

// Test 2: a name that is not a local branch gives ErrNoBranch (R10, R12):
// a missing branch, a tag, a commit hash, a remote branch by its short
// name and by its own name, a form such as main~1, HEAD, and names that
// git would read as an option if they were passed as they are.
func TestBranchCommitNotALocalBranch(t *testing.T) {
	r := gittest.New(t)
	hash := r.Git("rev-parse", "HEAD")
	r.Git("tag", "v1")
	r.Git("update-ref", "refs/remotes/origin/feature", "HEAD")

	for _, name := range []string{
		"missing", "v1", hash, "origin/feature", "feature",
		"remotes/origin/feature", "main~1", "main^{commit}", "HEAD",
		"-x", "--all", "--version", "",
	} {
		got, err := git.BranchCommit(context.Background(), r.Dir, name)
		if !errors.Is(err, git.ErrNoBranch) {
			t.Errorf("BranchCommit(%q) = %q, %v, want ErrNoBranch", name, got, err)
		}
	}
}

// Test 3: BranchExists finds a local branch, also one with a / in its
// name, and only a local one: a remote branch or a tag with the same
// name does not count (R24, R26).
func TestBranchExists(t *testing.T) {
	r := gittest.New(t)
	r.Branch("slice/login")
	r.Git("update-ref", "refs/remotes/origin/slice/remote-only", "HEAD")
	r.Git("tag", "slice/tagged")

	tests := []struct {
		branch string
		want   bool
	}{
		{"main", true},
		{"slice/login", true},
		{"slice/missing", false},
		{"slice/remote-only", false},
		{"origin/slice/remote-only", false},
		{"slice/tagged", false},
		{"slice", false},
		{"-x", false},
	}
	for _, tt := range tests {
		got, err := git.BranchExists(context.Background(), r.Dir, tt.branch)
		if err != nil {
			t.Errorf("BranchExists(%q) returned the error %v, want none", tt.branch, err)
			continue
		}
		if got != tt.want {
			t.Errorf("BranchExists(%q) = %v, want %v", tt.branch, got, tt.want)
		}
	}
}

// Test 4: a branch made in a linked worktree is a branch of the project,
// seen from the main folder too.
func TestBranchExistsFromAWorktree(t *testing.T) {
	r := gittest.New(t)
	worktree := r.AddWorktree("slice/from-worktree")

	for _, dir := range []string{r.Dir, worktree} {
		got, err := git.BranchExists(context.Background(), dir, "slice/from-worktree")
		if err != nil || !got {
			t.Errorf("BranchExists in %s = %v, %v, want true, no error", dir, got, err)
		}
	}
}

// Test 5: outside a repository, git fails with a code that is not 1.
// BranchExists returns that error, not "no", and BranchCommit returns
// it, not ErrNoBranch.
func TestBranchNotARepo(t *testing.T) {
	gittest.Isolate(t)
	dir := t.TempDir()

	got, err := git.BranchExists(context.Background(), dir, "main")
	var gitErr *git.Error
	if !errors.As(err, &gitErr) {
		t.Errorf("BranchExists = %v, %v (%T), want a *git.Error", got, err, err)
	}

	_, err = git.BranchCommit(context.Background(), dir, "main")
	if !errors.As(err, &gitErr) || errors.Is(err, git.ErrNoBranch) {
		t.Errorf("BranchCommit returned the error %v (%T), want a *git.Error", err, err)
	}
}
