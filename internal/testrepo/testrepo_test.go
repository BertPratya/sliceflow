package testrepo_test

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/BertPratya/sliceflow/internal/testrepo"
)

func TestSetupClearsRepoVars(t *testing.T) {
	testrepo.Setup(t)
	decoy := t.TempDir()
	testrepo.Git(t, decoy, "init")
	gitDir := filepath.Join(decoy, ".git")
	t.Setenv("GIT_DIR", gitDir)
	t.Setenv("GIT_INDEX_FILE", filepath.Join(gitDir, "index"))

	dir := testrepo.New(t)

	if got := testrepo.Git(t, dir, "log", "--format=%s"); got != "first\n" {
		t.Fatalf("new repo log: got %q", got)
	}
	cmd := exec.Command("git", "--git-dir", gitDir, "rev-parse", "--verify", "HEAD")
	cmd.Env = append(cmd.Environ(), "GIT_INDEX_FILE=")
	if err := cmd.Run(); err == nil {
		t.Fatal("decoy repo has a commit")
	}
}
