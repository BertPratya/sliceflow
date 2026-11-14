package cli

import (
	"context"
	"io"
)

// Run reads the command line in args, the words after the program's
// name, runs the command, and returns the exit code: 0 when it works, 2
// when the command line is wrong, 1 for any other failure (R76).
//
// For now, it does nothing and returns 0. Piece 2 of step 6 reads the
// words with the table (TD1).
func Run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	return 0
}
