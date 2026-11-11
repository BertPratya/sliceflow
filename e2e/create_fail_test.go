package e2e

// The e2e tests of a slice that cannot be made: a git step fails, the
// undo runs, and what is left is reported (R36, R37, R41). The tests
// that need a process group or a fake git are in
// create_fail_unix_test.go.

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/BertPratya/sliceflow/internal/gittest"
)

// Sliceflow's own lines of the errors below, word for word (R37, R39).
const (
	// makeCommitFailed is R37's first line for the commit, with nothing
	// left (S11, S22).
	makeCommitFailed = "Error: git could not make the commit. The slice was not created. Nothing was changed.\n"
	// commitFailedWithLeft is the same first line, when R41 adds lines
	// after git's text (S12).
	commitFailedWithLeft = "Error: git could not make the commit. The slice was not created.\n"
	// createWorktreeFailed is R37's first line for the worktree, with
	// nothing left (S10).
	createWorktreeFailed = "Error: git could not create the worktree. The slice was not created. Nothing was changed.\n"
	// stoppedLine is R39's line.
	stoppedLine = "Stopped. The slice was not created. Nothing was changed.\n"
	// gitSaid is the line that comes before git's own text (R37).
	gitSaid = "git said:\n"
)

// TestCreateWorktreeFolderNotEmpty checks S10: the worktree folder is
// there already, with a file in it, so git refuses to make the worktree
// (R25). The branch that the run made is removed. The folder and its file
// are not touched (R36).
func TestCreateWorktreeFolderNotEmpty(t *testing.T) {
	shop := newShop(t)
	mainTip := gittest.Git(t, shop, "rev-parse", "main")
	cmd, home := createCmd(t, shop, createOpts{}, "slice", "create", "card-payment")
	worktree := homeWorktree(home, "card-payment")
	writeFile(t, filepath.Join(worktree, "notes.txt"), "keep me\n")

	got := runCmd(t, cmd)

	checkGitStep(t, got, createWorktreeFailed, "")
	checkNoBranch(t, shop, "card-payment")
	checkNoExtraWorktree(t, shop)
	if notes := readFile(t, filepath.Join(worktree, "notes.txt")); notes != "keep me\n" {
		t.Errorf("notes.txt:\n%s\nwant it as it was: keep me", notes)
	}
	// The folder holds only the file that the test put there: git made
	// nothing in it.
	entries, err := os.ReadDir(worktree)
	if err != nil {
		t.Fatalf("read the worktree folder: %v", err)
	}
	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	if !slices.Equal(names, []string{"notes.txt"}) {
		t.Errorf("the worktree folder holds %v, want only notes.txt", names)
	}
	checkTip(t, shop, "main", mainTip)
}

// TestCreateCommitFails checks S11: git cannot work out a name and an
// email, so the commit fails. sliceflow removes the files, the worktree,
// the branch, and the folders it made for the worktree (R36, R37).
func TestCreateCommitFails(t *testing.T) {
	shop := newShop(t)
	mainTip := gittest.Git(t, shop, "rev-parse", "main")

	got, home := create(t, shop, createOpts{NoIdentity: true}, "card-payment")

	checkGitStep(t, got, makeCommitFailed, "")
	checkNothingLeft(t, shop, home, "card-payment")
	checkTip(t, shop, "main", mainTip)
}

// TestCreateHookRejects checks R6 and R37: a pre-commit hook that prints
// "no" and exits with 1 makes the commit fail. git's own text, with the
// hook's "no", is shown, and nothing is left. The hook is a shell script,
// so the test is skipped on Windows.
func TestCreateHookRejects(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("a shell hook may not run on Windows")
	}
	shop := newShop(t)
	mainTip := gittest.Git(t, shop, "rev-parse", "main")
	writeHook(t, shop, "pre-commit", "#!/bin/sh\necho no\nexit 1\n")

	got, home := create(t, shop, createOpts{}, "card-payment")

	text := checkGitStep(t, got, makeCommitFailed, "")
	if !slices.Contains(strings.Split(text, "\n"), "  no") {
		t.Errorf("git's text has no line \"  no\", the hook's word:\n%s", text)
	}
	checkNothingLeft(t, shop, home, "card-payment")
	checkTip(t, shop, "main", mainTip)
}

