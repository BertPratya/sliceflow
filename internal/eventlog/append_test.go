package eventlog_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/BertPratya/sliceflow/internal/eventlog"
	"github.com/BertPratya/sliceflow/internal/filelock"
)

// Test 16 needs several processes (R30): a lock that works between
// goroutines may still fail between programs. So the test program starts
// copies of itself. These environment variables tell a copy to act as a
// helper instead of running the tests.
const (
	envMode  = "EVENTLOG_TEST_HELPER" // "append"
	envPath  = "EVENTLOG_TEST_PATH"   // the log
	envTimes = "EVENTLOG_TEST_TIMES"  // how many events to add
)

func TestMain(m *testing.M) {
	if mode := os.Getenv(envMode); mode != "" {
		os.Exit(runHelper(mode))
	}
	os.Exit(m.Run())
}

// runHelper is what a helper process does. It returns the exit code.
//   - append: add envTimes events to the log at envPath, one Append each.
func runHelper(mode string) int {
	if mode != "append" {
		fmt.Fprintln(os.Stderr, "unknown helper mode:", mode)
		return 2
	}
	times, err := strconv.Atoi(os.Getenv(envTimes))
	if err != nil {
		fmt.Fprintln(os.Stderr, "append:", err)
		return 2
	}
	for i := 0; i < times; i++ {
		err := eventlog.Append(context.Background(), os.Getenv(envPath), "test.added", map[string]int{"pid": os.Getpid()})
		if err != nil {
			fmt.Fprintln(os.Stderr, "append:", err)
			return 2
		}
	}
	return 0
}

// timeLayout is how at is written, as R27 says.
const timeLayout = "2006-01-02T15:04:05Z"

// created is the data that the Create tests write.
var created = eventlog.SliceCreated{
	Name:        "card-payment",
	Title:       "Pay with a card",
	Description: "Let a customer pay.",
	Base:        "main",
	BaseCommit:  "3f9a1c2e8b7d4a6f0c1e5d9b2a7f3c8e4d6b1a09",
}

// newLogPath returns the path of a log in a new temporary folder. The log
// itself does not exist yet.
func newLogPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), ".events.jsonl")
}

// readFile returns the content of path. It fails the test on an error.
func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// readAll reads the log at path with Read. It fails the test if Read
// returns an error.
func readAll(t *testing.T, path string) ([]eventlog.Event, []eventlog.BadLine) {
	t.Helper()
	events, bad, err := eventlog.Read(path)
	if err != nil {
		t.Fatalf("Read returned the error %v, want no error", err)
	}
	return events, bad
}

// badLineReasons returns each bad line as "line <n>: <reason>", for
// failure messages.
func badLineReasons(bad []eventlog.BadLine) []string {
	out := make([]string, len(bad))
	for i, b := range bad {
		out[i] = fmt.Sprintf("line %d: %v", b.Line, b.Err)
	}
	return out
}

// mustAppend adds one event of the name event, with empty data, to the
// log at path. It fails the test if Append returns an error.
func mustAppend(t *testing.T, path, event string) {
	t.Helper()
	if err := eventlog.Append(context.Background(), path, event, struct{}{}); err != nil {
		t.Fatalf("Append returned the error %v, want no error", err)
	}
}

// holdLock takes the log's lock in this test, and releases it when the
// test ends.
func holdLock(t *testing.T, path string) {
	t.Helper()
	l, err := filelock.Acquire(context.Background(), path+".lock", time.Second, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("take the lock: %v", err)
	}
	t.Cleanup(func() { l.Release() })
}

