package state

import (
	"slices"
	"testing"

	"github.com/BertPratya/sliceflow/internal/eventlog"
)

// events makes a list of events with the given names, for tests. Each one
// is on its own line, from line 1.
func events(names ...string) []eventlog.Event {
	var es []eventlog.Event
	for i, name := range names {
		es = append(es, eventlog.Event{Version: 1, Seq: i + 1, Event: name, Line: i + 1})
	}
	return es
}

// TestReplay checks where a slice ends up after its events (R23), and the
// warning for an event with no move (R27).
func TestReplay(t *testing.T) {
	tests := []struct {
		name     string
		events   []eventlog.Event
		want     Place
		warnings []string
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
			warnings: []string{
				".events.jsonl line 1: something.else is not allowed in unknown; skipped",
			},
		},
		{
			name:   "a second slice.created changes nothing",
			events: events(eventlog.EventSliceCreated, eventlog.EventSliceCreated),
			want:   Place{Phase: "planning", State: "spec_draft"},
			warnings: []string{
				".events.jsonl line 2: slice.created is not allowed in planning (spec_draft); skipped",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, warnings := Replay(tt.events)
			if got != tt.want {
				t.Errorf("Replay() = %+v, want %+v", got, tt.want)
			}
			if !slices.Equal(warnings, tt.warnings) {
				t.Errorf("warnings = %q, want %q", warnings, tt.warnings)
			}
		})
	}
}
