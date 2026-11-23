package cli

import (
	"context"
	"io"
)

type CommandRegistry struct {
	Groups map[string]CommandGroup
}

type CommandGroup struct {
	Description string
	Actions     map[string]CommandAction
}

type CommandAction struct {
	Description     string
	LongDescription string
	Args            []CommandArg
	Options         map[string]CommandOption
	Run             func(ctx context.Context, in input, stdout, stderr io.Writer) int
}

type CommandArg struct {
	Name        string
	Description string
	Required    bool
}

type CommandOption struct {
	Description string
	ValueName   string
	Short       string
	Type        string
	Required    bool
	Default     any
}

// commands holds every command of sliceflow. The guide and the help of
// each group are built from it, so a new group or command shows up in
// them with no other change.
var commands = CommandRegistry{
	Groups: map[string]CommandGroup{
		"slice": sliceGroup,
	},
}

// sliceGroup holds the commands that make and show slices.
var sliceGroup = CommandGroup{
	Description: "Make and see slices",
	Actions: map[string]CommandAction{
		"create": {
			Description:     "Make a new slice",
			LongDescription: "Make a new slice on the base branch: a spec and an event log, in one commit.",
			Args:            []CommandArg{{Name: "name", Description: "Up to 40 letters, digits, - and _. Starts with a letter or a digit.", Required: true}},
			Options: map[string]CommandOption{
				"title":       {Description: "A short title for the slice. Default: the name.", ValueName: "text", Type: "string"},
				"description": {Description: "One or two sentences about the slice.", ValueName: "text", Type: "string"},
				"from":        {Description: "The base branch. Default: the sliceflow.base setting, else main.", ValueName: "branch", Type: "string"},
			},
			Run: runSliceCreate,
		},
		"list": {
			Description:     "Show all slices",
			LongDescription: "Show all slices in a table: name, title, and description.",
			Run:             runSliceList,
		},
		"status": {
			Description:     "Show one slice",
			LongDescription: "Show one slice: its title, description, creation time, base, and events.",
			Args:            []CommandArg{{Name: "name", Description: "The name of the slice to show.", Required: true}},
			Run:             runSliceStatus,
		},
	},
}
