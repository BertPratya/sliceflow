package git

import (
	"context"
	"errors"
	"strings"
)

// ErrNoBranch is returned when a local branch does not exist.
var ErrNoBranch = errors.New("git: no such local branch")

// BranchCommit returns the full ID of the commit at the tip of the local
// branch. It returns ErrNoBranch when the branch does not exist, by its
// exact name, as BranchExists checks it. It never fetches.
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
// name. Revision syntax such as main~1 is not a branch name, and a tag or
// a remote's branch is not a local branch, so each gives false. Any
// failure other than "no such branch" is returned as it is.
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

// BranchesWithPrefix returns the names of the local branches that start
// with prefix, with the prefix cut off, in git's order: prefix "slice/"
// gives "login" for the branch slice/login.
func BranchesWithPrefix(ctx context.Context, dir, prefix string) ([]string, error) {
	out, err := run(ctx, dir, "for-each-ref", "--format=%(refname)", "refs/heads/"+prefix)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, line := range strings.Split(out, "\n") {
		if name, ok := strings.CutPrefix(line, "refs/heads/"+prefix); ok && name != "" {
			names = append(names, name)
		}
	}
	return names, nil
}
