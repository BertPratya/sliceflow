package eventlog

import (
	"context"
	"os"
)

// Create makes a new log at logPath, and writes its first event,
// slice.created, with seq 1 (R27). Only Create makes a log, so the first
// line of a log is always slice.created.
//
// Steps:
//  1. Make the file. If it already exists, return an error that matches
//     fs.ErrExist, and change nothing.
//  2. Close it, then add the event with Append.
//  3. If a step after 1 fails, remove the file, and return the error, so
//     no empty log is left behind.
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
