package eventlog

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/BertPratya/sliceflow/internal/filelock"
)

// eventToWrite is a line as Append writes it. It is Event with Data as
// any, so the caller's own value is turned into JSON by the same encoder
// as the rest of the line. The fields are in the order of the keys in
// the line.
type eventToWrite struct {
	Version int    `json:"version"`
	Seq     int    `json:"seq"`
	Event   string `json:"event"`
	At      string `json:"at"`
	Data    any    `json:"data"`
}

// needsNewline reports whether the log in f must get a "\n" before a new
// line: when it is not empty, and its last byte is not "\n". Then a
// cut-off last line stays a line of its own. A file that ends with
// "\r\n" also ends with "\n".
func needsNewline(f *os.File) (bool, error) {
	info, err := f.Stat()
	if err != nil {
		return false, err
	}
	if info.Size() == 0 {
		return false, nil
	}

	last := make([]byte, 1)
	if _, err := f.ReadAt(last, info.Size()-1); err != nil {
		return false, err
	}
	return last[0] != '\n', nil
}

// Append adds one event to the end of the log at logPath, with the name
// event and the data data, as one line of JSON. The log must exist:
// Append never creates it (see Create), and never changes a line that is
// already in it.
//
// Two processes can append to the same log at the same time: Append
// holds the lock on logPath + ".lock" while it works, and waits for it up
// to 10 seconds, or until ctx is cancelled. Both events are kept, each
// whole and on its own line.
//
// The new event's seq is the highest seq that can be read in the log,
// plus one; lines that cannot be read do not count. Its time is now, in
// UTC, to the second. <, >, and & are written as they are. The line is
// written with one Write, so the log gets the whole line or nothing. When
// nothing else failed, an error from closing the log is returned: the
// line may not have been saved.
func Append(ctx context.Context, logPath string, event string, data any) (err error) {
	lock, err := filelock.Acquire(ctx, logPath+".lock", 10*time.Second, 50*time.Millisecond)
	if err != nil {
		return err
	}
	defer lock.Release()

	existingEvents, _, err := Read(logPath)
	if err != nil {
		return err
	}

	highest := 0
	for _, e := range existingEvents {
		highest = max(highest, e.Seq)
	}

	f, err := os.OpenFile(logPath, os.O_RDWR|os.O_APPEND, 0)
	if err != nil {
		return err
	}
	defer func() {
		cerr := f.Close()
		if err == nil {
			err = cerr
		}
	}()

	var buf bytes.Buffer
	newline, err := needsNewline(f)
	if err != nil {
		return err
	}
	if newline {
		buf.WriteByte('\n')
	}

	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	line := eventToWrite{
		Version: 1,
		Seq:     highest + 1,
		Event:   event,
		At:      time.Now().UTC().Format(timeLayout),
		Data:    data,
	}
	if err := encoder.Encode(line); err != nil {
		return err
	}

	_, err = f.Write(buf.Bytes())
	return err
}
