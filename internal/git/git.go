package git

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ErrNotRepo means the folder is not inside a git repository.
var ErrNotRepo = errors.New("not inside a git repository")

// Root returns the top folder of the git repository that dir is in. It
// returns ErrNotRepo when dir is not inside a repository.
func Root(dir string) (string, error) {
	args := []string{"rev-parse", "--show-toplevel"}
	out, err := command(dir, args...).Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && strings.Contains(string(exitErr.Stderr), "not a git repository") {
			return "", ErrNotRepo
		}
		return "", gitError(args, err)
	}

	return strings.TrimSpace(string(out)), nil
}

// HasCommits reports whether the repository that dir is in has at least one
// commit. A new repository has none until its first commit.
func HasCommits(dir string) (bool, error) {
	// With --verify --quiet, git exits with 0 when HEAD points at a commit,
	// and with 1, printing nothing, when there is no commit yet.
	args := []string{"rev-parse", "--verify", "--quiet", "HEAD"}
	_, err := command(dir, args...).Output()
	if err == nil {
		return true, nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return false, nil
	}
	return false, gitError(args, err)
}

// command prepares git with args, to run in dir. LC_ALL=C makes git print
// its messages in English, so sliceflow can read them on every computer.
func command(dir string, args ...string) *exec.Cmd {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "LC_ALL=C")
	return cmd
}

// gitError turns an error from running git with args into the reason of
// R18: the first line of git's message that is not empty, or
// "git <command> failed (exit code <n>)" when git printed nothing.
func gitError(args []string, err error) error {
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		// git could not run at all, e.g. it is not installed.
		return fmt.Errorf("run git: %w", err)
	}
	if line := firstLine(string(exitErr.Stderr)); line != "" {
		return errors.New(line)
	}
	return fmt.Errorf("git %s failed (exit code %d)", args[0], exitErr.ExitCode())
}

// firstLine returns the first line of s that is not empty.
func firstLine(s string) string {
	for _, line := range strings.Split(s, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			return line
		}
	}
	return ""
}
