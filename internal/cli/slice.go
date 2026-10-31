package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/BertPratya/sliceflow/internal/slice"
)

// sliceCreateUsage is the help text of "sliceflow slice create".
const sliceCreateUsage = `Create a new slice: a branch, a worktree, a spec to fill in, and an event log.

Usage:
  sliceflow slice create <name> [flags]

Flags:
  --from <branch>    The branch to start from (default: git config sliceflow.base, else main)
  --title <text>     The slice's title
  --json             Print the result as JSON
  -h, --help         Show this help
`

// sliceCreateFlags is the flags that "slice create" accepts.
var sliceCreateFlags = flagSpec{values: []string{"from", "title"}, bools: []string{"json"}}

// runSliceCreate runs "sliceflow slice create". args are the words after
// "create". It checks the words, then asks the slice package to create the slice.
func runSliceCreate(e *env, args []string) int {
	f, positional, err := parseFlags(args, sliceCreateFlags)
	if f.help {
		fmt.Fprint(e.stdout, sliceCreateUsage)
		return exitOK
	}
	if err != nil {
		return wrongUse(e, err.Error())
	}
	if len(positional) == 0 || positional[0] == "" {
		return wrongUse(e, "missing slice name")
	}
	if len(positional) > 1 {
		return wrongUse(e, "too many arguments")
	}

	dir, err := e.workDir()
	if err != nil {
		fmt.Fprintf(e.stderr, "Error: %v\n", err)
		return exitError
	}

	// Ctrl-C or a stop signal cancels ctx; Create then undoes its work (R18).
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	res, err := slice.Create(ctx, slice.CreateOptions{
		Dir:   dir,
		Name:  positional[0],
		From:  f.values["from"],
		Title: f.values["title"],
		Now:   time.Now(),
	})
	if err != nil {
		fmt.Fprintf(e.stderr, "Error: %v\n", err)
		return exitError
	}

	if f.bools["json"] {
		return printCreatedJSON(e, res)
	}
	printCreated(e, res)
	return exitOK
}

// printCreated prints the lines of R9 (S1). Worktree and Next are relative
// to the main checkout; Spec is the path inside the worktree.
func printCreated(e *env, r slice.Result) {
	fmt.Fprintf(e.stdout, "Created slice %s.\n", r.Name)
	fmt.Fprintf(e.stdout, "  Branch:   %s (from %s @ %s)\n", r.Branch, r.From.Branch, shortCommit(r.From.Commit))
	fmt.Fprintf(e.stdout, "  Worktree: %s\n", r.WorktreeRel)
	fmt.Fprintf(e.stdout, "  Spec:     %s\n", r.Spec)
	fmt.Fprintf(e.stdout, "Next: write the spec in %s/%s\n", r.WorktreeRel, r.Spec)
}

// shortCommit returns the first 7 characters of a commit id.
func shortCommit(id string) string {
	if len(id) > 7 {
		return id[:7]
	}
	return id
}

// createdJSON is the object that "slice create --json" prints (R10). The
// field order here is the order in the output.
type createdJSON struct {
	Slice    string      `json:"slice"`
	Branch   string      `json:"branch"`
	From     createdFrom `json:"from"`
	Worktree string      `json:"worktree"`
	Spec     string      `json:"spec"`
}

// createdFrom is the "from" object of the JSON: the base branch and the
// full commit id.
type createdFrom struct {
	Branch string `json:"branch"`
	Commit string `json:"commit"`
}

// printCreatedJSON prints the result as one line of JSON (R10), without
// escaping &, < or >. The encoder adds the new line at the end.
func printCreatedJSON(e *env, r slice.Result) int {
	enc := json.NewEncoder(e.stdout)
	enc.SetEscapeHTML(false)
	err := enc.Encode(createdJSON{
		Slice:    r.Name,
		Branch:   r.Branch,
		From:     createdFrom{Branch: r.From.Branch, Commit: r.From.Commit},
		Worktree: r.Worktree,
		Spec:     r.Spec,
	})
	if err != nil {
		fmt.Fprintf(e.stderr, "Error: %v\n", err)
		return exitError
	}
	return exitOK
}

