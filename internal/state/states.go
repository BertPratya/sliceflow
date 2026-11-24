package state

// Phase is the part of the workflow a slice is in.
type Phase string

// Step is where a slice is inside its phase.
type Step string

const (
	PhaseSpec      Phase = "spec"
	PhasePlan      Phase = "plan"
	PhaseImplement Phase = "implement"
)

const (
	StepDraft   Step = "draft"
	StepReview  Step = "review"
	StepApprove Step = "approve"
	StepExecute Step = "execute"
	StepTest    Step = "test"
)

// Position is where a slice is in the workflow, e.g. spec (draft).
type Position struct {
	Phase Phase
	Step  Step
}

// start is the position before the first event: no phase, no step.
var start = Position{}

// Detail is what a slice is about: its current title and description.
type Detail struct {
	Title       string
	Description string
}

// Reason says why the position could not be worked out.
type Reason string

const (
	// ReasonUnknownEvent is an event this version of sliceflow does not
	// know.
	ReasonUnknownEvent Reason = "unknown event"
	// ReasonNotAllowed is a known event that is not allowed in the
	// position the slice was in.
	ReasonNotAllowed Reason = "not allowed"
)

// Unknown says at which event the position stopped being worked out,
// and why.
type Unknown struct {
	Line   int    // the line of the log the event was read from
	Event  string // the event's name
	Reason Reason
}

// State is a slice's state, worked out from its events.
type State struct {
	// Position is where the slice is. When Unknown is not nil, it is the
	// last position known, the one the slice was in at that event.
	Position Position
	Detail   Detail
	// Unknown is nil while the position is known.
	Unknown *Unknown
}
