package eventlog

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/BertPratya/sliceflow/internal/filelock"
	"github.com/BertPratya/sliceflow/internal/gittest"
)

var titledAt = time.Date(2026, 10, 31, 9, 0, 0, 0, time.UTC)

// titled is the event the tests add. Seq is 0: AddEvent picks the seq.
func titled() Event {
	return Event{
		Version: Version,
		Event:   EventSliceTitled,
		At:      titledAt,
		Data:    json.RawMessage(`{"title":"Submit a spec"}`),
	}
}

// titledLine is the line AddEvent should write for titled() with seq.
func titledLine(t *testing.T, seq int) string {
	t.Helper()
	e := titled()
	e.Seq = seq
	b, err := Line(e)
	if err != nil {
		t.Fatalf("Line: %v", err)
	}
	return string(b)
}

// createdLine is a good slice.created line with the given seq.
func createdLine(t *testing.T, seq int) string {
	t.Helper()
	data, err := Data(SliceCreated{From: Ref{Branch: "main", Commit: commit}})
	if err != nil {
		t.Fatalf("Data: %v", err)
	}
	b, err := Line(Event{Version: Version, Seq: seq, Event: EventSliceCreated, At: titledAt, Data: data})
	if err != nil {
		t.Fatalf("Line: %v", err)
	}
	return string(b)
}

