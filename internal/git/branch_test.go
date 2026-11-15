package git

import (
	"context"
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Tests of BranchExists. Each one asks about a name in a repo made by
// newRepoWithMain: two commits on main, and a tag v1.

// Test 1: a local branch that exists.
func TestBranchExistsForALocalBranch(t *testing.T) {
	repo := newRepoWithMain(t)
	wantBranchExists(t, repo, "main", true)
}

// Test 2: a name with no branch.
func TestBranchExistsForNoSuchBranch(t *testing.T) {
	repo := newRepoWithMain(t)
	wantBranchExists(t, repo, "nosuch", false)
}

// Test 3: a tag is not a local branch (R10).
func TestBranchExistsForATag(t *testing.T) {
	repo := newRepoWithMain(t)
	wantBranchExists(t, repo, "v1", false)
}

// Test 4: main~1 is revision syntax, not a branch name (R10). Git's
// rev-parse would accept refs/heads/main~1 as the commit before main;
// main has two commits, so that commit exists.
func TestBranchExistsForRevisionSyntax(t *testing.T) {
	repo := newRepoWithMain(t)
	wantBranchExists(t, repo, "main~1", false)
}

// Test 5: outside any repo, BranchExists returns false and git's error
// as it is: a *Error with exit code 128.
func TestBranchExistsOutsideARepo(t *testing.T) {
	dir := t.TempDir()
	// Stop git from looking for a repo in the folders above dir.
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(dir))

	got, err := BranchExists(context.Background(), dir, "main")
	if got {
		t.Errorf("BranchExists returned true with its error, want false")
	}
	gitErr := wantGitError(t, err)
	if gitErr.ExitCode != 128 {
		t.Errorf("the exit code is %d, want 128", gitErr.ExitCode)
	}
}

// Test 5b: a branch that exists only on a remote is not a local branch
// (R26). The remote branch is made by hand, so no remote is needed.
func TestBranchExistsForARemoteBranch(t *testing.T) {
	repo := newRepoWithMain(t)
	gitIn(t, repo, "update-ref", "refs/remotes/origin/slice/login", "HEAD")
	wantBranchExists(t, repo, "slice/login", false)
}

// Tests of BranchCommit.

// Test 6: for a local branch, BranchCommit returns the commit at its
// tip, with no line break (R10, R11).
func TestBranchCommitForALocalBranch(t *testing.T) {
	repo := newRepoWithMain(t)
	want := gitOutput(t, repo, "rev-parse", "refs/heads/main")

	got, err := BranchCommit(context.Background(), repo, "main")
	if err != nil {
		t.Fatalf("BranchCommit returned the error %v, want none", err)
	}
	if got != want {
		t.Errorf("BranchCommit returned %q, want %q", got, want)
	}
}

// Test 7: for a name with no branch, BranchCommit returns ErrNoBranch
// (R12).
func TestBranchCommitForNoSuchBranch(t *testing.T) {
	repo := newRepoWithMain(t)
	wantNoBranch(t, repo, "nosuch")
}

// Test 8: revision syntax is not a branch, so BranchCommit returns
// ErrNoBranch, not the commit before main (R10, R12).
func TestBranchCommitForRevisionSyntax(t *testing.T) {
	repo := newRepoWithMain(t)
	wantNoBranch(t, repo, "main~1")
}

// newRepoWithMain makes a repo with two commits on a branch named main,
// and a tag v1, and returns its folder. The branch is renamed to main
// because git's default branch name depends on the user's settings.
func newRepoWithMain(t *testing.T) string {
	t.Helper()
	repo := newRepoWithCommit(t)
	gitIn(t, repo, "branch", "-M", "main")
	gitIn(t, repo, "commit", "--quiet", "--allow-empty", "-m", "second")
	gitIn(t, repo, "tag", "v1")
	return repo
}

// gitOutput runs real git in dir and returns its output with no line
// break at the end. It fails the test now if git fails.
func gitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git %v: %v", args, err)
	}
	return strings.TrimSuffix(string(out), "\n")
}

// wantBranchExists fails the test unless BranchExists gives want for
// branch, with no error.
func wantBranchExists(t *testing.T, repo, branch string, want bool) {
	t.Helper()
	got, err := BranchExists(context.Background(), repo, branch)
	if err != nil {
		t.Fatalf("BranchExists(%q) returned the error %v, want none", branch, err)
	}
	if got != want {
		t.Errorf("BranchExists(%q) returned %v, want %v", branch, got, want)
	}
}

// wantNoBranch fails the test unless BranchCommit returns "" and
// ErrNoBranch for branch.
func wantNoBranch(t *testing.T, repo, branch string) {
	t.Helper()
	got, err := BranchCommit(context.Background(), repo, branch)
	if !errors.Is(err, ErrNoBranch) {
		t.Errorf("BranchCommit(%q) returned the error %v, want ErrNoBranch", branch, err)
	}
	if got != "" {
		t.Errorf("BranchCommit(%q) returned %q with its error, want \"\"", branch, got)
	}
}

// CreateBranch makes the branch at the given commit, which need not be
// the tip of the current branch.
func TestCreateBranchAtACommit(t *testing.T) {
	repo := newRepoWithMain(t)
	commit := gitOutput(t, repo, "rev-parse", "main~1")

	if err := CreateBranch(context.Background(), repo, "slice/login", commit); err != nil {
		t.Fatalf("CreateBranch returned the error %v, want none", err)
	}
	if got := gitOutput(t, repo, "rev-parse", "refs/heads/slice/login"); got != commit {
		t.Errorf("slice/login is at %s, want %s", got, commit)
	}
	if got := gitOutput(t, repo, "branch", "--show-current"); got != "main" {
		t.Errorf("the checked-out branch is %q, want main", got)
	}
}

// CreateBranch returns git's error for a branch that exists.
func TestCreateBranchThatExists(t *testing.T) {
	repo := newRepoWithMain(t)
	commit := gitOutput(t, repo, "rev-parse", "main")

	err := CreateBranch(context.Background(), repo, "main", commit)
	wantGitError(t, err)
}

// DeleteBranch deletes a branch with commits that are on no other
// branch.
func TestDeleteBranch(t *testing.T) {
	repo := newRepoWithMain(t)
	gitIn(t, repo, "switch", "--quiet", "-c", "slice/login")
	gitIn(t, repo, "commit", "--quiet", "--allow-empty", "-m", "only here")
	gitIn(t, repo, "switch", "--quiet", "main")

	if err := DeleteBranch(context.Background(), repo, "slice/login"); err != nil {
		t.Fatalf("DeleteBranch returned the error %v, want none", err)
	}
	wantBranchExists(t, repo, "slice/login", false)
}

// DeleteBranch returns git's error for a branch that does not exist.
func TestDeleteBranchThatDoesNotExist(t *testing.T) {
	repo := newRepoWithMain(t)

	err := DeleteBranch(context.Background(), repo, "nosuch")
	wantGitError(t, err)
}

// DeleteBranch returns git's error for a branch that is checked out in a
// worktree, and the branch stays.
func TestDeleteBranchCheckedOutInAWorktree(t *testing.T) {
	repo := newRepoWithMain(t)
	gitIn(t, repo, "branch", "slice/login")
	gitIn(t, repo, "worktree", "add", "--quiet", filepath.Join(t.TempDir(), "login"), "slice/login")

	err := DeleteBranch(context.Background(), repo, "slice/login")
	wantGitError(t, err)
	wantBranchExists(t, repo, "slice/login", true)
}
