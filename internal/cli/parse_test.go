package cli

import (
	"errors"
	"github.com/BertPratya/sliceflow/internal/cli/slicecmd"
	"maps"
	"testing"
)

// sliceCommand returns the command actionName of the group slice, as
// resolveCommand finds it, with words after it.
func sliceCommand(actionName string, words ...string) found {
	return found{
		groupName:  "slice",
		group:      slicecmd.Group,
		actionName: actionName,
		action:     slicecmd.Group.Actions[actionName],
		rest:       words,
	}
}

// Each case parses words against slice create (one argument, three
// options) or slice list (none), and checks what parseWords gives back.
func TestParseWordsWorks(t *testing.T) {
	tests := []struct {
		about       string
		command     found
		wantArgs    map[string]string
		wantOptions map[string]string
	}{
		{"a name only", sliceCommand("create", "login"),
			map[string]string{"name": "login"}, map[string]string{}},
		{"an option after the name", sliceCommand("create", "login", "--title", "Log in"),
			map[string]string{"name": "login"}, map[string]string{"title": "Log in"}},
		{"an option before the name", sliceCommand("create", "--title", "Log in", "login"),
			map[string]string{"name": "login"}, map[string]string{"title": "Log in"}},
		{"--name=value", sliceCommand("create", "login", "--from=dev"),
			map[string]string{"name": "login"}, map[string]string{"from": "dev"}},
		{"three options", sliceCommand("create", "login", "--title", "T", "--description", "D", "--from", "dev"),
			map[string]string{"name": "login"}, map[string]string{"title": "T", "description": "D", "from": "dev"}},
		{"the last one counts", sliceCommand("create", "login", "--title", "A", "--title=B"),
			map[string]string{"name": "login"}, map[string]string{"title": "B"}},
		{"a value that starts with -", sliceCommand("create", "login", "--title", "--help"),
			map[string]string{"name": "login"}, map[string]string{"title": "--help"}},
		{"an empty value", sliceCommand("create", "login", "--title", ""),
			map[string]string{"name": "login"}, map[string]string{"title": ""}},
		{"an empty value after =", sliceCommand("create", "login", "--title="),
			map[string]string{"name": "login"}, map[string]string{"title": ""}},
		{"no words for a command with none", sliceCommand("list"),
			map[string]string{}, map[string]string{}},
	}
	for _, tt := range tests {
		t.Run(tt.about, func(t *testing.T) {
			words := tt.command.rest
			got, err := parseWords(tt.command)
			if err != nil {
				t.Fatalf("parseWords(%q) returned the error %v, want none", words, err)
			}
			if !maps.Equal(got.Args, tt.wantArgs) {
				t.Errorf("parseWords(%q) gave the arguments %v, want %v", words, got.Args, tt.wantArgs)
			}
			if !maps.Equal(got.Options, tt.wantOptions) {
				t.Errorf("parseWords(%q) gave the options %v, want %v", words, got.Options, tt.wantOptions)
			}
		})
	}
}

// Each case parses words that ask for help, or are wrong, and checks the
// error: errHelp, or a *usageError with the message want. The hints are
// checked by the tests of Run.
func TestParseWordsErrors(t *testing.T) {
	tests := []struct {
		about   string
		command found
		want    string // "help" for errHelp
	}{
		// --help comes before any error.
		{"-h", sliceCommand("create", "-h"), "help"},
		{"--help after a wrong word", sliceCommand("create", "login", "--titel", "--help"), "help"},
		{"--help with no name", sliceCommand("create", "--help"), "help"},

		// An unknown option.
		{"an unknown option", sliceCommand("create", "login", "--titel", "T"), `unknown option "--titel"`},
		{"an unknown option with =", sliceCommand("create", "login", "--titel=T"), `unknown option "--titel"`},
		{"a short option other than -h", sliceCommand("create", "login", "-t", "T"), `unknown option "-t"`},
		{"an option of another command", sliceCommand("list", "--title", "T"), `unknown option "--title"`},

		// An option with no value.
		{"an option as the last word", sliceCommand("create", "login", "--title"), `option "--title" needs a value`},

		// An unexpected word.
		{"a second word", sliceCommand("create", "login", "extra"), `unexpected argument "extra"`},
		{"a word for a command with none", sliceCommand("list", "extra"), `unexpected argument "extra"`},

		// The first wrong word from the left is reported.
		{"a word before an option", sliceCommand("create", "login", "extra", "--titel"), `unexpected argument "extra"`},
		{"an option before a word", sliceCommand("create", "login", "--titel", "x", "extra"), `unknown option "--titel"`},

		// A missing name is checked last.
		{"no words", sliceCommand("create"), `missing slice name`},
		{"only an option", sliceCommand("create", "--title", "T"), `missing slice name`},
		{"an empty name", sliceCommand("create", ""), `missing slice name`},
		{"a wrong option and no name", sliceCommand("create", "--titel"), `unknown option "--titel"`},
	}
	for _, tt := range tests {
		t.Run(tt.about, func(t *testing.T) {
			words := tt.command.rest
			_, err := parseWords(tt.command)
			if tt.want == "help" {
				if !errors.Is(err, errHelp) {
					t.Errorf("parseWords(%q) returned %v, want errHelp", words, err)
				}
				return
			}
			var usage *usageError
			if !errors.As(err, &usage) {
				t.Fatalf("parseWords(%q) returned %v, want a *usageError", words, err)
			}
			if usage.message != tt.want {
				t.Errorf("parseWords(%q) returned the message %q, want %q", words, usage.message, tt.want)
			}
		})
	}
}
