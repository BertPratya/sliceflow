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

// newCreateInput turns the call of `slice create` into createInput, with
// "main" when --from was not given (R10).
func newCreateInput(c call) createInput {
	from, fromGiven := c.options["from"]
	if !fromGiven {
		from = "main"
	}
	return createInput{
		name:        c.name,
		title:       c.options["title"],
		description: c.options["description"],
		from:        from,
		fromGiven:   fromGiven,
	}
}

// The work behind each command. Each one prints what the command shows
// when it works to stdout, and its warnings to stderr. It returns an
// error when the command fails: Run prints the error's text as it is,
// so the text is the whole message, one or more full lines with no line
// break at the end. An error that is wrong use (see wrongUser) gives
// exit code 2, any other error 1.
//
// TEMPORARY: the work is not built yet, so each one fails with exit
// code 1. Step 7 of the plan replaces createSlice with a call to the
// create flow in internal/slice, step 8 replaces showSlice, and step 9
// replaces listSlices.

// createSlice makes a new slice (Story 1).
func createSlice(_ context.Context, _ createInput, _, _ io.Writer) error {
	return errNotBuilt("slice create")
}

// showSlice shows the slice with the given name (S45).
func showSlice(_ context.Context, _ string, _, _ io.Writer) error {
	return errNotBuilt("slice status")
}

// listSlices shows all slices (R69).
func listSlices(_ context.Context, _, _ io.Writer) error {
	return errNotBuilt("slice list")
}

// errNotBuilt is the error of a command whose work is not built yet.
// TEMPORARY: it goes away with the last placeholder above.
func errNotBuilt(command string) error {
	return errors.New("Error: sliceflow " + command + " is not built yet.")
}
