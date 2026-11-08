package git

// These tests are in package git, not git_test, because run is not
// exported.

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// envFakeGit turns this test program into a fake git. A test copies the
// program into a folder, names it git, and puts that folder alone in
// PATH. When run sees it, the program starts again, finds this variable,
// and acts as the mode says, instead of running the tests:
//   - "sleep": wait for a minute, so a test can stop it while it runs;
//   - "fail": print a line to stdout and a line to stderr, and exit
//     with 3;
//   - "background": start one more copy of itself in the mode "linger",
//     which gets its stderr and so holds it open, print a line to
//     stdout, and exit with 0 at once, like a git whose hook started a
//     background program;
//   - "linger": wait for 30 seconds.
const envFakeGit = "SLICEFLOW_TEST_FAKE_GIT"

// The exact output of the fake git in the mode "fail".
const (
	fakeStdout   = "half an answer\n"
	fakeStderr   = "fatal: made up\n"
	fakeExitCode = 3
)

func TestMain(m *testing.M) {
	switch os.Getenv(envFakeGit) {
	case "":
		os.Exit(m.Run())
	case "sleep":
		time.Sleep(time.Minute)
		os.Exit(0)
	case "fail":
		fmt.Fprint(os.Stdout, fakeStdout)
		fmt.Fprint(os.Stderr, fakeStderr)
		os.Exit(fakeExitCode)
	case "background":
		self, err := os.Executable()
		if err != nil {
			os.Exit(98)
		}
		child := exec.Command(self)
		// The last value of a repeated variable wins.
		child.Env = append(os.Environ(), envFakeGit+"=linger")
		child.Stderr = os.Stderr
		if err := child.Start(); err != nil {
			os.Exit(97)
		}
		fmt.Fprint(os.Stdout, fakeStdout)
		os.Exit(0)
	case "linger":
		time.Sleep(30 * time.Second)
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "unknown %s mode %q\n", envFakeGit, os.Getenv(envFakeGit))
		os.Exit(99)
	}
}

// Test 1: when git works, run returns git's stdout exactly as git printed
// it, with its line break, and no error.
func TestRunReturnsStdout(t *testing.T) {
	repo := newRepo(t)

	out, err := run(context.Background(), repo, "rev-parse", "--is-inside-work-tree")
	if err != nil {
		t.Fatalf("run returned the error %v, want none", err)
	}
	if out != "true\n" {
		t.Errorf("run returned %q, want %q", out, "true\n")
	}
}

// Test 2: each argument reaches git as it is, even with a space and a ;
// in it. There is no shell to split it or to read the ; as a new
// command.
func TestRunPassesEachArgumentAsItIs(t *testing.T) {
	repo := newRepo(t)
	value := "a b; c"

	if _, err := run(context.Background(), repo, "config", "test.value", value); err != nil {
		t.Fatalf("run config returned the error %v, want none", err)
	}
	out, err := run(context.Background(), repo, "config", "--get", "test.value")
	if err != nil {
		t.Fatalf("run config --get returned the error %v, want none", err)
	}
	if out != value+"\n" {
		t.Errorf("git config --get gave %q, want %q", out, value+"\n")
	}
}

// Test 3: when git runs and fails, run returns a *Error with git's exit
// code and its stderr, and no stdout. The folder is not a repo, so this
// also shows that run starts git in dir: the same command works in a
// repo (test 1).
func TestRunReturnsGitsErrorWhenGitFails(t *testing.T) {
	dir := t.TempDir()
	// Stop git from looking for a repo in the folders above dir.
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(dir))
	// Git prints its messages in the user's language. The C locale makes
	// it print them in English, so the text below can be found.
	t.Setenv("LC_ALL", "C")

	out, err := run(context.Background(), dir, "rev-parse", "--is-inside-work-tree")

	gitErr := wantGitError(t, err)
	if gitErr.ExitCode != 128 {
		t.Errorf("ExitCode = %d, want 128", gitErr.ExitCode)
	}
	if !strings.Contains(gitErr.Stderr, "not a git repository") {
		t.Errorf("Stderr = %q, want it to contain %q", gitErr.Stderr, "not a git repository")
	}
	wantNoStdout(t, out)
}

// Test 4: Stderr holds everything git printed to its error output, word
// for word (R37), and ExitCode is git's own code. A fake git is used,
// because real git's text can change between versions. The fake also
// prints to stdout before it fails: run still returns no stdout.
func TestRunKeepsStderrWordForWord(t *testing.T) {
	useFakeGit(t, "fail")

	out, err := run(context.Background(), t.TempDir(), "anything")

	gitErr := wantGitError(t, err)
	if gitErr.ExitCode != fakeExitCode {
		t.Errorf("ExitCode = %d, want %d", gitErr.ExitCode, fakeExitCode)
	}
	if gitErr.Stderr != fakeStderr {
		t.Errorf("Stderr = %q, want %q", gitErr.Stderr, fakeStderr)
	}
	wantNoStdout(t, out)
}

