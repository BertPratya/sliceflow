package cli

import (
	"bytes"
	"context"
	"testing"
)

// runWords runs args with a recorder, and returns the exit code, the
// two outputs, and the recorder.
func runWords(args ...string) (int, string, string, *recorder) {
	r := &recorder{}
	var stdout, stderr bytes.Buffer
	code := run(context.Background(), args, &stdout, &stderr, r.flows())
	return code, stdout.String(), stderr.String(), r
}

// TestHelpComesFirst checks R40 and R71: --help or -h anywhere gives the
// help of the command, before any problem on the same line, and the
// work does not run.
func TestHelpComesFirst(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"-h", []string{"slice", "create", "-h"}, createHelp},
		{"before the name", []string{"slice", "create", "--help", "login"}, createHelp},
		{"after an unknown option", []string{"slice", "create", "login", "--titel", "--help"}, createHelp},
		{"after an option with no value yet", []string{"slice", "create", "--titel", "x", "-h"}, createHelp},
		{"after an unexpected argument", []string{"slice", "create", "a", "b", "--help"}, createHelp},
		{"with a missing name", []string{"slice", "status", "--help"}, statusHelp},
		{"after --", []string{"slice", "list", "--", "--help"}, listHelp},
		{"before the command", []string{"--help", "slice", "list"}, listHelp},
		{"slice", []string{"slice", "-h"}, guideHelp},
		{"after an unknown command", []string{"slice", "delete", "x", "--help"}, guideHelp},
		// After an unknown command, the words are not commands any more,
		// so the help is the guide, not the help of create.
		{"after an unknown command and real command words", []string{"foo", "slice", "create", "--help"}, guideHelp},
		{"after an unknown command and a wrong option", []string{"foo", "--bad", "slice", "list", "-h"}, guideHelp},
		{"after an unknown option at the top", []string{"--version", "--help"}, guideHelp},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, stdout, stderr, r := runWords(tt.args...)
			if code != 0 || stdout != tt.want || stderr != "" {
				t.Errorf("code %d, stdout %q, stderr %q; want 0 and the help", code, stdout, stderr)
			}
			if r.create != nil || r.status != nil || r.list {
				t.Error("the work ran")
			}
		})
	}
}

