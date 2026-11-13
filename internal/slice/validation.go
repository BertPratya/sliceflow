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
	// commit (R23).
	ErrFolderTaken = errors.New("slice: the slice's folder is in the base commit")
	// ErrBranchTaken: the local branch slice/<name> exists (R24).
	ErrBranchTaken = errors.New("slice: the slice's branch exists")
)

// CheckNameFree returns nil when the slice name is free in the repo in
// dir: no folder docs/slices/<name>/ in commit, the base commit, and no
// local branch slice/<name> (R23, R24). It never looks at a remote's
// branches (R26). commit must be a commit ID, e.g. from git.BranchCommit.
//
// Steps:
//  1. Ask git.DirectoryExists for docs/slices/<name> in commit. If it
//     fails, return its error as it is.
//  2. If the folder exists, return ErrFolderTaken (R23). The folder is
//     checked before the branch, as the spec's order says.
//  3. Ask git.BranchExists for slice/<name>. If it fails, return its
//     error as it is.
//  4. If the branch exists, return ErrBranchTaken (R24).
//  5. Return nil.
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
