package cli

import (
	"strings"

	"github.com/spf13/cobra"
)

// checkLine reads the words of the command line, args, from the left,
// and finds the command, whether help was asked for, and the first
// problem, as R40 and R71 ask. It changes nothing.
//
// Cobra reads the values of the options later, but it is not used to
// find the problems, because it does not follow R40: it stops at the
// first wrong option, so it misses a --help after it and a wrong word
// before it; it guesses that an unknown option before a command takes
// the next word as its value; and it lets `-`, `--`, and `-test.v`
// through.
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
//     `--title <text>`.
//  4. For `sliceflow` and `sliceflow slice`, any other word must be a
//     command of that level, or it is a problem (R70).
//  5. For `slice create` and `slice status`, the first such word is the
//     name. For all three commands, any other such word is a problem
//     (R75).
//
// Only the first problem is kept. After the last word, when there is no
// problem yet, a command that takes a name and got none, or got "", has
// the problem of R73: a missing name is checked last.
//
// cmd is the command the words name, e.g. `slice create`, or
// `sliceflow` or `slice` when no command was given.
func checkLine(root *cobra.Command, args []string) (cmd *cobra.Command, help bool, problem error) {
	cmd = root
	lost := false // an unknown command was found
	gotName := false
	name := ""

	keep := func(err error) {
		if problem == nil {
			problem = err
		}
	}

	for i := 0; i < len(args); i++ {
		word := args[i]
		switch {
		case word == "--help" || word == "-h":
			help = true

		case lost:
			// Only --help counts after an unknown command.

		case strings.HasPrefix(word, "-"):
			takesValue, err := checkOption(cmd, word)
			if err != nil {
				keep(err)
				continue
			}
			if takesValue {
				if i+1 < len(args) {
					i++ // the next word is the value
				} else {
					keep(needsValue(cmd, word))
				}
			}

		case cmd.HasSubCommands():
			sub := subcommand(cmd, word)
			if sub == nil {
				keep(unknownCommand(cmd, word))
				lost = true
				continue
			}
			cmd = sub

		case takesName(cmd) && !gotName:
			gotName = true
			name = word

		default:
			keep(unexpectedArgument(cmd, word))
		}
	}

	if problem == nil && !cmd.HasSubCommands() && takesName(cmd) && name == "" {
		problem = missingName(cmd)
	}
	return cmd, help, problem
}

// checkOption checks one option, word, of cmd. It returns whether the
// option takes the next word as its value, or the problem of R72 when
// the option does not exist.
//
// `--help` and `-h` are checked before, in checkLine. sliceflow has no
// other short options, so any other word with one `-` is unknown, e.g.
// `-`, `-x`, and `-hx`. `--help=<text>` is unknown too: --help takes no
// value.
func checkOption(cmd *cobra.Command, word string) (takesValue bool, err error) {
	if !strings.HasPrefix(word, "--") {
		return false, unknownOption(cmd, word)
	}

	name, _, hasValue := strings.Cut(word[2:], "=")
	if name == "help" {
		return false, unknownOption(cmd, word)
	}
	option := cmd.Flags().Lookup(name)
	if name == "" || option == nil {
		return false, unknownOption(cmd, "--"+name)
	}
	// NoOptDefVal is "" for an option that needs a value.
	return option.NoOptDefVal == "" && !hasValue, nil
}

// subcommand returns the command of cmd with the name word, or nil.
// Hidden commands, such as the one in place of Cobra's help command,
// are not counted.
func subcommand(cmd *cobra.Command, word string) *cobra.Command {
	for _, sub := range cmd.Commands() {
		if !sub.Hidden && sub.Name() == word {
			return sub
		}
	}
	return nil
}
