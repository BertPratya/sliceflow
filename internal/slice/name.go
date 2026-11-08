package slice

import (
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"
)

// The three problems of R20, in its order (SD28). The text of each error
// is the second line that the command prints.
var (
	errTooLong    = errors.New("A name has at most 40 characters.")
	errReserved   = errors.New("Windows reserves this name. Choose another.")
	errCharacters = errors.New("Allowed characters: letters, digits, hyphens, underscores. Start with a letter or a digit.")
)

// reservedNames are the names Windows does not allow for a file or a
// folder (R19).
var reservedNames = map[string]struct{}{
	"con": {}, "prn": {}, "aux": {}, "nul": {},
	"com1": {}, "com2": {}, "com3": {}, "com4": {}, "com5": {},
	"com6": {}, "com7": {}, "com8": {}, "com9": {},
	"lpt1": {}, "lpt2": {}, "lpt3": {}, "lpt4": {}, "lpt5": {},
	"lpt6": {}, "lpt7": {}, "lpt8": {}, "lpt9": {},
}

// isAlphaNum reports whether r is one of a to z, A to Z, or 0 to 9, the
// letters and digits of R19. unicode.IsLetter is not used, because it
// also accepts letters such as é.
func isAlphaNum(r rune) bool {
	return (r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9')
}

// lowercaseAZ makes only the letters A to Z lowercase, and leaves every
// other character as it is (R21). unicode.ToLower is called only inside
// the A to Z check: on its own, it also changes other characters, such
// as the Kelvin sign, which becomes a plain k and would then pass the
// character check.
func lowercaseAZ(given string) string {
	var result strings.Builder
	for _, r := range given {
		if r >= 'A' && r <= 'Z' {
			r = unicode.ToLower(r)
		}
		result.WriteRune(r)
	}
	return result.String()
}

// ParseName turns the name the user gave into the slice name, or returns
// the first problem in R20's order (SD28). On an error, the name is "".
//
// Steps:
//  1. If the name has more than 40 characters, return errTooLong. The
//     length counts characters, not bytes: the characters are not
//     checked yet, and é is 2 bytes.
//  2. Make A to Z lowercase (R21).
//  3. If the lowercase name is reserved, return errReserved.
//  4. If the name is empty, does not start with a letter or a digit, or
//     has a character that is not a letter, a digit, a hyphen, or an
//     underscore, return errCharacters.
//  5. Return the lowercase name.
func ParseName(given string) (string, error) {
	if utf8.RuneCountInString(given) > 40 {
		return "", errTooLong
	}

	name := lowercaseAZ(given)

	if _, isReserved := reservedNames[name]; isReserved {
		return "", errReserved
	}

	if name == "" {
		return "", errCharacters
	}
	for i, r := range name {
		allowed := isAlphaNum(r) || (i > 0 && (r == '-' || r == '_'))
		if !allowed {
			return "", errCharacters
		}
	}

	return name, nil
}
