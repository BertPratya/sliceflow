package slice

import (
	"context"
	"errors"

	"github.com/BertPratya/sliceflow/internal/git"
)

// The two ways a name can be taken, in the order CheckNameFree checks
// them. The command turns each one into its message.
var (
	// ErrFolderTaken: the folder docs/slices/<name>/ is in the base
	// commit.
	ErrFolderTaken = errors.New("slice: the slice's folder is in the base commit")
	// ErrBranchTaken: the local branch slice/<name> exists.
	ErrBranchTaken = errors.New("slice: the slice's branch exists")
)

// CheckNameFree returns nil when the slice name is free in the repo in
// dir: no folder docs/slices/<name>/ in commit, the base commit, and no
// local branch slice/<name>. It never looks at a remote's branches.
// commit must be a commit ID, e.g. from git.BranchCommit.
//
// When the name is taken, it returns ErrFolderTaken or ErrBranchTaken.
// When both apply, the folder is reported. A failure of git is returned
// as it is.
func CheckNameFree(ctx context.Context, dir, commit, name string) error {
	folder, err := git.DirectoryExists(ctx, dir, commit, "docs/slices/"+name)
	if err != nil {
		return err
	}
	if folder {
		return ErrFolderTaken
	}
	branch, err := git.BranchExists(ctx, dir, "slice/"+name)
	if err != nil {
		return err
	}
	if branch {
		return ErrBranchTaken
	}
	return nil
}
