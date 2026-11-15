package e2e

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// result is what one run of sliceflow printed, and its exit code.
type result struct {
	stdout string
	stderr string
	code   int
}

// gitConfig is the global git config of every test's home folder. It
// gives git a name and an email, so commits work on any machine, and
// makes main the first branch of a new repo.
const gitConfig = `[user]
	name = Test User
	email = test@example.com
[init]
	defaultBranch = main
`

// newHome returns a new, empty home folder for one test, with only
// gitConfig in it. The folder is deleted when the test ends.
func newHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	if err := os.WriteFile(filepath.Join(home, ".gitconfig"), []byte(gitConfig), 0o644); err != nil {
		t.Fatalf("could not write the test's git config: %v", err)
	}
	return home
}

// runSliceflow runs sliceflow with args in the folder dir, and returns
// what it printed and its exit code. Git sees home as the home folder,
// and reads no system config and none of the GIT_ variables of the shell
// that runs the tests, so the result is the same on every machine. An
// exit code other than 0 is a result, not a failure; the test fails only
// when sliceflow cannot be started.
func runSliceflow(t *testing.T, dir, home string, args ...string) result {
	t.Helper()
	cmd := exec.Command(program, args...)
	cmd.Dir = dir
	cmd.Env = append(withoutGitVariables(os.Environ()),
		"HOME="+home,
		"USERPROFILE="+home,
		"XDG_CONFIG_HOME="+filepath.Join(home, ".config"),
		"GIT_CONFIG_NOSYSTEM=1",
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	code := 0
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		code = exitErr.ExitCode()
	} else if err != nil {
		t.Fatalf("could not run sliceflow %q: %v", args, err)
	}
	return result{stdout: stdout.String(), stderr: stderr.String(), code: code}
}

// withoutGitVariables returns env, a list of NAME=value entries, without
// the ones whose name starts with GIT_. The check ignores case, because
// names do on Windows.
func withoutGitVariables(env []string) []string {
	var kept []string
	for _, entry := range env {
		if !strings.HasPrefix(strings.ToUpper(entry), "GIT_") {
			kept = append(kept, entry)
		}
	}
	return kept
}
