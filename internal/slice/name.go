package slice

import (
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"
)

// The three problems a name can have, in the order they are checked: when
// a name has more than one, the first is reported.
var (
	ErrNameTooLong    = errors.New("name has more than 40 characters")
	ErrNameReserved   = errors.New("name is reserved by Windows")
	ErrNameCharacters = errors.New("name has a character that is not allowed")
)

// reservedNames are the names Windows does not allow for a file or a
// folder.
var reservedNames = map[string]struct{}{
	"con": {}, "prn": {}, "aux": {}, "nul": {},
	"com1": {}, "com2": {}, "com3": {}, "com4": {}, "com5": {},
	"com6": {}, "com7": {}, "com8": {}, "com9": {},
	"lpt1": {}, "lpt2": {}, "lpt3": {}, "lpt4": {}, "lpt5": {},
	"lpt6": {}, "lpt7": {}, "lpt8": {}, "lpt9": {},
}

// isAlphaNum reports whether r is one of a to z, A to Z, or 0 to 9: the
// letters and digits a name may have. unicode.IsLetter is not used,
// because it also accepts letters such as é.
func isAlphaNum(r rune) bool {
	return (r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9')
}

// lowercaseAZ makes only the letters A to Z lowercase, and leaves every
// other character as it is. unicode.ToLower is called only inside
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

// ParseName turns the name the user gave into the slice name: the same
// name, with A to Z made lowercase. A slice name has at most 40
// characters, counted as characters, not bytes; is not a name that
// Windows reserves, such as con or lpt1; and has only letters a to z,
// digits, hyphens, and underscores, starting with a letter or a digit.
// When the name breaks more than one rule, the error is the first of
// ErrNameTooLong, ErrNameReserved, and ErrNameCharacters. On an error,
// the name is "".
func ParseName(given string) (string, error) {
	if utf8.RuneCountInString(given) > 40 {
		return "", ErrNameTooLong
	}

	name := lowercaseAZ(given)

	if _, isReserved := reservedNames[name]; isReserved {
		return "", ErrNameReserved
	}

	if name == "" {
		return "", ErrNameCharacters
	}
	for i, r := range name {
		allowed := isAlphaNum(r) || (i > 0 && (r == '-' || r == '_'))
		if !allowed {
			return "", ErrNameCharacters
		}
	}

	return name, nil
}