// TestCreateKeepsEmptyWorktreeFolder checks S22: the worktree folder is
// there and empty, and the commit fails. git removes the worktree, which
// removes the folder too, so sliceflow makes the folder again, empty
// (R36). The folders that were there before stay.
func TestCreateKeepsEmptyWorktreeFolder(t *testing.T) {
	shop := newShop(t)
	mainTip := gittest.Git(t, shop, "rev-parse", "main")
	cmd, home := createCmd(t, shop, createOpts{NoIdentity: true}, "slice", "create", "card-payment")
	worktree := homeWorktree(home, "card-payment")
	if err := os.MkdirAll(worktree, 0o755); err != nil {
		t.Fatalf("make the worktree folder: %v", err)
	}

	got := runCmd(t, cmd)

	checkGitStep(t, got, makeCommitFailed, "")
	checkEmpty(t, worktree)
	checkNoBranch(t, shop, "card-payment")
	checkNoExtraWorktree(t, shop)
	checkTip(t, shop, "main", mainTip)
}

// checkGitStep checks a run that failed in a git step (R37): exit code
// 1, nothing on the normal output, and on the error output the text
// header, then the line "git said:", then git's own text, then tail. The
// tail is the lines after git's text: "" when R41 adds none. It returns
// git's text.
func checkGitStep(t *testing.T, got result, header, tail string) string {
	t.Helper()
	if got.code != 1 {
		t.Errorf("exit code %d, want 1", got.code)
	}
	if got.stdout != "" {
		t.Errorf("normal output:\n%s\nwant nothing", got.stdout)
	}
	start := header + gitSaid
	if !strings.HasPrefix(got.stderr, start) || !strings.HasSuffix(got.stderr, tail) || len(got.stderr) < len(start)+len(tail) {
		t.Errorf("error output:\n%s\nwant it to start with:\n%s\nand to end with:\n%s", got.stderr, start, tail)
		return ""
	}
	text := got.stderr[len(start) : len(got.stderr)-len(tail)]
	checkGitText(t, text)
	return text
}

// checkGitText checks that git's own text is there, and shown as R37
// says: at least one line with text, and each line that is not empty
// indented by two spaces. The words are not compared, because git's text
// changes between git versions and languages.
func checkGitText(t *testing.T, text string) {
	t.Helper()
	withText := false
	for _, line := range strings.Split(strings.TrimSuffix(text, "\n"), "\n") {
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "  ") {
			t.Errorf("a line of git's text is not indented by two spaces: %q", line)
		}
		if strings.TrimSpace(line) != "" {
			withText = true
		}
	}
	if !withText {
		t.Errorf("git's text is empty:\n%q\nwant the text that git printed", text)
	}
}

// checkNoBranch checks that the repo has no branch slice/<name>.
func checkNoBranch(t *testing.T, repo, name string) {
	t.Helper()
	if branches := gittest.Git(t, repo, "branch", "--list", "slice/"+name); branches != "" {
		t.Errorf("the branch slice/%s is still there:\n%s", name, branches)
	}
}

// worktreePaths returns the paths that git lists as worktrees of the
// repo, from the lines "worktree <path>" of git worktree list.
func worktreePaths(t *testing.T, repo string) []string {
	t.Helper()
	var paths []string
	for _, line := range strings.Split(gittest.Git(t, repo, "worktree", "list", "--porcelain"), "\n") {
		if path, ok := strings.CutPrefix(line, "worktree "); ok {
			paths = append(paths, path)
		}
	}
	return paths
}

// checkNoExtraWorktree checks that git lists only the main folder of the
// repo as a worktree, so no worktree of the run is still registered (R36).
func checkNoExtraWorktree(t *testing.T, repo string) {
	t.Helper()
	if paths := worktreePaths(t, repo); len(paths) != 1 {
		t.Errorf("git lists %d worktrees, want only the main folder:\n%s", len(paths), strings.Join(paths, "\n"))
	}
}

// checkGone checks that nothing is at path, not even a broken link.
func checkGone(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Lstat(path); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("%s is still there (%v), want it gone", path, err)
	}
}

// checkNothingLeft checks that a run which removed what it made left
// nothing (R36): no branch slice/<name>, no worktree of the repo besides
// its main folder, no worktree folder, and no .sliceflow folder in home.
// The run made the .sliceflow folder, because the test's home was empty
// before the run.
func checkNothingLeft(t *testing.T, repo, home, name string) {
	t.Helper()
	checkNoBranch(t, repo, name)
	checkNoExtraWorktree(t, repo)
	checkGone(t, homeWorktree(home, name))
	checkGone(t, filepath.Join(home, ".sliceflow"))
}

// writeHook writes the git hook name in the hooks folder of the repo at
// dir, as the script script, and makes it executable. A worktree of the
// repo uses the same hooks.
func writeHook(t *testing.T, dir, name, script string) {
	t.Helper()
	path := filepath.Join(dir, ".git", "hooks", name)
	writeFile(t, path, script)
	if err := os.Chmod(path, 0o755); err != nil {
		t.Fatalf("make %s executable: %v", path, err)
	}
}
