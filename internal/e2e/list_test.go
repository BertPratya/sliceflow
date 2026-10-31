package e2e

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/BertPratya/sliceflow/internal/gittest"
)

// logTimePattern finds the time of an event line.
var logTimePattern = regexp.MustCompile(`"at":"[^"]*"`)

// setCreated changes the time in the one line of a slice's log to at (like
// "2026-10-24T07:42:00Z"), and commits that in the slice's worktree.
func setCreated(t *testing.T, worktree, name, at string) {
	t.Helper()
	path := logPath(worktree, name)
	line := readFile(t, filepath.ToSlash(path))
	changed := logTimePattern.ReplaceAllString(line, `"at":"`+at+`"`)
	if changed == line {
		t.Fatalf("cannot change the time in %q", line)
	}
	if err := os.WriteFile(path, []byte(changed), 0o644); err != nil {
		t.Fatal(err)
	}
	gittest.Run(t, worktree, "add", "-A")
	gittest.Run(t, worktree, "commit", "-m", "set time")
}

// listRepo makes the repo of S20: spec-submit and login-google in
// progress, and first-slice merged into main (its branch and worktree are
// removed).
func listRepo(t *testing.T) string {
	t.Helper()
	repo := gittest.NewRepo(t)

	setCreated(t, createSlice(t, repo, "spec-submit"), "spec-submit", "2026-10-24T07:42:00Z")
	setCreated(t, createSlice(t, repo, "login-google", "--title", "Login with Google"), "login-google", "2026-10-24T08:10:00Z")

	first := createSlice(t, repo, "first-slice")
	setCreated(t, first, "first-slice", "2026-10-24T07:50:00Z")
	gittest.Run(t, repo, "merge", "slice/first-slice")
	gittest.Run(t, repo, "worktree", "remove", "--force", first)
	gittest.Run(t, repo, "branch", "-D", "slice/first-slice")
	return repo
}

const listWant = "NAME          TITLE              PHASE     CREATED (UTC)\n" +
	"spec-submit   spec-submit        planning  2026-10-24 07:42\n" +
	"login-google  Login with Google  planning  2026-10-24 08:10\n"

// TestS20: the table of slices in progress; the merged one is left out.
func TestS20(t *testing.T) {
	repo := listRepo(t)

	r := run(t, repo, "slice", "list")
	if r.code != 0 || r.stderr != "" || r.stdout != listWant {
		t.Errorf("code = %d, stderr = %q, stdout = %q, want 0, empty, %q", r.code, r.stderr, r.stdout, listWant)
	}
}

// TestS21: --all adds the merged slice, at its place by created time.
func TestS21(t *testing.T) {
	repo := listRepo(t)

	want := "NAME          TITLE              PHASE     CREATED (UTC)\n" +
		"spec-submit   spec-submit        planning  2026-10-24 07:42\n" +
		"first-slice   first-slice        planning  2026-10-24 07:50\n" +
		"login-google  Login with Google  planning  2026-10-24 08:10\n"
	r := run(t, repo, "slice", "list", "--all")
	if r.code != 0 || r.stderr != "" || r.stdout != want {
		t.Errorf("code = %d, stderr = %q, stdout = %q, want 0, empty, %q", r.code, r.stderr, r.stdout, want)
	}
}

