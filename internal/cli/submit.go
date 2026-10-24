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

const submitUsage = `Usage: sliceflow spec submit [<name>]

Checks docs/slices/<name>/spec.md in the slice's worktree as
"spec validate --strict" does, and saves a spec.submitted event. With an
error or a warning, the submit is refused and spec.md stays a draft. It
does not start the review. Without a name, it uses the slice of the
current branch (slice/<name>).
`

func init() {
	register(command{Name: "spec submit", Usage: submitUsage, Run: runSubmit})
}

func runSubmit(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("spec submit", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	pos, err := parseFlags(fs, submitUsage, args, stdout)
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
	sub, err := slice.Submit(ctx, r, name)
	if err != nil {
		return statusFail(stderr, err)
	}
	printProblems(stdout, sub.Problems)
	if sub.Refused == nil {
		fmt.Fprintln(stdout, "Submitted spec.md.")
		fmt.Fprintln(stdout, "Next: sliceflow spec review")
		return 0
	}
	last := "It stays a draft."
	if sub.Before == "review_passed" {
		// A refused submit keeps the review, so the status still has its round.
		st, err := slice.Status(context.WithoutCancel(ctx), r, name)
		if err != nil {
			return statusFail(stderr, err)
		}
		last = fmt.Sprintf("It keeps the result of review %d.", st.State.Spec.Review.Round)
	}
	fmt.Fprintf(stdout, "Submit refused: spec.md is not finished (%s, %s). %s\n",
		plural(sub.Refused.Errors, "error"), plural(sub.Refused.Warnings, "warning"), last)
	return 1
}