// Test 1: Create writes one slice.created line: version 1, seq 1, at the
// time of the call, the five data keys, the keys in the order of R27,
// and a "\n" at the end.
func TestCreateWritesTheFirstLine(t *testing.T) {
	path := newLogPath(t)
	before := time.Now().UTC().Truncate(time.Second)

	if err := eventlog.Create(context.Background(), path, created); err != nil {
		t.Fatalf("Create returned the error %v, want no error", err)
	}
	after := time.Now().UTC()

	events, bad := readAll(t, path)
	if len(events) != 1 || len(bad) != 0 {
		t.Fatalf("got %d events and bad lines %v, want 1 event and none", len(events), badLineReasons(bad))
	}
	at := events[0].At
	got, err := time.Parse(timeLayout, at)
	if err != nil {
		t.Fatalf("at is %q, want a time like 2026-11-02T09:15:30Z", at)
	}
	if got.Before(before) || got.After(after) {
		t.Errorf("at is %s, want a time between %s and %s", at, before.Format(timeLayout), after.Format(timeLayout))
	}

	want := `{"version":1,"seq":1,"event":"slice.created","at":"` + at + `","data":` +
		`{"name":"card-payment","title":"Pay with a card","description":"Let a customer pay.",` +
		`"base":"main","base_commit":"3f9a1c2e8b7d4a6f0c1e5d9b2a7f3c8e4d6b1a09"}}` + "\n"
	if content := readFile(t, path); content != want {
		t.Errorf("the log is\n%s\nwant\n%s", content, want)
	}
}

// Test 2: Create on a log that already exists returns an error that
// matches fs.ErrExist, and does not change the log.
func TestCreateWhenTheLogExists(t *testing.T) {
	old := eventLine(1, "x") + "\n"
	path := writeLog(t, old)

	err := eventlog.Create(context.Background(), path, created)
	if !errors.Is(err, fs.ErrExist) {
		t.Errorf("Create returned %v, want an error that matches fs.ErrExist", err)
	}
	if content := readFile(t, path); content != old {
		t.Errorf("the log changed to %q, want %q", content, old)
	}
}

// Test 3: when Append fails inside Create, here because the lock is held
// and the context runs out, Create returns an error and removes the log
// it made.
func TestCreateRemovesTheLogWhenAppendFails(t *testing.T) {
	path := newLogPath(t)
	holdLock(t, path)
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	if err := eventlog.Create(ctx, path, created); err == nil {
		t.Fatal("Create returned no error, want one: the lock is held")
	}
	if _, err := os.Stat(path); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("the log is still there (Stat: %v), want it removed", err)
	}
}

// Test 4: <, >, and & are written as they are, not as \u003c, \u003e,
// and \u0026 (R27).
func TestCreateWritesHTMLCharactersAsTheyAre(t *testing.T) {
	path := newLogPath(t)
	data := created
	data.Title = "Fish & Chips <3>"

	if err := eventlog.Create(context.Background(), path, data); err != nil {
		t.Fatalf("Create returned the error %v, want no error", err)
	}
	if content := readFile(t, path); !strings.Contains(content, `"title":"Fish & Chips <3>"`) {
		t.Errorf("the log is %s, want it to have \"title\":\"Fish & Chips <3>\"", content)
	}
}

// Test 5: Append to a log that does not exist returns an error that
// matches fs.ErrNotExist, and does not make the log.
func TestAppendToAMissingLog(t *testing.T) {
	path := newLogPath(t)

	err := eventlog.Append(context.Background(), path, "x", struct{}{})
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("Append returned %v, want an error that matches fs.ErrNotExist", err)
	}
	if _, err := os.Stat(path); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("the log exists (Stat: %v), want no log", err)
	}
}

// Test 6: after Create, Append writes seq 2, with the event name and the
// data it was given. Both lines can be read.
func TestAppendTheSecondEvent(t *testing.T) {
	path := newLogPath(t)
	if err := eventlog.Create(context.Background(), path, created); err != nil {
		t.Fatalf("Create returned the error %v, want no error", err)
	}

	data := map[string]string{"by": "bert"}
	if err := eventlog.Append(context.Background(), path, "slice.started", data); err != nil {
		t.Fatalf("Append returned the error %v, want no error", err)
	}

	events, bad := readAll(t, path)
	if len(events) != 2 || len(bad) != 0 {
		t.Fatalf("got %d events and bad lines %v, want 2 events and none", len(events), badLineReasons(bad))
	}
	e := events[1]
	if e.Seq != 2 || e.Event != "slice.started" || e.Version != 1 {
		t.Errorf("second event: seq %d, event %q, version %d; want 2, slice.started, 1", e.Seq, e.Event, e.Version)
	}
	if string(e.Data) != `{"by":"bert"}` {
		t.Errorf("data is %s, want {\"by\":\"bert\"}", e.Data)
	}
}

