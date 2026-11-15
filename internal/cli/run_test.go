package cli

import (
	"bytes"
	"context"
	"testing"
)

// wantGuide is the guide of S41, word for word. The test keeps its own
// copy, so a change to the guide's wording must change the spec too.
const wantGuide = `sliceflow runs a spec-driven workflow with AI coding agents.

Usage:
  sliceflow <command> [options]

Commands:
  slice  Make and see slices

Options:
  -h, --help  Show this help.
`

// wantSliceHelp is the help of the group slice, in S43, word for word.
// Its commands are in alphabetical order (SD30).
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

// wantCreateHelp is the help of slice create, in S44, word for word.
const wantCreateHelp = `Make a new slice: a branch, a worktree, a spec, and an event log.

Usage:
  sliceflow slice create <name> [options]

Options:
  --description <text>  One or two sentences about the slice.
  --from <branch>       The branch to start from. Default: main.
  --title <text>        A short title for the slice. Default: the name.
  -h, --help            Show this help.
`

// wantStatusHelp is the help of slice status, in S45, word for word.
const wantStatusHelp = `Show one slice: its title, description, start, branch, worktree, and events.

Usage:
  sliceflow slice status <name> [options]

Options:
  -h, --help  Show this help.
`

// wantListHelp is the help of slice list, as R69 gives it.
const wantListHelp = `Show all slices in a table: name, title, and description.

Usage:
  sliceflow slice list [options]

Options:
  -h, --help  Show this help.
`

// unknownCommand is the error output of R70 for words.
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
		// R67: the guide, and the help of the group slice.
		{"no words", nil, 0, wantGuide, ""},
		{"--help", []string{"--help"}, 0, wantGuide, ""},
		{"-h", []string{"-h"}, 0, wantGuide, ""},
		{"slice alone", []string{"slice"}, 0, wantSliceHelp, ""},
		{"slice --help", []string{"slice", "--help"}, 0, wantSliceHelp, ""},
		{"slice -h", []string{"slice", "-h"}, 0, wantSliceHelp, ""},

		// R68, R69: the help of a command. --help may come after the
		// name, because it comes before any error (R40).
		{"slice create -h", []string{"slice", "create", "-h"}, 0, wantCreateHelp, ""},
		{"slice create --help", []string{"slice", "create", "--help"}, 0, wantCreateHelp, ""},
		{"slice create, a name, then --help", []string{"slice", "create", "login", "--help"}, 0, wantCreateHelp, ""},
		{"slice status --help", []string{"slice", "status", "--help"}, 0, wantStatusHelp, ""},
		{"slice list -h", []string{"slice", "list", "-h"}, 0, wantListHelp, ""},

		// R70: an unknown command.
		{"an unknown word", []string{"nosuch"}, 2, "", unknownCommand("nosuch")},
		{"an unknown word, then more", []string{"nosuch", "extra"}, 2, "", unknownCommand("nosuch")},
		{"slice, then an unknown word", []string{"slice", "nosuch"}, 2, "", unknownCommand("slice nosuch")},
		{"an unknown command with --help", []string{"slice", "nosuch", "--help"}, 2, "", unknownCommand("slice nosuch")},

		// A known command reaches its function. For now, every command
		// is the placeholder notBuilt.
		{"slice create", []string{"slice", "create", "login"}, 1, "", "Error: not built yet.\n"},
		{"slice status", []string{"slice", "status", "login"}, 1, "", "Error: not built yet.\n"},
		{"slice list", []string{"slice", "list"}, 1, "", "Error: not built yet.\n"},
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
