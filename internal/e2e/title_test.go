package e2e

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/BertPratya/sliceflow/internal/eventlog"
	"github.com/BertPratya/sliceflow/internal/filelock"
	"github.com/BertPratya/sliceflow/internal/gittest"
)

// titledLine matches the slice.titled line of 0002 S15, with the given seq
// and title.
func titledLine(seq, title string) *regexp.Regexp {
	return regexp.MustCompile(`^\{"version":1,"seq":` + seq + `,"event":"slice\.titled",` +
		`"at":"\d{4}-\d\d-\d\dT\d\d:\d\d:\d\dZ","data":\{"title":"` + regexp.QuoteMeta(title) + `"\}\}$`)
}

// logLines returns the lines of a slice's log in its worktree, without
// the \n at the end.
func logLines(t *testing.T, worktree, name string) []string {
	t.Helper()
	return strings.Split(strings.TrimSuffix(readFile(t, filepath.ToSlash(logPath(worktree, name))), "\n"), "\n")
}

// TestTitleS15: the title is set, only the log is committed, and status
// shows the new title (0002 R29, R32).
func TestTitleS15(t *testing.T) {
	repo := gittest.NewRepo(t)
	worktree := createSlice(t, repo, "spec-submit")
	spec := filepath.Join(worktree, "docs", "slices", "spec-submit", "spec.md")
	if err := os.WriteFile(spec, []byte("changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	before := gittest.Run(t, repo, "rev-parse", "slice/spec-submit")

	r := run(t, repo, "slice", "title", "spec-submit", "Submit a spec for review")
	want := "Title of spec-submit set: Submit a spec for review\n"
	if r.code != 0 || r.stderr != "" || r.stdout != want {
		t.Fatalf("code = %d, stderr = %q, stdout = %q, want 0, empty, %q", r.code, r.stderr, r.stdout, want)
	}

	if got := gittest.Run(t, repo, "rev-parse", "slice/spec-submit~1"); got != before {
		t.Errorf("want one new commit on slice/spec-submit")
	}
	if got := gittest.Run(t, repo, "log", "-1", "--format=%s", "slice/spec-submit"); got != "spec-submit: title set" {
		t.Errorf("commit message = %q", got)
	}
	if got := gittest.Run(t, repo, "show", "--name-only", "--format=", "slice/spec-submit"); got != "docs/slices/spec-submit/.events.jsonl" {
		t.Errorf("files in the commit = %q, want only the log", got)
	}
	lines := logLines(t, worktree, "spec-submit")
	if len(lines) != 2 || !titledLine("2", "Submit a spec for review").MatchString(lines[1]) {
		t.Errorf("log lines = %q, want a slice.titled line 2", lines)
	}
	if got := gittest.Run(t, worktree, "status", "--porcelain"); got != "M docs/slices/spec-submit/spec.md" {
		t.Errorf("git status = %q, want only spec.md changed, and no lock file", got)
	}

	status := run(t, repo, "slice", "status", "spec-submit")
	if !strings.Contains(status.stdout, "Title:    Submit a spec for review\n") {
		t.Errorf("status stdout = %q, want the new title", status.stdout)
	}
	list := run(t, repo, "slice", "list")
	if !strings.Contains(list.stdout, "Submit a spec for review") {
		t.Errorf("list stdout = %q, want the new title", list.stdout)
	}
}

// TestTitleNewestWins: a second title replaces the first, also over the
// title of slice create (0002 R32).
func TestTitleNewestWins(t *testing.T) {
	repo := gittest.NewRepo(t)
	createSlice(t, repo, "spec-submit", "--title", "First")
	for _, title := range []string{"Second", "Third"} {
		if r := run(t, repo, "slice", "title", "spec-submit", title); r.code != 0 {
			t.Fatalf("slice title %s: %q", title, r.stderr)
		}
	}

	r := run(t, repo, "slice", "status", "spec-submit")
	if !strings.Contains(r.stdout, "Title:    Third\n") || r.stderr != "" {
		t.Errorf("stdout = %q, stderr = %q, want the title Third and no warnings", r.stdout, r.stderr)
	}
}

// TestTitleJSON: --json prints the slice and the saved title (0002 R30, R31).
func TestTitleJSON(t *testing.T) {
	repo := gittest.NewRepo(t)
	createSlice(t, repo, "spec-submit")

	r := run(t, repo, "slice", "title", "spec-submit", "  Login  ", "--json")
	want := `{"slice":"spec-submit","title":"Login"}` + "\n"
	if r.code != 0 || r.stderr != "" || r.stdout != want {
		t.Errorf("code = %d, stderr = %q, stdout = %q, want 0, empty, %q", r.code, r.stderr, r.stdout, want)
	}
}

// TestTitleErrors: S16 and the errors of R6 and R7, which change nothing.
func TestTitleErrors(t *testing.T) {
	repo := gittest.NewRepo(t)
	worktree := createSlice(t, repo, "spec-submit")
	createSlice(t, repo, "no-tree")
	gittest.Run(t, repo, "worktree", "remove", filepath.Join(repo, ".sliceflow", "worktrees", "no-tree"))
	before := readFile(t, filepath.ToSlash(logPath(worktree, "spec-submit")))

	tests := []struct {
		name string
		args []string
		want string
	}{
		{"S16 empty title", []string{"spec-submit", "  "}, "Error: the title is empty. Nothing was changed.\n"},
		{"R6 no slice", []string{"nope", "T"}, "Error: no slice named nope. See all slices: sliceflow slice list\n"},
		{"R6 invalid name", []string{"Bad_Name", "T"}, "Error: no slice named Bad_Name. See all slices: sliceflow slice list\n"},
		{"R7 no worktree", []string{"no-tree", "T"}, "Error: slice no-tree has no worktree: branch slice/no-tree is not checked out anywhere.\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := run(t, repo, append([]string{"slice", "title"}, tt.args...)...)
			if r.code != 1 || r.stdout != "" || r.stderr != tt.want {
				t.Errorf("code = %d, stdout = %q, stderr = %q, want 1, empty, %q", r.code, r.stdout, r.stderr, tt.want)
			}
		})
	}
	if got := readFile(t, filepath.ToSlash(logPath(worktree, "spec-submit"))); got != before {
		t.Errorf("the log changed")
	}
}

// TestTitleWrongUse: a missing name or title is wrong use (0002 R31).
func TestTitleWrongUse(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"no name", nil, "missing slice name"},
		{"empty name", []string{""}, "missing slice name"},
		{"no title", []string{"spec-submit"}, "missing title"},
		{"too many", []string{"spec-submit", "a", "b"}, "too many arguments"},
		{"unknown flag", []string{"spec-submit", "a", "--colour"}, "unknown flag --colour"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := run(t, t.TempDir(), append([]string{"slice", "title"}, tt.args...)...)
			want := "Error: " + tt.want + ". Run sliceflow --help.\n"
			if r.code != 2 || r.stdout != "" || r.stderr != want {
				t.Errorf("code = %d, stdout = %q, stderr = %q, want 2, empty, %q", r.code, r.stdout, r.stderr, want)
			}
		})
	}
}

