// Package slicecmd holds the slice commands of sliceflow: create, list,
// and status. Each reads what the command line gave it, calls the slice
// package, and prints the result or the error.
package slicecmd

import "github.com/BertPratya/sliceflow/internal/cli/command"

// Group holds the commands that make and show slices.
var Group = command.Group{
	Description: "Make and see slices",
	Actions: map[string]command.Action{
		"create": {
			Description:     "Make a new slice",
			LongDescription: "Make a new slice on the base branch: a spec and an event log, in one commit.",
			Args:            []command.Arg{{Name: "name", Description: "Up to 40 letters, digits, - and _. Starts with a letter or a digit.", Required: true}},
			Options: map[string]command.Option{
				"title":       {Description: "A short title for the slice. Default: the name.", ValueName: "text", Type: "string"},
				"description": {Description: "One or two sentences about the slice.", ValueName: "text", Type: "string"},
				"from":        {Description: "The base branch. Default: the sliceflow.base setting, else main.", ValueName: "branch", Type: "string"},
			},
			Run: runSliceCreate,
		},
		"edit": {
			Description:     "Change a slice's title or description",
			LongDescription: "Change a slice's title, its description, or both, in one commit on its base branch.",
			Args:            []command.Arg{{Name: "name", Description: "The name of the slice to change.", Required: true}},
			Options: map[string]command.Option{
				"title":       {Description: "The new title. Empty sets it back to the name.", ValueName: "text", Type: "string"},
				"description": {Description: "The new description. Empty removes it.", ValueName: "text", Type: "string"},
			},
			Run: runSliceEdit,
		},
		"list": {
			Description:     "Show all slices",
			LongDescription: "Show all slices in a table: name, title, and description.",
			Run:             runSliceList,
		},
		"status": {
			Description:     "Show one slice",
			LongDescription: "Show one slice: its state, the next step, its title, description, creation time, base, and events.",
			Args:            []command.Arg{{Name: "name", Description: "The name of the slice to show.", Required: true}},
			Run:             runSliceStatus,
		},
	},
}
