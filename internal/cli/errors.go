package cli

import (
	"errors"
	"fmt"
)

// usageError is a wrong command line: an unknown command or option, a
// missing argument, an option with no value, or an extra word. Run
// prints it as "Error: <message>. Nothing was changed.", then the hint,
// and exits with code 2.
type usageError struct {
	message string // e.g. unknown option "--titel"
	hint    string // e.g. To see the options, run: sliceflow slice create --help
}

func (e *usageError) Error() string {
	return e.message
}

// errHelp is returned by parseWords when the words ask for help.
var errHelp = errors.New("help requested")

// unknownCommandError returns the error for words that name no command:
// the first word, or a group and the word after it.
func unknownCommandError(words string) *usageError {
	return &usageError{
		message: fmt.Sprintf(unknownCommandMessage, words),
		hint:    seeCommandsHint,
	}
}

// unknownOptionError returns the error for an option the command does not
// have. command is the command the user ran, e.g. "slice create", or ""
// for none; the hint points to its help.
func unknownOptionError(option, command string) *usageError {
	hint := seeAllOptionsHint
	if command != "" {
		hint = fmt.Sprintf(seeOptionsHint, command)
	}
	return &usageError{message: fmt.Sprintf(unknownOptionMessage, option), hint: hint}
}

// missingValueError returns the error for option, which needs a value but
// is the last word. The hint points to the help of command.
func missingValueError(option, command string) *usageError {
	return &usageError{
		message: fmt.Sprintf(missingValueMessage, option),
		hint:    fmt.Sprintf(seeOptionsHint, command),
	}
}

// unexpectedArgumentError returns the error for word, a word that comes
// after all the command's arguments. The hint is the usage line.
func unexpectedArgumentError(word, usage string) *usageError {
	return &usageError{
		message: fmt.Sprintf(unexpectedArgumentMessage, word),
		hint:    fmt.Sprintf(usageHint, usage),
	}
}

// missingArgumentError returns the error for the required argument name
// of a command in groupName, as in "missing slice name". The hint is the
// usage line.
func missingArgumentError(groupName, name, usage string) *usageError {
	return &usageError{
		message: fmt.Sprintf(missingArgumentMessage, groupName, name),
		hint:    fmt.Sprintf(usageHint, usage),
	}
}
