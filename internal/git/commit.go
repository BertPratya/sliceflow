package git

import "context"

// HasCommits reports whether the repo in dir has any commit, on any
// branch or tag (R15). It asks about the whole repo, not only the
// current branch: an orphan branch has no commit, but the repo may.
//
// Steps:
//  1. Run git in dir: rev-list -n 1 --all.
//  2. If it fails, return false and its error as it is.
//  3. If its output is empty, return false.
//  4. Otherwise, return true.
func HasCommits(ctx context.Context, dir string) (bool, error) {
	out, err := run(ctx, dir, "rev-list", "-n", "1", "--all")
	if err != nil {
		return false, err
	}
	if out == "" {
		return false, nil
	}
	return true, nil
}
