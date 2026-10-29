package slice

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/BertPratya/sliceflow/internal/eventlog"
	"github.com/BertPratya/sliceflow/internal/git"
	"github.com/BertPratya/sliceflow/internal/state"
)

// ErrNoSlice means no slice has the given name (R33). Use errors.Is to tell
// it from a git failure.
var ErrNoSlice = errors.New("no slice")

// noSliceError returns the R33 error for name. It wraps ErrNoSlice, and its
// text is the message of R33, made here only.
func noSliceError(name string) error {
	return &noSlice{name: name}
}

// noSlice is the error of noSliceError.
type noSlice struct{ name string }

// Error returns the text of R33.
func (e *noSlice) Error() string {
	return fmt.Sprintf("no slice named %s. See all slices: sliceflow slice list", e.name)
}

// Is makes errors.Is(err, ErrNoSlice) true.
func (e *noSlice) Is(target error) bool { return target == ErrNoSlice }

// MissingBaseWarning returns the R23 warning for a base branch that does
// not exist. source is as in R16, e.g. "the default". Status and list both
// use it.
func MissingBaseWarning(base, source string) string {
	return fmt.Sprintf("base branch %s does not exist (%s); merged slices are not shown", base, source)
}

// Info is what sliceflow knows about one slice, read from its event log
// (R29, R30).
type Info struct {
	Name  string
	Title string // the saved title, or "" when there is none
	Phase string // e.g. "planning", or "unknown" (0002 R23)
	State string // e.g. "spec_draft"; empty when Phase is "unknown"
	// From is the slice's base branch and commit. It is nil when the log
	// does not have them.
	From *eventlog.Ref
	// Created is when the slice was created, in UTC. It is the zero time
	// when the log does not have it.
	Created time.Time
	// Warnings are the warnings about this slice, without its name.
	Warnings []string
	// RepoWarnings are the warnings about the whole repo (R23). Status
	// fills it; ReadInfo does not.
	RepoWarnings []string
}

// Status finds the slice and reads its status (R21-R25, R29-R33). When it
// fails after the base branch was checked, the returned Info still holds
// RepoWarnings, so the caller can print them with the error. dir is
// the current folder. When name is empty, the slice is the one of the
// current branch slice/<name> (R31). It returns an error that ends with
// "Nothing was changed." outside a repo or in a repo with no commits (R22).
func Status(dir, name string) (Info, error) {
	root, err := openRepo(dir)
	if err != nil {
		return Info{}, err
	}
	base, baseExists, repoWarnings, err := loadBase(root)
	if err != nil {
		return Info{}, err
	}

	if name == "" {
		branch, err := git.CurrentBranch(root)
		if err != nil {
			return Info{RepoWarnings: repoWarnings}, err
		}
		rest, onSlice := strings.CutPrefix(branch, "slice/")
		if !onSlice {
			where := "on no branch"
			if branch != "" {
				where = "on " + branch
			}
			return Info{RepoWarnings: repoWarnings}, fmt.Errorf("not on a slice branch (you are %s).\n"+
				"Give a name: sliceflow slice status <name>\n"+
				"See all slices: sliceflow slice list", where)
		}
		name = rest
	}

	info, err := ReadInfo(root, base, baseExists, name)
	if err != nil {
		return Info{RepoWarnings: repoWarnings}, err
	}
	info.RepoWarnings = repoWarnings
	return info, nil
}

// loadBase reads the base branch (R21) and checks that it exists. When it
// does not, repoWarnings holds the R23 warning. Status and list both start
// with it.
func loadBase(root string) (base string, exists bool, repoWarnings []string, err error) {
	base, source, err := readBase(root)
	if err != nil {
		return "", false, nil, err
	}
	exists, err = git.BranchExists(root, base)
	if err != nil {
		return "", false, nil, err
	}
	if !exists {
		repoWarnings = append(repoWarnings, MissingBaseWarning(base, source))
	}
	return base, exists, repoWarnings, nil
}

// readBase returns the base branch for status and list (R21): the value of
// git config sliceflow.base, else main. source says where it came from, as
// R16 does. It does not check that the branch exists.
func readBase(root string) (branch, source string, err error) {
	value, ok, err := git.Config(root, "sliceflow.base")
	if err != nil {
		return "", "", err
	}
	// An empty value counts as not set.
	if ok && value != "" {
		return value, "from git config sliceflow.base", nil
	}
	return "main", "the default", nil
}

// ReadInfo reads one slice (R24, R25). root is the repo folder, base is the
// base branch, and baseExists says whether that branch exists. A slice
// exists when its name is valid and the local branch slice/<name> exists,
// or the folder docs/slices/<name>/ is on the base branch. Otherwise the
// error wraps ErrNoSlice (R33). The log is read as ReadLog does.
func ReadInfo(root, base string, baseExists bool, name string) (Info, error) {
	if ValidateName(name) != nil {
		return Info{}, noSliceError(name)
	}

	folder := "docs/slices/" + name
	rev := "refs/heads/slice/" + name
	hasBranch, err := git.BranchExists(root, "slice/"+name)
	if err != nil {
		return Info{}, err
	}
	if !hasBranch {
		if !baseExists {
			return Info{}, noSliceError(name)
		}
		rev = "refs/heads/" + base
		onBase, err := git.PathExists(root, rev, folder)
		if err != nil {
			return Info{}, err
		}
		if !onBase {
			return Info{}, noSliceError(name)
		}
	}
	return ReadLog(root, rev, name)
}

// ReadLog reads the log of a slice that is known to exist (R24, R25). rev
// is the newest commit to read from: refs/heads/slice/<name> when that
// branch exists, else the base branch. Files that are not committed are
// never read. A log that is missing or has bad lines gives warnings in
// Info, not an error.
func ReadLog(root, rev, name string) (Info, error) {
	info := Info{Name: name, Phase: "unknown"}
	content, found, err := git.ShowFile(root, rev, "docs/slices/"+name+"/.events.jsonl")
	if err != nil {
		return Info{}, err
	}
	if !found {
		info.Warnings = []string{eventlog.NotFound}
		return info, nil
	}

	log := eventlog.Parse(content)
	info.Warnings = log.Warnings
	place := state.Replay(log.Events)
	info.Phase, info.State = place.Phase, place.State
	if ev, data, ok := log.Created(); ok {
		from := data.From
		info.From = &from
		info.Title = data.Title
		info.Created = ev.At
	}
	return info, nil
}
