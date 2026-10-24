package slice

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/BertPratya/sliceflow/internal/eventlog"
	"github.com/BertPratya/sliceflow/internal/state"
	"github.com/BertPratya/sliceflow/internal/testrepo"
)

const saveName = "login-google"

func saveSlice(t *testing.T) (Repo, string) {
	t.Helper()
	r := newRepo(t)
	r.LockWait = 300 * time.Millisecond
	c, err := Create(context.Background(), r, CreateOptions{Name: saveName, Title: "Login with Google"})
	if err != nil {
		t.Fatal(err)
	}
	return r, filepath.FromSlash(c.Worktree)
}

func saveLogPath(wt string) string {
	return filepath.Join(wt, "docs", "slices", saveName, ".events.jsonl")
}

func titleOp(title string) saveOp {
	return saveOp{
		Event: "slice.titled",
		Prepare: func(state.State) (prepared, error) {
			return prepared{Data: eventlog.Titled{Title: title}, Message: saveName + ": title set"}, nil
		},
	}
}

func readFile(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func lastLine(s string) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	return lines[len(lines)-1]
}

func writeSaveHook(t *testing.T, r Repo, name, body string) {
	t.Helper()
	dir := filepath.Join(r.Dir, ".git", "hooks")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte("#!/bin/sh\n"+body+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
}

func commitCount(t *testing.T, wt string) string {
	return gitOut(t, wt, "rev-list", "--count", "HEAD")
}

func TestSaveCommit(t *testing.T) {
	r, wt := saveSlice(t)
	before := commitCount(t, wt)
	id, st, err := saveEvent(context.Background(), r, saveName, wt, titleOp("New"))
	if err != nil {
		t.Fatal(err)
	}
	if id != gitOut(t, wt, "rev-parse", "HEAD") || len(id) != 40 {
		t.Errorf("id = %q", id)
	}
	if st.Title != "New" {
		t.Errorf("state title = %q", st.Title)
	}
	if got := gitOut(t, wt, "rev-list", "--count", "HEAD"); got == before {
		t.Error("no new commit")
	}
	if got := gitOut(t, wt, "log", "-1", "--format=%s"); got != "login-google: title set" {
		t.Errorf("subject = %q", got)
	}
	files := gitOut(t, wt, "show", "--name-only", "--format=", "HEAD")
	if files != "docs/slices/login-google/.events.jsonl" {
		t.Errorf("files = %q", files)
	}
	if ll := lastLine(readFile(t, saveLogPath(wt))); !strings.HasPrefix(ll, `{"version":1,"seq":2,"event":"slice.titled"`) {
		t.Errorf("last line = %s", ll)
	}
	if _, err := os.Stat(lockPath(r, saveName)); err == nil {
		t.Error("lock not freed")
	}
}

func TestSaveOnlyOwnFiles(t *testing.T) {
	r, wt := saveSlice(t)
	if err := os.MkdirAll(filepath.Join(wt, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wt, "src", "app.go"), []byte("package src\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	testrepo.Git(t, wt, "add", "--", "src/app.go")
	spec := filepath.Join(wt, "docs", "slices", saveName, "spec.md")
	if err := os.WriteFile(spec, []byte("changed spec\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	op := titleOp("T")
	op.Keep = []string{"spec.md"}
	if _, _, err := saveEvent(context.Background(), r, saveName, wt, op); err != nil {
		t.Fatal(err)
	}
	files := gitOut(t, wt, "show", "--name-only", "--format=", "HEAD")
	want := "docs/slices/login-google/.events.jsonl\ndocs/slices/login-google/spec.md"
	if files != want {
		t.Errorf("files = %q", files)
	}
	if staged := gitOut(t, wt, "diff", "--cached", "--name-only"); staged != "src/app.go" {
		t.Errorf("staged = %q", staged)
	}
}

func commitRawLog(t *testing.T, wt, content string) {
	t.Helper()
	if err := os.WriteFile(saveLogPath(wt), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	testrepo.Git(t, wt, "add", "--", saveLogPath(wt))
	testrepo.Git(t, wt, "commit", "-m", "raw log")
}

func TestSaveNoNewline(t *testing.T) {
	r, wt := saveSlice(t)
	committed := readFile(t, saveLogPath(wt))
	half := committed + `{"version":1,"seq":2,"ev`
	commitRawLog(t, wt, half)
	if _, _, err := saveEvent(context.Background(), r, saveName, wt, titleOp("N")); err != nil {
		t.Fatal(err)
	}
	got := readFile(t, saveLogPath(wt))
	if !strings.HasPrefix(got, half+"\n") {
		t.Errorf("half line not kept: %q", got)
	}
	if !strings.HasPrefix(lastLine(got), `{"version":1,"seq":2,"event":"slice.titled"`) {
		t.Errorf("last = %s", lastLine(got))
	}
	if strings.Count(got, "\n") != strings.Count(half, "\n")+2 {
		t.Errorf("log = %q", got)
	}
}

func TestSavePrepareFails(t *testing.T) {
	r, wt := saveSlice(t)
	committed := readFile(t, saveLogPath(wt))
	before := commitCount(t, wt)
	boom := &Error{Message: "spec says no"}
	op := saveOp{Event: "slice.titled", Prepare: func(state.State) (prepared, error) { return prepared{}, boom }}
	_, _, err := saveEvent(context.Background(), r, saveName, wt, op)
	if err != boom {
		t.Errorf("err = %v", err)
	}
	if readFile(t, saveLogPath(wt)) != committed || commitCount(t, wt) != before {
		t.Error("something was saved")
	}
	if _, err := os.Stat(lockPath(r, saveName)); err == nil {
		t.Error("lock not freed")
	}
}

func TestSaveNotAllowed(t *testing.T) {
	r, wt := saveSlice(t)
	committed := readFile(t, saveLogPath(wt))
	before := commitCount(t, wt)
	op := saveOp{Event: "review.finished", Prepare: func(state.State) (prepared, error) {
		return prepared{Data: eventlog.ReviewFinished{Doc: "spec.md", Round: 1, Result: "passed"}, Message: "x"}, nil
	}}
	_, _, err := saveEvent(context.Background(), r, saveName, wt, op)
	if err == nil || !strings.Contains(err.Error(), "review.finished is not allowed when spec.md is draft") {
		t.Errorf("err = %v", err)
	}
	if readFile(t, saveLogPath(wt)) != committed || commitCount(t, wt) != before {
		t.Error("something was saved")
	}
}

func TestSaveHookFails(t *testing.T) {
	r, wt := saveSlice(t)
	committed := readFile(t, saveLogPath(wt))
	before := commitCount(t, wt)
	writeSaveHook(t, r, "pre-commit", "exit 1")
	op := titleOp("H")
	op.Prepare = func(state.State) (prepared, error) {
		return prepared{
			Data:    eventlog.Titled{Title: "H"},
			Files:   map[string][]byte{"reviews/spec-1.md": []byte("review\n")},
			Message: saveName + ": title set",
		}, nil
	}
	op.Keep = []string{"spec.md"}
	_, _, err := saveEvent(context.Background(), r, saveName, wt, op)
	want := "could not save the event: git commit failed (exit code 1). Nothing was changed."
	if err == nil || err.Error() != want {
		t.Errorf("err = %v", err)
	}
	if readFile(t, saveLogPath(wt)) != committed {
		t.Error("log not put back")
	}
	if commitCount(t, wt) != before {
		t.Error("a commit was made")
	}
	if _, err := os.Stat(filepath.Join(wt, "docs", "slices", saveName, "reviews", "spec-1.md")); err == nil {
		t.Error("review file still there")
	}
	if staged := gitOut(t, wt, "diff", "--cached", "--name-only"); staged != "" {
		t.Errorf("staged = %q", staged)
	}
	if _, err := os.Stat(lockPath(r, saveName)); err == nil {
		t.Error("lock not freed")
	}
}

func TestSaveTail(t *testing.T) {
	r, wt := saveSlice(t)
	writeSaveHook(t, r, "pre-commit", "exit 1")
	op := titleOp("H")
	op.Tail = "The steps before it were saved; spec.md is in_review."
	_, _, err := saveEvent(context.Background(), r, saveName, wt, op)
	if err == nil || !strings.HasSuffix(err.Error(), ". The steps before it were saved; spec.md is in_review.") {
		t.Errorf("err = %v", err)
	}
}

func TestSaveCancelled(t *testing.T) {
	r, wt := saveSlice(t)
	r.LockWait = 5 * time.Second
	committed := readFile(t, saveLogPath(wt))
	before := commitCount(t, wt)
	writeSaveHook(t, r, "pre-commit", "sleep 3")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	time.AfterFunc(500*time.Millisecond, cancel)
	op := titleOp("C")
	op.Prepare = func(state.State) (prepared, error) {
		return prepared{
			Data:    eventlog.Titled{Title: "C"},
			Files:   map[string][]byte{"reviews/spec-1.md": []byte("r\n")},
			Message: saveName + ": title set",
		}, nil
	}
	_, _, err := saveEvent(ctx, r, saveName, wt, op)
	want := "could not save the event: interrupted. Nothing was changed."
	if err == nil || err.Error() != want {
		t.Errorf("err = %v", err)
	}
	if readFile(t, saveLogPath(wt)) != committed || commitCount(t, wt) != before {
		t.Error("not undone")
	}
	if _, err := os.Stat(filepath.Join(wt, "docs", "slices", saveName, "reviews", "spec-1.md")); err == nil {
		t.Error("review file still there")
	}
	if staged := gitOut(t, wt, "diff", "--cached", "--name-only"); staged != "" {
		t.Errorf("staged = %q", staged)
	}
	if _, err := os.Stat(lockPath(r, saveName)); err == nil {
		t.Error("lock not freed")
	}
}

func TestSaveTwoAtOnce(t *testing.T) {
	r, wt := saveSlice(t)
	r.LockWait = 20 * time.Second
	var wg sync.WaitGroup
	errs := make([]error, 2)
	for i := range errs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, errs[i] = saveEvent(context.Background(), r, saveName, wt, titleOp("T"))
		}()
	}
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	lines := strings.Split(strings.TrimRight(readFile(t, saveLogPath(wt)), "\n"), "\n")
	if len(lines) != 3 || !strings.Contains(lines[1], `"seq":2,`) || !strings.Contains(lines[2], `"seq":3,`) {
		t.Errorf("lines = %q", lines)
	}
}

func TestSaveSeqAfterSkipped(t *testing.T) {
	r, wt := saveSlice(t)
	appendLog(t, wt, saveName, eventlog.New(2, fixedNow, "review.finished", eventlog.ReviewFinished{Doc: "spec.md", Round: 1, Result: "passed"}))
	if _, _, err := saveEvent(context.Background(), r, saveName, wt, titleOp("S")); err != nil {
		t.Fatal(err)
	}
	if ll := lastLine(readFile(t, saveLogPath(wt))); !strings.HasPrefix(ll, `{"version":1,"seq":3,"event":"slice.titled"`) {
		t.Errorf("last = %s", ll)
	}
}

func TestSaveLeftoverTemp(t *testing.T) {
	r, wt := saveSlice(t)
	tmp := saveLogPath(wt) + ".tmp"
	if err := os.WriteFile(tmp, []byte("junk"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := saveEvent(context.Background(), r, saveName, wt, titleOp("L")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(tmp); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("temp file still there: %v", err)
	}
}
