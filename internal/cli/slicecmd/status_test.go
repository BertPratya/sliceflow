package slicecmd

import (
	"testing"

	"github.com/BertPratya/sliceflow/internal/eventlog"
	"github.com/BertPratya/sliceflow/internal/slice"
	"github.com/BertPratya/sliceflow/internal/state"
)

// statusText shows the title from the state, the state, "(none)" for an
// empty description, the time in UTC to the minute, the base branch only,
// the number of events with the last one's name, and for spec (draft)
// the next step. An unknown state shows "unknown" and no next step.
func TestStatusText(t *testing.T) {
	known := slice.StatusResult{
		Name: "card-payment",
		Created: eventlog.SliceCreated{
			Title: "Pay with a card", Base: "main",
			BaseCommit: "3f9a1c2d4e5f60718293a4b5c6d7e8f901234567",
		},
		State: state.State{
			Position: state.Position{Phase: state.PhaseSpec, Step: state.StepDraft},
			Detail:   state.Detail{Title: "Pay by card"},
		},
		Spec:      "/home/you/shop/docs/slices/card-payment/spec.md",
		CreatedAt: "2026-11-02T09:15:30Z",
		Events:    3,
		LastEvent: "slice.edited",
	}
	unknown := known
	unknown.State.Unknown = &state.Unknown{Line: 3, Event: "plan.approved", Reason: state.ReasonUnknownEvent}

	tests := []struct {
		about  string
		status slice.StatusResult
		want   string
	}{
		{"a known state", known,
			"card-payment: Pay by card\n" +
				"  State:        spec (draft)\n" +
				"  Description:  (none)\n" +
				"  Created:      2026-11-02 09:15 UTC\n" +
				"  Base:         main\n" +
				"  Events:       3 (last: slice.edited)\n" +
				"\n" +
				"Next: write the spec in /home/you/shop/docs/slices/card-payment/spec.md\n" +
				"Then run: sliceflow spec validate card-payment\n"},
		{"an unknown state", unknown,
			"card-payment: Pay by card\n" +
				"  State:        unknown\n" +
				"  Description:  (none)\n" +
				"  Created:      2026-11-02 09:15 UTC\n" +
				"  Base:         main\n" +
				"  Events:       3 (last: slice.edited)\n"},
	}
	for _, tt := range tests {
		t.Run(tt.about, func(t *testing.T) {
			if got := statusText(tt.status); got != tt.want {
				t.Errorf("statusText gave:\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}
