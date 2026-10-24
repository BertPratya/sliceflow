package e2e

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/BertPratya/sliceflow/internal/testrepo"
)

var atRe = regexp.MustCompile(`"at":"[^"]+"`)

// setCreated commits the log of a slice again, with the given "at" in each
// event line (the created time of the scenario).
func setCreated(t *testing.T, dir, name, at string) {
	t.Helper()
	wt := filepath.Join(dir, ".sliceflow", "worktrees", name)
	logFile := filepath.Join(wt, "docs", "slices", name, ".events.jsonl")
	b, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatal(err)
	}
	out := atRe.ReplaceAllString(string(b), `"at":"`+at+`"`)
	if err := os.WriteFile(logFile, []byte(out), 0o644); err != nil {
		t.Fatal(err)
	}
	testrepo.Git(t, wt, "commit", "-am", "set created time")
}

// mergeAndRemove merges the slice into main, then removes its worktree and
// branch, as finishing a slice does.
func mergeAndRemove(t *testing.T, dir, name string) {
	t.Helper()
	testrepo.Git(t, dir, "merge", "--no-ff", "-m", "merge "+name, "slice/"+name)
	testrepo.Git(t, dir, "worktree", "remove", "--force", filepath.Join(dir, ".sliceflow", "worktrees", name))
	testrepo.Git(t, dir, "branch", "-D", "slice/"+name)
}

// s20Repo makes the repo of S20: spec-submit and login-google in progress,
// first-slice merged.
func s20Repo(t *testing.T) (dir, commit string) {
	t.Helper()
	dir = testrepo.New(t)
	createSlice(t, dir, "first-slice")
	setCreated(t, dir, "first-slice", "2026-10-24T07:00:00Z")
	mergeAndRemove(t, dir, "first-slice")
	// The other slices start from main after the merge.
	commit = strings.TrimSpace(testrepo.Git(t, dir, "rev-parse", "HEAD"))
	createSlice(t, dir, "spec-submit")
	setCreated(t, dir, "spec-submit", "2026-10-24T07:42:00Z")
	createSlice(t, dir, "login-google", "--title", "Login with Google")
	setCreated(t, dir, "login-google", "2026-10-24T08:10:00Z")
	return dir, commit
}

const s20Table = "NAME          TITLE              PHASE     CREATED (UTC)\n" +
	"spec-submit   spec-submit        planning  2026-10-24 07:42\n" +
	"login-google  Login with Google  planning  2026-10-24 08:10\n"

func TestListInProgress(t *testing.T) {
	dir, _ := s20Repo(t)
	out, errOut, code := run(t, dir, "slice", "list")
	if code != 0 || errOut != "" || out != s20Table {
		t.Fatalf("code %d\nstdout %q\nwant   %q\nstderr %q", code, out, s20Table, errOut)
	}
}

func TestListAll(t *testing.T) {
	dir, _ := s20Repo(t)
	out, errOut, code := run(t, dir, "slice", "list", "--all")
	want := "NAME          TITLE              PHASE     CREATED (UTC)\n" +
		"first-slice   first-slice        planning  2026-10-24 07:00\n" +
		"spec-submit   spec-submit        planning  2026-10-24 07:42\n" +
		"login-google  Login with Google  planning  2026-10-24 08:10\n"
	if code != 0 || errOut != "" || out != want {
		t.Fatalf("code %d\nstdout %q\nwant   %q\nstderr %q", code, out, want, errOut)
	}
}

