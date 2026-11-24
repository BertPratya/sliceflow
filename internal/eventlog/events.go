package eventlog

import "encoding/json"

// SliceCreated is the data of the slice.created event. Every field
// is always written, even when it is empty, so a slice.created line
// always has the five keys that Read checks.
type SliceCreated struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Base        string `json:"base"`
	BaseCommit  string `json:"base_commit"`
}

// SliceEdited is the data of the slice.edited event. It holds only the
// fields that changed: a field is nil when the event did not change it.
// A Description that points to "" means the slice now has no
// description. A nil field is not written.
type SliceEdited struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
}

// timeLayout is how at is written: UTC, to the second, for example
// 2026-11-02T09:15:30Z.
const timeLayout = "2006-01-02T15:04:05Z"

// Event is one readable line of the log. Read returns an Event only
// after every check has passed, so every field is set.
type Event struct {
	Version int             `json:"version"`
	Seq     int             `json:"seq"`
	Event   string          `json:"event"`
	At      string          `json:"at"`
	Data    json.RawMessage `json:"data"`

	// Line is the number of the line the event was read from, counted
	// from 1. It can differ from Seq when lines before it were skipped.
	Line int `json:"-"`

	// Payload is Data decoded into its event's own type, by value: a
	// SliceCreated for slice.created, a SliceEdited for slice.edited.
	// It is nil for an event this version of sliceflow does not know.
	Payload any `json:"-"`
}

// The names of the events that sliceflow writes.
const (
	// SliceCreatedEvent is the first event of every log, written by
	// Create.
	SliceCreatedEvent = "slice.created"

	// SliceEditedEvent changes a slice's title, its description, or
	// both.
	SliceEditedEvent = "slice.edited"
)
