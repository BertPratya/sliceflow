package git

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// AddWorktree makes the worktree's folder and the missing folders above
// it, and checks out the branch there.
func TestAddWorktreeMakesTheFolders(t *testing.T) {
	repo := newRepoWithBranch(t)
	path := filepath.Join(t.TempDir(), "worktrees", "shop", "login")

	if err := AddWorktree(context.Background(), repo, path, "slice/login"); err != nil {
		t.Fatalf("AddWorktree returned the error %v, want none", err)
	}
	if got := gitOutput(t, path, "branch", "--show-current"); got != "slice/login" {
		t.Errorf("the worktree has %q checked out, want slice/login", got)
	}
}

// AddWorktree uses a folder that exists and is empty.
func TestAddWorktreeInAnEmptyFolder(t *testing.T) {
	repo := newRepoWithBranch(t)
	path := filepath.Join(t.TempDir(), "login")
	if err := os.Mkdir(path, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := AddWorktree(context.Background(), repo, path, "slice/login"); err != nil {
		t.Fatalf("AddWorktree returned the error %v, want none", err)
	}
	if got := gitOutput(t, path, "branch", "--show-current"); got != "slice/login" {
		t.Errorf("the worktree has %q checked out, want slice/login", got)
	}
}

// AddWorktree returns git's error for a folder that is not empty, and
// leaves the folder's files as they were.
func TestAddWorktreeInAFolderThatIsNotEmpty(t *testing.T) {
	repo := newRepoWithBranch(t)
	path := filepath.Join(t.TempDir(), "login")
	writeFile(t, filepath.Join(path, "notes.txt"))

	err := AddWorktree(context.Background(), repo, path, "slice/login")
	wantGitError(t, err)
	if _, err := os.Stat(filepath.Join(path, "notes.txt")); err != nil {
		t.Errorf("the file in the folder is gone: %v", err)
	}
}

// RemoveWorktree deletes the worktree's folder, with a file that is not
// committed in it, and git forgets the worktree. The folder above it
// stays.
func TestRemoveWorktree(t *testing.T) {
	repo := newRepoWithBranch(t)
	parent := filepath.Join(t.TempDir(), "shop")
	path := filepath.Join(parent, "login")
	gitIn(t, repo, "worktree", "add", "--quiet", path, "slice/login")
	writeFile(t, filepath.Join(path, "draft.md"))

	if err := RemoveWorktree(context.Background(), repo, path); err != nil {
		t.Fatalf("RemoveWorktree returned the error %v, want none", err)
	}
	if _, err := os.Stat(path); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("the worktree's folder is still there, or cannot be checked: %v", err)
	}
	if _, err := os.Stat(parent); err != nil {
		t.Errorf("the folder above the worktree is gone: %v", err)
	}
	list := gitOutput(t, repo, "worktree", "list", "--porcelain")
	if n := strings.Count(list, "worktree "); n != 1 {
		t.Errorf("git lists %d worktrees, want 1, the main one:\n%s", n, list)
	}
}

// RemoveWorktree returns git's error for a folder that is not a worktree.
func TestRemoveWorktreeThatIsNotOne(t *testing.T) {
	repo := newRepoWithBranch(t)

	err := RemoveWorktree(context.Background(), repo, t.TempDir())
	wantGitError(t, err)
}

// newRepoWithBranch makes a repo with commits on main and a branch
// slice/login that is not checked out, and returns its folder.
func newRepoWithBranch(t *testing.T) string {
	t.Helper()
	repo := newRepoWithMain(t)
	gitIn(t, repo, "branch", "slice/login")
	return repo
}

// WorktreeOf finds the worktree that has the branch checked out, and
// gives "" for a branch that no worktree has.
func TestWorktreeOf(t *testing.T) {
	repo := newRepoWithBranch(t)
	gitIn(t, repo, "branch", "slice/other")
	path := filepath.Join(t.TempDir(), "login")
	gitIn(t, repo, "worktree", "add", "--quiet", path, "slice/login")

	got, err := WorktreeOf(context.Background(), repo, "slice/login")
	if err != nil {
		t.Fatalf("WorktreeOf returned the error %v, want none", err)
	}
	real, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != real {
		t.Errorf("WorktreeOf returned %q, want %q", got, real)
	}

	got, err = WorktreeOf(context.Background(), repo, "slice/other")
	if err != nil || got != "" {
		t.Errorf("WorktreeOf(slice/other) returned %q and the error %v, want \"\" and none", got, err)
	}
}
