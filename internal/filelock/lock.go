// Package filelock makes sure that only one process at a time does a piece
// of work, such as adding a line to a file. It uses the lock of the
// operating system on a separate lock file: flock on Linux and macOS, and
// LockFileEx on Windows. The operating system releases the lock when the
// process ends, even after a crash, so a lock is never left behind.
package filelock

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"
)

// Lock is a lock held by this process. Get one with Acquire, and give it
// back with Release.
type Lock struct{ file *os.File }

// ErrBusy means that another process held the lock for the whole wait
// time. Check for it with errors.Is.
var ErrBusy = errors.New("held by another process")

// Acquire takes the lock on the file at path, and returns it. It makes the
// file when it is missing, but not its folder. It never deletes the file.
//
// When another process holds the lock, Acquire tries again every interval,
// until timeout has passed, and then returns an error that wraps ErrBusy.
// An interval of 0 or less means: try again with no pause. The timeout
// must be more than 0: with 0 or less, Acquire makes no try at all, and
// returns ErrBusy even when the lock is free. Any other error is returned
// at once. When ctx is cancelled while Acquire waits, it
// returns ctx.Err().
//
// Acquire does not know about locks held by this same process. Asking for
// a lock that this process already holds waits until timeout, and returns
// ErrBusy.
func Acquire(ctx context.Context, path string, timeout, interval time.Duration) (*Lock, error) {

	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {

		lock, err := tryLock(path)

		if err == nil {
			return lock, nil
		} else if !errors.Is(err, ErrBusy) {
			return nil, err
		}

		select {

		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(interval):
		}

	}

	return nil, fmt.Errorf("lock %s: %w", path, ErrBusy)

}

// Release gives the lock back, and closes the lock file. Calling Release a
// second time returns an error.
func (lock *Lock) Release() error {
	defer lock.file.Close()
	if err := tryUnLock(lock); err != nil {
		return err
	}
	return nil

}
