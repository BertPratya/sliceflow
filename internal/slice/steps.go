package slice

import (
	"context"
	"slices"
)

// Step is one thing a command makes, with the way to remove it again.
type Step struct {
	// Name says what the step does, in words that fit "could not ...",
	// e.g. "write /home/you/shop/docs/slices/login/spec.md".
	Name string
	// Do makes the thing. It may be stopped through ctx.
	Do func(ctx context.Context) error
	// Undo removes what Do made in this run, and only that. It also runs
	// when Do failed, so it must work after a Do that made only part of
	// the thing, or nothing. It is nil when there is nothing to remove.
	// It cannot be stopped, so it must not wait for the user.
	Undo func() error
	// Left is what to report when Undo fails.
	Left Leftover
}

// Leftover is one thing that was made but could not be removed.
type Leftover struct {
	Kind string // "folder" or "staged"
	// Name is the folder's full path for "folder", and the folder's
	// path in the project, written with / and ending in /, for "staged",
	// e.g. "docs/slices/login/".
	Name string
	// Top is the full path of the top folder of the user's checkout,
	// where a "staged" thing can be unstaged. It is "" for "folder".
	Top string
}

// StepError is returned when a step fails. By then, the failed step and
// the steps before it have been undone, newest first, and Left lists
// what could not be removed.
type StepError struct {
	Step string     // the Name of the step that failed
	Err  error      // why it failed
	Left []Leftover // what could not be removed, each once, newest first; empty when all was removed
}

func (e *StepError) Error() string {
	return e.Step + ": " + e.Err.Error()
}

// Unwrap returns why the step failed, so errors.Is and errors.As can look
// inside it, e.g. for a *git.Error.
func (e *StepError) Unwrap() error {
	return e.Err
}

// runSteps runs the Do of each step in order, and returns nil when all of
// them work. When one fails, the steps after it do not run, and the
// failed step and the steps before it are undone, as undoSteps does. It
// then returns a *StepError with the failed step's name, its error, and
// what could not be removed. Each Do gets ctx; an Undo runs to the end
// even when ctx is cancelled.
func runSteps(ctx context.Context, steps []Step) error {
	for i := range steps {
		if err := steps[i].Do(ctx); err != nil {
			return &StepError{
				Step: steps[i].Name,
				Err:  err,
				Left: undoSteps(steps[:i+1]),
			}
		}
	}
	return nil
}

// undoSteps runs the Undo of each step, newest first: a step with no
// Undo is skipped, and a failed Undo does not stop the others. It returns
// the Left of each step whose Undo failed, newest first, each one once.
func undoSteps(steps []Step) []Leftover {
	var left []Leftover
	for i := len(steps) - 1; i >= 0; i-- {
		if steps[i].Undo == nil {
			continue
		}
		if err := steps[i].Undo(); err != nil {
			left = addLeftover(left, steps[i].Left)
		}
	}
	return left
}

// addLeftover returns left with thing added at its end, unless left
// already has it.
func addLeftover(left []Leftover, thing Leftover) []Leftover {
	if slices.Contains(left, thing) {
		return left
	}
	return append(left, thing)
}
