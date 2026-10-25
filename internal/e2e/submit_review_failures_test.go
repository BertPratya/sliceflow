package e2e

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/BertPratya/sliceflow/internal/testrepo"
)

const (
	refusedCommit  = "Error: could not save the event: git commit failed (exit code 1). Nothing was changed.\n"
	startedOnly    = "Review 1 started: reviews/spec-1.md (reviewer: placeholder)\n"
	statusSubmit   = "Spec:     submitted (review 1 not started)\nNext:     sliceflow spec review\n"
	hookRefuseWord = `grep -q '%s' "$1" && exit 1` + "\nexit 0"
)

func logPath(dir string) string {
	return filepath.Join(wtOf(dir), "docs", "slices", specName, ".events.jsonl")
}

// checkLogCommitted fails when the log on disk is not the one in HEAD.
func checkLogCommitted(t *testing.T, dir string) {
	t.Helper()
	disk, err := os.ReadFile(logPath(dir))
	if err != nil {
		t.Fatal(err)
	}
	head := testrepo.Git(t, wtOf(dir), "show", "HEAD:docs/slices/"+specName+"/.events.jsonl")
	if string(disk) != head {
		t.Fatalf("log on disk %q is not the committed log %q", disk, head)
	}
}

func commitMsgHook(t *testing.T, dir, word string) {
	t.Helper()
	writeUndoHook(t, dir, "commit-msg", strings.Replace(hookRefuseWord, "%s", word, 1))
}

func removeHook(t *testing.T, dir, name string) {
	t.Helper()
	hooks := strings.TrimSpace(testrepo.Git(t, dir, "rev-parse", "--git-path", "hooks"))
	if !filepath.IsAbs(hooks) {
		hooks = filepath.Join(dir, hooks)
	}
	if err := os.Remove(filepath.Join(hooks, name)); err != nil {
		t.Fatal(err)
	}
}

func lockFile(dir string) string {
	return filepath.Join(dir, ".sliceflow", "locks", specName+".lock")
}

