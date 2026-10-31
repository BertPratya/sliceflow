package eventlog

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/BertPratya/sliceflow/internal/filelock"
	"github.com/BertPratya/sliceflow/internal/git"
)

// logName is the log's file name inside a slice folder.
const logName = ".events.jsonl"

// ErrNotFound means the slice folder has no log. Nothing was changed.
var ErrNotFound = errors.New(NotFound)

// WriteError means AddEvent failed after it started to change the log.
// Restored tells whether the log was put back as it was (R35).
type WriteError struct {
	Err      error
	Restored bool
}

func (e *WriteError) Error() string { return e.Err.Error() }
func (e *WriteError) Unwrap() error { return e.Err }

// AddEvent adds e to the end of the log in sliceDir, and commits only the
// log, with msg (R29). It sets e.Seq itself: 1 more than the seq of the
// last good line, or 1 when there is none (R33). It holds the log's lock
// the whole time (R34).
//
// The warnings have no slice name; the caller adds it. The errors:
// filelock.ErrBusy and ErrNotFound change nothing; a *WriteError comes
// after the log was changed.
func AddEvent(sliceDir string, e Event, msg string) (warnings []string, err error) {
	logPath := filepath.Join(sliceDir, logName)

	// TODO: the .git/info/exclude lines (R34), before the lock.
	lock, err := filelock.Lock(filepath.Join(sliceDir, ".events.lock"))
	if err != nil {
		return nil, err
	}
	defer func() {
		// The event is already committed here, so a failed unlock is a
		// warning, not an error.
		if err := filelock.Unlock(lock); err != nil {
			warnings = append(warnings, fmt.Sprintf("could not free the lock %s: %v", lock.Name(), err))
		}
	}()

	original, err := os.ReadFile(logPath)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	good, warning := cutHalfLine(original)
	if warning != "" {
		warnings = append(warnings, warning)
	}
	e.Seq = nextSeq(good)
	line, err := Line(e)
	if err != nil {
		return warnings, err
	}

	// From here on the log changes, so every failure puts it back.
	if err := writeLog(logPath, len(good), len(original), line); err != nil {
		return warnings, putBack(logPath, original, err)
	}
	if err := git.CommitPaths(sliceDir, msg, logName); err != nil {
		return warnings, putBack(logPath, original, err)
	}
	return warnings, nil
}

// cutHalfLine returns the content up to and with its last \n, and the R33
// warning when it cut a half-written last line.
func cutHalfLine(content []byte) (good []byte, warning string) {
	if len(content) == 0 || bytes.HasSuffix(content, []byte("\n")) {
		return content, ""
	}
	good = content[:bytes.LastIndexByte(content, '\n')+1]
	n := bytes.Count(good, []byte("\n")) + 1
	return good, fmt.Sprintf("%s line %d: half-written last line; removed", logName, n)
}

// nextSeq returns 1 more than the seq of the last good line, or 1.
func nextSeq(content []byte) int {
	events := Parse(content).Events
	if len(events) == 0 {
		return 1
	}
	return events[len(events)-1].Seq + 1
}

// writeLog cuts the log to keep bytes when it is longer, then adds line at
// the end. It never makes the log.
func writeLog(path string, keep, size int, line []byte) error {
	if keep < size {
		if err := os.Truncate(path, int64(keep)); err != nil {
			return err
		}
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0)
	if err != nil {
		return err
	}
	if _, err := f.Write(line); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// putBack writes original back to the log, and returns cause as a
// *WriteError that tells whether that worked.
func putBack(path string, original []byte, cause error) error {
	if err := restore(path, original); err != nil {
		return &WriteError{Err: errors.Join(cause, err), Restored: false}
	}
	return &WriteError{Err: cause, Restored: true}
}

// restore writes content to a temp file beside path, then renames it onto
// path. The log is never empty or half-written on the way (R35).
func restore(path string, content []byte) error {
	tmpPath := path + ".tmp"
	t, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	if _, err := t.Write(content); err != nil {
		t.Close()
		return err
	}
	if err := t.Sync(); err != nil {
		t.Close()
		return err
	}
	if err := t.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}
