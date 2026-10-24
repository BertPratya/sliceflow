package slice

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func errMsg(t *testing.T, err error) string {
	t.Helper()
	if err == nil {
		t.Fatal("want an error")
	}
	e, ok := err.(*Error)
	if !ok {
		t.Fatalf("want *Error, got %T: %v", err, err)
	}
	return e.Message
}

func TestSetTitle(t *testing.T) {
	r, wt := saveSlice(t)
	before := commitCount(t, wt)
	if err := SetTitle(context.Background(), r, saveName, "Sign in with Google"); err != nil {
		t.Fatal(err)
	}
	if got := lastLine(readFile(t, saveLogPath(wt))); !strings.Contains(got, `"seq":2,"event":"slice.titled"`) ||
		!strings.HasSuffix(got, `"data":{"title":"Sign in with Google"}}`) {
		t.Errorf("last line %s", got)
	}
	if got := strings.TrimSpace(gitOut(t, wt, "log", "-1", "--format=%s")); got != saveName+": title set" {
		t.Errorf("message %q", got)
	}
	files := strings.TrimSpace(gitOut(t, wt, "show", "--name-only", "--format=", "HEAD"))
	if files != "docs/slices/"+saveName+"/.events.jsonl" {
		t.Errorf("files %q", files)
	}
	if commitCount(t, wt) == before {
		t.Error("no commit")
	}
}

func TestSetTitleSameAndEmpty(t *testing.T) {
	r, wt := saveSlice(t)
	before := commitCount(t, wt)
	ctx := context.Background()
	got := errMsg(t, SetTitle(ctx, r, saveName, "Login with Google"))
	if want := `the title of login-google is already "Login with Google". Nothing was changed.`; got != want {
		t.Errorf("same: %q", got)
	}
	got = errMsg(t, SetTitle(ctx, r, saveName, ""))
	if want := "the title must not be empty. Nothing was changed."; got != want {
		t.Errorf("empty: %q", got)
	}
	if commitCount(t, wt) != before {
		t.Error("commit made")
	}
}

func TestSetTitleUntitledOwnName(t *testing.T) {
	r := newRepo(t)
	c, err := Create(context.Background(), r, CreateOptions{Name: saveName})
	if err != nil {
		t.Fatal(err)
	}
	wt := filepath.FromSlash(c.Worktree)
	before := commitCount(t, wt)
	got := errMsg(t, SetTitle(context.Background(), r, saveName, saveName))
	if want := `the title of login-google is already "login-google". Nothing was changed.`; got != want {
		t.Errorf("got %q", got)
	}
	if commitCount(t, wt) != before {
		t.Error("commit made")
	}
}

func TestSetTitleCancelled(t *testing.T) {
	r, wt := saveSlice(t)
	before := commitCount(t, wt)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	got := errMsg(t, SetTitle(ctx, r, saveName, "New title"))
	if want := "could not save the event: interrupted. Nothing was changed."; got != want {
		t.Errorf("got %q", got)
	}
	if commitCount(t, wt) != before {
		t.Error("commit made")
	}
}

func TestSetTitleUnknownState(t *testing.T) {
	r, wt := saveSlice(t)
	commitRawLog(t, wt, "not json\n")
	before := commitCount(t, wt)
	got := errMsg(t, SetTitle(context.Background(), r, saveName, "X"))
	want := "slice login-google has no slice.created event, so its state is unknown. Nothing was changed."
	if got != want {
		t.Errorf("got %q", got)
	}
	if commitCount(t, wt) != before {
		t.Error("commit made")
	}
}
