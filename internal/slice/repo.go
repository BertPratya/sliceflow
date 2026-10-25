package slice

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"time"

	"github.com/BertPratya/sliceflow/internal/git"
)

// Repo is where a command runs. Dir is the folder the command runs in.
// Root is the root of the main checkout, with "/" on every system.
type Repo struct {
	Dir, Root string
	Now       func() time.Time
	// LockWait is how long takeLock tries again. 0 means 10 s.
	LockWait time.Duration
}

// FindRepo checks that dir is inside a git repo with at least one commit,
// and finds the main checkout's root.
func FindRepo(ctx context.Context, dir string) (Repo, error) {
	if _, err := git.Run(ctx, dir, "rev-parse", "--show-toplevel"); err != nil {
		var ge *git.Error
		if errors.As(err, &ge) {
			return Repo{}, &Error{Message: "not inside a git repository. Nothing was changed."}
		}
		return Repo{}, err
	}
	out, err := git.Run(ctx, dir, "for-each-ref", "--count=1", "refs/heads/")
	if err != nil {
		return Repo{}, err
	}
	if strings.TrimSpace(out) == "" {
		return Repo{}, &Error{Message: "the repository has no commits yet. Make a first commit, then try again. Nothing was changed."}
	}
	list, err := git.Run(ctx, dir, "worktree", "list", "--porcelain", "-z")
	if err != nil {
		return Repo{}, err
	}
	first, _, _ := strings.Cut(list, "\x00")
	root, ok := strings.CutPrefix(first, "worktree ")
	if !ok {
		return Repo{}, errors.New("unexpected output from git worktree list")
	}
	return Repo{Dir: dir, Root: filepath.ToSlash(root), Now: time.Now}, nil
}
