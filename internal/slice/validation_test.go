package slice

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

// Test 6: a folder on the disk with a file in it takes the name, even
// when it was never committed.
func TestCheckNameFreeFolderNotCommitted(t *testing.T) {
	repo, commit := newRepoWithSlice(t)
	writeSpec(t, repo, "signup")
	wantNameCheck(t, repo, commit, "signup", ErrFolderTaken)
}

// Test 7: outside any repo, CheckNameFree returns git's error as it is:
// a *git.Error with exit code 128.
func TestCheckNameFreeOutsideARepo(t *testing.T) {
	dir := t.TempDir()
	// Stop git from looking for a repo in the folders above dir.
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(dir))

	err := CheckNameFree(context.Background(), dir, dir, "HEAD", "login")
	var gitErr *git.Error
	if !errors.As(err, &gitErr) {
		t.Fatalf("CheckNameFree returned %v, want a *git.Error", err)
	}
	if gitErr.ExitCode != 128 {
		t.Errorf("the exit code is %d, want 128", gitErr.ExitCode)
	}
}

// Test 8: an empty folder on the disk does not take the name: a new slice
// is written into it.
func TestCheckNameFreeEmptyFolder(t *testing.T) {
	repo, commit := newRepoWithSlice(t)
	makeFolder(t, filepath.Join(repo, "docs", "slices", "signup"))
	wantNameCheck(t, repo, commit, "signup", nil)
}

// Test 9: a folder that holds only a hidden file takes the name.
func TestCheckNameFreeHiddenFile(t *testing.T) {
	repo, commit := newRepoWithSlice(t)
	folder := filepath.Join(repo, "docs", "slices", "signup")
	makeFolder(t, folder)
	writeFile(t, filepath.Join(folder, ".DS_Store"))
	wantNameCheck(t, repo, commit, "signup", ErrFolderTaken)
}

// Test 10: a file where the folder would be takes the name.
func TestCheckNameFreeFileAtThePath(t *testing.T) {
	repo, commit := newRepoWithSlice(t)
	makeFolder(t, filepath.Join(repo, "docs", "slices"))
	writeFile(t, filepath.Join(repo, "docs", "slices", "signup"))
	wantNameCheck(t, repo, commit, "signup", ErrFolderTaken)
}

// Test 11: a folder deleted from the disk, but still in the base commit,
// takes the name.
func TestCheckNameFreeFolderDeletedNotCommitted(t *testing.T) {
	repo, commit := newRepoWithSlice(t)
	if err := os.RemoveAll(filepath.Join(repo, "docs", "slices", "login")); err != nil {
		t.Fatal(err)
	}
	wantNameCheck(t, repo, commit, "login", ErrFolderTaken)
}

// Test 12: the disk is checked in top, not in dir. In a second worktree
// of the repo, the main folder's files do not take the name.
func TestCheckNameFreeInASecondWorktree(t *testing.T) {
	repo, commit := newRepoWithSlice(t)
	second := filepath.Join(t.TempDir(), "second")
	runGit(t, repo, "worktree", "add", "--quiet", "-b", "second", second)
	writeSpec(t, repo, "signup")

	err := CheckNameFree(context.Background(), repo, second, commit, "signup")
	if err != nil {
		t.Errorf("CheckNameFree returned %v, want nil", err)
	}
}

// Test 13: when the folder cannot be read, CheckNameFree returns the
// system's error as it is, not "taken" or "free". It is skipped on
// Windows, where chmod cannot block reading, and as root, who can read
// anything.
func TestCheckNameFreeFolderCannotBeRead(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("cannot block reading a folder here")
	}
	repo, commit := newRepoWithSlice(t)
	folder := filepath.Join(repo, "docs", "slices", "signup")
	makeFolder(t, folder)
	if err := os.Chmod(folder, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(folder, 0o755) })

	err := CheckNameFree(context.Background(), repo, repo, commit, "signup")
	var pathErr *fs.PathError
	if !errors.As(err, &pathErr) {
		t.Errorf("CheckNameFree returned %v, want a *fs.PathError", err)
	}
}

// makeFolder makes the folder path, with the folders above it.
func makeFolder(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("could not make %s: %v", path, err)
	}
}

// writeFile writes a short file at path. The folder above it must exist.
func writeFile(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("text\n"), 0o644); err != nil {
		t.Fatalf("could not write %s: %v", path, err)
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

// wantNameCheck fails the test unless CheckNameFree gives want for name,
// with repo as both the repo and the checkout's top: nil,
// ErrFolderTaken, or ErrBranchTaken.
func wantNameCheck(t *testing.T, repo, commit, name string, want error) {
	t.Helper()
	err := CheckNameFree(context.Background(), repo, repo, commit, name)
	if !errors.Is(err, want) {
		t.Errorf("CheckNameFree(%q) returned %v, want %v", name, err, want)
	}
}
