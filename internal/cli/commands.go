package cli

import (
	"github.com/BertPratya/sliceflow/internal/cli/command"
	"github.com/BertPratya/sliceflow/internal/cli/slicecmd"
	"github.com/BertPratya/sliceflow/internal/cli/speccmd"
)

// CommandRegistry is the command table: every command group by name.
type CommandRegistry struct {
	Groups map[string]command.Group
}

// commands holds every command of sliceflow. The guide and the help of
// each group are built from it, so a new group or command shows up in
// them with no other change.
var commands = CommandRegistry{
	Groups: map[string]command.Group{
		"slice": slicecmd.Group,
		"spec":  speccmd.Group,
	},
}