// sliceStatusUsage is the help text of "sliceflow slice status".
const sliceStatusUsage = `Show where a slice stands. Without a name, show the slice of the current branch.

Usage:
  sliceflow slice status [name] [flags]

Flags:
  --json         Print the result as JSON
  -h, --help     Show this help
`

// sliceStatusFlags is the flags that "slice status" accepts.
var sliceStatusFlags = flagSpec{bools: []string{"json"}}

// runSliceStatus runs "sliceflow slice status". args are the words after
// "status". Warnings go to stderr (R28), also with --json (R42).
func runSliceStatus(e *env, args []string) int {
	f, positional, err := parseFlags(args, sliceStatusFlags)
	if f.help {
		fmt.Fprint(e.stdout, sliceStatusUsage)
		return exitOK
	}
	if err != nil {
		return wrongUse(e, err.Error())
	}
	if len(positional) > 1 {
		return wrongUse(e, "too many arguments")
	}
	name := ""
	if len(positional) == 1 {
		name = positional[0]
	}

	dir, err := e.workDir()
	if err != nil {
		fmt.Fprintf(e.stderr, "Error: %v\n", err)
		return exitError
	}

	info, err := slice.Status(dir, name)
	// Repo warnings come first, also when the command fails.
	printWarnings(e, info.RepoWarnings, "", nil)
	if err != nil {
		fmt.Fprintf(e.stderr, "Error: %v\n", err)
		return exitError
	}

	printWarnings(e, nil, info.Name, info.Warnings)

	if f.bools["json"] {
		return printStatusJSON(e, info)
	}
	printStatus(e, info)
	return exitOK
}

// printStatus prints the lines of R29 (S13), with State: and Next: of
// 0002 R24 (0002 S11). A value the log does not have prints as "-", so
// State: and Next: are "-" when the phase is unknown.
func printStatus(e *env, info slice.Info) {
	title := info.Title
	if title == "" {
		title = info.Name
	}
	from, created := "-", "-"
	state, next := "-", "-"
	if info.State != "" {
		state = info.State
	}
	if info.Next != "" {
		next = info.Next
	}
	if info.From != nil {
		from = info.From.Branch + " @ " + shortCommit(info.From.Commit)
	}
	if !info.Created.IsZero() {
		created = info.Created.UTC().Format("2006-01-02 15:04") + " UTC"
	}
	fmt.Fprintf(e.stdout, "Slice:    %s\n", safeText(info.Name))
	fmt.Fprintf(e.stdout, "Title:    %s\n", safeText(title))
	fmt.Fprintf(e.stdout, "Phase:    %s\n", safeText(info.Phase))
	fmt.Fprintf(e.stdout, "State:    %s\n", safeText(state))
	fmt.Fprintf(e.stdout, "From:     %s\n", safeText(from))
	fmt.Fprintf(e.stdout, "Created:  %s\n", created)
	fmt.Fprintf(e.stdout, "Next:     %s\n", safeText(next))
}

