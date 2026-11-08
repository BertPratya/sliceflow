package slice

import (
	"fmt"
	"strings"
	"unicode"
)

// HasControlCharacter reports whether text holds a control character, as
// R29 defines it: a character with the code 0 to 31, 127, or 128 to 159.
// unicode.IsControl says yes for exactly these codes, and no for every
// other character.
//
// Steps:
//  1. Go through the text one character at a time.
//  2. If a character is a control character, return true.
//  3. After the loop, return false.
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
// written as a visible code, for the "Received:" line of R29. It also
// doubles each backslash, so that a plain backslash and the start of a
// code never look the same. It adds no quotes around the text: the line
// of R29 already has them.
//
// Steps:
//  1. Go through the text one character at a time.
//  2. For each character, write:
//     - a line break as \n, a carriage return as \r, a tab as \t;
//     - a backslash as \\;
//     - any other control character as \u and four hex digits: \u001b;
//     - any other character as it is.
//  3. Return the result.
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