func TestListBadLine(t *testing.T) {
	dir, _ := s20Repo(t)
	wt := filepath.Join(dir, ".sliceflow", "worktrees", "spec-submit")
	logFile := filepath.Join(wt, "docs", "slices", "spec-submit", ".events.jsonl")
	b, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatal(err)
	}
	line1 := strings.TrimSuffix(string(b), "\n")
	line3 := strings.Replace(line1, `"seq":1`, `"seq":4`, 1)
	if line3 == line1 {
		t.Fatalf("no seq in %q", line1)
	}
	if err := os.WriteFile(logFile, []byte(line1+"\nthis is not json\n"+line3+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	testrepo.Git(t, wt, "commit", "-am", "bad lines")

	out, errOut, code := run(t, dir, "slice", "list")
	want := "NAME          TITLE              PHASE       CREATED (UTC)\n" +
		"spec-submit   spec-submit        planning !  2026-10-24 07:42\n" +
		"login-google  Login with Google  planning    2026-10-24 08:10\n"
	wantErr := "Warning: spec-submit: .events.jsonl line 2: not valid JSON; skipped\n" +
		"Warning: spec-submit: .events.jsonl line 3: seq 4, expected 2\n"
	if code != 0 || out != want || errOut != wantErr {
		t.Fatalf("code %d\nstdout %q\nwant   %q\nstderr %q\nwant   %q", code, out, want, errOut, wantErr)
	}
}

func TestListJSON(t *testing.T) {
	dir, commit := s20Repo(t)
	out, errOut, code := run(t, dir, "slice", "list", "--json")
	want := `{"slices":[` +
		`{"slice":"spec-submit","phase":"planning","from":{"branch":"main","commit":"` + commit + `"},"created":"2026-10-24T07:42:00Z","warnings":[]},` +
		`{"slice":"login-google","title":"Login with Google","phase":"planning","from":{"branch":"main","commit":"` + commit + `"},"created":"2026-10-24T08:10:00Z","warnings":[]}` +
		`],"warnings":[]}` + "\n"
	if code != 0 || errOut != "" || out != want {
		t.Fatalf("code %d\nstdout %q\nwant   %q\nstderr %q", code, out, want, errOut)
	}
}

func TestListEmpty(t *testing.T) {
	dir := testrepo.New(t)
	cases := []struct {
		args []string
		want string
	}{
		{[]string{"slice", "list"}, "No slices in progress.\n"},
		{[]string{"slice", "list", "--all"}, "No slices.\n"},
		{[]string{"slice", "list", "--json"}, `{"slices":[],"warnings":[]}` + "\n"},
	}
	for _, c := range cases {
		out, errOut, code := run(t, dir, c.args...)
		if code != 0 || errOut != "" || out != c.want {
			t.Errorf("%v: code %d\nstdout %q\nwant   %q\nstderr %q", c.args, code, out, c.want, errOut)
		}
	}
}

func TestListBaseMissing(t *testing.T) {
	dir := testrepo.New(t)
	createSlice(t, dir, "spec-submit")
	at := createdAt(t, dir, "spec-submit")
	testrepo.Git(t, dir, "branch", "-m", "main", "master")
	out, errOut, code := run(t, dir, "slice", "list")
	want := "NAME         TITLE        PHASE     CREATED (UTC)\n" +
		"spec-submit  spec-submit  planning  " + at.Format("2006-01-02 15:04") + "\n"
	wantErr := "Warning: base branch main does not exist (the default); merged slices are not shown\n"
	if code != 0 || out != want || errOut != wantErr {
		t.Fatalf("code %d\nstdout %q\nwant   %q\nstderr %q\nwant   %q", code, out, want, errOut, wantErr)
	}
}

func TestListUnknownFlag(t *testing.T) {
	dir := testrepo.New(t)
	out, errOut, code := run(t, dir, "slice", "list", "--colour")
	if code != 2 || out != "" || errOut != "Error: unknown flag --colour. Run sliceflow --help.\n" {
		t.Fatalf("code %d out %q err %q", code, out, errOut)
	}
	out, errOut, code = run(t, dir, "slice", "list", "some-name")
	if code != 2 || out != "" || errOut != "Error: too many arguments. Run sliceflow --help.\n" {
		t.Fatalf("name: code %d out %q err %q", code, out, errOut)
	}
}

func TestListNotARepo(t *testing.T) {
	testrepo.Setup(t)
	dir := t.TempDir()
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(dir))
	out, errOut, code := run(t, dir, "slice", "list")
	if code != 1 || out != "" || errOut != "Error: not inside a git repository. Nothing was changed.\n" {
		t.Fatalf("code %d out %q err %q", code, out, errOut)
	}
}

func TestListErrorJSON(t *testing.T) {
	testrepo.Setup(t)
	dir := t.TempDir()
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(dir))
	out, errOut, code := run(t, dir, "slice", "list", "--json")
	if code != 1 || out != "" || errOut != "Error: not inside a git repository. Nothing was changed.\n" {
		t.Fatalf("code %d out %q err %q", code, out, errOut)
	}
}

func TestListBadBranchName(t *testing.T) {
	dir := testrepo.New(t)
	createSlice(t, dir, "spec-submit")
	testrepo.Git(t, dir, "branch", "slice/My_Thing")
	testrepo.Git(t, dir, "branch", "slice/team/login")
	w1 := "branch slice/My_Thing: not a valid slice name; skipped"
	w2 := "branch slice/team/login: not a valid slice name; skipped"
	wantErr := "Warning: " + w1 + "\nWarning: " + w2 + "\n"
	out, errOut, code := run(t, dir, "slice", "list")
	if code != 0 || errOut != wantErr || strings.Contains(out, "My_Thing") || strings.Contains(out, "login") ||
		strings.Contains(out, "!") || strings.Count(out, "\n") != 2 {
		t.Fatalf("code %d\nstdout %q\nstderr %q\nwant   %q", code, out, errOut, wantErr)
	}
	out, errOut, code = run(t, dir, "slice", "list", "--json")
	if code != 0 || errOut != wantErr || !strings.HasSuffix(out, `],"warnings":["`+w1+`","`+w2+`"]}`+"\n") {
		t.Fatalf("json: code %d\nstdout %q\nstderr %q", code, out, errOut)
	}
	if strings.Contains(out, `"slice":"My_Thing"`) || !strings.Contains(out, `"warnings":[]}`) {
		t.Fatalf("json slices: %q", out)
	}
}

// task-12-1/F1: run from a subfolder, the list is the same as from the top.
func TestListFromSubfolder(t *testing.T) {
	dir, _ := s20Repo(t)
	folders := map[string]string{
		"main checkout":  filepath.Join(dir, "docs"),
		"slice worktree": filepath.Join(dir, ".sliceflow", "worktrees", "spec-submit", "docs"),
	}
	for name, sub := range folders {
		for _, args := range [][]string{{"slice", "list"}, {"slice", "list", "--all"}} {
			wantOut, wantErr, wantCode := run(t, dir, args...)
			out, errOut, code := run(t, sub, args...)
			if code != wantCode || out != wantOut || errOut != wantErr {
				t.Errorf("%s, %v: from a subfolder\ncode %d, want %d\nstdout %q\nwant   %q\nstderr %q",
					name, args, code, wantCode, out, wantOut, errOut)
			}
		}
	}
}
