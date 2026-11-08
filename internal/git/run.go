package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"
)

type Error struct {
	ExitCode int
	Stderr   string
}

func (e *Error) Error() string {
	return fmt.Sprintf("exit code %d: %s", e.ExitCode, e.Stderr)
}

var ErrNotFound = errors.New("git not found.")

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
