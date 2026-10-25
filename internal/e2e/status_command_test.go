package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/BertPratya/sliceflow/internal/testrepo"
)

// createSlice runs "slice create" and fails the test when it does not work.
func createSlice(t *testing.T, dir string, args ...string) {
	t.Helper()
	all := append([]string{"slice", "create"}, args...)
	if _, e, c := run(t, dir, all...); c != 0 {
		t.Fatalf("create %v: code %d, stderr %q", args, c, e)
	}
}

// createdAt reads the "at" of the first event of a slice from its branch.
func createdAt(t *testing.T, dir, name string) time.Time {
	t.Helper()
	ev := testrepo.Git(t, dir, "show", "slice/"+name+":docs/slices/"+name+"/.events.jsonl")
	m := regexp.MustCompile(`"at":"([^"]+)"`).FindStringSubmatch(ev)
	if m == nil {
		t.Fatalf("no at in %q", ev)
	}
	at, err := time.Parse("2006-01-02T15:04:05Z", m[1])
	if err != nil {
		t.Fatal(err)
	}
	return at
}

func statusLines(slice, title, phase, from, created string) string {
	return statusLinesSpec(slice, title, phase, from, created, "draft", "write spec.md, then: sliceflow spec submit")
}

func statusLinesSpec(slice, title, phase, from, created, spec, next string) string {
	return "Slice:    " + slice + "\n" +
		"Title:    " + title + "\n" +
		"Phase:    " + phase + "\n" +
		"From:     " + from + "\n" +
		"Created:  " + created + "\n" +
		"Spec:     " + spec + "\n" +
		"Next:     " + next + "\n"
}

func TestStatusSpecSubmit(t *testing.T) {
	dir := testrepo.New(t)
	commit := strings.TrimSpace(testrepo.Git(t, dir, "rev-parse", "HEAD"))
	createSlice(t, dir, "spec-submit")
	at := createdAt(t, dir, "spec-submit")
	out, errOut, code := run(t, dir, "slice", "status", "spec-submit")
	want := statusLines("spec-submit", "spec-submit", "planning", "main @ "+commit[:7], at.Format("2006-01-02 15:04")+" UTC")
	if code != 0 || errOut != "" || out != want {
		t.Fatalf("code %d\nstdout %q\nwant   %q\nstderr %q", code, out, want, errOut)
	}
}

func TestStatusTitle(t *testing.T) {
	dir := testrepo.New(t)
	createSlice(t, dir, "spec-submit", "--title", "Submit the spec")
	out, errOut, code := run(t, dir, "slice", "status", "spec-submit")
	if code != 0 || errOut != "" || !strings.Contains(out, "Title:    Submit the spec\n") {
		t.Fatalf("code %d stdout %q stderr %q", code, out, errOut)
	}
}

func TestStatusCurrentBranch(t *testing.T) {
	dir := testrepo.New(t)
	commit := strings.TrimSpace(testrepo.Git(t, dir, "rev-parse", "HEAD"))
	createSlice(t, dir, "spec-submit")
	at := createdAt(t, dir, "spec-submit")
	inside := filepath.Join(dir, ".sliceflow", "worktrees", "spec-submit")
	out, errOut, code := run(t, inside, "slice", "status")
	want := statusLines("spec-submit", "spec-submit", "planning", "main @ "+commit[:7], at.Format("2006-01-02 15:04")+" UTC")
	if code != 0 || errOut != "" || out != want {
		t.Fatalf("code %d\nstdout %q\nwant   %q\nstderr %q", code, out, want, errOut)
	}
}

func TestStatusNotOnSlice(t *testing.T) {
	dir := testrepo.New(t)
	want := func(where string) string {
		return "Error: not on a slice branch (you are " + where + ").\n" +
			"Give a name: sliceflow slice status <name>\n" +
			"See all slices: sliceflow slice list\n"
	}
	out, errOut, code := run(t, dir, "slice", "status")
	if code != 1 || out != "" || errOut != want("on main") {
		t.Fatalf("code %d\nstdout %q\nstderr %q\nwant   %q", code, out, errOut, want("on main"))
	}
	testrepo.Git(t, dir, "checkout", "--detach")
	out, errOut, code = run(t, dir, "slice", "status")
	if code != 1 || out != "" || errOut != want("on no branch") {
		t.Fatalf("detached: code %d\nstdout %q\nstderr %q\nwant   %q", code, out, errOut, want("on no branch"))
	}
}

func TestStatusJSON(t *testing.T) {
	dir := testrepo.New(t)
	testrepo.Git(t, dir, "branch", "develop")
	testrepo.Git(t, dir, "config", "sliceflow.base", "develop")
	commit := strings.TrimSpace(testrepo.Git(t, dir, "rev-parse", "develop"))
	createSlice(t, dir, "login-google", "--title", "Login with Google")
	at := createdAt(t, dir, "login-google")
	out, errOut, code := run(t, dir, "slice", "status", "login-google", "--json")
	want := `{"slice":"login-google","title":"Login with Google","phase":"planning",` +
		`"from":{"branch":"develop","commit":"` + commit + `"},` +
		`"created":"` + at.Format("2006-01-02T15:04:05Z") + `",` +
		`"docs":[{"doc":"spec.md","status":"draft"}],"next":"write spec.md, then: sliceflow spec submit","warnings":[]}` + "\n"
	if code != 0 || errOut != "" || out != want {
		t.Fatalf("code %d\nstdout %q\nwant   %q\nstderr %q", code, out, want, errOut)
	}
}

