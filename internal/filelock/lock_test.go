package filelock_test

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/BertPratya/sliceflow/internal/filelock"
)

// Some tests need a second process (TD7): a lock that works between two
// goroutines may still fail between two programs. So the test program
// starts a second copy of itself. These environment variables tell the
// second copy to act as a helper instead of running the tests.
const (
	envMode    = "FILELOCK_TEST_HELPER"  // "hold", "crash", or "count"
	envPath    = "FILELOCK_TEST_PATH"    // the lock file
	envCounter = "FILELOCK_TEST_COUNTER" // "count" only: the counter file
	envTimes   = "FILELOCK_TEST_TIMES"   // "count" only: how many times to add 1
)

// interval is the wait between tries that every test passes to Acquire.
const interval = 50 * time.Millisecond

func TestMain(m *testing.M) {
	if mode := os.Getenv(envMode); mode != "" {
		os.Exit(runHelper(mode))
	}
	os.Exit(m.Run())
}

// runHelper is what the second process does. It returns the exit code.
//   - hold: take the lock, print "locked", wait until stdin is closed,
//     then release the lock.
//   - crash: take the lock, print "locked", and exit without releasing.
//   - count: add 1 to the number in the counter file, envTimes times,
//     taking the lock each time.
func runHelper(mode string) int {
	path := os.Getenv(envPath)
	switch mode {
	case "hold":
		l, err := filelock.Acquire(context.Background(), path, 10*time.Second, interval)
		if err != nil {
			fmt.Fprintln(os.Stderr, "hold: Acquire:", err)
			return 2
		}
		fmt.Println("locked")
		io.Copy(io.Discard, os.Stdin) // returns when the test closes stdin
		if err := l.Release(); err != nil {
			fmt.Fprintln(os.Stderr, "hold: Release:", err)
			return 2
		}
		return 0
	case "crash":
		if _, err := filelock.Acquire(context.Background(), path, 10*time.Second, interval); err != nil {
			fmt.Fprintln(os.Stderr, "crash: Acquire:", err)
			return 2
		}
		fmt.Println("locked")
		return 0 // exits with the lock still held
	case "count":
		times, err := strconv.Atoi(os.Getenv(envTimes))
		if err != nil {
			fmt.Fprintln(os.Stderr, "count:", err)
			return 2
		}
		for i := 0; i < times; i++ {
			if err := addOne(path, os.Getenv(envCounter)); err != nil {
				fmt.Fprintln(os.Stderr, "count:", err)
				return 2
			}
		}
		return 0
	}
	fmt.Fprintln(os.Stderr, "unknown helper mode:", mode)
	return 2
}

// addOne reads the number in counter, and writes it back plus one, while
// holding the lock. Without a working lock, two processes can read the
// same number, and one of the additions is lost.
func addOne(lockPath, counter string) error {
	l, err := filelock.Acquire(context.Background(), lockPath, 30*time.Second, interval)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(counter)
	if err != nil {
		l.Release()
		return err
	}
	n, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		l.Release()
		return err
	}
	if err := os.WriteFile(counter, []byte(strconv.Itoa(n+1)), 0o644); err != nil {
		l.Release()
		return err
	}
	return l.Release()
}

// helperCmd makes the command that starts a second copy of this test
// program in the given helper mode.
func helperCmd(t *testing.T, mode, path string, env ...string) *exec.Cmd {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("find the test program: %v", err)
	}
	// -test.run=^$ matches no test, in case the helper mode is not seen.
	cmd := exec.Command(exe, "-test.run=^$")
	cmd.Env = append(os.Environ(), envMode+"="+mode, envPath+"="+path)
	cmd.Env = append(cmd.Env, env...)
	return cmd
}

// startHolder starts a second process that takes the lock at path, and
// returns once it holds it. Calling release makes the process release the
// lock and exit. When the test ends, release is called if it was not, and
// the test fails if the second process failed.
func startHolder(t *testing.T, path string) (release func()) {
	t.Helper()
	cmd := helperCmd(t, "hold", path)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		t.Fatalf("start the second process: %v", err)
	}

	line, err := bufio.NewReader(stdout).ReadString('\n')
	if strings.TrimSpace(line) != "locked" {
		cmd.Wait()
		t.Fatalf("the second process did not take the lock (read %q, %v):\n%s",
			line, err, stderr.String())
	}

	var once sync.Once
	release = func() { once.Do(func() { stdin.Close() }) }
	t.Cleanup(func() {
		release()
		if err := cmd.Wait(); err != nil {
			t.Errorf("the second process failed: %v\n%s", err, stderr.String())
		}
	})
	return release
}

