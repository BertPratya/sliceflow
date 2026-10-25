package slice

import "fmt"

// reservedNames are the names Windows does not allow for a file or folder,
// so a slice cannot use them (R2).
var reservedNames = map[string]struct{}{
	"con": {}, "prn": {}, "aux": {}, "nul": {},
	"com1": {}, "com2": {}, "com3": {}, "com4": {},
	"com5": {}, "com6": {}, "com7": {}, "com8": {}, "com9": {},
	"lpt1": {}, "lpt2": {}, "lpt3": {}, "lpt4": {},
	"lpt5": {}, "lpt6": {}, "lpt7": {}, "lpt8": {}, "lpt9": {},
}

// ValidateName reports whether name can be used as a slice name (R2). A
// valid name has 1 to 64 characters, uses only lowercase letters a-z,
// digits, and dashes, starts with a letter or a digit, and is not a name
// Windows reserves. It returns nil for a valid name, and otherwise an
// error that says what is wrong (R14).
func ValidateName(name string) error {
	if name == "" {
		return invalidShape(name)
	}

	if len(name) > 64 {
		return invalidShape(name)
	}

	if !isLowerLetter(name[0]) && !isDigit(name[0]) {
		return invalidShape(name)
	}

	if _, exists := reservedNames[name]; exists {
		return fmt.Errorf("invalid slice name %q: it is a reserved name on Windows", name)
	}

	for _, ch := range name {
		switch {
		case ch >= 'a' && ch <= 'z':
		case ch >= '0' && ch <= '9':
		case ch == '-':
		default:
			return invalidShape(name)
		}
	}

	return nil
}

// invalidShape is the error for a name that breaks the length or character
// rules of R2 (R14).
func invalidShape(name string) error {
	return fmt.Errorf("invalid slice name %q: use 1-64 lowercase letters, digits, and dashes, starting with a letter or digit", name)
}

func isLowerLetter(ch byte) bool {
	return ch >= 'a' && ch <= 'z'
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}
