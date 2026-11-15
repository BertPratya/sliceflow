package git

import (
	"context"
	"errors"
)

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

// IsIgnored reports whether a .gitignore rule, or another ignore rule of
// git, matches path, which is relative to dir. The file does not need to
// exist. Any failure other than "no rule matches" is returned as it is.
func IsIgnored(ctx context.Context, dir, path string) (bool, error) {
	_, err := run(ctx, dir, "check-ignore", "--quiet", "--", path)
	if err == nil {
		return true, nil
	}
	var gitErr *Error
	if errors.As(err, &gitErr) && gitErr.ExitCode == 1 {
		return false, nil
	}
	return false, err
}

// Commit adds the files at paths, relative to dir, even when an ignore
// rule matches them, and commits them with message. The commit uses the
// user's own git settings, so hooks and commit signing run as the user
// set them up. When adding or committing fails, it returns git's error
// as a *Error, and nothing is committed.
func Commit(ctx context.Context, dir, message string, paths ...string) error {
	add := append([]string{"add", "--force", "--"}, paths...)
	if _, err := run(ctx, dir, add...); err != nil {
		return err
	}
	_, err := run(ctx, dir, "commit", "--quiet", "-m", message)
	return err
}