// TestFirstProblem checks R40 and R70 to R75: the first wrong word from
// the left is reported, a missing name is checked last, and the work
// does not run.
func TestFirstProblem(t *testing.T) {
	const (
		createOptions = "To see the options, run: sliceflow slice create --help"
		createUsage   = "Usage: sliceflow slice create <name> [options]"
	)
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"unexpected argument before an unknown option",
			[]string{"slice", "create", "a", "extra", "--titel"},
			"Error: unexpected argument \"extra\". Nothing was changed.\n" + createUsage + "\n"},
		{"unknown option before an unexpected argument",
			[]string{"slice", "create", "a", "--titel", "extra"},
			"Error: unknown option \"--titel\". Nothing was changed.\n" + createOptions + "\n"},
		{"unknown option before a missing value",
			[]string{"slice", "create", "--bad", "--title"},
			"Error: unknown option \"--bad\". Nothing was changed.\n" + createOptions + "\n"},
		{"unknown option before a missing name",
			[]string{"slice", "create", "--bad"},
			"Error: unknown option \"--bad\". Nothing was changed.\n" + createOptions + "\n"},
		{"missing value before a missing name",
			[]string{"slice", "create", "--from"},
			"Error: option \"--from\" needs a value. Nothing was changed.\n" + createOptions + "\n"},
		{"unknown option with =",
			[]string{"slice", "create", "a", "--titel=Pay"},
			"Error: unknown option \"--titel\". Nothing was changed.\n" + createOptions + "\n"},
		{"--help with a value",
			[]string{"slice", "create", "a", "--help=yes"},
			"Error: unknown option \"--help=yes\". Nothing was changed.\n" + createOptions + "\n"},
		{"a single -",
			[]string{"slice", "create", "a", "-"},
			"Error: unknown option \"-\". Nothing was changed.\n" + createOptions + "\n"},
		{"--",
			[]string{"slice", "create", "a", "--", "b"},
			"Error: unknown option \"--\". Nothing was changed.\n" + createOptions + "\n"},
		{"a short option group",
			[]string{"slice", "create", "a", "-hx"},
			"Error: unknown option \"-hx\". Nothing was changed.\n" + createOptions + "\n"},
		{"a go test option",
			[]string{"slice", "create", "a", "-test.v"},
			"Error: unknown option \"-test.v\". Nothing was changed.\n" + createOptions + "\n"},
		{"three dashes",
			[]string{"slice", "create", "a", "---title", "x"},
			"Error: unknown option \"---title\". Nothing was changed.\n" + createOptions + "\n"},
		{"empty name and an extra word",
			[]string{"slice", "create", "", "extra"},
			"Error: unexpected argument \"extra\". Nothing was changed.\n" + createUsage + "\n"},
		{"empty name",
			[]string{"slice", "create", "", "--title", "x"},
			"Error: missing slice name. Nothing was changed.\n" + createUsage + "\n"},
		{"status option",
			[]string{"slice", "status", "a", "--title", "x"},
			"Error: unknown option \"--title\". Nothing was changed.\nTo see the options, run: sliceflow slice status --help\n"},
		{"list argument",
			[]string{"slice", "list", ""},
			"Error: unexpected argument \"\". Nothing was changed.\nUsage: sliceflow slice list [options]\n"},
		{"option before the command",
			[]string{"--title", "x", "slice", "create", "a"},
			"Error: unknown option \"--title\". Nothing was changed.\nTo see the options, run: sliceflow --help\n"},
		{"option of slice",
			[]string{"slice", "--from", "x", "create", "a"},
			"Error: unknown option \"--from\". Nothing was changed.\nTo see the options, run: sliceflow slice --help\n"},
		{"unknown command before an unknown option",
			[]string{"foo", "--version"},
			"Error: unknown command \"foo\". Nothing was changed.\nTo see the commands, run: sliceflow --help\n"},
		{"help as a command",
			[]string{"help"},
			"Error: unknown command \"help\". Nothing was changed.\nTo see the commands, run: sliceflow --help\n"},
		{"completion as a command",
			[]string{"completion", "bash"},
			"Error: unknown command \"completion\". Nothing was changed.\nTo see the commands, run: sliceflow --help\n"},
		{"__complete as a command",
			[]string{"__complete", "slice", ""},
			"Error: unknown command \"__complete\". Nothing was changed.\nTo see the commands, run: sliceflow --help\n"},
		{"two command words in one argument",
			[]string{"slice create", "a"},
			"Error: unknown command \"slice create\". Nothing was changed.\nTo see the commands, run: sliceflow --help\n"},
		{"an empty command word",
			[]string{"slice", "", "list"},
			"Error: unknown command \"slice \". Nothing was changed.\nTo see the commands, run: sliceflow --help\n"},
		{"a command word with a tab after it",
			[]string{"slice", "list\t"},
			"Error: unknown command \"slice list\t\". Nothing was changed.\nTo see the commands, run: sliceflow --help\n"},
		{"the start of a command word",
			[]string{"slice", "cre", "a"},
			"Error: unknown command \"slice cre\". Nothing was changed.\nTo see the commands, run: sliceflow --help\n"},
		{"the start of a group word",
			[]string{"sli"},
			"Error: unknown command \"sli\". Nothing was changed.\nTo see the commands, run: sliceflow --help\n"},
		{"a command in capitals",
			[]string{"slice", "List"},
			"Error: unknown command \"slice List\". Nothing was changed.\nTo see the commands, run: sliceflow --help\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, stdout, stderr, r := runWords(tt.args...)
			if code != 2 || stdout != "" || stderr != tt.want {
				t.Errorf("code %d, stdout %q, stderr %q; want 2, nothing, %q", code, stdout, stderr, tt.want)
			}
			if r.create != nil || r.status != nil || r.list {
				t.Error("the work ran")
			}
		})
	}
}

// TestNoCommand checks R67: `sliceflow` and `sliceflow slice` with no
// command print the guide.
func TestNoCommand(t *testing.T) {
	for _, args := range [][]string{nil, {}, {"slice"}} {
		code, stdout, stderr, _ := runWords(args...)
		if code != 0 || stdout != guideHelp || stderr != "" {
			t.Errorf("run %q: code %d, stdout %q, stderr %q; want 0 and the guide", args, code, stdout, stderr)
		}
	}
}