// TestS22: a bad log line gives " !" on that row only, and a warning.
func TestS22(t *testing.T) {
	repo := listRepo(t)
	worktree := filepath.Join(repo, ".sliceflow", "worktrees", "spec-submit")
	path := logPath(worktree, "spec-submit")
	good := readFile(t, filepath.ToSlash(path))
	if err := os.WriteFile(path, []byte(good+"this is not json\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gittest.Run(t, worktree, "add", "-A")
	gittest.Run(t, worktree, "commit", "-m", "bad log")

	want := "NAME          TITLE              PHASE       CREATED (UTC)\n" +
		"spec-submit   spec-submit        planning !  2026-10-24 07:42\n" +
		"login-google  Login with Google  planning    2026-10-24 08:10\n"
	r := run(t, repo, "slice", "list")
	wantErr := "Warning: spec-submit: .events.jsonl line 2: not valid JSON; skipped\n"
	if r.code != 0 || r.stderr != wantErr || r.stdout != want {
		t.Errorf("code = %d, stderr = %q, stdout = %q, want 0, %q, %q", r.code, r.stderr, r.stdout, wantErr, want)
	}
}

// TestS23: --json, one object with slices and warnings. Each slice holds
// state and next after phase (0002 R26).
func TestS23(t *testing.T) {
	repo := listRepo(t)

	r := run(t, repo, "slice", "list", "--json")
	re := regexp.MustCompile(`^\{"slices":\[\{"slice":"spec-submit","phase":"planning",` +
		`"state":"spec_draft","next":"write the spec, then check it: sliceflow spec validate spec-submit",` +
		`"from":\{"branch":"main","commit":"[0-9a-f]{40}"\},"created":"2026-10-24T07:42:00Z","warnings":\[\]\},` +
		`\{"slice":"login-google","title":"Login with Google","phase":"planning",` +
		`"state":"spec_draft","next":"write the spec, then check it: sliceflow spec validate login-google",` +
		`.*"created":"2026-10-24T08:10:00Z","warnings":\[\]\}\],"warnings":\[\]\}\n$`)
	if r.code != 0 || r.stderr != "" || !re.MatchString(r.stdout) {
		t.Errorf("code = %d, stderr = %q, stdout = %q, does not match %v", r.code, r.stderr, r.stdout, re)
	}
}

// TestS24: nothing in progress.
func TestS24(t *testing.T) {
	repo := gittest.NewRepo(t)

	r := run(t, repo, "slice", "list")
	if r.code != 0 || r.stderr != "" || r.stdout != "No slices in progress.\n" {
		t.Errorf("code = %d, stderr = %q, stdout = %q, want 0, empty, the no-slices line", r.code, r.stderr, r.stdout)
	}

	r = run(t, repo, "slice", "list", "--json")
	if r.code != 0 || r.stdout != "{\"slices\":[],\"warnings\":[]}\n" {
		t.Errorf("--json: code = %d, stdout = %q, want empty lists", r.code, r.stdout)
	}
}

// TestListAllNothing: --all with no slices.
func TestListAllNothing(t *testing.T) {
	repo := gittest.NewRepo(t)

	r := run(t, repo, "slice", "list", "--all")
	if r.code != 0 || r.stderr != "" || r.stdout != "No slices.\n" {
		t.Errorf("code = %d, stderr = %q, stdout = %q, want 0, empty, \"No slices.\"", r.code, r.stderr, r.stdout)
	}
}

// TestS25: the base branch main does not exist (the repo uses master).
func TestS25(t *testing.T) {
	repo := gittest.NewRepo(t)
	gittest.Run(t, repo, "branch", "-m", "main", "master")
	createSlice(t, repo, "spec-submit", "--from", "master")

	r := run(t, repo, "slice", "list")
	wantErr := "Warning: base branch main does not exist (the default); merged slices are not shown\n"
	re := regexp.MustCompile(`^NAME         TITLE        PHASE     CREATED \(UTC\)\n` +
		`spec-submit  spec-submit  planning  \d{4}-\d\d-\d\d \d\d:\d\d\n$`)
	if r.code != 0 || r.stderr != wantErr || !re.MatchString(r.stdout) {
		t.Errorf("code = %d, stderr = %q, stdout = %q, want 0, %q, the row of spec-submit", r.code, r.stderr, r.stdout, wantErr)
	}

	// The repo warning is in the JSON too, and not on a slice.
	r = run(t, repo, "slice", "list", "--json")
	if !regexp.MustCompile(`"warnings":\[\]\}\],"warnings":\["base branch main does not exist \(the default\); merged slices are not shown"\]\}\n$`).MatchString(r.stdout) {
		t.Errorf("--json stdout = %q, want the repo warning at the top level only", r.stdout)
	}
}

// TestS27: an unknown flag.
func TestS27(t *testing.T) {
	repo := gittest.NewRepo(t)

	r := run(t, repo, "slice", "list", "--colour")
	want := "Error: unknown flag --colour. Run sliceflow --help.\n"
	if r.code != 2 || r.stdout != "" || r.stderr != want {
		t.Errorf("code = %d, stdout = %q, stderr = %q, want 2, empty, %q", r.code, r.stdout, r.stderr, want)
	}
}

// TestS28: outside a git repo.
func TestS28(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(dir))

	r := run(t, dir, "slice", "list")
	want := "Error: not inside a git repository. Nothing was changed.\n"
	if r.code != 1 || r.stdout != "" || r.stderr != want {
		t.Errorf("code = %d, stdout = %q, stderr = %q, want 1, empty, %q", r.code, r.stdout, r.stderr, want)
	}
}

// TestS29: an invalid name for create.
func TestS29(t *testing.T) {
	repo := gittest.NewRepo(t)

	r := run(t, repo, "slice", "create", "My-Slice", "--json")
	want := "Error: invalid slice name \"My-Slice\": use 1-64 lowercase letters, digits, and dashes, " +
		"starting with a letter or digit. Nothing was changed.\n"
	if r.code != 1 || r.stdout != "" || r.stderr != want {
		t.Errorf("code = %d, stdout = %q, stderr = %q, want 1, empty, %q", r.code, r.stdout, r.stderr, want)
	}
}

// TestListInvalidBranches: R43. Two branches with bad names are left out,
// each warning is printed once.
func TestListInvalidBranches(t *testing.T) {
	repo := gittest.NewRepo(t)
	gittest.Run(t, repo, "branch", "slice/My_Thing")
	gittest.Run(t, repo, "branch", "slice/team/login")

	r := run(t, repo, "slice", "list")
	wantErr := "Warning: branch slice/My_Thing: not a valid slice name; skipped\n" +
		"Warning: branch slice/team/login: not a valid slice name; skipped\n"
	if r.code != 0 || r.stderr != wantErr || r.stdout != "No slices in progress.\n" {
		t.Errorf("code = %d, stderr = %q, stdout = %q, want 0, %q, the no-slices line", r.code, r.stderr, r.stdout, wantErr)
	}

	r = run(t, repo, "slice", "list", "--json")
	wantJSON := "{\"slices\":[],\"warnings\":[\"branch slice/My_Thing: not a valid slice name; skipped\"," +
		"\"branch slice/team/login: not a valid slice name; skipped\"]}\n"
	if r.stdout != wantJSON {
		t.Errorf("--json stdout = %q, want %q", r.stdout, wantJSON)
	}
}

// TestListTooManyArguments: slice list takes no names.
func TestListTooManyArguments(t *testing.T) {
	repo := gittest.NewRepo(t)

	r := run(t, repo, "slice", "list", "x")
	want := "Error: too many arguments. Run sliceflow --help.\n"
	if r.code != 2 || r.stdout != "" || r.stderr != want {
		t.Errorf("code = %d, stdout = %q, stderr = %q, want 2, empty, %q", r.code, r.stdout, r.stderr, want)
	}
}

// TestListMergedWithBranch: a merged slice whose branch still exists is
// left out, and --all shows it.
func TestListMergedWithBranch(t *testing.T) {
	repo := gittest.NewRepo(t)
	createSlice(t, repo, "done")
	gittest.Run(t, repo, "merge", "slice/done")

	r := run(t, repo, "slice", "list")
	if r.code != 0 || r.stderr != "" || r.stdout != "No slices in progress.\n" {
		t.Errorf("list: code = %d, stderr = %q, stdout = %q, want the no-slices line", r.code, r.stderr, r.stdout)
	}

	r = run(t, repo, "slice", "list", "--all")
	if r.code != 0 || r.stderr != "" || !regexp.MustCompile(`\ndone  done   planning  \d{4}-`).MatchString(r.stdout) {
		t.Errorf("list --all: code = %d, stderr = %q, stdout = %q, want a row for done", r.code, r.stderr, r.stdout)
	}
}

// TestListAllMissingBase: --all with no base branch gives the R23 warning
// and still lists the local slices.
func TestListAllMissingBase(t *testing.T) {
	repo := gittest.NewRepo(t)
	gittest.Run(t, repo, "branch", "-m", "main", "master")
	createSlice(t, repo, "spec-submit", "--from", "master")

	r := run(t, repo, "slice", "list", "--all")
	wantErr := "Warning: base branch main does not exist (the default); merged slices are not shown\n"
	if r.code != 0 || r.stderr != wantErr || !regexp.MustCompile(`\nspec-submit  spec-submit  planning  `).MatchString(r.stdout) {
		t.Errorf("code = %d, stderr = %q, stdout = %q, want 0, %q, a row for spec-submit", r.code, r.stderr, r.stdout, wantErr)
	}
}
