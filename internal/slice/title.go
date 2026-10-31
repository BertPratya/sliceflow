package slice

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/BertPratya/sliceflow/internal/eventlog"
	"github.com/BertPratya/sliceflow/internal/filelock"
	"github.com/BertPratya/sliceflow/internal/git"
)

// validateTitle returns the title to save: title without the white space
// at both ends (0002 R31). It returns an error when nothing is left.
func validateTitle(title string) (string, error) {
	trimmed := strings.TrimSpace(title)
	if trimmed == "" {
		return "", errors.New("the title is empty")
	}
	return trimmed, nil
}

// TitleResult is what SetTitle did. The cli prints it (0002 R29, R30).
type TitleResult struct {
	Name     string
	Title    string   // the saved title, without white space at the ends
	Warnings []string // warnings about the log, without the slice name
}

// SetTitle sets the title of the slice name (0002 R29-R35). dir is the
// current folder. It adds a slice.titled event to the log in the slice's
// worktree, and commits only the log, on the slice branch. The result
// holds the warnings also when it fails, so the caller can print them.
func SetTitle(dir, name, title string, now time.Time) (TitleResult, error) {
	root, err := openRepo(dir)
	if err != nil {
		return TitleResult{}, err
	}
	title, err = validateTitle(title)
	if err != nil {
		return TitleResult{}, unchanged(err)
	}
	worktree, err := findWorktree(root, name)
	if err != nil {
		return TitleResult{}, err
	}
	sliceDir := filepath.Join(worktree, "docs", "slices", name)

	data, err := eventlog.Data(eventlog.SliceTitled{Title: title})
	if err != nil {
		return TitleResult{}, unchanged(err)
	}
	e := eventlog.Event{Version: eventlog.Version, Event: eventlog.EventSliceTitled, At: now, Data: data}
	warnings, err := eventlog.AddEvent(sliceDir, e, name+": title set")
	res := TitleResult{Name: name, Title: title, Warnings: warnings}
	if err != nil {
		return res, titleFailed(name, sliceDir, err)
	}
	return res, nil
}

// findWorktree returns the worktree where the branch slice/<name> is
// checked out, or the error of 0002 R6 or R7.
func findWorktree(root, name string) (string, error) {
	branch := "slice/" + name
	if ValidateName(name) != nil {
		return "", noSliceError(name)
	}
	exists, err := git.BranchExists(root, branch)
	if err != nil {
		return "", unchanged(err)
	}
	if !exists {
		return "", noSliceError(name)
	}
	path, ok, err := git.WorktreeOf(root, branch)
	if err != nil {
		return "", unchanged(err)
	}
	if !ok {
		return "", fmt.Errorf("slice %s has no worktree: branch %s is not checked out anywhere.", name, branch)
	}
	return path, nil
}

// titleFailed turns an error of AddEvent into the user's message: busy
// (0002 R34), put back (R35), or nothing changed.
func titleFailed(name, sliceDir string, err error) error {
	if errors.Is(err, filelock.ErrBusy) {
		msg := fmt.Sprintf("the event log of %s is busy: another sliceflow is writing to it. Try again. Nothing was changed.", name)
		if filelock.StaleLockBlocks {
			msg += " If no sliceflow is running, delete " + eventlog.LockPath(sliceDir) + "."
		}
		return errors.New(msg)
	}
	if errors.Is(err, eventlog.ErrNotFound) {
		// Assumed: the spec has no text for this.
		return fmt.Errorf("slice %s has no event log in its worktree. Nothing was changed.", name)
	}

	var we *eventlog.WriteError
	if errors.As(err, &we) {
		reason, details := failReason(we.Err)
		msg := fmt.Sprintf("could not set the title of %s: %s.", name, reason)
		if we.Restored {
			msg += " The event log was put back as it was."
		} else {
			// Assumed: the spec has no text for this.
			msg += fmt.Sprintf(" Could not put the event log back: %v.", we.RestoreErr)
		}
		return withDetails(msg, details)
	}

	// The log was not changed yet, e.g. the exclude file or the read failed.
	reason, details := failReason(err)
	return withDetails(fmt.Sprintf("could not set the title of %s: %s. Nothing was changed.", name, reason), details)
}
