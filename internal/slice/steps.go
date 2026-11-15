package slice

import (
	"context"
)

// Step is one thing a command makes, with the way to remove it again.
type Step struct {
	// Name says what the step does, in words that fit "could not ...",
	// e.g. "create the branch".
	Name string
	// Do makes the thing. It may be stopped through ctx.
	Do func(ctx context.Context) error
	// Undo removes what Do made. It is nil when there is nothing to
	// remove, e.g. a file inside a worktree, which goes with the
	// worktree. It cannot be stopped, so it must not wait for the user.
	Undo func() error
	// Left is what to report when Undo fails.
	Left Leftover
}

// Leftover is one thing that was made but could not be removed.
type Leftover struct {
	Kind string // "worktree" or "branch"
	Name string // the worktree's full path, or the branch's name
}

// StepError is returned when a step fails. By then, the steps before it
// have been undone, newest first, and Left lists what could not be
// removed.
type StepError struct {
	Step string     // the Name of the step that failed
	Err  error      // why it failed
	Left []Leftover // what could not be removed, newest first; empty when all was removed
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
// them work. When one fails, the steps after it do not run, and the steps
// before it are undone, newest first: a step with no Undo is skipped, and
// a failed Undo does not stop the others. It then returns a *StepError
// with the failed step's name, its error, and the Left of each step whose
// Undo failed. Each Do gets ctx; an Undo runs to the end even when ctx is
// cancelled.
func runSteps(ctx context.Context, steps []Step) error {
	for i := range steps {
		if doErr := steps[i].Do(ctx); doErr != nil {
			var leftovers []Leftover

			for j := i - 1; j >= 0; j-- {
				if steps[j].Undo == nil {
					continue
				}

				if err := steps[j].Undo(); err != nil {
					leftovers = append(leftovers, steps[j].Left)
				}
			}

			return &StepError{
				Step: steps[i].Name,
				Err:  doErr,
				Left: leftovers,
			}
		}
	}

	return nil
}
