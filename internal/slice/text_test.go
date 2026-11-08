package slice_test

import (
	"testing"

	"github.com/BertPratya/sliceflow/internal/slice"
)

// Test 1: HasControlCharacter says yes for the codes 0 to 31, 127, and
// 128 to 159 (R29), and no for every other character. The cases sit on
// the edges of each range.
func TestHasControlCharacter(t *testing.T) {
	tests := []struct {
		name string
		text string
		want bool
	}{
		{"empty", "", false},
		{"normal text", "Pay with a card", false},
		{"code 0", "\x00", true},
		{"code 31, the last of 0 to 31", "\x1f", true},
		{"code 32, a space", " ", false},
		{"code 126", "~", false},
		{"code 127", "\x7f", true},
		{"code 128", "\u0080", true},
		{"code 159", "\u009f", true},
		{"code 160, just past the range", "\u00a0", false},
		{"a letter outside a to z", "café", false},
		{"Thai letters", "ก", false},
		{"a line break at the start", "\nPay", true},
		{"a line break at the end", "Pay\n", true},
		{"a tab", "Pay\twith", true},
		// Not valid UTF-8: range gives U+FFFD, which is not a control
		// character. Left as it is on purpose (choice 1, 2026-11-08).
		{"a broken byte", "\x9b", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := slice.HasControlCharacter(tt.text); got != tt.want {
				t.Errorf("HasControlCharacter(%q) = %v, want %v", tt.text, got, tt.want)
			}
		})
	}
}

// Test 2: EscapeControlCharacters writes each control character as the
// code of R29, doubles each backslash, and leaves every other character
// as it is. Both texts are Go string literals: the quotes are not part
// of the value.
func TestEscapeControlCharacters(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{"empty", "", ""},
		{"normal text", "hello", "hello"},
		{"a line break", "a\nb", `a\nb`},
		{"a carriage return", "a\rb", `a\rb`},
		{"a tab", "a\tb", `a\tb`},
		{"a backslash", `a\b`, `a\\b`},
		{"the bell, code 7", "a\x07b", `a\u0007b`},
		{"escape, code 27", "a\x1bb", `a\u001bb`},
		{"code 0", "\x00", `\u0000`},
		{"code 31", "\x1f", `\u001f`},
		{"code 32, a space", " ", " "},
		{"code 126", "~", "~"},
		{"code 127", "\x7f", `\u007f`},
		{"code 128", "\u0080", `\u0080`},
		{"code 159", "\u009f", `\u009f`},
		{"code 160, just past the range", "\u00a0", "\u00a0"},
		{"quotes stay as they are", `say "hi"`, `say "hi"`},
		{"Thai letters", "ก", "ก"},
		{"the example of S20", "Pay with a card\n", `Pay with a card\n`},
		{"a backslash and a tab", "C:\\temp\t", `C:\\temp\t`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := slice.EscapeControlCharacters(tt.text); got != tt.want {
				t.Errorf("EscapeControlCharacters(%q) = %q, want %q", tt.text, got, tt.want)
			}
		})
	}
}
