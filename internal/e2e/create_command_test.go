package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/BertPratya/sliceflow/internal/testrepo"
)

// realDir returns dir with links resolved and "/" between parts.
func realDir(t *testing.T, dir string) string {
	t.Helper()
	d, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatal(err)
	}
	return filepath.ToSlash(d)
}

func TestCreateSpecSubmit(t *testing.T) {
	dir := testrepo.New(t)
	commit := strings.TrimSpace(testrepo.Git(t, dir, "rev-parse", "HEAD"))
	out, errOut, code := run(t, dir, "slice", "create", "spec-submit")
	want := "Created slice spec-submit.\n" +
		"  Branch:   slice/spec-submit (from main @ " + commit[:7] + ")\n" +
		"  Worktree: .sliceflow/worktrees/spec-submit\n" +
		"  Spec:     docs/slices/spec-submit/spec.md\n" +
		"Next: write the spec in .sliceflow/worktrees/spec-submit/docs/slices/spec-submit/spec.md\n"
	if code != 0 || errOut != "" || out != want {
		t.Fatalf("code %d\nstdout %q\nwant   %q\nstderr %q", code, out, want, errOut)
	}
	ig, err := os.ReadFile(filepath.Join(dir, ".sliceflow", ".gitignore"))
	if err != nil || string(ig) != "*\n" {
		t.Errorf(".gitignore %q err %v", ig, err)
	}
	files := strings.Fields(testrepo.Git(t, dir, "diff", "--name-only", "main", "slice/spec-submit"))
	if len(files) != 2 || files[0] != "docs/slices/spec-submit/.events.jsonl" || files[1] != "docs/slices/spec-submit/spec.md" {
		t.Errorf("files %v", files)
	}
	ev := testrepo.Git(t, dir, "show", "slice/spec-submit:docs/slices/spec-submit/.events.jsonl")
	if strings.Count(ev, "\n") != 1 || !strings.HasSuffix(ev, "\n") {
		t.Errorf("event log not one line: %q", ev)
	}
	mainFull := strings.TrimSpace(testrepo.Git(t, dir, "rev-parse", "main"))
	reEv := regexp.MustCompile(`^\{"version":1,"seq":1,"event":"slice\.created","at":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z",`)
	if !reEv.MatchString(ev) {
		t.Errorf("event line %q", ev)
	}
	if !strings.Contains(ev, `"from":{"branch":"main","commit":"`+mainFull+`"}`) {
		t.Errorf("event from %q want commit %s", ev, mainFull)
	}
	log := strings.TrimSpace(testrepo.Git(t, dir, "log", "--format=%s", "main..slice/spec-submit"))
	if log != "spec-submit: slice created" {
		t.Errorf("log %q", log)
	}
}

func TestCreateJSON(t *testing.T) {
	dir := testrepo.New(t)
	testrepo.Git(t, dir, "branch", "develop")
	testrepo.Git(t, dir, "config", "sliceflow.base", "develop")
	commit := strings.TrimSpace(testrepo.Git(t, dir, "rev-parse", "develop"))
	out, errOut, code := run(t, dir, "slice", "create", "login-google", "--title", "Login with Google", "--json")
	if code != 0 || errOut != "" {
		t.Fatalf("code %d stderr %q", code, errOut)
	}
	if strings.Count(out, "\n") != 1 {
		t.Fatalf("want one line: %q", out)
	}
	var got struct {
		Slice, Branch string
		From          struct{ Branch, Commit string }
		Worktree      string
		Spec          string
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatal(err)
	}
	wt := realDir(t, dir) + "/.sliceflow/worktrees/login-google"
	if got.Slice != "login-google" || got.Branch != "slice/login-google" ||
		got.From.Branch != "develop" || got.From.Commit != commit ||
		got.Spec != "docs/slices/login-google/spec.md" {
		t.Errorf("got %+v", got)
	}
	if realDir(t, got.Worktree) != wt || strings.Contains(got.Worktree, "\\") {
		t.Errorf("worktree %q want %q", got.Worktree, wt)
	}
	ev, err := os.ReadFile(filepath.Join(dir, ".sliceflow", "worktrees", "login-google", "docs", "slices", "login-google", ".events.jsonl"))
	if err != nil || !strings.Contains(string(ev), `"title":"Login with Google"`) {
		t.Errorf("events %q err %v", ev, err)
	}
}

