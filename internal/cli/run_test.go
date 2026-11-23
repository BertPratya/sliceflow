package cli

import (
	"bytes"
	"context"
	"testing"
)

// wantGuide is the guide, word for word. The test keeps its own copy, so
// a change to the wording fails until the test is changed too.
const wantGuide = `sliceflow runs a spec-driven workflow with AI coding agents.

Usage:
  sliceflow <command> [options]

Commands:
  slice  Make and see slices

Options:
  -h, --help  Show this help.
`

// wantSliceHelp is the help of the group slice, word for word. Its
// commands are in alphabetical order.
const wantSliceHelp = `Make and see slices.

Usage:
  sliceflow slice <command> [options]

Commands:
  create <name>  Make a new slice
  list           Show all slices
  status <name>  Show one slice

Options:
  -h, --help  Show this help.
`

// wantCreateHelp is the help of slice create, word for word.
const wantCreateHelp = `Make a new slice on the base branch: a spec and an event log, in one commit.

Usage:
  sliceflow slice create <name> [options]

Arguments:
  <name>  Up to 40 letters, digits, - and _. Starts with a letter or a digit.

Options:
  --description <text>  One or two sentences about the slice.
  --from <branch>       The base branch. Default: the sliceflow.base setting, else main.
  --title <text>        A short title for the slice. Default: the name.
  -h, --help            Show this help.
`

// wantStatusHelp is the help of slice status, word for word.
const wantStatusHelp = `Show one slice: its title, description, creation time, base, and events.

Usage:
  sliceflow slice status <name> [options]

Arguments:
  <name>  The name of the slice to show.

Options:
  -h, --help  Show this help.
`

// wantListHelp is the help of slice list, word for word.
const wantListHelp = `Show all slices in a table: name, title, and description.

Usage:
  sliceflow slice list [options]

Options:
  -h, --help  Show this help.
`

// unknownCommand is the error output for words that name no command.
func unknownCommand(words string) string {
	return "Error: unknown command \"" + words + "\". Nothing was changed.\n" +
		"To see the commands, run: sliceflow --help\n"
}

// Each case runs Run with words, and checks its exit code, its normal
// output, and its error output.
func TestRunRouting(t *testing.T) {
	tests := []struct {
		about      string
		words      []string
		wantCode   int
		wantStdout string
		wantStderr string
	}{
		// The guide, and the help of the group slice.
		{"no words", nil, 0, wantGuide, ""},
		{"--help", []string{"--help"}, 0, wantGuide, ""},
		{"-h", []string{"-h"}, 0, wantGuide, ""},
		{"slice alone", []string{"slice"}, 0, wantSliceHelp, ""},
		{"slice --help", []string{"slice", "--help"}, 0, wantSliceHelp, ""},
		{"slice -h", []string{"slice", "-h"}, 0, wantSliceHelp, ""},

		// The help of a command. --help may come after the name,
		// because it comes before any error.
		{"slice create -h", []string{"slice", "create", "-h"}, 0, wantCreateHelp, ""},
		{"slice create --help", []string{"slice", "create", "--help"}, 0, wantCreateHelp, ""},
		{"slice create, a name, then --help", []string{"slice", "create", "login", "--help"}, 0, wantCreateHelp, ""},
		{"slice status --help", []string{"slice", "status", "--help"}, 0, wantStatusHelp, ""},
		{"slice list -h", []string{"slice", "list", "-h"}, 0, wantListHelp, ""},

		// An unknown command.
		{"an unknown word", []string{"nosuch"}, 2, "", unknownCommand("nosuch")},
		{"an unknown word, then more", []string{"nosuch", "extra"}, 2, "", unknownCommand("nosuch")},
		{"slice, then an unknown word", []string{"slice", "nosuch"}, 2, "", unknownCommand("slice nosuch")},
		{"an unknown command with --help", []string{"slice", "nosuch", "--help"}, 2, "", unknownCommand("slice nosuch")},

		// A wrong command line, with exit code 2.
		{"an option with no command", []string{"--version"}, 2, "",
			"Error: unknown option \"--version\". Nothing was changed.\nTo see the options, run: sliceflow --help\n"},
		{"an unknown option of a group", []string{"slice", "--bogus"}, 2, "",
			"Error: unknown option \"--bogus\". Nothing was changed.\nTo see the options, run: sliceflow slice --help\n"},
		{"an unknown option", []string{"slice", "create", "card-payment", "--titel", "Pay with a card"}, 2, "",
			"Error: unknown option \"--titel\". Nothing was changed.\nTo see the options, run: sliceflow slice create --help\n"},
		{"no name", []string{"slice", "create"}, 2, "",
			"Error: missing slice name. Nothing was changed.\nUsage: sliceflow slice create <name> [options]\n"},
		{"an empty name", []string{"slice", "status", ""}, 2, "",
			"Error: missing slice name. Nothing was changed.\nUsage: sliceflow slice status <name> [options]\n"},
		{"an option with no value", []string{"slice", "create", "card-payment", "--title"}, 2, "",
			"Error: option \"--title\" needs a value. Nothing was changed.\nTo see the options, run: sliceflow slice create --help\n"},
		{"an unexpected argument", []string{"slice", "create", "card-payment", "extra"}, 2, "",
			"Error: unexpected argument \"extra\". Nothing was changed.\nUsage: sliceflow slice create <name> [options]\n"},
		{"an unexpected argument for list", []string{"slice", "list", "extra"}, 2, "",
			"Error: unexpected argument \"extra\". Nothing was changed.\nUsage: sliceflow slice list [options]\n"},

		// The commands themselves are tested in e2e, each in a repo of
		// its own: run from here, they would work on sliceflow's own repo.
	}
	for _, tt := range tests {
		t.Run(tt.about, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := Run(context.Background(), tt.words, &stdout, &stderr)

			if code != tt.wantCode {
				t.Errorf("Run(%q) returned %d, want %d", tt.words, code, tt.wantCode)
			}
			if got := stdout.String(); got != tt.wantStdout {
				t.Errorf("Run(%q) printed to stdout:\n%s\nwant:\n%s", tt.words, got, tt.wantStdout)
			}
			if got := stderr.String(); got != tt.wantStderr {
				t.Errorf("Run(%q) printed to stderr:\n%s\nwant:\n%s", tt.words, got, tt.wantStderr)
			}
		})
	}
}
