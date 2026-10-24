package cli

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/BertPratya/sliceflow/internal/slice"
)

const statusUsage = `Usage: sliceflow slice status [<name>] [--json]

Shows the status of a slice. Without a name, it shows the slice of the
current branch (slice/<name>).

Flags:
  --json   print one JSON object instead of text
`

func init() {
	register(command{Name: "slice status", Usage: statusUsage, Run: runStatus})
}

func runStatus(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("slice status", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	asJSON := fs.Bool("json", false, "")
	pos, err := parseFlags(fs, statusUsage, args, stdout)
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
	st, err := slice.Status(ctx, r, name)
	if err != nil {
		return statusFail(stderr, err)
	}
	if *asJSON {
		obj := newSliceJSON(st)
		b, err := json.Marshal(obj)
		if err != nil {
			return fail(stderr, err)
		}
		fmt.Fprintf(stdout, "%s\n", b)
		printWarnings(stderr, st.Warnings)
		return 0
	}
	title := st.Title
	if title == "" {
		title = st.Slice
	}
	from, created := "-", "-"
	if st.From != nil {
		from = st.From.Branch + " @ " + shortCommit(st.From.Commit)
	}
	if !st.Created.IsZero() {
		created = st.Created.UTC().Format("2006-01-02 15:04") + " UTC"
	}
	fmt.Fprintf(stdout, "Slice:    %s\n", st.Slice)
	fmt.Fprintf(stdout, "Title:    %s\n", title)
	fmt.Fprintf(stdout, "Phase:    %s\n", st.Phase)
	fmt.Fprintf(stdout, "From:     %s\n", from)
	fmt.Fprintf(stdout, "Created:  %s\n", created)
	printWarnings(stderr, st.Warnings)
	return 0
}

// statusFail prints an error of the core as "Error: " and its message, so
// the lines of R32 keep no indent. Any other error goes through fail.
func statusFail(stderr io.Writer, err error) int {
	var se *slice.Error
	if errors.As(err, &se) {
		fmt.Fprintf(stderr, "Error: %s\n", se.Message)
		return 1
	}
	return fail(stderr, err)
}

// shortCommit returns the first 7 characters of a commit id.
func shortCommit(id string) string {
	if len(id) > 7 {
		return id[:7]
	}
	return id
}

type fromJSON struct {
	Branch string `json:"branch"`
	Commit string `json:"commit"`
}

// sliceJSON is the JSON object of one slice (R30). `slice status --json` and
// each item of `slice list --json` (R37) both use it.
type sliceJSON struct {
	Slice    string    `json:"slice"`
	Title    string    `json:"title,omitempty"`
	Phase    string    `json:"phase"`
	From     *fromJSON `json:"from,omitempty"`
	Created  string    `json:"created,omitempty"`
	Warnings []string  `json:"warnings"`
}

// warningTexts returns the text of each warning; never nil, so JSON gets [].
func warningTexts(ws []slice.Warning) []string {
	texts := []string{}
	for _, w := range ws {
		texts = append(texts, w.Text)
	}
	return texts
}

func newSliceJSON(st slice.SliceStatus) sliceJSON {
	obj := sliceJSON{Slice: st.Slice, Title: st.Title, Phase: st.Phase, Warnings: warningTexts(st.Warnings)}
	if st.From != nil {
		obj.From = &fromJSON{st.From.Branch, st.From.Commit}
	}
	if !st.Created.IsZero() {
		obj.Created = st.Created.UTC().Format("2006-01-02T15:04:05Z")
	}
	return obj
}
