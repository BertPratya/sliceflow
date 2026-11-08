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
//  2. If git exited with 0, return the normal output and no error. This
//     also holds when the 5 seconds ran out (exec.ErrWaitDelay): git
//     itself exited with 0.
//  3. If ctx ended (Ctrl+C, or a time limit), return ctx.Err(). This
//     comes before step 5: Go kills git when ctx ends, and a killed git
//     looks like a git that failed.
//  4. If git was not found, return ErrNotFound.
//  5. If git exited with another code, return an *Error with that code
//     and git's error output.
//  6. Otherwise, return the error as it is.
//
// On an error, the normal output is "".
func run(ctx context.Context, dir string, args ...string) (string, error) {

	var stdout, stderr bytes.Buffer
	var exitErr *exec.ExitError

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	cmd.Dir = dir
	cmd.WaitDelay = 5 * time.Second

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
