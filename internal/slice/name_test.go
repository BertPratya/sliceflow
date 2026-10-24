package slice_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/BertPratya/sliceflow/internal/slice"
)

func TestCheckName(t *testing.T) {
	bad := func(n string) string {
		return fmt.Sprintf(`invalid slice name "%s": use 1-64 lowercase letters, digits, and dashes, starting with a letter or digit. Nothing was changed.`, n)
	}
	reserved := func(n string) string {
		return fmt.Sprintf(`invalid slice name "%s": it is a reserved name on Windows. Nothing was changed.`, n)
	}
	a64 := strings.Repeat("a", 64)
	a65 := strings.Repeat("a", 65)
	tests := []struct {
		name string
		want string // empty = passes
	}{
		{"spec-submit", ""},
		{"a", ""},
		{"0001-x", ""},
		{a64, ""},
		{"com10", ""},
		{"My Slice", bad("My Slice")},
		{"-a", bad("-a")},
		{"a_b", bad("a_b")},
		{"a/b", bad("a/b")},
		{a65, bad(a65)},
		{"", bad("")},
		{"con", reserved("con")},
		{"nul", reserved("nul")},
		{"prn", reserved("prn")},
		{"aux", reserved("aux")},
		{"com1", reserved("com1")},
		{"lpt9", reserved("lpt9")},
	}
	for _, tt := range tests {
		err := slice.CheckName(tt.name)
		if tt.want == "" {
			if err != nil {
				t.Errorf("%q: unexpected error %v", tt.name, err)
			}
			continue
		}
		var se *slice.Error
		if err == nil {
			t.Errorf("%q: want error", tt.name)
			continue
		}
		if e, ok := err.(*slice.Error); ok {
			se = e
		} else {
			t.Errorf("%q: error type %T", tt.name, err)
			continue
		}
		if se.Message != tt.want {
			t.Errorf("%q: got %q, want %q", tt.name, se.Message, tt.want)
		}
	}
}
