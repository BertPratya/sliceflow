package command

import (
	"strings"
	"testing"

	"github.com/BertPratya/sliceflow/internal/state"
)

// PrintBadLines names one line, or several, or the first 10 and how many
// more there are.
func TestPrintBadLines(t *testing.T) {
	const path = "docs/slices/card-payment/.events.jsonl"
	tests := []struct {
		about string
		lines []int
		want  string
	}{
		{"none", nil, ""},
		{"one", []int{2}, "Warning: line 2 of " + path + " could not be read, so it was skipped.\n"},
		{"two", []int{2, 3}, "Warning: lines 2, 3 of " + path + " could not be read, so they were skipped.\n"},
		{"ten", []int{2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
			"Warning: lines 2, 3, 4, 5, 6, 7, 8, 9, 10, 11 of " + path + " could not be read, so they were skipped.\n"},
		{"twelve", []int{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13},
			"Warning: lines 2, 3, 4, 5, 6, 7, 8, 9, 10, 11 (2 more not listed) of " + path + " could not be read, so they were skipped.\n"},
	}
	for _, tt := range tests {
		t.Run(tt.about, func(t *testing.T) {
			var out strings.Builder
			PrintBadLines(&out, path, tt.lines)
			if out.String() != tt.want {
				t.Errorf("PrintBadLines printed %q, want %q", out.String(), tt.want)
			}
		})
	}
}

// PrintStateWarning prints nothing for a known state, and the warning of
// the reason for an unknown one, with the last known position for an
// event that is not allowed.
func TestPrintStateWarning(t *testing.T) {
	const path = "docs/slices/login/.events.jsonl"
	draft := state.Position{Phase: state.PhaseSpec, Step: state.StepDraft}
	tests := []struct {
		about string
		state state.State
		want  string
	}{
		{"known", state.State{Position: draft}, ""},
		{"unknown event", state.State{Position: draft,
			Unknown: &state.Unknown{Line: 2, Event: "spec.submitted", Reason: state.ReasonUnknownEvent}},
			"Warning: line 2 of " + path + " has the event spec.submitted, which this version of sliceflow does not know, so the state cannot be worked out.\n"},
		{"not allowed", state.State{Position: draft,
			Unknown: &state.Unknown{Line: 2, Event: "slice.created", Reason: state.ReasonNotAllowed}},
			"Warning: line 2 of " + path + " has the event slice.created, which is not allowed while the slice is in spec (draft), so the state cannot be worked out.\n"},
	}
	for _, tt := range tests {
		t.Run(tt.about, func(t *testing.T) {
			var out strings.Builder
			PrintStateWarning(&out, path, tt.state)
			if out.String() != tt.want {
				t.Errorf("PrintStateWarning printed %q, want %q", out.String(), tt.want)
			}
		})
	}
}
