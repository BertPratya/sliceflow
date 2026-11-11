package e2e

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/BertPratya/sliceflow/internal/gittest"
)

// checkRefused checks a run that was refused: the normal output, the
// error output, and the exit code are want's, word for word.
func checkRefused(t *testing.T, got, want result) {
	t.Helper()
	if got.stdout != want.stdout {
		t.Errorf("normal output:\n%s\nwant:\n%s", got.stdout, want.stdout)
	}
	if got.stderr != want.stderr {
		t.Errorf("error output:\n%s\nwant:\n%s", got.stderr, want.stderr)
	}
	if got.code != want.code {
		t.Errorf("exit code %d, want %d", got.code, want.code)
	}
}

// checkNothingMade checks that a refused run made nothing. When repo is
// not "", the repo has no branch slice/<name>. The worktrees of home
// are checked by checkNoWorktree.
func checkNothingMade(t *testing.T, repo, home, name string) {
	t.Helper()
	if repo != "" {
		if branches := gittest.Git(t, repo, "branch", "--list", "slice/"+name); branches != "" {
			t.Errorf("the branch slice/%s exists:\n%s\nwant none", name, branches)
		}
	}
	checkNoWorktree(t, home)
}

// checkTip checks that the branch still points at the commit want, the
// commit it pointed at before the run.
func checkTip(t *testing.T, repo, branch, want string) {
	t.Helper()
	if got := gittest.Git(t, repo, "rev-parse", "--verify", "refs/heads/"+branch); got != want {
		t.Errorf("branch %s points at %s, want %s", branch, got, want)
	}
}

// checkNoWorktree checks that home has no worktree folder. The worktrees
// are in home/.sliceflow/worktrees (R17), so that folder must be missing
// or empty.
func checkNoWorktree(t *testing.T, home string) {
	t.Helper()
	root := filepath.Join(home, ".sliceflow", "worktrees")
	entries, err := os.ReadDir(root)
	if errors.Is(err, fs.ErrNotExist) {
		return
	}
	if err != nil {
		t.Fatalf("read %s: %v", root, err)
	}
	if len(entries) != 0 {
		t.Errorf("%s has %d entries, want no worktree folder", root, len(entries))
	}
}

// checkEmpty checks that the folder dir has no entries, so a refused run
// wrote nothing: no .git and no file.
func checkEmpty(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read %s: %v", dir, err)
	}
	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	if len(names) != 0 {
		t.Errorf("%s is not empty: %v", dir, names)
	}
}

// createOutside runs `sliceflow slice create` with args in a new empty
// folder. It returns the result, the folder, and the home folder. The
// folder is not inside a git repo (R13).
//
// GIT_CEILING_DIRECTORIES stops git from looking above the test's
// temporary folder. It stops git only. sliceflow's own search for a .git
// entry (isInRepo in internal/git) still walks up to the root, so this
// test assumes that the temporary folder is not inside a repo: no .git
// is in it or in a folder above it.
func createOutside(t *testing.T, args ...string) (result, string, string) {
	t.Helper()
	parent, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("find the real path of the temporary folder: %v", err)
	}
	dir := filepath.Join(parent, "plain")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatalf("make the folder: %v", err)
	}
	cmd, home := createCmd(t, dir, createOpts{}, append([]string{"slice", "create"}, args...)...)
	cmd.Env = append(cmd.Env, "GIT_CEILING_DIRECTORIES="+filepath.ToSlash(parent))
	return runCmd(t, cmd), dir, home
}

// TestCreateBadName checks S2: a name with a dot is refused, with exit
// code 2 (R19, R20).
func TestCreateBadName(t *testing.T) {
	shop := newShop(t)
	mainTip := gittest.Git(t, shop, "rev-parse", "main")

	got, home := create(t, shop, createOpts{}, "card.payment")

	checkRefused(t, got, result{
		stderr: "Error: invalid slice name \"card.payment\". Nothing was changed.\n" +
			"Allowed characters: letters, digits, hyphens, underscores. Start with a letter or a digit.\n",
		code: 2,
	})
	checkNothingMade(t, shop, home, "card.payment")
	checkTip(t, shop, "main", mainTip)
}

// TestCreateBranchTaken checks S3: the branch slice/card-payment is there
// already, so the run is refused (R24). The branch was made before the
// run, so the check is that it still points at the same commit.
func TestCreateBranchTaken(t *testing.T) {
	shop := newShop(t)
	gittest.Git(t, shop, "branch", "slice/card-payment")
	before := gittest.Git(t, shop, "rev-parse", "slice/card-payment")
	mainTip := gittest.Git(t, shop, "rev-parse", "main")

	got, home := create(t, shop, createOpts{}, "card-payment")

	checkRefused(t, got, result{
		stderr: "Error: slice card-payment already exists (branch slice/card-payment). Nothing was changed.\n",
		code:   1,
	})
	if after := gittest.Git(t, shop, "rev-parse", "slice/card-payment"); after != before {
		t.Errorf("the branch moved from %s to %s", before, after)
	}
	checkNoWorktree(t, home)
	checkTip(t, shop, "main", mainTip)
}

