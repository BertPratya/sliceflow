package git

import (
	"context"
	"errors"
)

// DirectoryExists reports whether path is a folder in commit. A file at
// path does not count. Any failure other than "no such folder" is
// returned as it is.
//
// commit must be a commit ID that is known to exist, e.g. from
// BranchCommit: git gives the same answer for a commit it cannot find as
// for a missing path. path is from the repo's top folder, with "/"
// between its parts and none at its end, e.g. docs/slices/login.
func DirectoryExists(ctx context.Context, dir, commit, path string) (bool, error) {
	_, err := run(ctx, dir, "rev-parse", "--verify", "--quiet", commit+":"+path+"/")
	if err == nil {
		return true, nil
	}
	var gitErr *Error
	if errors.As(err, &gitErr) && gitErr.ExitCode == 1 {
		return false, nil
	}
	return false, err
}

// PathExists reports whether commit has a file or a folder at path. Any
// failure other than "no such path" is returned as it is. commit and path
// are as for DirectoryExists, and commit may also be a name, e.g. HEAD.
func PathExists(ctx context.Context, dir, commit, path string) (bool, error) {
	_, err := run(ctx, dir, "rev-parse", "--verify", "--quiet", commit+":"+path)
	if err == nil {
		return true, nil
	}
	var gitErr *Error
	if errors.As(err, &gitErr) && gitErr.ExitCode == 1 {
		return false, nil
	}
	return false, err
}
