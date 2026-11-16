package git

import (
	"context"
	"path/filepath"
	"strings"
)

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

// WorktreeOf returns the full path of the worktree that has the local
// branch checked out, written with the separator of this system, or ""
// when no worktree has it. The path is as git records it: the folder may
// be gone.
func WorktreeOf(ctx context.Context, dir, branch string) (string, error) {
	out, err := run(ctx, dir, "worktree", "list", "--porcelain")
	if err != nil {
		return "", err
	}
	var path string
	for _, line := range strings.Split(out, "\n") {
		if value, ok := strings.CutPrefix(line, "worktree "); ok {
			path = value
		}
		if line == "branch refs/heads/"+branch {
			return filepath.FromSlash(path), nil
		}
	}
	return "", nil
}
