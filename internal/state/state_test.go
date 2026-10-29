package state

import (
	"testing"

	"github.com/BertPratya/sliceflow/internal/eventlog"
)

// events makes a list of events with the given names, for tests.
func events(names ...string) []eventlog.Event {
	var es []eventlog.Event
	for i, name := range names {
		es = append(es, eventlog.Event{Version: 1, Seq: i + 1, Event: name})
	}
	return es
}

// TestReplay checks where a slice ends up after its events (R23).
func TestReplay(t *testing.T) {
	tests := []struct {
		name   string
		events []eventlog.Event
		want   Place
	}{
		{
			name:   "slice.created goes to planning, spec_draft",
			events: events(eventlog.EventSliceCreated),
			want:   Place{Phase: "planning", State: "spec_draft"},
		},
		{
			name:   "no events stays unknown",
			events: nil,
			want:   Unknown,
		},
		{
			name:   "an event with no move stays unknown",
			events: events("something.else"),
			want:   Unknown,
		},
		{
			name:   "a second slice.created changes nothing",
			events: events(eventlog.EventSliceCreated, eventlog.EventSliceCreated),
			want:   Place{Phase: "planning", State: "spec_draft"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Replay(tt.events); got != tt.want {
				t.Errorf("Replay() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
