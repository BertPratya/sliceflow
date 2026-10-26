package slice

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/BertPratya/sliceflow/internal/git"
)

// CreateOptions is what Create needs to make a slice.
type CreateOptions struct {
	Dir   string    // the folder the command runs in
	Name  string    // the slice name
	From  string    // the --from branch; empty if not given
	Title string    // the --title text; empty if not given
	Now   time.Time // the time of the command
}

// Result is what Create made. Later chunks fill it.
type Result struct{}

// Create makes a new slice. It first runs the checks of R11 in order, and
// stops at the first that fails. Every error is a finished sentence that
// ends with "Nothing was changed." (PD1).
func Create(opts CreateOptions) (Result, error) {
	if err := checkCreate(opts); err != nil {
		return Result{}, err
	}

	// TODO(chunk 4): create the branch, the worktree, the files, and the commit.
	return Result{}, errors.New("creating a slice is not built yet")
}

// checkCreate runs the checks R12, R13, R14, R16, R15, R17, in this order.
// It changes nothing.
func checkCreate(opts CreateOptions) error {
	root, err := git.Root(opts.Dir)
	if err != nil {
		return unchanged(err)
	}

	hasCommits, err := git.HasCommits(root)
	if err != nil {
		return unchanged(err)
	}
	if !hasCommits {
		return unchanged(errors.New("the repository has no commits yet. Make a first commit, then try again"))
	}

	if err := ValidateName(opts.Name); err != nil {
		return unchanged(err)
	}

	base, err := baseBranch(root, opts.From)
	if err != nil {
		return unchanged(err)
	}

	if err := checkNameFree(root, opts.Name, base); err != nil {
		return err
	}

	mainRoot, err := git.MainRoot(root)
	if err != nil {
		return unchanged(err)
	}
	old := filepath.Join(mainRoot, ".sliceflow", "worktrees", opts.Name)
	_, err = os.Stat(old)
	if err == nil {
		return unchanged(fmt.Errorf("folder .sliceflow/worktrees/%s already exists. Remove it, or pick another name", opts.Name))
	}
	if !errors.Is(err, os.ErrNotExist) {
		return unchanged(err)
	}
	return nil
}

// checkNameFree returns the R15 error when the slice name is already used:
// by a local branch, by a branch of a remote, or by a folder on the base
// branch. It names the first place it finds, in that order.
func checkNameFree(root, name, base string) error {
	exists := func(where string) error {
		return unchanged(fmt.Errorf("slice %s already exists: %s", name, where))
	}
	branch := "slice/" + name

	local, err := git.BranchExists(root, branch)
	if err != nil {
		return unchanged(err)
	}
	if local {
		return exists("branch " + branch)
	}

	remotes, err := git.RemoteBranches(root, branch)
	if err != nil {
		return unchanged(err)
	}
	if len(remotes) > 0 {
		return exists("branch " + remotes[0])
	}

	folder := "docs/slices/" + name
	onBase, err := git.PathExists(root, "refs/heads/"+base, folder)
	if err != nil {
		return unchanged(err)
	}
	if onBase {
		return exists(fmt.Sprintf("folder %s on %s", folder, base))
	}
	return nil
}

// unchanged turns the error of a check into the user's message: the error
// text, then ". Nothing was changed." (PD1). git.ErrNotRepo already reads
// as R12 asks: "not inside a git repository".
func unchanged(err error) error {
	return fmt.Errorf("%w. Nothing was changed.", err)
}
