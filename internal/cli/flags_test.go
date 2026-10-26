package cli

import (
	"maps"
	"slices"
	"testing"
)

func TestParseFlags(t *testing.T) {
	spec := flagSpec{values: []string{"from", "title"}, bools: []string{"json"}}

	tests := []struct {
		name       string
		args       []string
		values     map[string]string
		bools      map[string]bool
		positional []string
		help       bool
		err        string
	}{
		{name: "nothing", args: nil},
		{name: "positional only", args: []string{"a", "b"}, positional: []string{"a", "b"}},
		{name: "value with space", args: []string{"--from", "dev", "x"},
			values: map[string]string{"from": "dev"}, positional: []string{"x"}},
		{name: "value with equals", args: []string{"--from=dev", "x"},
			values: map[string]string{"from": "dev"}, positional: []string{"x"}},
		{name: "equals keeps later equals", args: []string{"--title=a=b"},
			values: map[string]string{"title": "a=b"}},
		{name: "empty value with equals", args: []string{"--title="},
			values: map[string]string{"title": ""}},
		{name: "bool flag", args: []string{"--json", "x"},
			bools: map[string]bool{"json": true}, positional: []string{"x"}},
		{name: "flags after positional", args: []string{"x", "--json", "--from", "dev"},
			values: map[string]string{"from": "dev"}, bools: map[string]bool{"json": true},
			positional: []string{"x"}},
		{name: "-h", args: []string{"-h"}, help: true},
		{name: "--help after words", args: []string{"x", "--help"}, positional: []string{"x"}, help: true},
		{name: "double dash ends flags", args: []string{"--", "--json", "-h"},
			positional: []string{"--json", "-h"}},
		{name: "flag before double dash", args: []string{"--json", "--", "--from"},
			bools: map[string]bool{"json": true}, positional: []string{"--from"}},
		{name: "unknown long flag", args: []string{"x", "--colour"}, positional: []string{"x"},
			err: "unknown flag --colour"},
		{name: "unknown flag with equals", args: []string{"--x=v"}, err: "unknown flag --x"},
		{name: "unknown short flag", args: []string{"-v"}, err: "unknown flag -v"},
		{name: "value flag at the end", args: []string{"x", "--from"},
			positional: []string{"x"}, err: "flag needs a value: --from"},
		{name: "bool flag with a value", args: []string{"--json=x"}, err: "unknown flag --json"},
		{name: "lone dash is positional", args: []string{"-"}, positional: []string{"-"}},
		{name: "empty positional", args: []string{""}, positional: []string{""}},
		{name: "repeated value flag, last wins", args: []string{"--from", "a", "--from=b"},
			values: map[string]string{"from": "b"}},
		{name: "value that starts with a dash", args: []string{"--from", "--json"},
			err: "flag needs a value: --from", bools: map[string]bool{"json": true}},
		{name: "equals form allows a dash value", args: []string{"--from=-x"},
			values: map[string]string{"from": "-x"}},
		{name: "lone dash as a value", args: []string{"--from", "-"},
			values: map[string]string{"from": "-"}},
		{name: "help wins after an error", args: []string{"--colour", "--help"}, help: true,
			err: "unknown flag --colour"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, positional, err := parseFlags(tt.args, spec)

			gotErr := ""
			if err != nil {
				gotErr = err.Error()
			}
			if gotErr != tt.err {
				t.Errorf("error = %q, want %q", gotErr, tt.err)
			}
			if f.help != tt.help {
				t.Errorf("help = %v, want %v", f.help, tt.help)
			}
			if !maps.Equal(f.values, tt.values) {
				t.Errorf("values = %v, want %v", f.values, tt.values)
			}
			if !maps.Equal(f.bools, tt.bools) {
				t.Errorf("bools = %v, want %v", f.bools, tt.bools)
			}
			if !slices.Equal(positional, tt.positional) {
				t.Errorf("positional = %v, want %v", positional, tt.positional)
			}
		})
	}
}
