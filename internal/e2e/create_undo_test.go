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

func writeUndoHook(t *testing.T, dir, name, body string) {
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

func checkNothingLeft(t *testing.T, dir string) {
	t.Helper()
	if b := testrepo.Git(t, dir, "branch", "--list", "slice/spec-submit"); strings.TrimSpace(b) != "" {
		t.Errorf("branch left: %q", b)
	}
	wl := testrepo.Git(t, dir, "worktree", "list", "--porcelain")
	if n := strings.Count(wl, "worktree "); n != 1 {
		t.Errorf("worktrees left: %q", wl)
	}
	if _, err := os.Lstat(filepath.Join(dir, ".sliceflow")); err == nil {
		t.Error(".sliceflow left")
	}
}

func TestCreateHookRefuses(t *testing.T) {
	dir := testrepo.New(t)
	writeUndoHook(t, dir, "pre-commit", "exit 1")
	out, errOut, code := run(t, dir, "slice", "create", "spec-submit")
	want := "Error: could not create slice spec-submit: git commit failed (exit code 1). Everything sliceflow made was removed.\n"
	if code != 1 || out != "" || errOut != want {
		t.Fatalf("code %d stdout %q stderr %q", code, out, errOut)
	}
	checkNothingLeft(t, dir)
}

func TestCreateHookOutput(t *testing.T) {
	dir := testrepo.New(t)
	writeUndoHook(t, dir, "pre-commit", "echo 'lint: 2 errors.' >&2\necho 'a.go:1: bad' >&2\nexit 1")
	_, errOut, code := run(t, dir, "slice", "create", "spec-submit")
	want := "Error: could not create slice spec-submit: lint: 2 errors. Everything sliceflow made was removed.\n  a.go:1: bad\n"
	if code != 1 || errOut != want {
		t.Fatalf("code %d stderr %q", code, errOut)
	}
	checkNothingLeft(t, dir)
}

func TestCreateInterrupt(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Go cannot send os.Interrupt to another program on Windows")
	}
	dir := testrepo.New(t)
	marker := filepath.Join(t.TempDir(), "started")
	writeUndoHook(t, dir, "pre-commit", ": > '"+marker+"'\nsleep 30")
	cmd := exec.Command(binPath, "slice", "create", "spec-submit")
	cmd.Dir = dir
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
	want := "Error: could not create slice spec-submit: interrupted. Everything sliceflow made was removed.\n"
	if code != 1 || o.String() != "" || e.String() != want {
		t.Fatalf("code %d stdout %q stderr %q", code, o.String(), e.String())
	}
	checkNothingLeft(t, dir)
}
