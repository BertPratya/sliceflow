package filelock

import (
	"errors"
	"os"
	"time"
)

// ErrBusy means another process holds the lock.
var ErrBusy = errors.New("the lock is held by another process")
var reTryAgainTime = 50 * time.Millisecond
var limitRetryTime = 10 * time.Second

// Lock takes the lock at path, and makes the file if it is missing. While
// another process holds it, Lock tries again until the deadline, then
// returns ErrBusy. Keep the file open until Unlock: closing it frees the lock.
func Lock(path string) (*os.File, error) {
	start := time.Now()
	for time.Since(start) < limitRetryTime {
		f, err := tryLock(path)

		if err == nil {
			return f, nil
		}

		if !errors.Is(err, ErrBusy) {
			return nil, err
		}

		time.Sleep(reTryAgainTime)

	}

	return nil, ErrBusy

}

// Unlock frees a lock that Lock took.
func Unlock(f *os.File) error {
	err := tryUnlock(f)
	return err
}
