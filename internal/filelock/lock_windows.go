package filelock

import (
	"errors"
	"io/fs"
	"os"
)

// StaleLockBlocks tells whether a lock file left by a process that stopped,
// e.g. after a crash, blocks later locks. Here it does, until someone
// deletes the file: the lock is the file itself.
const StaleLockBlocks = true

func tryLock(path string) (*os.File, error) {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if errors.Is(err, fs.ErrExist) {
		return nil, ErrBusy
	}
	if err != nil {
		return nil, err
	}
	return f, nil
}

// tryUnlock closes the file and deletes it. If the close fails, e.g. on a
// second unlock, it does not delete: by then the file may be another
// process's lock.
func tryUnlock(f *os.File) error {
	if err := f.Close(); err != nil {
		return err
	}
	return os.Remove(f.Name())
}
