// Package state works out where a slice is, by replaying its event log
// through a table of moves (R23).
package state

import (
	"fmt"

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

// next moves current along the move for event, and reports whether there
// was one.
func next(current *Place, event eventlog.Event) bool {
	for _, m := range moves {
		if *current == m.from && event.Event == m.event {
			*current = m.to
			return true
		}
	}
	return false
}

// Replay returns the place a slice ends up in after its events, in file
// order, following the moves in the table (R23). An event with no move is
// skipped, with a warning (R27).
func Replay(events []eventlog.Event) (Place, []string) {
	current := Unknown
	var warnings []string
	for _, event := range events {
		if !next(&current, event) {
			warnings = append(warnings, fmt.Sprintf(
				".events.jsonl line %d: %s is not allowed in %s; skipped",
				event.Line, event.Event, describe(current)))
		}
	}
	return current, warnings
}

// describe writes a place as in R27's warning: "planning (spec_draft)", or
// "unknown" when there is no state.
func describe(p Place) string {
	if p.State == "" {
		return p.Phase
	}
	return p.Phase + " (" + p.State + ")"
}
