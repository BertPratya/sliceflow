package git

import (
	"context"
	"path/filepath"
	"strings"
)

// RepoRoot returns the folder of the main repo that dir is in, from any
// subfolder of it and from inside any of its worktrees. The name of the
// repo, used in a worktree's path, is filepath.Base of it. A failure is
// returned as it is: ErrNotFound when git cannot be run, or a *Error,
// e.g. with exit code 128 outside a repo.
//
// It is the folder above the repo's shared .git folder. In a submodule,
// or when .git is not inside the repo's folder, that is not the repo's
// folder (see the backlog).
func RepoRoot(ctx context.Context, dir string) (string, error) {
	result, err := run(ctx, dir, "rev-parse", "--path-format=absolute", "--git-common-dir")

	if err != nil {
		return "", err
	}

	path := strings.TrimSpace(result)
	repoRoot := filepath.Dir(path)
	return repoRoot, nil
}

// CheckoutTop returns the top folder of the checkout that dir is in: the
// repo's main folder, or a worktree's own folder when dir is inside a
// worktree. Inside a worktree it differs from RepoRoot. The path is in
// this system's form, e.g. with \ on Windows. A failure is returned as it
// is: ErrNotFound when git cannot be run, or a *Error, e.g. with exit
// code 128 outside a repo or in a repo with no checkout.
func CheckoutTop(ctx context.Context, dir string) (string, error) {
	out, err := run(ctx, dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	return filepath.FromSlash(strings.TrimRight(out, "\r\n")), nil
}
