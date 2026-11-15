package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
)

// Run reads the command line in args, the words after the program's
// name, runs the command, and returns the exit code: 0 when it works, 2
// when the command line is wrong, 1 for any other failure.
//
// When the words ask for help, it prints the help to stdout and returns
// 0. A wrong command line gets two lines on stderr: what is wrong, then
// where to look. Otherwise the command's Run gets what was parsed, and
// its exit code is returned.
func Run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	var usage *usageError
	line, err := parseCommandLine(commands, args)
	if errors.As(err, &usage) {
		fmt.Fprintf(stderr, "Error: %s. Nothing was changed.\n%s\n", usage.message, usage.hint)
		return 2
	}
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	if line.help != "" {
		fmt.Fprint(stdout, line.help)
		return 0
	}
	return line.action.Run(ctx, line.input, stdout, stderr)
}
