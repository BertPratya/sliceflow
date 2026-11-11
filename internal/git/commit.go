package git

import (
	"context"
	"errors"
)

// IsIgnored reports whether a .gitignore rule, or another rule of git's
// ignore files, matches the file at path (R5). path is read from dir.
//
// Steps:
//  1. Run git check-ignore -q --end-of-options <path>.
//  2. Exit 0 means yes, and exit 1 means no, as in BranchExists. Any
//     other failure is returned as it is.
func IsIgnored(ctx context.Context, dir, path string) (bool, error) {
	_, err := run(ctx, dir, "check-ignore", "-q", "--end-of-options", path)
	if err == nil {
		return true, nil
	}
	var gitErr *Error
	if errors.As(err, &gitErr) && gitErr.ExitCode == 1 {
		return false, nil
	}
	return false, err
}

// AddForce adds the files at paths to the next commit, even when a
// .gitignore rule matches them, and only those files (R5). Each path is
// read from dir.
//
// Steps:
//  1. Run git add -f --end-of-options <paths>. There is no --: after
//     --end-of-options, git would read -- as a path.
//  2. Return git's error as it is.
func AddForce(ctx context.Context, dir string, paths ...string) error {
	args := append([]string{"add", "-f", "--end-of-options"}, paths...)
	_, err := run(ctx, dir, args...)
	return err
}

// Commit makes a commit of what was added, with the message, in the
// worktree at dir (R37). It runs with the user's own git settings: it
// never passes --no-verify or -c, so the user's hooks and commit signing
// run as the user set them up (R6).
//
// Steps:
//  1. Run git commit -q -m <message>.
//  2. Return git's error as it is, for example a hook that refused.
func Commit(ctx context.Context, dir, message string) error {
	_, err := run(ctx, dir, "commit", "-q", "-m", message)
	return err
}

// ExcludeFile returns the full path of the repository's info/exclude
// file (R2, SD31). It is the shared file of the main repository, also
// when dir is in a linked worktree. The file may not exist yet.
//
// Steps:
//  1. Run git rev-parse --path-format=absolute --git-path info/exclude.
//  2. Read the answer with absPaths: one full path, with the system's
//     separators.
func ExcludeFile(ctx context.Context, dir string) (string, error) {
	out, err := run(ctx, dir, "rev-parse", "--path-format=absolute", "--git-path", "info/exclude")
	if err != nil {
		return "", err
	}
	paths, err := absPaths(out, 1)
	if err != nil {
		return "", err
	}
	return paths[0], nil
}
