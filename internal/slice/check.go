package slice

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/BertPratya/sliceflow/internal/git"
)

// checkCreate runs the checks of R14, R16, R15 and R17, in that order,
// before anything is made. R12 and R13 come from FindRepo.
func checkCreate(ctx context.Context, r Repo, opts CreateOptions, base, source string) error {
	name := opts.Name
	if err := CheckName(name); err != nil {
		return err
	}
	ok, err := git.RefExists(ctx, r.Dir, "refs/heads/"+base)
	if err != nil {
		return err
	}
	if !ok {
		return &Error{Message: fmt.Sprintf("base branch %s does not exist (%s). Nothing was changed.", base, source)}
	}
	where, err := existsWhere(ctx, r, name, base)
	if err != nil {
		return err
	}
	if where != "" {
		return &Error{Message: fmt.Sprintf("slice %s already exists: %s. Nothing was changed.", name, where)}
	}
	if _, err := os.Lstat(worktreeDir(r.Root, name)); err == nil {
		return &Error{Message: fmt.Sprintf("folder .sliceflow/worktrees/%s already exists. Remove it, or pick another name. Nothing was changed.", name)}
	} else if !os.IsNotExist(err) {
		return err
	}
	return nil
}

// existsWhere returns where the slice name is taken (R15's <where>), or ""
// when it is free.
func existsWhere(ctx context.Context, r Repo, name, base string) (string, error) {
	branch := "slice/" + name
	ok, err := git.RefExists(ctx, r.Dir, "refs/heads/"+branch)
	if err != nil {
		return "", err
	}
	if ok {
		return "branch " + branch, nil
	}
	out, err := git.Run(ctx, r.Dir, "remote")
	if err != nil {
		return "", err
	}
	for _, remote := range strings.Fields(out) {
		ok, err := git.RefExists(ctx, r.Dir, "refs/remotes/"+remote+"/"+branch)
		if err != nil {
			return "", err
		}
		if ok {
			return "branch " + remote + "/" + branch, nil
		}
	}
	folder := path.Join("docs", "slices", name)
	if _, err := git.Run(ctx, r.Dir, "cat-file", "-e", "refs/heads/"+base+":"+folder); err == nil {
		return "folder " + folder + " on " + base, nil
	} else if _, isGit := err.(*git.Error); !isGit {
		return "", err
	}
	return "", nil
}
