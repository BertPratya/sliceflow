package slicecmd

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/BertPratya/sliceflow/internal/cli/command"
	"github.com/BertPratya/sliceflow/internal/eventlog"
	"github.com/BertPratya/sliceflow/internal/git"
	"github.com/BertPratya/sliceflow/internal/slice"
)

// printCreateError prints the message for err, an error of slice.Create,
// to w, and returns the exit code: 2 for a title, a description, or a
// name that cannot be used, and 1 for any other failure. When a tool
// failed, its own text follows sliceflow's first line, word for word.
// When something that was made could not be removed, the message ends
// with what is left and the commands that remove it.
func printCreateError(w io.Writer, err error) int {
	var (
		control *slice.ControlCharacterError
		invalid *slice.InvalidNameError
		step    *slice.StepError
		done    *slice.CommittedError
		noBase  *slice.NoBaseError
		wrong   *slice.WrongBranchError
		taken   *slice.NameTakenError
		check   *slice.CheckError
	)
	switch {
	case errors.As(err, &done):
		printCommittedError(w, done)
	case errors.As(err, &control):
		fmt.Fprintf(w, controlCharacterLines, control.Option, slice.EscapeControlCharacters(control.Text))
		return 2
	case errors.As(err, &invalid):
		fmt.Fprintf(w, invalidNameLines, invalid.Given, invalid.Problem)
		return 2
	case errors.As(err, &step):
		printStepError(w, step)
	case command.IsStop(err):
		fmt.Fprintf(w, stoppedLine, notCreated+command.NothingChanged)
	case errors.Is(err, git.ErrNotFound):
		fmt.Fprint(w, gitNotFoundLines)
	case errors.Is(err, slice.ErrNoCommits):
		fmt.Fprint(w, noCommitsLines)
	case errors.As(err, &noBase):
		fmt.Fprintf(w, noBaseLine, noBase.Base)
		switch noBase.Source {
		case slice.BaseFromFlag:
			fmt.Fprint(w, flagBaseHint)
		case slice.BaseFromSetting:
			fmt.Fprint(w, settingBaseHint)
		default:
			fmt.Fprintf(w, defaultBaseHint, noBase.Name)
		}
	case errors.As(err, &wrong):
		if wrong.Branch == "" {
			fmt.Fprintf(w, noBranchLine, wrong.Base)
		} else {
			fmt.Fprintf(w, wrongBranchLine, wrong.Base, wrong.Branch)
		}
		fmt.Fprintf(w, switchToBaseLine, wrong.Base)
	case errors.As(err, &taken):
		if errors.Is(taken, slice.ErrFolderTaken) {
			fmt.Fprintf(w, folderTakenLine, taken.Name, taken.Name)
		} else {
			fmt.Fprintf(w, branchTakenLine, taken.Name, taken.Name)
		}
	case errors.As(err, &check):
		command.PrintFailure(w, check.Check, check.Err, command.NothingChanged)
	default:
		fmt.Fprintf(w, "Error: %v\n", err)
	}
	return 1
}

// printStepError prints the message for a step of slice create that
// failed: which step, or that the user stopped it, then the tool's own
// text, then what could not be removed, if anything. When the lock of
// the event log could not be taken, the message names the lock's file,
// not the log.
func printStepError(w io.Writer, step *slice.StepError) {
	ending := notCreated
	if len(step.Left) == 0 {
		ending += command.NothingChanged
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
	printLeftovers(w, step.Left)
}

// printCommittedError prints the message for a slice whose commit was
// made, but git or the user stopped it afterwards: the slice exists.
func printCommittedError(w io.Writer, done *slice.CommittedError) {
	if command.IsStop(done.Err) {
		fmt.Fprintf(w, stoppedAfterLine, done.Name, done.Base)
		return
	}
	command.PrintFailure(w, "finish the commit", done.Err, fmt.Sprintf(createdAnyway, done.Name, done.Base))
}

// leftoverKinds is the order in which leftovers are printed.
var leftoverKinds = []string{"folder", "staged"}

// printLeftovers prints what could not be removed, then how to remove
// each thing, both in the order of leftoverKinds. It prints nothing for
// an empty list.
func printLeftovers(w io.Writer, left []slice.Leftover) {
	if len(left) == 0 {
		return
	}
	fmt.Fprint(w, leftoversLine)
	for _, kind := range leftoverKinds {
		for _, thing := range left {
			if thing.Kind == kind {
				fmt.Fprintf(w, leftoverLine, thing.Kind, thing.Name)
			}
		}
	}
	fmt.Fprint(w, cleanUpLine)
	for _, kind := range leftoverKinds {
		for _, thing := range left {
			if thing.Kind != kind {
				continue
			}
			switch kind {
			case "folder":
				fmt.Fprintf(w, deleteFolderCmd, thing.Name)
			case "staged":
				fmt.Fprintf(w, unstageCmd, thing.Top, strings.TrimSuffix(thing.Name, "/"))
			}
		}
	}
}
