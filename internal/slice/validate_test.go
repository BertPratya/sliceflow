package slice

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BertPratya/sliceflow/internal/testrepo"
)

func TestOpenSliceOrder(t *testing.T) {
	r := newRepo(t)
	ctx := context.Background()
	if _, err := Create(ctx, r, CreateOptions{Name: "has-branch"}); err != nil {
		t.Fatal(err)
	}
	c, err := Create(ctx, r, CreateOptions{Name: "merged-one"})
	if err != nil {
		t.Fatal(err)
	}
	testrepo.Git(t, r.Dir, "merge", "--no-ff", "-m", "merge", "slice/merged-one")
	testrepo.Git(t, r.Dir, "worktree", "remove", "--force", c.Worktree)
	testrepo.Git(t, r.Dir, "branch", "-D", "slice/merged-one")

	wt, err := openSlice(ctx, r, "has-branch")
	if err != nil || wt == "" {
		t.Fatalf("open: %q %v", wt, err)
	}
	testrepo.Git(t, r.Dir, "worktree", "remove", "--force", wt)

	_, err = openSlice(ctx, r, "Bad_Name")
	wantErrMsg(t, err, `invalid slice name "Bad_Name": use 1-64 lowercase letters, digits, and dashes, starting with a letter or digit. Nothing was changed.`)
	_, err = openSlice(ctx, r, "no-such-slice")
	wantErrMsg(t, err, "no slice named no-such-slice. See all slices: sliceflow slice list")
	_, err = openSlice(ctx, r, "has-branch")
	wantErrMsg(t, err, "slice has-branch has no worktree: no worktree has the branch slice/has-branch checked out. Nothing was changed.\n"+
		"Make one: git worktree add "+r.Root+"/.sliceflow/worktrees/has-branch slice/has-branch")
	_, err = openSlice(ctx, r, "merged-one")
	wantErrMsg(t, err, "slice merged-one is merged into main, so it has no worktree. Nothing was changed.")
}

func TestValidateCounts(t *testing.T) {
	r := newRepo(t)
	ctx := context.Background()
	c, err := Create(ctx, r, CreateOptions{Name: "demo"})
	if err != nil {
		t.Fatal(err)
	}
	// The fresh template has placeholders and comments: warnings, no panic.
	v, err := Validate(ctx, r, "demo", false)
	if err != nil {
		t.Fatal(err)
	}
	if v.Errors+v.Warnings+v.Notes != len(v.Problems) || len(v.Problems) == 0 {
		t.Fatalf("got %+v", v)
	}
	// No name: the slice of the current branch. Main is not a slice branch.
	if _, err := Validate(ctx, r, "", false); err == nil || !strings.Contains(err.Error(), "not on a slice branch") {
		t.Fatalf("err %v", err)
	}
	// Missing spec.md.
	if err := os.Remove(filepath.Join(filepath.FromSlash(c.Worktree), "docs", "slices", "demo", "spec.md")); err != nil {
		t.Fatal(err)
	}
	_, err = Validate(ctx, r, "demo", false)
	wantErrMsg(t, err, ".sliceflow/worktrees/demo/docs/slices/demo/spec.md not found. Nothing was changed.")
}

func TestValidateMovedWorktree(t *testing.T) {
	r := newRepo(t)
	ctx := context.Background()
	c, err := Create(ctx, r, CreateOptions{Name: "demo"})
	if err != nil {
		t.Fatal(err)
	}
	moved := filepath.Join(r.Dir, ".worktrees", "elsewhere", "demo")
	if err := os.MkdirAll(filepath.Dir(moved), 0o755); err != nil {
		t.Fatal(err)
	}
	testrepo.Git(t, r.Dir, "worktree", "move", filepath.FromSlash(c.Worktree), moved)
	if err := os.Remove(filepath.Join(moved, "docs", "slices", "demo", "spec.md")); err != nil {
		t.Fatal(err)
	}
	_, err = Validate(ctx, r, "demo", false)
	wantErrMsg(t, err, ".worktrees/elsewhere/demo/docs/slices/demo/spec.md not found. Nothing was changed.")
}

func wantErrMsg(t *testing.T, err error, msg string) {
	t.Helper()
	se, ok := err.(*Error)
	if !ok {
		t.Fatalf("want *Error, got %T: %v", err, err)
	}
	if se.Message != msg {
		t.Fatalf("got %q, want %q", se.Message, msg)
	}
}
