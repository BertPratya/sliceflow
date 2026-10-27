package cli

import (
	"fmt"
	"io"
	"os"
)

// The exit codes of sliceflow.
const (
	exitOK       = 0 // the command worked
	exitError    = 1 // the command failed
	exitWrongUse = 2 // the user typed something wrong
)

// env is what a command needs from the outside world: where to write, and
// which folder it runs in. Tests give it their own writers. If sliceflow
// could not find the current folder, dirErr says why.
type env struct {
	stdout, stderr io.Writer
	dir            string
	dirErr         error
}

// workDir returns the current folder. A command calls it when it needs the
// folder; if the folder is unknown, the command gets the error. Commands
// that do not need the folder, like --help, work without it.
func (e *env) workDir() (string, error) {
	return e.dir, e.dirErr
}

// Run runs the command the user typed, and returns the exit code. args are
// the user's words, without the program path: os.Args[1:].
func Run(args []string, stdout, stderr io.Writer) int {
	// The folder may be gone (deleted while we are in it). That must not
	// stop --help, so keep the error for the commands that need the folder.
	dir, err := os.Getwd()
	e := &env{stdout: stdout, stderr: stderr, dir: dir, dirErr: err}
	return rootCommands.route(e, args)
}

// wrongUse prints a wrong-use error on stderr, and returns the exit code 2.
// msg says what is wrong, e.g. `unknown flag --colour`.
func wrongUse(e *env, msg string) int {
	fmt.Fprintf(e.stderr, "Error: %s. Run sliceflow --help.\n", msg)
	return exitWrongUse
}

// rootCommands is the top level: the words that can follow "sliceflow".
var rootCommands = group{
	path:  "sliceflow",
	about: "sliceflow runs a spec-driven workflow with AI coding agents.",
	commands: []command{
		{name: "slice", args: "<command>", about: "Create and inspect slices", run: sliceCommands.route},
	},
}

// sliceCommands is the words that can follow "sliceflow slice".
var sliceCommands = group{
	path:  "sliceflow slice",
	about: "Work with slices: one piece of work, with its own branch and event log.",
	commands: []command{
		{name: "create", args: "<name>", about: "Create a new slice", usage: sliceCreateUsage, run: runSliceCreate},
		{name: "status", args: "[name]", about: "Show where a slice stands", usage: sliceStatusUsage, run: runSliceStatus},
		{name: "list", about: "List the slices in progress", usage: sliceListUsage, run: runSliceList},
	},
}
