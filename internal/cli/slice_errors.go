package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"strings"

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
		noBase  *slice.NoBaseError
		taken   *slice.NameTakenError
		check   *slice.CheckError
	)
	switch {
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
		if noBase.Chosen {
			fmt.Fprint(w, chosenBaseHint)
		} else {
			fmt.Fprintf(w, defaultBaseHint, noBase.Name)
		}
	case errors.As(err, &taken):
		if errors.Is(taken, slice.ErrFolderTaken) {
			fmt.Fprintf(w, folderTakenLine, taken.Name, taken.Name, taken.Base)
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
// text, then what could not be removed, if anything.
func printStepError(w io.Writer, step *slice.StepError) {
	ending := notCreated
	if len(step.Left) == 0 {
		ending += nothingChanged
	}
	if isStop(step.Err) {
		fmt.Fprintf(w, stoppedLine, ending)
	} else {
		printFailure(w, step.Step, step.Err, ending)
	}
	printLeftovers(w, step.Left)
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

// printLeftovers prints what could not be removed, then a command that
// removes each thing, the worktree first: git refuses to delete a branch
// that a worktree has checked out. It prints nothing for an empty list.
func printLeftovers(w io.Writer, left []slice.Leftover) {
	if len(left) == 0 {
		return
	}
	fmt.Fprint(w, leftoversLine)
	for _, thing := range left {
		fmt.Fprintf(w, leftoverLine, thing.Kind, thing.Name)
	}
	fmt.Fprint(w, cleanUpLine)
	for _, thing := range left {
		if thing.Kind == "worktree" {
			fmt.Fprintf(w, removeWorktreeCmd, thing.Name)
		}
	}
	for _, thing := range left {
		if thing.Kind == "branch" {
			fmt.Fprintf(w, deleteBranchCmd, thing.Name)
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
