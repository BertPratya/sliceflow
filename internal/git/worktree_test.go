package git_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/BertPratya/sliceflow/internal/git"
	"github.com/BertPratya/sliceflow/internal/gittest"
)

// Test 1: AddWorktree makes the worktree on the branch (R37). In a
// folder that exists and is not empty, git refuses, and AddWorktree
// returns git's error; a folder that exists and is empty is used (R25).
func TestAddWorktree(t *testing.T) {
	r := gittest.New(t)
	ctx := context.Background()
	r.Branch("slice/login")
	r.Branch("slice/full")
	r.Branch("slice/empty")

	path := filepath.Join(t.TempDir(), "login")
	if err := git.AddWorktree(ctx, r.Dir, path, "slice/login"); err != nil {
		t.Fatalf("AddWorktree returned the error %v, want none", err)
	}
	wantOnBranch(t, path, "slice/login")

	full := mkdir(t, t.TempDir(), "full")
	if err := os.WriteFile(filepath.Join(full, "a.txt"), []byte("a\n"), 0o644); err != nil {
		t.Fatalf("write a.txt: %v", err)
	}
	err := git.AddWorktree(ctx, r.Dir, full, "slice/full")
	var gitErr *git.Error
	if !errors.As(err, &gitErr) {
		t.Errorf("AddWorktree in a folder that is not empty returned %v (%T), want a *git.Error", err, err)
	}

	empty := mkdir(t, t.TempDir(), "empty")
	if err := git.AddWorktree(ctx, r.Dir, empty, "slice/empty"); err != nil {
		t.Fatalf("AddWorktree in an empty folder returned the error %v, want none", err)
	}
	wantOnBranch(t, empty, "slice/empty")
}

// Test 2: WorktreeRegistered finds the worktree just added, also when its
// path is given through a link, or when its folder is gone, as for a
// half made worktree (R36). Another folder is not registered.
func TestWorktreeRegistered(t *testing.T) {
	r := gittest.New(t)
	ctx := context.Background()
	r.Branch("slice/login")
	parent := t.TempDir()
	path := filepath.Join(parent, "login")
	if err := git.AddWorktree(ctx, r.Dir, path, "slice/login"); err != nil {
		t.Fatalf("AddWorktree returned the error %v, want none", err)
	}

	wantRegistered(t, r.Dir, path, true)
	wantRegistered(t, r.Dir, t.TempDir(), false)
	wantRegistered(t, r.Dir, filepath.Join(parent, "other"), false)

	t.Run("through a link", func(t *testing.T) {
		link := filepath.Join(t.TempDir(), "link")
		if err := os.Symlink(parent, link); err != nil {
			t.Skipf("cannot make a link here: %v", err)
		}
		wantRegistered(t, r.Dir, filepath.Join(link, "login"), true)
	})

	t.Run("folder gone", func(t *testing.T) {
		if err := os.RemoveAll(path); err != nil {
			t.Fatalf("remove the worktree's folder: %v", err)
		}
		wantRegistered(t, r.Dir, path, true)
	})
}

// Test 3: RemoveWorktree removes a worktree with an untracked file in
// it, and a locked worktree (R36). Afterwards git does not know it, and
// its folder is gone.
func TestRemoveWorktree(t *testing.T) {
	r := gittest.New(t)
	ctx := context.Background()

	untracked := r.AddWorktree("slice/untracked")
	if err := os.WriteFile(filepath.Join(untracked, "new.txt"), []byte("new\n"), 0o644); err != nil {
		t.Fatalf("write new.txt: %v", err)
	}
	locked := r.AddWorktree("slice/locked")
	r.Git("worktree", "lock", locked)

	for _, path := range []string{untracked, locked} {
		if err := git.RemoveWorktree(ctx, r.Dir, path); err != nil {
			t.Errorf("RemoveWorktree(%s) returned the error %v, want none", path, err)
			continue
		}
		wantRegistered(t, r.Dir, path, false)
		if _, err := os.Lstat(path); !errors.Is(err, os.ErrNotExist) {
			t.Errorf("after RemoveWorktree, %s is still there (%v)", path, err)
		}
	}
}

// wantOnBranch fails the test unless the worktree at path has branch
// checked out.
func wantOnBranch(t *testing.T, path, branch string) {
	t.Helper()
	if got := gittest.Git(t, path, "symbolic-ref", "--short", "HEAD"); got != branch {
		t.Errorf("the worktree %s is on %q, want %q", path, got, branch)
	}
}

// wantRegistered fails the test unless WorktreeRegistered(path) is want,
// with no error.
func wantRegistered(t *testing.T, dir, path string, want bool) {
	t.Helper()
	got, err := git.WorktreeRegistered(context.Background(), dir, path)
	if err != nil {
		t.Fatalf("WorktreeRegistered(%s) returned the error %v, want none", path, err)
	}
	if got != want {
		t.Errorf("WorktreeRegistered(%s) = %v, want %v", path, got, want)
	}
}
