package speccmd

// The lines of spec validate.
const (
	noProblemsLine = "%s: no problems found.\n"
	problemLines   = "%s:%d: error: %s\n  %s\n"
	oneProblemLine = "1 problem found.\n"
	problemsLine   = "%d problems found.\n"
	noSpecLines    = "Error: the spec of slice %s was not found at %s. Nothing was changed.\n" +
		"To start from the template, run: sliceflow spec template\n"
)
