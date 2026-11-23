package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/BertPratya/sliceflow/internal/slice"
)

// runSliceStatus prints the status of the slice named in in, from the
// user's checkout of the current folder, and returns 0. Lines of its log
// that cannot be read are named in a warning first, on stderr. When the slice
// cannot be shown, it prints why, and returns 2 for a name that cannot
// be used, or 1 for any other failure.
func runSliceStatus(ctx context.Context, in input, stdout, stderr io.Writer) int {
	dir, err := os.Getwd()
	if err != nil {
		printFailure(stderr, "find the current folder", err, nothingChanged)
		return 1
	}
	status, err := slice.Status(ctx, dir, in.args["name"])
	if err != nil {
		return printSliceError(stderr, err)
	}
	printBadLines(stderr, status.LogPath, status.BadLines)
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

// printBadLines prints one warning that names the lines of the log at
// path that could not be read, the first 10 of them at most. It prints
// nothing when there are none.
func printBadLines(w io.Writer, path string, lines []int) {
	switch {
	case len(lines) == 0:
		return
	case len(lines) == 1:
		fmt.Fprintf(w, oneBadLineWarning, lines[0], path)
		return
	}
	shown := lines[:min(len(lines), 10)]
	numbers := make([]string, len(shown))
	for i, line := range shown {
		numbers[i] = strconv.Itoa(line)
	}
	list := strings.Join(numbers, ", ")
	if len(lines) > 10 {
		list += fmt.Sprintf(moreBadLines, len(lines)-10)
	}
	fmt.Fprintf(w, badLinesWarning, list, path)
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
		printBadLines(w, noEvent.LogPath, noEvent.BadLines)
		fmt.Fprintf(w, noCreatedEventLines, noEvent.Name)
		return 1
	}
	return printCreateError(w, err)
}
