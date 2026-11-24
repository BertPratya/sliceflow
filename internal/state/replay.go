package state

import (
	"github.com/BertPratya/sliceflow/internal/eventlog"
)

// Replay works out a slice's state from its events, in order. It reads
// every event. The title and description follow every event, even after
// the position became unknown. The position follows transitions; at the
// first event that is unknown, or not allowed in the position the slice
// is in, the position stops changing, and State.Unknown says where and
// why.
func Replay(events []eventlog.Event) State {
	var s State

	for _, e := range events {
		updateSliceDetail(&s, e)
		updatePosition(&s, e)
	}

	return s
}
