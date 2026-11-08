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
	_, err := run(ctx, dir, "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
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
