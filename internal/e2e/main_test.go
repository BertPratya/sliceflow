// Package e2e runs the real sliceflow program and checks what it prints.
package e2e

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// binary is the path of the sliceflow program that TestMain builds.
var binary string

// TestMain builds sliceflow once, runs the tests, and removes the build.
func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

// testMain does the work of TestMain, so that the deferred clean-up runs
// before os.Exit.
func testMain(m *testing.M) int {
	dir, err := os.MkdirTemp("", "sliceflow-e2e")
	if err != nil {
		fmt.Fprintln(os.Stderr, "cannot make a temp folder:", err)
		return 1
	}
	defer os.RemoveAll(dir)

	name := "sliceflow"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	binary = filepath.Join(dir, name)

	build := exec.Command("go", "build", "-o", binary, "../../cmd/sliceflow")
	if out, err := build.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "go build failed: %v\n%s", err, out)
		return 1
	}
	return m.Run()
}

// result is what one run of sliceflow gave back.
type result struct {
	stdout, stderr string
	code           int
}

// run runs sliceflow with args in dir. It stops the test when the program
// cannot start; a non-zero exit code is not a failure here.
func run(t *testing.T, dir string, args ...string) result {
	t.Helper()
	cmd := exec.Command(binary, args...)
	cmd.Dir = dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	code := 0
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			t.Fatalf("cannot run sliceflow: %v", err)
		}
		code = exitErr.ExitCode()
	}
	return result{stdout: stdout.String(), stderr: stderr.String(), code: code}
}
