package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"
)

// readLine reads args with checkLine, and fails the test now when the
// command line has a problem or asks for help.
func readLine(t *testing.T, args ...string) line {
	t.Helper()
	found, err := checkLine(commands, args)
	if err != nil || found.wantsHelp {
		t.Fatalf("checkLine %q: problem %v, help %v; want neither", args, err, found.wantsHelp)
	}
	return found
}

// TestCreateOptions checks R71: options before or after the name,
// --option=<text>, the word after an option is its value, and the last
// one counts. It also checks that newCreateInput gives "main" when
// --from was not given (R10), and tells an empty --from= from none.
func TestCreateOptions(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want createInput
	}{
		{
			name: "no options",
			args: []string{"slice", "create", "login"},
			want: createInput{name: "login", from: "main"},
		},
		{
			name: "options after the name",
			args: []string{"slice", "create", "login", "--title", "Log in", "--description", "A form.", "--from", "dev"},
			want: createInput{name: "login", title: "Log in", description: "A form.", from: "dev", fromGiven: true},
		},
		{
			name: "options before the name",
			args: []string{"slice", "create", "--title", "Log in", "--from", "dev", "login"},
			want: createInput{name: "login", title: "Log in", from: "dev", fromGiven: true},
		},
		{
			name: "option with =",
			args: []string{"slice", "create", "login", "--title=Log in", "--description=A form.", "--from=dev"},
			want: createInput{name: "login", title: "Log in", description: "A form.", from: "dev", fromGiven: true},
		},
		{
			name: "empty value with =",
			args: []string{"slice", "create", "login", "--title="},
			want: createInput{name: "login", from: "main"},
		},
		{
			name: "value with = inside",
			args: []string{"slice", "create", "login", "--title=a=b"},
			want: createInput{name: "login", title: "a=b", from: "main"},
		},
		{
			name: "value that starts with -",
			args: []string{"slice", "create", "login", "--title", "-x"},
			want: createInput{name: "login", title: "-x", from: "main"},
		},
		{
			name: "value that is --help",
			args: []string{"slice", "create", "login", "--title", "--help"},
			want: createInput{name: "login", title: "--help", from: "main"},
		},
		{
			name: "value that is -h",
			args: []string{"slice", "create", "--description", "-h", "login"},
			want: createInput{name: "login", description: "-h", from: "main"},
		},
		{
			name: "value that is a command word",
			args: []string{"slice", "create", "--title", "create", "slice"},
			want: createInput{name: "slice", title: "create", from: "main"},
		},
		{
			name: "last one counts",
			args: []string{"slice", "create", "login", "--title", "A", "--title=B", "--from", "x", "--from", "y"},
			want: createInput{name: "login", title: "B", from: "y", fromGiven: true},
		},
		{
			name: "--from main given",
			args: []string{"slice", "create", "login", "--from", "main"},
			want: createInput{name: "login", from: "main", fromGiven: true},
		},
		{
			name: "--from empty given",
			args: []string{"slice", "create", "login", "--from="},
			want: createInput{name: "login", from: "", fromGiven: true},
		},
		{
			name: "name not checked here",
			args: []string{"slice", "create", "Bad Name!", "--title", "a\tb"},
			want: createInput{name: "Bad Name!", title: "a\tb", from: "main"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found := readLine(t, tt.args...)
			if found.words != "slice create" {
				t.Fatalf("checkLine found %q, want %q", found.words, "slice create")
			}
			if got := newCreateInput(found.call); got != tt.want {
				t.Errorf("newCreateInput got %+v, want %+v", got, tt.want)
			}
		})
	}
}

// TestStatusAndList checks that status gets the given name, and list
// none.
func TestStatusAndList(t *testing.T) {
	found := readLine(t, "slice", "status", "Login")
	if found.words != "slice status" || found.call.name != "Login" {
		t.Errorf("checkLine found %q with name %q, want %q with %q", found.words, found.call.name, "slice status", "Login")
	}

	found = readLine(t, "slice", "list")
	if found.words != "slice list" || found.call.name != "" {
		t.Errorf("checkLine found %q with name %q, want %q with none", found.words, found.call.name, "slice list")
	}
}

// wrongName is an error of another package that is wrong use, as the
// create flow's errors for R20 and R29 will be.
type wrongName struct{}

func (wrongName) Error() string {
	return "Error: invalid slice name \"x y\". Nothing was changed.\nA second line."
}
func (wrongName) WrongUse() bool { return true }

// notWrong has a WrongUse method that returns false.
type notWrong struct{}

func (notWrong) Error() string  { return "Error: it failed." }
func (notWrong) WrongUse() bool { return false }

// TestReport checks R76 for errors from a command's work: wrong use
// gives 2, also when wrapped, and any other error 1. The error's text is
// printed as it is, with a line break.
func TestReport(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode int
		wantText string
	}{
		{"no error", nil, 0, ""},
		{"wrong use", wrongName{}, 2, "Error: invalid slice name \"x y\". Nothing was changed.\nA second line.\n"},
		{"wrapped wrong use", fmt.Errorf("wrapped: %w", wrongName{}), 2, "wrapped: Error: invalid slice name \"x y\". Nothing was changed.\nA second line.\n"},
		{"WrongUse false", notWrong{}, 1, "Error: it failed.\n"},
		{"plain error", errors.New("Error: git failed."), 1, "Error: git failed.\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stderr bytes.Buffer
			if code := report(&stderr, tt.err); code != tt.wantCode {
				t.Errorf("exit code = %d, want %d", code, tt.wantCode)
			}
			if stderr.String() != tt.wantText {
				t.Errorf("stderr = %q, want %q", stderr.String(), tt.wantText)
			}
		})
	}
}

// TestPlaceholders checks that each command whose work is not built yet
// fails with exit code 1. TEMPORARY: remove a case when its step is
// built.
func TestPlaceholders(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{[]string{"slice", "create", "x"}, "Error: sliceflow slice create is not built yet.\n"},
		{[]string{"slice", "status", "x"}, "Error: sliceflow slice status is not built yet.\n"},
		{[]string{"slice", "list"}, "Error: sliceflow slice list is not built yet.\n"},
	}
	for _, tt := range tests {
		var stdout, stderr bytes.Buffer
		code := Run(context.Background(), tt.args, &stdout, &stderr)
		if code != 1 || stderr.String() != tt.want || stdout.Len() != 0 {
			t.Errorf("Run %q: code %d, stdout %q, stderr %q; want 1, nothing, %q",
				tt.args, code, stdout.String(), stderr.String(), tt.want)
		}
	}
}
