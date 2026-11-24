package cli

import (
	"strings"

	"github.com/BertPratya/sliceflow/internal/cli/command"
)

// found is what resolveCommand finds in the words of a command line.
type found struct {
	groupName  string
	group      command.Group
	actionName string // "" when the words name only a group
	action     command.Action
	rest       []string // the words after the command words
}

// resolveCommand finds the command that args names in registry. The
// command words are the words before the first option: a group, then one
// of its commands. When args names only a group, as in "slice" or
// "slice --help", the result has no action. When the words name no group,
// or no command of the group, it returns a *usageError. args must not be
// empty.
func resolveCommand(registry CommandRegistry, args []string) (found, error) {
	group, ok := registry.Groups[args[0]]
	if !ok {
		return found{}, unknownCommandError(args[0])
	}
	if len(args) == 1 || isOption(args[1]) {
		return found{groupName: args[0], group: group, rest: args[1:]}, nil
	}

	action, ok := group.Actions[args[1]]
	if !ok {
		return found{}, unknownCommandError(args[0] + " " + args[1])
	}
	return found{
		groupName:  args[0],
		group:      group,
		actionName: args[1],
		action:     action,
		rest:       args[2:],
	}, nil
}

// isOption reports whether word is an option: a word that starts with
// "-".
func isOption(word string) bool {
	return strings.HasPrefix(word, "-")
}
