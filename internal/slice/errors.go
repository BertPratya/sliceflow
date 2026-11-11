package slice

import (
	"errors"
	"strings"

	"github.com/BertPratya/sliceflow/internal/git"
)

// The errors of slice create. The text of each error is the whole
// message the command prints: one or more full lines, with no line break
// at the end (see runCreate in internal/cli).

// wrongUseError is an error of the command line (R20, R29). Its method
// WrongUse makes internal/cli exit with code 2 (R76), without this
// package importing internal/cli.
type wrongUseError struct {
	text string
}

func (e *wrongUseError) Error() string { return e.text }

// WrongUse reports that the command line is wrong. See wrongUseMarker in
// internal/cli.
func (e *wrongUseError) WrongUse() bool { return true }

// errControl is the problem of R29: option, "--title" or
// "--description", holds a control character. text is the option's
// value, as the user gave it; it is printed with its control characters
// written as codes.
func errControl(option, text string) error {
	return &wrongUseError{text: "Error: " + option + " has a control character. Nothing was changed.\n" +
		`Received: "` + EscapeControlCharacters(text) + `"` + "\n" +
		"Write it as plain text on one line, with no control characters."}
}

// errBadName is the problem of R20: the name the user gave, given,
// breaks R19. err is the error of ParseName, whose text is the second
// line.
func errBadName(given string, err error) error {
	return &wrongUseError{text: `Error: invalid slice name "` + given + `". Nothing was changed.` + "\n" +
		err.Error()}
}

// errGitNotFound is the problem of R14: git cannot be run.
func errGitNotFound() error {
	return errors.New("Error: git was not found. Nothing was changed.\n" +
		"Install git, then try again.")
}

// errNotARepo is the problem of R13: the folder is not inside a git
// repository.
func errNotARepo() error {
	return errors.New("Error: this folder is not in a git repository. Nothing was changed.\n" +
		"Run sliceflow inside your project.")
}

// errNoCommits is the problem of R15: the repository has no commits.
func errNoCommits() error {
	return errors.New("Error: this repository has no commits yet. Nothing was changed.\n" +
		"Make a first commit, then try again.")
}

// errNoBase is the problem of R12: the branch base does not exist. name
// is the slice name, for the command to run. fromGiven is true when the
// user gave --from.
func errNoBase(name, base string, fromGiven bool) error {
	first := "Error: base branch " + base + " does not exist. Nothing was changed.\n"
	if fromGiven {
		return errors.New(first +
			"Check the name. To see your branches, run:\n" +
			"  git branch")
	}
	return errors.New(first +
		"sliceflow starts from main unless you choose another branch. To choose one, run:\n" +
		"  sliceflow slice create " + name + " --from <branch>")
}

// errFolderTaken is the problem of R23: the folder of the slice name is
// in the base commit of the branch base. The folder is a path in git, so
// it is written with "/" on every system.
func errFolderTaken(name, base string) error {
	return errors.New("Error: slice " + name + " already exists (folder docs/slices/" + name + "/ in " + base + "). Nothing was changed.")
}

// errBranchTaken is the problem of R24: the branch of the slice name
// exists.
func errBranchTaken(name string) error {
	return errors.New("Error: slice " + name + " already exists (branch slice/" + name + "). Nothing was changed.")
}

// failedCheck is a check before sliceflow makes anything that failed in
// a way with no message of its own. question is what the check asks,
// e.g. "find the base branch".
//
// Steps:
//  1. If err is a git error, the first line is "Error: git could not
//     <question>. Nothing was changed.", and git's text follows, as in
//     R37 (gitLines).
//  2. Otherwise, the first line is "Error: could not <question>. Nothing
//     was changed.", and the error's text follows, as in R38
//     (systemLines).
func failedCheck(question string, err error) error {
	var gitErr *git.Error
	if errors.As(err, &gitErr) {
		return joinLines("Error: git could not "+question+". Nothing was changed.", gitLines(gitErr.Stderr), nil)
	}
	return joinLines("Error: could not "+question+". Nothing was changed.", systemLines(err), nil)
}

