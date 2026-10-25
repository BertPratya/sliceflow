package cli

import "fmt"

// command is one word the user can type, like "create".
type command struct {
	name  string
	args  string
	about string
	run   func(args []string) int
}

// group is one level of commands, like "sliceflow" or "sliceflow slice".
type group struct {
	path     string
	about    string
	commands []command
}

// route finds the command whose name is the first word in args, and runs
// it with the rest of the words. With no words, it prints the group's help
// and returns 0. With an unknown word, it returns 2.
//
// For example, with args [create spec-submit], it runs the create command
// with [spec-submit].
func (g group) route(args []string) int {
	if len(args) == 0 {
		fmt.Print(g.help())
		return 0
	}

	for _, c := range g.commands {
		if args[0] == c.name {
			return c.run(args[1:])
		}
	}

	fmt.Println("unknown command")
	return 2
}
