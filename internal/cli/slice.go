package cli

import (
	"encoding/json"
	"fmt"
	"time"

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

	res, err := slice.Create(slice.CreateOptions{
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
