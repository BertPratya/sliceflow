package cli

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// wrongUseError is a command line that is wrong (R70 to R75). Its text
// is the two lines the command prints, with no line break at the end.
// It gives exit code 2 (R76).
type wrongUseError struct {
	text string
}

func (e *wrongUseError) Error() string { return e.text }

// WrongUse reports that the command line is wrong. See wrongUser.
func (e *wrongUseError) WrongUse() bool { return true }

// unknownCommand is the problem of R70: word is not a command of cmd.
// The words in the message are the command path of cmd without
// "sliceflow", then word, e.g. "slice delete".
func unknownCommand(cmd *cobra.Command, word string) error {
	words := word
	if cmd.HasParent() {
		words = cmd.Name() + " " + word
	}
	return &wrongUseError{text: `Error: unknown command "` + words + `". Nothing was changed.` + "\n" +
		"To see the commands, run: sliceflow --help"}
}

// unknownOption is the problem of R72: option does not exist for cmd.
// The second line names cmd, e.g. "sliceflow slice create --help", or
// "sliceflow --help" when no command was given.
func unknownOption(cmd *cobra.Command, option string) error {
	return &wrongUseError{text: `Error: unknown option "` + option + `". Nothing was changed.` + "\n" +
		"To see the options, run: " + cmd.CommandPath() + " --help"}
}

// needsValue is the problem of R74: option of cmd got no value.
func needsValue(cmd *cobra.Command, option string) error {
	return &wrongUseError{text: `Error: option "` + option + `" needs a value. Nothing was changed.` + "\n" +
		"To see the options, run: " + cmd.CommandPath() + " --help"}
}

// missingName is the problem of R73: cmd got no name, or "".
func missingName(cmd *cobra.Command) error {
	return &wrongUseError{text: "Error: missing slice name. Nothing was changed.\n" +
		"Usage: " + usageLine(cmd)}
}

// unexpectedArgument is the problem of R75: word is not a name and not
// an option of cmd.
func unexpectedArgument(cmd *cobra.Command, word string) error {
	return &wrongUseError{text: `Error: unexpected argument "` + word + `". Nothing was changed.` + "\n" +
		"Usage: " + usageLine(cmd)}
}

// flagError turns an error of pflag, which Cobra uses to read the
// options, into the problem of R72 or R74. Cobra calls it when reading
// the options of cmd fails. checkLine stops every wrong command line
// before Cobra runs, so Cobra never calls it today, and no e2e test can
// reach it. It is a second guard, in case Cobra's or pflag's behaviour
// changes: it keeps pflag's own text from being printed. TestFlagError
// calls it directly, with pflag's real errors.
//
// The kind of error is found with errors.As on pflag's error types, not
// from its text:
//   - NotExistError: an unknown option (R72);
//   - ValueRequiredError: an option with no value (R74);
//   - InvalidSyntaxError, e.g. `---x`: an unknown option (R72);
//   - InvalidValueError, e.g. `--help=x`: an unknown option (R72).
//
// Any other error is still wrong use, with pflag's text in the line.
func flagError(cmd *cobra.Command, err error) error {
	var notExist *pflag.NotExistError
	if errors.As(err, &notExist) {
		return unknownOption(cmd, optionWord(notExist.GetSpecifiedName(), notExist.GetSpecifiedShortnames()))
	}

	var noValue *pflag.ValueRequiredError
	if errors.As(err, &noValue) {
		return needsValue(cmd, optionWord(noValue.GetSpecifiedName(), noValue.GetSpecifiedShortnames()))
	}

	var badSyntax *pflag.InvalidSyntaxError
	if errors.As(err, &badSyntax) {
		return unknownOption(cmd, badSyntax.GetSpecifiedFlag())
	}

	var badValue *pflag.InvalidValueError
	if errors.As(err, &badValue) {
		return unknownOption(cmd, "--"+badValue.GetFlag().Name+"="+badValue.GetValue())
	}

	return &wrongUseError{text: "Error: " + err.Error() + ". Nothing was changed.\n" +
		"To see the options, run: " + cmd.CommandPath() + " --help"}
}

// optionWord writes an option as the user typed it, from the parts that
// pflag gives: "--title" for a long option, or "-x" when shorthands, the
// group of short options it was in, is not empty.
func optionWord(name, shorthands string) string {
	if shorthands != "" {
		return "-" + name
	}
	return "--" + name
}
