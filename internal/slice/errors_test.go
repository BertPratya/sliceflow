package slice

import (
	"errors"
	"testing"

	"github.com/BertPratya/sliceflow/internal/git"
)

// The two things S12 leaves behind, built with the same constructors as
// task 03, so the tests check the R41 text that sliceflow prints. Their
// undo is never run here.
var (
	leftWorktree = worktreeStep("/home/you/.sliceflow/worktrees/shop/card-payment", nil)
	leftBranch   = branchStep("card-payment", nil)
	// leftFile is a step with no R41 line, e.g. a file in the worktree.
	leftFile = undoStep{}
)

// checkText fails the test when err is nil or its text is not want.
func checkText(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("got no error, want:\n%s", want)
	}
	if got := err.Error(); got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

// isWrongUse reports whether err has the method WrongUse, and it returns
// true. It is the check of exitCode in internal/cli.
func isWrongUse(err error) bool {
	var wrong interface{ WrongUse() bool }
	return errors.As(err, &wrong) && wrong.WrongUse()
}

// R37: git's text comes after "git said:", each line indented by two
// spaces, an empty line with no spaces, and no line for the last line
// break. With no text, "git printed no message." takes its place.
func TestFailedStepText(t *testing.T) {
	tests := []struct {
		name   string
		stderr string
		want   string
	}{
		{
			"git's text",
			"line 1\n\nline 2\n",
			"Error: git could not make the commit. The slice was not created. Nothing was changed.\n" +
				"git said:\n" +
				"  line 1\n" +
				"\n" +
				"  line 2",
		},
		{
			"no last line break",
			"line 1",
			"Error: git could not make the commit. The slice was not created. Nothing was changed.\n" +
				"git said:\n" +
				"  line 1",
		},
		{
			"Windows line ends",
			"line 1\r\n\r\nline 2\r\n",
			"Error: git could not make the commit. The slice was not created. Nothing was changed.\n" +
				"git said:\n" +
				"  line 1\n" +
				"\n" +
				"  line 2",
		},
		{
			"no text",
			"",
			"Error: git could not make the commit. The slice was not created. Nothing was changed.\n" +
				"git printed no message.",
		},
		{
			"only a line break",
			"\n",
			"Error: git could not make the commit. The slice was not created. Nothing was changed.\n" +
				"git printed no message.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := failedStep("make the commit", &git.Error{ExitCode: 1, Stderr: tt.stderr}, nil)
			checkText(t, err, tt.want)
			if isWrongUse(err) {
				t.Error("WrongUse is true, want an error that gives exit code 1")
			}
		})
	}
}

// S10: the step's name is in the first line.
func TestFailedStepS10(t *testing.T) {
	stderr := "Preparing worktree (checking out 'slice/card-payment')\n" +
		"fatal: '/home/you/.sliceflow/worktrees/shop/card-payment' already exists\n"
	err := failedStep("create the worktree", &git.Error{ExitCode: 128, Stderr: stderr}, nil)
	checkText(t, err, "Error: git could not create the worktree. The slice was not created. Nothing was changed.\n"+
		"git said:\n"+
		"  Preparing worktree (checking out 'slice/card-payment')\n"+
		"  fatal: '/home/you/.sliceflow/worktrees/shop/card-payment' already exists")
}

// An error that is not git's: its text comes after "The system said:"
// (Assumed: not in the spec).
func TestFailedStepNotGit(t *testing.T) {
	err := failedStep("create the branch", errors.New("broken pipe"), nil)
	checkText(t, err, "Error: git could not create the branch. The slice was not created. Nothing was changed.\n"+
		"The system said:\n"+
		"  broken pipe")
}

