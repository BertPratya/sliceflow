package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/BertPratya/sliceflow/internal/slice"
)

const titleUsage = `Usage: sliceflow slice title <name> <title>

Saves a new title for the slice, as a slice.titled event. It works in any
phase.
`

func init() {
	register(command{Name: "slice title", Usage: titleUsage, Run: runTitle})
}

func runTitle(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("slice title", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	pos, err := parseFlags(fs, titleUsage, args, stdout)
	if errors.Is(err, errHelp) {
		return 0
	}
	if err != nil {
		return usageError(stderr, err.Error())
	}
	switch {
	case len(pos) == 0:
		return usageError(stderr, "missing slice name")
	case len(pos) == 1:
		return usageError(stderr, "missing title")
	case len(pos) > 2:
		return usageError(stderr, "too many arguments")
	}
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	wd, err := os.Getwd()
	if err != nil {
		return fail(stderr, err)
	}
	r, err := slice.FindRepo(ctx, wd)
	if err != nil {
		return statusFail(stderr, err)
	}
	if err := slice.SetTitle(ctx, r, pos[0], pos[1]); err != nil {
		return statusFail(stderr, err)
	}
	fmt.Fprintf(stdout, "Title of %s is now \"%s\".\n", pos[0], pos[1])
	return 0
}
