package state

import (
	"fmt"

	"github.com/BertPratya/sliceflow/internal/eventlog"
)

// Replay applies each allowed event in order. Each other event is skipped,
// with R4's warning.
func Replay(events []eventlog.Event) (State, []string) {
	s := State{Phase: "unknown"}
	var warnings []string
	for _, e := range events {
		if reason := Allowed(s, e); reason != "" {
			warnings = append(warnings, fmt.Sprintf(".events.jsonl line %d: %s; skipped", e.Line, reason))
			continue
		}
		s = apply(s, e)
	}
	return s, warnings
}
