package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The warnings slice status prints about the log of login.
const (
	unknownEventLine2 = "Warning: line 2 of docs/slices/login/.events.jsonl has the event spec.submitted, which this version of sliceflow does not know, so the state cannot be worked out.\n"
	notAllowedLine2   = "Warning: line 2 of docs/slices/login/.events.jsonl has the event slice.created, which is not allowed while the slice is in spec (draft), so the state cannot be worked out.\n"
)

// slice status prints the State: line after the first line, then, after
// an empty line, the next step for spec (draft).
func TestStatusShowsStateAndNext(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)
	createSlice(t, repo, home, "login", "--title", "Log in", "--description", "Lets a user log in.")

	got := runSliceflow(t, repo, home, "slice", "status", "login")

	want := "login: Log in\n" +
		"  State:        spec (draft)\n" +
		"  Description:  Lets a user log in.\n" +
		"  Created:      " + createdTime(t, repo, "login") + "\n" +
		"  Base:         main\n" +
		"  Events:       1 (last: slice.created)\n" +
		"\n" +
		"Next: write the spec in " + specFile(t, repo, home, "login") + "\n" +
		"Then run: sliceflow spec validate login\n"
	wantOutput(t, got, want, "")
}

// An event this version does not know makes the state unknown, with a
// warning that names its line and its event.
func TestStatusUnknownEvent(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)
	createSlice(t, repo, home, "login")
	appendText(t, sliceLog(repo, "login"), eventLine(2, "spec.submitted", `{}`))

	got := runSliceflow(t, repo, home, "slice", "status", "login")

	wantCode(t, got, 0)
	if got.stderr != unknownEventLine2 {
		t.Errorf("stderr is:\n%s\nwant:\n%s", got.stderr, unknownEventLine2)
	}
}

// A second slice.created is not allowed in spec (draft): the warning
// names the state the slice was in.
func TestStatusEventNotAllowed(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)
	createSlice(t, repo, home, "login")
	appendText(t, sliceLog(repo, "login"), secondCreatedLine(t, repo))

	got := runSliceflow(t, repo, home, "slice", "status", "login")

	wantCode(t, got, 0)
	if got.stderr != notAllowedLine2 {
		t.Errorf("stderr is:\n%s\nwant:\n%s", got.stderr, notAllowedLine2)
	}
	if !strings.Contains(got.stdout, "  State:        unknown\n") {
		t.Errorf("stdout is:\n%s\nwant State: unknown", got.stdout)
	}
}

// With an unknown state, slice status prints State: unknown and its other
// lines, with no next step. Events: counts the event that stopped the
// state.
func TestStatusOfUnknownState(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)
	createSlice(t, repo, home, "login", "--title", "Log in")
	appendText(t, sliceLog(repo, "login"), eventLine(2, "spec.submitted", `{}`))

	got := runSliceflow(t, repo, home, "slice", "status", "login")

	want := "login: Log in\n" +
		"  State:        unknown\n" +
		"  Description:  (none)\n" +
		"  Created:      " + createdTime(t, repo, "login") + "\n" +
		"  Base:         main\n" +
		"  Events:       2 (last: spec.submitted)\n"
	wantOutput(t, got, want, unknownEventLine2)
}

