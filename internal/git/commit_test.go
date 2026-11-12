package git

import (
	"context"
	"path/filepath"
	"testing"
)

// Test 1: a repo with no commits has none.
func TestHasCommitsWithNoCommits(t *testing.T) {
	repo := newRepo(t)

	got, err := HasCommits(context.Background(), repo)
	if err != nil {
		t.Fatalf("HasCommits returned the error %v, want none", err)
	}
	if got {
		t.Errorf("HasCommits returned true, want false")
	}
}

// Test 2: a repo with one commit has one.
func TestHasCommitsWithACommit(t *testing.T) {
	repo := newRepoWithCommit(t)

	got, err := HasCommits(context.Background(), repo)
	if err != nil {
		t.Fatalf("HasCommits returned the error %v, want none", err)
	}
	if !got {
		t.Errorf("HasCommits returned false, want true")
	}
}

// Test 3: on an orphan branch, the current branch has no commit, but
// main does. R15 asks about the repo, so HasCommits returns true.
func TestHasCommitsOnAnOrphanBranch(t *testing.T) {
	repo := newRepoWithCommit(t)
	gitIn(t, repo, "checkout", "--quiet", "--orphan", "other")

	got, err := HasCommits(context.Background(), repo)
	if err != nil {
		t.Fatalf("HasCommits returned the error %v, want none", err)
	}
	if !got {
		t.Errorf("HasCommits returned false, want true: main has a commit")
	}
}

// Test 4: outside any repo, HasCommits returns false and git's error as
// it is: a *Error with exit code 128.
func TestHasCommitsOutsideARepo(t *testing.T) {
	dir := t.TempDir()
	// Stop git from looking for a repo in the folders above dir.
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(dir))

	got, err := HasCommits(context.Background(), dir)
	if got {
		t.Errorf("HasCommits returned true with its error, want false")
	}
	gitErr := wantGitError(t, err)
	if gitErr.ExitCode != 128 {
		t.Errorf("the exit code is %d, want 128", gitErr.ExitCode)
	}
}
