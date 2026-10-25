package e2e

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/BertPratya/sliceflow/internal/testrepo"
)

func titledSlice(t *testing.T) string {
	t.Helper()
	dir := testrepo.New(t)
	createSlice(t, dir, "login-google")
	setCreated(t, dir, "login-google", "2026-10-24T07:42:00Z")
	return dir
}

func TestTitleSet(t *testing.T) {
	dir := titledSlice(t)
	before := strings.TrimSpace(testrepo.Git(t, dir, "rev-list", "--count", "slice/login-google"))
	out, errOut, code := run(t, dir, "slice", "title", "login-google", "Sign in with Google")
	if want := "Title of login-google is now \"Sign in with Google\".\n"; code != 0 || errOut != "" || out != want {
		t.Fatalf("code %d stdout %q stderr %q", code, out, errOut)
	}
	after := strings.TrimSpace(testrepo.Git(t, dir, "rev-list", "--count", "slice/login-google"))
	if after == before {
		t.Fatal("no new commit")
	}
	if got := strings.TrimSpace(testrepo.Git(t, dir, "log", "-1", "--format=%s", "slice/login-google")); got != "login-google: title set" {
		t.Errorf("message %q", got)
	}
	if got := strings.TrimSpace(testrepo.Git(t, dir, "show", "--name-only", "--format=", "slice/login-google")); got != "docs/slices/login-google/.events.jsonl" {
		t.Errorf("files %q", got)
	}
	ev := testrepo.Git(t, dir, "show", "slice/login-google:docs/slices/login-google/.events.jsonl")
	line := strings.Split(strings.TrimRight(ev, "\n"), "\n")[1]
	re := regexp.MustCompile(`^\{"version":1,"seq":2,"event":"slice\.titled","at":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z","data":\{"title":"Sign in with Google"\}\}$`)
	if !re.MatchString(line) {
		t.Errorf("line %q", line)
	}
	list, _, code := run(t, dir, "slice", "list")
	if code != 0 || !strings.Contains(list, "login-google  Sign in with Google  planning  2026-10-24 07:42\n") {
		t.Errorf("list %q", list)
	}
}

func TestTitleSame(t *testing.T) {
	dir := titledSlice(t)
	run(t, dir, "slice", "title", "login-google", "Sign in with Google")
	before := testrepo.Git(t, dir, "rev-parse", "slice/login-google")
	out, errOut, code := run(t, dir, "slice", "title", "login-google", "Sign in with Google")
	want := "Error: the title of login-google is already \"Sign in with Google\". Nothing was changed.\n"
	if code != 1 || out != "" || errOut != want {
		t.Fatalf("code %d stdout %q stderr %q", code, out, errOut)
	}
	if testrepo.Git(t, dir, "rev-parse", "slice/login-google") != before {
		t.Error("new commit")
	}
}

func TestTitleEmpty(t *testing.T) {
	dir := titledSlice(t)
	before := testrepo.Git(t, dir, "rev-parse", "slice/login-google")
	out, errOut, code := run(t, dir, "slice", "title", "login-google", "")
	if want := "Error: the title must not be empty. Nothing was changed.\n"; code != 1 || out != "" || errOut != want {
		t.Fatalf("code %d stdout %q stderr %q", code, out, errOut)
	}
	if testrepo.Git(t, dir, "rev-parse", "slice/login-google") != before {
		t.Error("new commit")
	}
}

func TestTitleMissing(t *testing.T) {
	dir := titledSlice(t)
	out, errOut, code := run(t, dir, "slice", "title", "login-google")
	if want := "Error: missing title. Run sliceflow --help.\n"; code != 2 || out != "" || errOut != want {
		t.Fatalf("code %d stdout %q stderr %q", code, out, errOut)
	}
}

func TestTitleUnknownState(t *testing.T) {
	dir := titledSlice(t)
	appendLogRewrite(t, dir, "login-google")
	out, errOut, code := run(t, dir, "slice", "title", "login-google", "X")
	want := "Error: slice login-google has no slice.created event, so its state is unknown. Nothing was changed.\n"
	if code != 1 || out != "" || errOut != want {
		t.Fatalf("code %d stdout %q stderr %q", code, out, errOut)
	}
}

// appendLogRewrite replaces the log with a line that is not an event, and commits.
func appendLogRewrite(t *testing.T, dir, name string) {
	t.Helper()
	wt := sliceWT(dir, name)
	p := filepath.Join(wt, "docs", "slices", name, ".events.jsonl")
	if err := os.WriteFile(p, []byte("not json\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	testrepo.Git(t, wt, "commit", "-am", "bad log")
}
