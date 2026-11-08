package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"testing"
)

// recorder is a flows that keeps what each command got, and returns err.
type recorder struct {
	create *createInput
	status *statusInput
	list   bool
	err    error
}

func (r *recorder) flows() flows {
	return flows{
		create: func(_ context.Context, in createInput, _, _ io.Writer) error {
			r.create = &in
			return r.err
		},
		status: func(_ context.Context, in statusInput, _, _ io.Writer) error {
			r.status = &in
			return r.err
		},
		list: func(context.Context, io.Writer, io.Writer) error {
			r.list = true
			return r.err
		},
	}
}

// runRecorded runs args with a recorder, and fails the test when the
// exit code is not 0 or something was printed.
func runRecorded(t *testing.T, args ...string) *recorder {
	t.Helper()
	r := &recorder{}
	var stdout, stderr bytes.Buffer
	code := run(context.Background(), args, &stdout, &stderr, r.flows())
	if code != 0 || stdout.Len() != 0 || stderr.Len() != 0 {
		t.Fatalf("run %q: code %d, stdout %q, stderr %q; want 0 and no output",
			args, code, stdout.String(), stderr.String())
	}
	return r
}

// TestCreateOptions checks R71: options before or after the name,
// --option=<text>, the word after an option is its value, and the last
// one counts.
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
			r := runRecorded(t, tt.args...)
			if r.create == nil {
				t.Fatal("create was not called")
			}
			if *r.create != tt.want {
				t.Errorf("create got %+v, want %+v", *r.create, tt.want)
			}
		})
	}
}

// TestStatusAndList checks that status gets the given name, and list is
// called.
func TestStatusAndList(t *testing.T) {
	r := runRecorded(t, "slice", "status", "Login")
	if r.status == nil || r.status.name != "Login" {
		t.Errorf("status got %+v, want name %q", r.status, "Login")
	}

	r = runRecorded(t, "slice", "list")
	if !r.list {
		t.Error("list was not called")
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

// TestFlowErrors checks R76 for errors from a command's work: wrong use
// gives 2, also when wrapped, and any other error 1. The error's text is
// printed as it is, with a line break.
func TestFlowErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode int
		wantText string
	}{
		{"wrong use", wrongName{}, 2, "Error: invalid slice name \"x y\". Nothing was changed.\nA second line.\n"},
		{"wrapped wrong use", fmt.Errorf("wrapped: %w", wrongName{}), 2, "wrapped: Error: invalid slice name \"x y\". Nothing was changed.\nA second line.\n"},
		{"WrongUse false", notWrong{}, 1, "Error: it failed.\n"},
		{"plain error", errors.New("Error: git failed."), 1, "Error: git failed.\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &recorder{err: tt.err}
			var stdout, stderr bytes.Buffer
			code := run(context.Background(), []string{"slice", "create", "x"}, &stdout, &stderr, r.flows())
			if code != tt.wantCode {
				t.Errorf("exit code = %d, want %d", code, tt.wantCode)
			}
			if stderr.String() != tt.wantText {
				t.Errorf("stderr = %q, want %q", stderr.String(), tt.wantText)
			}
			if stdout.Len() != 0 {
				t.Errorf("stdout = %q, want nothing", stdout.String())
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

// TestContextReachesTheWork checks TD5: the context given to Run is the
// one the work gets, so Ctrl+C reaches it.
func TestContextReachesTheWork(t *testing.T) {
	type key struct{}
	ctx := context.WithValue(context.Background(), key{}, "mine")
	var got any
	f := flows{
		list: func(ctx context.Context, _, _ io.Writer) error {
			got = ctx.Value(key{})
			return nil
		},
	}
	var stdout, stderr bytes.Buffer
	if code := run(ctx, []string{"slice", "list"}, &stdout, &stderr, f); code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr %q", code, stderr.String())
	}
	if got != "mine" {
		t.Errorf("the work got a context without the value; got %v", got)
	}
}
