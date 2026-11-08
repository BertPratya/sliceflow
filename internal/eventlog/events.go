package eventlog

import "encoding/json"

// SliceCreated is the data of the slice.created event (R27). Every field
// is always written, even when it is empty, so a slice.created line
// always has the five keys that Read checks.
type SliceCreated struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Base        string `json:"base"`
	BaseCommit  string `json:"base_commit"`
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
}

// The names of the events that sliceflow writes.
const (
	// SliceCreatedEvent is the first event of every log, written by
	// Create.
	SliceCreatedEvent = "slice.created"
)
