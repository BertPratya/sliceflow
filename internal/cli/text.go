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
	createdLine       = "Created slice %s.\n"
	madeLowercaseNote = "  Note:     \"%s\" was made lowercase.\n"
	ignoredFileNote   = "  Note:     %s matches a .gitignore rule, so sliceflow added it anyway.\n"
	branchLine        = "  Branch:   slice/%s (from %s, at %s)\n"
	worktreeLine      = "  Worktree: %s\n"
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
	noBaseLine        = "Error: base branch %s does not exist. Nothing was changed.\n"
	defaultBaseHint   = "sliceflow starts from main unless you choose another branch. To choose one, run:\n  sliceflow slice create %s --from <branch>\n"
	chosenBaseHint    = "Check the name. To see your branches, run:\n  git branch\n"
	folderTakenLine   = "Error: slice %s already exists (folder docs/slices/%s/ in %s). Nothing was changed.\n"
	branchTakenLine   = "Error: slice %s already exists (branch slice/%s). Nothing was changed.\n"
	gitFailedLine     = "Error: git could not %s.%s\n"
	systemFailedLine  = "Error: could not %s.%s\n"
	stoppedLine       = "Stopped.%s\n"
	notCreated        = " The slice was not created."
	nothingChanged    = " Nothing was changed."
	gitSaid           = "git said:\n"
	gitSaidNothing    = "git printed no message.\n"
	systemSaid        = "The system said:\n"
	leftoversLine     = "Some of what sliceflow made could not be removed:\n"
	leftoverLine      = "  %-8s  %s\n"
	cleanUpLine       = "To clean up, run:\n"
	removeWorktreeCmd = "  git worktree remove --force \"%s\"\n"
	deleteBranchCmd   = "  git branch -D \"%s\"\n"
)
