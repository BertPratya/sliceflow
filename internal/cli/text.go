package cli

// intro is the first line of the guide.
const intro = "sliceflow runs a spec-driven workflow with AI coding agents."

// helpOption is the line for -h and --help that ends every help text.
var helpOption = row{"-h, --help", "Show this help."}

// The two lines of a wrong command line. Run prints the first one as
// "Error: <message>. Nothing was changed.", then the hint.
// Each %s is printed as typed: %q would escape some characters.
const (
	unknownCommandMessage     = "unknown command \"%s\""
	unknownOptionMessage      = "unknown option \"%s\""
	missingValueMessage       = "option \"%s\" needs a value"
	unexpectedArgumentMessage = "unexpected argument \"%s\""
	// missingArgumentMessage gets the group and the argument's name, as
	// in "missing slice name".
	missingArgumentMessage = "missing %s %s"

	seeCommandsHint = "To see the commands, run: sliceflow --help"
	// seeOptionsHint gets the command the user ran, e.g. "slice create".
	seeOptionsHint = "To see the options, run: sliceflow %s --help"
	// seeAllOptionsHint is for an option given with no command.
	seeAllOptionsHint = "To see the options, run: sliceflow --help"
	usageHint         = "Usage: %s"
)

// The lines that slice create prints when the slice is made. Each Note
// comes right after the first line, when it applies.
const (
	createdLine       = "Created slice %s on branch %s.\n"
	madeLowercaseNote = "  Note: \"%s\" was made lowercase.\n"
	ignoredFileNote   = "  Note: %s matches a .gitignore rule, so sliceflow added it anyway.\n"
	nextLine          = "Next: write the spec in %s\n"
)

// The messages of slice create when it cannot make the slice. Each goes
// to stderr. "Nothing was changed." is left out of a first line when
// something sliceflow made could not be removed.
const (
	controlCharacterLines = "Error: --%s has a control character. Nothing was changed.\n" +
		"Received: \"%s\"\n" +
		"Write it as plain text on one line, with no control characters.\n"
	invalidNameLines = "Error: invalid slice name \"%s\". Nothing was changed.\n%s\n"
	gitNotFoundLines = "Error: git was not found. Nothing was changed.\n" +
		"Install git, then try again.\n"
	noCommitsLines = "Error: this repository has no commits yet. Nothing was changed.\n" +
		"Make a first commit, then try again.\n"
	noBaseLine       = "Error: base branch %s does not exist. Nothing was changed.\n"
	defaultBaseHint  = "sliceflow starts from main unless you choose another branch. To choose one for this slice, run:\n  sliceflow slice create %s --from <branch>\nTo choose one for every slice, run:\n  git config sliceflow.base <branch>\n"
	flagBaseHint     = "Check the name. To see your branches, run:\n  git branch\n"
	settingBaseHint  = "The base branch comes from the git setting sliceflow.base. To see it, run:\n  git config sliceflow.base\n"
	wrongBranchLine  = "Error: the base branch is %s, but your checkout is on %s. Nothing was changed.\n"
	noBranchLine     = "Error: the base branch is %s, but your checkout is not on a branch. Nothing was changed.\n"
	switchToBaseLine = "Switch to %s, then try again.\n"
	folderTakenLine  = "Error: slice %s already exists (folder docs/slices/%s/). Nothing was changed.\n"
	branchTakenLine  = "Error: slice %s already exists (branch slice/%s). Nothing was changed.\n"
	gitFailedLine    = "Error: git could not %s.%s\n"
	systemFailedLine = "Error: could not %s.%s\n"
	stoppedLine      = "Stopped.%s\n"
	notCreated       = " The slice was not created."
	nothingChanged   = " Nothing was changed."
	gitSaid          = "git said:\n"
	gitSaidNothing   = "git printed no message.\n"
	systemSaid       = "The system said:\n"
	leftoversLine    = "Some of what sliceflow made could not be removed:\n"
	leftoverLine     = "  %-8s  %s\n"
	cleanUpLine      = "To clean up:\n"
	deleteFolderCmd  = "  delete the folder %s\n"
	unstageCmd       = "  run: git -C \"%s\" reset -- \"%s\"\n"
	stoppedAfterLine = "Stopped after the commit. Slice %s was created on branch %s.\n"
	createdAnyway    = " Slice %s was created on branch %s."
)

// The lines of slice status. Each label line has its value in one column.
const (
	statusFirstLine     = "%s: %s\n"
	statusLine          = "  %-13s %s\n"
	noneValue           = "(none)"
	notFoundValue       = "(not found)"
	createdTimeLayout   = "2006-01-02 15:04 UTC"
	startedFromValue    = "%s, at %s"
	eventsValue         = "%d (last: %s)"
	oneBadLineWarning   = "Warning: line %d of %s could not be read, so it was skipped.\n"
	badLinesWarning     = "Warning: lines %s of %s could not be read, so they were skipped.\n"
	moreBadLines        = " (%d more not listed)"
	noSliceLines        = "Error: no slice named %s. Nothing was changed.\nNo log was found in any of these places:\n  worktree  %s\n  branch    slice/%s\n  main      docs/slices/%s/.events.jsonl\nCheck the name, or run: sliceflow slice list\n"
	noCreatedEventLines = "Error: no slice.created event could be read for %s. Nothing was changed.\nFix or restore the log, then try again.\n"
)

// The lines of slice list.
const (
	notValidWarning       = "Warning: %s is not listed as a slice. It is not a valid slice name.\n"
	noLogOnBranchWarning  = "Warning: %s is not listed as a slice. The branch slice/%s has no event log.\n"
	noLogInMainWarning    = "Warning: %s is not listed as a slice. The folder docs/slices/%s/ in main has no event log.\n"
	noLogInBothWarning    = "Warning: %s is not listed as a slice. Neither the branch slice/%s nor the folder docs/slices/%s/ in main has an event log.\n"
	noCreatedEventWarning = "Warning: %s is not listed as a slice. No slice.created event could be read from %s.\n"
	noSlicesLine          = "No slices yet. To make one, run: sliceflow slice create <name>\n"
)

// listHeader is the header line of the table of slice list.
var listHeader = [3]string{"NAME", "TITLE", "DESCRIPTION"}
