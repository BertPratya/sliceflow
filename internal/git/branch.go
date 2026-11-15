package git

import (
	"context"
	"errors"
	"strings"
)

// ErrNoBranch is returned when a local branch does not exist (R12).
var ErrNoBranch = errors.New("git: no such local branch")

// BranchCommit returns the commit ID at the tip of the local branch
// (R10, R11). It never fetches.
//
// Steps:
//  1. Check BranchExists. If it fails, return its error as it is.
//  2. If the branch does not exist, return ErrNoBranch (R12).
//  3. Run git in dir: rev-parse --verify refs/heads/<branch>.
//  4. Remove the line break at the end, and return the ID.
func BranchCommit(ctx context.Context, dir, branch string) (string, error) {
	exists, err := BranchExists(ctx, dir, branch)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", ErrNoBranch
	}
	out, err := run(ctx, dir, "rev-parse", "--verify", "refs/heads/"+branch)
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(out, "\n"), nil
}

// BranchExists reports whether the local branch exists, by its exact
// name, with no revision syntax such as main~1 (R10).
//
// Steps:
//  1. Run git in dir: show-ref --verify --quiet refs/heads/<branch>.
//  2. If it works, return true.
//  3. If it fails with exit code 1, return false: that is git's answer
//     "no such ref".
//  4. Any other error: return false and the error as it is.
func BranchExists(ctx context.Context, dir, branch string) (bool, error) {
	ref := "refs/heads/" + branch
	_, err := run(ctx, dir, "show-ref", "--verify", "--quiet", ref)

	if err == nil {
		return true, nil
	}
	var gitError *Error
	if errors.As(err, &gitError) && gitError.ExitCode == 1 {
		return false, nil
	}

	return false, err

}

// CreateBranch makes the local branch at commit, without checking it
// out. When git cannot make it, for example because the branch exists,
// it returns git's error as a *Error, with git's message word for word.
func CreateBranch(ctx context.Context, dir, branch, commit string) error {
	_, err := run(ctx, dir, "branch", branch, commit)
	return err
}

// DeleteBranch deletes the local branch, even when its commits are on no
// other branch. Git refuses, with a *Error, when the branch does not
// exist or is checked out in a worktree, so a worktree of the branch
// must be removed first.
func DeleteBranch(ctx context.Context, dir, branch string) error {
	_, err := run(ctx, dir, "branch", "-D", branch)
	return err
}
