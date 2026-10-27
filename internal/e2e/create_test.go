package e2e

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/BertPratya/sliceflow/internal/gittest"
)

// realPath returns the path with symlinks resolved and / between the parts.
func realPath(t *testing.T, path string) string {
	t.Helper()
	real, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatal(err)
	}
	return filepath.ToSlash(real)
}

// readFile returns the text of the file at path (with / between the parts).
func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.FromSlash(path))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// atPattern matches the time of an event: UTC, to the second.
var atPattern = regexp.MustCompile(`^\d{4}-\d\d-\d\dT\d\d:\d\d:\d\dZ$`)

// eventLine is the first line of the event log, as the test reads it.
type eventLine struct {
	Version int             `json:"version"`
	Seq     int             `json:"seq"`
	Event   string          `json:"event"`
	At      string          `json:"at"`
	Data    json.RawMessage `json:"data"`
}

// readEvent reads .events.jsonl of a slice from its branch, checks that it
// is exactly one line, and returns that line parsed, and as text.
func readEvent(t *testing.T, repo, name string) (eventLine, string) {
	t.Helper()
	// gittest.Run trims the end, so add back the \n to compare below.
	raw := gittest.Run(t, repo, "show", "slice/"+name+":docs/slices/"+name+"/.events.jsonl")
	if strings.Contains(raw, "\n") {
		t.Fatalf("event log has more than one line: %q", raw)
	}
	var ev eventLine
	if err := json.Unmarshal([]byte(raw), &ev); err != nil {
		t.Fatal(err)
	}
	return ev, raw
}

// TestS1: the plain case. Output, files, commit, and the log line.
func TestS1(t *testing.T) {
	repo := gittest.NewRepo(t)
	full := gittest.Run(t, repo, "rev-parse", "main")
	short := full[:7]

	r := run(t, repo, "slice", "create", "spec-submit")
	want := "Created slice spec-submit.\n" +
		"  Branch:   slice/spec-submit (from main @ " + short + ")\n" +
		"  Worktree: .sliceflow/worktrees/spec-submit\n" +
		"  Spec:     docs/slices/spec-submit/spec.md\n" +
		"Next: write the spec in .sliceflow/worktrees/spec-submit/docs/slices/spec-submit/spec.md\n"
	if r.code != 0 || r.stderr != "" || r.stdout != want {
		t.Fatalf("code = %d, stderr = %q, stdout = %q, want 0, \"\", %q", r.code, r.stderr, r.stdout, want)
	}

	if got := readFile(t, filepath.ToSlash(repo)+"/.sliceflow/.gitignore"); got != "*\n" {
		t.Errorf(".sliceflow/.gitignore = %q, want %q", got, "*\n")
	}

	// One new commit after main, with the R8 message.
	if got := gittest.Run(t, repo, "log", "--format=%s", "main..slice/spec-submit"); got != "spec-submit: slice created" {
		t.Errorf("commits after main = %q, want exactly one: spec-submit: slice created", got)
	}
	if got := gittest.Run(t, repo, "rev-parse", "slice/spec-submit~1"); got != full {
		t.Errorf("parent of the commit = %q, want %q", got, full)
	}
	files := gittest.Run(t, repo, "show", "--name-only", "--format=", "slice/spec-submit")
	if wantFiles := "docs/slices/spec-submit/.events.jsonl\ndocs/slices/spec-submit/spec.md"; files != wantFiles {
		t.Errorf("files in the commit = %q, want %q", files, wantFiles)
	}
	if got := gittest.Run(t, repo, "log", "-1", "--format=%an <%ae>", "slice/spec-submit"); got != "sliceflow test <test@example.com>" {
		t.Errorf("author = %q, want the repo's configured user", got)
	}

	// The log line.
	ev, raw := readEvent(t, repo, "spec-submit")
	if !atPattern.MatchString(ev.At) {
		t.Errorf("at = %q, want a UTC time to the second", ev.At)
	}
	wantLine := `{"version":1,"seq":1,"event":"slice.created","at":"` + ev.At + `","data":{"from":{"branch":"main","commit":"` + full + `"}}}`
	if raw != wantLine {
		t.Errorf("log line = %s\nwant %s", raw, wantLine)
	}

	// The spec: front matter with today's UTC date.
	spec := gittest.Run(t, repo, "show", "slice/spec-submit:docs/slices/spec-submit/spec.md")
	// Today's or yesterday's date, so the test cannot fail at midnight.
	now := time.Now().UTC()
	okFront := false
	for _, day := range []time.Time{now, now.AddDate(0, 0, -1)} {
		front := "---\nslice: spec-submit\ncreated: " + day.Format("2006-01-02") + "\nideas: []\n---"
		okFront = okFront || strings.HasPrefix(spec, front)
	}
	if !okFront {
		t.Errorf("spec starts with %q, want today's UTC date in the front matter", spec[:min(len(spec), 100)])
	}

	// The committed bytes, untrimmed: no \r, and one \n at the end of the log.
	events := blob(t, repo, "slice/spec-submit:docs/slices/spec-submit/.events.jsonl")
	if strings.Contains(events, "\r") || !strings.HasSuffix(events, "}\n") || strings.Count(events, "\n") != 1 {
		t.Errorf("event log bytes = %q, want one line ending in a single \\n", events)
	}
	if strings.Contains(blob(t, repo, "slice/spec-submit:docs/slices/spec-submit/spec.md"), "\r") {
		t.Error("spec.md has a \\r")
	}

	// The worktree is on the slice branch and clean; the main checkout is
	// still on main and does not show .sliceflow.
	wt := filepath.Join(repo, ".sliceflow", "worktrees", "spec-submit")
	if got := gittest.Run(t, wt, "branch", "--show-current"); got != "slice/spec-submit" {
		t.Errorf("worktree branch = %q", got)
	}
	if got := gittest.Run(t, wt, "status", "--porcelain"); got != "" {
		t.Errorf("worktree status = %q, want clean", got)
	}
	if got := gittest.Run(t, repo, "branch", "--show-current"); got != "main" {
		t.Errorf("main checkout branch = %q, want main", got)
	}
	if got := gittest.Run(t, repo, "status", "--porcelain"); got != "" {
		t.Errorf("main checkout status = %q, want clean (.sliceflow ignored)", got)
	}
}

