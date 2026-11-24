package state

import (
	"testing"

	"github.com/BertPratya/sliceflow/internal/eventlog"
)

// These tests are inside the package, so they can add moves that 0002
// does not have yet, and check the table itself.

var (
	specDraft  = Position{Phase: PhaseSpec, Step: StepDraft}
	specReview = Position{Phase: PhaseSpec, Step: StepReview}
)

// withMoves adds rows to transitions for one test, and puts the table
// back when the test ends.
func withMoves(t *testing.T, rows ...Transition) {
	t.Helper()
	saved := transitions
	transitions = append(append([]Transition{}, transitions...), rows...)
	t.Cleanup(func() { transitions = saved })
}

// event is an event read from line, with no data.
func event(line int, name string) eventlog.Event {
	return eventlog.Event{Version: 1, Seq: line, Line: line, Event: name}
}

// createdEvent is a slice.created event read from line 1.
func createdEvent() eventlog.Event {
	return eventlog.Event{
		Version: 1, Seq: 1, Line: 1, Event: eventlog.SliceCreatedEvent,
		Payload: eventlog.SliceCreated{Name: "login", Title: "Log in"},
	}
}

// Test 10: a second move in the table is followed.
func TestReplayFollowsAnAddedMove(t *testing.T) {
	withMoves(t, Transition{From: specDraft, Event: "test.submitted", To: specReview})

	s := Replay([]eventlog.Event{createdEvent(), event(2, "test.submitted")})

	if s.Unknown != nil || s.Position != specReview {
		t.Errorf("got Position %+v and Unknown %+v, want %+v and nil", s.Position, s.Unknown, specReview)
	}
}

// Test 11: once the position is unknown, a later move that would be
// allowed does not make it known again.
func TestReplayUnknownStaysUnknown(t *testing.T) {
	withMoves(t, Transition{From: specDraft, Event: "test.submitted", To: specReview})

	s := Replay([]eventlog.Event{
		createdEvent(),
		event(2, "spec.reopened"),
		event(3, "test.submitted"),
	})

	if s.Unknown == nil {
		t.Fatalf("Unknown is nil, want the event on line 2")
	}
	if s.Unknown.Line != 2 || s.Unknown.Reason != ReasonUnknownEvent {
		t.Errorf("Unknown is %+v, want line 2, %q", *s.Unknown, ReasonUnknownEvent)
	}
	if s.Position != specDraft {
		t.Errorf("Position is %+v, want %+v: the move on line 3 must not apply", s.Position, specDraft)
	}
}

// Test 12: a move whose guard says no is not allowed, and a known event
// that is not allowed gives ReasonNotAllowed, not ReasonUnknownEvent.
func TestReplayGuardThatSaysNo(t *testing.T) {
	withMoves(t, Transition{
		From: specDraft, Event: "test.submitted", To: specReview,
		Guard: func(State) bool { return false },
	})

	s := Replay([]eventlog.Event{createdEvent(), event(2, "test.submitted")})

	if s.Unknown == nil || s.Unknown.Reason != ReasonNotAllowed {
		t.Fatalf("Unknown is %+v, want line 2, %q", s.Unknown, ReasonNotAllowed)
	}
	if s.Position != specDraft {
		t.Errorf("Position is %+v, want %+v", s.Position, specDraft)
	}
}

// Test 13: the table is well formed: no two rows without a guard have
// the same From and Event, no row moves back to the start, and no event
// is both a move and allowed everywhere.
func TestTransitionsTable(t *testing.T) {
	seen := map[[2]any]bool{}
	for i, row := range transitions {
		if row.Guard == nil {
			key := [2]any{row.From, row.Event}
			if seen[key] {
				t.Errorf("row %d: another row without a guard has From %+v and Event %q", i, row.From, row.Event)
			}
			seen[key] = true
		}
		if row.To == start {
			t.Errorf("row %d: To is the start position", i)
		}
		for _, everywhere := range allowedEverywhere {
			if row.Event == everywhere {
				t.Errorf("row %d: %q is also in allowedEverywhere, so the row never applies", i, row.Event)
			}
		}
	}
}
