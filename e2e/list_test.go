package e2e

import (
	"path/filepath"
	"testing"
)

// slice list shows every slice, wherever its log is: in its worktree, on
// its branch, or in main. The table is sorted by name, and a row with no
// description ends after the title.
func TestListSlices(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)
	createSlice(t, repo, home, "old-search", "--title", "Search by name")
	removeWorktree(t, repo, home, "old-search")
	mergeIntoMain(t, repo, home, "old-search", true)
	createSlice(t, repo, home, "login-v2", "--title", "Login with a passkey",
		"--description", "Replace the password login with a passkey.")
	removeWorktree(t, repo, home, "login-v2")
	createSlice(t, repo, home, "card-payment", "--title", "Pay with a card",
		"--description", "Let a customer pay with a credit or debit card.")

	got := runSliceflow(t, repo, home, "slice", "list")

	want := "NAME          TITLE                 DESCRIPTION\n" +
		"card-payment  Pay with a card       Let a customer pay with a credit or debit card.\n" +
		"login-v2      Login with a passkey  Replace the password login with a passkey.\n" +
		"old-search    Search by name\n"
	wantOutput(t, got, want, "")
}

// A project with no slices gets a line that says how to make one.
func TestListNoSlices(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)

	got := runSliceflow(t, repo, home, "slice", "list")
	wantOutput(t, got, "No slices yet. To make one, run: sliceflow slice create <name>\n", "")
}

// A slice merged into main whose branch is still there is listed once.
func TestListMergedWithItsBranch(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)
	createSlice(t, repo, home, "login", "--title", "Log in")
	removeWorktree(t, repo, home, "login")
	mergeIntoMain(t, repo, home, "login", false)

	got := runSliceflow(t, repo, home, "slice", "list")
	wantOutput(t, got, "NAME   TITLE   DESCRIPTION\nlogin  Log in\n", "")
}

// Names that look like slices but are not are left out, each with one
// warning, sorted by name, by character code: a branch with no log, a
// folder in main with no log, both, and names that are not valid.
func TestListSkipsCandidates(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)
	writeText(t, filepath.Join(repo, "docs", "slices", "bar", "notes.md"), "notes\n")
	writeText(t, filepath.Join(repo, "docs", "slices", "baz", "notes.md"), "notes\n")
	gitOutput(t, repo, home, "add", ".")
	gitOutput(t, repo, home, "commit", "--quiet", "-m", "folders with no log")
	for _, branch := range []string{"slice/foo", "slice/baz", "slice/Bad", "slice/x/y"} {
		gitOutput(t, repo, home, "branch", branch)
	}
	createSlice(t, repo, home, "card-payment", "--title", "Pay with a card")

	got := runSliceflow(t, repo, home, "slice", "list")

	wantStderr := "Warning: Bad is not listed as a slice. It is not a valid slice name.\n" +
		"Warning: bar is not listed as a slice. The folder docs/slices/bar/ in main has no event log.\n" +
		"Warning: baz is not listed as a slice. Neither the branch slice/baz nor the folder docs/slices/baz/ in main has an event log.\n" +
		"Warning: foo is not listed as a slice. The branch slice/foo has no event log.\n" +
		"Warning: x/y is not listed as a slice. It is not a valid slice name.\n"
	wantOutput(t, got, "NAME          TITLE            DESCRIPTION\ncard-payment  Pay with a card\n", wantStderr)
}

// A log whose first line cannot be read leaves its slice out, with a
// warning. A bad line later in a log is not mentioned: the list reads
// only the first line.
func TestListDamagedLogs(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)
	createSlice(t, repo, home, "card-payment", "--title", "Pay with a card")
	createSlice(t, repo, home, "login-v2", "--title", "Login with a passkey")
	writeText(t, worktreeLog(home, "card-payment"), "{\"version\":1,\"seq\n")
	appendText(t, worktreeLog(home, "login-v2"), "cut off\n")

	got := runSliceflow(t, repo, home, "slice", "list")

	wantStderr := "Warning: card-payment is not listed as a slice. No slice.created event could be read from docs/slices/card-payment/.events.jsonl.\n"
	wantOutput(t, got, "NAME      TITLE                 DESCRIPTION\nlogin-v2  Login with a passkey\n", wantStderr)
}