func TestSubmitBusy(t *testing.T) {
	dir := draftSlice(t, s15Text())
	if err := os.MkdirAll(filepath.Dir(lockFile(dir)), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(lockFile(dir), []byte(`{"pid":4312,"since":"2026-10-24T10:02:00Z"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	before := headOf(t, dir)
	start := time.Now()
	out, errOut, code := run(t, wtOf(dir), "spec", "submit")
	want := "Error: slice login-google is busy: another sliceflow command holds .sliceflow/locks/login-google.lock (pid 4312, since 10:02 UTC). If no sliceflow command is running, delete that file.\n"
	if code != 1 || out != "" || errOut != want {
		t.Fatalf("code %d\nstdout %q\nstderr %q\nwant   %q", code, out, errOut, want)
	}
	if d := time.Since(start); d < 9*time.Second {
		t.Fatalf("waited only %v", d)
	}
	if headOf(t, dir) != before {
		t.Fatal("a commit was made")
	}
}

func TestSubmitHookFails(t *testing.T) {
	dir := draftSlice(t, s15Text())
	writeUndoHook(t, dir, "pre-commit", "exit 1")
	before := headOf(t, dir)
	out, errOut, code := run(t, wtOf(dir), "spec", "submit")
	if code != 1 || errOut != refusedCommit || strings.Contains(out, "Submitted") {
		t.Fatalf("code %d\nstdout %q\nstderr %q", code, out, errOut)
	}
	if headOf(t, dir) != before {
		t.Fatal("a commit was made")
	}
	checkLogCommitted(t, dir)
	if got := testrepo.Git(t, wtOf(dir), "diff", "--cached", "--name-only"); strings.TrimSpace(got) != "" {
		t.Fatalf("staged: %q", got)
	}
}

func TestSubmitOwnFiles(t *testing.T) {
	dir := draftSlice(t, s15Text())
	wt := wtOf(dir)
	if err := os.MkdirAll(filepath.Join(wt, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wt, "src", "app.go"), []byte("package src\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	testrepo.Git(t, wt, "add", "src/app.go")
	before := headOf(t, dir)
	if _, e, c := run(t, wt, "spec", "submit"); c != 0 {
		t.Fatalf("submit: code %d stderr %q", c, e)
	}
	if got := testrepo.Git(t, wt, "log", "--name-only", "--format=", before+"..HEAD"); strings.Contains(got, "src/app.go") || got == "" {
		t.Fatalf("files in new commits: %q", got)
	}
	if got := strings.TrimSpace(testrepo.Git(t, wt, "diff", "--cached", "--name-only")); got != "src/app.go" {
		t.Fatalf("staged: %q", got)
	}
}

func TestReviewFileExistsCmd(t *testing.T) {
	dir := submittedSlice(t)
	p := filepath.Join(wtOf(dir), "docs", "slices", specName, "reviews")
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(p, "spec-1.md"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	failsWith(t, dir, "Error: docs/slices/login-google/reviews/spec-1.md already exists, but review 1 has not started. Move it away, then review again. Nothing was changed.", "spec", "review")
}

func TestReviewCommitRefused(t *testing.T) {
	dir := submittedSlice(t)
	commitMsgHook(t, dir, "review")
	before := headOf(t, dir)
	out, errOut, code := run(t, wtOf(dir), "spec", "review")
	if code != 1 || out != "" || errOut != refusedCommit {
		t.Fatalf("code %d\nstdout %q\nstderr %q", code, out, errOut)
	}
	if headOf(t, dir) != before {
		t.Fatal("a commit was made")
	}
	if got := statusTail(t, dir); got != statusSubmit {
		t.Fatalf("status %q", got)
	}
}

func TestReviewAgain(t *testing.T) {
	dir := submittedSlice(t)
	commitMsgHook(t, dir, "review")
	if _, _, c := run(t, wtOf(dir), "spec", "review"); c != 1 {
		t.Fatal("review should fail")
	}
	removeHook(t, dir, "commit-msg")
	before := headOf(t, dir)
	out, errOut, code := run(t, wtOf(dir), "spec", "review")
	if code != 0 || out != reviewed || errOut != "" {
		t.Fatalf("code %d\nstdout %q\nstderr %q", code, out, errOut)
	}
	if n := testrepo.Git(t, wtOf(dir), "rev-list", "--count", before+"..HEAD"); strings.TrimSpace(n) != "2" {
		t.Fatalf("commits: %s", n)
	}
	s := subjects(t, dir, "2")
	if len(s) != 2 || s[0] != "login-google: review 1 passed" || s[1] != "login-google: review 1 started" {
		t.Fatalf("subjects %v", s)
	}
}

func TestReviewStopsHalfway(t *testing.T) {
	dir := submittedSlice(t)
	commitMsgHook(t, dir, "passed")
	out, errOut, code := run(t, wtOf(dir), "spec", "review")
	wantErr := "Error: could not save the event: git commit failed (exit code 1). The steps before it were saved; spec.md is in_review.\n"
	if code != 1 || out != startedOnly || errOut != wantErr {
		t.Fatalf("code %d\nstdout %q\nstderr %q", code, out, errOut)
	}
	want := "Spec:     in_review (review 1, placeholder: reviews/spec-1.md)\nNext:     wait for review 1\n"
	if got := statusTail(t, dir); got != want {
		t.Fatalf("status %q", got)
	}
}

func TestSubmitInterrupted(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Go cannot send os.Interrupt to another program on Windows")
	}
	dir := draftSlice(t, s15Text())
	marker := filepath.Join(t.TempDir(), "started")
	writeUndoHook(t, dir, "pre-commit", ": > '"+marker+"'\nsleep 30")
	testrepo.Setup(t)
	cmd := exec.Command(binPath, "spec", "submit")
	cmd.Dir = wtOf(dir)
	var o, e bytes.Buffer
	cmd.Stdout = &o
	cmd.Stderr = &e
	cmd.WaitDelay = 2 * time.Second
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(10 * time.Second)
	for {
		if _, err := os.Stat(marker); err == nil {
			break
		}
		if time.Now().After(deadline) {
			_ = cmd.Process.Kill()
			t.Fatal("hook did not start")
		}
		time.Sleep(20 * time.Millisecond)
	}
	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	var err error
	select {
	case err = <-done:
	case <-time.After(10 * time.Second):
		_ = cmd.Process.Kill()
		t.Fatal("did not stop within 10 seconds")
	}
	code := 0
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		code = ee.ExitCode()
	} else if err != nil {
		t.Fatal(err)
	}
	want := "Error: could not save the event: interrupted. Nothing was changed.\n"
	if code != 1 || e.String() != want || strings.Contains(o.String(), "Submitted") {
		t.Fatalf("code %d stdout %q stderr %q", code, o.String(), e.String())
	}
	checkLogCommitted(t, dir)
	if _, err := os.Lstat(lockFile(dir)); err == nil {
		t.Fatal("lock file left")
	}
}
