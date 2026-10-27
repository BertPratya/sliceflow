package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/BertPratya/sliceflow/internal/gittest"
)

// createSlice runs "slice create" and stops the test if it fails. It
// returns the worktree folder of the slice.
func createSlice(t *testing.T, repo, name string, extra ...string) string {
	t.Helper()
	args := append([]string{"slice", "create", name}, extra...)
	if r := run(t, repo, args...); r.code != 0 {
		t.Fatalf("slice create failed: %q", r.stderr)
	}
	return filepath.Join(repo, ".sliceflow", "worktrees", name)
}

// logPath returns the path of a slice's event log inside its worktree.
func logPath(worktree, name string) string {
	return filepath.Join(worktree, "docs", "slices", name, ".events.jsonl")
}

// TestS13: the plain status of a new slice, from main.
func TestS13(t *testing.T) {
	repo := gittest.NewRepo(t)
	short := gittest.Run(t, repo, "rev-parse", "main")[:7]
	createSlice(t, repo, "spec-submit")

	r := run(t, repo, "slice", "status", "spec-submit")
	if r.code != 0 || r.stderr != "" {
		t.Fatalf("code = %d, stderr = %q, want 0, empty", r.code, r.stderr)
	}
	want := regexp.MustCompile(`^Slice:    spec-submit\n` +
		`Title:    spec-submit\n` +
		`Phase:    planning\n` +
		`From:     main @ ` + short + `\n` +
		`Created:  \d{4}-\d\d-\d\d \d\d:\d\d UTC\n$`)
	if !want.MatchString(r.stdout) {
		t.Errorf("stdout = %q, does not match %v", r.stdout, want)
	}
}

// TestS14: no name, inside the worktree of the slice.
func TestS14(t *testing.T) {
	repo := gittest.NewRepo(t)
	worktree := createSlice(t, repo, "spec-submit")

	byName := run(t, repo, "slice", "status", "spec-submit")
	r := run(t, worktree, "slice", "status")
	if r.code != 0 || r.stderr != "" || r.stdout != byName.stdout {
		t.Errorf("code = %d, stderr = %q, stdout = %q, want the status of spec-submit %q",
			r.code, r.stderr, r.stdout, byName.stdout)
	}
}

// TestS15: no name, and not on a slice branch.
func TestS15(t *testing.T) {
	repo := gittest.NewRepo(t)

	r := run(t, repo, "slice", "status")
	want := "Error: not on a slice branch (you are on main).\n" +
		"Give a name: sliceflow slice status <name>\n" +
		"See all slices: sliceflow slice list\n"
	if r.code != 1 || r.stdout != "" || r.stderr != want {
		t.Errorf("code = %d, stdout = %q, stderr = %q, want 1, empty, %q", r.code, r.stdout, r.stderr, want)
	}
}

// TestStatusDetachedHead: R32 with no branch.
func TestStatusDetachedHead(t *testing.T) {
	repo := gittest.NewRepo(t)
	gittest.Run(t, repo, "checkout", "--detach")

	r := run(t, repo, "slice", "status")
	want := "Error: not on a slice branch (you are on no branch).\n" +
		"Give a name: sliceflow slice status <name>\n" +
		"See all slices: sliceflow slice list\n"
	if r.code != 1 || r.stdout != "" || r.stderr != want {
		t.Errorf("code = %d, stdout = %q, stderr = %q, want 1, empty, %q", r.code, r.stdout, r.stderr, want)
	}
}