func TestCreateFromFlag(t *testing.T) {
	dir := testrepo.New(t)
	testrepo.Git(t, dir, "checkout", "-b", "develop")
	testrepo.Git(t, dir, "commit", "--allow-empty", "-m", "dev work")
	testrepo.Git(t, dir, "checkout", "main")
	testrepo.Git(t, dir, "config", "sliceflow.base", "develop")
	commit := strings.TrimSpace(testrepo.Git(t, dir, "rev-parse", "main"))
	out, errOut, code := run(t, dir, "slice", "create", "hotfix-ui", "--from", "main")
	want := "  Branch:   slice/hotfix-ui (from main @ " + commit[:7] + ")\n"
	if code != 0 || errOut != "" || !strings.Contains(out, want) {
		t.Fatalf("code %d out %q err %q", code, out, errOut)
	}
	parent := strings.TrimSpace(testrepo.Git(t, dir, "rev-parse", "slice/hotfix-ui~1"))
	if parent != commit {
		t.Errorf("parent %s want %s", parent, commit)
	}
}

func TestCreateInsideWorktree(t *testing.T) {
	dir := testrepo.New(t)
	if _, e, c := run(t, dir, "slice", "create", "other"); c != 0 {
		t.Fatalf("first create: %d %q", c, e)
	}
	inside := filepath.Join(dir, ".sliceflow", "worktrees", "other")
	commit := strings.TrimSpace(testrepo.Git(t, dir, "rev-parse", "main"))
	out, errOut, code := run(t, inside, "slice", "create", "spec-submit")
	want := "Created slice spec-submit.\n" +
		"  Branch:   slice/spec-submit (from main @ " + commit[:7] + ")\n" +
		"  Worktree: .sliceflow/worktrees/spec-submit\n" +
		"  Spec:     docs/slices/spec-submit/spec.md\n" +
		"Next: write the spec in .sliceflow/worktrees/spec-submit/docs/slices/spec-submit/spec.md\n"
	if code != 0 || errOut != "" || out != want {
		t.Fatalf("code %d\nstdout %q\nwant   %q\nstderr %q", code, out, want, errOut)
	}
	if _, err := os.Stat(filepath.Join(dir, ".sliceflow", "worktrees", "spec-submit", "docs", "slices", "spec-submit", "spec.md")); err != nil {
		t.Error(err)
	}
	if _, err := os.Stat(filepath.Join(inside, ".sliceflow")); err == nil {
		t.Error("worktree made inside the other worktree")
	}
}

func TestCreateHelp(t *testing.T) {
	dir := testrepo.New(t)
	out, errOut, code := run(t, dir, "slice", "create", "--help")
	if code != 0 || errOut != "" {
		t.Fatalf("code %d stderr %q", code, errOut)
	}
	for _, f := range []string{"--title", "--from", "--json"} {
		if !strings.Contains(out, f) {
			t.Errorf("usage lacks %s: %q", f, out)
		}
	}
}

func TestCreateWrongUse(t *testing.T) {
	dir := testrepo.New(t)
	cases := []struct {
		args []string
		want string
	}{
		{[]string{"slice", "create"}, "Error: missing slice name. Run sliceflow --help.\n"},
		{[]string{"slice", "create", "a", "b"}, "Error: too many arguments. Run sliceflow --help.\n"},
	}
	for _, c := range cases {
		out, errOut, code := run(t, dir, c.args...)
		if code != 2 || out != "" || errOut != c.want {
			t.Errorf("%v: code %d out %q err %q", c.args, code, out, errOut)
		}
	}
}
