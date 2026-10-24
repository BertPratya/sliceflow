package slice_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BertPratya/sliceflow/internal/slice"
	"github.com/BertPratya/sliceflow/internal/testrepo"
)

func wantError(t *testing.T, err error, msg string) {
	t.Helper()
	se, ok := err.(*slice.Error)
	if !ok {
		t.Fatalf("want *slice.Error, got %T: %v", err, err)
	}
	if se.Message != msg {
		t.Fatalf("got %q, want %q", se.Message, msg)
	}
}

func TestFindRepoNotARepo(t *testing.T) {
	testrepo.Setup(t)
	dir := t.TempDir()
	// Stop git from looking in parent folders of the temp folder.
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(dir))
	_, err := slice.FindRepo(context.Background(), dir)
	wantError(t, err, "not inside a git repository. Nothing was changed.")
}

func TestFindRepoNoCommits(t *testing.T) {
	testrepo.Setup(t)
	dir := t.TempDir()
	testrepo.Git(t, dir, "init")
	_, err := slice.FindRepo(context.Background(), dir)
	wantError(t, err, "the repository has no commits yet. Make a first commit, then try again. Nothing was changed.")
}

func realPath(t *testing.T, p string) string {
	t.Helper()
	r, err := filepath.EvalSymlinks(filepath.FromSlash(p))
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestFindRepoFromWorktree(t *testing.T) {
	main := testrepo.New(t)
	wt := filepath.Join(t.TempDir(), "wt")
	testrepo.Git(t, main, "worktree", "add", "-b", "other", wt)
	sub := filepath.Join(wt, "a", "b")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	r, err := slice.FindRepo(context.Background(), sub)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(r.Root, `\`) {
		t.Errorf("Root has a backslash: %q", r.Root)
	}
	if realPath(t, r.Root) != realPath(t, main) {
		t.Errorf("Root = %q, want %q", r.Root, main)
	}
	if r.Dir != sub {
		t.Errorf("Dir = %q, want %q", r.Dir, sub)
	}
	if r.Now == nil {
		t.Error("Now is nil")
	}
}
