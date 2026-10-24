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

	"github.com/BertPratya/sliceflow/internal/testrepo"
)

var binPath string

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func testMain(m *testing.M) int {
	tmp, err := os.MkdirTemp("", "sliceflow-e2e")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer os.RemoveAll(tmp)
	name := "sliceflow"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	binPath = filepath.Join(tmp, name)
	build := exec.Command("go", "build", "-buildvcs=false", "-o", binPath, "./cmd/sliceflow")
	build.Dir = filepath.Join("..", "..")
	if out, err := build.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "go build: %v\n%s", err, out)
		return 1
	}
	return m.Run()
}

// run runs the sliceflow binary in dir. It calls testrepo.Setup first, so the
// binary never sees git's hook variables and uses the test git config.
func run(t *testing.T, dir string, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	testrepo.Setup(t)
	cmd := exec.Command(binPath, args...)
	cmd.Dir = dir
	var o, e bytes.Buffer
	cmd.Stdout = &o
	cmd.Stderr = &e
	err := cmd.Run()
	var ee *exec.ExitError
	switch {
	case err == nil:
	case errors.As(err, &ee):
		code = ee.ExitCode()
	default:
		t.Fatalf("run %v: %v", args, err)
	}
	return o.String(), e.String(), code
}
