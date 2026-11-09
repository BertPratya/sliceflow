package cli

import (
	"context"
	"errors"
	"io"
)

// The help texts, word for word. printHelp prints them as they are.
const (
	// guideHelp is the guide of S41, for `sliceflow` and `sliceflow
	// slice` (R67).
	guideHelp = `sliceflow runs a spec-driven workflow with AI coding agents.

Usage:
  sliceflow <command> [options]

Commands:
  slice create <name>  Make a new slice
  slice status <name>  Show one slice
  slice list           Show all slices

For help with a command, run: sliceflow <command> --help
`

	// createHelp is the help of `slice create`, as in S44 (R68).
	createHelp = `Make a new slice: a branch, a worktree, a spec, and an event log.

Usage:
  sliceflow slice create <name> [options]

Options:
  --title <text>        A short title for the slice. Default: the name.
  --description <text>  One or two sentences about the slice.
  --from <branch>       The branch to start from. Default: main.
  --help                Show this help.
`

	// statusHelp is the help of `slice status`, as in S45 (R69).
	statusHelp = `Show one slice: its title, description, start, branch, worktree, and events.

Usage:
  sliceflow slice status <name> [options]

Options:
  --help  Show this help.
`

	// listHelp is the help of `slice list`, as in R69.
	listHelp = `Show all slices in a table: name, title, and description.

Usage:
  sliceflow slice list [options]

Options:
  --help  Show this help.
`
)

// command is one command of sliceflow, e.g. `slice create`. Its words
// are its key in commands.
type command struct {
	help      string   // its help, printed as it is (R68, R69)
	takesName bool     // it takes one slice name: `slice create <name>`
	options   []string // its options without "--", e.g. "title"; each takes a value
	run       func(ctx context.Context, c call, stdout, stderr io.Writer) error
}

// call is what the command line gives to a command: the name, and the
// value of each option that was given.
type call struct {
	name string // the given name, or "" for a command that takes none
	// options holds the value of each option that was given, by its name
	// without "--", e.g. "title". The last value counts when an option is
	// given more than once (R71). An option that was not given has no
	// key, so a command can tell `--from=` (given, "") from no --from.
	options map[string]string
}

// commands is the table of sliceflow's commands, by the words that name
// them. It is a plain map, like the table of Terraform's commands:
// sliceflow has three commands, so it needs no library to find them.
//
// The words `sliceflow` and `sliceflow slice` alone are not commands:
// they print the guide (R67).
var commands = map[string]command{
	"slice create": {help: createHelp, takesName: true, options: []string{"title", "description", "from"}, run: runCreate},
	"slice status": {help: statusHelp, takesName: true, run: runStatus},
	"slice list":   {help: listHelp, run: runList},
}

// The work behind each command. Each one prints what the command shows
// when it works to stdout, and its warnings to stderr. It returns an
// error when the command fails: Run prints the error's text as it is,
// so the text is the whole message, one or more full lines with no line
// break at the end. An error that is wrong use (see wrongUser) gives
// exit code 2, any other error 1.
//
// TEMPORARY: the work is not built yet, so each one fails with exit
// code 1. Step 7 of the plan replaces runCreate with a call to the
// create flow, step 8 replaces runStatus, and step 9 replaces runList.

// runCreate runs `slice create`: it makes a new slice (Story 1).
func runCreate(_ context.Context, _ call, _, _ io.Writer) error {
	return errNotBuilt("slice create")
}

// runStatus runs `slice status`: it shows one slice (S45).
func runStatus(_ context.Context, _ call, _, _ io.Writer) error {
	return errNotBuilt("slice status")
}

// runList runs `slice list`: it shows all slices (R69).
func runList(_ context.Context, _ call, _, _ io.Writer) error {
	return errNotBuilt("slice list")
}

// errNotBuilt is the error of a command whose work is not built yet.
// TEMPORARY: it goes away with the last placeholder above.
func errNotBuilt(command string) error {
	return errors.New("Error: sliceflow " + command + " is not built yet.")
}

// commandPath returns the command line that words name, e.g.
// "sliceflow slice create", or "sliceflow" for no words.
func commandPath(words string) string {
	if words == "" {
		return "sliceflow"
	}
	return "sliceflow " + words
}

// usageLine returns the usage line of cmd, which words name, as in R73
// and R75, e.g. "sliceflow slice create <name> [options]".
func usageLine(words string, cmd *command) string {
	line := commandPath(words)
	if cmd.takesName {
		line += " <name>"
	}
	return line + " [options]"
}