// TestS16: --json, a base from git config, and a title.
func TestS16(t *testing.T) {
	repo := gittest.NewRepo(t)
	gittest.Run(t, repo, "branch", "develop")
	gittest.Run(t, repo, "config", "sliceflow.base", "develop")
	full := gittest.Run(t, repo, "rev-parse", "develop")
	createSlice(t, repo, "login-google", "--title", "Login with Google")

	r := run(t, repo, "slice", "status", "login-google", "--json")
	if r.code != 0 || r.stderr != "" {
		t.Fatalf("code = %d, stderr = %q, want 0, empty", r.code, r.stderr)
	}
	var got struct {
		Slice string `json:"slice"`
		Title string `json:"title"`
		Phase string `json:"phase"`
		From  struct {
			Branch string `json:"branch"`
			Commit string `json:"commit"`
		} `json:"from"`
		Created  string   `json:"created"`
		Warnings []string `json:"warnings"`
	}
	if err := json.Unmarshal([]byte(r.stdout), &got); err != nil {
		t.Fatalf("stdout %q is not JSON: %v", r.stdout, err)
	}
	if got.Slice != "login-google" || got.Title != "Login with Google" || got.Phase != "planning" {
		t.Errorf("slice, title, phase = %q, %q, %q", got.Slice, got.Title, got.Phase)
	}
	if got.From.Branch != "develop" || got.From.Commit != full {
		t.Errorf("from = %+v, want develop, %s", got.From, full)
	}
	if !atPattern.MatchString(got.Created) {
		t.Errorf("created = %q, want UTC to the second", got.Created)
	}
	if got.Warnings == nil || len(got.Warnings) != 0 {
		t.Errorf("warnings = %#v, want []", got.Warnings)
	}
	// The key order is part of R30, and warnings is [] in the text.
	order := regexp.MustCompile(`^\{"slice":.*,"title":.*,"phase":.*,"from":.*,"created":.*,"warnings":\[\]\}\n$`)
	if !order.MatchString(r.stdout) {
		t.Errorf("stdout = %q, fields are not in the order of R30", r.stdout)
	}
}

// TestS17: a bad line and a wrong seq give warnings, and exit 0.
func TestS17(t *testing.T) {
	repo := gittest.NewRepo(t)
	worktree := createSlice(t, repo, "spec-submit")

	path := logPath(worktree, "spec-submit")
	good := readFile(t, filepath.ToSlash(path))
	// Line 3 is the good line again, with seq 4.
	line3 := regexp.MustCompile(`"seq":1`).ReplaceAllString(good, `"seq":4`)
	if line3 == good {
		t.Fatalf("cannot change seq in %q", good)
	}
	if err := os.WriteFile(path, []byte(good+"this is not json\n"+line3), 0o644); err != nil {
		t.Fatal(err)
	}
	gittest.Run(t, worktree, "add", "-A")
	gittest.Run(t, worktree, "commit", "-m", "bad log")

	r := run(t, repo, "slice", "status", "spec-submit")
	want := "Warning: spec-submit: .events.jsonl line 2: not valid JSON; skipped\n" +
		"Warning: spec-submit: .events.jsonl line 3: seq 4, expected 2\n"
	if r.code != 0 || r.stderr != want {
		t.Errorf("code = %d, stderr = %q, want 0, %q", r.code, r.stderr, want)
	}
	if !regexp.MustCompile(`Phase:    planning\n`).MatchString(r.stdout) {
		t.Errorf("stdout = %q, want the status from the good lines", r.stdout)
	}

	// With --json the warnings are texts only, and also go to stderr.
	j := run(t, repo, "slice", "status", "spec-submit", "--json")
	var got struct {
		Warnings []string `json:"warnings"`
	}
	if err := json.Unmarshal([]byte(j.stdout), &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Warnings) != 2 || got.Warnings[0] != ".events.jsonl line 2: not valid JSON; skipped" {
		t.Errorf("warnings = %q", got.Warnings)
	}
	if j.stderr != want {
		t.Errorf("stderr with --json = %q, want %q", j.stderr, want)
	}
}

