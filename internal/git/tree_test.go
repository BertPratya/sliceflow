package git_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/BertPratya/sliceflow/internal/git"
	"github.com/BertPratya/sliceflow/internal/gittest"
)

// Test 1: FolderInCommit finds a folder in the commit, and says no for a
// file with that name, a missing folder, and a folder whose name only
// starts the same way (R23).
func TestFolderInCommit(t *testing.T) {
	r := gittest.New(t)
	r.WriteFile("docs/slices/login/spec.md", "spec\n")
	r.WriteFile("docs/slices/card", "a file, not a folder\n")
	r.WriteFile("docs/slices/payment-v2/spec.md", "spec\n")
	r.Git("add", "--all")
	r.Git("commit", "--quiet", "-m", "add slices")
	commit := r.Git("rev-parse", "HEAD")

	tests := []struct {
		path string
		want bool
	}{
		{"docs/slices/login", true},
		{"docs/slices", true},
		{"docs/slices/card", false},
		{"docs/slices/missing", false},
		{"docs/slices/payment", false},
		{"docs/slices/log", false},
	}
	for _, tt := range tests {
		got, err := git.FolderInCommit(context.Background(), r.Dir, commit, tt.path)
		if err != nil {
			t.Errorf("FolderInCommit(%q) returned the error %v, want none", tt.path, err)
			continue
		}
		if got != tt.want {
			t.Errorf("FolderInCommit(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

// Test 2: only the given commit counts (R23). The folder is in a later
// commit, and in the working folder, but not in the first commit. It is
// also not in a commit when it is only in the working folder.
func TestFolderInCommitOnlyThatCommit(t *testing.T) {
	r := gittest.New(t)
	first := r.Git("rev-parse", "HEAD")
	second := r.Commit("docs/slices/login/spec.md", "spec\n")
	r.WriteFile("docs/slices/draft/spec.md", "not committed\n")

	check := func(commit, path string, want bool) {
		t.Helper()
		got, err := git.FolderInCommit(context.Background(), r.Dir, commit, path)
		if err != nil {
			t.Fatalf("FolderInCommit(%s, %q) returned the error %v, want none", commit, path, err)
		}
		if got != want {
			t.Errorf("FolderInCommit(%s, %q) = %v, want %v", commit, path, got, want)
		}
	}
	check(first, "docs/slices/login", false)
	check(second, "docs/slices/login", true)
	check(second, "docs/slices/draft", false)
}

// Test 3: from a subfolder of the project, the path still starts at the
// project's top folder.
func TestFolderInCommitFromASubfolder(t *testing.T) {
	r := gittest.New(t)
	commit := r.Commit("docs/slices/login/spec.md", "spec\n")
	sub := filepath.Join(r.Dir, "docs")

	got, err := git.FolderInCommit(context.Background(), sub, commit, "docs/slices/login")
	if err != nil || !got {
		t.Errorf("FolderInCommit from docs/ = %v, %v, want true, no error", got, err)
	}
}

// Test 4: a commit that does not exist, or that looks like an option,
// gives git's error, not "no folder". The -x and --help cases would fail
// without --end-of-options too; in sliceflow the commit always comes
// from BranchCommit, so it is always a hash.
func TestFolderInCommitBadCommit(t *testing.T) {
	r := gittest.New(t)

	for _, commit := range []string{"0123456789012345678901234567890123456789", "-x", "--help"} {
		got, err := git.FolderInCommit(context.Background(), r.Dir, commit, "docs")
		var gitErr *git.Error
		if !errors.As(err, &gitErr) {
			t.Errorf("FolderInCommit(%q) = %v, %v (%T), want a *git.Error", commit, got, err, err)
		}
	}
}
