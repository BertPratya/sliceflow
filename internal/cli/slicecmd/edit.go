package slicecmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/BertPratya/sliceflow/internal/cli/command"
	"github.com/BertPratya/sliceflow/internal/eventlog"
	"github.com/BertPratya/sliceflow/internal/slice"
	"github.com/BertPratya/sliceflow/internal/state"
)

// runSliceEdit changes the title, the description, or both, of the
// slice named in in, in the user's checkout of the current folder. When
// it works, it prints the title and description the slice has now, or
// that nothing had to change, and returns 0. Lines of the log that
// cannot be read are named in a warning first, on stderr. Otherwise it
// prints what went wrong, and returns 2 for a command line, a title, a
// description, or a name that cannot be used, or 1 for any other
// failure.
func runSliceEdit(ctx context.Context, in command.Input, stdout, stderr io.Writer) int {
	dir, err := os.Getwd()
	if err != nil {
		command.PrintFailure(stderr, "find the current folder", err, command.NothingChanged)
		return 1
	}

	opts := slice.EditOptions{Name: in.Args["name"]}
	if title, ok := in.Options["title"]; ok {
		opts.Title = &title
	}
	if description, ok := in.Options["description"]; ok {
		opts.Description = &description
	}
	result, err := slice.Edit(ctx, dir, opts)
	if result.LogPath != "" {
		command.PrintBadLines(stderr, result.LogPath, result.BadLines)
	}
	if err != nil {
		return printEditError(stderr, err)
	}

	if !result.Changed {
		fmt.Fprintf(stdout, nothingToChange, result.Name)
		return 0
	}
	description := result.Description
	if description == "" {
		description = noneValue
	}
	fmt.Fprintf(stdout, editedLine, result.Name)
	fmt.Fprintf(stdout, editValueLine, "Title:", result.Title)
	fmt.Fprintf(stdout, editValueLine, "Description:", description)
	return 0
}

// printEditError prints the message for err, an error of slice.Edit, and
// returns the exit code.
func printEditError(w io.Writer, err error) int {
	var (
		done      *slice.CommittedError
		step      *slice.StepError
		notOnBase *slice.NotOnBaseError
		noState   *slice.NoStateError
	)
	switch {
	case errors.Is(err, slice.ErrNothingToEdit):
		fmt.Fprint(w, needsAnOptionLines)
		return 2
	case errors.As(err, &done):
		printCommittedError(w, done)
	case errors.As(err, &step):
		printEditStepError(w, step)
	case errors.As(err, &notOnBase):
		if notOnBase.Branch == "" {
			fmt.Fprintf(w, notOnBranchLine, notOnBase.Name, notOnBase.Base)
		} else {
			fmt.Fprintf(w, notOnBaseLine, notOnBase.Name, notOnBase.Base, notOnBase.Branch)
		}
		fmt.Fprintf(w, switchToBaseLine, notOnBase.Base)
	case errors.As(err, &noState):
		command.PrintStateWarning(w, noState.LogPath, noState.State)
		fmt.Fprintf(w, noStateLine, noState.Name)
		if noState.State.Unknown.Reason == state.ReasonNotAllowed {
			fmt.Fprintf(w, logHistoryHint, noState.LogPath)
		} else {
			fmt.Fprint(w, updateHint)
		}
	case command.IsStop(err):
		fmt.Fprintf(w, stoppedLine, command.NothingChanged)
	default:
		return PrintSliceError(w, err)
	}
	return 1
}

// printEditStepError prints the message for a step of slice edit that
// failed: which step, or that the user stopped it, then the tool's own
// text, then what could not be put back, if anything. When the lock of
// the log could not be taken, the message names the lock's file.
func printEditStepError(w io.Writer, step *slice.StepError) {
	ending := ""
	if len(step.Left) == 0 {
		ending = command.NothingChanged
	}
	what := step.Step
	var lockErr *eventlog.LockError
	if errors.As(step.Err, &lockErr) {
		what = "write " + lockErr.Path
	}
	if command.IsStop(step.Err) {
		fmt.Fprintf(w, stoppedLine, ending)
	} else {
		command.PrintFailure(w, what, step.Err, ending)
	}
	printEditLeftovers(w, step.Left)
}

// printEditLeftovers prints what slice edit could not put back: the log
// that still holds the added line, then the log that is still staged,
// and the one command that puts both back. It prints nothing for an
// empty list.
func printEditLeftovers(w io.Writer, left []slice.Leftover) {
	if len(left) == 0 {
		return
	}
	fmt.Fprint(w, leftoversLine)
	var top, logInProject string
	for _, kind := range []string{"log", "staged"} {
		for _, thing := range left {
			if thing.Kind != kind {
				continue
			}
			top = thing.Top
			if kind == "log" {
				fmt.Fprintf(w, logLeftoverLine, thing.Name)
			} else {
				fmt.Fprintf(w, logStagedLine, thing.Name)
				logInProject = thing.Name
			}
		}
	}
	if logInProject == "" {
		logInProject = projectPath(top, left)
	}
	fmt.Fprint(w, cleanUpLine)
	fmt.Fprintf(w, restoreLogCmd, top, logInProject)
}

// projectPath returns the path in the project, written with /, of the
// log among left, from its full path and top, the checkout's top folder.
func projectPath(top string, left []slice.Leftover) string {
	for _, thing := range left {
		if thing.Kind != "log" {
			continue
		}
		if rel, err := filepath.Rel(top, thing.Name); err == nil {
			return filepath.ToSlash(rel)
		}
		return thing.Name
	}
	return ""
}