// failedStep is the problem of R37: the git step, e.g. "make the
// commit", failed. left is what the undo could not remove (R41).
//
// Steps:
//  1. The first line is "Error: git could not <step>. The slice was not
//     created.", and "Nothing was changed." when nothing is left.
//  2. If err is a git error, git's text follows (gitLines). Otherwise,
//     the error's text follows (systemLines).
//  3. The R41 lines follow (leftLines).
func failedStep(step string, err error, left []undoStep) error {
	first := "Error: git could not " + step + ". The slice was not created."
	var gitErr *git.Error
	if errors.As(err, &gitErr) {
		return withLeft(first, gitLines(gitErr.Stderr), left)
	}
	return withLeft(first, systemLines(err), left)
}

// failedWrite is the problem of R38: sliceflow could not write the file
// path, or take the lock of the log. left is what the undo could not
// remove (R41).
func failedWrite(path string, err error, left []undoStep) error {
	return withLeft("Error: could not write "+path+". The slice was not created.", systemLines(err), left)
}

// stopped is the message of R39: the user pressed Ctrl+C. left is what
// the undo could not remove (R41).
func stopped(left []undoStep) error {
	return withLeft("Stopped. The slice was not created.", nil, left)
}

// withLeft builds the error of a step that failed after sliceflow began
// to make the slice.
//
// Steps:
//  1. Make the R41 lines of left (leftLines).
//  2. If there are none, add " Nothing was changed." to first (R41).
//  3. Return first, then body, then the R41 lines.
func withLeft(first string, body []string, left []undoStep) error {
	leftover := leftLines(left)
	if len(leftover) == 0 {
		first += " Nothing was changed."
	}
	return joinLines(first, body, leftover)
}

// joinLines returns an error whose text is first, then the lines of
// body, then the lines of tail, one per line, with no line break at the
// end.
func joinLines(first string, body, tail []string) error {
	lines := append([]string{first}, body...)
	lines = append(lines, tail...)
	return errors.New(strings.Join(lines, "\n"))
}

// gitLines returns the lines that show git's error output, stderr, as
// R37 says: "git said:", then git's lines (indent). When git printed no
// line, it returns "git printed no message." alone.
func gitLines(stderr string) []string {
	lines := indent(stderr)
	if len(lines) == 0 {
		return []string{"git printed no message."}
	}
	return append([]string{"git said:"}, lines...)
}

// systemLines returns the lines that show the text of err, as R38 says:
// "The system said:", then the text's lines (indent).
func systemLines(err error) []string {
	return append([]string{"The system said:"}, indent(err.Error())...)
}

// indent returns the lines of text, each one indented by two spaces.
//
// Steps:
//  1. Remove one line break at the end of text, "\n" or "\r\n": a last
//     line break is not a line. If text is then "", return nil.
//  2. Split text at each "\n", and remove a "\r" at the end of each line,
//     so a line that ends with "\r\n" on Windows is the same line.
//  3. Put two spaces before each line that is not empty. An empty line
//     stays empty, with no spaces (R37).
func indent(text string) []string {
	text = strings.TrimSuffix(text, "\n")
	text = strings.TrimSuffix(text, "\r")
	if text == "" {
		return nil
	}
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		line = strings.TrimSuffix(line, "\r")
		if line != "" {
			line = "  " + line
		}
		lines[i] = line
	}
	return lines
}

// leftLines returns the lines of R41 for the steps in left, in their
// order, or nil when no step has a left.
//
// Steps:
//  1. Keep only the steps whose left is not "". If none are kept,
//     return nil.
//  2. Return "Some of what sliceflow made could not be removed:", the
//     left of each kept step, "To clean up, run:", and the cleanup of
//     each kept step. Each left and each cleanup is indented by two
//     spaces.
func leftLines(left []undoStep) []string {
	var kept []undoStep
	for _, step := range left {
		if step.left != "" {
			kept = append(kept, step)
		}
	}
	if len(kept) == 0 {
		return nil
	}
	lines := []string{"Some of what sliceflow made could not be removed:"}
	for _, step := range kept {
		lines = append(lines, "  "+step.left)
	}
	lines = append(lines, "To clean up, run:")
	for _, step := range kept {
		lines = append(lines, "  "+step.cleanup)
	}
	return lines
}
