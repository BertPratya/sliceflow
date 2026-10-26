package cli

import (
	"fmt"
	"time"

	"github.com/BertPratya/sliceflow/internal/slice"
)

// sliceCreateUsage is the help text of "sliceflow slice create".
const sliceCreateUsage = `Create a new slice: a branch, a worktree, a spec to fill in, and an event log.

Usage:
  sliceflow slice create <name> [flags]

Flags:
  --from <branch>    The branch to start from (default: git config sliceflow.base, else main)
  --title <text>     The slice's title
  --json             Print the result as JSON
  -h, --help         Show this help
`

// sliceCreateFlags is the flags that "slice create" accepts.
var sliceCreateFlags = flagSpec{values: []string{"from", "title"}, bools: []string{"json"}}

// runSliceCreate runs "sliceflow slice create". args are the words after
// "create". It checks the words, then asks the slice package to create the slice.
func runSliceCreate(e *env, args []string) int {
	f, positional, err := parseFlags(args, sliceCreateFlags)
	if f.help {
		fmt.Fprint(e.stdout, sliceCreateUsage)
		return exitOK
	}
	if err != nil {
		return wrongUse(e, err.Error())
	}
	if len(positional) == 0 || positional[0] == "" {
		return wrongUse(e, "missing slice name")
	}
	if len(positional) > 1 {
		return wrongUse(e, "too many arguments")
	}

	dir, err := e.workDir()
	if err != nil {
		fmt.Fprintf(e.stderr, "Error: %v\n", err)
		return exitError
	}

	_, err = slice.Create(slice.CreateOptions{
		Dir:   dir,
		Name:  positional[0],
		From:  f.values["from"],
		Title: f.values["title"],
		Now:   time.Now(),
	})
	if err != nil {
		fmt.Fprintf(e.stderr, "Error: %v\n", err)
		return exitError
	}

	// TODO(chunk 4): print the result.
	return exitOK
}
