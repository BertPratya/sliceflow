package cli

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/BertPratya/sliceflow/internal/slice"
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
	// key, so a command can tell `--title=` (given, "") from no --title.
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
// break at the end. An error that is wrong use (see wrongUseMarker) gives
// exit code 2, any other error 1.
//
// TEMPORARY: the work of status and list is not built yet, so each one
// fails with exit code 1. Step 8 of the plan replaces runStatus, and
// step 9 replaces runList.

// runCreate runs `slice create`: it makes a new slice (Story 1).
//
// Steps:
//  1. Get the current folder with os.Getwd. A failure is an error of
//     two lines: sliceflow's own, then the system's text.
//  2. Build slice.CreateInput from c: the name, and the values of
//     --title, --description, and --from ("" when not given).
//  3. Call slice.Create in that folder. Its error is returned as it is:
//     its text is the whole message.
//  4. Print the lines of R8 to stdout, with printCreated.
func runCreate(ctx context.Context, c call, stdout, _ io.Writer) error {
	dir, err := os.Getwd()
	if err != nil {
		return errors.New("Error: sliceflow could not find the current folder. Nothing was changed.\n" + err.Error())
	}
	created, err := slice.Create(ctx, dir, slice.CreateInput{
		GivenName:   c.name,
		Title:       c.options["title"],
		Description: c.options["description"],
		From:        c.options["from"],
	})
	if err != nil {
		return err
	}
	return printCreated(stdout, c.name, created)
}

// printCreated prints to w the lines of R8 for the slice c, which was
// made from the given name givenName:
//  1. `Created slice <name>.`
//  2. When c.Lowered, the Note of R22, with givenName.
//  3. One Note of R5 for each file of c.IgnoredFiles, in its order.
//  4. `  Branch:   slice/<name> (from <base>, at <commit>)`, with the
//     first 7 characters of c.BaseCommit.
//  5. `  Worktree: <path>`, then `Next: write the spec in <path>`.
//
// The lines are written at once. It returns the error of the write.
func printCreated(w io.Writer, givenName string, c slice.Created) error {
	var b strings.Builder
	b.WriteString("Created slice " + c.Name + ".\n")
	if c.Lowered {
		b.WriteString(`  Note:     "` + givenName + `" was made lowercase.` + "\n")
	}
	for _, file := range c.IgnoredFiles {
		b.WriteString("  Note:     " + file + " matches a .gitignore rule, so sliceflow added it anyway.\n")
	}
	commit := c.BaseCommit
	if len(commit) > 7 {
		commit = commit[:7]
	}
	b.WriteString("  Branch:   slice/" + c.Name + " (from " + c.Base + ", at " + commit + ")\n")
	b.WriteString("  Worktree: " + c.Worktree + "\n")
	b.WriteString("Next: write the spec in " + c.SpecPath + "\n")
	_, err := io.WriteString(w, b.String())
	return err
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
