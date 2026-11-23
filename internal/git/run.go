package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
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
	cmd.Cancel = func() error { return stop(cmd.Process) }
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

// stop asks the process p to end, as Ctrl+C does, so that git can remove
// its lock files, e.g. .git/index.lock, before it ends. A git that is
// killed leaves them, and then every git command in the repo fails until
// the user deletes them. Where Ctrl+C cannot be sent, e.g. on Windows, it
// kills p. When p does not end within the command's WaitDelay, it is
// killed.
func stop(p *os.Process) error {
	if err := p.Signal(os.Interrupt); err != nil {
		return p.Kill()
	}
	return nil
}
