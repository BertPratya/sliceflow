// Package eventlog writes and reads the event log of a slice (R19).
package eventlog

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// Version is the version of the event format. Every event carries it.
const Version = 1

// EventSliceCreated is the name of the first event of a slice.
const EventSliceCreated = "slice.created"

// atLayout is how an event's time is written: UTC, to the second, with a
// Z at the end.
const atLayout = "2006-01-02T15:04:05Z"

// Event is one entry of the event log. Data holds the event's own fields
// as ready-made JSON.
type Event struct {
	Version int
	Seq     int
	Event   string
	At      time.Time
	Data    json.RawMessage
}

// Ref names a place in git: a branch and the full id of a commit on it.
type Ref struct {
	Branch string `json:"branch"`
	Commit string `json:"commit"`
}

// SliceCreated is the data of a slice.created event (R20). Title is left
// out when it is empty.
type SliceCreated struct {
	From  Ref    `json:"from"`
	Title string `json:"title,omitempty"`
}

// line is the shape of one log line. The field order here is the order in
// the file (R19). At is a string so we control its format.
type line struct {
	Version int             `json:"version"`
	Seq     int             `json:"seq"`
	Event   string          `json:"event"`
	At      string          `json:"at"`
	Data    json.RawMessage `json:"data"`
}

// Data turns a value into JSON for the Data field of an Event. The value
// must encode to a JSON object, for example a SliceCreated. It does not
// escape &, < or >, and it adds no newline at the end.
func Data(v any) (json.RawMessage, error) {
	b, err := marshal(v)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(bytes.TrimSuffix(b, []byte("\n"))), nil
}

// Line returns the event as one line of the log: JSON with the fields in
// R19's order, and one \n at the end. The time is written in UTC, to the
// second. It returns an error if Version is not 1, or if Data is not a
// JSON object. It does not check Seq: the caller picks it, and reading
// the log checks it (R27).
func Line(e Event) ([]byte, error) {
	if e.Version != Version {
		return nil, fmt.Errorf("event version is %d, want %d", e.Version, Version)
	}
	trimmed := bytes.TrimSpace(e.Data)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return nil, errors.New("event data must be a JSON object")
	}
	return marshal(line{
		Version: e.Version,
		Seq:     e.Seq,
		Event:   e.Event,
		At:      e.At.UTC().Format(atLayout),
		Data:    e.Data,
	})
}

// marshal encodes v as JSON without HTML escaping. The encoder adds one
// \n at the end, which Line keeps and Data removes.
func marshal(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
