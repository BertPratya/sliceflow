package cli

import (
	"slices"
	"strings"
)

// line is what checkLine found on the command line.
type line struct {
	// words are the command words found so far: "" for `sliceflow`,
	// "slice" for `sliceflow slice`, or a command's words.
	words string
	// cmd is the command that words name, or nil when words name no
	// command yet (`sliceflow`, `sliceflow slice`).
	cmd       *command
	call      call // the name and the options' values, for cmd
	wantsHelp bool // --help or -h was given
}

// helpText returns the help for what the words name: the help of the
// command, or the guide when no command was named (R67, R68, R69).
func (l line) helpText() string {
	if l.cmd == nil {
		return guideHelp
	}
	return l.cmd.help
}

// checkLine reads the words of the command line, args, from the left,
// and finds the command in commands, its name and its options' values,
// whether help was asked for, and the first problem, as R40 and R71
// ask. It changes nothing.
//
// Steps, for each word from the left:
//  1. `--help` or `-h`: help was asked for (R71). It is the help of the
//     last command named before or after it.
//  2. After an unknown command, the other words are only searched for
//     --help.
//  3. A word that starts with `-` is an option (R71). An option that
//     does not exist is a problem (R72). An option that needs a value
//     takes the next word as its value, whatever it is; with no next
//     word, it is a problem (R74). `--title=<text>` is the same as
//     `--title <text>`. The last value of an option counts.
//  4. For `sliceflow` and `sliceflow slice`, any other word must be a
//     command word of that level, or it is a problem (R70).
//  5. For `slice create` and `slice status`, the first such word is the
//     name. For all three commands, any other such word is a problem
//     (R75).
//
// Only the first problem is kept. After the last word, when there is no
// problem yet, a command that takes a name and got none, or got "", has
// the problem of R73: a missing name is checked last.
func checkLine(commands []command, args []string) (found line, problem error) {
	found.call.options = map[string]string{}
	lost := false // an unknown command was found
	gotName := false

	keep := func(err error) {
		if problem == nil {
			problem = err
		}
	}

	for i := 0; i < len(args); i++ {
		word := args[i]
		switch {
		case word == "--help" || word == "-h":
			found.wantsHelp = true

		case lost:
			// Only --help counts after an unknown command.

		case strings.HasPrefix(word, "-"):
			name, value, hasValue, err := checkOption(found, word)
			if err != nil {
				keep(err)
				continue
			}
			if !hasValue {
				if i+1 == len(args) {
					keep(needsValue(found.words, word))
					continue
				}
				i++ // the next word is the value
				value = args[i]
			}
			found.call.options[name] = value

		case found.cmd == nil:
			words, cmd, ok := nextWord(commands, found.words, word)
			if !ok {
				keep(unknownCommand(found.words, word))
				lost = true
				continue
			}
			found.words, found.cmd = words, cmd

		case found.cmd.takesName && !gotName:
			gotName = true
			found.call.name = word

		default:
			keep(unexpectedArgument(found.cmd, word))
		}
	}

	if problem == nil && found.cmd != nil && found.cmd.takesName && found.call.name == "" {
		problem = missingName(found.cmd)
	}
	return found, problem
}

// checkOption checks one option, word, of the command found so far. It
// returns the option's name without "--", and its value when the word
// has one, as in `--title=<text>`; otherwise hasValue is false and the
// value is the next word. It returns the problem of R72 when the option
// does not exist.
//
// `--help` and `-h` are checked before, in checkLine. sliceflow has no
// other short options, so any other word with one `-` is unknown, e.g.
// `-`, `-x`, and `-hx`. `--help=<text>` is unknown too: --help takes no
// value. `sliceflow` and `sliceflow slice` have no options.
func checkOption(found line, word string) (name, value string, hasValue bool, err error) {
	if !strings.HasPrefix(word, "--") {
		return "", "", false, unknownOption(found.words, word)
	}
	name, value, hasValue = strings.Cut(word[2:], "=")
	if name == "help" {
		return "", "", false, unknownOption(found.words, word)
	}
	if found.cmd == nil || !slices.Contains(found.cmd.options, name) {
		return "", "", false, unknownOption(found.words, "--"+name)
	}
	return name, value, hasValue, nil
}

// nextWord finds word as the next command word after words, e.g. "create"
// after "slice". It returns the new words, and the command they name, or
// nil when they name only the start of commands, like "slice". ok is
// false when no command starts with them.
//
// A word with a space is never a command word, so that one word
// "slice create" is not taken for two.
func nextWord(commands []command, words, word string) (next string, cmd *command, ok bool) {
	if word == "" || strings.Contains(word, " ") {
		return "", nil, false
	}
	next = word
	if words != "" {
		next = words + " " + word
	}
	for i := range commands {
		if commands[i].words == next {
			return next, &commands[i], true
		}
		if strings.HasPrefix(commands[i].words, next+" ") {
			return next, nil, true
		}
	}
	return "", nil, false
}
