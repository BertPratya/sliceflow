package slice

import (
	"context"
	"os"
	"path/filepath"

	"github.com/BertPratya/sliceflow/internal/spec"
)

// Validation is the result of Validate: the problems in line order, and
// how many there are of each level.
type Validation struct {
	Problems                []spec.Problem
	Errors, Warnings, Notes int
}

// Validate checks docs/slices/<name>/spec.md in the slice's worktree, as it
// is on disk (R13). An empty name means the slice of the current branch.
// It writes nothing and takes no lock. strict only changes the exit code
// (R14), which the caller decides from the counts.
func Validate(ctx context.Context, r Repo, name string, strict bool) (Validation, error) {
	if name == "" {
		n, err := currentSliceName(ctx, r, "spec validate")
		if err != nil {
			return Validation{}, err
		}
		name = n
	}
	wt, err := openSlice(ctx, r, name)
	if err != nil {
		return Validation{}, err
	}
	data, err := os.ReadFile(filepath.Join(wt, "docs", "slices", name, "spec.md"))
	if os.IsNotExist(err) {
		return Validation{}, specNotFoundError(r, wt, name)
	}
	if err != nil {
		return Validation{}, err
	}
	v := Validation{Problems: spec.Check(spec.Parse(data), name)}
	for _, p := range v.Problems {
		switch p.Level {
		case spec.LevelError:
			v.Errors++
		case spec.LevelWarning:
			v.Warnings++
		default:
			v.Notes++
		}
	}
	return v, nil
}

// specNotFoundError is R19's error: spec.md is missing in the worktree wt.
// The path is from the root of the main checkout, with "/" on every system
// (R41 of 0001). A worktree outside the main checkout gives a path that
// starts with "../". Task 14 (spec submit) uses it too.
func specNotFoundError(r Repo, wt, name string) *Error {
	p := filepath.Join(wt, "docs", "slices", name, "spec.md")
	if rel, err := filepath.Rel(filepath.FromSlash(r.Root), p); err == nil {
		p = rel
	}
	return &Error{Message: filepath.ToSlash(p) + " not found. Nothing was changed."}
}
