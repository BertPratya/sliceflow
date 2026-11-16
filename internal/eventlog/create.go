package eventlog

import (
	"context"
	"os"
)

// Create makes a new log at logPath, and writes its first event,
// slice.created, with seq 1. Only Create makes a log, so the first line
// of a log is always slice.created. When the file already exists, it
// returns an error that matches fs.ErrExist, and changes nothing. When
// the event cannot be written, it removes the file it made, so no empty
// log is left behind.
func Create(ctx context.Context, logPath string, data SliceCreated) error {
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}

	// The file is closed before Append and Remove, because Windows
	// cannot remove a file that is still open. Remove's own error is not
	// returned: the error that made Create fail tells the caller more.
	if err := f.Close(); err != nil {
		os.Remove(logPath)
		return err
	}

	if err := Append(ctx, logPath, SliceCreatedEvent, data); err != nil {
		os.Remove(logPath)
		return err
	}

	return nil
}
