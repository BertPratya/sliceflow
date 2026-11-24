package slice

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"

	"github.com/BertPratya/sliceflow/internal/state"
)

// The reasons a candidate is not listed.
const (
	NotValidName   = "name"  // the name has a capital letter, or breaks the name rules
	NoLog          = "log"   // the folder has no log
	NoCreatedEvent = "event" // the log's first line is not a readable slice.created event
)

// ListResult is what List found: the slices, and the candidates it left
// out, both sorted by name, by character code.
type ListResult struct {
	Slices  []ListedSlice
	Skipped []SkippedCandidate
}

// ListedSlice is one slice of the list.
type ListedSlice struct {
	Name, Title, Description string
}

// SkippedCandidate is a name that looked like a slice, but is not listed.
type SkippedCandidate struct {
	Name    string // as it was found
	Reason  string // NotValidName, NoLog, or NoCreatedEvent
	LogPath string // for NoCreatedEvent: where the log is in the project, written with /
}

// List finds the slices in the user's checkout, the checkout that dir is
// in. The candidates are the folders docs/slices/<name>/ in that
// checkout. A candidate is listed when its name is valid, its folder has
// the log .events.jsonl, and the log's first line is a slice.created
// event. A listed slice has its current title and description: the ones
// from slice.created, changed by each slice.edited after it that can be
// read. Other lines that cannot be read are skipped with no warning, and
// the state is not worked out. It never looks at other branches, other
// worktrees, or commits, and does not check whether another branch has
// newer copies. It changes nothing.
//
// It checks the project first, and returns ErrNoCommits, or a
// *CheckError when git fails, which holds git.ErrNotFound when git is not
// found.
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
		_, _, events, _, ok, err := readCreated(data)
		if err != nil || !ok {
			result.Skipped = append(result.Skipped, SkippedCandidate{Name: name, Reason: NoCreatedEvent, LogPath: logPath(name)})
			continue
		}
		detail := state.Replay(events).Detail
		result.Slices = append(result.Slices, ListedSlice{
			Name:        name,
			Title:       detail.Title,
			Description: detail.Description,
		})
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
