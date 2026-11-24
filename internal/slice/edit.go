package slice

import (
	"context"
	"os"
	"path/filepath"

	"github.com/BertPratya/sliceflow/internal/eventlog"
	"github.com/BertPratya/sliceflow/internal/git"
	"github.com/BertPratya/sliceflow/internal/state"
)

// EditOptions is what Edit changes. A nil field is not changed.
type EditOptions struct {
	Name        string  // the name as the user gave it
	Title       *string // the new title; "" sets it back to the slice name
	Description *string // the new description; "" removes it
}

// EditResult says what Edit found and did.
type EditResult struct {
	Name        string // the slice name, in lowercase
	Base        string // the base branch, where the commit was made
	Title       string // the title the slice has now
	Description string // the description the slice has now, or ""
	Changed     bool   // false when no value differed, so nothing was written
	// LogPath is where the log is in the project, written with /, and
	// BadLines are the numbers of its lines that could not be read. They
	// are set as soon as the log was read, even when Edit then fails, so
	// the caller can warn about the log first.
	LogPath  string
	BadLines []int
}

// Edit changes the title, the description, or both, of the slice with
// the given name, in the user's checkout, the checkout that dir is in,
// which must be on the slice's base branch. It adds one slice.edited
// event to the log, holding only the values that differ from the
// current ones, and makes one commit of only the log on the base
// branch. The user's other files, what is staged, and the checked-out
// branch stay as they were. When no value differs, it writes nothing
// and returns a result with Changed false.
//
// It checks, in this order, and stops at the first problem: a control
// character in the title, then in the description
// (*ControlCharacterError), no value to change (ErrNothingToEdit), the
// name (*InvalidNameError), the project (ErrNoCommits), the log
// (*NoLogError), its first line (*NoCreatedEventError), the checkout on
// the base branch (*NotOnBaseError), and the state (*NoStateError). When
// a check cannot be done, e.g. because git fails, the error is a
// *CheckError. When writing the event or the commit fails before the
// commit is made, the log is put back as it was, and the error is a
// *StepError that lists what could not be put back. When the commit was
// made, but git or the user stopped it afterwards, nothing is put back,
// and the error is a *CommittedError. A stop through ctx holds ctx's
// error.
func Edit(ctx context.Context, dir string, opts EditOptions) (EditResult, error) {
	if opts.Title != nil && HasControlCharacter(*opts.Title) {
		return EditResult{}, &ControlCharacterError{Option: "title", Text: *opts.Title}
	}
	if opts.Description != nil && HasControlCharacter(*opts.Description) {
		return EditResult{}, &ControlCharacterError{Option: "description", Text: *opts.Description}
	}
	if opts.Title == nil && opts.Description == nil {
		return EditResult{}, ErrNothingToEdit
	}
	name, err := ParseName(opts.Name)
	if err != nil {
		return EditResult{}, &InvalidNameError{Given: opts.Name, Problem: err}
	}
	_, top, err := findCheckout(ctx, dir)
	if err != nil {
		return EditResult{}, err
	}

	data, found, err := readLog(top, name)
	if err != nil {
		return EditResult{}, &CheckError{Check: "read the log of " + name, Err: err}
	}
	if !found {
		branch, err := git.CheckedOutBranch(ctx, top)
		if err != nil {
			return EditResult{}, &CheckError{Check: "read the current branch", Err: err}
		}
		return EditResult{}, &NoLogError{Name: name, Branch: branch}
	}
	created, _, events, bad, ok, err := readCreated(data)
	if err != nil {
		return EditResult{}, &CheckError{Check: "read the log of " + name, Err: err}
	}
	if !ok {
		return EditResult{}, &NoCreatedEventError{Name: name, LogPath: logPath(name), BadLines: lineNumbers(bad)}
	}
	result := EditResult{Name: name, Base: created.Base, LogPath: logPath(name), BadLines: lineNumbers(bad)}

	branch, err := git.CheckedOutBranch(ctx, top)
	if err != nil {
		return result, &CheckError{Check: "read the current branch", Err: err}
	}
	if branch != created.Base {
		return result, &NotOnBaseError{Name: name, Base: created.Base, Branch: branch}
	}
	s := state.Replay(events)
	if s.Unknown != nil {
		return result, &NoStateError{Name: name, LogPath: logPath(name), State: s}
	}

	edited := changedValues(name, s.Detail, opts)
	result.Title, result.Description = s.Detail.Title, s.Detail.Description
	if edited.Title != nil {
		result.Title = *edited.Title
	}
	if edited.Description != nil {
		result.Description = *edited.Description
	}
	if edited.Title == nil && edited.Description == nil {
		return result, nil
	}

	lockPath, err := git.GitPath(ctx, top, "sliceflow/locks/"+name+".lock")
	if err != nil {
		return result, &CheckError{Check: "find git's folder", Err: err}
	}
	before, err := git.BranchCommit(ctx, top, created.Base)
	if err != nil {
		return result, &CheckError{Check: "read the branch " + created.Base, Err: err}
	}

	logFile := filepath.Join(top, filepath.FromSlash(logPath(name)))
	steps, err := editSteps(top, logFile, lockPath, edited)
	if err != nil {
		return result, &CheckError{Check: "read " + logFile, Err: err}
	}
	if err := runSteps(ctx, steps); err != nil {
		return result, err
	}
	if err := git.Commit(ctx, top, editMessage(name, edited), logPath(name)); err != nil {
		return result, editCommitFailed(top, name, created.Base, before, steps, err)
	}
	result.Changed = true
	return result, nil
}

