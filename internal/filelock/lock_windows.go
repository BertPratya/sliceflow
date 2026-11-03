package filelock

import (
	"errors"
	"fmt"
	"os"

	"golang.org/x/sys/windows"
)

// tryLock opens the file at path, making it when it is missing, and takes
// its lock with LockFileEx, without waiting. When another process holds
// the lock, it returns ErrBusy. On any error, it closes the file.
//
// Windows locks a range of bytes, not a whole file. tryLock locks the
// first byte, from offset 0, which stands for the whole file. Locking past
// the end of the file is allowed, so this works on an empty file too.
func tryLock(path string) (*Lock, error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return nil, err
	}

	// A new Overlapped is all zeros, so the range starts at offset 0.
	ol := new(windows.Overlapped)
	flags := uint32(windows.LOCKFILE_EXCLUSIVE_LOCK | windows.LOCKFILE_FAIL_IMMEDIATELY)
	if err := windows.LockFileEx(windows.Handle(f.Fd()), flags, 0, 1, 0, ol); err != nil {
		f.Close()
		if errors.Is(err, windows.ERROR_LOCK_VIOLATION) {
			return nil, ErrBusy
		}
		return nil, fmt.Errorf("lock %s: %w", path, err)
	}

	return &Lock{file: f}, nil
}

// tryUnLock releases the lock that tryLock took: the same first byte, from
// offset 0. It does not close the file: Release does.
func tryUnLock(lock *Lock) error {
	ol := new(windows.Overlapped)
	if err := windows.UnlockFileEx(windows.Handle(lock.file.Fd()), 0, 1, 0, ol); err != nil {
		return fmt.Errorf("unlock %s: %w", lock.file.Name(), err)
	}
	return nil
}
