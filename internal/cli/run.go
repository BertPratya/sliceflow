package cli

import (
	"context"
	"fmt"
	"io"
	"slices"
)

// Run reads the command line in args, the words after the program's
// name, runs the command, and returns the exit code: 0 when it works, 2
// when the command line is wrong, 1 for any other failure (R76).
//
// For now, it only shows the guide, the help of a group (R67), and the
// help of a command (R68).
func Run(ctx context.Context, args []string, stdout, stderr io.Writer) int {

	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		return showHelp(stdout, guideText(commands))
	}

	group, ok := commands.Groups[args[0]]
	if ok && (len(args) == 1 || args[1] == "-h" || args[1] == "--help") {
		return showHelp(stdout, groupHelpText(args[0], group))
	}

	if ok {
		action, found := group.Actions[args[1]]
		if found && asksForHelp(args[2:]) {
			return showHelp(stdout, commandHelpText(args[0], args[1], action))
		}
	}
	return 0
}

// asksForHelp reports whether words has -h or --help in it. --help comes
// before any error of the command line (R40).
func asksForHelp(words []string) bool {
	return slices.Contains(words, "-h") || slices.Contains(words, "--help")
}

func showHelp(stdout io.Writer, text string) int {
	fmt.Fprint(stdout, text)
	return 0
}
