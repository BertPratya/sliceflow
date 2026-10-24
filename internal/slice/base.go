package slice

import (
	"context"
	"errors"
	"strings"

	"github.com/BertPratya/sliceflow/internal/git"
)

// BaseBranch picks the base branch: from (--from), else git config
// sliceflow.base, else main. source says which one won. It does not check
// that the branch exists.
func BaseBranch(ctx context.Context, r Repo, from string) (branch, source string, err error) {
	if from != "" {
		return from, "from --from", nil
	}
	out, err := git.Run(ctx, r.Dir, "config", "--get", "sliceflow.base")
	if err != nil {
		var ge *git.Error
		if errors.As(err, &ge) && ge.ExitCode == 1 {
			return "main", "the default", nil
		}
		return "", "", err
	}
	if v := strings.TrimSpace(out); v != "" {
		return v, "from git config sliceflow.base", nil
	}
	return "main", "the default", nil
}
