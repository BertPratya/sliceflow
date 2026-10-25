package slice

import (
	"context"
	"fmt"
	"path"

	"github.com/BertPratya/sliceflow/internal/git"
)

// openSlice runs the first checks of R11 for a slice that has a name, after
// FindRepo (R12 and R13 of 0001): the name (R14 of 0001), that the slice
// exists (R33 of 0001), then that it has a worktree (R11). It returns the
// full path of the worktree, in the system's form.
func openSlice(ctx context.Context, r Repo, name string) (string, error) {
	if err := CheckName(name); err != nil {
		return "", err
	}
	branched, err := git.RefExists(ctx, r.Dir, "refs/heads/slice/"+name)
	if err != nil {
		return "", err
	}
	if !branched {
		// No local branch: the slice is merged if its folder is on the base
		// branch, else there is no such slice.
		base, _, err := BaseBranch(ctx, r, "")
		if err != nil {
			return "", err
		}
		onBase, err := baseExists(ctx, r, base)
		if err != nil {
			return "", err
		}
		if onBase {
			folder := path.Join("docs", "slices", name)
			_, err := git.Run(ctx, r.Dir, "cat-file", "-e", "refs/heads/"+base+":"+folder)
			if err == nil {
				return "", &Error{Message: fmt.Sprintf("slice %s is merged into %s, so it has no worktree. Nothing was changed.", name, base)}
			}
			if _, isGit := err.(*git.Error); !isGit {
				return "", err
			}
		}
		return "", noSliceError(name)
	}
	wt, ok, err := findWorktree(ctx, r, name)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", &Error{Message: fmt.Sprintf("slice %s has no worktree: no worktree has the branch slice/%s checked out. Nothing was changed.\nMake one: git worktree add %s/.sliceflow/worktrees/%s slice/%s",
			name, name, r.Root, name, name)}
	}
	return wt, nil
}
