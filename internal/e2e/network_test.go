package e2e

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/BertPratya/sliceflow/internal/gittest"
)

// networkTrace matches a line of GIT_TRACE that shows git running a command
// that talks to a remote. It looks at the command word right after "git",
// so a branch named "push" does not match.
var networkTrace = regexp.MustCompile(`trace: (built-in|run_command|exec): '?git'? '?(fetch|push|pull|ls-remote|clone|remote'? '?update)\b`)

// TestNoNetwork: R40. The repo has a remote that cannot be reached, and a
// remote-tracking branch. create, status, and list still work, and git's
// own trace shows that no command for the network was run.
func TestNoNetwork(t *testing.T) {
	repo := gittest.NewRepo(t)
	gittest.Run(t, repo, "remote", "add", "origin", filepath.Join(t.TempDir(), "does-not-exist"))
	gittest.Run(t, repo, "update-ref", "refs/remotes/origin/main", "refs/heads/main")
	gittest.Run(t, repo, "branch", "push")

	// git appends every command it runs, also the child ones, to this file.
	trace := filepath.Join(t.TempDir(), "trace")
	t.Setenv("GIT_TRACE", trace)

	if r := run(t, repo, "slice", "create", "spec-submit"); r.code != 0 || r.stderr != "" {
		t.Fatalf("slice create: code = %d, stderr = %q, want 0, empty", r.code, r.stderr)
	}
	if r := run(t, repo, "slice", "status", "spec-submit"); r.code != 0 || r.stderr != "" {
		t.Errorf("slice status: code = %d, stderr = %q, want 0, empty", r.code, r.stderr)
	}
	if r := run(t, repo, "slice", "list"); r.code != 0 || r.stderr != "" {
		t.Errorf("slice list: code = %d, stderr = %q, want 0, empty", r.code, r.stderr)
	}

	content, err := os.ReadFile(trace)
	if err != nil {
		t.Fatal(err)
	}
	if len(content) == 0 {
		t.Fatal("the git trace is empty: GIT_TRACE did not work")
	}
	if m := networkTrace.Find(content); m != nil {
		t.Errorf("git ran a command for the network: %q", m)
	}
}
