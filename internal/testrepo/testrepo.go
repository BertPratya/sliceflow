// Package testrepo makes temporary git repos for tests. Only test files
// import it.
package testrepo

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Setup removes the repo-location variables git sets when it runs a hook
// (the ones listed by "git rev-parse --local-env-vars"), so tests never touch
// the repo that runs them. It then points git at a temporary global config with a test name, email,
// and default branch, and turns off the system config.
func Setup(t testing.TB) {
	t.Helper()
	clearLocalEnv(t)
	cfg := filepath.Join(t.TempDir(), "gitconfig")
	content := "[user]\n\tname = Test User\n\temail = test@example.com\n[init]\n\tdefaultBranch = main\n"
	if err := os.WriteFile(cfg, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", cfg)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
}

// New makes a repo on main with one empty commit "first" and returns its
// folder.
func New(t testing.TB) string {
	t.Helper()
	Setup(t)
	dir := t.TempDir()
	Git(t, dir, "init")
	Git(t, dir, "commit", "--allow-empty", "-m", "first")
	return dir
}

// Git runs git in dir, fails the test on an error, and returns stdout.
func Git(t testing.TB, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, stderr.String())
	}
	return stdout.String()
}

// clearLocalEnv unsets every variable git lists as repo-local. The test
// restores them when it ends.
func clearLocalEnv(t testing.TB) {
	t.Helper()
	out, err := exec.Command("git", "rev-parse", "--local-env-vars").Output()
	if err != nil {
		t.Fatalf("git rev-parse --local-env-vars: %v", err)
	}
	for _, k := range strings.Fields(string(out)) {
		t.Setenv(k, "")
		os.Unsetenv(k)
	}
}
