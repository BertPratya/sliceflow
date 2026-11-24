package state

import (
	"slices"

	"github.com/BertPratya/sliceflow/internal/eventlog"
)

// updatePosition moves s by the event e, as transitions says. Once the
// position is unknown, it changes nothing more. An event that is
// unknown, or not allowed in the position s is in, leaves the position
// as it is and sets s.Unknown.
func updatePosition(s *State, e eventlog.Event) {
	if s.Unknown != nil {
		return
	}
	if slices.Contains(allowedEverywhere, e.Event) {
		return
	}

	for _, t := range transitions {
		if t.From == s.Position && t.Event == e.Event && (t.Guard == nil || t.Guard(*s)) {
			s.Position = t.To
			return
		}
	}

	reason := ReasonNotAllowed
	if !isKnown(e.Event) {
		reason = ReasonUnknownEvent
	}
	s.Unknown = &Unknown{Line: e.Line, Event: e.Event, Reason: reason}
}

// updateSliceDetail sets the title and description from slice.created,
// and changes each field that a slice.edited holds. It ignores every
// other event.
func updateSliceDetail(s *State, e eventlog.Event) {
	switch data := e.Payload.(type) {
	case eventlog.SliceCreated:
		s.Detail.Title = data.Title
		s.Detail.Description = data.Description

	case eventlog.SliceEdited:
		if data.Title != nil {
			s.Detail.Title = *data.Title
		}

		if data.Description != nil {
			s.Detail.Description = *data.Description
		}
	}
}
