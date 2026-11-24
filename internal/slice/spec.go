package slice

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/BertPratya/sliceflow/internal/git"
)

// SpecResult says where a slice's spec is.
type SpecResult struct {
	Name string // the slice name, in lowercase
	Path string // the full path of docs/slices/<name>/spec.md in the user's checkout
}

// SpecPath finds the spec of the slice with the given name in the user's
// checkout, the checkout that dir is in. It needs the slice's log only
// to exist there: it does not read it, so it works whatever the log
// holds. It does not check that the spec exists. It changes nothing.
//
// It checks, in this order, and stops at the first problem: the name
// (*InvalidNameError), the project (ErrNoCommits, or a *CheckError when
// git fails, which holds git.ErrNotFound when git is not found), and the
// log (*NoLogError when the checkout has none).
func SpecPath(ctx context.Context, dir, given string) (SpecResult, error) {
	name, err := ParseName(given)
	if err != nil {
		return SpecResult{}, &InvalidNameError{Given: given, Problem: err}
	}
	_, top, err := findCheckout(ctx, dir)
	if err != nil {
		return SpecResult{}, err
	}

	_, err = os.Stat(filepath.Join(top, filepath.FromSlash(logPath(name))))
	if errors.Is(err, fs.ErrNotExist) {
		branch, err := git.CheckedOutBranch(ctx, top)
		if err != nil {
			return SpecResult{}, &CheckError{Check: "read the current branch", Err: err}
		}
		return SpecResult{}, &NoLogError{Name: name, Branch: branch}
	}
	if err != nil {
		return SpecResult{}, &CheckError{Check: "read the log of " + name, Err: err}
	}
	return SpecResult{Name: name, Path: filepath.Join(top, "docs", "slices", name, "spec.md")}, nil
}
