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

func TestStatusTagIsNotBase(t *testing.T) {
	r := newRepo(t)
	ctx := context.Background()
	if _, err := Create(ctx, r, CreateOptions{Name: "spec-submit"}); err != nil {
		t.Fatal(err)
	}
	testrepo.Git(t, r.Dir, "branch", "-m", "main", "master")
	testrepo.Git(t, r.Dir, "tag", "main")
	got, err := Status(ctx, r, "spec-submit")
	if err != nil {
		t.Fatal(err)
	}
	want := Warning{Text: "base branch main does not exist (the default); merged slices are not shown"}
	if len(got.Warnings) != 1 || got.Warnings[0] != want {
		t.Errorf("Warnings = %+v", got.Warnings)
	}
}

func TestStatusOldTagSameNameAsBase(t *testing.T) {
	r := newRepo(t)
	ctx := context.Background()
	testrepo.Git(t, r.Dir, "tag", "main")
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
	if got.Phase != "planning" || got.From == nil {
		t.Errorf("got %+v", got)
	}
}

func TestStatusBaseConfigHEAD(t *testing.T) {
	r := newRepo(t)
	ctx := context.Background()
	if _, err := Create(ctx, r, CreateOptions{Name: "spec-submit"}); err != nil {
		t.Fatal(err)
	}
	testrepo.Git(t, r.Dir, "config", "sliceflow.base", "HEAD")
	got, err := Status(ctx, r, "spec-submit")
	if err != nil {
		t.Fatal(err)
	}
	want := Warning{Text: "base branch HEAD does not exist (from git config sliceflow.base); merged slices are not shown"}
	if len(got.Warnings) != 1 || got.Warnings[0] != want {
		t.Errorf("Warnings = %+v", got.Warnings)
	}
}

func appendLog(t *testing.T, worktree, name string, evs ...eventlog.Event) {
	t.Helper()
	logFile := filepath.Join(filepath.FromSlash(worktree), "docs", "slices", name, ".events.jsonl")
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	for _, e := range evs {
		b, err := eventlog.Encode(e)
		if err != nil {
			t.Fatal(err)
		}
		f.Write(b)
		if len(b) == 0 || b[len(b)-1] != '\n' {
			f.WriteString("\n")
		}
	}
	f.Close()
	testrepo.Git(t, worktree, "add", "--", logFile)
	testrepo.Git(t, worktree, "commit", "-m", "events")
}

func TestStatusState(t *testing.T) {
	r := newRepo(t)
	ctx := context.Background()
	c, err := Create(ctx, r, CreateOptions{Name: "login-google", Title: "Login"})
	if err != nil {
		t.Fatal(err)
	}
	appendLog(t, c.Worktree, "login-google", eventlog.New(2, fixedNow, "slice.titled", eventlog.Titled{Title: "Sign in with Google"}))
	// An uncommitted line on disk must not count.
	logFile := filepath.Join(filepath.FromSlash(c.Worktree), "docs", "slices", "login-google", ".events.jsonl")
	f, _ := os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY, 0o644)
	f.WriteString("garbage\n")
	f.Close()
	got, err := Status(ctx, r, "login-google")
	if err != nil {
		t.Fatal(err)
	}
	if got.Phase != "planning" || got.Title != "Sign in with Google" || got.State.Title != got.Title {
		t.Errorf("got %+v", got)
	}
	if got.SpecText != "draft" || got.Next != "write spec.md, then: sliceflow spec submit" || got.Merged || len(got.Warnings) != 0 {
		t.Errorf("got %+v", got)
	}
}

func TestStatusMergedNext(t *testing.T) {
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
	if !got.Merged || got.Next != "-" || got.SpecText != "draft" {
		t.Errorf("got %+v", got)
	}
}

func TestStatusChangedAfterReview(t *testing.T) {
	r := newRepo(t)
	ctx := context.Background()
	c, err := Create(ctx, r, CreateOptions{Name: "login-google"})
	if err != nil {
		t.Fatal(err)
	}
	wt := filepath.FromSlash(c.Worktree)
	specFile := filepath.Join(wt, "docs", "slices", "login-google", "spec.md")
	if err := os.WriteFile(specFile, []byte("spec\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	testrepo.Git(t, wt, "add", "--", specFile)
	testrepo.Git(t, wt, "commit", "-m", "spec")
	commit := gitOut(t, wt, "rev-parse", "HEAD")
	appendLog(t, c.Worktree, "login-google",
		eventlog.New(2, fixedNow, "spec.submitted", eventlog.Submitted{Result: "accepted"}),
		eventlog.New(3, fixedNow, "review.started", eventlog.ReviewStarted{Doc: "spec.md", Round: 1, Reviewer: "placeholder", File: "reviews/spec-1.md", Commit: commit}),
		eventlog.New(4, fixedNow, "review.finished", eventlog.ReviewFinished{Doc: "spec.md", Round: 1, Result: "passed"}))
	got, err := Status(ctx, r, "login-google")
	if err != nil {
		t.Fatal(err)
	}
	if got.ChangedAfterReview || got.SpecText != "review_passed (review 1, placeholder)" || got.Next != "write the plan" {
		t.Errorf("unchanged: %+v", got)
	}
	if err := os.WriteFile(specFile, []byte("spec changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err = Status(ctx, r, "login-google")
	if err != nil {
		t.Fatal(err)
	}
	if !got.ChangedAfterReview || got.SpecText != "passed review 1, but spec.md was changed after it" || got.Next != "sliceflow spec submit" {
		t.Errorf("changed: %+v", got)
	}
}

func TestCurrentSliceNameCommand(t *testing.T) {
	r := newRepo(t)
	_, err := currentSliceName(context.Background(), r, "spec validate")
	want := "not on a slice branch (you are on main).\nGive a name: sliceflow spec validate <name>\nSee all slices: sliceflow slice list"
	var e *Error
	if !errors.As(err, &e) || e.Message != want {
		t.Errorf("err = %v", err)
	}
}
