// Package slice is the core of sliceflow's slice commands. The CLI calls it.
package slice

// Error is a failure the spec names. Message holds the spec's words,
// without "Error: ".
type Error struct{ Message string }

// Error returns the message.
func (e *Error) Error() string { return e.Message }

// Warning is a note a result carries. Slice is empty when it is about the
// whole repo.
type Warning struct{ Slice, Text string }
