package cli

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/BertPratya/sliceflow/internal/slice"
)

const createUsage = `Usage: sliceflow slice create <name> [--title <title>] [--from <branch>] [--json]

Creates a slice: a branch slice/<name>, a worktree in
.sliceflow/worktrees/<name>, a spec, and an event log.

Flags:
  --title <title>   a title for the slice
  --from <branch>   the base branch (default: git config sliceflow.base, else main)
  --json            print one JSON object instead of text
`

func init() {
	register(command{Name: "slice create", Usage: createUsage, Run: runCreate})
}

func runCreate(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("slice create", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	title := fs.String("title", "", "")
	from := fs.String("from", "", "")
	asJSON := fs.Bool("json", false, "")
	pos, err := parseFlags(fs, createUsage, args, stdout)
	if errors.Is(err, errHelp) {
		return 0
	}
	if err != nil {
		return usageError(stderr, err.Error())
	}
	switch {
	case len(pos) == 0:
		return usageError(stderr, "missing slice name")
	case len(pos) > 1:
		return usageError(stderr, "too many arguments")
	}
	wd, err := os.Getwd()
	if err != nil {
		return fail(stderr, err)
	}
	r, err := slice.FindRepo(ctx, wd)
	if err != nil {
		return fail(stderr, err)
	}
	c, err := slice.Create(ctx, r, slice.CreateOptions{Name: pos[0], Title: *title, From: *from})
	if err != nil {
		return fail(stderr, err)
	}
	if *asJSON {
		type fromJSON struct {
			Branch string `json:"branch"`
			Commit string `json:"commit"`
		}
		b, err := json.Marshal(struct {
			Slice    string   `json:"slice"`
			Branch   string   `json:"branch"`
			From     fromJSON `json:"from"`
			Worktree string   `json:"worktree"`
			Spec     string   `json:"spec"`
		}{c.Slice, c.Branch, fromJSON{c.From.Branch, c.From.Commit}, c.Worktree, c.Spec})
		if err != nil {
			return fail(stderr, err)
		}
		fmt.Fprintf(stdout, "%s\n", b)
		return 0
	}
	short := c.From.Commit
	if len(short) > 7 {
		short = short[:7]
	}
	wt := path.Join(".sliceflow", "worktrees", c.Slice)
	fmt.Fprintf(stdout, "Created slice %s.\n", c.Slice)
	fmt.Fprintf(stdout, "  Branch:   %s (from %s @ %s)\n", c.Branch, c.From.Branch, short)
	fmt.Fprintf(stdout, "  Worktree: %s\n", wt)
	fmt.Fprintf(stdout, "  Spec:     %s\n", c.Spec)
	fmt.Fprintf(stdout, "Next: write the spec in %s/%s\n", wt, c.Spec)
	return 0
}
