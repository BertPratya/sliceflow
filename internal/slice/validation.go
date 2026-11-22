package slice

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/BertPratya/sliceflow/internal/git"
)

// The two ways a name can be taken, in the order CheckNameFree checks
// them. The command turns each one into its message.
var (
	// ErrFolderTaken: the folder docs/slices/<name>/ is in the checkout
	// and not empty, or it is in the base commit.
	ErrFolderTaken = errors.New("slice: the slice's folder exists")
	// ErrBranchTaken: the local branch slice/<name> exists.
	ErrBranchTaken = errors.New("slice: the slice's branch exists")
)

// CheckNameFree returns nil when the slice name is free. The name is free
// when all of these hold:
//   - in the checkout whose top folder is top, docs/slices/<name> is
//     missing or an empty folder;
//   - commit, the base commit, has no folder docs/slices/<name>/;
//   - the repo in dir has no local branch slice/<name>.
//
// It never looks at a remote's branches. commit must be a commit ID, e.g.
// from git.BranchCommit.
//
// When the name is taken, it returns ErrFolderTaken or ErrBranchTaken.
// When both apply, the folder is reported. A failure to read the disk or
// of git is returned as it is.
func CheckNameFree(ctx context.Context, dir, top, commit, name string) error {
	used, err := pathInUse(filepath.Join(top, "docs", "slices", name))
	if err != nil {
		return err
	}
	if used {
		return ErrFolderTaken
	}
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

// pathInUse reports whether something is at path that a new slice must
// not write into: a folder with any entry in it, hidden ones too, or
// anything that is not a folder, e.g. a file or a link. Nothing at path,
// or an empty folder, gives false. A failure to read the disk is
// returned as it is, often as a *fs.PathError.
func pathInUse(path string) (bool, error) {
	info, err := os.Lstat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !info.IsDir() {
		return true, nil
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return false, err
	}
	return len(entries) > 0, nil
}
