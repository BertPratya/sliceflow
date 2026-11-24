// Package speccmd holds the spec commands of sliceflow: template, which
// prints the built-in spec template, and validate, which finds the
// problems in a slice's spec.
package speccmd

import "github.com/BertPratya/sliceflow/internal/cli/command"

// Group holds the commands that print the spec template and check specs.
var Group = command.Group{
	Description: "Print the spec template, and find problems in specs",
	Actions: map[string]command.Action{
		"template": {
			Description:     "Print the spec template",
			LongDescription: "Print the spec template built into sliceflow, with its guidance comments.",
			Run:             runTemplate,
		},
	},
}
