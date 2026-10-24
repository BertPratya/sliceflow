package slice

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/BertPratya/sliceflow/internal/git"
	"github.com/BertPratya/sliceflow/internal/testrepo"
)

func writeHook(t *testing.T, dir, name, body string) {
	t.Helper()
	hooks := strings.TrimSpace(testrepo.Git(t, dir, "rev-parse", "--git-path", "hooks"))
	if !filepath.IsAbs(hooks) {
		hooks = filepath.Join(dir, hooks)
	}
	if err := os.MkdirAll(hooks, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hooks, name), []byte("#!/bin/sh\n"+body+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
}

func setAfterStep(t *testing.T, f func(step string)) {
	t.Helper()
	afterStep = f
	t.Cleanup(func() { afterStep = nil })
}

func branchExists(t *testing.T, dir, branch string) bool {
	t.Helper()
	return gitOut(t, dir, "branch", "--list", branch) != ""
}

const interruptedMsg = "could not create slice x: interrupted. Everything sliceflow made was removed."

func TestUndoCancelledContext(t *testing.T) {
	r := newRepo(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	setAfterStep(t, func(step string) {
		if step == "worktree" {
			cancel()
		}
	})
	_, err := Create(ctx, r, CreateOptions{Name: "x"})
	wantSliceError(t, err, interruptedMsg)
	if branchExists(t, r.Dir, "slice/x") {
		t.Error("branch left")
	}
	if strings.Contains(gitOut(t, r.Dir, "worktree", "list", "--porcelain"), "worktrees/x") {
		t.Error("worktree left")
	}
	if exists(filepath.Join(r.Root, ".sliceflow", "worktrees")) || exists(filepath.Join(r.Root, ".sliceflow")) {
		t.Error(".sliceflow left")
	}
}

func TestUndoCancelledDuringChecks(t *testing.T) {
	r := newRepo(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := Create(ctx, r, CreateOptions{Name: "x"})
	wantSliceError(t, err, interruptedMsg)
	if exists(filepath.Join(r.Root, ".sliceflow")) {
		t.Error(".sliceflow made")
	}
}

func TestUndoInterruptedWorktreeAdd(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("hooks that sleep are not portable here")
	}
	r := newRepo(t)
	marker := filepath.Join(t.TempDir(), "marker")
	writeHook(t, r.Dir, "post-checkout", "echo started > '"+filepath.ToSlash(marker)+"'\nsleep 30")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		for i := 0; i < 400 && ctx.Err() == nil; i++ {
			if exists(marker) {
				cancel()
				return
			}
			time.Sleep(25 * time.Millisecond)
		}
	}()
	start := time.Now()
	_, err := Create(ctx, r, CreateOptions{Name: "spec-submit"})
	if d := time.Since(start); d > 10*time.Second {
		t.Errorf("took %v", d)
	}
	wantSliceError(t, err, "could not create slice spec-submit: interrupted. Everything sliceflow made was removed.")
	if branchExists(t, r.Dir, "slice/spec-submit") {
		t.Error("branch left")
	}
	if strings.Contains(gitOut(t, r.Dir, "worktree", "list", "--porcelain"), "spec-submit") {
		t.Error("worktree left")
	}
	if exists(filepath.Join(r.Root, ".sliceflow", "worktrees", "spec-submit")) || exists(filepath.Join(r.Root, ".sliceflow")) {
		t.Error(".sliceflow left")
	}
}

func TestUndoFailedWorktreeAdd(t *testing.T) {
	r := newRepo(t)
	writeHook(t, r.Dir, "post-checkout", "exit 3")
	_, err := Create(context.Background(), r, CreateOptions{Name: "spec-submit"})
	wantSliceError(t, err, "could not create slice spec-submit: git worktree add failed (exit code 3). Everything sliceflow made was removed.")
	if branchExists(t, r.Dir, "slice/spec-submit") {
		t.Error("branch left")
	}
	if strings.Contains(gitOut(t, r.Dir, "worktree", "list", "--porcelain"), "spec-submit") {
		t.Error("worktree left")
	}
	if exists(filepath.Join(r.Root, ".sliceflow", "worktrees", "spec-submit")) || exists(filepath.Join(r.Root, ".sliceflow")) {
		t.Error(".sliceflow left")
	}
}

func TestUndoKeepsStaleEntry(t *testing.T) {
	r := newRepo(t)
	wt := filepath.Join(r.Root, ".sliceflow", "worktrees", "spec-submit")
	testrepo.Git(t, r.Dir, "worktree", "add", "--detach", wt)
	if err := os.RemoveAll(wt); err != nil {
		t.Fatal(err)
	}
	_, err := Create(context.Background(), r, CreateOptions{Name: "spec-submit"})
	if err == nil || !strings.HasPrefix(err.Error(), "could not create slice spec-submit: ") {
		t.Fatalf("err = %v", err)
	}
	if !strings.Contains(filepath.ToSlash(gitOut(t, r.Dir, "worktree", "list", "--porcelain")), ".sliceflow/worktrees/spec-submit") {
		t.Error("the stale entry was removed")
	}
	if branchExists(t, r.Dir, "slice/spec-submit") {
		t.Error("branch left")
	}
}

func TestUndoKeepsOthers(t *testing.T) {
	r := newRepo(t)
	testrepo.Git(t, r.Dir, "commit", "--allow-empty", "-m", "second")
	testrepo.Git(t, r.Dir, "branch", "slice/x", "main~1")
	old := gitOut(t, r.Dir, "rev-parse", "slice/x")
	_, err := Create(context.Background(), r, CreateOptions{Name: "x", skipChecks: true})
	if err == nil || !strings.HasPrefix(err.Error(), "could not create slice x: ") {
		t.Fatalf("err = %v", err)
	}
	if got := gitOut(t, r.Dir, "rev-parse", "slice/x"); got != old {
		t.Errorf("slice/x = %s, want %s", got, old)
	}
	if exists(filepath.Join(r.Root, ".sliceflow")) {
		t.Error(".sliceflow left")
	}
}

func TestUndoKeepsExistingFolders(t *testing.T) {
	r := newRepo(t)
	wts := filepath.Join(r.Root, ".sliceflow", "worktrees")
	if err := os.MkdirAll(wts, 0o755); err != nil {
		t.Fatal(err)
	}
	ig := filepath.Join(r.Root, ".sliceflow", ".gitignore")
	if err := os.WriteFile(ig, []byte("*\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeHook(t, r.Dir, "pre-commit", "exit 1")
	_, err := Create(context.Background(), r, CreateOptions{Name: "x"})
	wantSliceError(t, err, "could not create slice x: git commit failed (exit code 1). Everything sliceflow made was removed.")
	if !exists(wts) || !exists(ig) {
		t.Error("existing .sliceflow parts were removed")
	}
	if exists(filepath.Join(wts, "x")) || branchExists(t, r.Dir, "slice/x") {
		t.Error("slice left")
	}
}

func TestUndoRemovesOwnFolders(t *testing.T) {
	r := newRepo(t)
	writeHook(t, r.Dir, "pre-commit", "exit 1")
	_, err := Create(context.Background(), r, CreateOptions{Name: "x"})
	if err == nil {
		t.Fatal("want error")
	}
	if exists(filepath.Join(r.Root, ".sliceflow")) {
		t.Error(".sliceflow left")
	}
}

func TestUndoCouldNotRemove(t *testing.T) {
	r := newRepo(t)
	writeHook(t, r.Dir, "pre-commit", "exit 1")
	setAfterStep(t, func(step string) {
		if step == "worktree" {
			testrepo.Git(t, r.Dir, "worktree", "lock", worktreeDir(r.Root, "x"))
		}
	})
	_, err := Create(context.Background(), r, CreateOptions{Name: "x"})
	if err == nil {
		t.Fatal("want error")
	}
	want := "Could not remove: .sliceflow/worktrees/x, branch slice/x."
	if !strings.HasSuffix(err.Error(), want) {
		t.Errorf("err = %q, want suffix %q", err, want)
	}
}

func TestReasonFileError(t *testing.T) {
	err := &fs.PathError{Op: "open", Path: filepath.FromSlash("C:/a/b/spec.md"), Err: errors.New("Access is denied.")}
	got, extra := failureReason(err, "")
	if got != "open C:/a/b/spec.md: Access is denied" || len(extra) != 0 {
		t.Errorf("got %q %q", got, extra)
	}
}

func TestReasonGitCommand(t *testing.T) {
	got, _ := failureReason(&git.Error{ExitCode: 128}, "worktree add")
	if got != "git worktree add failed (exit code 128)" {
		t.Errorf("got %q", got)
	}
	got, extra := failureReason(&git.Error{ExitCode: 1, Stderr: "\nlint: 2 errors.\na.go:1: bad"}, "commit")
	if got != "lint: 2 errors" || len(extra) != 1 || extra[0] != "a.go:1: bad" {
		t.Errorf("got %q %q", got, extra)
	}
}

func wantSliceError(t *testing.T, err error, msg string) {
	t.Helper()
	var se *Error
	if !errors.As(err, &se) {
		t.Fatalf("want *Error, got %T: %v", err, err)
	}
	if se.Message != msg {
		t.Fatalf("got %q, want %q", se.Message, msg)
	}
}
