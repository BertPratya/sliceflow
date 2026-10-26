package e2e

import (
	"strings"
	"testing"
)

// TestS26: "slice create --help" prints the usage of "slice create".
func TestS26(t *testing.T) {
	r := run(t, t.TempDir(), "slice", "create", "--help")

	if r.code != 0 {
		t.Errorf("code = %d, want 0", r.code)
	}
	if !strings.Contains(r.stdout, "Usage:\n  sliceflow slice create <name> [flags]") {
		t.Errorf("stdout has no usage of slice create:\n%s", r.stdout)
	}
	if r.stderr != "" {
		t.Errorf("stderr = %q, want empty", r.stderr)
	}
}

// TestHelpShown checks that every way to ask for a group's help prints the
// usage on stdout and nothing on stderr.
func TestHelpShown(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"sliceflow alone", nil, "Usage:\n  sliceflow <command>"},
		{"sliceflow --help", []string{"--help"}, "Usage:\n  sliceflow <command>"},
		{"sliceflow -h", []string{"-h"}, "Usage:\n  sliceflow <command>"},
		{"slice alone", []string{"slice"}, "Usage:\n  sliceflow slice <command>"},
		{"slice --help", []string{"slice", "--help"}, "Usage:\n  sliceflow slice <command>"},
		{"slice create -h", []string{"slice", "create", "-h"}, "Usage:\n  sliceflow slice create"},
		{"unknown flag then -h at the top", []string{"--colour", "-h"}, "Usage:\n  sliceflow <command>"},
		{"unknown flag then -h under slice", []string{"slice", "--colour", "-h"}, "Usage:\n  sliceflow slice <command>"},
		{"help wins over wrong use", []string{"slice", "create", "a", "b", "--colour", "-h"}, "Usage:\n  sliceflow slice create"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := run(t, t.TempDir(), tt.args...)
			if r.code != 0 {
				t.Errorf("code = %d, want 0", r.code)
			}
			if !strings.Contains(r.stdout, tt.want) {
				t.Errorf("stdout does not contain %q:\n%s", tt.want, r.stdout)
			}
			if r.stderr != "" {
				t.Errorf("stderr = %q, want empty", r.stderr)
			}
		})
	}
}

// TestWrongUse checks each wrong-use text of R38 (and S27's text).
func TestWrongUse(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"unknown command", []string{"frobnicate"}, `unknown command "frobnicate"`},
		{"unknown command under slice", []string{"slice", "frobnicate"}, `unknown command "frobnicate"`},
		{"unknown flag at the top", []string{"--colour"}, "unknown flag --colour"},
		{"unknown flag", []string{"slice", "create", "x", "--colour"}, "unknown flag --colour"},
		{"flag needs a value", []string{"slice", "create", "x", "--from"}, "flag needs a value: --from"},
		{"missing slice name", []string{"slice", "create"}, "missing slice name"},
		{"empty slice name", []string{"slice", "create", ""}, "missing slice name"},
		{"too many arguments", []string{"slice", "create", "a", "b"}, "too many arguments"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := run(t, t.TempDir(), tt.args...)
			want := "Error: " + tt.want + ". Run sliceflow --help.\n"
			if r.stderr != want {
				t.Errorf("stderr = %q, want %q", r.stderr, want)
			}
			if r.stdout != "" {
				t.Errorf("stdout = %q, want empty", r.stdout)
			}
			if r.code != 2 {
				t.Errorf("code = %d, want 2", r.code)
			}
		})
	}
}
