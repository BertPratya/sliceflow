package cli

import (
	"errors"
	"io"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// pflagError returns the error that pflag itself gives for args, with
// the options of `slice create`, so the test uses pflag's real error
// types and not copies of them.
func pflagError(t *testing.T, args ...string) error {
	t.Helper()
	fs := pflag.NewFlagSet("create", pflag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.String("title", "", "")
	fs.BoolP("help", "h", false, "")
	err := fs.Parse(args)
	if err == nil {
		t.Fatalf("pflag gave no error for %q", args)
	}
	return err
}

// TestFlagError checks that each kind of pflag error becomes the
// message of R72 or R74, with exit code 2, and none of pflag's text.
func TestFlagError(t *testing.T) {
	cmd := &cobra.Command{Use: "create"}
	(&cobra.Command{Use: "slice"}).AddCommand(cmd)
	const options = "\nTo see the options, run: slice create --help"

	tests := []struct {
		name string
		err  error
		want string
	}{
		{"unknown long option", pflagError(t, "--titel"),
			`Error: unknown option "--titel". Nothing was changed.` + options},
		{"unknown long option with =", pflagError(t, "--titel=x"),
			`Error: unknown option "--titel". Nothing was changed.` + options},
		{"unknown short option", pflagError(t, "-x"),
			`Error: unknown option "-x". Nothing was changed.` + options},
		{"unknown short option in a group", pflagError(t, "-hx"),
			`Error: unknown option "-x". Nothing was changed.` + options},
		{"no value", pflagError(t, "--title"),
			`Error: option "--title" needs a value. Nothing was changed.` + options},
		{"bad syntax", pflagError(t, "---x"),
			`Error: unknown option "---x". Nothing was changed.` + options},
		{"bad value", pflagError(t, "--help=maybe"),
			`Error: unknown option "--help=maybe". Nothing was changed.` + options},
		{"other error", errors.New("something new"),
			"Error: something new. Nothing was changed." + options},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := flagError(cmd, tt.err)
			if got.Error() != tt.want {
				t.Errorf("text = %q, want %q", got.Error(), tt.want)
			}
			if code := exitCode(got); code != 2 {
				t.Errorf("exit code = %d, want 2", code)
			}
		})
	}
}

// TestExitCode checks R76 for nil and for an error that is not wrong
// use.
func TestExitCode(t *testing.T) {
	if code := exitCode(nil); code != 0 {
		t.Errorf("exitCode(nil) = %d, want 0", code)
	}
	if code := exitCode(errors.New("x")); code != 1 {
		t.Errorf("exitCode(plain error) = %d, want 1", code)
	}
}
