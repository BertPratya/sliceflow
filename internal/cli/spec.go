package cli

import (
	"fmt"

	"github.com/BertPratya/sliceflow/internal/slice"
)

// specTemplateUsage is the help text of "sliceflow spec template".
const specTemplateUsage = `Print the built-in spec template, the one "slice create" writes, with {{slice}} and {{date}} left as they are.

Usage:
  sliceflow spec template [flags]

Flags:
  -h, --help     Show this help
`

// runSpecTemplate runs "sliceflow spec template" (0002 R28). args are the
// words after "template". It needs no git repo, so it never asks for the
// current folder.
func runSpecTemplate(e *env, args []string) int {
	f, positional, err := parseFlags(args, flagSpec{})
	if f.help {
		fmt.Fprint(e.stdout, specTemplateUsage)
		return exitOK
	}
	if err != nil {
		return wrongUse(e, err.Error())
	}
	if len(positional) > 0 {
		return wrongUse(e, "too many arguments")
	}

	text, err := slice.SpecTemplate()
	if err != nil {
		fmt.Fprintf(e.stderr, "Error: %v\n", err)
		return exitError
	}
	fmt.Fprint(e.stdout, text)
	return exitOK
}
