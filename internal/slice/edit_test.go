package slice

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/BertPratya/sliceflow/internal/eventlog"
	"github.com/BertPratya/sliceflow/internal/filelock"
)

// newLog writes a log with one slice.created line, or start when it is
// not "", and returns its path and the path of its lock.
func newLog(t *testing.T, start string) (logFile, lockPath string) {
	t.Helper()
	dir := t.TempDir()
	logFile = filepath.Join(dir, ".events.jsonl")
	lockPath = filepath.Join(dir, "locks", "login.lock")
	if start == "" {
		created := eventlog.SliceCreated{Name: "login", Title: "Log in", Base: "main", BaseCommit: "3f9a1c2e8b7d4a6f0c1e5d9b2a7f3c8e4d6b1a09"}
		if err := eventlog.Create(context.Background(), logFile, lockPath, created); err != nil {
			t.Fatal(err)
		}
		return logFile, lockPath
	}
	if err := os.WriteFile(logFile, []byte(start), 0o644); err != nil {
		t.Fatal(err)
	}
	return logFile, lockPath
}

// readFile returns the text of the file at path.
func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// appendLine adds line at the end of the file at path, as another
// process would.
func appendLine(t *testing.T, path, line string) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := f.WriteString(line); err != nil {
		t.Fatal(err)
	}
}

const otherLine = `{"version":1,"seq":2,"event":"slice.edited","at":"2026-11-18T10:00:00Z","data":{"title":"Other"}}` + "\n"

// The Undo of edit's step cuts the log back only when what follows the
// old size is exactly the line the step wrote. It never removes a line
// another process added, and does nothing when the lock was not taken.
func TestEditStepUndo(t *testing.T) {
	title := "Sign in"
	edited := eventlog.SliceEdited{Title: &title}

	t.Run("its own line is removed", func(t *testing.T) {
		logFile, lockPath := newLog(t, "")
		before := readFile(t, logFile)
		steps, err := editSteps("/top", logFile, lockPath, edited)
		if err != nil {
			t.Fatal(err)
		}
		if err := steps[0].Do(context.Background()); err != nil {
			t.Fatal(err)
		}
		if err := steps[0].Undo(); err != nil {
			t.Fatalf("Undo: %v", err)
		}
		if after := readFile(t, logFile); after != before {
			t.Errorf("the log is:\n%s\nwant it as before:\n%s", after, before)
		}
	})

	t.Run("its own line and the line break before it, in a log with no last line break", func(t *testing.T) {
		start := `{"version":1,"seq":1,"event":"slice.created","at":"2026-11-18T10:00:00Z","data":{"name":"login","title":"Log in","description":"","base":"main","base_commit":"3f9a"}}`
		logFile, lockPath := newLog(t, start)
		steps, _ := editSteps("/top", logFile, lockPath, edited)
		if err := steps[0].Do(context.Background()); err != nil {
			t.Fatal(err)
		}
		if err := steps[0].Undo(); err != nil {
			t.Fatalf("Undo: %v", err)
		}
		if after := readFile(t, logFile); after != start {
			t.Errorf("the log is %q, want %q", after, start)
		}
	})

	t.Run("a line another process added after it stays", func(t *testing.T) {
		logFile, lockPath := newLog(t, "")
		steps, _ := editSteps("/top", logFile, lockPath, edited)
		if err := steps[0].Do(context.Background()); err != nil {
			t.Fatal(err)
		}
		appendLine(t, logFile, otherLine)
		grown := readFile(t, logFile)
		if err := steps[0].Undo(); err == nil {
			t.Error("Undo returned nil, want an error, so the log is reported as left over")
		}
		if after := readFile(t, logFile); after != grown {
			t.Errorf("the log is:\n%s\nwant it unchanged:\n%s", after, grown)
		}
	})

	t.Run("a line another process added while it waited stays", func(t *testing.T) {
		logFile, lockPath := newLog(t, "")
		steps, _ := editSteps("/top", logFile, lockPath, edited)
		appendLine(t, logFile, otherLine)
		if err := steps[0].Do(context.Background()); err != nil {
			t.Fatal(err)
		}
		grown := readFile(t, logFile)
		if err := steps[0].Undo(); err == nil {
			t.Error("Undo returned nil, want an error, so the log is reported as left over")
		}
		if after := readFile(t, logFile); after != grown {
			t.Errorf("the log is:\n%s\nwant it unchanged:\n%s", after, grown)
		}
	})

	t.Run("nothing is cut when the lock was not taken", func(t *testing.T) {
		logFile, lockPath := newLog(t, "")
		steps, _ := editSteps("/top", logFile, lockPath, edited)
		held, err := filelock.Acquire(context.Background(), lockPath, time.Second, 10*time.Millisecond)
		if err != nil {
			t.Fatal(err)
		}
		defer held.Release()
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		var lockErr *eventlog.LockError
		if err := steps[0].Do(ctx); err == nil || !errors.As(err, &lockErr) {
			t.Fatalf("Do gave %v, want a *eventlog.LockError", err)
		}
		appendLine(t, logFile, otherLine)
		grown := readFile(t, logFile)
		if err := steps[0].Undo(); err != nil {
			t.Errorf("Undo: %v, want nil: nothing of its own to remove", err)
		}
		if after := readFile(t, logFile); after != grown {
			t.Errorf("the log is:\n%s\nwant it unchanged:\n%s", after, grown)
		}
	})
}
