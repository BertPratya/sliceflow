package slice

import (
	"context"
	"errors"

	"github.com/BertPratya/sliceflow/internal/eventlog"
	"github.com/BertPratya/sliceflow/internal/git"
)

// StatusResult is what Status found about one slice.
type StatusResult struct {
	Name      string                // the slice name, in lowercase
	Created   eventlog.SliceCreated // the data of the slice.created event
	CreatedAt string                // when it happened, in UTC, written like 2026-11-02T09:15:30Z
	Events    int                   // the number of lines of the log that can be read
	LastEvent string                // the event name of the last of them
	LogPath   string                // where the log is in the project, written with /
	BadLines  []int                 // the numbers of the lines that could not be read, from 1
	// OldCopy is whether the user's checkout lacks a commit of the
	// slice's base branch, so the log may be out of date. It is false
	// when the base branch does not exist.
	OldCopy bool
	// OldCopyErr is git's error, a *git.Error, when git could not tell
	// whether the log is an old copy; OldCopy is then false.
	OldCopyErr error
	// Branch is the branch the user's checkout is on, or "" when it is
	// on no branch. It is read only when OldCopy is true.
	Branch string
}

// Status reads the slice with the given name from the user's checkout,
// the checkout that dir is in, and returns what its log says. It reads
// only docs/slices/<name>/.events.jsonl in that checkout, never another
// branch, another worktree, or a commit. It changes nothing.
//
// It checks, in this order, and stops at the first problem: the name
// (*InvalidNameError), and the project (ErrNoCommits, or a *CheckError
// when git fails, which holds git.ErrNotFound when git is not found).
// When the checkout has no log, the error is a *NoLogError; when the log
// has no slice.created event as its first line, a *NoCreatedEventError.
// When git cannot tell whether the log is an old copy, Status still
// returns the status, with OldCopyErr set.
func Status(ctx context.Context, dir, given string) (StatusResult, error) {
	name, err := ParseName(given)
	if err != nil {
		return StatusResult{}, &InvalidNameError{Given: given, Problem: err}
	}
	_, top, err := findCheckout(ctx, dir)
	if err != nil {
		return StatusResult{}, err
	}

	data, found, err := readLog(top, name)
	if err != nil {
		return StatusResult{}, &CheckError{Check: "read the log of " + name, Err: err}
	}
	if !found {
		branch, err := git.CheckedOutBranch(ctx, top)
		if err != nil {
			return StatusResult{}, &CheckError{Check: "read the current branch", Err: err}
		}
		return StatusResult{}, &NoLogError{Name: name, Branch: branch}
	}

	created, at, events, bad, ok, err := readCreated(data)
	if err != nil {
		return StatusResult{}, &CheckError{Check: "read the log of " + name, Err: err}
	}
	badLines := lineNumbers(bad)
	if !ok {
		return StatusResult{}, &NoCreatedEventError{Name: name, LogPath: logPath(name), BadLines: badLines}
	}
	result := StatusResult{
		Name:      name,
		Created:   created,
		CreatedAt: at,
		Events:    len(events),
		LastEvent: events[len(events)-1].Event,
		LogPath:   logPath(name),
		BadLines:  badLines,
	}

	result.OldCopy, err = missesBase(ctx, top, created.Base)
	var gitErr *git.Error
	if errors.As(err, &gitErr) {
		result.OldCopyErr = err
	} else if err != nil {
		return StatusResult{}, &CheckError{Check: "check whether " + name + " is out of date", Err: err}
	}
	if result.OldCopy {
		result.Branch, err = git.CheckedOutBranch(ctx, top)
		if err != nil {
			return StatusResult{}, &CheckError{Check: "read the current branch", Err: err}
		}
	}
	return result, nil
}

// findProject returns the main folder of the project that dir is in,
// once it knows the project has a commit.
func findProject(ctx context.Context, dir string) (string, error) {
	root, err := git.RepoRoot(ctx, dir)
	if err != nil {
		return "", &CheckError{Check: "find the project", Err: err}
	}
	hasCommits, err := git.HasCommits(ctx, root)
	if err != nil {
		return "", &CheckError{Check: "read the commits", Err: err}
	}
	if !hasCommits {
		return "", ErrNoCommits
	}
	return root, nil
}

// lineNumbers returns the numbers of the bad lines, in order.
func lineNumbers(bad []eventlog.BadLine) []int {
	var numbers []int
	for _, line := range bad {
		numbers = append(numbers, line.Line)
	}
	return numbers
}