// Test 7: the new seq is the highest seq plus one, not the last seq plus
// one (R31).
func TestAppendUsesTheHighestSeq(t *testing.T) {
	path := writeLog(t, eventLine(1, "x")+"\n"+eventLine(5, "x")+"\n"+eventLine(3, "x")+"\n")

	mustAppend(t, path, "y")

	events, _ := readAll(t, path)
	if got := events[len(events)-1].Seq; got != 6 {
		t.Errorf("new seq is %d, want 6", got)
	}
}

// Test 8: a bad line does not count when Append finds the highest seq,
// even when it holds a higher number.
func TestAppendIgnoresBadLines(t *testing.T) {
	path := writeLog(t, eventLine(1, "x")+"\n"+`{"version":1,"seq":9,"ev`+"\n")

	mustAppend(t, path, "y")

	events, _ := readAll(t, path)
	if got := events[len(events)-1].Seq; got != 2 {
		t.Errorf("new seq is %d, want 2", got)
	}
}

// Test 9: when the last line is cut off, with no "\n", Append adds "\n"
// first. The cut-off line stays one bad line, and the new event can be
// read (R34).
func TestAppendAfterACutOffLastLine(t *testing.T) {
	old := eventLine(1, "x") + "\n" + `{"version":1,"se`
	path := writeLog(t, old)

	mustAppend(t, path, "y")

	content := readFile(t, path)
	if !strings.HasPrefix(content, old+"\n{") {
		t.Errorf("the log is %q, want it to start with %q", content, old+"\n{")
	}
	events, bad := readAll(t, path)
	if len(events) != 2 || events[1].Seq != 2 {
		t.Errorf("got %d events, want 2, the new one with seq 2", len(events))
	}
	if got := badLineNumbers(bad); len(got) != 1 || got[0] != 2 {
		t.Errorf("bad lines: %v, want [2]", got)
	}
}

// Test 10: when the last line ends with "\r\n", Append adds no "\n"
// first, so no empty line appears (R33, R34).
func TestAppendAfterAWindowsLineEnding(t *testing.T) {
	old := eventLine(1, "x") + "\r\n"
	path := writeLog(t, old)

	mustAppend(t, path, "y")

	content := readFile(t, path)
	if !strings.HasPrefix(content, old+"{") {
		t.Errorf("the log is %q, want the new line right after %q", content, old)
	}
	if _, bad := readAll(t, path); len(bad) != 0 {
		t.Errorf("bad lines: %v, want none", badLineReasons(bad))
	}
}

// Test 11: Append to an empty log writes seq 1, with no "\n" before it.
func TestAppendToAnEmptyLog(t *testing.T) {
	path := writeLog(t, "")

	mustAppend(t, path, "y")

	content := readFile(t, path)
	if !strings.HasPrefix(content, `{"version":1,"seq":1,`) {
		t.Errorf("the log is %q, want it to start with {\"version\":1,\"seq\":1,", content)
	}
	if !strings.HasSuffix(content, "}\n") || strings.Count(content, "\n") != 1 {
		t.Errorf("the log is %q, want one line that ends with \"\\n\"", content)
	}
}

// Test 12: Append never changes the lines already in the log: the old
// content is exactly the start of the new content (R35).
func TestAppendKeepsOldLines(t *testing.T) {
	old := createdLine + "\n" + eventLine(2, "x") + "\n" + "hello\n"
	path := writeLog(t, old)

	mustAppend(t, path, "y")

	if content := readFile(t, path); !strings.HasPrefix(content, old) {
		t.Errorf("the log is %q, want it to start with %q", content, old)
	}
}

// Test 13: when Read stops with an error, here a line over 4 MiB,
// Append returns an error and writes nothing.
func TestAppendWritesNothingWhenReadFails(t *testing.T) {
	old := longCreatedLine(5<<20) + "\n"
	path := writeLog(t, old)

	if err := eventlog.Append(context.Background(), path, "y", struct{}{}); err == nil {
		t.Error("Append returned no error, want one: the log has a line over 4 MiB")
	}
	if content := readFile(t, path); content != old {
		t.Errorf("the log changed: %d bytes, want %d", len(content), len(old))
	}
}

