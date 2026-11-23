package git

import (
	"context"
	"errors"
)

// HasCommits reports whether the repo in dir has any commit, on any
// branch or tag. It asks about the whole repo, not only the current
// branch: an orphan branch has no commit, but the repo may. A failure is
// returned as it is.
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
// rule matches them, and commits only them with message, on the branch
// checked out in dir. Other changes, staged or not, stay as they were
// and are not in the commit. The commit uses the user's own git
// settings, so hooks and commit signing run as the user set them up.
// When adding or committing fails, it returns git's error as a *Error,
// and nothing is committed; the files may still be staged. Git refuses
// to commit only some files during a merge, so Commit fails then.
func Commit(ctx context.Context, dir, message string, paths ...string) error {
	add := append([]string{"add", "--force", "--"}, paths...)
	if _, err := run(ctx, dir, add...); err != nil {
		return err
	}
	commit := append([]string{"commit", "--quiet", "--only", "-m", message, "--"}, paths...)
	_, err := run(ctx, dir, commit...)
	return err
}

// Unstage puts the files at path, relative to dir, back in git's staging
// area as they are in the commit checked out in dir: a file that is not
// in that commit is no longer staged. The files on the disk do not
// change, and nothing outside path is unstaged. A path with nothing
// staged is not an error. When git fails, it returns git's error as a
// *Error.
func Unstage(ctx context.Context, dir, path string) error {
	_, err := run(ctx, dir, "reset", "--quiet", "--", path)
	return err
}
