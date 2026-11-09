package e2e

import "testing"

// The second lines of the messages of R70 to R75.
const (
	seeCommands   = "To see the commands, run: sliceflow --help\n"
	createOptions = "To see the options, run: sliceflow slice create --help\n"
	createUsage   = "Usage: sliceflow slice create <name> [options]\n"
	statusUsage   = "Usage: sliceflow slice status <name> [options]\n"
	listUsage     = "Usage: sliceflow slice list [options]\n"
)

// TestWrongUse checks S46 to S50 and R70 to R76: a wrong command line
// prints the two lines of its requirement to the error output, nothing
// to the normal output, changes nothing, and exits with code 2.
func TestWrongUse(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		stderr string
	}{
		// R70: an unknown command, at both levels.
		{"S46 slice delete",
			[]string{"slice", "delete", "card-payment"},
			"Error: unknown command \"slice delete\". Nothing was changed.\n" + seeCommands},
		{"unknown command at the top",
			[]string{"foo"},
			"Error: unknown command \"foo\". Nothing was changed.\n" + seeCommands},
		{"help as a command",
			[]string{"help"},
			"Error: unknown command \"help\". Nothing was changed.\n" + seeCommands},
		{"completion as a command",
			[]string{"completion", "bash"},
			"Error: unknown command \"completion\". Nothing was changed.\n" + seeCommands},
		{"no guess for a near miss",
			[]string{"slice", "lst"},
			"Error: unknown command \"slice lst\". Nothing was changed.\n" + seeCommands},

		// R72: an unknown option.
		{"S47 --titel",
			[]string{"slice", "create", "card-payment", "--titel", "Pay with a card"},
			"Error: unknown option \"--titel\". Nothing was changed.\n" + createOptions},
		{"--titel before the name",
			[]string{"slice", "create", "--titel", "Pay", "card-payment"},
			"Error: unknown option \"--titel\". Nothing was changed.\n" + createOptions},
		{"--titel=",
			[]string{"slice", "create", "card-payment", "--titel=Pay"},
			"Error: unknown option \"--titel\". Nothing was changed.\n" + createOptions},
		{"--version",
			[]string{"--version"},
			"Error: unknown option \"--version\". Nothing was changed.\nTo see the options, run: sliceflow --help\n"},
		{"unknown option of status",
			[]string{"slice", "status", "card-payment", "--from", "main"},
			"Error: unknown option \"--from\". Nothing was changed.\nTo see the options, run: sliceflow slice status --help\n"},
		{"unknown option of list",
			[]string{"slice", "list", "-v"},
			"Error: unknown option \"-v\". Nothing was changed.\nTo see the options, run: sliceflow slice list --help\n"},

		// R73: a missing name.
		{"S48 create with no name",
			[]string{"slice", "create"},
			"Error: missing slice name. Nothing was changed.\n" + createUsage},
		{"status with no name",
			[]string{"slice", "status"},
			"Error: missing slice name. Nothing was changed.\n" + statusUsage},
		{"create with an empty name",
			[]string{"slice", "create", ""},
			"Error: missing slice name. Nothing was changed.\n" + createUsage},
		{"status with an empty name",
			[]string{"slice", "status", ""},
			"Error: missing slice name. Nothing was changed.\n" + statusUsage},
		{"create with only options",
			[]string{"slice", "create", "--title", "Pay"},
			"Error: missing slice name. Nothing was changed.\n" + createUsage},

		// R74: an option with no value.
		{"S49 --title with no value",
			[]string{"slice", "create", "card-payment", "--title"},
			"Error: option \"--title\" needs a value. Nothing was changed.\n" + createOptions},
		{"--from with no value",
			[]string{"slice", "create", "card-payment", "--from"},
			"Error: option \"--from\" needs a value. Nothing was changed.\n" + createOptions},

		// R75: an unexpected argument.
		{"S50 extra",
			[]string{"slice", "create", "card-payment", "extra"},
			"Error: unexpected argument \"extra\". Nothing was changed.\n" + createUsage},
		{"extra for status",
			[]string{"slice", "status", "card-payment", "extra"},
			"Error: unexpected argument \"extra\". Nothing was changed.\n" + statusUsage},
		{"extra for list",
			[]string{"slice", "list", "extra"},
			"Error: unexpected argument \"extra\". Nothing was changed.\n" + listUsage},
		{"the first extra word",
			[]string{"slice", "create", "card-payment", "one", "two"},
			"Error: unexpected argument \"one\". Nothing was changed.\n" + createUsage},

		// R40: the first wrong word from the left; a missing name last.
		{"unexpected argument before an unknown option",
			[]string{"slice", "create", "card-payment", "extra", "--titel"},
			"Error: unexpected argument \"extra\". Nothing was changed.\n" + createUsage},
		{"unknown option before an unexpected argument",
			[]string{"slice", "create", "card-payment", "--titel", "extra"},
			"Error: unknown option \"--titel\". Nothing was changed.\n" + createOptions},
		{"unknown option before a missing name",
			[]string{"slice", "create", "--titel"},
			"Error: unknown option \"--titel\". Nothing was changed.\n" + createOptions},
		{"unknown command before an unknown option",
			[]string{"foo", "--version"},
			"Error: unknown command \"foo\". Nothing was changed.\n" + seeCommands},
		{"unknown option before a command",
			[]string{"--version", "slice", "list"},
			"Error: unknown option \"--version\". Nothing was changed.\nTo see the options, run: sliceflow --help\n"},
		{"option of create before the command",
			[]string{"slice", "--title", "Pay", "create", "card-payment"},
			"Error: unknown option \"--title\". Nothing was changed.\nTo see the options, run: sliceflow slice --help\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check(t, tt.args, result{stderr: tt.stderr, code: 2})
		})
	}
}
