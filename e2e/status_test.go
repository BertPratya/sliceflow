package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// With its worktree there, slice status reads the log from the worktree
// and prints no Read from line. A name with capital letters is made
// lowercase, with no Note.
func TestStatusWithItsWorktree(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)
	base := gitOutput(t, repo, home, "rev-parse", "main")
	createSlice(t, repo, home, "login", "--title", "Log in", "--description", "Lets a user log in.")

	got := runSliceflow(t, repo, home, "slice", "status", "Login")

	want := "login: Log in\n" +
		"  Description:  Lets a user log in.\n" +
		"  Created:      " + createdTime(t, home, "login") + "\n" +
		"  Started from: main, at " + base[:7] + "\n" +
		"  Branch:       slice/login\n" +
		"  Worktree:     " + defaultWorktree(home, "login") + "\n" +
		"  Events:       1 (last: slice.created)\n"
	wantOutput(t, got, want, "")
}

// With the worktree removed, slice status reads the log from the branch.
func TestStatusFromTheBranch(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)
	base := gitOutput(t, repo, home, "rev-parse", "main")
	createSlice(t, repo, home, "login", "--title", "Log in")
	created := createdTime(t, home, "login")
	removeWorktree(t, repo, home, "login")

	got := runSliceflow(t, repo, home, "slice", "status", "login")

	want := "login: Log in\n" +
		"  Description:  (none)\n" +
		"  Created:      " + created + "\n" +
		"  Started from: main, at " + base[:7] + "\n" +
		"  Branch:       slice/login\n" +
		"  Worktree:     (not found)\n" +
		"  Events:       1 (last: slice.created)\n" +
		"  Read from:    branch slice/login\n"
	wantOutput(t, got, want, "")
}

// Merged into main, with its branch and worktree gone, slice status reads
// the log from main.
func TestStatusFromMain(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)
	base := gitOutput(t, repo, home, "rev-parse", "main")
	createSlice(t, repo, home, "login", "--title", "Log in")
	created := createdTime(t, home, "login")
	removeWorktree(t, repo, home, "login")
	mergeIntoMain(t, repo, home, "login", true)

	got := runSliceflow(t, repo, home, "slice", "status", "login")

	want := "login: Log in\n" +
		"  Description:  (none)\n" +
		"  Created:      " + created + "\n" +
		"  Started from: main, at " + base[:7] + "\n" +
		"  Branch:       slice/login (not found)\n" +
		"  Worktree:     (not found)\n" +
		"  Events:       1 (last: slice.created)\n" +
		"  Read from:    main\n"
	wantOutput(t, got, want, "")
}

// With no log anywhere, slice status names the places it looked.
func TestStatusOfNoSlice(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)

	got := runSliceflow(t, repo, home, "slice", "status", "login")
	wantError(t, got, 1, "Error: no slice named login. Nothing was changed.\n"+
		"No log was found in any of these places:\n"+
		"  worktree  "+defaultWorktree(home, "login")+"\n"+
		"  branch    slice/login\n"+
		"  main      docs/slices/login/.events.jsonl\n"+
		"Check the name, or run: sliceflow slice list\n")
}

// Lines of the log that cannot be read are skipped, with one warning.
func TestStatusWithBadLines(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)
	createSlice(t, repo, home, "login")
	appendText(t, worktreeLog(home, "login"), "cut off\n{\n")

	got := runSliceflow(t, repo, home, "slice", "status", "login")
	wantCode(t, got, 0)
	wantWarning := "Warning: lines 2, 3 of docs/slices/login/.events.jsonl could not be read, so they were skipped.\n"
	if got.stderr != wantWarning {
		t.Errorf("stderr is %q, want %q", got.stderr, wantWarning)
	}
	if !strings.Contains(got.stdout, "  Events:       1 (last: slice.created)\n") {
		t.Errorf("stdout is:\n%s\nwant 1 event", got.stdout)
	}
}

// A log whose first line cannot be read has no slice.created event.
func TestStatusWithNoCreatedEvent(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)
	createSlice(t, repo, home, "login")
	writeText(t, worktreeLog(home, "login"), "{\"version\":1,\"seq\n")

	got := runSliceflow(t, repo, home, "slice", "status", "login")
	wantError(t, got, 1, "Warning: line 1 of docs/slices/login/.events.jsonl could not be read, so it was skipped.\n"+
		"Error: no slice.created event could be read for login. Nothing was changed.\n"+
		"Fix or restore the log, then try again.\n")
}

// slice status checks the name like slice create, with exit code 2.
func TestStatusOfAnInvalidName(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)

	got := runSliceflow(t, repo, home, "slice", "status", "a.b")
	wantError(t, got, 2, "Error: invalid slice name \"a.b\". Nothing was changed.\n"+
		"Allowed characters: letters, digits, hyphens, underscores. Start with a letter or a digit.\n")
}

// createSlice runs slice create with name and options in repo, and fails
// the test now if it does not work.
func createSlice(t *testing.T, repo, home, name string, options ...string) {
	t.Helper()
	args := append([]string{"slice", "create", name}, options...)
	wantCode(t, runSliceflow(t, repo, home, args...), 0)
}

// removeWorktree removes the worktree of the slice name, and keeps its
// branch.
func removeWorktree(t *testing.T, repo, home, name string) {
	t.Helper()
	gitOutput(t, repo, home, "worktree", "remove", "--force", defaultWorktree(home, name))
}

// mergeIntoMain merges the branch of the slice name into main, which the
// project has checked out, and then deletes the branch when deleteBranch
// is true.
func mergeIntoMain(t *testing.T, repo, home, name string, deleteBranch bool) {
	t.Helper()
	gitOutput(t, repo, home, "merge", "--quiet", "--ff-only", "slice/"+name)
	if deleteBranch {
		gitOutput(t, repo, home, "branch", "--quiet", "-d", "slice/"+name)
	}
}

// worktreeLog returns the path of the log in the worktree of the slice
// name.
func worktreeLog(home, name string) string {
	return filepath.Join(defaultWorktree(home, name), "docs", "slices", name, ".events.jsonl")
}

// createdTime returns the time of the slice.created event of the slice
// name, from the log in its worktree, as slice status prints it.
func createdTime(t *testing.T, home, name string) string {
	t.Helper()
	data, err := os.ReadFile(worktreeLog(home, name))
	if err != nil {
		t.Fatalf("could not read the log: %v", err)
	}
	first, _, _ := strings.Cut(string(data), "\n")
	var event struct{ At string }
	if err := json.Unmarshal([]byte(first), &event); err != nil {
		t.Fatalf("could not read the first event: %v", err)
	}
	at, err := time.Parse(time.RFC3339, event.At)
	if err != nil {
		t.Fatalf("could not read the time %q: %v", event.At, err)
	}
	return at.UTC().Format("2006-01-02 15:04") + " UTC"
}

// appendText adds text at the end of the file path.
func appendText(t *testing.T, path, text string) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatalf("could not open %s: %v", path, err)
	}
	if _, err := f.WriteString(text); err != nil {
		f.Close()
		t.Fatalf("could not write to %s: %v", path, err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("could not close %s: %v", path, err)
	}
}

// wantOutput fails the test unless sliceflow exited with 0 and printed
// exactly stdout and stderr.
func wantOutput(t *testing.T, got result, stdout, stderr string) {
	t.Helper()
	wantCode(t, got, 0)
	if got.stdout != stdout {
		t.Errorf("stdout is:\n%s\nwant:\n%s", got.stdout, stdout)
	}
	if got.stderr != stderr {
		t.Errorf("stderr is:\n%s\nwant:\n%s", got.stderr, stderr)
	}
}
