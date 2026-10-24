package cli

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/BertPratya/sliceflow/internal/slice"
)

const listUsage = `Usage: sliceflow slice list [--all] [--json]

Lists the slices in progress, oldest first.

Flags:
  --all    list merged slices too
  --json   print one JSON object instead of text
`

func init() {
	register(command{Name: "slice list", Usage: listUsage, Run: runList})
}

func runList(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("slice list", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	all := fs.Bool("all", false, "")
	asJSON := fs.Bool("json", false, "")
	pos, err := parseFlags(fs, listUsage, args, stdout)
	if errors.Is(err, errHelp) {
		return 0
	}
	if err != nil {
		return usageError(stderr, err.Error())
	}
	if len(pos) > 0 {
		return usageError(stderr, "too many arguments")
	}
	wd, err := os.Getwd()
	if err != nil {
		return fail(stderr, err)
	}
	r, err := slice.FindRepo(ctx, wd)
	if err != nil {
		return statusFail(stderr, err)
	}
	list, err := slice.List(ctx, r, *all)
	if err != nil {
		return statusFail(stderr, err)
	}
	if *asJSON {
		obj := struct {
			Slices   []sliceJSON `json:"slices"`
			Warnings []string    `json:"warnings"`
		}{Slices: []sliceJSON{}, Warnings: warningTexts(list.Warnings)}
		for _, st := range list.Slices {
			obj.Slices = append(obj.Slices, newSliceJSON(st))
		}
		b, err := json.Marshal(obj)
		if err != nil {
			return fail(stderr, err)
		}
		fmt.Fprintf(stdout, "%s\n", b)
	} else {
		fmt.Fprint(stdout, listTable(list.Slices, *all))
	}
	printWarnings(stderr, list.Warnings)
	for _, st := range list.Slices {
		printWarnings(stderr, st.Warnings)
	}
	return 0
}

// listTable builds the text of R34 and R35: a header and a row per slice,
// columns 2 spaces apart, each as wide as its widest value.
func listTable(slices []slice.SliceStatus, all bool) string {
	if len(slices) == 0 {
		if all {
			return "No slices.\n"
		}
		return "No slices in progress.\n"
	}
	rows := [][]string{{"NAME", "TITLE", "PHASE", "CREATED (UTC)"}}
	for _, st := range slices {
		title := st.Title
		if title == "" {
			title = st.Slice
		}
		phase := st.Phase
		if len(st.Warnings) > 0 {
			phase += " !"
		}
		created := "-"
		if !st.Created.IsZero() {
			created = st.Created.UTC().Format("2006-01-02 15:04")
		}
		rows = append(rows, []string{st.Slice, title, phase, created})
	}
	widths := make([]int, 4)
	for _, row := range rows {
		for i, cell := range row {
			if n := utf8.RuneCountInString(cell); n > widths[i] {
				widths[i] = n
			}
		}
	}
	var sb strings.Builder
	for _, row := range rows {
		for i, cell := range row {
			sb.WriteString(cell)
			// The last column gets no padding: no trailing spaces.
			if i < len(row)-1 {
				sb.WriteString(strings.Repeat(" ", widths[i]-utf8.RuneCountInString(cell)+2))
			}
		}
		sb.WriteString("\n")
	}
	return sb.String()
}
