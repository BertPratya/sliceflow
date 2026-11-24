package command

// NothingChanged ends a first line of an error when sliceflow changed
// nothing.
const NothingChanged = " Nothing was changed."

// The lines that say a tool or the system failed.
const (
	gitFailedLine    = "Error: git could not %s.%s\n"
	systemFailedLine = "Error: could not %s.%s\n"
	gitSaid          = "git said:\n"
	gitSaidNothing   = "git printed no message.\n"
	systemSaid       = "The system said:\n"
)

// The warning for lines of a log that could not be read.
const (
	oneBadLineWarning = "Warning: line %d of %s could not be read, so it was skipped.\n"
	badLinesWarning   = "Warning: lines %s of %s could not be read, so they were skipped.\n"
	moreBadLines      = " (%d more not listed)"
)