// S12 (R41): with things left, the first line has no "Nothing was
// changed.", and the R41 lines follow git's text, the worktree first.
// A step with an empty left gets no line.
func TestLeftoverLines(t *testing.T) {
	stderr := "Author identity unknown\n" +
		"\n" +
		"*** Please tell me who you are.\n" +
		"\n" +
		"Run\n" +
		"\n" +
		"  git config --global user.email \"you@example.com\"\n" +
		"  git config --global user.name \"Your Name\"\n" +
		"\n" +
		"to set your account's default identity.\n" +
		"Omit --global to set the identity only in this repository.\n" +
		"\n" +
		"fatal: unable to auto-detect email address (got 'you@laptop.(none)')\n"
	left := []undoStep{leftWorktree, leftFile, leftBranch}

	err := failedStep("make the commit", &git.Error{ExitCode: 128, Stderr: stderr}, left)

	checkText(t, err, `Error: git could not make the commit. The slice was not created.
git said:
  Author identity unknown

  *** Please tell me who you are.

  Run

    git config --global user.email "you@example.com"
    git config --global user.name "Your Name"

  to set your account's default identity.
  Omit --global to set the identity only in this repository.

  fatal: unable to auto-detect email address (got 'you@laptop.(none)')
Some of what sliceflow made could not be removed:
  worktree  /home/you/.sliceflow/worktrees/shop/card-payment
  branch    slice/card-payment
To clean up, run:
  git worktree remove --force --force "/home/you/.sliceflow/worktrees/shop/card-payment"
  git branch -D "slice/card-payment"`)
}

// R39, with and without R41.
func TestStoppedText(t *testing.T) {
	tests := []struct {
		name string
		left []undoStep
		want string
	}{
		{"nothing left", nil, "Stopped. The slice was not created. Nothing was changed."},
		{
			"a branch left",
			[]undoStep{leftBranch},
			"Stopped. The slice was not created.\n" +
				"Some of what sliceflow made could not be removed:\n" +
				"  branch    slice/card-payment\n" +
				"To clean up, run:\n" +
				`  git branch -D "slice/card-payment"`,
		},
		{"only a step with no left", []undoStep{leftFile}, "Stopped. The slice was not created. Nothing was changed."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := stopped(tt.left)
			checkText(t, err, tt.want)
			if isWrongUse(err) {
				t.Error("WrongUse is true, want an error that gives exit code 1")
			}
		})
	}
}

// R38, with and without R41.
func TestFailedWriteText(t *testing.T) {
	path := "/home/you/.sliceflow/worktrees/shop/card-payment/docs/slices/card-payment/spec.md"
	cause := errors.New("open " + path + ": permission denied")

	t.Run("nothing left", func(t *testing.T) {
		err := failedWrite(path, cause, nil)
		checkText(t, err, "Error: could not write "+path+". The slice was not created. Nothing was changed.\n"+
			"The system said:\n"+
			"  open "+path+": permission denied")
		if isWrongUse(err) {
			t.Error("WrongUse is true, want an error that gives exit code 1")
		}
	})

	t.Run("a worktree left", func(t *testing.T) {
		err := failedWrite(path, cause, []undoStep{leftWorktree})
		checkText(t, err, "Error: could not write "+path+". The slice was not created.\n"+
			"The system said:\n"+
			"  open "+path+": permission denied\n"+
			"Some of what sliceflow made could not be removed:\n"+
			"  worktree  /home/you/.sliceflow/worktrees/shop/card-payment\n"+
			"To clean up, run:\n"+
			`  git worktree remove --force --force "/home/you/.sliceflow/worktrees/shop/card-payment"`)
	})
}

