// Package gittest makes temporary git repos for tests (plan TD6).
//
// It runs git with os/exec itself, not with the git package of
// sliceflow, so a bug in that package cannot break the setup of its own
// tests. Every helper fails the test at once when git fails, and shows
// everything git printed.
//
// The helpers change environment variables with t.Setenv, so a test
// that uses them cannot call t.Parallel: Go would stop the test.
package gittest

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The name and the email that the test's global git config sets, so git
// can make commits.
const (
	UserName  = "Test User"
	UserEmail = "test@example.com"
)

// leakedVars are git's variables that change what git does: which repo
// it uses, who makes a commit and when, and what it traces. git sets
// some of them when it runs a hook, so `go test` started from a hook
// would see them. Isolate removes them for the test.
var leakedVars = []string{
	"GIT_DIR",
	"GIT_WORK_TREE",
	"GIT_COMMON_DIR",
	"GIT_INDEX_FILE",
	"GIT_OBJECT_DIRECTORY",
	"GIT_ALTERNATE_OBJECT_DIRECTORIES",
	"GIT_CEILING_DIRECTORIES",
	"GIT_CONFIG",
	"GIT_CONFIG_PARAMETERS",
	"GIT_CONFIG_COUNT",
	"GIT_AUTHOR_NAME",
	"GIT_AUTHOR_EMAIL",
	"GIT_AUTHOR_DATE",
	"GIT_COMMITTER_NAME",
	"GIT_COMMITTER_EMAIL",
	"GIT_COMMITTER_DATE",
	"GIT_NAMESPACE",
	"GIT_TRACE",
	"GIT_TRACE2",
	"GIT_TRACE2_EVENT",
	"GIT_TRACE2_PERF",
}

// globalConfig is the test's global git config. It sets a name and an
// email, so git can make commits, and turns off git's work in the
// background (gc.auto, maintenance.auto): a git left running could hold
// a temporary folder open, and Windows cannot delete an open folder.
const globalConfig = "[user]\n" +
	"\tname = " + UserName + "\n" +
	"\temail = " + UserEmail + "\n" +
	"[gc]\n\tauto = 0\n" +
	"[maintenance]\n\tauto = false\n"

// Isolate keeps git, for this test, away from the settings of the user's
// machine, and returns the test's home folder.
//
// Steps:
//  1. Make a new home folder, and set HOME and USERPROFILE to it.
//     os.UserHomeDir reads HOME on Linux and macOS, and USERPROFILE on
//     Windows.
//  2. Write globalConfig in a temporary file, and point
//     GIT_CONFIG_GLOBAL at it.
//  3. Set GIT_CONFIG_NOSYSTEM=1, so git reads no system config.
//  4. Set LC_ALL=C, so git prints its messages in English.
//  5. Remove the variables in leakedVars.
//
// Each call makes a new home folder and a new global config.
func Isolate(t testing.TB) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	global := filepath.Join(t.TempDir(), "gitconfig")
	if err := os.WriteFile(global, []byte(globalConfig), 0o644); err != nil {
		t.Fatalf("write the global git config: %v", err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", global)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("LC_ALL", "C")

	for _, key := range leakedVars {
		unsetenv(t, key)
	}
	return home
}

// unsetenv removes the variable key for this test. t.Setenv first saves
// the old value, so Go puts it back when the test ends.
func unsetenv(t testing.TB, key string) {
	t.Helper()
	t.Setenv(key, "")
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("unset %s: %v", key, err)
	}
}

// Git runs git with args in the folder dir, and returns what git printed
// to its normal output, without the line break at its end. It fails the
// test when git fails.
func Git(t testing.TB, dir string, args ...string) string {
	t.Helper()
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("git %s in %s: %v\nstdout:\n%s\nstderr:\n%s",
			strings.Join(args, " "), dir, err, stdout.String(), stderr.String())
	}
	return strings.TrimRight(stdout.String(), "\r\n")
}

