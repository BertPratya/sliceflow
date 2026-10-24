package slice

import (
	"context"
	"strings"

	"github.com/BertPratya/sliceflow/internal/git"
)

// findWorktree returns the full path of the worktree whose branch is
// refs/heads/slice/<name>. The bool is false when no worktree has it.
// name must pass CheckName.
func findWorktree(ctx context.Context, r Repo, name string) (string, bool, error) {
	out, err := git.Run(ctx, r.Dir, "worktree", "list", "--porcelain", "-z")
	if err != nil {
		return "", false, err
	}
	want := "refs/heads/slice/" + name
	// Records are separated by an empty field; each field is one line.
	var dir string
	for _, field := range strings.Split(out, "\x00") {
		switch {
		case field == "":
			dir = ""
		case strings.HasPrefix(field, "worktree "):
			dir = strings.TrimPrefix(field, "worktree ")
		case field == "branch "+want && dir != "":
			return dir, true, nil
		}
	}
	return "", false, nil
}
