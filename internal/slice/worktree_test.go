package slice

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/BertPratya/sliceflow/internal/testrepo"
)

func TestFindWorktree(t *testing.T) {
	ctx := context.Background()
	r := newRepo(t)
	c, err := Create(ctx, r, CreateOptions{Name: "login-google"})
	if err != nil {
		t.Fatal(err)
	}
	got, ok, err := findWorktree(ctx, r, "login-google")
	if err != nil || !ok || !sameFile(t, got, c.Worktree) {
		t.Errorf("got %q %v %v, want %q", got, ok, err, c.Worktree)
	}
	_, ok, err = findWorktree(ctx, r, "other")
	if err != nil || ok {
		t.Errorf("other: %v %v", ok, err)
	}
	if filepath.ToSlash(got) != filepath.ToSlash(filepath.Clean(got)) || got != filepath.Clean(got) {
		t.Errorf("not clean: %q", got)
	}
	testrepo.Git(t, r.Dir, "worktree", "remove", "--force", c.Worktree)
	_, ok, err = findWorktree(ctx, r, "login-google")
	if err != nil || ok {
		t.Errorf("removed: %v %v", ok, err)
	}
}

func TestFindWorktreeDeletedByHand(t *testing.T) {
	ctx := context.Background()
	r := newRepo(t)
	c, err := Create(ctx, r, CreateOptions{Name: "login-google"})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(c.Worktree); err != nil {
		t.Fatal(err)
	}
	_, ok, err := findWorktree(ctx, r, "login-google")
	if err != nil || ok {
		t.Errorf("deleted: %v %v", ok, err)
	}
}
