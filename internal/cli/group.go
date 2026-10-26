package cli

import (
	"fmt"
	"strings"
)

// command is one word the user can type, like "create". A leaf command has
// a usage text, which it prints for --help.
type command struct {
	name  string
	args  string
	about string
	usage string
	run   func(e *env, args []string) int
}

// group is one level of commands, like "sliceflow" or "sliceflow slice".
type group struct {
	path     string
	about    string
	commands []command
}

// route finds the command whose name is the first word in args, and runs
// it with the rest of the words. With no words, or with -h or --help, it
// prints the group's help and returns 0. With an unknown flag or an
// unknown word, it returns 2.
//
// For example, with args [create spec-submit], it runs the create command
// with [spec-submit].
func (g group) route(e *env, args []string) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Fprint(e.stdout, g.help())
		return exitOK
	}

	word := args[0]
	if strings.HasPrefix(word, "-") {
		// Help wins over an unknown flag, if -h or --help is one of the
		// leading flags: "sliceflow --colour -h" shows the help.
		for _, w := range args {
			if w == "--" || !strings.HasPrefix(w, "-") {
				break
			}
			if w == "-h" || w == "--help" {
				fmt.Fprint(e.stdout, g.help())
				return exitOK
			}
		}
		return wrongUse(e, "unknown flag "+word)
	}

	for _, c := range g.commands {
		if word == c.name {
			return c.run(e, args[1:])
		}
	}

	return wrongUse(e, fmt.Sprintf("unknown command %q", word))
}