// A line that cannot be read is skipped, and the state is worked out from
// the others. Its warning comes before the warning about the state,
// whatever their line numbers.
func TestStatusDamagedLineFirst(t *testing.T) {
	t.Run("a damaged line alone", func(t *testing.T) {
		home := newHome(t)
		repo := newProject(t, home)
		createSlice(t, repo, home, "login")
		appendText(t, sliceLog(repo, "login"), "cut off\n")

		got := runSliceflow(t, repo, home, "slice", "status", "login")

		wantCode(t, got, 0)
		wantStderr := "Warning: line 2 of docs/slices/login/.events.jsonl could not be read, so it was skipped.\n"
		if got.stderr != wantStderr {
			t.Errorf("stderr is:\n%s\nwant:\n%s", got.stderr, wantStderr)
		}
		if !strings.Contains(got.stdout, "  State:        spec (draft)\n") {
			t.Errorf("stdout is:\n%s\nwant State: spec (draft)", got.stdout)
		}
	})

	t.Run("a damaged line after the unknown event", func(t *testing.T) {
		home := newHome(t)
		repo := newProject(t, home)
		createSlice(t, repo, home, "login")
		appendText(t, sliceLog(repo, "login"),
			eventLine(2, "slice.edited", `{"title":"Sign in"}`)+
				eventLine(3, "spec.submitted", `{}`)+
				eventLine(4, "slice.edited", `{"title":"Log in"}`)+
				"cut off\n")

		got := runSliceflow(t, repo, home, "slice", "status", "login")

		wantCode(t, got, 0)
		wantStderr := "Warning: line 5 of docs/slices/login/.events.jsonl could not be read, so it was skipped.\n" +
			"Warning: line 3 of docs/slices/login/.events.jsonl has the event spec.submitted, which this version of sliceflow does not know, so the state cannot be worked out.\n"
		if got.stderr != wantStderr {
			t.Errorf("stderr is:\n%s\nwant:\n%s", got.stderr, wantStderr)
		}
	})
}

// slice status shows the title and description after every slice.edited
// that can be read, even after an event that makes the state unknown.
func TestStatusShowsEdits(t *testing.T) {
	t.Run("an edit", func(t *testing.T) {
		home := newHome(t)
		repo := newProject(t, home)
		createSlice(t, repo, home, "login", "--title", "Log in")
		appendText(t, sliceLog(repo, "login"), eventLine(2, "slice.edited", `{"title":"Sign in"}`))

		got := runSliceflow(t, repo, home, "slice", "status", "login")

		wantCode(t, got, 0)
		if first, _, _ := strings.Cut(got.stdout, "\n"); first != "login: Sign in" {
			t.Errorf("first line is %q, want %q", first, "login: Sign in")
		}
	})

	t.Run("an edit after an unknown event", func(t *testing.T) {
		home := newHome(t)
		repo := newProject(t, home)
		createSlice(t, repo, home, "login", "--title", "Log in")
		appendText(t, sliceLog(repo, "login"),
			eventLine(2, "spec.submitted", `{}`)+
				eventLine(3, "slice.edited", `{"title":"Sign in"}`))

		got := runSliceflow(t, repo, home, "slice", "status", "login")

		wantCode(t, got, 0)
		if first, _, _ := strings.Cut(got.stdout, "\n"); first != "login: Sign in" {
			t.Errorf("first line is %q, want %q", first, "login: Sign in")
		}
		if !strings.Contains(got.stdout, "  State:        unknown\n") {
			t.Errorf("stdout is:\n%s\nwant State: unknown", got.stdout)
		}
	})
}

// eventLine returns one line of a log, with its line break: an event
// with the seq, the name event, and data, a JSON object.
func eventLine(seq int, event, data string) string {
	return fmt.Sprintf(`{"version":1,"seq":%d,"event":"%s","at":"2026-11-18T10:00:00Z","data":%s}`+"\n", seq, event, data)
}

// secondCreatedLine returns the first line of the log of login in the
// checkout at repo, with seq 2: a second slice.created that can be read.
func secondCreatedLine(t *testing.T, repo string) string {
	t.Helper()
	data, err := os.ReadFile(sliceLog(repo, "login"))
	if err != nil {
		t.Fatalf("could not read the log: %v", err)
	}
	first, _, _ := strings.Cut(string(data), "\n")
	return strings.Replace(first, `"seq":1`, `"seq":2`, 1) + "\n"
}

// specFile returns the full path of the spec of the slice name in the
// checkout at repo, as sliceflow prints it.
func specFile(t *testing.T, repo, home, name string) string {
	t.Helper()
	return filepath.Join(topFolder(t, repo, home), "docs", "slices", name, "spec.md")
}
