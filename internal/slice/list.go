package slice

import (
	"context"
	"sort"
	"strings"

	"github.com/BertPratya/sliceflow/internal/git"
)

// SliceList is the result of List. Slices are in the order of R34. Warnings
// are about the whole repo (R23, R43). Neither is nil.
type SliceList struct {
	Slices   []SliceStatus
	Warnings []Warning
}

// List returns the slices in progress, or all slices when all is true
// (Flow 5).
func List(ctx context.Context, r Repo, all bool) (SliceList, error) {
	res := SliceList{Slices: []SliceStatus{}, Warnings: []Warning{}}
	base, exists, w, err := baseWarning(ctx, r)
	if err != nil {
		return SliceList{}, err
	}
	if w != nil {
		res.Warnings = append(res.Warnings, *w)
	}

	out, err := git.Run(ctx, r.Dir, "for-each-ref", "--format=%(refname)", "refs/heads/slice/")
	if err != nil {
		return SliceList{}, err
	}
	var branches []string // valid names, in git's order
	for _, ref := range strings.Split(out, "\n") {
		ref = strings.TrimSpace(ref)
		x, ok := strings.CutPrefix(ref, "refs/heads/slice/")
		if !ok || x == "" {
			continue
		}
		// R43: a bad name is skipped, and never reaches another git command.
		if CheckName(x) != nil {
			res.Warnings = append(res.Warnings, Warning{Text: "branch slice/" + x + ": not a valid slice name; skipped"})
			continue
		}
		branches = append(branches, x)
	}

	merged := map[string]bool{}
	if exists {
		// -z: git does not quote odd names. A name that breaks R2 is no slice.
		// --full-tree: the path is from the top of the repo, not the current folder.
		out, err := git.Run(ctx, r.Dir, "ls-tree", "--full-tree", "-d", "--name-only", "-z", "refs/heads/"+base, "docs/slices/")
		if err != nil {
			return SliceList{}, err
		}
		for _, p := range strings.Split(out, "\x00") {
			if n, ok := strings.CutPrefix(p, "docs/slices/"); ok && CheckName(n) == nil {
				merged[n] = true
			}
		}
	}

	names := []string{}
	seen := map[string]bool{}
	for _, n := range branches {
		if !merged[n] {
			names = append(names, n)
			seen[n] = true
		}
	}
	if all {
		for n := range merged {
			if !seen[n] {
				names = append(names, n)
			}
		}
	}

	for _, n := range names {
		st, found, err := readSlice(ctx, r, n, base, exists)
		if err != nil {
			return SliceList{}, err
		}
		if found {
			res.Slices = append(res.Slices, st)
		}
	}
	// R34: oldest first; the same time by name; no created time last, by name.
	sort.SliceStable(res.Slices, func(i, j int) bool {
		a, b := res.Slices[i], res.Slices[j]
		switch {
		case a.Created.IsZero() != b.Created.IsZero():
			return !a.Created.IsZero()
		case !a.Created.Equal(b.Created):
			return a.Created.Before(b.Created)
		}
		return a.Slice < b.Slice
	})
	return res, nil
}
