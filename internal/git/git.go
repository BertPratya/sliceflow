// Package git runs git commands as their own processes. Every later task
// runs git through Run.
package git

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// Error is a failed git command.
type Error struct {
	Args     []string
	ExitCode int
	Stderr   string // git's stderr, trimmed at both ends
}

// Error returns git's stderr text.
func (e *Error) Error() string { return e.Stderr }

var (
	pathOnce sync.Once
	gitPath  string
	pathErr  error
)

func findGit() (string, error) {
	pathOnce.Do(func() {
		gitPath, pathErr = exec.LookPath("git")
	})
	return gitPath, pathErr
}

// Run runs one git command in dir and returns its stdout. A failed command
// returns a *Error. When ctx is done, it returns ctx.Err().
func Run(ctx context.Context, dir string, args ...string) (string, error) {
	path, err := findGit()
	if err != nil {
		return "", err
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	cmd := exec.CommandContext(ctx, path, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0", "GIT_NO_LAZY_FETCH=1")
	cmd.WaitDelay = 2 * time.Second
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()
	if runErr == nil {
		return stdout.String(), nil
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	var exitErr *exec.ExitError
	if errors.As(runErr, &exitErr) {
		return "", &Error{
			Args:     append([]string(nil), args...),
			ExitCode: exitErr.ExitCode(),
			Stderr:   strings.TrimSpace(stderr.String()),
		}
	}
	return "", runErr
}

// RefExists reports whether ref names a commit.
func RefExists(ctx context.Context, dir, ref string) (bool, error) {
	_, err := Run(ctx, dir, "rev-parse", "--verify", "--quiet", ref+"^{commit}")
	if err == nil {
		return true, nil
	}
	var ge *Error
	if errors.As(err, &ge) && ge.ExitCode == 1 {
		return false, nil
	}
	return false, err
}

// Show returns the file at path in the commit ref. The bool is false when
// the file is not in that commit. A bad ref or a folder that is not a repo
// also returns false, nil, so callers check those first.
func Show(ctx context.Context, dir, ref, path string) ([]byte, bool, error) {
	spec := ref + ":" + path
	if _, err := Run(ctx, dir, "cat-file", "-e", spec); err != nil {
		var ge *Error
		if errors.As(err, &ge) {
			return nil, false, nil
		}
		return nil, false, err
	}
	out, err := Run(ctx, dir, "show", spec)
	if err != nil {
		return nil, false, err
	}
	return []byte(out), true, nil
}
