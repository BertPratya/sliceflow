package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// slice status prints the slice's lines, read from the log in the user's
// checkout. A name with capital letters is made lowercase, with no Note.
func TestStatusOfASlice(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)
	createSlice(t, repo, home, "login", "--title", "Log in", "--description", "Lets a user log in.")

	got := runSliceflow(t, repo, home, "slice", "status", "Login")

	want := "login: Log in\n" +
		"  State:        spec (draft)\n" +
		"  Description:  Lets a user log in.\n" +
		"  Created:      " + createdTime(t, repo, "login") + "\n" +
		"  Base:         main\n" +
		"  Events:       1 (last: slice.created)\n" +
		nextSteps(t, repo, home, "login")
	wantOutput(t, got, want, "")
}

// slice status works from a subfolder of the checkout: it reads the log
// at the top of the checkout.
func TestStatusFromASubfolder(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)
	createSlice(t, repo, home, "login", "--title", "Log in")
	sub := filepath.Join(repo, "internal", "cli")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	got := runSliceflow(t, sub, home, "slice", "status", "login")

	want := "login: Log in\n" +
		"  State:        spec (draft)\n" +
		"  Description:  (none)\n" +
		"  Created:      " + createdTime(t, repo, "login") + "\n" +
		"  Base:         main\n" +
		"  Events:       1 (last: slice.created)\n" +
		nextSteps(t, repo, home, "login")
	wantOutput(t, got, want, "")
}

// In a second worktree that the user made, slice status reads the log in
// that worktree. The main checkout does not have the slice, so there
// slice status finds none: it never reads another worktree.
func TestStatusInASecondWorktree(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)
	second := addWorktree(t, repo, home, "dev/bert")
	gitOutput(t, repo, home, "config", "sliceflow.base", "dev/bert")
	createSlice(t, second, home, "login", "--title", "Log in")

	got := runSliceflow(t, second, home, "slice", "status", "login")
	want := "login: Log in\n" +
		"  State:        spec (draft)\n" +
		"  Description:  (none)\n" +
		"  Created:      " + createdTime(t, second, "login") + "\n" +
		"  Base:         dev/bert\n" +
		"  Events:       1 (last: slice.created)\n" +
		nextSteps(t, second, home, "login")
	wantOutput(t, got, want, "")

	got = runSliceflow(t, repo, home, "slice", "status", "login")
	wantError(t, got, 1, "Error: no slice named login. Nothing was changed.\n"+
		"No log was found at docs/slices/login/.events.jsonl.\n"+
		"Your checkout is on branch main. Check the name, or switch to the branch the slice is planned on.\n")
}

// With no log in the checkout, slice status says so, and names the branch
// the checkout is on. It never looks on other branches: here the slice is
// planned on dev/bert, and the checkout is on fix/old, made before the
// slice.
func TestStatusOfNoSlice(t *testing.T) {
	t.Run("on a branch", func(t *testing.T) {
		home := newHome(t)
		repo := newProject(t, home)
		gitOutput(t, repo, home, "config", "sliceflow.base", "dev/bert")
		gitOutput(t, repo, home, "branch", "fix/old")
		gitOutput(t, repo, home, "switch", "--quiet", "-c", "dev/bert")
		createSlice(t, repo, home, "login")
		gitOutput(t, repo, home, "switch", "--quiet", "fix/old")

		got := runSliceflow(t, repo, home, "slice", "status", "login")
		wantError(t, got, 1, "Error: no slice named login. Nothing was changed.\n"+
			"No log was found at docs/slices/login/.events.jsonl.\n"+
			"Your checkout is on branch fix/old. Check the name, or switch to the branch the slice is planned on.\n")
	})
	t.Run("on no branch", func(t *testing.T) {
		home := newHome(t)
		repo := newProject(t, home)
		gitOutput(t, repo, home, "checkout", "--quiet", "--detach", "main")

		got := runSliceflow(t, repo, home, "slice", "status", "login")
		wantError(t, got, 1, "Error: no slice named login. Nothing was changed.\n"+
			"No log was found at docs/slices/login/.events.jsonl.\n"+
			"Your checkout is not on a branch. Check the name, or switch to the branch the slice is planned on.\n")
	})
}

// When the checkout holds an old copy of a slice, here from fix/new made
// before dev/bert, the base, got one more commit, slice status shows that
// copy as it is, with no warning about other branches: the user decides
// which branch to look at. The only warning is the one for lines that
// cannot be read.
func TestStatusOfAnOldCopy(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)
	newOldCopy(t, repo, home, "login")
	appendText(t, sliceLog(repo, "login"), "cut off\n")

	got := runSliceflow(t, repo, home, "slice", "status", "login")

	wantStdout := "login: login\n" +
		"  State:        spec (draft)\n" +
		"  Description:  (none)\n" +
		"  Created:      " + createdTime(t, repo, "login") + "\n" +
		"  Base:         dev/bert\n" +
		"  Events:       1 (last: slice.created)\n" +
		nextSteps(t, repo, home, "login")
	wantStderr := "Warning: line 2 of docs/slices/login/.events.jsonl could not be read, so it was skipped.\n"
	wantOutput(t, got, wantStdout, wantStderr)
}

// Lines of the log that cannot be read are skipped, with one warning.
func TestStatusWithBadLines(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)
	createSlice(t, repo, home, "login")
	appendText(t, sliceLog(repo, "login"), "cut off\n{\n")

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
	writeText(t, sliceLog(repo, "login"), "{\"version\":1,\"seq\n")

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

// newOldCopy makes the checkout at repo hold an old copy of the slices
// names: it creates them on dev/bert, the base branch, makes fix/new
// from dev/bert, gives dev/bert one more commit, and leaves fix/new
// checked out. Each slice's title is its name.
func newOldCopy(t *testing.T, repo, home string, names ...string) {
	t.Helper()
	gitOutput(t, repo, home, "config", "sliceflow.base", "dev/bert")
	gitOutput(t, repo, home, "switch", "--quiet", "-c", "dev/bert")
	for _, name := range names {
		createSlice(t, repo, home, name)
	}
	gitOutput(t, repo, home, "branch", "fix/new")
	gitOutput(t, repo, home, "commit", "--quiet", "--allow-empty", "-m", "more on dev/bert")
	gitOutput(t, repo, home, "switch", "--quiet", "fix/new")
}

// nextSteps returns the empty line and the next step that slice status
// prints for the slice name in spec (draft), in the checkout at repo.
func nextSteps(t *testing.T, repo, home, name string) string {
	t.Helper()
	return "\n" +
		"Next: write the spec in " + specFile(t, repo, home, name) + "\n" +
		"Then run: sliceflow spec validate " + name + "\n"
}

// sliceLog returns the path of the log of the slice name in the checkout
// at repo.
func sliceLog(repo, name string) string {
	return filepath.Join(repo, "docs", "slices", name, ".events.jsonl")
}

// createdTime returns the time of the slice.created event of the slice
// name, from the log in the checkout at repo, as slice status prints it.
func createdTime(t *testing.T, repo, name string) string {
	t.Helper()
	data, err := os.ReadFile(sliceLog(repo, name))
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
