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
	for _, w := range info.RepoWarnings {
		fmt.Fprintf(e.stderr, "Warning: %s\n", w)
	}
	if err != nil {
		fmt.Fprintf(e.stderr, "Error: %v\n", err)
		return exitError
	}

	for _, w := range info.Warnings {
		fmt.Fprintf(e.stderr, "Warning: %s: %s\n", info.Name, w)
	}

	if f.bools["json"] {
		return printStatusJSON(e, info)
	}
	printStatus(e, info)
	return exitOK
}

// printStatus prints the lines of R29 (S13). A value the log does not
// have prints as "-".
func printStatus(e *env, info slice.Info) {
	title := info.Title
	if title == "" {
		title = info.Name
	}
	from, created := "-", "-"
	if info.From != nil {
		from = info.From.Branch + " @ " + shortCommit(info.From.Commit)
	}
	if !info.Created.IsZero() {
		created = info.Created.UTC().Format("2006-01-02 15:04") + " UTC"
	}
	fmt.Fprintf(e.stdout, "Slice:    %s\n", safeText(info.Name))
	fmt.Fprintf(e.stdout, "Title:    %s\n", safeText(title))
	fmt.Fprintf(e.stdout, "Phase:    %s\n", safeText(info.Phase))
	fmt.Fprintf(e.stdout, "From:     %s\n", safeText(from))
	fmt.Fprintf(e.stdout, "Created:  %s\n", created)
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
// field is left out when the log does not have the value.
type statusJSON struct {
	Slice    string       `json:"slice"`
	Title    string       `json:"title,omitempty"`
	Phase    string       `json:"phase"`
	From     *createdFrom `json:"from,omitempty"`
	Created  string       `json:"created,omitempty"`
	Warnings []string     `json:"warnings"`
}

// printStatusJSON prints the status as one line of JSON (R30). The list of
// warnings has the repo warnings first, and is [] when there are none.
func printStatusJSON(e *env, info slice.Info) int {
	out := statusJSON{
		Slice:    info.Name,
		Title:    info.Title,
		Phase:    info.Phase,
		Warnings: append(append([]string{}, info.RepoWarnings...), info.Warnings...),
	}
	if info.From != nil {
		out.From = &createdFrom{Branch: info.From.Branch, Commit: info.From.Commit}
	}
	if !info.Created.IsZero() {
		out.Created = info.Created.UTC().Format("2006-01-02T15:04:05Z")
	}

	enc := json.NewEncoder(e.stdout)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(out); err != nil {
		fmt.Fprintf(e.stderr, "Error: %v\n", err)
		return exitError
	}
	return exitOK
}
