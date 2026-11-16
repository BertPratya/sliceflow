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
