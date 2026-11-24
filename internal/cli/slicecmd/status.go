package slicecmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/BertPratya/sliceflow/internal/cli/command"
	"github.com/BertPratya/sliceflow/internal/slice"
)

// runSliceStatus prints the status of the slice named in in, from the
// user's checkout of the current folder, and returns 0. Lines of its log
// that cannot be read are named in a warning first, on stderr. When the slice
// cannot be shown, it prints why, and returns 2 for a name that cannot
// be used, or 1 for any other failure.
func runSliceStatus(ctx context.Context, in command.Input, stdout, stderr io.Writer) int {
	dir, err := os.Getwd()
	if err != nil {
		command.PrintFailure(stderr, "find the current folder", err, command.NothingChanged)
		return 1
	}
	status, err := slice.Status(ctx, dir, in.Args["name"])
	if err != nil {
		return printSliceError(stderr, err)
	}
	command.PrintBadLines(stderr, status.LogPath, status.BadLines)
	fmt.Fprint(stdout, statusText(status))
	return 0
}

// statusText returns the status of a slice as it is printed.
func statusText(s slice.StatusResult) string {
	var b strings.Builder
	fmt.Fprintf(&b, statusFirstLine, s.Name, s.Created.Title)

	description := s.Created.Description
	if description == "" {
		description = noneValue
	}
	fmt.Fprintf(&b, statusLine, "Description:", description)

	created := s.CreatedAt
	if at, err := time.Parse(time.RFC3339, s.CreatedAt); err == nil {
		created = at.UTC().Format(createdTimeLayout)
	}
	fmt.Fprintf(&b, statusLine, "Created:", created)
	fmt.Fprintf(&b, statusLine, "Base:", s.Created.Base)
	fmt.Fprintf(&b, statusLine, "Events:", fmt.Sprintf(eventsValue, s.Events, s.LastEvent))
	return b.String()
}

// printSliceError prints the message for err, an error of slice.Status,
// slice.List, or slice.Create, and returns the exit code.
func printSliceError(w io.Writer, err error) int {
	var (
		noLog   *slice.NoLogError
		noEvent *slice.NoCreatedEventError
	)
	switch {
	case errors.As(err, &noLog):
		fmt.Fprintf(w, noSliceLine, noLog.Name, noLog.Name)
		if noLog.Branch == "" {
			fmt.Fprint(w, noSliceNoBranchLine)
		} else {
			fmt.Fprintf(w, noSliceOnBranchLine, noLog.Branch)
		}
		return 1
	case errors.As(err, &noEvent):
		command.PrintBadLines(w, noEvent.LogPath, noEvent.BadLines)
		fmt.Fprintf(w, noCreatedEventLines, noEvent.Name)
		return 1
	}
	return printCreateError(w, err)
}
