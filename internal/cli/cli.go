// Package cli reads sliceflow's command line, picks the command, and
// turns the result into an exit code. It knows the commands and their
// options, but not how a slice is made: that work lives in
// internal/slice (plan, "How the pieces connect").
package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
)

// The exit codes of R76 and SD17.
const (
	exitOK       = 0 // the command worked
	exitFailed   = 1 // the command failed for any other reason
	exitWrongUse = 2 // the command line is wrong
)

// wrongUser is an error that says the command line is wrong. An error
// from any package is wrong use when it, or an error it wraps, has the
// method WrongUse and the method returns true. The create flow uses it
// for a bad name (R20) and a control character (R29), without importing
// internal/cli.
type wrongUser interface {
	WrongUse() bool
}

// Run runs sliceflow with args, the words of the command line without
// the program's name. It prints to stdout and stderr, and returns the
// exit code (R76, SD17): 0 when the command works, 2 when the command
// line is wrong, and 1 when it fails for any other reason.
//
// When ctx is cancelled, e.g. by Ctrl+C, the command stops (TD5): the
// command's work gets ctx.
//
// Steps:
//  1. Check the command line from the left, against the table of
//     commands (R40, R71).
//  2. If --help or -h was given, print the help of the command, and
//     return 0 (R67, R68). It comes before any problem (R40).
//  3. If the command line has a problem, print it, and return 2.
//  4. If no command was given (`sliceflow`, `sliceflow slice`), print
//     the guide, and return 0 (R67).
//  5. Run the command. Print its error, if any, and return its exit
//     code.
func Run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	found, err := checkLine(commands, args)
	if found.wantsHelp {
		return printHelp(stdout, found.helpText())
	}
	if err != nil {
		return report(stderr, err)
	}
	if found.cmd == nil {
		return printHelp(stdout, found.helpText())
	}
	return report(stderr, found.cmd.run(ctx, found.call, stdout, stderr))
}

// printHelp prints help to out, and returns the exit code: 0, or 1 when
// the help could not be written.
func printHelp(out io.Writer, help string) int {
	if _, err := io.WriteString(out, help); err != nil {
		return exitFailed
	}
	return exitOK
}

// report prints err to errOut, as it is, followed by a line break, and
// returns its exit code. A nil err prints nothing, and returns 0.
func report(errOut io.Writer, err error) int {
	if err == nil {
		return exitOK
	}
	// The error of Fprintln is not checked: the exit code already says
	// that the command failed, and there is no other place to write to.
	fmt.Fprintln(errOut, err.Error())
	return exitCode(err)
}

// exitCode returns the exit code for err (R76): 0 for nil, 2 when err is
// wrong use (see wrongUser), and 1 for any other error.
func exitCode(err error) int {
	if err == nil {
		return exitOK
	}
	var wrong wrongUser
	if errors.As(err, &wrong) && wrong.WrongUse() {
		return exitWrongUse
	}
	return exitFailed
}
