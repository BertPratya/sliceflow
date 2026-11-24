package command

import (
	"strings"
	"testing"
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
