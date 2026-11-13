package slice

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BertPratya/sliceflow/internal/git"
)

// Tests of CheckNameFree. Each one uses a repo made by newRepoWithSlice:
// one commit with the folder docs/slices/login.

// Test 1: a name with no folder and no branch is free.
func TestCheckNameFreeWhenFree(t *testing.T) {
	repo, commit := newRepoWithSlice(t)
	wantNameCheck(t, repo, commit, "signup", nil)
}

// Test 2: a folder in the base commit takes the name (R23).
func TestCheckNameFreeFolderTaken(t *testing.T) {
	repo, commit := newRepoWithSlice(t)
	wantNameCheck(t, repo, commit, "login", ErrFolderTaken)
}

// Test 3: a local branch slice/<name> takes the name (R24).
func TestCheckNameFreeBranchTaken(t *testing.T) {
	repo, commit := newRepoWithSlice(t)
	runGit(t, repo, "branch", "slice/signup")
	wantNameCheck(t, repo, commit, "signup", ErrBranchTaken)
}

// Test 4: when both the folder and the branch exist, the folder is
// reported, because it is checked first.
func TestCheckNameFreeBothTaken(t *testing.T) {
	repo, commit := newRepoWithSlice(t)
	runGit(t, repo, "branch", "slice/login")
	wantNameCheck(t, repo, commit, "login", ErrFolderTaken)
}

// Test 5: a branch slice/<name> on a remote only does not take the name
// (R26). The remote branch is made by hand, so no remote is needed.
func TestCheckNameFreeRemoteBranch(t *testing.T) {
	repo, commit := newRepoWithSlice(t)
	runGit(t, repo, "update-ref", "refs/remotes/origin/slice/signup", "HEAD")
	wantNameCheck(t, repo, commit, "signup", nil)
}

// Test 6: a folder on the disk that is not committed does not take the
// name (R23 asks about the base commit).
func TestCheckNameFreeFolderNotCommitted(t *testing.T) {
	repo, commit := newRepoWithSlice(t)
	writeSpec(t, repo, "signup")
	wantNameCheck(t, repo, commit, "signup", nil)
}

// Test 7: outside any repo, CheckNameFree returns git's error as it is:
// a *git.Error with exit code 128.
func TestCheckNameFreeOutsideARepo(t *testing.T) {
	dir := t.TempDir()
	// Stop git from looking for a repo in the folders above dir.
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(dir))

	err := CheckNameFree(context.Background(), dir, "HEAD", "login")
	var gitErr *git.Error
	if !errors.As(err, &gitErr) {
		t.Fatalf("CheckNameFree returned %v, want a *git.Error", err)
	}
	if gitErr.ExitCode != 128 {
		t.Errorf("the exit code is %d, want 128", gitErr.ExitCode)
	}
}

// newRepoWithSlice makes a repo with one commit that has the folder
// docs/slices/login. It returns the repo's folder and the commit's ID.
func newRepoWithSlice(t *testing.T) (repo, commit string) {
	t.Helper()
	repo = t.TempDir()
	runGit(t, repo, "init", "--quiet")
	writeSpec(t, repo, "login")
	runGit(t, repo, "add", ".")
	runGit(t, repo, "commit", "--quiet", "-m", "login")
	return repo, strings.TrimSuffix(runGit(t, repo, "rev-parse", "HEAD"), "\n")
}

// writeSpec writes docs/slices/<name>/spec.md in repo, with the folders
// above it.
func writeSpec(t *testing.T, repo, name string) {
	t.Helper()
	folder := filepath.Join(repo, "docs", "slices", name)
	if err := os.MkdirAll(folder, 0o755); err != nil {
		t.Fatalf("could not make %s: %v", folder, err)
	}
	if err := os.WriteFile(filepath.Join(folder, "spec.md"), []byte("text\n"), 0o644); err != nil {
		t.Fatalf("could not write the spec of %s: %v", name, err)
	}
}

// runGit runs real git in dir to set up a test, and returns its output.
// It fails the test now if git fails. The user's own git settings for a
// name, an email, and signing are replaced, so commits always work.
func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	full := append([]string{
		"-c", "user.name=sliceflow test",
		"-c", "user.email=test@example.com",
		"-c", "commit.gpgsign=false",
	}, args...)
	cmd := exec.Command("git", full...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git %v: %v", args, err)
	}
	return string(out)
}

// wantNameCheck fails the test unless CheckNameFree gives want for name:
// nil, ErrFolderTaken, or ErrBranchTaken.
func wantNameCheck(t *testing.T, repo, commit, name string, want error) {
	t.Helper()
	err := CheckNameFree(context.Background(), repo, commit, name)
	if !errors.Is(err, want) {
		t.Errorf("CheckNameFree(%q) returned %v, want %v", name, err, want)
	}
}
