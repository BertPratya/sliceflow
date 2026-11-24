// Package command holds what every command group of sliceflow shares:
// the shape of a command in the command table, what a command gets from
// the command line, and the helpers that print failures and warnings.
package command

import (
	"context"
	"io"
)

// Group is one command group, e.g. slice: its description in the guide,
// and its commands by name.
type Group struct {
	Description string
	Actions     map[string]Action
}

// Action is one command of a group: its descriptions for the help, its
// arguments and options, and the function that runs it. Run returns the
// exit code.
type Action struct {
	Description     string
	LongDescription string
	Args            []Arg
	Options         map[string]Option
	Run             func(ctx context.Context, in Input, stdout, stderr io.Writer) int
}

// Arg is one argument of a command, in its order on the command line.
type Arg struct {
	Name        string
	Description string
	Required    bool
}

// Option is one option of a command. An option with a ValueName takes a
// value, e.g. --title <text>.
type Option struct {
	Description string
	ValueName   string
	Short       string
	Type        string
	Required    bool
	Default     any
}

// Input is what was read from the words after a command. A command's
// Run gets it, so it never reads the words itself.
type Input struct {
	// Args holds each argument by its name in the command table, e.g.
	// Args["name"] is "login".
	Args map[string]string
	// Options holds each option that was given, by its name without
	// "--", e.g. Options["title"] is "Log in". An option that was not
	// given is not in it, so v, ok := Options["title"] tells "not given"
	// from "given as an empty text".
	Options map[string]string
}
