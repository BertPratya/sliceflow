package slice

import "fmt"

// CheckName returns an *Error when name breaks the slice name rule (R2).
func CheckName(name string) error {
	if !validChars(name) {
		return &Error{Message: fmt.Sprintf(`invalid slice name "%s": use 1-64 lowercase letters, digits, and dashes, starting with a letter or digit. Nothing was changed.`, name)}
	}
	if isReserved(name) {
		return &Error{Message: fmt.Sprintf(`invalid slice name "%s": it is a reserved name on Windows. Nothing was changed.`, name)}
	}
	return nil
}

func validChars(name string) bool {
	if len(name) < 1 || len(name) > 64 {
		return false
	}
	for i := 0; i < len(name); i++ {
		c := name[i]
		switch {
		case c >= 'a' && c <= 'z', c >= '0' && c <= '9':
		case c == '-' && i > 0:
		default:
			return false
		}
	}
	return true
}

func isReserved(name string) bool {
	switch name {
	case "con", "prn", "aux", "nul":
		return true
	}
	if len(name) == 4 && (name[:3] == "com" || name[:3] == "lpt") {
		return name[3] >= '1' && name[3] <= '9'
	}
	return false
}
