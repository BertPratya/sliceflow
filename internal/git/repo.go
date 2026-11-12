package git

import (
	"context"
	"path/filepath"
	"strings"
)

// RepoRoot returns the folder of the main repo that dir is in, from any
// subfolder of it and from inside any of its worktrees (R16). The name
// of the repo, used in a worktree's path, is filepath.Base of it.
//
// Steps:
//  1. Run git in dir: rev-parse --path-format=absolute --git-common-dir.
//     It prints the main repo's .git folder, even from a worktree.
//  2. If it fails, return its error as it is: ErrNotFound when git
//     cannot be run, or a *Error, e.g. exit code 128 outside a repo.
//  3. Remove the spaces and line breaks around git's output.
//  4. Return the parent folder of that .git folder.
//
// In a submodule, or when .git is not inside the repo's folder, the
// result is not the repo's folder (see the backlog).
func RepoRoot(ctx context.Context, dir string) (string, error) {
	result, err := run(ctx, dir, "rev-parse", "--path-format=absolute", "--git-common-dir")

	if err != nil {
		return "", err
	}

	path := strings.TrimSpace(result)
	repoRoot := filepath.Dir(path)
	return repoRoot, nil
}
