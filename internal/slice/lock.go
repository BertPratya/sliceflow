package slice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/BertPratya/sliceflow/internal/eventlog"
)

const (
	defaultLockWait = 10 * time.Second
	lockRetry       = 100 * time.Millisecond
)

// lockCreate and lockRemove are variables so tests can fake the Windows
// failures (sharing violation, access denied) on every system.
var (
	lockCreate = os.OpenFile
	lockRemove = os.Remove
)

const (
	freeTries = 50
	freePause = 10 * time.Millisecond
)

// freeLock removes the lock file. On Windows the remove fails while another
// process has the file open, so it tries again for a short time. A missing
// file counts as freed.
func freeLock(path string) {
	for i := 0; i < freeTries; i++ {
		err := lockRemove(path)
		if err == nil || errors.Is(err, fs.ErrNotExist) {
			return
		}
		if _, serr := os.Stat(path); errors.Is(serr, fs.ErrNotExist) {
			return
		}
		time.Sleep(freePause)
	}
}

type lockContent struct {
	Pid   int    `json:"pid"`
	Since string `json:"since"`
}

// takeLock takes the lock of one slice: it creates
// .sliceflow/locks/<name>.lock in the main checkout, only if it does not
// exist, and tries again every 100 ms for up to r.LockWait. The returned
// function removes the lock.
func takeLock(ctx context.Context, r Repo, name string) (func(), error) {
	sliceflowDir := filepath.Join(filepath.FromSlash(r.Root), ".sliceflow")
	locksDir := filepath.Join(sliceflowDir, "locks")
	if err := os.MkdirAll(locksDir, 0o755); err != nil {
		return nil, err
	}
	if err := ensureIgnore(filepath.Join(sliceflowDir, ".gitignore")); err != nil {
		return nil, err
	}
	path := filepath.Join(locksDir, name+".lock")
	line, err := json.Marshal(lockContent{Pid: os.Getpid(), Since: r.Now().UTC().Format(eventlog.TimeLayout)})
	if err != nil {
		return nil, err
	}
	line = append(line, '\n')

	wait := r.LockWait
	if wait == 0 {
		wait = defaultLockWait
	}
	deadline := time.Now().Add(wait)
	for {
		f, err := lockCreate(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err == nil {
			_, werr := f.Write(line)
			if cerr := f.Close(); werr == nil {
				werr = cerr
			}
			if werr != nil {
				lockRemove(path)
				return nil, werr
			}
			return func() { freeLock(path) }, nil
		}
		// On Windows a file that is being deleted can give "access denied".
		// That counts as busy, until the wait is over.
		other := !errors.Is(err, fs.ErrExist)
		if other && !errors.Is(err, fs.ErrPermission) {
			if _, serr := os.Stat(path); serr != nil {
				return nil, err
			}
		}
		if !time.Now().Before(deadline) {
			if other {
				return nil, err
			}
			return nil, busyError(name, path)
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(lockRetry):
		}
	}
}

func ensureIgnore(path string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if errors.Is(err, fs.ErrExist) {
		return nil
	}
	if err != nil {
		return err
	}
	_, werr := f.WriteString("*\n")
	if cerr := f.Close(); werr == nil {
		werr = cerr
	}
	return werr
}

func busyError(name, path string) error {
	pid, since := "unknown", "unknown"
	if b, err := os.ReadFile(path); err == nil {
		var c lockContent
		if json.Unmarshal(b, &c) == nil && c.Pid != 0 {
			if t, err := time.Parse(eventlog.TimeLayout, c.Since); err == nil {
				pid, since = fmt.Sprint(c.Pid), t.UTC().Format("15:04")
			}
		}
	}
	return &Error{Message: fmt.Sprintf("slice %s is busy: another sliceflow command holds .sliceflow/locks/%s.lock (pid %s, since %s UTC). If no sliceflow command is running, delete that file.", name, name, pid, since)}
}
