package cli

// wrongUseError is a command line that is wrong (R70 to R75). Its text
// is the two lines the command prints, with no line break at the end.
// It gives exit code 2 (R76).
type wrongUseError struct {
	text string
}

func (e *wrongUseError) Error() string { return e.text }

// WrongUse reports that the command line is wrong. See wrongUseMarker.
func (e *wrongUseError) WrongUse() bool { return true }

// unknownCommand is the problem of R70: word is not a command word after
// words. The words in the message are words, then word, e.g. "slice
// delete".
func unknownCommand(words, word string) error {
	if words != "" {
		word = words + " " + word
	}
	return &wrongUseError{text: `Error: unknown command "` + word + `". Nothing was changed.` + "\n" +
		"To see the commands, run: sliceflow --help"}
}

// unknownOption is the problem of R72: option does not exist for what
// words name. The second line names it, e.g. "sliceflow slice create
// --help", or "sliceflow --help" when no command was given.
func unknownOption(words, option string) error {
	return &wrongUseError{text: `Error: unknown option "` + option + `". Nothing was changed.` + "\n" +
		"To see the options, run: " + commandPath(words) + " --help"}
}

// needsValue is the problem of R74: option of the command that words
// name got no value.
func needsValue(words, option string) error {
	return &wrongUseError{text: `Error: option "` + option + `" needs a value. Nothing was changed.` + "\n" +
		"To see the options, run: " + commandPath(words) + " --help"}
}

// missingName is the problem of R73: cmd, which words name, got no
// name, or "".
func missingName(words string, cmd *command) error {
	return &wrongUseError{text: "Error: missing slice name. Nothing was changed.\n" +
		"Usage: " + usageLine(words, cmd)}
}

// unexpectedArgument is the problem of R75: word is not a name and not
// an option of cmd, which words name.
func unexpectedArgument(words string, cmd *command, word string) error {
	return &wrongUseError{text: `Error: unexpected argument "` + word + `". Nothing was changed.` + "\n" +
		"Usage: " + usageLine(words, cmd)}
}
