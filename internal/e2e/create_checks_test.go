package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BertPratya/sliceflow/internal/gittest"
)

// wantFail checks that a run failed the way R11 says: the exact message on
// stderr, nothing on stdout, and exit code 1.
func wantFail(t *testing.T, r result, msg string) {
	t.Helper()
	if r.code != 1 {
		t.Errorf("code = %d, want 1", r.code)
	}
	if r.stdout != "" {
		t.Errorf("stdout = %q, want empty", r.stdout)
	}
	if want := msg + "\n"; r.stderr != want {
		t.Errorf("stderr = %q, want %q", r.stderr, want)
	}
}

// makeDir makes the folder path (with / between the parts) in repo.
func makeDir(t *testing.T, repo, path string) string {
	t.Helper()
	dir := filepath.Join(repo, filepath.FromSlash(path))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

// TestS5: a name with a space gives the shape error.
func TestS5(t *testing.T) {
	repo := gittest.NewRepo(t)
	r := run(t, repo, "slice", "create", "My Slice")
	wantFail(t, r, `Error: invalid slice name "My Slice": use 1-64 lowercase letters, digits, and dashes, starting with a letter or digit. Nothing was changed.`)
}

// TestS6: a Windows-reserved name gives the reserved error.
func TestS6(t *testing.T) {
	repo := gittest.NewRepo(t)
	r := run(t, repo, "slice", "create", "con")
	wantFail(t, r, `Error: invalid slice name "con": it is a reserved name on Windows. Nothing was changed.`)
}

// TestS7: a teammate's branch, which the last fetch brought, blocks the name.
// The test pushes to a bare repo; sliceflow itself never does.
func TestS7(t *testing.T) {
	repo := gittest.NewRepo(t)
	bare := t.TempDir()
	gittest.Run(t, bare, "init", "--bare")
	gittest.Run(t, repo, "remote", "add", "origin", bare)
	gittest.Run(t, repo, "branch", "slice/spec-submit")
	gittest.Run(t, repo, "push", "origin", "slice/spec-submit")
	gittest.Run(t, repo, "branch", "-D", "slice/spec-submit")

	r := run(t, repo, "slice", "create", "spec-submit")
	wantFail(t, r, "Error: slice spec-submit already exists: branch origin/slice/spec-submit. Nothing was changed.")
}

// TestS8: git config sliceflow.base names a branch that does not exist.
func TestS8(t *testing.T) {
	repo := gittest.NewRepo(t)
	gittest.Run(t, repo, "config", "sliceflow.base", "develop")
	r := run(t, repo, "slice", "create", "spec-submit")
	wantFail(t, r, "Error: base branch develop does not exist (from git config sliceflow.base). Nothing was changed.")
}

// TestS9: an old worktree folder was left behind.
func TestS9(t *testing.T) {
	repo := gittest.NewRepo(t)
	makeDir(t, repo, ".sliceflow/worktrees/spec-submit")
	r := run(t, repo, "slice", "create", "spec-submit")
	wantFail(t, r, "Error: folder .sliceflow/worktrees/spec-submit already exists. Remove it, or pick another name. Nothing was changed.")
}

// TestS10: not inside a git repository.
func TestS10(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(dir))
	r := run(t, dir, "slice", "create", "spec-submit")
	wantFail(t, r, "Error: not inside a git repository. Nothing was changed.")
}

// TestS11: a repository with no commits.
func TestS11(t *testing.T) {
	repo := gittest.Init(t)
	r := run(t, repo, "slice", "create", "spec-submit")
	wantFail(t, r, "Error: the repository has no commits yet. Make a first commit, then try again. Nothing was changed.")
}

// TestCreateLocalBranchExists: R15, first case.
func TestCreateLocalBranchExists(t *testing.T) {
	repo := gittest.NewRepo(t)
	gittest.Run(t, repo, "branch", "slice/spec-submit")
	r := run(t, repo, "slice", "create", "spec-submit")
	wantFail(t, r, "Error: slice spec-submit already exists: branch slice/spec-submit. Nothing was changed.")
}

// TestCreateFolderOnBase: R15, third case.
func TestCreateFolderOnBase(t *testing.T) {
	repo := gittest.NewRepo(t)
	dir := makeDir(t, repo, "docs/slices/spec-submit")
	if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	gittest.Run(t, repo, "add", ".")
	gittest.Run(t, repo, "commit", "-m", "add spec-submit")

	r := run(t, repo, "slice", "create", "spec-submit")
	wantFail(t, r, "Error: slice spec-submit already exists: folder docs/slices/spec-submit on main. Nothing was changed.")
}

// TestCreateFromMissing: --from names a branch that does not exist.
func TestCreateFromMissing(t *testing.T) {
	repo := gittest.NewRepo(t)
	r := run(t, repo, "slice", "create", "spec-submit", "--from", "develop")
	wantFail(t, r, "Error: base branch develop does not exist (from --from). Nothing was changed.")
}

// TestCreateOrderNotRepoBeforeName: R12 comes before R14.
func TestCreateOrderNotRepoBeforeName(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(dir))
	r := run(t, dir, "slice", "create", "My Slice")
	wantFail(t, r, "Error: not inside a git repository. Nothing was changed.")
}

// TestCreateOrderBaseBeforeExists: R16 comes before R15.
func TestCreateOrderBaseBeforeExists(t *testing.T) {
	repo := gittest.NewRepo(t)
	gittest.Run(t, repo, "branch", "slice/spec-submit")
	r := run(t, repo, "slice", "create", "spec-submit", "--from", "develop")
	wantFail(t, r, "Error: base branch develop does not exist (from --from). Nothing was changed.")
}

// TestCreateOldFolderFromSubfolder: R17 looks in the main checkout, also
// when the command runs in a subfolder of it.
func TestCreateOldFolderFromSubfolder(t *testing.T) {
	repo := gittest.NewRepo(t)
	makeDir(t, repo, ".sliceflow/worktrees/x")
	sub := makeDir(t, repo, "sub/deeper")

	r := run(t, sub, "slice", "create", "x")
	wantFail(t, r, "Error: folder .sliceflow/worktrees/x already exists. Remove it, or pick another name. Nothing was changed.")
}

// TestCreateOldFolderFromWorktree: R17 looks in the main checkout, also
// when the command runs in a linked worktree.
func TestCreateOldFolderFromWorktree(t *testing.T) {
	repo := gittest.NewRepo(t)
	makeDir(t, repo, ".sliceflow/worktrees/x")
	wt := filepath.Join(t.TempDir(), "wt")
	gittest.Run(t, repo, "worktree", "add", "-b", "other", wt)

	r := run(t, wt, "slice", "create", "x")
	wantFail(t, r, "Error: folder .sliceflow/worktrees/x already exists. Remove it, or pick another name. Nothing was changed.")
}
