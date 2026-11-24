package slicecmd

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
	stoppedLine      = "Stopped.%s\n"
	notCreated       = " The slice was not created."
	leftoversLine    = "Some of what sliceflow made could not be removed:\n"
	leftoverLine     = "  %-8s  %s\n"
	cleanUpLine      = "To clean up:\n"
	deleteFolderCmd  = "  delete the folder %s\n"
	unstageCmd       = "  run: git -C \"%s\" reset -- \"%s\"\n"
	stoppedAfterLine = "Stopped after the commit. Slice %s was %s on branch %s.\n"
	doneAnyway       = " Slice %s was %s on branch %s."
)

// The lines of slice status. Each label line has its value in one column.
const (
	statusFirstLine     = "%s: %s\n"
	statusLine          = "  %-13s %s\n"
	noneValue           = "(none)"
	unknownValue        = "unknown"
	nextWriteSpecLines  = "Next: write the spec in %s\nThen run: sliceflow spec validate %s\n"
	createdTimeLayout   = "2006-01-02 15:04 UTC"
	eventsValue         = "%d (last: %s)"
	noSliceLine         = "Error: no slice named %s. Nothing was changed.\nNo log was found at docs/slices/%s/.events.jsonl.\n"
	noSliceOnBranchLine = "Your checkout is on branch %s. Check the name, or switch to the branch the slice is planned on.\n"
	noSliceNoBranchLine = "Your checkout is not on a branch. Check the name, or switch to the branch the slice is planned on.\n"
	noCreatedEventLines = "Error: no slice.created event could be read for %s. Nothing was changed.\nFix or restore the log, then try again.\n"
)

// The lines of slice list.
const (
	notValidWarning       = "Warning: %s is not listed as a slice. It is not a valid slice name.\n"
	noLogWarning          = "Warning: %s is not listed as a slice. The folder docs/slices/%s/ has no event log.\n"
	noCreatedEventWarning = "Warning: %s is not listed as a slice. No slice.created event could be read from %s.\n"
	noSlicesLine          = "No slices yet. To make one, run: sliceflow slice create <name>\n"
)

// listHeader is the header line of the table of slice list.
var listHeader = [3]string{"NAME", "TITLE", "DESCRIPTION"}

// The lines of slice edit.
const (
	editedLine         = "Edited slice %s.\n"
	editValueLine      = "  %-13s %s\n"
	nothingToChange    = "Nothing to change: slice %s already has this title and description.\n"
	needsAnOptionLines = "Error: slice edit needs --title, --description, or both. Nothing was changed.\n" +
		"To see the options, run: sliceflow slice edit --help\n"
	notOnBaseLine   = "Error: slice %s is planned on branch %s, but your checkout is on %s. Nothing was changed.\n"
	notOnBranchLine = "Error: slice %s is planned on branch %s, but your checkout is not on a branch. Nothing was changed.\n"
	noStateLine     = "Error: the state of slice %s cannot be worked out, so it cannot be changed. Nothing was changed.\n"
	updateHint      = "Update sliceflow, then try again.\n"
	logHistoryHint  = "To see how the log changed, run: git log -p -- %s\n"
	logLeftoverLine = "  log       %s\n"
	logStagedLine   = "  staged    %s\n"
	restoreLogCmd   = "  run: git -C \"%s\" restore --staged --worktree -- \"%s\"\n"
)
