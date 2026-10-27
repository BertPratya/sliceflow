package slice

import (
	"fmt"
	"slices"
	"strings"

	"github.com/BertPratya/sliceflow/internal/git"
)

// Listing is the result of List.
type Listing struct {
	// Slices are the slices to show, in the order of R34.
	Slices []Info
	// Warnings are the warnings about the whole repo (R23, R43): R23
	// first, then R43 by branch name.
	Warnings []string
}

// List finds the slices of the repo that dir is in (R21-R24, R34-R36,
// R43). A slice is in progress when its local branch slice/<name> exists
// and docs/slices/<name>/ is not on the base branch yet. It is merged when
// that folder is on the base branch. Merged slices are only in the result
// when all is true. It returns an error that ends with "Nothing was
// changed." outside a repo or in a repo with no commits (R22).
func List(dir string, all bool) (Listing, error) {
	root, err := openRepo(dir)
	if err != nil {
		return Listing{}, err
	}
	base, baseExists, repoWarnings, err := loadBase(root)
	if err != nil {
		return Listing{}, err
	}
	result := Listing{Warnings: repoWarnings}

	// hasBranch holds the names of the valid slice branches.
	hasBranch := map[string]bool{}
	branches, err := git.Branches(root, "slice/")
	if err != nil {
		return Listing{}, err
	}
	for _, branch := range branches {
		name := strings.TrimPrefix(branch, "slice/")
		if ValidateName(name) != nil {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("branch %s: not a valid slice name; skipped", branch))
			continue
		}
		hasBranch[name] = true
	}

	// onBase holds the names of the valid folders on the base branch.
	onBase := map[string]bool{}
	baseRev := "refs/heads/" + base
	if baseExists {
		folders, err := git.ListDirs(root, baseRev, "docs/slices")
		if err != nil {
			return Listing{}, err
		}
		for _, name := range folders {
			// A folder with an invalid name is skipped without a warning.
			if ValidateName(name) == nil {
				onBase[name] = true
			}
		}
	}

	var names []string
	for name := range hasBranch {
		names = append(names, name)
	}
	for name := range onBase {
		if !hasBranch[name] {
			names = append(names, name)
		}
	}

	for _, name := range names {
		if onBase[name] && !all {
			continue
		}
		// R24: the branch first, else the base branch.
		rev := baseRev
		if hasBranch[name] {
			rev = "refs/heads/slice/" + name
		}
		info, err := ReadLog(root, rev, name)
		if err != nil {
			return Listing{}, err
		}
		result.Slices = append(result.Slices, info)
	}

	sortSlices(result.Slices)
	return result, nil
}

// sortSlices puts the slices in the order of R34: oldest created time
// first, the same time by name. Slices with no created time come last, by
// name.
func sortSlices(list []Info) {
	slices.SortFunc(list, func(a, b Info) int {
		switch {
		case a.Created.IsZero() && !b.Created.IsZero():
			return 1
		case !a.Created.IsZero() && b.Created.IsZero():
			return -1
		}
		if c := a.Created.Compare(b.Created); c != 0 {
			return c
		}
		return strings.Compare(a.Name, b.Name)
	})
}