// Test 1: lock, release, then lock again. The lock file is made when it
// is missing, and is not deleted by Release.
func TestAcquireReleaseAcquire(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.lock")
	ctx := context.Background()

	l, err := filelock.Acquire(ctx, path, time.Second, interval)
	if err != nil {
		t.Fatalf("first Acquire: %v", err)
	}
	if err := l.Release(); err != nil {
		t.Fatalf("first Release: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("the lock file should still exist after Release: %v", err)
	}

	l, err = filelock.Acquire(ctx, path, time.Second, interval)
	if err != nil {
		t.Fatalf("second Acquire: %v", err)
	}
	if err := l.Release(); err != nil {
		t.Fatalf("second Release: %v", err)
	}
}

// Test 2: while a second process holds the lock, Acquire keeps trying
// until its wait time is over, then returns ErrBusy and no lock.
func TestAcquireBusyWhileAnotherProcessHolds(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.lock")
	startHolder(t, path)

	const wait = 300 * time.Millisecond
	start := time.Now()
	l, err := filelock.Acquire(context.Background(), path, wait, interval)
	elapsed := time.Since(start)

	if !errors.Is(err, filelock.ErrBusy) {
		t.Fatalf("Acquire returned %v, want ErrBusy", err)
	}
	if l != nil {
		t.Errorf("Acquire returned a lock together with an error")
	}
	// Allow one interval of slack below the wait time.
	if elapsed < wait-interval {
		t.Errorf("Acquire gave up after %v, want it to keep trying for about %v", elapsed, wait)
	}
	if elapsed > 3*time.Second {
		t.Errorf("Acquire took %v, far more than the wait time %v", elapsed, wait)
	}
}

// Test 3: when the second process releases the lock while Acquire is
// waiting, Acquire gets the lock. This tests the retry.
func TestAcquireWaitsForRelease(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.lock")
	release := startHolder(t, path)
	time.AfterFunc(200*time.Millisecond, release)

	l, err := filelock.Acquire(context.Background(), path, 5*time.Second, interval)
	if err != nil {
		t.Fatalf("Acquire returned %v, want the lock once the other process released it", err)
	}
	if err := l.Release(); err != nil {
		t.Fatalf("Release: %v", err)
	}
}

// Test 4: when the process holding the lock exits without releasing it,
// as in a crash, the OS releases the lock, and Acquire gets it.
func TestAcquireAfterHolderExitsWithoutRelease(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.lock")

	cmd := helperCmd(t, "crash", path)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil || strings.TrimSpace(string(out)) != "locked" {
		t.Fatalf("the second process did not take the lock (%q, %v):\n%s", out, err, stderr.String())
	}

	l, err := filelock.Acquire(context.Background(), path, time.Second, interval)
	if err != nil {
		t.Fatalf("Acquire returned %v, want the lock left by the exited process", err)
	}
	if err := l.Release(); err != nil {
		t.Fatalf("Release: %v", err)
	}
}

// Test 5: when the context is cancelled while Acquire waits (Ctrl+C),
// Acquire stops soon, and returns the context's error and no lock.
func TestAcquireStopsWhenContextCancelled(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.lock")
	startHolder(t, path)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	time.AfterFunc(100*time.Millisecond, cancel)

	start := time.Now()
	l, err := filelock.Acquire(ctx, path, 10*time.Second, interval)
	elapsed := time.Since(start)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Acquire returned %v, want context.Canceled", err)
	}
	if l != nil {
		t.Errorf("Acquire returned a lock together with an error")
	}
	if elapsed > 2*time.Second {
		t.Errorf("Acquire took %v to stop after the cancel, want it to stop soon", elapsed)
	}
}

// Test 6: when the folder of the lock file does not exist, Acquire
// returns an error at once. It is not ErrBusy, and no folder is made.
func TestAcquireMissingFolder(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "missing")
	path := filepath.Join(dir, "test.lock")

	start := time.Now()
	l, err := filelock.Acquire(context.Background(), path, 2*time.Second, interval)
	elapsed := time.Since(start)

	if err == nil {
		l.Release()
		t.Fatal("Acquire returned no error, want an error for a missing folder")
	}
	if errors.Is(err, filelock.ErrBusy) {
		t.Errorf("Acquire returned ErrBusy (%v), want the error of the missing folder", err)
	}
	if l != nil {
		t.Errorf("Acquire returned a lock together with an error")
	}
	if elapsed > time.Second {
		t.Errorf("Acquire took %v, want it to return at once, not retry", elapsed)
	}
	if _, err := os.Stat(dir); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("Acquire should not make the folder; Stat returned %v", err)
	}
}