// TestTitleS17: a half-written last line is removed with a warning, and
// the new event takes line 2 (0002 R33).
func TestTitleS17(t *testing.T) {
	repo := gittest.NewRepo(t)
	worktree := createSlice(t, repo, "spec-submit")
	f, err := os.OpenFile(logPath(worktree, "spec-submit"), os.O_WRONLY|os.O_APPEND, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(`{"version":1,"seq":2,"ev`); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	r := run(t, repo, "slice", "title", "spec-submit", "Submit a spec")
	wantErr := "Warning: spec-submit: .events.jsonl line 2: half-written last line; removed\n"
	if r.code != 0 || r.stderr != wantErr || r.stdout != "Title of spec-submit set: Submit a spec\n" {
		t.Fatalf("code = %d, stderr = %q, stdout = %q, want 0 and the warning", r.code, r.stderr, r.stdout)
	}
	lines := logLines(t, worktree, "spec-submit")
	if len(lines) != 2 || !titledLine("2", "Submit a spec").MatchString(lines[1]) {
		t.Errorf("log lines = %q, want the slice.titled line on line 2", lines)
	}
}

// TestTitleS18: another sliceflow holds the lock, so after 10 seconds the
// command gives up and changes nothing (0002 R34).
func TestTitleS18(t *testing.T) {
	if testing.Short() {
		t.Skip("waits 10 seconds")
	}
	repo := gittest.NewRepo(t)
	worktree := createSlice(t, repo, "spec-submit")
	sliceDir := filepath.Join(worktree, "docs", "slices", "spec-submit")
	before := readFile(t, filepath.ToSlash(logPath(worktree, "spec-submit")))

	held, err := filelock.Lock(eventlog.LockPath(sliceDir))
	if err != nil {
		t.Fatal(err)
	}
	defer filelock.Unlock(held)

	r := run(t, repo, "slice", "title", "spec-submit", "Submit a spec")
	want := "Error: the event log of spec-submit is busy: another sliceflow is writing to it. Try again. Nothing was changed."
	if filelock.StaleLockBlocks {
		want += " If no sliceflow is running, delete " + eventlog.LockPath(sliceDir) + "."
	}
	if r.code != 1 || r.stdout != "" || r.stderr != want+"\n" {
		t.Errorf("code = %d, stdout = %q, stderr = %q, want 1, empty, %q", r.code, r.stdout, r.stderr, want)
	}
	if got := readFile(t, filepath.ToSlash(logPath(worktree, "spec-submit"))); got != before {
		t.Errorf("the log changed while busy")
	}
}

// TestTitleS19: the commit fails, so the log is put back (0002 R35).
func TestTitleS19(t *testing.T) {
	repo := gittest.NewRepo(t)
	worktree := createSlice(t, repo, "spec-submit")
	before := readFile(t, filepath.ToSlash(logPath(worktree, "spec-submit")))
	hook := filepath.Join(repo, ".git", "hooks", "pre-commit")
	if err := os.MkdirAll(filepath.Dir(hook), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(hook, []byte("#!/bin/sh\necho 'error: commits are blocked' >&2\nexit 1\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	r := run(t, repo, "slice", "title", "spec-submit", "Submit a spec")
	want := "Error: could not set the title of spec-submit: error: commits are blocked. The event log was put back as it was.\n"
	if r.code != 1 || r.stdout != "" || r.stderr != want {
		t.Errorf("code = %d, stdout = %q, stderr = %q, want 1, empty, %q", r.code, r.stdout, r.stderr, want)
	}
	if got := readFile(t, filepath.ToSlash(logPath(worktree, "spec-submit"))); got != before {
		t.Errorf("the log was not put back:\ngot  %q\nwant %q", got, before)
	}
}
