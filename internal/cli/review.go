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

const reviewUsage = `Usage: sliceflow spec review [<name>]

Reviews docs/slices/<name>/spec.md, which must be submitted. It saves a
review.started event, runs the reviewer, and saves a review.finished
event. Without a name, it uses the slice of the current branch
(slice/<name>).
`

func init() {
	register(command{Name: "spec review", Usage: reviewUsage, Run: runReview})
}

func runReview(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("spec review", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	pos, err := parseFlags(fs, reviewUsage, args, stdout)
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
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	wd, err := os.Getwd()
	if err != nil {
		return fail(stderr, err)
	}
	r, err := slice.FindRepo(ctx, wd)
	if err != nil {
		if ctx.Err() != nil {
			err = slice.InterruptedSave()
		}
		return statusFail(stderr, err)
	}
	err = slice.ReviewSpec(ctx, r, name, func(s slice.Step) {
		switch s.Kind {
		case "started":
			fmt.Fprintf(stdout, "Review %d started: %s (reviewer: %s)\n", s.Round, s.File, s.Reviewer)
		case "finished":
			if s.Result == "passed" {
				fmt.Fprintf(stdout, "Review %d passed.\n", s.Round)
				fmt.Fprintln(stdout, "Next: write the plan.")
			}
		}
	})
	if err != nil {
		return statusFail(stderr, err)
	}
	return 0
}
