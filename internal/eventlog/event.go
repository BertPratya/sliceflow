// Package eventlog writes and reads the slice event log. It does no file
// or git access.
package eventlog

import (
	"bytes"
	"encoding/json"
	"time"
)

// TimeLayout is the time format of the "at" field of an event (UTC, whole seconds).
const TimeLayout = "2006-01-02T15:04:05Z"

// Event is one line of the log. Fields are declared in R19's order, so
// encoding/json writes them in that order.
type Event struct {
	Version int    `json:"version"`
	Seq     int    `json:"seq"`
	Event   string `json:"event"`
	At      string `json:"at"`
	Data    any    `json:"data"`
	// Line is the line number in the log, set by Parse. Never written.
	Line int `json:"-"`
}

// Titled is the data of a slice.titled event.
type Titled struct {
	Title string `json:"title"`
}

// Submitted is the data of a spec.submitted event.
type Submitted struct {
	Result   string `json:"result"`
	Errors   int    `json:"errors"`
	Warnings int    `json:"warnings"`
}

// ReviewStarted is the data of a review.started event.
type ReviewStarted struct {
	Doc      string `json:"doc"`
	Round    int    `json:"round"`
	Reviewer string `json:"reviewer"`
	File     string `json:"file"`
	Commit   string `json:"commit"`
}

// ReviewFinished is the data of a review.finished event.
type ReviewFinished struct {
	Doc    string `json:"doc"`
	Round  int    `json:"round"`
	Result string `json:"result"`
}

// New returns an event with version 1 and the time in UTC.
func New(seq int, at time.Time, event string, data any) Event {
	return Event{Version: 1, Seq: seq, Event: event, At: at.UTC().Format(TimeLayout), Data: data}
}

// Created is the data of a slice.created event.
type Created struct {
	From  From   `json:"from"`
	Title string `json:"title,omitempty"`
}

// From is the branch and commit a slice started from.
type From struct {
	Branch string `json:"branch"`
	Commit string `json:"commit"`
}

// NewCreated returns the first event of a slice.
func NewCreated(at time.Time, from From, title string) Event {
	return Event{
		Version: 1,
		Seq:     1,
		Event:   "slice.created",
		At:      at.UTC().Format(TimeLayout),
		Data:    Created{From: from, Title: title},
	}
}

// Encode returns one line of JSON, ending in "\n".
func Encode(e Event) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(e); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