// Test 14: when the data cannot be turned into JSON, here a function,
// Append returns an error and writes nothing, not even the "\n" for the
// cut-off last line.
func TestAppendWritesNothingWhenEncodingFails(t *testing.T) {
	old := eventLine(1, "x") + "\n" + `{"version":1,"se`
	path := writeLog(t, old)

	if err := eventlog.Append(context.Background(), path, "y", func() {}); err == nil {
		t.Error("Append returned no error, want one: a function has no JSON")
	}
	if content := readFile(t, path); content != old {
		t.Errorf("the log changed to %q, want %q", content, old)
	}
}

// Test 15: when the lock is held and the context runs out, Append
// returns the context's error at once, and writes nothing.
func TestAppendStopsWhenTheContextRunsOut(t *testing.T) {
	old := eventLine(1, "x") + "\n"
	path := writeLog(t, old)
	holdLock(t, path)
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	start := time.Now()
	err := eventlog.Append(ctx, path, "y", struct{}{})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Append returned %v, want an error that matches context.DeadlineExceeded", err)
	}
	if waited := time.Since(start); waited > 5*time.Second {
		t.Errorf("Append waited %v, want it to stop soon after the context ran out", waited)
	}
	if content := readFile(t, path); content != old {
		t.Errorf("the log changed to %q, want %q", content, old)
	}
}

// Test 16: 4 processes add 25 events each to one log at the same time.
// Every event is kept, whole and on its own line, and the seqs are 1 to
// 100 in file order, each one once (R30, R31). Without the lock, two
// processes can read the same highest seq and write the same number.
func TestAppendFromFourProcessesAtOnce(t *testing.T) {
	const procs, times = 4, 25
	path := writeLog(t, "")
	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("find the test program: %v", err)
	}

	cmds := make([]*exec.Cmd, procs)
	outs := make([]*bytes.Buffer, procs)
	for i := range cmds {
		// -test.run=^$ matches no test, in case the helper mode is not seen.
		cmd := exec.Command(exe, "-test.run=^$")
		cmd.Env = append(os.Environ(), envMode+"=append", envPath+"="+path, envTimes+"="+strconv.Itoa(times))
		outs[i] = new(bytes.Buffer)
		cmd.Stderr = outs[i]
		if err := cmd.Start(); err != nil {
			t.Fatalf("start process %d: %v", i+1, err)
		}
		cmds[i] = cmd
	}
	for i, cmd := range cmds {
		if err := cmd.Wait(); err != nil {
			t.Errorf("process %d failed: %v\n%s", i+1, err, outs[i])
		}
	}

	events, bad := readAll(t, path)
	if len(bad) != 0 {
		t.Errorf("bad lines: %v, want none", badLineReasons(bad))
	}
	if len(events) != procs*times {
		t.Fatalf("got %d events, want %d", len(events), procs*times)
	}
	for i, e := range events {
		if e.Seq != i+1 {
			t.Fatalf("event %d has seq %d, want %d: seqs must be 1 to %d in file order", i+1, e.Seq, i+1, procs*times)
		}
	}
}

// Test 17: when Append fails after taking the lock, here because Read
// stops on a line over 4 MiB, it still releases the lock. Otherwise the
// next Append in the same process waits the full 10 seconds, then fails.
func TestAppendReleasesTheLockWhenItFails(t *testing.T) {
	path := writeLog(t, longCreatedLine(5<<20)+"\n")

	if err := eventlog.Append(context.Background(), path, "y", struct{}{}); err == nil {
		t.Fatal("Append returned no error, want one: the log has a line over 4 MiB")
	}

	l, err := filelock.Acquire(context.Background(), path+".lock", 500*time.Millisecond, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("take the lock after the failed Append: %v, want the lock to be free", err)
	}
	l.Release()
}

// Test 18: when the log cannot be opened for writing, here because it
// is read-only, Append returns an error, and does not crash.
func TestAppendToAReadOnlyLog(t *testing.T) {
	old := eventLine(1, "x") + "\n"
	path := writeLog(t, old)
	if err := os.Chmod(path, 0o444); err != nil {
		t.Fatal(err)
	}
	// Make it writable again, so the temporary folder can be removed on
	// Windows.
	t.Cleanup(func() { os.Chmod(path, 0o644) })

	if err := eventlog.Append(context.Background(), path, "y", struct{}{}); err == nil {
		t.Error("Append returned no error, want one: the log is read-only")
	}
	if content := readFile(t, path); content != old {
		t.Errorf("the log changed to %q, want %q", content, old)
	}
}