// changedValues returns the event data for opts: each value that is
// given and differs from current. An empty title means the slice name.
func changedValues(name string, current state.Detail, opts EditOptions) eventlog.SliceEdited {
	var edited eventlog.SliceEdited
	if opts.Title != nil {
		title := *opts.Title
		if title == "" {
			title = name
		}
		if title != current.Title {
			edited.Title = &title
		}
	}
	if opts.Description != nil && *opts.Description != current.Description {
		description := *opts.Description
		edited.Description = &description
	}
	return edited
}

// editMessage returns the message of the commit that edits the slice
// name, after what edited holds.
func editMessage(name string, edited eventlog.SliceEdited) string {
	switch {
	case edited.Title != nil && edited.Description != nil:
		return name + ": edit the title and description"
	case edited.Title != nil:
		return name + ": edit the title"
	default:
		return name + ": edit the description"
	}
}

// editSteps returns the one step of Edit: append edited to the log at
// logFile, in the checkout whose top folder is top, under the lock at
// lockPath. Its Undo cuts the log back to the
// size it has now, which removes the line the step added, and the line
// break it may have added before it. The Undo never makes the log longer.
func editSteps(top, logFile, lockPath string, edited eventlog.SliceEdited) ([]Step, error) {
	info, err := os.Stat(logFile)
	if err != nil {
		return nil, err
	}
	size := info.Size()
	return []Step{{
		Name: "write " + logFile,
		Do: func(ctx context.Context) error {
			return eventlog.Append(ctx, logFile, lockPath, eventlog.SliceEditedEvent, edited)
		},
		Undo: func() error {
			info, err := os.Stat(logFile)
			if err != nil {
				return err
			}
			if info.Size() <= size {
				return nil // nothing was added: a cut must never make the log longer
			}
			return os.Truncate(logFile, size)
		},
		Left: Leftover{Kind: "log", Name: logFile, Top: top},
	}}, nil
}

// editCommitFailed returns the error of Edit when the commit of the log
// of the slice name failed with err, in the checkout whose top folder is
// top, after steps added the event. before is the commit base pointed
// to before the commit.
//
// When base now points to another commit, the commit was made: it puts
// nothing back and returns a *CommittedError. Otherwise it unstages the
// log, undoes steps, and returns a *StepError for the step "make the
// commit". When it cannot tell whether the commit was made, it puts the
// log back, as if it was not. It never waits for the user: ctx may
// already be cancelled.
func editCommitFailed(top, name, base, before string, steps []Step, err error) error {
	ctx := context.Background()
	after, checkErr := git.BranchCommit(ctx, top, base)
	if checkErr == nil && after != before {
		return &CommittedError{Name: name, Base: base, Done: "edited", Err: err}
	}

	var left []Leftover
	if unstageErr := git.Unstage(ctx, top, logPath(name)); unstageErr != nil {
		left = append(left, Leftover{Kind: "staged", Name: logPath(name), Top: top})
	}
	for _, thing := range undoSteps(steps) {
		left = addLeftover(left, thing)
	}
	return &StepError{Step: "make the commit", Err: err, Left: left}
}
