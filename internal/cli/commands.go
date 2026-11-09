package cli

import (
	"context"
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
	// run turns the call into the input of the command's work in f, and
	// runs it. f is given, not fixed, so that tests can give their own.
	run func(ctx context.Context, f flows, c call, stdout, stderr io.Writer) error
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

// runCreate runs `slice create`: it gives the name and the options to
// f.create, with "main" when --from was not given (R10).
func runCreate(ctx context.Context, f flows, c call, stdout, stderr io.Writer) error {
	from, fromGiven := c.options["from"]
	if !fromGiven {
		from = "main"
	}
	in := createInput{
		name:        c.name,
		title:       c.options["title"],
		description: c.options["description"],
		from:        from,
		fromGiven:   fromGiven,
	}
	return f.create(ctx, in, stdout, stderr)
}

// runStatus runs `slice status`: it gives the name to f.status.
func runStatus(ctx context.Context, f flows, c call, stdout, stderr io.Writer) error {
	return f.status(ctx, statusInput{name: c.name}, stdout, stderr)
}

// runList runs `slice list`: it calls f.list.
func runList(ctx context.Context, f flows, _ call, stdout, stderr io.Writer) error {
	return f.list(ctx, stdout, stderr)
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
