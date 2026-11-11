// Package e2e runs the real sliceflow program, and checks what it
// prints and its exit code (plan, Testing).
package e2e

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/BertPratya/sliceflow/internal/gittest"
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
	return runCmd(t, cmd)
}

// createOpts are the choices of createCmd.
type createOpts struct {
	// NoIdentity gives git no name and no email, and sets
	// user.useConfigOnly=true, so git cannot make a commit on any
	// machine (TD8).
	NoIdentity bool
	// Config holds more settings of the global git config, by key, e.g.
	// "sliceflow.worktree.dir".
	Config map[string]string
	// Path is the PATH of the program; "" means the normal PATH.
	Path string
}

// createCmd prepares sliceflow with args, in the folder dir, and does
// not start it. It returns the command and its home folder.
//
// Steps:
//  1. Make a new home folder, with filepath.EvalSymlinks, so it is the
//     real path that git also gives (/private/var on macOS).
//  2. Write the global git config in it, with `git config --file`: a
//     name and an email, or, with opts.NoIdentity, only
//     user.useConfigOnly=true; gc.auto=0 and maintenance.auto=false, so
//     no git is left running in the background; then each setting of
//     opts.Config, in the order of their keys.
//  3. The environment is cleanEnv, with HOME and USERPROFILE set to the
//     home folder, GIT_CONFIG_NOSYSTEM=1, and GIT_CONFIG_GLOBAL set to
//     the config. With opts.NoIdentity, EMAIL is removed too. When
//     opts.Path is not "", PATH is opts.Path.
//  4. The outputs go to two new bytes.Buffer, which runCmd reads.
//
// A test may set SysProcAttr, and call Start itself, before runCmd.
func createCmd(t *testing.T, dir string, opts createOpts, args ...string) (cmd *exec.Cmd, home string) {
	t.Helper()
	home, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("find the real path of the home folder: %v", err)
	}

	gitConfig := filepath.Join(home, "gitconfig")
	settings := [][2]string{{"gc.auto", "0"}, {"maintenance.auto", "false"}}
	if opts.NoIdentity {
		settings = append(settings, [2]string{"user.useConfigOnly", "true"})
	} else {
		settings = append(settings, [2]string{"user.name", "Test User"}, [2]string{"user.email", "test@example.com"})
	}
	for _, key := range slices.Sorted(maps.Keys(opts.Config)) {
		settings = append(settings, [2]string{key, opts.Config[key]})
	}
	for _, kv := range settings {
		gittest.Git(t, home, "config", "--file", gitConfig, kv[0], kv[1])
	}

	var env []string
	for _, kv := range cleanEnv() {
		key, _, _ := strings.Cut(kv, "=")
		upper := strings.ToUpper(key)
		if (opts.NoIdentity && upper == "EMAIL") || (opts.Path != "" && upper == "PATH") {
			continue
		}
		env = append(env, kv)
	}
	env = append(env,
		"HOME="+home,
		"USERPROFILE="+home,
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_CONFIG_GLOBAL="+gitConfig,
	)
	if opts.Path != "" {
		env = append(env, "PATH="+opts.Path)
	}

	cmd = exec.Command(program, args...)
	cmd.Dir = dir
	cmd.Env = env
	cmd.Stdout = &bytes.Buffer{}
	cmd.Stderr = &bytes.Buffer{}
	return cmd, home
}

// runCmd starts cmd, unless the test has started it already, waits for
// it, and returns what it printed and its exit code. Each output is read
// when it is a *bytes.Buffer; when cmd.Stdout or cmd.Stderr is nil, it
// gets a new one first. The exit code is -1 when a signal ended it.
func runCmd(t *testing.T, cmd *exec.Cmd) result {
	t.Helper()
	if cmd.Process == nil {
		if cmd.Stdout == nil {
			cmd.Stdout = &bytes.Buffer{}
		}
		if cmd.Stderr == nil {
			cmd.Stderr = &bytes.Buffer{}
		}
		if err := cmd.Start(); err != nil {
			t.Fatalf("start sliceflow %q: %v", cmd.Args[1:], err)
		}
	}

	err := cmd.Wait()
	code := 0
	if err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			t.Fatalf("run sliceflow %q: %v", cmd.Args[1:], err)
		}
		code = exitErr.ExitCode()
	}
	return result{stdout: bufferText(cmd.Stdout), stderr: bufferText(cmd.Stderr), code: code}
}

// bufferText returns the text of out when it is a *bytes.Buffer, and ""
// otherwise.
func bufferText(out io.Writer) string {
	if b, ok := out.(*bytes.Buffer); ok {
		return b.String()
	}
	return ""
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
