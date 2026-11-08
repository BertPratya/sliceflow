package cli

import (
	"context"
	"errors"
	"io"
)

// createInput holds what the user typed for `slice create`. Nothing in
// it is checked yet: the create flow checks the control characters and
// the name, in the order of R40.
type createInput struct {
	name        string // the given name, as typed (R21)
	title       string // the text of --title, or "" (R28)
	description string // the text of --description, or "" (R28)
	from        string // the text of --from, or "main" (R10)
	fromGiven   bool   // whether --from was given, for the lines of R12
}

// statusInput holds what the user typed for `slice status`.
type statusInput struct {
	name string // the given name, as typed, not checked yet (R40)
}

// flows holds the work behind each command. Each one prints what the
// command shows when it works to stdout, and its warnings to stderr. It
// returns an error when the command fails: run prints the error's text
// as it is, so the text is the whole message, one or more full lines
// with no line break at the end. An error that is wrong use (see
// wrongUser) gives exit code 2, any other error 1.
type flows struct {
	create func(ctx context.Context, in createInput, stdout, stderr io.Writer) error
	status func(ctx context.Context, in statusInput, stdout, stderr io.Writer) error
	list   func(ctx context.Context, stdout, stderr io.Writer) error
}

// builtFlows returns the work behind each command.
//
// TEMPORARY: the work is not built yet. Step 7 of the plan replaces
// create with a call to the create flow in internal/slice, step 8
// replaces status, and step 9 replaces list. Until then, each one fails
// with exit code 1.
func builtFlows() flows {
	return flows{
		create: func(context.Context, createInput, io.Writer, io.Writer) error {
			return errNotBuilt("slice create")
		},
		status: func(context.Context, statusInput, io.Writer, io.Writer) error {
			return errNotBuilt("slice status")
		},
		list: func(context.Context, io.Writer, io.Writer) error {
			return errNotBuilt("slice list")
		},
	}
}

// errNotBuilt is the error of a command whose work is not built yet.
// TEMPORARY: it goes away with the last placeholder in builtFlows.
func errNotBuilt(command string) error {
	return errors.New("Error: sliceflow " + command + " is not built yet.")
}
