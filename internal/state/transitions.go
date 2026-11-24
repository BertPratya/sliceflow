package state

import (
	"slices"

	"github.com/BertPratya/sliceflow/internal/eventlog"
)

// Transition is one move: in the position From, the event Event moves a
// slice to the position To. Guard, when not nil, is an extra condition
// the state must meet for the move to apply.
type Transition struct {
	From  Position
	Event string
	To    Position
	Guard func(State) bool
}

// transitions holds every move, one per row. A known event that no row
// allows in the slice's position is not allowed there.
var transitions = []Transition{
	{From: start, Event: eventlog.SliceCreatedEvent, To: Position{Phase: PhaseSpec, Step: StepDraft}},
}

// allowedEverywhere holds the events that are allowed in every position
// and do not move the slice.
var allowedEverywhere = []string{
	eventlog.SliceEditedEvent,
}

// isKnown reports whether event is one this version of sliceflow knows:
// it is in a row of transitions, or in allowedEverywhere.
func isKnown(event string) bool {
	if slices.Contains(allowedEverywhere, event) {
		return true
	}
	return slices.ContainsFunc(transitions, func(t Transition) bool {
		return t.Event == event
	})
}