// TestCreateFolderTaken checks S4: the slice was merged into main, so its
// folder is in main and its branch is gone. The run is refused (R23).
func TestCreateFolderTaken(t *testing.T) {
	shop := newShop(t)
	mainTip := commitFile(t, shop, "docs/slices/card-payment/spec.md", "# card-payment\n")

	got, home := create(t, shop, createOpts{}, "card-payment")

	checkRefused(t, got, result{
		stderr: "Error: slice card-payment already exists (folder docs/slices/card-payment/ in main). Nothing was changed.\n",
		code:   1,
	})
	checkNothingMade(t, shop, home, "card-payment")
	checkTip(t, shop, "main", mainTip)
}

// TestCreateNotARepo checks S6: the folder is not in a git repo (R13).
func TestCreateNotARepo(t *testing.T) {
	got, dir, home := createOutside(t, "card-payment")

	checkRefused(t, got, result{
		stderr: "Error: this folder is not in a git repository. Nothing was changed.\n" +
			"Run sliceflow inside your project.\n",
		code: 1,
	})
	checkNothingMade(t, "", home, "card-payment")
	checkEmpty(t, dir)
}

// TestCreateNoCommits checks S7: the repo has no commits yet (R15). There
// is no commit, so no branch can be there before or after the run.
func TestCreateNoCommits(t *testing.T) {
	repo := gittest.NewEmpty(t)

	got, home := create(t, repo.Dir, createOpts{}, "card-payment")

	checkRefused(t, got, result{
		stderr: "Error: this repository has no commits yet. Nothing was changed.\n" +
			"Make a first commit, then try again.\n",
		code: 1,
	})
	checkNothingMade(t, repo.Dir, home, "card-payment")
	if refs := gittest.Git(t, repo.Dir, "for-each-ref", "refs/heads"); refs != "" {
		t.Errorf("the repo has branches:\n%s\nwant none", refs)
	}
}

// TestCreateNoMain checks S8: the repo's branch is called master, so
// there is no main, and the base is missing (R12). The check is that
// master still points at the same commit.
func TestCreateNoMain(t *testing.T) {
	shop := newShop(t)
	gittest.Git(t, shop, "branch", "-m", "main", "master")
	masterTip := gittest.Git(t, shop, "rev-parse", "master")

	got, home := create(t, shop, createOpts{}, "card-payment")

	checkRefused(t, got, result{
		stderr: "Error: base branch main does not exist. Nothing was changed.\n" +
			"sliceflow starts from main unless you choose another branch. To choose one, run:\n" +
			"  sliceflow slice create card-payment --from <branch>\n",
		code: 1,
	})
	checkNothingMade(t, shop, home, "card-payment")
	checkTip(t, shop, "master", masterTip)
}

// TestCreateFromMissing checks S17: --from names a branch that is not
// there (R12). The base is missing, so the check is on main, which is
// still in the repo.
func TestCreateFromMissing(t *testing.T) {
	shop := newShop(t)
	mainTip := gittest.Git(t, shop, "rev-parse", "main")

	got, home := create(t, shop, createOpts{}, "card-payment", "--from", "release")

	checkRefused(t, got, result{
		stderr: "Error: base branch release does not exist. Nothing was changed.\n" +
			"Check the name. To see your branches, run:\n" +
			"  git branch\n",
		code: 1,
	})
	checkNothingMade(t, shop, home, "card-payment")
	checkTip(t, shop, "main", mainTip)
}

// TestCreateControlCharacter checks S20: the title has a line break, so
// the run is refused, with exit code 2 (R29, R76). The received text is
// written with \n, as R29 says.
func TestCreateControlCharacter(t *testing.T) {
	shop := newShop(t)
	mainTip := gittest.Git(t, shop, "rev-parse", "main")

	got, home := create(t, shop, createOpts{}, "card-payment", "--title", "Pay with a card\n")

	checkRefused(t, got, result{
		stderr: "Error: --title has a control character. Nothing was changed.\n" +
			`Received: "Pay with a card\n"` + "\n" +
			"Write it as plain text on one line, with no control characters.\n",
		code: 2,
	})
	checkNothingMade(t, shop, home, "card-payment")
	checkTip(t, shop, "main", mainTip)
}

// TestCreateNoGit checks S21: PATH is an empty folder, so sliceflow
// finds no git (R14).
func TestCreateNoGit(t *testing.T) {
	shop := newShop(t)
	noGit := t.TempDir()
	mainTip := gittest.Git(t, shop, "rev-parse", "main")

	got, home := create(t, shop, createOpts{Path: noGit}, "card-payment")

	checkRefused(t, got, result{
		stderr: "Error: git was not found. Nothing was changed.\n" +
			"Install git, then try again.\n",
		code: 1,
	})
	checkNothingMade(t, shop, home, "card-payment")
	checkTip(t, shop, "main", mainTip)
}

// TestCreateTwoProblems checks R40: the name is bad, the --from branch
// is missing, and the folder is not a repo. Only the name is reported,
// because R20 comes before R13 and R12 (exit code 2, R76).
func TestCreateTwoProblems(t *testing.T) {
	got, dir, home := createOutside(t, "bad name", "--from", "nope")

	checkRefused(t, got, result{
		stderr: "Error: invalid slice name \"bad name\". Nothing was changed.\n" +
			"Allowed characters: letters, digits, hyphens, underscores. Start with a letter or a digit.\n",
		code: 2,
	})
	checkNothingMade(t, "", home, "bad name")
	checkEmpty(t, dir)
}
