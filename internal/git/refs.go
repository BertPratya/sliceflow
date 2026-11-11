package git

import (
	"context"
	"errors"
)

// ErrNoBranch means there is no local branch with that name (R12). A
// tag, a commit, or a remote branch such as origin/main is not a local
// branch (R10).
var ErrNoBranch = errors.New("no such local branch")

// BranchExists reports whether the local branch refs/heads/<branch>
// exists. It never looks at remote branches or tags (R24, R26).
//
// It asks git show-ref --verify --quiet refs/heads/<branch>, which
// takes only that exact, full ref name: exit 0 means yes, and exit 1
// means no. Any other failure is returned as it is. The argument always
// starts with refs/heads/, so a branch such as -x is never read as an
// option.
func BranchExists(ctx context.Context, dir, branch string) (bool, error) {
	return branchExists(ctx, dir, branch, run)
}

// BranchExistsShielded is BranchExists, with git shielded from Ctrl+C
// (runShielded). The undo asks it, after a failure or a Ctrl+C, whether
// the branch was made anyway (R36, R39).
//
// Steps: as BranchExists, with runShielded.
func BranchExistsShielded(ctx context.Context, dir, branch string) (bool, error) {
	return branchExists(ctx, dir, branch, runShielded)
}

// branchExists is the code of BranchExists and BranchExistsShielded. It
// runs git with runner: run or runShielded.
func branchExists(ctx context.Context, dir, branch string, runner func(context.Context, string, ...string) (string, error)) (bool, error) {
	_, err := runner(ctx, dir, "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
	if err == nil {
		return true, nil
	}
	var gitErr *Error
	if errors.As(err, &gitErr) && gitErr.ExitCode == 1 {
		return false, nil
	}
	return false, err
}

// BranchCommit returns the hash of the commit at the tip of the local
// branch refs/heads/<branch>, as it is now (R10, R11). It never fetches
// (SD2).
//
// Steps:
//  1. Check that the local branch exists, with BranchExists. If it does
//     not, return ErrNoBranch. This step comes first because git
//     rev-parse also accepts tags, hashes, remote branches, and forms
//     such as main~1.
//  2. Ask git rev-parse --verify refs/heads/<branch>^{commit} for the
//     hash. ^{commit} makes sure the answer is a commit.
func BranchCommit(ctx context.Context, dir, branch string) (string, error) {
	exists, err := BranchExists(ctx, dir, branch)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", ErrNoBranch
	}
	out, err := run(ctx, dir, "rev-parse", "--verify", "refs/heads/"+branch+"^{commit}")
	if err != nil {
		return "", err
	}
	return trimNewline(out), nil
}

// CreateBranch makes the local branch <branch> at the commit, and does
// not check it out (R37). The branch that is checked out stays as it is
// (R7).
//
// Steps:
//  1. Run git branch --no-track --end-of-options <branch> <commit>.
//     --no-track keeps the user's branch.autoSetupMerge setting from
//     making the new branch follow another one. --end-of-options makes
//     git read a name such as -x as a name, not as an option.
//  2. Return git's error as it is, for example when the branch exists
//     already.
func CreateBranch(ctx context.Context, dir, branch, commit string) error {
	_, err := run(ctx, dir, "branch", "--no-track", "--end-of-options", branch, commit)
	return err
}

// DeleteBranch removes the local branch <branch>, even when its commits
// are on no other branch (R36). The undo calls it, so git is shielded
// from Ctrl+C (runShielded).
//
// Steps:
//  1. Run git branch -D --end-of-options <branch>. -D removes the branch
//     even when it is not merged: the undo removes only a branch that
//     this run made.
//  2. Return git's error as it is.
func DeleteBranch(ctx context.Context, dir, branch string) error {
	_, err := runShielded(ctx, dir, "branch", "-D", "--end-of-options", branch)
	return err
}
