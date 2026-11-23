package cli

import (
	"strings"
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

// printBadLines names one line, or several, or the first 10 and how many
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
			printBadLines(&out, path, tt.lines)
			if out.String() != tt.want {
				t.Errorf("printBadLines printed %q, want %q", out.String(), tt.want)
			}
		})
	}
}