// SetGlobalConfig sets key to value in the test's global git config.
// Call Isolate (or New) first, or it changes the user's own config.
func SetGlobalConfig(t testing.TB, key, value string) {
	t.Helper()
	if os.Getenv("GIT_CONFIG_GLOBAL") == "" {
		t.Fatalf("SetGlobalConfig: call Isolate first")
	}
	Git(t, t.TempDir(), "config", "--global", key, value)
}

// BreakGlobalConfig adds a line that git cannot read to the test's
// global config, so every git command fails. Call Isolate (or New)
// first.
func BreakGlobalConfig(t testing.TB) {
	t.Helper()
	path := os.Getenv("GIT_CONFIG_GLOBAL")
	if path == "" {
		t.Fatalf("BreakGlobalConfig: call Isolate first")
	}
	appendLine(t, path, "[broken")
}

// appendLine adds line, and a line break, at the end of the file path.
func appendLine(t testing.TB, path, line string) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	if _, err := f.WriteString(line + "\n"); err != nil {
		f.Close()
		t.Fatalf("write %s: %v", path, err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close %s: %v", path, err)
	}
}

// Repo is a git repo in a temporary folder.
type Repo struct {
	t testing.TB
	// Dir is the full path of the repo's folder.
	Dir string
	// Home is the test's home folder, from Isolate.
	Home string
}

// NewEmpty calls Isolate, then makes a new repo with the branch main and
// no commits.
func NewEmpty(t testing.TB) *Repo {
	t.Helper()
	home := Isolate(t)
	dir := t.TempDir()
	Git(t, dir, "init", "--quiet", "-b", "main")
	return &Repo{t: t, Dir: dir, Home: home}
}

// New is NewEmpty with a first commit on main, which adds README.md.
func New(t testing.TB) *Repo {
	t.Helper()
	r := NewEmpty(t)
	r.Commit("README.md", "# test\n")
	return r
}

// Git runs git with args in the repo's folder, as the function Git does.
func (r *Repo) Git(args ...string) string {
	r.t.Helper()
	return Git(r.t, r.Dir, args...)
}

// WriteFile writes content to the file at path, a path inside the repo
// with "/" between its parts, and makes the folders it needs. It does
// not add the file to git.
func (r *Repo) WriteFile(path, content string) {
	r.t.Helper()
	full := filepath.Join(r.Dir, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		r.t.Fatalf("make the folders of %s: %v", path, err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		r.t.Fatalf("write %s: %v", path, err)
	}
}

// Commit writes the file at path (as WriteFile does), commits it on the
// branch that is checked out, and returns the new commit's hash.
func (r *Repo) Commit(path, content string) string {
	r.t.Helper()
	r.WriteFile(path, content)
	r.Git("add", "--", path)
	r.Git("commit", "--quiet", "-m", "add "+path)
	return r.Git("rev-parse", "HEAD")
}

// Branch makes the branch name at the commit that is checked out, and
// stays on the branch it was on.
func (r *Repo) Branch(name string) {
	r.t.Helper()
	r.Git("branch", "--", name)
}

// AddWorktree makes a new branch at the commit that is checked out, and
// a linked worktree for it in a new temporary folder. It returns the
// worktree's full path.
func (r *Repo) AddWorktree(branch string) string {
	r.t.Helper()
	path := filepath.Join(r.t.TempDir(), "worktree")
	r.Git("worktree", "add", "--quiet", "-b", branch, path)
	return path
}

// SetConfig sets key to value in the repo's own git config.
func (r *Repo) SetConfig(key, value string) {
	r.t.Helper()
	r.Git("config", key, value)
}

// BreakConfig adds a line that git cannot read to the repo's own config
// file, .git/config, so git fails in the repo.
func (r *Repo) BreakConfig() {
	r.t.Helper()
	appendLine(r.t, filepath.Join(r.Dir, ".git", "config"), "[broken")
}
