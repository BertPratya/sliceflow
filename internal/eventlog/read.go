// Package eventlog reads and writes a slice's event log, .events.jsonl:
// one JSON object per line, one line per event, oldest first.
package eventlog

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os"
	"time"
)

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
// It does not check the fields inside data; see decodeEvent.
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

// decoders maps an event name to the function that checks its data and
// decodes it into the event's own type. An event with no entry is not
// decoded.
var decoders = map[string]func(json.RawMessage) (any, error){
	SliceCreatedEvent: decodeSliceCreated,
	SliceEditedEvent:  decodeSliceEdited,
}

// decodeEvent checks the data of one event and decodes it, with the
// function that decoders has for its name. For an event name that is
// not in decoders it returns nil and no error: a newer sliceflow may
// have written it.
func decodeEvent(event string, data json.RawMessage) (any, error) {
	decode, ok := decoders[event]
	if !ok {
		return nil, nil
	}
	return decode(data)
}

// Read reads the log at path. It returns the readable lines as events, in
// file order, and the lines it skipped as bad lines. Plain reads need no
// lock. When the file cannot be opened, it returns that error, for
// example one that matches fs.ErrNotExist. Otherwise it reads as ReadFrom
// does.
func Read(path string) ([]Event, []BadLine, error) {
	file, err := os.OpenFile(path, os.O_RDONLY, 0400)
	if err != nil {
		return []Event{}, []BadLine{}, err
	}
	defer file.Close()
	return ReadFrom(file)
}

// ReadFrom reads a log from r, e.g. a log that git keeps in a commit. It
// returns the readable lines as events, in order, and the lines it
// skipped as bad lines, each with its number, counted from 1. Lines may
// be up to 4 MiB long, and end with "\n" or "\r\n"; the text after the
// last "\n" is a line when it is not empty. A line is readable when it is
// an event of the right form, and its data fits its event. When reading
// stops early, e.g. on a line longer than 4 MiB, it returns the error:
// the events are then incomplete, so a caller that writes must not go on.
func ReadFrom(r io.Reader) ([]Event, []BadLine, error) {
	scanner := bufio.NewScanner(r)
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

		event.Line = c
		event.Payload, err = decodeEvent(event.Event, event.Data)
		if err != nil {
			badLine := BadLine{Line: c, Err: err}
			badLines = append(badLines, badLine)
			continue
		}

		events = append(events, event)
	}

	if err := scanner.Err(); err != nil {
		return []Event{}, []BadLine{}, err
	}

	return events, badLines, nil
}
