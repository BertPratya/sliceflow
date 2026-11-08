// Package e2e runs the real sliceflow program, and checks what it
// prints and its exit code (plan, Testing).
package e2e

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// program is the path of the sliceflow program that TestMain built.
var program string

// TestMain builds sliceflow once, into a temporary folder, runs the
// tests, and removes the folder.
func TestMain(m *testing.M) {
	os.Exit(buildAndRun(m))
}

// buildAndRun is TestMain's work, in its own function so that its
// deferred clean-up runs before os.Exit.
func buildAndRun(m *testing.M) int {
	dir, err := os.MkdirTemp("", "sliceflow-e2e-")
	if err != nil {
		fmt.Fprintln(os.Stderr, "e2e: make a temporary folder:", err)
		return 1
	}
	defer os.RemoveAll(dir)

	name := "sliceflow"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	program = filepath.Join(dir, name)

	// go test puts the go program of this Go first in PATH.
	build := exec.Command("go", "build", "-o", program, "github.com/BertPratya/sliceflow/cmd/sliceflow")
	if out, err := build.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "e2e: go build: %v\n%s", err, out)
		return 1
	}
	return m.Run()
}

// result is what one run of sliceflow printed, and its exit code.
type result struct {
	stdout string
	stderr string
	code   int
}

// runIn runs sliceflow with args in the folder dir, and returns what it
// printed and its exit code. It runs with a home folder of its own, so
// the user's files and git settings are not used: HOME (USERPROFILE on
// Windows) is a new temporary folder, git reads no system settings
// (GIT_CONFIG_NOSYSTEM=1), and git's global settings are a file that
// sets only a name and an email (GIT_CONFIG_GLOBAL).
func runIn(t *testing.T, dir string, args ...string) result {
	t.Helper()
	home := t.TempDir()
	gitConfig := filepath.Join(home, "gitconfig")
	settings := "[user]\n\tname = Test User\n\temail = test@example.com\n"
	if err := os.WriteFile(gitConfig, []byte(settings), 0o600); err != nil {
		t.Fatalf("write the git settings: %v", err)
	}

	cmd := exec.Command(program, args...)
	cmd.Dir = dir
	cmd.Env = append(cleanEnv(),
		"HOME="+home,
		"USERPROFILE="+home,
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_CONFIG_GLOBAL="+gitConfig,
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	code := 0
	if err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			t.Fatalf("run sliceflow %q: %v", args, err)
		}
		code = exitErr.ExitCode()
	}
	return result{stdout: stdout.String(), stderr: stderr.String(), code: code}
}

// cleanEnv returns the environment of the tests without the variables
// that runIn sets, and without git's own variables (GIT_*), so that no
// setting of the user's reaches sliceflow. Names are compared without
// case, because Windows does not tell Home from HOME.
func cleanEnv() []string {
	var env []string
	for _, kv := range os.Environ() {
		key, _, _ := strings.Cut(kv, "=")
		upper := strings.ToUpper(key)
		if upper == "HOME" || upper == "USERPROFILE" || upper == "XDG_CONFIG_HOME" || strings.HasPrefix(upper, "GIT_") {
			continue
		}
		env = append(env, kv)
	}
	return env
}

// check runs sliceflow with args in a new empty folder, and checks the
// whole normal output, the whole error output, and the exit code, word
// for word. It also checks that the folder is still empty: the command
// changed nothing.
func check(t *testing.T, args []string, want result) {
	t.Helper()
	dir := t.TempDir()
	got := runIn(t, dir, args...)
	if got.stdout != want.stdout {
		t.Errorf("sliceflow %q: normal output:\n%s\nwant:\n%s", args, got.stdout, want.stdout)
	}
	if got.stderr != want.stderr {
		t.Errorf("sliceflow %q: error output:\n%s\nwant:\n%s", args, got.stderr, want.stderr)
	}
	if got.code != want.code {
		t.Errorf("sliceflow %q: exit code %d, want %d", args, got.code, want.code)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read the folder: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("sliceflow %q: the folder is not empty: %d entries", args, len(entries))
	}
}
