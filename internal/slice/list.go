package slice

import (
	"bytes"
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"

	"github.com/BertPratya/sliceflow/internal/git"
)

// The reasons a candidate is not listed.
const (
	NotValidName   = "name"  // the name has a capital letter, or breaks the name rules
	NoLog          = "log"   // the folder has no log
	NoCreatedEvent = "event" // the log's first line is not a readable slice.created event
)

// ListResult is what List found: the slices, and the candidates it left
// out, both sorted by name, by character code, and whether the slices
// may be out of date.
type ListResult struct {
	Slices  []ListedSlice
	Skipped []SkippedCandidate
	// OldCopies is how many of the slices have a base branch whose
	// commits the user's checkout does not all hold, so they may be out
	// of date.
	OldCopies int
	// CheckFailed lists the base branches for which git could not tell
	// whether the checkout holds all their commits, sorted by name.
	CheckFailed []BaseCheckError
	// Branch is the branch the user's checkout is on, or "" when it is
	// on no branch. It is read only when OldCopies is not 0.
	Branch string
}

// ListedSlice is one slice of the list.
type ListedSlice struct {
	Name, Title, Description string
	base                     string // the base branch, from the slice.created event
}

// SkippedCandidate is a name that looked like a slice, but is not listed.
type SkippedCandidate struct {
	Name    string // as it was found
	Reason  string // NotValidName, NoLog, or NoCreatedEvent
	LogPath string // for NoCreatedEvent: where the log is in the project, written with /
}

// BaseCheckError says that git could not tell whether the user's
// checkout holds all the commits of the base branch Base.
type BaseCheckError struct {
	Base string
	Err  error // git's error, a *git.Error
}

// List finds the slices in the user's checkout, the checkout that dir is
// in. The candidates are the folders docs/slices/<name>/ in that
// checkout. A candidate is listed when its name is valid, its folder has
// the log .events.jsonl, and the log's first line is a slice.created
// event. Only that line is used. It never looks at other branches, other
// worktrees, or commits, and it changes nothing.
//
// It checks the project first, and returns ErrNoCommits, or a
// *CheckError when git fails, which holds git.ErrNotFound when git is not
// found. When git cannot tell whether some slices are out of date, List
// still returns the list, with CheckFailed set.
func List(ctx context.Context, dir string) (ListResult, error) {
	_, top, err := findCheckout(ctx, dir)
	if err != nil {
		return ListResult{}, err
	}
	names, err := sliceFolders(top)
	if err != nil {
		return ListResult{}, &CheckError{Check: "list the folders in docs/slices", Err: err}
	}

	var result ListResult
	for _, name := range names {
		if valid, err := ParseName(name); err != nil || valid != name {
			result.Skipped = append(result.Skipped, SkippedCandidate{Name: name, Reason: NotValidName})
			continue
		}
		data, found, err := readLog(top, name)
		if err != nil {
			return ListResult{}, &CheckError{Check: "read the log of " + name, Err: err}
		}
		if !found {
			result.Skipped = append(result.Skipped, SkippedCandidate{Name: name, Reason: NoLog})
			continue
		}
		first, _, _ := bytes.Cut(data, []byte("\n"))
		created, _, _, _, ok, err := readCreated(first)
		if err != nil || !ok {
			result.Skipped = append(result.Skipped, SkippedCandidate{Name: name, Reason: NoCreatedEvent, LogPath: logPath(name)})
			continue
		}
		result.Slices = append(result.Slices, ListedSlice{
			Name:        name,
			Title:       created.Title,
			Description: created.Description,
			base:        created.Base,
		})
	}

	if err := checkOldCopies(ctx, top, &result); err != nil {
		return ListResult{}, err
	}
	return result, nil
}

// sliceFolders returns the names of the folders in docs/slices/ of the
// checkout whose top folder is top, sorted by name, by character code.
// It returns none when docs/slices/ does not exist. Files, and links to
// folders, are left out.
func sliceFolders(top string) ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(top, "docs", "slices"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	slices.Sort(names)
	return names, nil
}

// checkOldCopies counts the slices of result whose base branch has
// commits that the checkout whose top folder is top does not hold. It
// asks git once per base branch. When git cannot answer for a base, that
// base goes in CheckFailed, and its slices are not counted. Any other
// failure is returned as a *CheckError.
func checkOldCopies(ctx context.Context, top string, result *ListResult) error {
	misses := map[string]bool{}
	var bases []string
	for _, s := range result.Slices {
		if _, seen := misses[s.base]; !seen {
			misses[s.base] = false
			bases = append(bases, s.base)
		}
	}
	slices.Sort(bases)
	for _, base := range bases {
		missing, err := missesBase(ctx, top, base)
		var gitErr *git.Error
		if errors.As(err, &gitErr) {
			result.CheckFailed = append(result.CheckFailed, BaseCheckError{Base: base, Err: err})
			continue
		}
		if err != nil {
			return &CheckError{Check: "check whether the slices are out of date", Err: err}
		}
		misses[base] = missing
	}
	for _, s := range result.Slices {
		if misses[s.base] {
			result.OldCopies++
		}
	}
	if result.OldCopies == 0 {
		return nil
	}
	branch, err := git.CheckedOutBranch(ctx, top)
	if err != nil {
		return &CheckError{Check: "read the current branch", Err: err}
	}
	result.Branch = branch
	return nil
}
