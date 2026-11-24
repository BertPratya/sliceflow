package cli

import (
	"errors"
	"github.com/BertPratya/sliceflow/internal/cli/command"
	"slices"
	"strings"
)

// commandLine is what parseCommandLine read from a command line: either
// help to show, or a command to run with its input.
type commandLine struct {
	help   string         // the help text to show, or "" for none
	action command.Action // the command to run, when help is ""
	input  command.Input  // what was parsed for the command
}

// parseCommandLine reads args, the words after the program's name,
// against registry. It returns the help to show when the words ask for
// it: the guide, a group's help, or a command's help. Help comes before
// any error. Otherwise it returns the command to run and its input, or a
// *usageError for a wrong command line.
func parseCommandLine(registry CommandRegistry, args []string) (commandLine, error) {
	if len(args) == 0 {
		return commandLine{help: guideText(registry)}, nil
	}
	if isOption(args[0]) {
		if asksForHelp(args) {
			return commandLine{help: guideText(registry)}, nil
		}
		return commandLine{}, unknownOptionError(optionWord(args[0]), "")
	}

	f, err := resolveCommand(registry, args)
	if err != nil {
		return commandLine{}, err
	}
	if f.actionName == "" {
		if len(f.rest) == 0 || asksForHelp(f.rest) {
			return commandLine{help: groupHelpText(f.groupName, f.group)}, nil
		}
		// resolveCommand stops at the first option, so rest[0] is one.
		return commandLine{}, unknownOptionError(optionWord(f.rest[0]), f.groupName)
	}

	in, err := parseWords(f)
	if errors.Is(err, errHelp) {
		return commandLine{help: commandHelpText(f.groupName, f.actionName, f.action)}, nil
	}
	if err != nil {
		return commandLine{}, err
	}
	return commandLine{action: f.action, input: in}, nil
}

// parseWords reads f.rest, the words after the command words, against
// the arguments and options of f.action in the command table:
//   - a word that starts with "-" is an option; -h means --help;
//   - options may come before or after the arguments;
//   - --name=value means the same as --name value, and the word after an
//     option that needs a value is its value, even when it starts with
//     "-";
//   - when an option is given more than once, the last one counts.
//
// It returns errHelp when the words ask for help, which comes before any
// error. Otherwise it returns a *usageError for the first wrong word from
// the left: an unknown option, an option with no value, or an unexpected
// word. A missing argument is checked last; an empty word counts as
// missing.
func parseWords(f found) (command.Input, error) {
	ran := f.groupName + " " + f.actionName
	usage := usageLine(f.groupName, f.actionName, f.action)
	words := f.rest

	in := command.Input{Args: map[string]string{}, Options: map[string]string{}}
	var firstErr error
	keep := func(err error) {
		if firstErr == nil {
			firstErr = err
		}
	}
	help := false
	nextArg := 0

	for i := 0; i < len(words); i++ {
		word := words[i]

		if !isOption(word) {
			if nextArg < len(f.action.Args) {
				in.Args[f.action.Args[nextArg].Name] = word
				nextArg++
			} else {
				keep(unexpectedArgumentError(word, usage))
			}
			continue
		}

		if isHelp(word) {
			help = true
			continue
		}

		name, value, hasValue := strings.Cut(strings.TrimPrefix(word, "--"), "=")
		option, known := f.action.Options[name]
		if !strings.HasPrefix(word, "--") || !known || option.ValueName == "" {
			keep(unknownOptionError(optionWord(word), ran))
			continue
		}
		if !hasValue {
			if i+1 == len(words) {
				keep(missingValueError(word, ran))
				continue
			}
			i++
			value = words[i]
		}
		in.Options[name] = value
	}

	if help {
		return command.Input{}, errHelp
	}
	if firstErr != nil {
		return command.Input{}, firstErr
	}
	for _, arg := range f.action.Args {
		if arg.Required && in.Args[arg.Name] == "" {
			return command.Input{}, missingArgumentError(f.groupName, arg.Name, usage)
		}
	}
	return in, nil
}

// optionWord returns an option word without its "=value", if it has one:
// --titel=x gives --titel.
func optionWord(word string) string {
	option, _, _ := strings.Cut(word, "=")
	return option
}

// isHelp reports whether word asks for help: -h or --help.
func isHelp(word string) bool {
	return word == "-h" || word == "--help"
}

// asksForHelp reports whether words has -h or --help in it. It is for the
// guide and the help of a group, where no option takes a value. A
// command's words are read by parseWords instead.
func asksForHelp(words []string) bool {
	return slices.ContainsFunc(words, isHelp)
}
