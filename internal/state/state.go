// Package state works out where a slice is, by replaying its event log
// through a table of moves (R23).
package state

import (
	"github.com/BertPratya/sliceflow/internal/eventlog"
)

// Place is where a slice is: its phase, and its state inside that phase.
type Place struct {
	Phase string // e.g. "planning"
	State string // e.g. "spec_draft"; empty when the phase is "unknown"
}

// Unknown is the place of a slice before its first good event, and of a
// slice whose log has no good slice.created event.
var Unknown = Place{Phase: "unknown"}

// move is one allowed step: the event moves a slice from one place to
// another.
type move struct {
	event string
	from  Place
	to    Place
}

// moves is every allowed move. Later slices add their own moves here.
var moves = []move{
	{
		event: eventlog.EventSliceCreated,
		from:  Unknown,
		to:    Place{Phase: "planning", State: "spec_draft"},
	},
}

func next(current *Place, event eventlog.Event) {
	for _, m := range moves {
		if *current == m.from && event.Event == m.event {
			*current = m.to
			return
		}
	}
}

// Replay returns the place a slice ends up in after its events, in file
// order, following the moves in the table (R23).
func Replay(events []eventlog.Event) Place {

	current := Unknown

	for _, event := range events {
		next(&current, event)
	}
	return current

}
