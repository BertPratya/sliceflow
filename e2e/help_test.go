package e2e

import "testing"

// The help texts of S41, S44, S45, and R69, written out here again, and
// not taken from internal/cli, so that a change there is caught.
const (
	guide = `sliceflow runs a spec-driven workflow with AI coding agents.

Usage:
  sliceflow <command> [options]

Commands:
  slice create <name>  Make a new slice
  slice status <name>  Show one slice
  slice list           Show all slices

For help with a command, run: sliceflow <command> --help
`

	createHelp = `Make a new slice: a branch, a worktree, a spec, and an event log.

Usage:
  sliceflow slice create <name> [options]

Options:
  --title <text>        A short title for the slice. Default: the name.
  --description <text>  One or two sentences about the slice.
  --from <branch>       The branch to start from. Default: main.
  --help                Show this help.
`

	statusHelp = `Show one slice: its title, description, start, branch, worktree, and events.

Usage:
  sliceflow slice status <name> [options]

Options:
  --help  Show this help.
`

	listHelp = `Show all slices in a table: name, title, and description.

Usage:
  sliceflow slice list [options]

Options:
  --help  Show this help.
`
)

// TestGuide checks S41, S42, S43, and R67: the guide, to the normal
// output, with exit code 0.
func TestGuide(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"S41 --help", []string{"--help"}},
		{"-h", []string{"-h"}},
		{"S42 no command", nil},
		{"S43 slice", []string{"slice"}},
		{"slice --help", []string{"slice", "--help"}},
		{"slice -h", []string{"slice", "-h"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check(t, tt.args, result{stdout: guide, code: 0})
		})
	}
}

// TestCommandHelp checks S44, S45, R68, and R69: the help of each
// command, with --help and -h, to the normal output, with exit code 0.
func TestCommandHelp(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"S44 create --help", []string{"slice", "create", "--help"}, createHelp},
		{"create -h", []string{"slice", "create", "-h"}, createHelp},
		{"S45 status --help", []string{"slice", "status", "--help"}, statusHelp},
		{"status -h", []string{"slice", "status", "-h"}, statusHelp},
		{"R69 list --help", []string{"slice", "list", "--help"}, listHelp},
		{"list -h", []string{"slice", "list", "-h"}, listHelp},
		{"create --help after the name and options", []string{"slice", "create", "card-payment", "--title", "Pay", "--help"}, createHelp},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check(t, tt.args, result{stdout: tt.want, code: 0})
		})
	}
}

// TestHelpBeforeAnError checks R40: --help comes before any error on the
// same line.
func TestHelpBeforeAnError(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"after an unknown option", []string{"slice", "create", "card-payment", "--titel", "x", "--help"}, createHelp},
		{"before an unknown option", []string{"slice", "create", "--help", "--titel"}, createHelp},
		{"after an unexpected argument", []string{"slice", "create", "card-payment", "extra", "-h"}, createHelp},
		{"with a missing name", []string{"slice", "status", "--help"}, statusHelp},
		{"with an option with no value", []string{"slice", "create", "card-payment", "-h", "--title"}, createHelp},
		{"after an unexpected argument of list", []string{"slice", "list", "extra", "--help"}, listHelp},
		{"after an unknown option at the top", []string{"--version", "--help"}, guide},
		{"after an unknown command", []string{"slice", "delete", "--help"}, guide},
		{"after an unknown command and real command words", []string{"foo", "slice", "create", "--help"}, guide},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check(t, tt.args, result{stdout: tt.want, code: 0})
		})
	}
}
