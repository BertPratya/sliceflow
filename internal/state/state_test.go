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
			name:   "slice.titled stays in planning, spec_draft (R32)",
			events: events(eventlog.EventSliceCreated, eventlog.EventSliceTitled, eventlog.EventSliceTitled),
			want:   Place{Phase: "planning", State: "spec_draft"},
		},
		{
			name:   "slice.titled before slice.created is skipped",
			events: events(eventlog.EventSliceTitled, eventlog.EventSliceCreated),
			want:   Place{Phase: "planning", State: "spec_draft"},
			warnings: []string{
				".events.jsonl line 1: slice.titled is not allowed in unknown; skipped",
			},
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

// TestNext checks the next step for each place (0002 R25).
func TestNext(t *testing.T) {
	tests := []struct {
		name  string
		place Place
		want  string
	}{
		{"planning, spec_draft", Place{Phase: "planning", State: "spec_draft"}, "write the spec, then check it: sliceflow spec validate spec-submit"},
		{"unknown", Unknown, ""},
		{"planning with no state", Place{Phase: "planning"}, ""},
		{"a place with no move", Place{Phase: "build", State: "coding"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Next(tt.place, "spec-submit"); got != tt.want {
				t.Errorf("Next(%+v) = %q, want %q", tt.place, got, tt.want)
			}
		})
	}
}