// Test 5: when git is not installed or not on the path, run returns
// ErrNotFound (R14).
func TestRunReturnsErrNotFound(t *testing.T) {
	t.Setenv("PATH", "")

	out, err := run(context.Background(), t.TempDir(), "status")

	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("run returned the error %v, want ErrNotFound", err)
	}
	wantNoStdout(t, out)
}

// Test 6: when the context ends while git runs, run stops git and returns
// the context's error. Go kills git, and a killed git looks like git that
// failed (an *exec.ExitError). So this test fails if run asks about the
// exit code before it asks about the context. A context that is already
// cancelled before git starts would not catch that, so the fake git
// sleeps, and the time limit ends while it runs.
func TestRunStopsGitWhenTheContextEnds(t *testing.T) {
	useFakeGit(t, "sleep")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()
	out, err := run(ctx, t.TempDir(), "anything")
	took := time.Since(start)

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("run returned the error %v (%T), want context.DeadlineExceeded", err, err)
	}
	var gitErr *Error
	if errors.As(err, &gitErr) {
		t.Errorf("run returned a *Error %v, want only the context's error", gitErr)
	}
	if took > 10*time.Second {
		t.Errorf("run took %v, want it to stop git at once", took)
	}
	wantNoStdout(t, out)
}

// Test 8: when git exits with 0, but a program it started (such as a
// hook's background job) still holds git's stderr open, run waits for
// at most a few seconds, then returns git's stdout and no error: git
// did its job. Without WaitDelay, run would wait for the program to
// end, 30 seconds here, and forever for a program that never ends.
func TestRunDoesNotWaitForProgramsGitStarted(t *testing.T) {
	useFakeGit(t, "background")

	start := time.Now()
	out, err := run(context.Background(), t.TempDir(), "anything")
	took := time.Since(start)

	if err != nil {
		t.Fatalf("run returned the error %v (%T), want none: git exited with 0", err, err)
	}
	if out != fakeStdout {
		t.Errorf("run returned %q, want %q", out, fakeStdout)
	}
	if took > 15*time.Second {
		t.Errorf("run took %v, want it to stop waiting after a few seconds", took)
	}
}

// Test 7: any other failure is returned as it is. Here dir does not
// exist, so git never starts.
func TestRunReturnsOtherErrorsAsTheyAre(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing")

	out, err := run(context.Background(), missing, "status")

	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("run returned the error %v (%T), want one that is fs.ErrNotExist", err, err)
	}
	if errors.Is(err, ErrNotFound) {
		t.Errorf("run returned ErrNotFound, want the error as it is")
	}
	var gitErr *Error
	if errors.As(err, &gitErr) {
		t.Errorf("run returned a *Error %v, want the error as it is", gitErr)
	}
	wantNoStdout(t, out)
}

// newRepo makes a new git repo in a temporary folder and returns its
// path. It runs git with os/exec itself, not with run, so a broken run
// cannot break the setup of its own tests.
func newRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if out, err := exec.Command("git", "init", "--quiet", dir).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	return dir
}

// useFakeGit makes run find the fake git, in the given mode, and no
// other git. It copies this test program into a new folder as git (or
// git.exe on Windows), and makes that folder the only one in PATH, for
// this test only.
func useFakeGit(t *testing.T, mode string) {
	t.Helper()
	self, err := os.Executable()
	if err != nil {
		t.Fatalf("find the test program: %v", err)
	}
	data, err := os.ReadFile(self)
	if err != nil {
		t.Fatalf("read the test program: %v", err)
	}
	name := "git"
	if runtime.GOOS == "windows" {
		name = "git.exe"
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, name), data, 0o755); err != nil {
		t.Fatalf("write the fake git: %v", err)
	}
	t.Setenv("PATH", dir)
	t.Setenv(envFakeGit, mode)
}

// wantGitError fails the test now unless err is a *Error, and returns it.
func wantGitError(t *testing.T, err error) *Error {
	t.Helper()
	var gitErr *Error
	if !errors.As(err, &gitErr) {
		t.Fatalf("run returned the error %v (%T), want a *Error", err, err)
	}
	return gitErr
}

// wantNoStdout fails the test unless out is empty: a failure returns no
// stdout.
func wantNoStdout(t *testing.T, out string) {
	t.Helper()
	if out != "" {
		t.Errorf("run returned the stdout %q with its error, want \"\"", out)
	}
}
