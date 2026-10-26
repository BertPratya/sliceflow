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

// Config returns the value of the git setting key, as git config --get
// prints it. ok is false when the setting is not set.
func Config(dir, key string) (value string, ok bool, err error) {
	// git config --get exits with 1 when the key is not set.
	args := []string{"config", "--get", key}
	out, err := command(dir, args...).Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return "", false, nil
		}
		return "", false, gitError(args, err)
	}
	return strings.TrimSpace(string(out)), true, nil
}

// BranchExists reports whether the repository that dir is in has a local
// branch with this name. A branch of a remote, like origin/main, does not
// count.
func BranchExists(dir, branch string) (bool, error) {
	args := []string{"rev-parse", "--verify", "--quiet", "refs/heads/" + branch}
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
	for line := range strings.SplitSeq(s, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			return line
		}
	}
	return ""
}
