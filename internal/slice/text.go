package slice

import (
	"fmt"
	"strings"
	"unicode"
)

// HasControlCharacter reports whether text holds a control character: a
// character with the code 0 to 31, 127, or 128 to 159, such as a line
// break or a tab. unicode.IsControl says yes for exactly these codes, and
// no for every other character.
//
// A byte that is not valid UTF-8 is not a character, so it is not
// checked: range gives U+FFFD for it, and U+FFFD is not a control
// character. On Linux and macOS a command-line argument can hold such a
// byte; this was left as it is on purpose (decided, the human,
// 2026-11-08).
func HasControlCharacter(text string) bool {
	for _, r := range text {
		if unicode.IsControl(r) {
			return true
		}
	}
	return false
}

// EscapeControlCharacters returns text with each control character
// written as a visible code, for the "Received:" line of the error about
// a control character. It also doubles each backslash, so that a plain
// backslash and the start of a code never look the same. A line break
// is written \n, a carriage
// return \r, a tab \t, a backslash \\, and any other control character
// \u and four hex digits, e.g. \u001b. Every other character stays as it
// is. It adds no quotes around the text: the line it is printed in
// already has them.
//
// The four cases before unicode.IsControl must stay before it: a switch
// runs the first case that matches, and a line break, a carriage return,
// and a tab are control characters too.
func EscapeControlCharacters(text string) string {
	var result strings.Builder
	for _, r := range text {
		switch {
		case r == '\n':
			result.WriteString(`\n`)
		case r == '\r':
			result.WriteString(`\r`)
		case r == '\t':
			result.WriteString(`\t`)
		case r == '\\':
			result.WriteString(`\\`)
		case unicode.IsControl(r):
			// Fprintf's error is not checked: writing to a
			// strings.Builder never fails.
			fmt.Fprintf(&result, `\u%04x`, r)
		default:
			result.WriteRune(r)
		}
	}
	return result.String()
}
