package filelock

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// shortTimes makes the waits short for one test, and puts them back after.
func shortTimes(t *testing.T) {
	t.Helper()
	oldLimit, oldRetry := limitRetryTime, reTryAgainTime
	limitRetryTime, reTryAgainTime = 200*time.Millisecond, 10*time.Millisecond
	t.Cleanup(func() { limitRetryTime, reTryAgainTime = oldLimit, oldRetry })
}

// result is what one call to Lock gave back, and how long it took.
type result struct {
	f    *os.File
	err  error
	took time.Duration
}

// lockWithin calls Lock in a goroutine, and fails the test if Lock has not
// returned after max. So a Lock that waits forever fails the test instead
// of hanging it.
func lockWithin(t *testing.T, path string, max time.Duration) result {
	t.Helper()
	done := make(chan result, 1)
	go func() {
		start := time.Now()
		f, err := Lock(path)
		done <- result{f, err, time.Since(start)}
	}()
	select {
	case r := <-done:
		return r
	case <-time.After(max):
		t.Fatalf("Lock(%q) did not return within %v", path, max)
		return result{}
	}
}

// mustLock takes the lock, and fails the test if it cannot.
func mustLock(t *testing.T, path string) *os.File {
	t.Helper()
	r := lockWithin(t, path, 2*time.Second)
	if r.err != nil {
		t.Fatalf("Lock: %v", r.err)
	}
	return r.f
}

// Case 1a: no lock file yet.
func TestLockNewFile(t *testing.T) {
	shortTimes(t)
	path := filepath.Join(t.TempDir(), ".events.lock")

	f := mustLock(t, path)
	defer Unlock(f)

	if _, err := os.Stat(path); err != nil {
		t.Errorf("the lock file was not made: %v", err)
	}
}

// Case 1b: a lock file is left behind, e.g. after a crash, and nobody holds it.
func TestLockLeftoverFile(t *testing.T) {
	shortTimes(t)
	path := filepath.Join(t.TempDir(), ".events.lock")
	if err := os.WriteFile(path, nil, 0o644); err != nil {
		t.Fatal(err)
	}

	r := lockWithin(t, path, 2*time.Second)
	if StaleLockBlocks {
		if !errors.Is(r.err, ErrBusy) {
			t.Errorf("got %v, want ErrBusy (a left lock file blocks here)", r.err)
		}
		return
	}
	if r.err != nil {
		t.Fatalf("got %v, want the lock (a left lock file never blocks here)", r.err)
	}
	Unlock(r.f)
}

// Case 2: someone holds the lock until the deadline passes.
func TestLockBusy(t *testing.T) {
	shortTimes(t)
	path := filepath.Join(t.TempDir(), ".events.lock")
	held := mustLock(t, path)
	defer Unlock(held)

	r := lockWithin(t, path, 2*time.Second)
	if !errors.Is(r.err, ErrBusy) {
		t.Fatalf("got %v, want ErrBusy", r.err)
	}
	if r.f != nil {
		t.Errorf("got a file with ErrBusy, want nil")
	}
	if r.took < limitRetryTime {
		t.Errorf("gave up after %v, want at least the deadline %v", r.took, limitRetryTime)
	}
}

// Case 3: someone holds the lock, and frees it while we wait.
func TestLockWaitsForUnlock(t *testing.T) {
	shortTimes(t)
	path := filepath.Join(t.TempDir(), ".events.lock")
	held := mustLock(t, path)

	// Free the lock 50 ms from now, while the Lock below is waiting.
	time.AfterFunc(50*time.Millisecond, func() {
		if err := Unlock(held); err != nil {
			t.Errorf("Unlock: %v", err)
		}
	})

	r := lockWithin(t, path, 2*time.Second)
	if r.err != nil {
		t.Fatalf("got %v, want the lock once it was freed", r.err)
	}
	Unlock(r.f)
}

// Case 4: lock, unlock, then lock again.
func TestLockAgainAfterUnlock(t *testing.T) {
	shortTimes(t)
	path := filepath.Join(t.TempDir(), ".events.lock")

	f := mustLock(t, path)
	if err := Unlock(f); err != nil {
		t.Fatalf("Unlock: %v", err)
	}
	f = mustLock(t, path)
	Unlock(f)
}

// Case 5: the folder does not exist. This is not "busy", so it must fail at
// once, without waiting for the deadline.
func TestLockMissingFolder(t *testing.T) {
	shortTimes(t)
	path := filepath.Join(t.TempDir(), "no-such-folder", ".events.lock")

	r := lockWithin(t, path, 2*time.Second)
	if r.err == nil {
		Unlock(r.f)
		t.Fatal("got the lock, want an error")
	}
	if errors.Is(r.err, ErrBusy) {
		t.Errorf("got ErrBusy, want the real error")
	}
	if r.took >= limitRetryTime {
		t.Errorf("took %v, want it to fail at once", r.took)
	}
}

// Case 6: after Unlock, Unix keeps the lock file, and Windows deletes it.
func TestUnlockLockFile(t *testing.T) {
	shortTimes(t)
	path := filepath.Join(t.TempDir(), ".events.lock")

	f := mustLock(t, path)
	if err := Unlock(f); err != nil {
		t.Fatalf("Unlock: %v", err)
	}

	_, err := os.Stat(path)
	if runtime.GOOS == "windows" {
		if !errors.Is(err, os.ErrNotExist) {
			t.Errorf("the lock file is still there after Unlock (err %v)", err)
		}
		return
	}
	if err != nil {
		t.Errorf("the lock file is gone after Unlock, want it kept: %v", err)
	}
}

// Case 7: a second Unlock returns an error, and does not crash.
func TestUnlockTwice(t *testing.T) {
	shortTimes(t)
	path := filepath.Join(t.TempDir(), ".events.lock")

	f := mustLock(t, path)
	if err := Unlock(f); err != nil {
		t.Fatalf("first Unlock: %v", err)
	}
	if err := Unlock(f); err == nil {
		t.Errorf("second Unlock gave no error")
	}
}
