package e2e

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// watchSources makes the first run of sliceflow look at its source
// files, so go test runs the tests again when the code changes.
var watchSources sync.Once

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
	return runSliceflowWith(t, dir, home, nil, args...)
}

// runSliceflowWith is runSliceflow with more environment variables, e.g.
// "PATH=" + an empty folder. A variable in env replaces one of the same
// name.
func runSliceflowWith(t *testing.T, dir, home string, env []string, args ...string) result {
	t.Helper()
	watchSources.Do(func() {
		if err := statSources(".."); err != nil {
			t.Fatalf("could not look at the source files: %v", err)
		}
	})
	cmd := exec.Command(program, args...)
	cmd.Dir = dir
	cmd.Env = append(testEnv(home), env...)
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

// testEnv returns the environment of a program that a test runs: the
// test runner's own, without its GIT_ variables, with home as the home
// folder, and with no system git config.
func testEnv(home string) []string {
	return append(withoutGitVariables(os.Environ()),
		"HOME="+home,
		"USERPROFILE="+home,
		"XDG_CONFIG_HOME="+filepath.Join(home, ".config"),
		"GIT_CONFIG_NOSYSTEM=1",
	)
}

// newProject makes a git repo called shop in a new temp folder, with one
// commit on main, and returns its folder. Git runs with home as the home
// folder, so it uses the test's git config.
func newProject(t *testing.T, home string) string {
	t.Helper()
	repo := filepath.Join(t.TempDir(), "shop")
	if err := os.Mkdir(repo, 0o755); err != nil {
		t.Fatalf("could not make the project folder: %v", err)
	}
	gitOutput(t, repo, home, "init", "--quiet")
	gitOutput(t, repo, home, "commit", "--quiet", "--allow-empty", "-m", "first")
	return repo
}

// gitOutput runs git with args in the folder dir, with home as the home
// folder, and returns its output without the line break at the end. It
// fails the test now if git fails.
func gitOutput(t *testing.T, dir, home string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = testEnv(home)
	out, err := cmd.Output()
	if err != nil {
		var stderr []byte
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			stderr = exitErr.Stderr
		}
		t.Fatalf("git %q in %s: %v\n%s", args, dir, err, stderr)
	}
	return strings.TrimSuffix(string(out), "\n")
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
