package state_test

import (
	"testing"

	"github.com/BertPratya/sliceflow/internal/eventlog"
	"github.com/BertPratya/sliceflow/internal/state"
)

// specDraft is the position of a slice right after slice.created.
var specDraft = state.Position{Phase: state.PhaseSpec, Step: state.StepDraft}

// created is a slice.created event for the slice login, read from line,
// as Read returns it.
func created(line int) eventlog.Event {
	return eventlog.Event{
		Version: 1,
		Seq:     line,
		Line:    line,
		Event:   eventlog.SliceCreatedEvent,
		At:      "2026-11-02T09:15:30Z",
		Payload: eventlog.SliceCreated{
			Name:        "login",
			Title:       "Log in",
			Description: "Lets a user log in.",
			Base:        "main",
			BaseCommit:  "3f9a1c2e8b7d4a6f0c1e5d9b2a7f3c8e4d6b1a09",
		},
	}
}

// edited is a slice.edited event read from line. A nil title or
// description is a field the event does not change.
func edited(line int, title, description *string) eventlog.Event {
	return eventlog.Event{
		Version: 1,
		Seq:     line,
		Line:    line,
		Event:   eventlog.SliceEditedEvent,
		At:      "2026-11-02T10:00:00Z",
		Payload: eventlog.SliceEdited{Title: title, Description: description},
	}
}

// unknown is an event this version of sliceflow does not know, read from
// line: Read keeps it, with no Payload.
func unknown(line int, name string) eventlog.Event {
	return eventlog.Event{
		Version: 1,
		Seq:     line,
		Line:    line,
		Event:   name,
		At:      "2026-11-02T11:00:00Z",
		Data:    []byte("{}"),
	}
}

// text returns a pointer to s.
func text(s string) *string { return &s }

// checkDetail fails the test when s does not have the title and
// description wanted.
func checkDetail(t *testing.T, s state.State, wantTitle, wantDescription string) {
	t.Helper()
	if s.Detail.Title != wantTitle {
		t.Errorf("Title is %q, want %q", s.Detail.Title, wantTitle)
	}
	if s.Detail.Description != wantDescription {
		t.Errorf("Description is %q, want %q", s.Detail.Description, wantDescription)
	}
}

// checkKnown fails the test when the position of s is unknown, or not
// want.
func checkKnown(t *testing.T, s state.State, want state.Position) {
	t.Helper()
	if s.Unknown != nil {
		t.Errorf("Unknown is %+v, want nil", *s.Unknown)
	}
	if s.Position != want {
		t.Errorf("Position is %+v, want %+v", s.Position, want)
	}
}

// checkUnknown fails the test when the position of s is known, when
// Unknown is not want, or when the last known position is not
// lastKnown.
func checkUnknown(t *testing.T, s state.State, want state.Unknown, lastKnown state.Position) {
	t.Helper()
	if s.Unknown == nil {
		t.Fatalf("Unknown is nil, want %+v", want)
	}
	if *s.Unknown != want {
		t.Errorf("Unknown is %+v, want %+v", *s.Unknown, want)
	}
	if s.Position != lastKnown {
		t.Errorf("Position is %+v, want the last known one, %+v", s.Position, lastKnown)
	}
}

// Test 1: slice.created alone gives the title and description it holds.
func TestReplayCreatedGivesDetail(t *testing.T) {
	s := state.Replay([]eventlog.Event{created(1)})

	checkDetail(t, s, "Log in", "Lets a user log in.")
}

// Test 2: slice.created alone moves the slice into spec (draft).
func TestReplayCreatedGivesSpecDraft(t *testing.T) {
	s := state.Replay([]eventlog.Event{created(1)})

	checkKnown(t, s, specDraft)
}

// Test 3: slice.edited with only a title changes the title, keeps the
// description from slice.created, and does not move the slice.
func TestReplayEditedTitleKeepsTheRest(t *testing.T) {
	events := []eventlog.Event{
		created(1),
		edited(2, text("Sign in"), nil),
	}

	s := state.Replay(events)

	checkDetail(t, s, "Sign in", "Lets a user log in.")
	checkKnown(t, s, specDraft)
}

// Test 4: with two edits, each field has its latest value. An edit that
// holds only the description does not wipe out the title, and a
// description set to "" means no description.
func TestReplayTwoEditsKeepEachLatestField(t *testing.T) {
	events := []eventlog.Event{
		created(1),
		edited(2, text("Sign in"), nil),
		edited(3, nil, text("")),
	}

	s := state.Replay(events)

	checkDetail(t, s, "Sign in", "")
	checkKnown(t, s, specDraft)
}

// Test 5: an event this version does not know makes the position
// unknown, with the line it was read from, its name, and the reason. The
// line is the line in the file, not the seq.
func TestReplayUnknownEventMakesThePositionUnknown(t *testing.T) {
	submitted := unknown(3, "spec.submitted")
	submitted.Seq = 2 // a damaged line 2 was skipped
	events := []eventlog.Event{created(1), submitted}

	s := state.Replay(events)

	checkUnknown(t, s,
		state.Unknown{Line: 3, Event: "spec.submitted", Reason: state.ReasonUnknownEvent},
		specDraft)
}

// Test 6: after an unknown event, the title and description still follow
// every slice.edited, and the position stays unknown at the first
// problem.
func TestReplayDetailGoesOnAfterAnUnknownEvent(t *testing.T) {
	events := []eventlog.Event{
		created(1),
		unknown(2, "spec.submitted"),
		edited(3, text("Sign in"), nil),
	}

	s := state.Replay(events)

	checkUnknown(t, s,
		state.Unknown{Line: 2, Event: "spec.submitted", Reason: state.ReasonUnknownEvent},
		specDraft)
	checkDetail(t, s, "Sign in", "Lets a user log in.")
}

// Test 7: slice.created is allowed only as the first event; a second one
// is not allowed in spec (draft).
func TestReplaySecondCreatedIsNotAllowed(t *testing.T) {
	events := []eventlog.Event{
		created(1),
		created(2),
	}

	s := state.Replay(events)

	checkUnknown(t, s,
		state.Unknown{Line: 2, Event: eventlog.SliceCreatedEvent, Reason: state.ReasonNotAllowed},
		specDraft)
}

// Test 8: only the first problem is kept: a later unknown event does not
// replace it.
func TestReplayKeepsTheFirstProblem(t *testing.T) {
	events := []eventlog.Event{
		created(1),
		created(2),
		unknown(3, "spec.submitted"),
	}

	s := state.Replay(events)

	checkUnknown(t, s,
		state.Unknown{Line: 2, Event: eventlog.SliceCreatedEvent, Reason: state.ReasonNotAllowed},
		specDraft)
}

// Test 9: no events give the empty state: no position, no detail, and a
// position that is known.
func TestReplayNoEvents(t *testing.T) {
	s := state.Replay(nil)

	checkKnown(t, s, state.Position{})
	checkDetail(t, s, "", "")
}