// A check that fails with no message of its own (Assumed: not in the
// spec).
func TestFailedCheckText(t *testing.T) {
	t.Run("git's error", func(t *testing.T) {
		err := failedCheck("find the base branch", &git.Error{ExitCode: 128, Stderr: "fatal: bad thing\n"})
		checkText(t, err, "Error: git could not find the base branch. Nothing was changed.\n"+
			"git said:\n"+
			"  fatal: bad thing")
	})
	t.Run("git's error with no text", func(t *testing.T) {
		err := failedCheck("check the name", &git.Error{ExitCode: 128})
		checkText(t, err, "Error: git could not check the name. Nothing was changed.\n"+
			"git printed no message.")
	})
	t.Run("another error", func(t *testing.T) {
		err := failedCheck("find the worktree folder", errors.New("$HOME is not defined"))
		checkText(t, err, "Error: could not find the worktree folder. Nothing was changed.\n"+
			"The system said:\n"+
			"  $HOME is not defined")
		if isWrongUse(err) {
			t.Error("WrongUse is true, want an error that gives exit code 1")
		}
	})
}

// Each check error has its spec text word for word, with the values of
// the scenarios. Only R20's and R29's are wrong use.
func TestCheckErrorTexts(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		want     string
		wrongUse bool
	}{
		{
			"S2, R20: a character",
			errBadName("card.payment", errCharacters),
			"Error: invalid slice name \"card.payment\". Nothing was changed.\n" +
				"Allowed characters: letters, digits, hyphens, underscores. Start with a letter or a digit.",
			true,
		},
		{
			"R20: too long",
			errBadName("a123456789a123456789a123456789a123456789x", errTooLong),
			"Error: invalid slice name \"a123456789a123456789a123456789a123456789x\". Nothing was changed.\n" +
				"A name has at most 40 characters.",
			true,
		},
		{
			"R20: reserved",
			errBadName("CON", errReserved),
			"Error: invalid slice name \"CON\". Nothing was changed.\n" +
				"Windows reserves this name. Choose another.",
			true,
		},
		{
			"S3, R24",
			errBranchTaken("card-payment"),
			"Error: slice card-payment already exists (branch slice/card-payment). Nothing was changed.",
			false,
		},
		{
			"S4, R23",
			errFolderTaken("card-payment", "main"),
			"Error: slice card-payment already exists (folder docs/slices/card-payment/ in main). Nothing was changed.",
			false,
		},
		{
			"S6, R13",
			errNotARepo(),
			"Error: this folder is not in a git repository. Nothing was changed.\n" +
				"Run sliceflow inside your project.",
			false,
		},
		{
			"S7, R15",
			errNoCommits(),
			"Error: this repository has no commits yet. Nothing was changed.\n" +
				"Make a first commit, then try again.",
			false,
		},
		{
			"S8, R12: no --from",
			errNoBase("card-payment", "main", false),
			"Error: base branch main does not exist. Nothing was changed.\n" +
				"sliceflow starts from main unless you choose another branch. To choose one, run:\n" +
				"  sliceflow slice create card-payment --from <branch>",
			false,
		},
		{
			"S17, R12: --from given",
			errNoBase("card-payment", "release", true),
			"Error: base branch release does not exist. Nothing was changed.\n" +
				"Check the name. To see your branches, run:\n" +
				"  git branch",
			false,
		},
		{
			"S20, R29: --title",
			errControl("--title", "Pay with a card\n"),
			"Error: --title has a control character. Nothing was changed.\n" +
				`Received: "Pay with a card\n"` + "\n" +
				"Write it as plain text on one line, with no control characters.",
			true,
		},
		{
			"R29: --description, escaped",
			errControl("--description", "a\tb\\c\x1b"),
			"Error: --description has a control character. Nothing was changed.\n" +
				`Received: "a\tb\\c\u001b"` + "\n" +
				"Write it as plain text on one line, with no control characters.",
			true,
		},
		{
			"S21, R14",
			errGitNotFound(),
			"Error: git was not found. Nothing was changed.\n" +
				"Install git, then try again.",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkText(t, tt.err, tt.want)
			if got := isWrongUse(tt.err); got != tt.wrongUse {
				t.Errorf("wrong use = %v, want %v", got, tt.wrongUse)
			}
			var marker interface{ WrongUse() bool }
			if !tt.wrongUse && errors.As(tt.err, &marker) {
				t.Error("the error has a WrongUse method, want none")
			}
		})
	}
}
