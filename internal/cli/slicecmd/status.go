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
	"github.com/BertPratya/sliceflow/internal/state"
)

// runSliceStatus prints the status of the slice named in in, from the
// user's checkout of the current folder, and returns 0. Lines of its log
// that cannot be read are named in a warning first, on stderr, then the
// event that stopped the state from being worked out, if one did. When the slice
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
		return PrintSliceError(stderr, err)
	}
	command.PrintBadLines(stderr, status.LogPath, status.BadLines)
	command.PrintStateWarning(stderr, status.LogPath, status.State)
	fmt.Fprint(stdout, statusText(status))
	return 0
}

// statusText returns the status of a slice as it is printed: the title,
// the state, the other lines, and, when the state is known and has one,
// the next step after an empty line.
func statusText(s slice.StatusResult) string {
	var b strings.Builder
	fmt.Fprintf(&b, statusFirstLine, s.Name, s.State.Detail.Title)

	position := unknownValue
	if s.State.Unknown == nil {
		position = s.State.Position.String()
	}
	fmt.Fprintf(&b, statusLine, "State:", position)

	description := s.State.Detail.Description
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

	if next := nextStep(s); s.State.Unknown == nil && next != "" {
		b.WriteString("\n")
		b.WriteString(next)
	}
	return b.String()
}

// nextStep returns the lines that say what to do next in the slice's
// state, or "" when sliceflow has none for it.
func nextStep(s slice.StatusResult) string {
	switch s.State.Position {
	case state.Position{Phase: state.PhaseSpec, Step: state.StepDraft}:
		return fmt.Sprintf(nextWriteSpecLines, s.Spec, s.Name)
	}
	return ""
}

// PrintSliceError prints the message for err, an error of slice.Status,
// slice.List, slice.SpecPath, or slice.Create, and returns the exit
// code. Every command that finds a slice by its name prints its errors
// with it, so the same mistake gets the same message.
func PrintSliceError(w io.Writer, err error) int {
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
