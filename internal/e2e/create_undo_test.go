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

	"github.com/BertPratya/sliceflow/internal/gittest"
)

// writeHook writes a pre-commit hook (a shell script) into the repo.
func writeHook(t *testing.T, repo, body string) {
	t.Helper()
	path := filepath.Join(repo, ".git", "hooks", "pre-commit")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
}

// checkNothingLeft checks that the slice has no branch and no worktree.
func checkNothingLeft(t *testing.T, repo, name string) {
	t.Helper()
	if out := gittest.Run(t, repo, "branch", "--list", "slice/"+name); out != "" {
		t.Errorf("branch left: %q", out)
	}
	if _, err := os.Stat(filepath.Join(repo, ".sliceflow", "worktrees", name)); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("worktree folder left: %v", err)
	}
	out := gittest.Run(t, repo, "worktree", "list", "--porcelain")
	if n := strings.Count(out, "worktree "); n != 1 {
		t.Errorf("git worktree list shows %d worktrees, want only the main checkout:\n%s", n, out)
	}
}

// TestS12: a hook that refuses every commit and prints nothing.
func TestS12(t *testing.T) {
	repo := gittest.NewRepo(t)
	writeHook(t, repo, "exit 1")

	r := run(t, repo, "slice", "create", "spec-submit")
	want := "Error: could not create slice spec-submit: git commit failed (exit code 1). Everything sliceflow made was removed.\n"
	if r.code != 1 || r.stdout != "" || r.stderr != want {
		t.Fatalf("code = %d, stdout = %q, stderr = %q, want 1, \"\", %q", r.code, r.stdout, r.stderr, want)
	}
	checkNothingLeft(t, repo, "spec-submit")
}

// TestCreateFailureShowsHookOutput: the first line of the hook's output is
// the reason; the next lines follow, indented by 2 spaces.
func TestCreateFailureShowsHookOutput(t *testing.T) {
	repo := gittest.NewRepo(t)
	writeHook(t, repo, "echo first >&2; echo second >&2; exit 1")

	r := run(t, repo, "slice", "create", "x")
	want := "Error: could not create slice x: first. Everything sliceflow made was removed.\n  second\n"
	if r.code != 1 || r.stdout != "" || r.stderr != want {
		t.Fatalf("code = %d, stdout = %q, stderr = %q, want 1, \"\", %q", r.code, r.stdout, r.stderr, want)
	}

	// R42: with --json the failure is the same, and stdout stays empty.
	j := run(t, repo, "slice", "create", "x", "--json")
	if j.code != 1 || j.stdout != "" || j.stderr != want {
		t.Fatalf("--json: code = %d, stdout = %q, stderr = %q, want 1, \"\", %q", j.code, j.stdout, j.stderr, want)
	}
	checkNothingLeft(t, repo, "x")
}

// TestCreateInterrupted: Ctrl-C while the hook runs. The hook touches a
// marker, so the test knows sliceflow is in the commit step. This covers
// "a step succeeded, but a signal came", not a step killed by the signal:
// only the program gets the signal, so git and the hook finish normally.
func TestCreateInterrupted(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("os.Interrupt cannot be sent to another process on Windows")
	}
	repo := gittest.NewRepo(t)
	marker := filepath.ToSlash(filepath.Join(t.TempDir(), "marker"))
	writeHook(t, repo, "touch '"+marker+"'\nsleep 1\nexit 0")

	cmd := exec.Command(binary, "slice", "create", "x")
	cmd.Dir = repo
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(10 * time.Second)
	for {
		if _, err := os.Stat(marker); err == nil {
			break
		}
		if time.Now().After(deadline) {
			cmd.Process.Kill()
			t.Fatal("the hook did not start")
		}
		time.Sleep(10 * time.Millisecond)
	}
	// Signal only the program, not git or the hook.
	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}

	err := cmd.Wait()
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
		t.Fatalf("Wait error = %v, want exit code 1", err)
	}
	want := "Error: could not create slice x: interrupted. Everything sliceflow made was removed.\n"
	if stderr.String() != want || stdout.String() != "" {
		t.Errorf("stdout = %q, stderr = %q, want \"\", %q", stdout.String(), stderr.String(), want)
	}
	checkNothingLeft(t, repo, "x")
}
