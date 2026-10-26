package slice

import (
	"fmt"

	"github.com/BertPratya/sliceflow/internal/git"
)

// baseBranch returns the branch a new slice starts from (R3): from when it
// is not empty, else the value of git config sliceflow.base, else main. It
// returns an error when that branch is not a local branch of the repository
// at root (R16).
func baseBranch(root, from string) (string, error) {
	branch, source := from, "from --from"
	if branch == "" {
		value, ok, err := git.Config(root, "sliceflow.base")
		if err != nil {
			return "", err
		}
		// An empty value counts as not set.
		if ok && value != "" {
			branch, source = value, "from git config sliceflow.base"
		} else {
			branch, source = "main", "the default"
		}
	}

	exists, err := git.BranchExists(root, branch)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", fmt.Errorf("base branch %s does not exist (%s)", branch, source)
	}
	return branch, nil
}
