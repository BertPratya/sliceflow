package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/BertPratya/sliceflow/internal/slice"
)

const validateUsage = `Usage: sliceflow spec validate [<name>] [--strict]

Checks docs/slices/<name>/spec.md in the slice's worktree, and prints one
line per problem and a summary. Without a name, it checks the slice of the
current branch (slice/<name>). It writes nothing.

Exit code 1 means an error was found; with --strict, a warning too.

Flags:
  --strict   a warning also gives exit code 1
`

func init() {
	register(command{Name: "spec validate", Usage: validateUsage, Run: runValidate})
}

func runValidate(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("spec validate", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	strict := fs.Bool("strict", false, "")
	pos, err := parseFlags(fs, validateUsage, args, stdout)
	if errors.Is(err, errHelp) {
		return 0
	}
	if err != nil {
		return usageError(stderr, err.Error())
	}
	if len(pos) > 1 {
		return usageError(stderr, "too many arguments")
	}
	name := ""
	if len(pos) == 1 {
		name = pos[0]
	}
	wd, err := os.Getwd()
	if err != nil {
		return fail(stderr, err)
	}
	r, err := slice.FindRepo(ctx, wd)
	if err != nil {
		return statusFail(stderr, err)
	}
	v, err := slice.Validate(ctx, r, name, *strict)
	if err != nil {
		return statusFail(stderr, err)
	}
	printProblems(stdout, v.Problems)
	fmt.Fprintf(stdout, "%s, %s, %s\n", plural(v.Errors, "error"), plural(v.Warnings, "warning"), plural(v.Notes, "note"))
	if v.Errors > 0 || (*strict && v.Warnings > 0) {
		return 1
	}
	return 0
}
