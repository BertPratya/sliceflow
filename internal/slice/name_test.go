package slice_test

import (
	"strings"
	"testing"

	"github.com/BertPratya/sliceflow/internal/slice"
)

// The second lines of R20. ParseName's error text is the line itself, so
// the command can print it as it is.
const (
	msgCharacters = "Allowed characters: letters, digits, hyphens, underscores. Start with a letter or a digit."
	msgTooLong    = "A name has at most 40 characters."
	msgReserved   = "Windows reserves this name. Choose another."
)

// kelvin is the Kelvin sign, U+212A. It looks like K, but it is not one
// of A to Z. strings.ToLower and unicode.ToLower turn it into a plain k.
const kelvin = "K"

// Test 1: a valid name comes back lowercase: only A to Z change (R21).
func TestParseNameGoodNames(t *testing.T) {
	tests := []struct {
		given string
		want  string
	}{
		{"card-payment", "card-payment"},
		{"Card-Payment", "card-payment"},
		{"ABCDEFGHIJKLMNOPQRSTUVWXYZ", "abcdefghijklmnopqrstuvwxyz"},
		{"R2-D2_v1", "r2-d2_v1"},
		{"a", "a"},
		{"7", "7"},
		{"1st-try", "1st-try"},
		{"my_slice", "my_slice"},
		{"ends-with-hyphen-", "ends-with-hyphen-"},
		{"ends_with_underscore_", "ends_with_underscore_"},
		{strings.Repeat("a", 40), strings.Repeat("a", 40)},
		{strings.Repeat("A", 40), strings.Repeat("a", 40)},
		// Close to a reserved name, but not one.
		{"con1", "con1"},
		{"console", "console"},
		{"com0", "com0"},
		{"com10", "com10"},
		{"lpt0", "lpt0"},
		{"lpt10", "lpt10"},
		{"nul-x", "nul-x"},
		{"aux_", "aux_"},
	}
	for _, tt := range tests {
		t.Run(tt.given, func(t *testing.T) {
			got, err := slice.ParseName(tt.given)
			if err != nil {
				t.Fatalf("ParseName(%q) returned the error %q, want none", tt.given, err)
			}
			if got != tt.want {
				t.Errorf("ParseName(%q) = %q, want %q", tt.given, got, tt.want)
			}
		})
	}
}

// Test 2: a character or a start that R19 does not allow gives the
// first message of R20. A capital outside A to Z is not made lowercase,
// so it is rejected too.
func TestParseNameRejectsCharacters(t *testing.T) {
	names := []string{
		"-card",        // starts with a hyphen
		"_card",        // starts with an underscore
		"card payment", // a space
		"card.payment", // a dot
		"card/payment", // a slash
		`card\payment`, // a backslash
		"café",         // a letter outside a to z
		"CAFÉ",         // É is not one of A to Z, so it stays, and is rejected
		kelvin + "art", // looks like Kart
		kelvin + "ART", // looks like KART
		"การ์ด",        // Thai letters
		"card\n",       // a line break
		"card:payment", // Windows does not allow : in a path
		"card*", "card?", "card\"", "card<", "card>", "card|",
	}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			checkError(t, name, msgCharacters)
		})
	}
}

// Test 3: a name over 40 characters gives the second message of R20.
func TestParseNameRejectsTooLong(t *testing.T) {
	checkError(t, strings.Repeat("a", 41), msgTooLong)
	checkError(t, strings.Repeat("A", 41), msgTooLong)
	checkError(t, strings.Repeat("ab-", 20), msgTooLong) // 60
}

// Test 4: every name that Windows reserves gives the third message of
// R20, also when it is given in capitals, because the name is made
// lowercase first (R21).
func TestParseNameRejectsReserved(t *testing.T) {
	names := []string{"con", "prn", "aux", "nul", "CON", "Nul", "COM1", "Lpt9"}
	for i := 1; i <= 9; i++ {
		names = append(names, "com"+string(rune('0'+i)), "lpt"+string(rune('0'+i)))
	}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			checkError(t, name, msgReserved)
		})
	}
}

// Test 5: when a name has more than one problem, the message is for the
// first one in R20's order: length, then reserved, then characters
// (SD28). The length counts characters, not bytes: a name of exactly 40
// characters with é in it is 41 bytes long, and must be rejected for the
// é, not for its length.
func TestParseNameReportsTheFirstProblem(t *testing.T) {
	checkError(t, "café"+strings.Repeat("a", 40), msgTooLong)
	checkError(t, "-"+strings.Repeat("a", 45), msgTooLong)
	checkError(t, "é"+strings.Repeat("a", 39), msgCharacters)
	checkError(t, strings.Repeat("é", 40), msgCharacters)
	checkError(t, strings.Repeat("é", 41), msgTooLong)
}

// Test 6: an empty name is not valid. The commands report it as a
// missing name before they parse it (R73), so no exact message is
// checked here.
func TestParseNameRejectsEmpty(t *testing.T) {
	if _, err := slice.ParseName(""); err == nil {
		t.Error(`ParseName("") returned no error, want one`)
	}
}

// checkError fails the test unless ParseName(given) returns an error
// whose text is want, and an empty name.
func checkError(t *testing.T, given, want string) {
	t.Helper()
	got, err := slice.ParseName(given)
	if err == nil {
		t.Errorf("ParseName(%q) = %q and no error, want the error %q", given, got, want)
		return
	}
	if err.Error() != want {
		t.Errorf("ParseName(%q) returned the error %q, want %q", given, err, want)
	}
	if got != "" {
		t.Errorf("ParseName(%q) returned the name %q with its error, want \"\"", given, got)
	}
}
