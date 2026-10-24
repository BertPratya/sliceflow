package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BertPratya/sliceflow/internal/testrepo"
)

// wantFail runs the binary and checks the failure and that nothing was made.
func wantFail(t *testing.T, dir string, wantErr string, args ...string) {
	t.Helper()
	out, errOut, code := run(t, dir, args...)
	if code != 1 || out != "" || errOut != "Error: "+wantErr+"\n" {
		t.Fatalf("code %d\nstdout %q\nstderr %q\nwant   %q", code, out, errOut, wantErr)
	}
	if _, err := os.Lstat(filepath.Join(dir, ".sliceflow")); err == nil && !strings.Contains(wantErr, "folder .sliceflow") {
		t.Error(".sliceflow was made")
	}
	if b := testrepo.Git(t, dir, "branch", "--list", "slice/*"); strings.TrimSpace(b) != "" {
		t.Errorf("branch made: %q", b)
	}
	if w := strings.Count(testrepo.Git(t, dir, "worktree", "list"), "\n"); w != 1 {
		t.Errorf("worktrees: %d", w)
	}
}

const nameErr = `invalid slice name "%s": use 1-64 lowercase letters, digits, and dashes, starting with a letter or digit. Nothing was changed.`

func TestCreateBadName(t *testing.T) {
	dir := testrepo.New(t)
	wantFail(t, dir, strings.Replace(nameErr, "%s", "My Slice", 1), "slice", "create", "My Slice")
}

func TestCreateReservedName(t *testing.T) {
	dir := testrepo.New(t)
	wantFail(t, dir, `invalid slice name "con": it is a reserved name on Windows. Nothing was changed.`, "slice", "create", "con")
}

func TestCreateRemoteExists(t *testing.T) {
	dir := testrepo.New(t)
	testrepo.Git(t, dir, "remote", "add", "origin", "https://example.invalid/x.git")
	testrepo.Git(t, dir, "update-ref", "refs/remotes/origin/slice/spec-submit", "HEAD")
	wantFail(t, dir, "slice spec-submit already exists: branch origin/slice/spec-submit. Nothing was changed.", "slice", "create", "spec-submit")
}

func TestCreateBaseMissing(t *testing.T) {
	dir := testrepo.New(t)
	testrepo.Git(t, dir, "config", "sliceflow.base", "develop")
	wantFail(t, dir, "base branch develop does not exist (from git config sliceflow.base). Nothing was changed.", "slice", "create", "spec-submit")
}

func TestCreateFolderExists(t *testing.T) {
	dir := testrepo.New(t)
	if err := os.MkdirAll(filepath.Join(dir, ".sliceflow", "worktrees", "spec-submit"), 0o755); err != nil {
		t.Fatal(err)
	}
	wantFail(t, dir, "folder .sliceflow/worktrees/spec-submit already exists. Remove it, or pick another name. Nothing was changed.", "slice", "create", "spec-submit")
}

func TestCreateNotARepo(t *testing.T) {
	testrepo.Setup(t)
	dir := t.TempDir()
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(dir))
	out, errOut, code := run(t, dir, "slice", "create", "spec-submit")
	if code != 1 || out != "" || errOut != "Error: not inside a git repository. Nothing was changed.\n" {
		t.Fatalf("code %d out %q err %q", code, out, errOut)
	}
	if _, err := os.Lstat(filepath.Join(dir, ".sliceflow")); err == nil {
		t.Error(".sliceflow was made")
	}
}

func TestCreateNoCommits(t *testing.T) {
	testrepo.Setup(t)
	dir := t.TempDir()
	testrepo.Git(t, dir, "init")
	out, errOut, code := run(t, dir, "slice", "create", "spec-submit")
	if code != 1 || out != "" || errOut != "Error: the repository has no commits yet. Make a first commit, then try again. Nothing was changed.\n" {
		t.Fatalf("code %d out %q err %q", code, out, errOut)
	}
	if _, err := os.Lstat(filepath.Join(dir, ".sliceflow")); err == nil {
		t.Error(".sliceflow was made")
	}
}

func TestCreateBadNameJSON(t *testing.T) {
	dir := testrepo.New(t)
	wantFail(t, dir, strings.Replace(nameErr, "%s", "My-Slice", 1), "slice", "create", "My-Slice", "--json")
}
