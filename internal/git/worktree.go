package git

import "context"

// AddWorktree checks out the existing local branch in a new worktree at
// path. Git makes the folder and any missing folders above it. A folder
// that exists and is empty is used. When git cannot make the worktree,
// for example because the folder is not empty or the branch is checked
// out elsewhere, it returns git's error as a *Error.
func AddWorktree(ctx context.Context, dir, path, branch string) error {
	_, err := run(ctx, dir, "worktree", "add", path, branch)
	return err
}

// RemoveWorktree deletes the worktree at path and git's record of it,
// even when it has changes that are not committed. It deletes the
// worktree's folder, even one that existed before the worktree was
// made, but not the folders above it. When git cannot remove it, it
// returns git's error as a *Error.
func RemoveWorktree(ctx context.Context, dir, path string) error {
	_, err := run(ctx, dir, "worktree", "remove", "--force", path)
	return err
}
