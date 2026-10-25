package slice

import (
	"fmt"
	"strings"
	"testing"
)

func TestValidateName(t *testing.T) {
	// shape and reserved build the two error texts of R14.
	shape := func(name string) string {
		return fmt.Sprintf("invalid slice name %q: use 1-64 lowercase letters, digits, and dashes, starting with a letter or digit", name)
	}
	reserved := func(name string) string {
		return fmt.Sprintf("invalid slice name %q: it is a reserved name on Windows", name)
	}

	tests := []struct {
		name  string
		input string
		want  string // the exact error text, or "" for a valid name
	}{
		// Valid names
		{"simple name", "feature", ""},
		{"starts with digit", "123feature", ""},
		{"digits only", "12345", ""},
		{"with hyphen", "my-feature", ""},
		{"mixed valid characters", "feature-123-test", ""},
		{"single letter", "a", ""},
		{"single digit", "1", ""},
		{"exactly 64 characters", strings.Repeat("a", 64), ""},

		// The texts of S5 and S6, written out in full, so a mistake in
		// shape or reserved above cannot hide one in ValidateName.
		{"spec S5", "My Slice", `invalid slice name "My Slice": use 1-64 lowercase letters, digits, and dashes, starting with a letter or digit`},
		{"spec S6", "con", `invalid slice name "con": it is a reserved name on Windows`},

		// Empty and length validation
		{"empty name", "", shape("")},
		{"exceeds 64 characters", strings.Repeat("a", 65), shape(strings.Repeat("a", 65))},

		// Invalid starting characters
		{"starts with space", " feature", shape(" feature")},
		{"starts with hyphen", "-feature", shape("-feature")},
		{"starts with underscore", "_feature", shape("_feature")},
		{"starts with uppercase", "Feature", shape("Feature")},
		{"starts with special character", "@feature", shape("@feature")},

		// Invalid characters
		{"contains underscore", "my_feature", shape("my_feature")},
		{"contains space", "my feature", shape("my feature")},
		{"contains uppercase", "myFeature", shape("myFeature")},
		{"contains dot", "my.feature", shape("my.feature")},
		{"contains slash", "my/feature", shape("my/feature")},
		{"contains backslash", `my\feature`, shape(`my\feature`)},
		{"contains at sign", "my@feature", shape("my@feature")},
		{"contains hash", "my#feature", shape("my#feature")},
		{"contains exclamation mark", "my!feature", shape("my!feature")},
		{"contains colon", "my:feature", shape("my:feature")},
		{"contains tab", "my\tfeature", shape("my\tfeature")},
		{"contains newline", "my\nfeature", shape("my\nfeature")},
		{"contains unicode", "café", shape("café")},
		{"contains emoji", "feature🚀", shape("feature🚀")},
		{"trailing space", "feature ", shape("feature ")},

		// Cross-platform reserved names
		{"reserved prn", "prn", reserved("prn")},
		{"reserved aux", "aux", reserved("aux")},
		{"reserved nul", "nul", reserved("nul")},
		{"reserved com1", "com1", reserved("com1")},
		{"reserved com9", "com9", reserved("com9")},
		{"reserved lpt1", "lpt1", reserved("lpt1")},
		{"reserved lpt9", "lpt9", reserved("lpt9")},

		// Names resembling reserved names
		{"reserved name with suffix", "con-test", ""},
		{"reserved name with prefix", "my-con", ""},
		{"reserved name with digit suffix", "con1", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ""
			if err := ValidateName(tt.input); err != nil {
				got = err.Error()
			}

			if got != tt.want {
				t.Errorf("ValidateName(%q) error = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
