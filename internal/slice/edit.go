package slice

import (
	"bytes"
	"context"
	"errors"
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
// lockPath.
//
// Its Undo removes the line the step added, and the line break it may
// have added before it, by cutting the log back to the size it has now.
// It cuts only when what follows that size is exactly one line, the
// event this step wrote. When the lock was never taken, Append wrote
// nothing, so the Undo does nothing. When anything else follows, e.g. a
// line another process added under the lock, the Undo cuts nothing and
// fails, so the log is reported as left over: it never removes a line
// that is not its own, and never makes the log longer.
func editSteps(top, logFile, lockPath string, edited eventlog.SliceEdited) ([]Step, error) {
	info, err := os.Stat(logFile)
	if err != nil {
		return nil, err
	}
	size := info.Size()
	locked := false
	return []Step{{
		Name: "write " + logFile,
		Do: func(ctx context.Context) error {
			err := eventlog.Append(ctx, logFile, lockPath, eventlog.SliceEditedEvent, edited)
			var lockErr *eventlog.LockError
			locked = !errors.As(err, &lockErr)
			return err
		},
		Undo: func() error {
			if !locked {
				return nil
			}
			data, err := os.ReadFile(logFile)
			if err != nil {
				return err
			}
			if int64(len(data)) <= size {
				return nil
			}
			if !onlyOurLine(data[size:], edited) {
				return errNotOurLine
			}
			return os.Truncate(logFile, size)
		},
		Left: Leftover{Kind: "log", Name: logFile, Top: top},
	}}, nil
}

// errNotOurLine is the Undo's error when the log holds more than the one
// line the step added, so it cannot be cut back safely.
var errNotOurLine = errors.New("slice: the log holds lines that this edit did not add")

// onlyOurLine reports whether tail, what follows the log's old size, is
// one whole line, after the line break Append may add first, holding a
// slice.edited event with exactly the data edited.
func onlyOurLine(tail []byte, edited eventlog.SliceEdited) bool {
	tail = bytes.TrimPrefix(tail, []byte("\n"))
	if bytes.Count(tail, []byte("\n")) != 1 || !bytes.HasSuffix(tail, []byte("\n")) {
		return false
	}
	events, bad, err := eventlog.ReadFrom(bytes.NewReader(tail))
	if err != nil || len(bad) != 0 || len(events) != 1 || events[0].Event != eventlog.SliceEditedEvent {
		return false
	}
	got, ok := events[0].Payload.(eventlog.SliceEdited)
	return ok && sameText(got.Title, edited.Title) && sameText(got.Description, edited.Description)
}

// sameText reports whether a and b are both nil, or point to the same
// text.
func sameText(a, b *string) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
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
