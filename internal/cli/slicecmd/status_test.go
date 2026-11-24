package slicecmd

import (
	"testing"

	"github.com/BertPratya/sliceflow/internal/eventlog"
	"github.com/BertPratya/sliceflow/internal/slice"
)

// statusText shows "(none)" for an empty description, the time in UTC to
// the minute, the base branch only, and the number of events with the
// last one's name.
func TestStatusText(t *testing.T) {
	status := slice.StatusResult{
		Name: "card-payment",
		Created: eventlog.SliceCreated{
			Title: "Pay with a card", Base: "main",
			BaseCommit: "3f9a1c2d4e5f60718293a4b5c6d7e8f901234567",
		},
		CreatedAt: "2026-11-02T09:15:30Z",
		Events:    3,
		LastEvent: "plan.approved",
	}

	want := "card-payment: Pay with a card\n" +
		"  Description:  (none)\n" +
		"  Created:      2026-11-02 09:15 UTC\n" +
		"  Base:         main\n" +
		"  Events:       3 (last: plan.approved)\n"
	if got := statusText(status); got != want {
		t.Errorf("statusText gave:\n%s\nwant:\n%s", got, want)
	}
}
