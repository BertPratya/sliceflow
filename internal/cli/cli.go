package cli

// Run runs the command the user typed, and returns the exit code. args are
// the user's words, without the program path: os.Args[1:].
func Run(args []string) int {
	return rootCommands.route(args)
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
		{name: "create", args: "<name>", about: "Create a new slice", run: runSliceCreate},
	},
}
