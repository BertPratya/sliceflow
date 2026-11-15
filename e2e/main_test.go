// Package e2e tests sliceflow the way a user runs it: it builds the real
// program, runs it as its own process, and checks what it prints and its
// exit code.
package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// program is the path to the sliceflow program that TestMain built.
var program string

// TestMain builds sliceflow once for all the tests in this package, runs
// them, then deletes the built program. When the build fails, it prints
// go build's output and runs no tests.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "sliceflow-e2e")
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not make a folder for the program:", err)
		os.Exit(1)
	}

	name := "sliceflow"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	program = filepath.Join(dir, name)

	build := exec.Command("go", "build", "-o", program, "github.com/BertPratya/sliceflow/cmd/sliceflow")
	if out, err := build.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "could not build sliceflow: %v\n%s", err, out)
		os.RemoveAll(dir)
		os.Exit(1)
	}

	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}
