//go:build unix

package filelock

import (
	"errors"
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

// tryLock opens the file at path, making it when it is missing, and takes
// its lock with flock, without waiting. When another process holds the
// lock, it returns ErrBusy.
func tryLock(path string) (*Lock, error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o644)

	lock := Lock{file: f}

	if err != nil {
		return nil, err
	}

	if err := unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		f.Close()
		if errors.Is(err, unix.EWOULDBLOCK) {
			return nil, ErrBusy
		}
		return nil, fmt.Errorf("lock %s: %w", path, err)
	}

	return &lock, nil

}

// tryUnLock releases the flock lock of lock's file. It does not close the
// file: Release does.
func tryUnLock(lock *Lock) error {
	err := unix.Flock(int(lock.file.Fd()), unix.LOCK_UN)
	if err != nil {
		return fmt.Errorf("unlock %s: %w", lock.file.Name(), err)
	}

	return nil
}
