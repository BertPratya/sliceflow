package slice

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/BertPratya/sliceflow/internal/git"
)

// findWorktree returns the full path of the worktree whose branch is
// refs/heads/slice/<name>. The bool is false when no worktree has it, or when its folder is gone
// (a prunable record). The path is in the system's form.
// name must pass CheckName.
func findWorktree(ctx context.Context, r Repo, name string) (string, bool, error) {
	out, err := git.Run(ctx, r.Dir, "worktree", "list", "--porcelain", "-z")
	if err != nil {
		return "", false, err
	}
	want := "refs/heads/slice/" + name
	// Records are separated by an empty field; each field is one line.
	// A record is judged at its end, when all its fields are known.
	var found string
	var dir, branch string
	var prunable bool
	check := func() {
		if found == "" && dir != "" && branch == want && !prunable {
			if fi, err := os.Stat(dir); err == nil && fi.IsDir() {
				found = filepath.Clean(dir)
			}
		}
		dir, branch, prunable = "", "", false
	}
	for _, field := range strings.Split(out, "\x00") {
		switch {
		case field == "":
			check()
		case strings.HasPrefix(field, "worktree "):
			dir = strings.TrimPrefix(field, "worktree ")
		case strings.HasPrefix(field, "branch "):
			branch = strings.TrimPrefix(field, "branch ")
		case field == "prunable" || strings.HasPrefix(field, "prunable "):
			prunable = true
		}
	}
	check()
	return found, found != "", nil
}
