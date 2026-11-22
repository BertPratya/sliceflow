package e2e

import (
	"path/filepath"
	"testing"
)

// slice list shows every slice in the user's checkout. The table is
// sorted by name, and a row with no description ends after the title.
func TestListSlices(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)
	createSlice(t, repo, home, "old-search", "--title", "Search by name")
	createSlice(t, repo, home, "login-v2", "--title", "Login with a passkey",
		"--description", "Replace the password login with a passkey.")
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

// slice list reads only the user's checkout: a slice planned on another
// branch is not listed, and gets no warning.
func TestListReadsOnlyYourCheckout(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)
	gitOutput(t, repo, home, "config", "sliceflow.base", "dev/bert")
	gitOutput(t, repo, home, "switch", "--quiet", "-c", "dev/bert")
	createSlice(t, repo, home, "login")
	gitOutput(t, repo, home, "switch", "--quiet", "main")

	got := runSliceflow(t, repo, home, "slice", "list")
	wantOutput(t, got, "No slices yet. To make one, run: sliceflow slice create <name>\n", "")
}

// Folders in docs/slices/ that are not slices are left out, each with one
// warning, sorted by name, by character code: folders whose names are
// not valid, and folders with no log, committed or not.
func TestListSkipsCandidates(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)
	writeText(t, filepath.Join(repo, "docs", "slices", "bar", "notes.md"), "notes\n")
	gitOutput(t, repo, home, "add", ".")
	gitOutput(t, repo, home, "commit", "--quiet", "-m", "a folder with no log")
	for _, name := range []string{"baz", "Bad", "a.b"} {
		writeText(t, filepath.Join(repo, "docs", "slices", name, "notes.md"), "notes\n")
	}
	createSlice(t, repo, home, "card-payment", "--title", "Pay with a card")

	got := runSliceflow(t, repo, home, "slice", "list")

	wantStderr := "Warning: Bad is not listed as a slice. It is not a valid slice name.\n" +
		"Warning: a.b is not listed as a slice. It is not a valid slice name.\n" +
		"Warning: bar is not listed as a slice. The folder docs/slices/bar/ has no event log.\n" +
		"Warning: baz is not listed as a slice. The folder docs/slices/baz/ has no event log.\n"
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
	writeText(t, sliceLog(repo, "card-payment"), "{\"version\":1,\"seq\n")
	appendText(t, sliceLog(repo, "login-v2"), "cut off\n")

	got := runSliceflow(t, repo, home, "slice", "list")

	wantStderr := "Warning: card-payment is not listed as a slice. No slice.created event could be read from docs/slices/card-payment/.events.jsonl.\n"
	wantOutput(t, got, "NAME      TITLE                 DESCRIPTION\nlogin-v2  Login with a passkey\n", wantStderr)
}

// When the checkout does not hold every commit of the base branch of some
// slices, slice list gives one warning before the table, which says how
// many slices.
func TestListOldCopies(t *testing.T) {
	tests := []struct {
		about      string
		names      []string
		detach     bool
		wantStdout string
		wantStderr string
	}{
		{"two slices", []string{"login", "report"}, false,
			"NAME    TITLE   DESCRIPTION\nlogin   login\nreport  report\n",
			"Warning: your checkout is on branch fix/new, but 2 slices are planned on other branches, so they may be out of date.\n"},
		{"one slice", []string{"login"}, false,
			"NAME   TITLE  DESCRIPTION\nlogin  login\n",
			"Warning: your checkout is on branch fix/new, but 1 slice is planned on another branch, so it may be out of date.\n"},
		{"on no branch", []string{"login"}, true,
			"NAME   TITLE  DESCRIPTION\nlogin  login\n",
			"Warning: your checkout is not on a branch, but 1 slice is planned on another branch, so it may be out of date.\n"},
	}
	for _, tt := range tests {
		t.Run(tt.about, func(t *testing.T) {
			home := newHome(t)
			repo := newProject(t, home)
			newOldCopy(t, repo, home, tt.names...)
			if tt.detach {
				gitOutput(t, repo, home, "checkout", "--quiet", "--detach", "fix/new")
			}

			got := runSliceflow(t, repo, home, "slice", "list")
			wantOutput(t, got, tt.wantStdout, tt.wantStderr)
		})
	}
}
