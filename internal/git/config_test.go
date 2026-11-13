package git

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// The setting that the tests of Config read.
const worktreeSetting = "sliceflow.worktree.dir"

// Test 1: a setting that is not set gives "".
func TestConfigNotSet(t *testing.T) {
	useTestConfig(t)
	repo := newRepo(t)
	wantConfig(t, repo, "")
}

// Test 2: a setting set for the repo gives its value.
func TestConfigSetForTheRepo(t *testing.T) {
	useTestConfig(t)
	repo := newRepo(t)
	gitIn(t, repo, "config", worktreeSetting, "../wt")
	wantConfig(t, repo, "../wt")
}

// Test 3: a setting set only globally gives its value.
func TestConfigSetGlobally(t *testing.T) {
	useTestConfig(t)
	repo := newRepo(t)
	gitIn(t, repo, "config", "--global", worktreeSetting, "~/wt")
	wantConfig(t, repo, "~/wt")
}

// Test 4: when the setting is set both for the repo and globally, the
// repo's value wins (R18).
func TestConfigRepoWinsOverGlobal(t *testing.T) {
	useTestConfig(t)
	repo := newRepo(t)
	gitIn(t, repo, "config", "--global", worktreeSetting, "~/wt")
	gitIn(t, repo, "config", worktreeSetting, "../wt")
	wantConfig(t, repo, "../wt")
}

// Test 5: a setting set to "" gives "", the same as not set (R18). Git
// prints an empty line for it, and exits with 0.
func TestConfigSetToEmpty(t *testing.T) {
	useTestConfig(t)
	repo := newRepo(t)
	gitIn(t, repo, "config", worktreeSetting, "")
	wantConfig(t, repo, "")
}

// Test 6: a broken config file gives "" and git's error as it is: a
// *Error with exit code 128.
func TestConfigBrokenFile(t *testing.T) {
	useTestConfig(t)
	repo := newRepo(t)
	config := filepath.Join(repo, ".git", "config")
	if err := os.WriteFile(config, []byte("not [ a config\n"), 0o644); err != nil {
		t.Fatalf("could not break the config file: %v", err)
	}

	got, err := Config(context.Background(), repo, worktreeSetting)
	if got != "" {
		t.Errorf("Config returned %q with its error, want \"\"", got)
	}
	gitErr := wantGitError(t, err)
	if gitErr.ExitCode != 128 {
		t.Errorf("the exit code is %d, want 128", gitErr.ExitCode)
	}
}

// useTestConfig makes git read a new, empty global config for the rest
// of the test, and no system config, so the user's own settings cannot
// change the result. git --global then writes to that new config.
func useTestConfig(t *testing.T) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
}

// wantConfig fails the test unless Config gives want for the worktree
// setting in repo, with no error.
func wantConfig(t *testing.T, repo, want string) {
	t.Helper()
	got, err := Config(context.Background(), repo, worktreeSetting)
	if err != nil {
		t.Fatalf("Config returned the error %v, want none", err)
	}
	if got != want {
		t.Errorf("Config returned %q, want %q", got, want)
	}
}
