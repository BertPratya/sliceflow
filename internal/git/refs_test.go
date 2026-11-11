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

// Test 6: CreateBranch makes the branch at the given commit, and the
// branch that is checked out stays main (R37). A second branch at an
// older commit shows that the commit is used, not the commit that is
// checked out. On a branch that exists, git refuses, and CreateBranch
// returns git's error.
func TestCreateBranch(t *testing.T) {
	r := gittest.New(t)
	first := r.Git("rev-parse", "HEAD")
	mainTip := r.Commit("b.txt", "b\n")

	if err := git.CreateBranch(context.Background(), r.Dir, "slice/login", mainTip); err != nil {
		t.Fatalf("CreateBranch returned the error %v, want none", err)
	}
	if got := r.Git("rev-parse", "refs/heads/slice/login"); got != mainTip {
		t.Errorf("slice/login is at %q, want %q", got, mainTip)
	}
	if err := git.CreateBranch(context.Background(), r.Dir, "slice/old", first); err != nil {
		t.Fatalf("CreateBranch at the first commit returned the error %v, want none", err)
	}
	if got := r.Git("rev-parse", "refs/heads/slice/old"); got != first {
		t.Errorf("slice/old is at %q, want %q", got, first)
	}
	if got := r.Git("symbolic-ref", "--short", "HEAD"); got != "main" {
		t.Errorf("the checked-out branch is %q, want main", got)
	}

	err := git.CreateBranch(context.Background(), r.Dir, "slice/login", first)
	var gitErr *git.Error
	if !errors.As(err, &gitErr) {
		t.Fatalf("CreateBranch on a branch that exists returned %v (%T), want a *git.Error", err, err)
	}
	if got := r.Git("rev-parse", "refs/heads/slice/login"); got != mainTip {
		t.Errorf("after the refusal, slice/login is at %q, want %q", got, mainTip)
	}
}

// Test 7: DeleteBranch removes the branch that CreateBranch made (R36),
// and BranchExistsShielded sees it before and not after.
func TestDeleteBranch(t *testing.T) {
	r := gittest.New(t)
	ctx := context.Background()
	if err := git.CreateBranch(ctx, r.Dir, "slice/login", r.Git("rev-parse", "HEAD")); err != nil {
		t.Fatalf("CreateBranch returned the error %v, want none", err)
	}
	if got, err := git.BranchExistsShielded(ctx, r.Dir, "slice/login"); err != nil || !got {
		t.Fatalf("before the delete, BranchExistsShielded = %v, %v, want true, no error", got, err)
	}

	if err := git.DeleteBranch(ctx, r.Dir, "slice/login"); err != nil {
		t.Fatalf("DeleteBranch returned the error %v, want none", err)
	}

	got, err := git.BranchExistsShielded(ctx, r.Dir, "slice/login")
	if err != nil {
		t.Fatalf("BranchExistsShielded returned the error %v, want none", err)
	}
	if got {
		t.Errorf("after DeleteBranch, BranchExistsShielded = true, want false")
	}
}
