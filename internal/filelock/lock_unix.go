//go:build linux || darwin

package filelock

import (
	"errors"
	"os"
	"syscall"
)

// StaleLockBlocks tells whether a lock file left by a process that stopped,
// e.g. after a crash, blocks later locks. Here it never does: the system
// frees the lock when the process ends.
const StaleLockBlocks = false

func tryLock(path string) (*os.File, error) {

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return nil, err
	}

	err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)

	if err != nil {
		f.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) {
			return nil, ErrBusy
		} else {
			return nil, err
		}
	}

	return f, nil

}

// tryUnlock frees the lock and closes the file. It keeps the file: deleting
// it would let two processes hold the lock at once.
func tryUnlock(f *os.File) error {
	err := syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	closeErr := f.Close()
	if err != nil {
		return err
	}
	return closeErr

}
