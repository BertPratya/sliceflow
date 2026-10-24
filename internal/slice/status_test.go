package slice

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BertPratya/sliceflow/internal/eventlog"
	"github.com/BertPratya/sliceflow/internal/testrepo"
)

func TestStatusReadsBranch(t *testing.T) {
	r := newRepo(t)
	ctx := context.Background()
	mainCommit := gitOut(t, r.Dir, "rev-parse", "main")
	if _, err := Create(ctx, r, CreateOptions{Name: "spec-submit"}); err != nil {
		t.Fatal(err)
	}
	got, err := Status(ctx, r, "spec-submit")
	if err != nil {
		t.Fatal(err)
	}
	if got.Slice != "spec-submit" || got.Phase != "planning" || got.Title != "" {
		t.Errorf("got %+v", got)
	}
	if got.From == nil || *got.From != (eventlog.From{Branch: "main", Commit: mainCommit}) {
		t.Errorf("From = %+v", got.From)
	}
	if !got.Created.Equal(fixedNow) {
		t.Errorf("Created = %v", got.Created)
	}
	if len(got.Warnings) != 0 {
		t.Errorf("Warnings = %+v", got.Warnings)
	}
}

func TestStatusReadsCommitted(t *testing.T) {
	r := newRepo(t)
	ctx := context.Background()
	c, err := Create(ctx, r, CreateOptions{Name: "spec-submit", Title: "Spec submit"})
	if err != nil {
		t.Fatal(err)
	}
	logFile := filepath.Join(filepath.FromSlash(c.Worktree), "docs", "slices", "spec-submit", ".events.jsonl")
	if err := os.WriteFile(logFile, []byte("garbage\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := Status(ctx, r, "spec-submit")
	if err != nil {
		t.Fatal(err)
	}
	if got.Phase != "planning" || got.Title != "Spec submit" || len(got.Warnings) != 0 {
		t.Errorf("got %+v", got)
	}
}

func TestStatusMergedFromBase(t *testing.T) {
	r := newRepo(t)
	ctx := context.Background()
	c, err := Create(ctx, r, CreateOptions{Name: "spec-submit"})
	if err != nil {
		t.Fatal(err)
	}
	testrepo.Git(t, r.Dir, "merge", "--no-ff", "-m", "merge", "slice/spec-submit")
	testrepo.Git(t, r.Dir, "worktree", "remove", "--force", c.Worktree)
	testrepo.Git(t, r.Dir, "branch", "-D", "slice/spec-submit")
	got, err := Status(ctx, r, "spec-submit")
	if err != nil {
		t.Fatal(err)
	}
	if got.Phase != "planning" || got.From == nil || !got.Created.Equal(fixedNow) {
		t.Errorf("got %+v", got)
	}
}

func TestStatusNoLog(t *testing.T) {
	r := newRepo(t)
	testrepo.Git(t, r.Dir, "branch", "slice/old-idea")
	got, err := Status(context.Background(), r, "old-idea")
	if err != nil {
		t.Fatal(err)
	}
	if got.Phase != "unknown" || got.From != nil || !got.Created.IsZero() {
		t.Errorf("got %+v", got)
	}
	want := []Warning{{Slice: "old-idea", Text: ".events.jsonl not found"}}
	if len(got.Warnings) != 1 || got.Warnings[0] != want[0] {
		t.Errorf("Warnings = %+v", got.Warnings)
	}
}

func TestStatusBadName(t *testing.T) {
	r := newRepo(t)
	trace := filepath.Join(t.TempDir(), "trace")
	t.Setenv("GIT_TRACE", trace)
	_, err := Status(context.Background(), r, "../x")
	want := "no slice named ../x. See all slices: sliceflow slice list"
	var e *Error
	if err == nil || !errors.As(err, &e) || e.Message != want {
		t.Fatalf("err = %v, want %q", err, want)
	}
	data, _ := os.ReadFile(trace)
	if strings.Contains(string(data), "../x") {
		t.Errorf("a git command used the name:\n%s", data)
	}
}

func TestStatusUnknownName(t *testing.T) {
	r := newRepo(t)
	_, err := Status(context.Background(), r, "no-such-slice")
	want := "no slice named no-such-slice. See all slices: sliceflow slice list"
	var e *Error
	if err == nil || !errors.As(err, &e) || e.Message != want {
		t.Fatalf("err = %v, want %q", err, want)
	}
}

func TestStatusBaseMissing(t *testing.T) {
	r := newRepo(t)
	ctx := context.Background()
	if _, err := Create(ctx, r, CreateOptions{Name: "spec-submit"}); err != nil {
		t.Fatal(err)
	}
	testrepo.Git(t, r.Dir, "branch", "-m", "main", "master")
	got, err := Status(ctx, r, "spec-submit")
	if err != nil {
		t.Fatal(err)
	}
	want := Warning{Text: "base branch main does not exist (the default); merged slices are not shown"}
	if len(got.Warnings) != 1 || got.Warnings[0] != want {
		t.Errorf("Warnings = %+v", got.Warnings)
	}
	if got.Phase != "planning" {
		t.Errorf("Phase = %q", got.Phase)
	}
	base, exists, w, err := baseWarning(ctx, r)
	if err != nil || base != "main" || exists || w == nil || *w != want {
		t.Errorf("baseWarning = %q %v %+v %v", base, exists, w, err)
	}
}
