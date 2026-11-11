// Package git runs git commands for sliceflow, each one as its own
// process, with its arguments as a list and no shell.
package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"
)

// Error is what run returns when git ran and exited with a code that is
// not 0.
type Error struct {
	// ExitCode is git's exit code. Some commands answer a question with
	// it: for example, git show-ref exits with 1 to say "no".
	ExitCode int
	// Stderr is everything git printed to its error output, word for
	// word, for the message of R37. Never decide anything by its text:
	// it changes with the user's language and git's version.
	Stderr string
}

func (e *Error) Error() string {
	return fmt.Sprintf("exit code %d: %s", e.ExitCode, e.Stderr)
}

// ErrNotFound means git could not be run: it is not installed, or not on
// the path (R14).
var ErrNotFound = errors.New("git not found.")

// run runs git with args in the folder dir, and returns what git printed
// to its normal output, exactly as git printed it.
//
// Steps:
//  1. Run git in dir, with its normal output and its error output in two
//     buffers. After git exits, wait at most 5 seconds for its output
//     (WaitDelay): a program that git started, such as a hook's
//     background job, may hold the output open long after git is done.
//  2. When ctx ends while git runs, send git an interrupt, not a kill
//     (interrupt), so git can clean up after itself, for example remove
//     its own lock files (R39). If git has not exited 5 seconds after
//     the interrupt, Go kills it (WaitDelay).
//  3. If git exited with 0, return the normal output and no error. This
//     also holds when the 5 seconds ran out (exec.ErrWaitDelay): git
//     itself exited with 0.
//  4. If ctx ended (Ctrl+C, or a time limit), return ctx.Err(). This
//     comes before step 6: a git that was stopped looks like a git that
//     failed. A git that exits with 0 after the interrupt also returns
//     ctx.Err(): it may have stopped half way. The caller then asks git
//     what is there.
//  5. If git was not found, return ErrNotFound.
//  6. If git exited with another code, return an *Error with that code
//     and git's error output.
//  7. Otherwise, return the error as it is.
//
// On an error, the normal output is "".
func run(ctx context.Context, dir string, args ...string) (string, error) {
	return runGit(ctx, dir, false, args)
}

// runShielded is run, with git in a process group of its own (shield).
// A Ctrl+C in the terminal goes to the terminal's process group, so it
// does not reach this git, even when it is pressed many times. The undo
// after a failure, and the questions it asks git, use it (R36, R39).
//
// It does not shield git from ctx: give it a context that a Ctrl+C does
// not end.
//
// Steps:
//  1. Run git with runGit, shielded.
//  2. If git was ended by a signal, and ctx has not ended, go back to
//     step 1, at most 3 runs in all. Until the child calls setpgid, it is
//     still in sliceflow's process group, so a Ctrl+C in that short time
//     can kill it before git runs. Git did nothing then, so a new run is
//     safe. The check cannot tell that case from another signal, so any
//     signal leads to a new run.
//  3. Otherwise, return what the last run returned.
func runShielded(ctx context.Context, dir string, args ...string) (string, error) {
	var out string
	var err error
	for attempt := 1; attempt <= 3; attempt++ {
		out, err = runGit(ctx, dir, true, args)
		if !killedBySignal(err) || ctx.Err() != nil {
			break
		}
	}
	return out, err
}

// runGit is the code of run and runShielded, in the steps of run. When
// shielded is true, it starts git in a process group of its own, with
// shield.
func runGit(ctx context.Context, dir string, shielded bool, args []string) (string, error) {

	var stdout, stderr bytes.Buffer
	var exitErr *exec.ExitError

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	cmd.Dir = dir
	cmd.WaitDelay = 5 * time.Second
	cmd.Cancel = interrupt(cmd)
	if shielded {
		shield(cmd)
	}

	err := cmd.Run()

	if errors.Is(err, exec.ErrWaitDelay) {
		return stdout.String(), nil
	}

	if err == nil {
		return stdout.String(), nil
	}

	if ctx.Err() != nil {
		return "", ctx.Err()
	}

	if errors.Is(err, exec.ErrNotFound) {
		return "", ErrNotFound
	}

	if errors.As(err, &exitErr) {
		return "", &Error{ExitCode: exitErr.ExitCode(), Stderr: stderr.String()}
	}

	return "", err

}
