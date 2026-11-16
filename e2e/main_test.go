// Package e2e tests sliceflow the way a user runs it: it builds the real
// program, runs it as its own process, and checks what it prints and its
// exit code.
package e2e

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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

// statSources looks at every file under root that the program is built
// from: the Go files and the templates. go test keeps a package's result
// until a file that one of its tests looked at changes. The program is
// built by another process, which go test does not watch, so without
// this, a change to the code could give an old result. It must be called
// from a test: go test does not watch TestMain.
func statSources(root string) error {
	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && strings.HasPrefix(entry.Name(), ".") && path != root {
			return filepath.SkipDir
		}
		if entry.IsDir() || strings.HasSuffix(path, ".go") || strings.HasSuffix(path, ".md") {
			_, err := os.Stat(path)
			return err
		}
		return nil
	})
}
