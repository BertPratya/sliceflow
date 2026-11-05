// Package eventlog reads and writes a slice's event log, .events.jsonl:
// one JSON object per line, one line per event, oldest first.
package eventlog

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"time"
)

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

// rawEvent is a line as json.Unmarshal fills it. The fields are pointers,
// so a missing key stays nil and can be told apart from a zero value.
type rawEvent struct {
	Version *int            `json:"version"`
	Seq     *int            `json:"seq"`
	Event   *string         `json:"event"`
	At      *string         `json:"at"`
	Data    json.RawMessage `json:"data"`
}

// BadLine is a line that Read skipped: its number, counted from 1, and
// the reason it cannot be read.
type BadLine struct {
	Line int
	Err  error
}

// jsonlLineToEvent turns one line of the log into an Event. It returns
// the first problem it finds:
//  1. The line is not a JSON object with fields of the right types: the
//     error from json.Unmarshal.
//  2. version is missing, or is not 1.
//  3. seq is missing.
//  4. event is missing.
//  5. at is missing.
//  6. data is missing, or is not an object.
//  7. at is not written exactly like timeLayout: UTC, to the second.
//
// It does not check the fields inside data; see validateEvent.
func jsonlLineToEvent(line []byte) (Event, error) {
	var r rawEvent

	if err := json.Unmarshal(line, &r); err != nil {
		return Event{}, err
	}

	switch {
	case r.Version == nil:
		return Event{}, errors.New("version is missing")
	case *r.Version != 1:
		return Event{}, errors.New("version is not 1")
	case r.Seq == nil:
		return Event{}, errors.New("seq is missing")
	case r.Event == nil:
		return Event{}, errors.New("event is missing")
	case r.At == nil:
		return Event{}, errors.New("at is not a valid time")
	case len(r.Data) == 0 || r.Data[0] != '{' || !json.Valid(r.Data):
		return Event{}, errors.New("data is not an object")
	}

	// time.Parse alone accepts fractions of a second, like .123, even
	// though timeLayout has none. Writing the time back out and comparing
	// it with the text rejects anything not written exactly like
	// timeLayout.
	t, err := time.Parse(timeLayout, *r.At)
	if err != nil || t.Format(timeLayout) != *r.At {
		return Event{}, errors.New("at is not a valid time")
	}

	return Event{Version: *r.Version, Seq: *r.Seq, Event: *r.Event, At: *r.At, Data: r.Data}, nil

}

// validationRouter maps an event name to the function that checks its
// data. An event with no entry has no extra checks.
var validationRouter = map[string]func(json.RawMessage) error{
	"slice.created": validateSliceCreated,
}

// validateEvent checks the data of one event, with the function that
// validationRouter has for its name. An event name that is not in
// validationRouter is not an error: a newer sliceflow may have written
// it (R32).
func validateEvent(event string, data json.RawMessage) error {
	check, ok := validationRouter[event]
	if !ok {
		return nil
	}

	return check(data)

}

// Read reads the log at path. It returns the readable lines as events, in
// file order, and the lines it skipped as bad lines. Plain reads need no
// lock.
//
// Steps:
//  1. Open the file. If that fails, return the error, for example one
//     that matches fs.ErrNotExist.
//  2. Read it line by line, with lines of up to 4 MiB. "\r\n" counts as
//     a line end, and the text after the last "\n" is a line when it is
//     not empty.
//  3. Count each line from 1, and turn it into an event with
//     jsonlLineToEvent, then check its data with validateEvent. A line
//     that fails either one becomes a bad line, and reading goes on.
//  4. If reading stops early, for example on a line longer than 4 MiB,
//     return the error. The events are then incomplete, so a caller
//     that writes must not go on.
func Read(path string) ([]Event, []BadLine, error) {
	file, err := os.OpenFile(path, os.O_RDONLY, 0400)
	if err != nil {
		return []Event{}, []BadLine{}, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)

	var events []Event
	var badLines []BadLine
	var c int = 0

	for scanner.Scan() {
		c += 1
		line := scanner.Bytes()
		event, err := jsonlLineToEvent(line)

		if err != nil {
			badLine := BadLine{Line: c, Err: err}
			badLines = append(badLines, badLine)
			continue
		}

		if err = validateEvent(event.Event, event.Data); err != nil {
			badLine := BadLine{Line: c, Err: err}
			badLines = append(badLines, badLine)
			continue
		}

		events = append(events, event)
	}

	if err := scanner.Err(); err != nil {
		return []Event{}, []BadLine{}, err
	}

	return events, badLines, err

}
