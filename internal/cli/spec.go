package cli

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/BertPratya/sliceflow/internal/slice"
)

// specTemplateUsage is the help text of "sliceflow spec template".
const specTemplateUsage = `Print the built-in spec template, the one "slice create" writes, with {{slice}} and {{date}} left as they are.

Usage:
  sliceflow spec template [flags]

Flags:
  -h, --help     Show this help
`

// runSpecTemplate runs "sliceflow spec template" (0002 R28). args are the
// words after "template". It needs no git repo, so it never asks for the
// current folder.
func runSpecTemplate(e *env, args []string) int {
	f, positional, err := parseFlags(args, flagSpec{})
	if f.help {
		fmt.Fprint(e.stdout, specTemplateUsage)
		return exitOK
	}
	if err != nil {
		return wrongUse(e, err.Error())
	}
	if len(positional) > 0 {
		return wrongUse(e, "too many arguments")
	}

	text, err := slice.SpecTemplate()
	if err != nil {
		fmt.Fprintf(e.stderr, "Error: %v\n", err)
		return exitError
	}
	fmt.Fprint(e.stdout, text)
	return exitOK
}

// specValidateUsage is the help text of "sliceflow spec validate".
const specValidateUsage = `Check a slice's spec, and print each problem with its line and a hint. Without a name, check the spec of the current branch's slice.

Usage:
  sliceflow spec validate [name] [flags]

Flags:
  --json         Print the result as JSON
  -h, --help     Show this help
`

// specValidateFlags is the flags that "spec validate" accepts.
var specValidateFlags = flagSpec{bools: []string{"json"}}

// runSpecValidate runs "sliceflow spec validate" (0002 R1-R8). args are
// the words after "validate". A spec with problems prints them on stdout
// and exits with 1 (R3); an error (R5-R8) prints only on stderr.
func runSpecValidate(e *env, args []string) int {
	f, positional, err := parseFlags(args, specValidateFlags)
	if f.help {
		fmt.Fprint(e.stdout, specValidateUsage)
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

	// UTC, like the created date that "slice create" writes.
	res, err := slice.ValidateSpec(dir, name, time.Now().UTC())
	if err != nil {
		fmt.Fprintf(e.stderr, "Error: %v\n", err)
		return exitError
	}

	if f.bools["json"] {
		if code := printSpecJSON(e, res); code != exitOK {
			return code
		}
	} else {
		printSpecProblems(e, res)
	}
	if len(res.Problems) > 0 {
		return exitError
	}
	return exitOK
}

// printSpecProblems prints the lines of R2 and R3: two lines per problem,
// then the count, or only "No problems in <path>." The message and the
// path come from the user's files, so they go through safeText.
func printSpecProblems(e *env, r slice.SpecResult) {
	path := safeText(r.Path)
	if len(r.Problems) == 0 {
		fmt.Fprintf(e.stdout, "No problems in %s.\n", path)
		return
	}
	for _, p := range r.Problems {
		where := path
		if p.Line > 0 {
			where = fmt.Sprintf("%s:%d", path, p.Line)
		}
		fmt.Fprintf(e.stdout, "%s: error: %s\n", where, safeText(p.Message))
		fmt.Fprintf(e.stdout, "  hint: %s.\n", safeText(p.Hint))
	}
	if len(r.Problems) == 1 {
		fmt.Fprintln(e.stdout, "1 error.")
	} else {
		fmt.Fprintf(e.stdout, "%d errors.\n", len(r.Problems))
	}
}

// specJSON is the object that "spec validate --json" prints (R4). The
// field order here is the order in the output.
type specJSON struct {
	Slice    string        `json:"slice"`
	File     string        `json:"file"`
	Problems []problemJSON `json:"problems"`
	Errors   int           `json:"errors"`
}

// problemJSON is one problem in the JSON of R4. Line is left out when the
// problem has no line.
type problemJSON struct {
	Line    int    `json:"line,omitempty"`
	Level   string `json:"level"`
	Message string `json:"message"`
	Hint    string `json:"hint"`
}

// printSpecJSON prints the result as one line of JSON (R4). The list of
// problems is [] when there are none, and each hint ends with a period.
func printSpecJSON(e *env, r slice.SpecResult) int {
	out := specJSON{Slice: r.Name, File: r.File, Problems: []problemJSON{}, Errors: len(r.Problems)}
	for _, p := range r.Problems {
		out.Problems = append(out.Problems, problemJSON{
			Line:    p.Line,
			Level:   "error",
			Message: p.Message,
			Hint:    p.Hint + ".",
		})
	}

	enc := json.NewEncoder(e.stdout)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(out); err != nil {
		fmt.Fprintf(e.stderr, "Error: %v\n", err)
		return exitError
	}
	return exitOK
}