// blob returns the bytes of a file in git, as git prints them, untrimmed.
func blob(t *testing.T, repo, object string) string {
	t.Helper()
	cmd := exec.Command("git", "cat-file", "blob", object)
	cmd.Dir = repo
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

// TestCreateRunsHooks: R8, the user's pre-commit hook runs on the commit.
func TestCreateRunsHooks(t *testing.T) {
	repo := gittest.NewRepo(t)
	marker := filepath.ToSlash(filepath.Join(t.TempDir(), "marker"))
	hook := "#!/bin/sh\necho ran >> '" + marker + "'\n"
	path := filepath.Join(repo, ".git", "hooks", "pre-commit")
	if err := os.WriteFile(path, []byte(hook), 0o755); err != nil {
		t.Fatal(err)
	}

	if r := run(t, repo, "slice", "create", "x"); r.code != 0 {
		t.Fatalf("create: %+v", r)
	}
	if got := readFile(t, marker); got != "ran\n" {
		t.Errorf("marker = %q, want the hook to run once", got)
	}
}

// TestS2: base from git config, a title, and --json.
func TestS2(t *testing.T) {
	repo := gittest.NewRepo(t)
	gittest.Run(t, repo, "branch", "develop")
	gittest.Run(t, repo, "checkout", "develop")
	gittest.Commit(t, repo, "on develop")
	gittest.Run(t, repo, "checkout", "main")
	gittest.Run(t, repo, "config", "sliceflow.base", "develop")
	full := gittest.Run(t, repo, "rev-parse", "develop")

	r := run(t, repo, "slice", "create", "login-google", "--title", "Login with Google", "--json")
	if r.code != 0 || r.stderr != "" {
		t.Fatalf("code = %d, stderr = %q, want 0 and empty", r.code, r.stderr)
	}
	if strings.Count(r.stdout, "\n") != 1 || !strings.HasSuffix(r.stdout, "}\n") {
		t.Errorf("stdout = %q, want one JSON line", r.stdout)
	}

	var got struct {
		Slice, Branch string
		From          struct{ Branch, Commit string }
		Worktree      string
		Spec          string
	}
	if err := json.Unmarshal([]byte(r.stdout), &got); err != nil {
		t.Fatal(err)
	}
	wantWorktree := realPath(t, repo) + "/.sliceflow/worktrees/login-google"
	if got.Slice != "login-google" || got.Branch != "slice/login-google" ||
		got.From.Branch != "develop" || got.From.Commit != full ||
		got.Spec != "docs/slices/login-google/spec.md" {
		t.Errorf("JSON = %+v", got)
	}
	if realPath(t, got.Worktree) != wantWorktree || strings.Contains(got.Worktree, `\`) {
		t.Errorf("worktree = %q, want %q with /", got.Worktree, wantWorktree)
	}
	// The fields come in the order of R10.
	if !strings.HasPrefix(r.stdout, `{"slice":"login-google","branch":"slice/login-google","from":{"branch":"develop","commit":"`) {
		t.Errorf("stdout = %q, wrong field order", r.stdout)
	}

	ev, _ := readEvent(t, repo, "login-google")
	wantData := `{"from":{"branch":"develop","commit":"` + full + `"},"title":"Login with Google"}`
	if string(ev.Data) != wantData {
		t.Errorf("data = %s, want %s", ev.Data, wantData)
	}
	if parent := gittest.Run(t, repo, "rev-parse", "slice/login-google~1"); parent != full {
		t.Errorf("parent = %q, want develop's commit %q", parent, full)
	}
}

// TestS3: --from beats git config sliceflow.base.
func TestS3(t *testing.T) {
	repo := gittest.NewRepo(t)
	gittest.Run(t, repo, "branch", "develop")
	gittest.Run(t, repo, "config", "sliceflow.base", "develop")
	main := gittest.Run(t, repo, "rev-parse", "main")

	r := run(t, repo, "slice", "create", "hotfix-ui", "--from", "main")
	if r.code != 0 || r.stderr != "" {
		t.Fatalf("code = %d, stderr = %q", r.code, r.stderr)
	}
	if want := "  Branch:   slice/hotfix-ui (from main @ " + main[:7] + ")\n"; !strings.Contains(r.stdout, want) {
		t.Errorf("stdout = %q, want it to hold %q", r.stdout, want)
	}
	ev, _ := readEvent(t, repo, "hotfix-ui")
	if want := `{"from":{"branch":"main","commit":"` + main + `"}}`; string(ev.Data) != want {
		t.Errorf("data = %s, want %s (no title)", ev.Data, want)
	}
}

// TestS4: inside another slice's worktree, the new worktree still goes to
// the main checkout.
func TestS4(t *testing.T) {
	repo := gittest.NewRepo(t)
	if r := run(t, repo, "slice", "create", "first"); r.code != 0 {
		t.Fatalf("first create: %+v", r)
	}
	inside := filepath.Join(repo, ".sliceflow", "worktrees", "first")

	r := run(t, inside, "slice", "create", "spec-submit")
	if r.code != 0 || r.stderr != "" {
		t.Fatalf("code = %d, stderr = %q", r.code, r.stderr)
	}
	if !strings.Contains(r.stdout, "  Worktree: .sliceflow/worktrees/spec-submit\n") {
		t.Errorf("stdout = %q", r.stdout)
	}
	if _, err := os.Stat(filepath.Join(repo, ".sliceflow", "worktrees", "spec-submit", "docs", "slices", "spec-submit", "spec.md")); err != nil {
		t.Errorf("worktree not at the main checkout: %v", err)
	}
	if _, err := os.Stat(filepath.Join(inside, ".sliceflow")); err == nil {
		t.Error("a .sliceflow folder was made inside the other worktree")
	}
}

// TestCreateFromSubfolder: Worktree and Next are the same in a subfolder.
func TestCreateFromSubfolder(t *testing.T) {
	repo := gittest.NewRepo(t)
	sub := makeDir(t, repo, "sub/deeper")

	r := run(t, sub, "slice", "create", "x")
	if r.code != 0 || r.stderr != "" {
		t.Fatalf("code = %d, stderr = %q", r.code, r.stderr)
	}
	for _, want := range []string{
		"  Worktree: .sliceflow/worktrees/x\n",
		"Next: write the spec in .sliceflow/worktrees/x/docs/slices/x/spec.md\n",
	} {
		if !strings.Contains(r.stdout, want) {
			t.Errorf("stdout = %q, want it to hold %q", r.stdout, want)
		}
	}
	if _, err := os.Stat(filepath.Join(repo, ".sliceflow", "worktrees", "x")); err != nil {
		t.Error(err)
	}
}

// TestCreateKeepsGitignore: R6, an existing .sliceflow/.gitignore stays.
func TestCreateKeepsGitignore(t *testing.T) {
	repo := gittest.NewRepo(t)
	if r := run(t, repo, "slice", "create", "one"); r.code != 0 {
		t.Fatalf("first create: %+v", r)
	}
	path := filepath.Join(repo, ".sliceflow", ".gitignore")
	if err := os.WriteFile(path, []byte("mine\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if r := run(t, repo, "slice", "create", "two"); r.code != 0 {
		t.Fatalf("second create: %+v", r)
	}
	if got := readFile(t, filepath.ToSlash(path)); got != "mine\n" {
		t.Errorf(".sliceflow/.gitignore = %q, want it unchanged", got)
	}
}

// TestCreateJSONError: R42, with --json an error looks the same.
func TestCreateJSONError(t *testing.T) {
	repo := gittest.NewRepo(t)
	r := run(t, repo, "slice", "create", "My Slice", "--json")
	wantFail(t, r, `Error: invalid slice name "My Slice": use 1-64 lowercase letters, digits, and dashes, starting with a letter or digit. Nothing was changed.`)
}
