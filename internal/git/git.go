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
	out, err := command(dir, "rev-parse", "--show-toplevel").Output()
	if err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			// git could not run at all, e.g. it is not installed.
			return "", fmt.Errorf("run git: %w", err)
		}

		msg := string(exitErr.Stderr)
		if strings.Contains(msg, "not a git repository") {
			return "", ErrNotRepo
		}

		// git ran and said no for another reason, e.g. "detected dubious
		// ownership": pass git's own message on.
		return "", fmt.Errorf("git rev-parse: %s", firstLine(msg))
	}

	return strings.TrimSpace(string(out)), nil
}

// command prepares git with args, to run in dir. LC_ALL=C makes git print
// its messages in English, so sliceflow can read them on every computer.
func command(dir string, args ...string) *exec.Cmd {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "LC_ALL=C")
	return cmd
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