// TestS18: a slice branch with no log.
func TestS18(t *testing.T) {
	repo := gittest.NewRepo(t)
	gittest.Run(t, repo, "branch", "slice/old-idea")

	r := run(t, repo, "slice", "status", "old-idea")
	want := "Slice:    old-idea\n" +
		"Title:    old-idea\n" +
		"Phase:    unknown\n" +
		"From:     -\n" +
		"Created:  -\n"
	if r.code != 0 || r.stdout != want || r.stderr != "Warning: old-idea: .events.jsonl not found\n" {
		t.Errorf("code = %d, stdout = %q, stderr = %q", r.code, r.stdout, r.stderr)
	}

	j := run(t, repo, "slice", "status", "old-idea", "--json")
	wantJSON := `{"slice":"old-idea","phase":"unknown","warnings":[".events.jsonl not found"]}` + "\n"
	if j.stdout != wantJSON {
		t.Errorf("json stdout = %q, want %q", j.stdout, wantJSON)
	}
}

// TestS19: no such slice.
func TestS19(t *testing.T) {
	repo := gittest.NewRepo(t)

	for _, name := range []string{"no-such-slice", "Bad Name"} {
		want := "Error: no slice named " + name + ". See all slices: sliceflow slice list\n"
		r := run(t, repo, "slice", "status", name)
		if r.code != 1 || r.stdout != "" || r.stderr != want {
			t.Errorf("%q: code = %d, stdout = %q, stderr = %q, want 1, empty, %q", name, r.code, r.stdout, r.stderr, want)
		}
	}
}

// TestStatusNotARepo: R22, outside a repo.
func TestStatusNotARepo(t *testing.T) {
	r := run(t, t.TempDir(), "slice", "status", "x")
	want := "Error: not inside a git repository. Nothing was changed.\n"
	if r.code != 1 || r.stdout != "" || r.stderr != want {
		t.Errorf("code = %d, stdout = %q, stderr = %q, want 1, empty, %q", r.code, r.stdout, r.stderr, want)
	}
}

// TestStatusNoCommits: R22, a repo with no commits.
func TestStatusNoCommits(t *testing.T) {
	repo := gittest.Init(t)
	r := run(t, repo, "slice", "status", "x")
	want := "Error: the repository has no commits yet. Make a first commit, then try again. Nothing was changed.\n"
	if r.code != 1 || r.stdout != "" || r.stderr != want {
		t.Errorf("code = %d, stdout = %q, stderr = %q, want 1, empty, %q", r.code, r.stdout, r.stderr, want)
	}
}

// TestStatusMissingBase: R23, the base branch is missing; the status still shows.
func TestStatusMissingBase(t *testing.T) {
	repo := gittest.NewRepo(t)
	createSlice(t, repo, "spec-submit")
	gittest.Run(t, repo, "config", "sliceflow.base", "develop")

	r := run(t, repo, "slice", "status", "spec-submit", "--json")
	want := "Warning: base branch develop does not exist (from git config sliceflow.base); merged slices are not shown\n"
	if r.code != 0 || r.stderr != want {
		t.Fatalf("code = %d, stderr = %q, want 0, %q", r.code, r.stderr, want)
	}
	var got struct {
		Warnings []string `json:"warnings"`
	}
	if err := json.Unmarshal([]byte(r.stdout), &got); err != nil {
		t.Fatal(err)
	}
	wantText := "base branch develop does not exist (from git config sliceflow.base); merged slices are not shown"
	if len(got.Warnings) != 1 || got.Warnings[0] != wantText {
		t.Errorf("warnings = %q, want [%q]", got.Warnings, wantText)
	}

	// Without a config, the source is the default.
	gittest.Run(t, repo, "config", "--unset", "sliceflow.base")
	gittest.Run(t, repo, "branch", "-m", "main", "trunk")
	r = run(t, repo, "slice", "status", "spec-submit")
	want = "Warning: base branch main does not exist (the default); merged slices are not shown\n"
	if r.code != 0 || r.stderr != want {
		t.Errorf("code = %d, stderr = %q, want 0, %q", r.code, r.stderr, want)
	}
}