func TestStatusJSONWarnings(t *testing.T) {
	dir := testrepo.New(t)
	testrepo.Git(t, dir, "branch", "slice/old-idea")
	testrepo.Git(t, dir, "config", "sliceflow.base", "develop")
	out, errOut, code := run(t, dir, "slice", "status", "old-idea", "--json")
	want := `{"slice":"old-idea","phase":"unknown","docs":[],"warnings":[` +
		`"base branch develop does not exist (from git config sliceflow.base); merged slices are not shown",` +
		`".events.jsonl not found"]}` + "\n"
	wantErr := "Warning: base branch develop does not exist (from git config sliceflow.base); merged slices are not shown\n" +
		"Warning: old-idea: .events.jsonl not found\n"
	if code != 0 || out != want || errOut != wantErr {
		t.Fatalf("code %d\nstdout %q\nwant   %q\nstderr %q\nwant   %q", code, out, want, errOut, wantErr)
	}
}

func TestStatusBadLines(t *testing.T) {
	dir := testrepo.New(t)
	commit := strings.TrimSpace(testrepo.Git(t, dir, "rev-parse", "HEAD"))
	createSlice(t, dir, "spec-submit")
	at := createdAt(t, dir, "spec-submit")
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
	content := line1 + "\nthis is not json\n" + line3 + "\n"
	if err := os.WriteFile(logFile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	testrepo.Git(t, wt, "commit", "-am", "bad lines")

	out, errOut, code := run(t, dir, "slice", "status", "spec-submit")
	want := statusLines("spec-submit", "spec-submit", "planning", "main @ "+commit[:7], at.Format("2006-01-02 15:04")+" UTC")
	wantErr := "Warning: spec-submit: .events.jsonl line 2: not valid JSON; skipped\n" +
		"Warning: spec-submit: .events.jsonl line 3: seq 4, expected 2\n" +
		"Warning: spec-submit: .events.jsonl line 3: slice.created is not allowed when spec.md is draft; skipped\n"
	if code != 0 || out != want || errOut != wantErr {
		t.Fatalf("code %d\nstdout %q\nwant   %q\nstderr %q\nwant   %q", code, out, want, errOut, wantErr)
	}

	out, errOut, code = run(t, dir, "slice", "status", "spec-submit", "--json")
	var got struct {
		Warnings []string `json:"warnings"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("json: %v: %q", err, out)
	}
	if code != 0 || errOut != wantErr || len(got.Warnings) != 3 ||
		got.Warnings[0] != ".events.jsonl line 2: not valid JSON; skipped" ||
		got.Warnings[1] != ".events.jsonl line 3: seq 4, expected 2" ||
		got.Warnings[2] != ".events.jsonl line 3: slice.created is not allowed when spec.md is draft; skipped" {
		t.Fatalf("code %d json %q stderr %q", code, out, errOut)
	}
}

func TestStatusNoLog(t *testing.T) {
	dir := testrepo.New(t)
	testrepo.Git(t, dir, "branch", "slice/old-idea")
	out, errOut, code := run(t, dir, "slice", "status", "old-idea")
	want := statusLinesSpec("old-idea", "old-idea", "unknown", "-", "-", "-", "-")
	if code != 0 || out != want || errOut != "Warning: old-idea: .events.jsonl not found\n" {
		t.Fatalf("code %d\nstdout %q\nwant   %q\nstderr %q", code, out, want, errOut)
	}
}

func TestStatusNoSuchSlice(t *testing.T) {
	dir := testrepo.New(t)
	out, errOut, code := run(t, dir, "slice", "status", "no-such-slice")
	want := "Error: no slice named no-such-slice. See all slices: sliceflow slice list\n"
	if code != 1 || out != "" || errOut != want {
		t.Fatalf("code %d\nstdout %q\nstderr %q\nwant   %q", code, out, errOut, want)
	}
}

func TestStatusErrorJSON(t *testing.T) {
	dir := testrepo.New(t)
	out, errOut, code := run(t, dir, "slice", "status", "no-such-slice", "--json")
	want := "Error: no slice named no-such-slice. See all slices: sliceflow slice list\n"
	if code != 1 || out != "" || errOut != want {
		t.Fatalf("code %d\nstdout %q\nstderr %q\nwant   %q", code, out, errOut, want)
	}
}

func TestStatusNotARepo(t *testing.T) {
	testrepo.Setup(t)
	dir := t.TempDir()
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(dir))
	out, errOut, code := run(t, dir, "slice", "status", "spec-submit")
	if code != 1 || out != "" || errOut != "Error: not inside a git repository. Nothing was changed.\n" {
		t.Fatalf("code %d out %q err %q", code, out, errOut)
	}
}

func TestStatusWrongUse(t *testing.T) {
	dir := testrepo.New(t)
	out, errOut, code := run(t, dir, "slice", "status", "a", "b")
	if code != 2 || out != "" || errOut != "Error: too many arguments. Run sliceflow --help.\n" {
		t.Fatalf("code %d out %q err %q", code, out, errOut)
	}
	out, errOut, code = run(t, dir, "slice", "status", "--help")
	if code != 0 || errOut != "" || !strings.Contains(out, "--json") {
		t.Fatalf("help: code %d out %q err %q", code, out, errOut)
	}
}