// safeText makes text from a log safe to print in a terminal: each control
// character (unicode.IsControl, e.g. ESC, \r, \n, \t) becomes \uXXXX, so
// it cannot send escape codes.
func safeText(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsControl(r) {
			fmt.Fprintf(&b, "\\u%04x", r)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// statusJSON is the object that "slice status --json" prints (R30). The
// field order here is the order in the output. A pointer or an omitempty
// field is left out when the log does not have the value. State and Next
// come after Phase, and are left out when the phase is unknown (0002 R26).
type statusJSON struct {
	Slice    string       `json:"slice"`
	Title    string       `json:"title,omitempty"`
	Phase    string       `json:"phase"`
	State    string       `json:"state,omitempty"`
	Next     string       `json:"next,omitempty"`
	From     *createdFrom `json:"from,omitempty"`
	Created  string       `json:"created,omitempty"`
	Warnings []string     `json:"warnings"`
}

// toStatusJSON builds the R30 object for a slice. Its warnings are the
// repoWarnings first, then its own; the list is [] when there are none.
func toStatusJSON(info slice.Info, repoWarnings []string) statusJSON {
	out := statusJSON{
		Slice:    info.Name,
		Title:    info.Title,
		Phase:    info.Phase,
		State:    info.State,
		Next:     info.Next,
		Warnings: append(append([]string{}, repoWarnings...), info.Warnings...),
	}
	if info.From != nil {
		out.From = &createdFrom{Branch: info.From.Branch, Commit: info.From.Commit}
	}
	if !info.Created.IsZero() {
		out.Created = info.Created.UTC().Format("2006-01-02T15:04:05Z")
	}
	return out
}

// printStatusJSON prints the status as one line of JSON (R30). The list of
// warnings has the repo warnings first, and is [] when there are none.
func printStatusJSON(e *env, info slice.Info) int {
	out := toStatusJSON(info, info.RepoWarnings)

	enc := json.NewEncoder(e.stdout)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(out); err != nil {
		fmt.Fprintf(e.stderr, "Error: %v\n", err)
		return exitError
	}
	return exitOK
}

// sliceListUsage is the help text of "sliceflow slice list".
const sliceListUsage = `List the slices in progress.

Usage:
  sliceflow slice list [flags]

Flags:
  --all          Also list the merged slices
  --json         Print the result as JSON
  -h, --help     Show this help
`

// sliceListFlags is the flags that "slice list" accepts.
var sliceListFlags = flagSpec{bools: []string{"all", "json"}}

// runSliceList runs "sliceflow slice list". args are the words after
// "list". Warnings go to stderr (R28), also with --json.
func runSliceList(e *env, args []string) int {
	f, positional, err := parseFlags(args, sliceListFlags)
	if f.help {
		fmt.Fprint(e.stdout, sliceListUsage)
		return exitOK
	}
	if err != nil {
		return wrongUse(e, err.Error())
	}
	if len(positional) > 0 {
		return wrongUse(e, "too many arguments")
	}

	dir, err := e.workDir()
	if err != nil {
		fmt.Fprintf(e.stderr, "Error: %v\n", err)
		return exitError
	}

	all := f.bools["all"]
	listing, err := slice.List(dir, all)
	if err != nil {
		fmt.Fprintf(e.stderr, "Error: %v\n", err)
		return exitError
	}

	// Repo warnings first, then the warnings of each slice in row order.
	printWarnings(e, listing.Warnings, "", nil)
	for _, info := range listing.Slices {
		printWarnings(e, nil, info.Name, info.Warnings)
	}

	if f.bools["json"] {
		return printListJSON(e, listing)
	}
	printList(e, listing.Slices, all)
	return exitOK
}

// printList prints the table of R34, or the line for no rows (R34, R35).
// Each column is as wide as its widest value, header included.
func printList(e *env, infos []slice.Info, all bool) {
	if len(infos) == 0 {
		if all {
			fmt.Fprintln(e.stdout, "No slices.")
		} else {
			fmt.Fprintln(e.stdout, "No slices in progress.")
		}
		return
	}

	rows := [][]string{{"NAME", "TITLE", "PHASE", "CREATED (UTC)"}}
	for _, info := range infos {
		title := info.Title
		if title == "" {
			title = info.Name
		}
		phase := safeText(info.Phase)
		if len(info.Warnings) > 0 {
			phase += " !"
		}
		created := "-"
		if !info.Created.IsZero() {
			created = info.Created.UTC().Format("2006-01-02 15:04")
		}
		rows = append(rows, []string{safeText(info.Name), safeText(title), phase, created})
	}

	// widths[i] is the width of column i. It counts runes, so wide
	// characters (CJK, emoji) can misalign. This is accepted.
	widths := make([]int, 4)
	for _, row := range rows {
		for i, cell := range row {
			widths[i] = max(widths[i], utf8.RuneCountInString(cell))
		}
	}
	for _, row := range rows {
		var line strings.Builder
		for i, cell := range row {
			line.WriteString(cell)
			// The last column gets no spaces after it.
			if i < len(row)-1 {
				line.WriteString(strings.Repeat(" ", widths[i]-utf8.RuneCountInString(cell)+2))
			}
		}
		fmt.Fprintln(e.stdout, line.String())
	}
}

// listJSON is the object that "slice list --json" prints (R37).
type listJSON struct {
	Slices   []statusJSON `json:"slices"`
	Warnings []string     `json:"warnings"`
}

// printListJSON prints the listing as one line of JSON (R37). Each slice
// holds only its own warnings; the repo warnings are at the top. Empty
// lists are [].
func printListJSON(e *env, l slice.Listing) int {
	out := listJSON{Slices: []statusJSON{}, Warnings: append([]string{}, l.Warnings...)}
	for _, info := range l.Slices {
		out.Slices = append(out.Slices, toStatusJSON(info, nil))
	}

	enc := json.NewEncoder(e.stdout)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(out); err != nil {
		fmt.Fprintf(e.stderr, "Error: %v\n", err)
		return exitError
	}
	return exitOK
}

// sliceTitleUsage is the help text of "sliceflow slice title".
const sliceTitleUsage = `Set the title of a slice. The title is saved in the slice's event log.

Usage:
  sliceflow slice title <name> <title> [flags]

Flags:
  --json         Print the result as JSON
  -h, --help     Show this help
`

// sliceTitleFlags is the flags that "slice title" accepts.
var sliceTitleFlags = flagSpec{bools: []string{"json"}}

// runSliceTitle runs "sliceflow slice title" (0002 R29-R35). args are the
// words after "title". Warnings go to stderr, also when it fails.
func runSliceTitle(e *env, args []string) int {
	f, positional, err := parseFlags(args, sliceTitleFlags)
	if f.help {
		fmt.Fprint(e.stdout, sliceTitleUsage)
		return exitOK
	}
	if err != nil {
		return wrongUse(e, err.Error())
	}
	if len(positional) == 0 || positional[0] == "" {
		return wrongUse(e, "missing slice name")
	}
	// A title of "" is given, so it is not wrong use; SetTitle gives R31's
	// error.
	if len(positional) == 1 {
		return wrongUse(e, "missing title")
	}
	if len(positional) > 2 {
		return wrongUse(e, "too many arguments")
	}

	dir, err := e.workDir()
	if err != nil {
		fmt.Fprintf(e.stderr, "Error: %v\n", err)
		return exitError
	}

	name := positional[0]
	res, err := slice.SetTitle(dir, name, positional[1], time.Now())
	printWarnings(e, nil, name, res.Warnings)
	if err != nil {
		fmt.Fprintf(e.stderr, "Error: %v\n", err)
		return exitError
	}

	if f.bools["json"] {
		return printTitleJSON(e, res)
	}
	fmt.Fprintf(e.stdout, "Title of %s set: %s\n", res.Name, safeText(res.Title))
	return exitOK
}

// titleJSON is the object that "slice title --json" prints (0002 R30).
type titleJSON struct {
	Slice string `json:"slice"`
	Title string `json:"title"`
}

// printTitleJSON prints the result as one line of JSON (0002 R30).
func printTitleJSON(e *env, r slice.TitleResult) int {
	enc := json.NewEncoder(e.stdout)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(titleJSON{Slice: r.Name, Title: r.Title}); err != nil {
		fmt.Fprintf(e.stderr, "Error: %v\n", err)
		return exitError
	}
	return exitOK
}

// printWarnings prints warnings on stderr (R28). Each repo warning prints
// as "Warning: <warning>", then each warning of the slice name prints as
// "Warning: <name>: <warning>".
func printWarnings(e *env, repoWarnings []string, name string, warnings []string) {
	for _, w := range repoWarnings {
		fmt.Fprintf(e.stderr, "Warning: %s\n", w)
	}
	for _, w := range warnings {
		fmt.Fprintf(e.stderr, "Warning: %s: %s\n", name, w)
	}
}
