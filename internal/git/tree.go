package git

import (
	"context"
	"errors"
)

// DirectoryExists reports whether path is a folder in commit (R23). A
// file at path does not count.
//
// commit must be a commit ID that is known to exist, e.g. from
// BranchCommit: git gives the same answer for a commit it cannot find as
// for a missing path. path is from the repo's top folder, with "/"
// between its parts and none at its end, e.g. docs/slices/login.
//
// Steps:
//  1. Run git in dir: rev-parse --verify --quiet <commit>:<path>/. The
//     "/" at the end makes git accept a folder only.
//  2. If it works, return true.
//  3. If it fails with exit code 1, return false: path is not in commit,
//     or is not a folder.
//  4. Any other error: return false and the error as it is.
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
