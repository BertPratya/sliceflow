// Package cli is sliceflow's command line: it finds the command and runs it.
package cli

import (
	"context"
	"fmt"
	"io"
	"strings"
)

// command is one sliceflow command. Name is e.g. "slice create".
type command struct {
	Name, Usage string
	Run         func(ctx context.Context, args []string, stdout, stderr io.Writer) int
}

var registry = map[string]command{}

// register adds a command. Command files call it from init().
func register(c command) {
	if _, dup := registry[c.Name]; dup {
		panic("cli: command registered twice: " + c.Name)
	}
	registry[c.Name] = c
}

const topUsage = `Usage: sliceflow <command> [flags]

Commands:
  slice create <name> [--title <title>] [--from <branch>] [--json]
  slice status [<name>] [--json]
  slice list [--all] [--json]
  slice title <name> <title>
  spec validate [<name>] [--strict]
  spec submit [<name>]
  spec review [<name>]
  template show <template>

Run "sliceflow <command> --help" for the usage of one command.
`

// Run runs sliceflow with the arguments after the program name. It returns
// the exit code.
func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || isHelp(args[0]) || (len(args) == 1 && isGroup(args[0])) ||
		(isGroup(args[0]) && len(args) > 1 && isHelp(args[1])) {
		fmt.Fprint(stdout, topUsage)
		return 0
	}
	if strings.HasPrefix(args[0], "-") {
		return usageError(stderr, "unknown flag "+flagText(args[0]))
	}
	if isGroup(args[0]) && len(args) > 1 && strings.HasPrefix(args[1], "-") {
		return usageError(stderr, "unknown flag "+flagText(args[1]))
	}
	// A command is one word, or a group and a second word.
	n := 1
	if isGroup(args[0]) {
		n = 2
	}
	name := strings.Join(args[:n], " ")
	c, ok := registry[name]
	if !ok {
		return usageError(stderr, fmt.Sprintf("unknown command %q", name))
	}
	return c.Run(context.Background(), args[n:], stdout, stderr)
}

// isGroup tells if a is a command group: "slice", "spec", or "template".
func isGroup(a string) bool { return a == "slice" || a == "spec" || a == "template" }

func isHelp(a string) bool { return a == "-h" || a == "--help" }

// flagText returns the flag as typed, without a "=value".
func flagText(a string) string {
	if i := strings.Index(a, "="); i >= 0 {
		return a[:i]
	}
	return a
}
