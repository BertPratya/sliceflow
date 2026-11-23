package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"strings"

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
	case isStop(err):
		fmt.Fprintf(w, stoppedLine, notCreated+nothingChanged)
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
		printFailure(w, check.Check, check.Err, nothingChanged)
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
		ending += nothingChanged
	}
	what := step.Step
	var lockErr *eventlog.LockError
	if errors.As(step.Err, &lockErr) {
		what = "write " + lockErr.Path
	}
	if isStop(step.Err) {
		fmt.Fprintf(w, stoppedLine, ending)
	} else {
		printFailure(w, what, step.Err, ending)
	}
	printLeftovers(w, step.Left)
}

// printCommittedError prints the message for a slice whose commit was
// made, but git or the user stopped it afterwards: the slice exists.
func printCommittedError(w io.Writer, done *slice.CommittedError) {
	if isStop(done.Err) {
		fmt.Fprintf(w, stoppedAfterLine, done.Name, done.Base)
		return
	}
	printFailure(w, "finish the commit", done.Err, fmt.Sprintf(createdAnyway, done.Name, done.Base))
}

// printFailure prints that sliceflow could not do what, followed by
// ending, then why: git's own error output when git failed, or else the
// system's error text.
func printFailure(w io.Writer, what string, err error, ending string) {
	var gitErr *git.Error
	if errors.As(err, &gitErr) {
		fmt.Fprintf(w, gitFailedLine, what, ending)
		printGitText(w, gitErr.Stderr)
		return
	}
	fmt.Fprintf(w, systemFailedLine, what, ending)
	fmt.Fprint(w, systemSaid)
	fmt.Fprintf(w, "  %s\n", systemText(err))
}

// printGitText prints what git printed to its error output, word for
// word, with each line indented by two spaces. An empty line stays empty.
// When git printed nothing, it says so.
func printGitText(w io.Writer, stderr string) {
	text := strings.TrimSuffix(strings.ReplaceAll(stderr, "\r\n", "\n"), "\n")
	if text == "" {
		fmt.Fprint(w, gitSaidNothing)
		return
	}
	fmt.Fprint(w, gitSaid)
	for _, line := range strings.Split(text, "\n") {
		if line == "" {
			fmt.Fprintln(w)
			continue
		}
		fmt.Fprintf(w, "  %s\n", line)
	}
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

// systemText returns the system's own words for err. For a file error,
// that is the reason only, e.g. "permission denied": the first line
// already names the file.
func systemText(err error) string {
	var pathErr *fs.PathError
	if errors.As(err, &pathErr) {
		return pathErr.Err.Error()
	}
	return err.Error()
}

// isStop reports whether err comes from the user stopping sliceflow,
// e.g. with Ctrl+C.
func isStop(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}