// Test 7: two processes each add 1 to a number in a file, many times,
// taking the lock each time. If the lock works, no addition is lost.
func TestTwoProcessesNeverLoseAnAddition(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "test.lock")
	counter := filepath.Join(dir, "counter.txt")
	if err := os.WriteFile(counter, []byte("0"), 0o644); err != nil {
		t.Fatal(err)
	}

	const processes, times = 2, 500
	cmds := make([]*exec.Cmd, processes)
	errs := make([]*bytes.Buffer, processes)
	for i := range cmds {
		cmds[i] = helperCmd(t, "count", lockPath,
			envCounter+"="+counter, envTimes+"="+strconv.Itoa(times))
		errs[i] = new(bytes.Buffer)
		cmds[i].Stderr = errs[i]
		if err := cmds[i].Start(); err != nil {
			t.Fatalf("start process %d: %v", i+1, err)
		}
	}
	for i, cmd := range cmds {
		if err := cmd.Wait(); err != nil {
			t.Errorf("process %d failed: %v\n%s", i+1, err, errs[i].String())
		}
	}

	data, err := os.ReadFile(counter)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := strings.TrimSpace(string(data)), strconv.Itoa(processes*times); got != want {
		t.Errorf("counter is %s, want %s: some additions were lost", got, want)
	}
}

// Test 9: an interval of 0 or less is allowed, and means "try again with
// no pause". Acquire still gets a free lock, and still returns ErrBusy at
// the end of the wait time when the lock is held.
func TestAcquireWithNoPause(t *testing.T) {
	for _, iv := range []time.Duration{0, -time.Second} {
		t.Run(iv.String(), func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "test.lock")

			l, err := filelock.Acquire(context.Background(), path, time.Second, iv)
			if err != nil {
				t.Fatalf("Acquire of a free lock returned %v", err)
			}
			if err := l.Release(); err != nil {
				t.Fatalf("Release: %v", err)
			}

			startHolder(t, path)
			start := time.Now()
			l, err = filelock.Acquire(context.Background(), path, 200*time.Millisecond, iv)
			elapsed := time.Since(start)
			if !errors.Is(err, filelock.ErrBusy) {
				t.Fatalf("Acquire of a held lock returned %v, want ErrBusy", err)
			}
			if l != nil {
				t.Errorf("Acquire returned a lock together with an error")
			}
			if elapsed > 3*time.Second {
				t.Errorf("Acquire took %v, far more than the wait time 200ms", elapsed)
			}
		})
	}
}

// Test 10: a busy Acquire leaves no file open. Each try opens the lock
// file, so a try that fails must close it, or the process runs out of
// files. This test counts the open files of this process, which only Linux
// shows in /proc/self/fd. On Windows, a leaked file is caught anyway: the
// test's temporary folder cannot be deleted while a file in it is open.
func TestBusyAcquireLeavesNoFileOpen(t *testing.T) {
	if _, err := os.Stat("/proc/self/fd"); err != nil {
		t.Skip("this system does not list open files in /proc/self/fd")
	}
	openFiles := func() int {
		entries, err := os.ReadDir("/proc/self/fd")
		if err != nil {
			t.Fatal(err)
		}
		return len(entries)
	}

	path := filepath.Join(t.TempDir(), "test.lock")
	startHolder(t, path)

	before := openFiles()
	// About 50 busy tries: 500 ms, one try every 10 ms.
	_, err := filelock.Acquire(context.Background(), path, 500*time.Millisecond, 10*time.Millisecond)
	if !errors.Is(err, filelock.ErrBusy) {
		t.Fatalf("Acquire returned %v, want ErrBusy", err)
	}
	if after := openFiles(); after != before {
		t.Errorf("open files: %d before Acquire, %d after; each busy try must close its file", before, after)
	}
}

// Test 8: calling Release a second time returns an error.
func TestReleaseTwice(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.lock")
	l, err := filelock.Acquire(context.Background(), path, time.Second, interval)
	if err != nil {
		t.Fatalf("Acquire: %v", err)
	}
	if err := l.Release(); err != nil {
		t.Fatalf("first Release: %v", err)
	}
	err = l.Release()
	if err == nil {
		t.Fatal("second Release returned nil, want an error")
	}
	if errors.Is(err, filelock.ErrBusy) {
		t.Errorf("second Release returned ErrBusy (%v), want an error that says the lock was already released", err)
	}
}