// sliceWithLog makes a repo with a slice folder whose log holds content,
// commits the log, and returns the repo and the slice folder.
func sliceWithLog(t *testing.T, content string) (repo, sliceDir string) {
	t.Helper()
	repo = gittest.NewRepo(t)
	sliceDir = filepath.Join(repo, "docs", "slices", "demo")
	if err := os.MkdirAll(sliceDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(sliceDir, ".events.jsonl"), content)
	gittest.Run(t, repo, "add", ".")
	gittest.Run(t, repo, "commit", "-m", "demo: slice created")
	return repo, sliceDir
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// R29: the event is added at the end, with the next seq, and only the log
// is committed, with the message given.
func TestAddEventCommitsTheLog(t *testing.T) {
	first := makeLine(t, titledAt, "")
	repo, sliceDir := sliceWithLog(t, first)
	logPath := filepath.Join(sliceDir, ".events.jsonl")

	// Another file, staged by the user, must stay staged and not committed.
	writeFile(t, filepath.Join(repo, "notes.txt"), "mine\n")
	gittest.Run(t, repo, "add", "notes.txt")

	if _, err := AddEvent(sliceDir, titled(), "demo: title set"); err != nil {
		t.Fatalf("AddEvent: %v", err)
	}

	if got, want := readFile(t, logPath), first+titledLine(t, 2); got != want {
		t.Errorf("log:\ngot  %q\nwant %q", got, want)
	}
	if got := gittest.Run(t, repo, "log", "-1", "--format=%s"); got != "demo: title set" {
		t.Errorf("commit message: got %q", got)
	}
	if got := gittest.Run(t, repo, "show", "--name-only", "--format=", "HEAD"); got != "docs/slices/demo/.events.jsonl" {
		t.Errorf("files in the commit: got %q, want only the log", got)
	}
	if got := gittest.Run(t, repo, "diff", "--cached", "--name-only"); got != "notes.txt" {
		t.Errorf("staged after: got %q, want notes.txt still staged", got)
	}
}

// R33: the seq is 1 more than the seq of the last good line, not the
// number of lines.
func TestAddEventSeqFromLastLine(t *testing.T) {
	first := makeLine(t, titledAt, "")
	third := createdLine(t, 3) // a gap: seq 3 after seq 1
	_, sliceDir := sliceWithLog(t, first+third)
	logPath := filepath.Join(sliceDir, ".events.jsonl")

	warnings, err := AddEvent(sliceDir, titled(), "demo: title set")
	if err != nil {
		t.Fatalf("AddEvent: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("warnings: got %q, want none", warnings)
	}
	if got, want := readFile(t, logPath), first+third+titledLine(t, 4); got != want {
		t.Errorf("log:\ngot  %q\nwant %q", got, want)
	}
}

// R33: an empty log gets seq 1.
func TestAddEventEmptyLog(t *testing.T) {
	_, sliceDir := sliceWithLog(t, "")
	logPath := filepath.Join(sliceDir, ".events.jsonl")

	if _, err := AddEvent(sliceDir, titled(), "demo: title set"); err != nil {
		t.Fatalf("AddEvent: %v", err)
	}
	if got, want := readFile(t, logPath), titledLine(t, 1); got != want {
		t.Errorf("log:\ngot  %q\nwant %q", got, want)
	}
}

// R33 / S17: a half-written last line is removed, and the new event takes
// its place, with the seq after the last good line.
func TestAddEventHalfWrittenLine(t *testing.T) {
	first := makeLine(t, titledAt, "")
	_, sliceDir := sliceWithLog(t, first+`{"version":1,"seq":2,"ev`)
	logPath := filepath.Join(sliceDir, ".events.jsonl")

	warnings, err := AddEvent(sliceDir, titled(), "demo: title set")
	if err != nil {
		t.Fatalf("AddEvent: %v", err)
	}
	want := []string{".events.jsonl line 2: half-written last line; removed"}
	if !slices.Equal(warnings, want) {
		t.Errorf("warnings: got %q, want %q", warnings, want)
	}
	if got, want := readFile(t, logPath), first+titledLine(t, 2); got != want {
		t.Errorf("log:\ngot  %q\nwant %q", got, want)
	}
}

// R33: a log that is only one half-written line becomes the new event.
func TestAddEventOnlyHalfWrittenLine(t *testing.T) {
	_, sliceDir := sliceWithLog(t, `{"version":1,"se`)
	logPath := filepath.Join(sliceDir, ".events.jsonl")

	warnings, err := AddEvent(sliceDir, titled(), "demo: title set")
	if err != nil {
		t.Fatalf("AddEvent: %v", err)
	}
	want := []string{".events.jsonl line 1: half-written last line; removed"}
	if !slices.Equal(warnings, want) {
		t.Errorf("warnings: got %q, want %q", warnings, want)
	}
	if got, want := readFile(t, logPath), titledLine(t, 1); got != want {
		t.Errorf("log:\ngot  %q\nwant %q", got, want)
	}
}

// A missing log is ErrNotFound, and AddEvent does not make one.
func TestAddEventNoLog(t *testing.T) {
	repo := gittest.NewRepo(t)
	sliceDir := filepath.Join(repo, "docs", "slices", "demo")
	if err := os.MkdirAll(sliceDir, 0o755); err != nil {
		t.Fatal(err)
	}

	_, err := AddEvent(sliceDir, titled(), "demo: title set")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("got %v, want ErrNotFound", err)
	}
	if _, err := os.Stat(filepath.Join(sliceDir, ".events.jsonl")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("AddEvent made a log file")
	}
}

// R35 / S19: the commit fails, so the log is put back byte for byte, and
// AddEvent returns an error.
func TestAddEventCommitFails(t *testing.T) {
	first := makeLine(t, titledAt, "")
	original := first + `{"version":1,"seq":2,"ev` // the half line is put back too
	repo, sliceDir := sliceWithLog(t, original)
	logPath := filepath.Join(sliceDir, ".events.jsonl")
	head := gittest.Run(t, repo, "rev-parse", "HEAD")

	hook := filepath.Join(repo, ".git", "hooks", "pre-commit")
	if err := os.WriteFile(hook, []byte("#!/bin/sh\necho 'error: commits are blocked' >&2\nexit 1\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	_, err := AddEvent(sliceDir, titled(), "demo: title set")
	var we *WriteError
	if !errors.As(err, &we) {
		t.Fatalf("got %v, want a *WriteError", err)
	}
	if !we.Restored {
		t.Errorf("Restored is false, want true")
	}
	if !strings.Contains(err.Error(), "error: commits are blocked") {
		t.Errorf("error %q does not hold git's reason", err)
	}
	if got := readFile(t, logPath); got != original {
		t.Errorf("log not put back:\ngot  %q\nwant %q", got, original)
	}
	if got := gittest.Run(t, repo, "rev-parse", "HEAD"); got != head {
		t.Errorf("a commit was made")
	}
	if _, err := os.Stat(logPath + ".tmp"); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("the .tmp file is left behind after a good restore")
	}
}

// R34 / S18: another process holds the lock, so AddEvent gives up with
// ErrBusy and changes nothing. It waits the full 10 seconds.
func TestAddEventBusy(t *testing.T) {
	if testing.Short() {
		t.Skip("waits 10 seconds")
	}
	first := makeLine(t, titledAt, "")
	_, sliceDir := sliceWithLog(t, first)
	logPath := filepath.Join(sliceDir, ".events.jsonl")

	held, err := filelock.Lock(filepath.Join(sliceDir, ".events.lock"))
	if err != nil {
		t.Fatal(err)
	}
	defer filelock.Unlock(held)

	_, err = AddEvent(sliceDir, titled(), "demo: title set")
	if !errors.Is(err, filelock.ErrBusy) {
		t.Fatalf("got %v, want ErrBusy", err)
	}
	if got := readFile(t, logPath); got != first {
		t.Errorf("the log changed while busy")
	}
}