// TestStatusReadsCommitted: R24, an uncommitted edit of the log changes nothing.
func TestStatusReadsCommitted(t *testing.T) {
	repo := gittest.NewRepo(t)
	worktree := createSlice(t, repo, "spec-submit")
	before := run(t, repo, "slice", "status", "spec-submit")

	if err := os.WriteFile(logPath(worktree, "spec-submit"), []byte("garbage\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	after := run(t, repo, "slice", "status", "spec-submit")
	if after != before {
		t.Errorf("after an uncommitted edit: %+v, want %+v", after, before)
	}
}

// TestStatusMergedSlice: a folder on main and no branch is read from main.
func TestStatusMergedSlice(t *testing.T) {
	repo := gittest.NewRepo(t)
	createSlice(t, repo, "spec-submit")
	want := run(t, repo, "slice", "status", "spec-submit")

	gittest.Run(t, repo, "merge", "--no-ff", "-m", "merge", "slice/spec-submit")
	gittest.Run(t, repo, "worktree", "remove", "--force", filepath.Join(repo, ".sliceflow", "worktrees", "spec-submit"))
	gittest.Run(t, repo, "branch", "-D", "slice/spec-submit")

	got := run(t, repo, "slice", "status", "spec-submit")
	if got != want {
		t.Errorf("merged slice: %+v, want %+v", got, want)
	}
}

// TestStatusTooManyArguments: two names are a wrong use.
func TestStatusTooManyArguments(t *testing.T) {
	repo := gittest.NewRepo(t)
	r := run(t, repo, "slice", "status", "a", "b")
	want := "Error: too many arguments. Run sliceflow --help.\n"
	if r.code != 2 || r.stdout != "" || r.stderr != want {
		t.Errorf("code = %d, stdout = %q, stderr = %q, want 2, empty, %q", r.code, r.stdout, r.stderr, want)
	}
}

// TestStatusMissingBaseUnknownName: the R23 warning is shown before the R33 error.
func TestStatusMissingBaseUnknownName(t *testing.T) {
	repo := gittest.NewRepo(t)
	gittest.Run(t, repo, "config", "sliceflow.base", "develop")

	r := run(t, repo, "slice", "status", "nope")
	want := "Warning: base branch develop does not exist (from git config sliceflow.base); merged slices are not shown\n" +
		"Error: no slice named nope. See all slices: sliceflow slice list\n"
	if r.code != 1 || r.stdout != "" || r.stderr != want {
		t.Errorf("code = %d, stdout = %q, stderr = %q, want 1, empty, %q", r.code, r.stdout, r.stderr, want)
	}
}

// TestStatusEmptyBaseConfig: an empty sliceflow.base counts as not set.
func TestStatusEmptyBaseConfig(t *testing.T) {
	repo := gittest.NewRepo(t)
	createSlice(t, repo, "spec-submit")
	gittest.Run(t, repo, "config", "sliceflow.base", "")

	r := run(t, repo, "slice", "status", "spec-submit")
	if r.code != 0 || r.stderr != "" {
		t.Errorf("code = %d, stderr = %q, want 0, empty (base main)", r.code, r.stderr)
	}
}

// TestStatusTitleEscapeCodes: control characters of a title are not printed raw.
func TestStatusTitleEscapeCodes(t *testing.T) {
	repo := gittest.NewRepo(t)
	createSlice(t, repo, "spec-submit", "--title", "red\x1b[31m text")

	r := run(t, repo, "slice", "status", "spec-submit")
	if strings.ContainsRune(r.stdout, 0x1b) {
		t.Fatalf("stdout has a raw escape character: %q", r.stdout)
	}
	if !strings.Contains(r.stdout, "Title:    red\\u001b[31m text\n") {
		t.Errorf("stdout = %q, want the escape written as \\u001b", r.stdout)
	}
}
