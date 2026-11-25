package slicecmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/BertPratya/sliceflow/internal/cli/command"
	"github.com/BertPratya/sliceflow/internal/slice"
)

// runSliceList prints the slices in the user's checkout of the current
// folder in a table, and returns 0. A name that looked like a slice but
// is not listed gets a warning first, on stderr. When the project cannot be
// read, it prints why, and returns 1.
func runSliceList(ctx context.Context, in command.Input, stdout, stderr io.Writer) int {
	dir, err := os.Getwd()
	if err != nil {
		command.PrintFailure(stderr, "find the current folder", err, command.NothingChanged)
		return 1
	}
	list, err := slice.List(ctx, dir)
	if err != nil {
		return PrintSliceError(stderr, err)
	}
	for _, skipped := range list.Skipped {
		printSkipped(stderr, skipped)
	}
	if len(list.Slices) == 0 {
		fmt.Fprint(stdout, noSlicesLine)
		return 0
	}
	fmt.Fprint(stdout, listTable(list.Slices))
	return 0
}

// printSkipped prints the warning for a name that is not listed, which
// says why.
func printSkipped(w io.Writer, s slice.SkippedCandidate) {
	switch s.Reason {
	case slice.NotValidName:
		fmt.Fprintf(w, notValidWarning, s.Name)
	case slice.NoCreatedEvent:
		fmt.Fprintf(w, noCreatedEventWarning, s.Name, s.LogPath)
	default:
		fmt.Fprintf(w, noLogWarning, s.Name, s.Name)
	}
}

// listTable returns the table of slices: a header, then one row for each
// slice. The name and the title columns are padded to their longest
// value, counted in characters, with two spaces between columns. No line
// ends with a space.
func listTable(slices []slice.ListedSlice) string {
	nameWidth := utf8.RuneCountInString(listHeader[0])
	titleWidth := utf8.RuneCountInString(listHeader[1])
	for _, s := range slices {
		nameWidth = max(nameWidth, utf8.RuneCountInString(s.Name))
		titleWidth = max(titleWidth, utf8.RuneCountInString(s.Title))
	}

	var b strings.Builder
	row := func(name, title, description string) {
		line := fmt.Sprintf("%-*s  %-*s  %s", nameWidth, name, titleWidth, title, description)
		b.WriteString(strings.TrimRight(line, " "))
		b.WriteString("\n")
	}
	row(listHeader[0], listHeader[1], listHeader[2])
	for _, s := range slices {
		row(s.Name, s.Title, s.Description)
	}
	return b.String()
}
