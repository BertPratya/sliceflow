package slice

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BertPratya/sliceflow/internal/eventlog"
)

// submittedSlice makes a slice with an accepted spec.submitted. It returns
// the repo, the worktree, and the commit that added the submit.
func submittedSlice(t *testing.T) (Repo, string, string) {
	t.Helper()
	r, wt := saveSlice(t)
	writeSpec(t, wt, "")
	sub, err := Submit(context.Background(), r, saveName)
	if err != nil || sub.Refused != nil {
		t.Fatalf("submit: %v %+v", err, sub.Refused)
	}
	return r, wt, gitOut(t, wt, "rev-parse", "HEAD")
}

func reviewFilePath(wt string, n string) string {
	return filepath.Join(wt, "docs", "slices", saveName, "reviews", "spec-"+n+".md")
}

func TestReviewPasses(t *testing.T) {
	ctx := context.Background()
	r, wt, submitCommit := submittedSlice(t)
	if err := SetTitle(ctx, r, saveName, "Another title"); err != nil {
		t.Fatal(err)
	}
	before := commitCount(t, wt)
	var steps []Step
	if err := ReviewSpec(ctx, r, saveName, func(s Step) { steps = append(steps, s) }); err != nil {
		t.Fatal(err)
	}
	want := []Step{
		{Kind: "started", Round: 1, Reviewer: "placeholder", File: "reviews/spec-1.md"},
		{Kind: "finished", Round: 1, Reviewer: "placeholder", File: "reviews/spec-1.md", Result: "passed"},
	}
	if len(steps) != 2 || steps[0] != want[0] || steps[1] != want[1] {
		t.Fatalf("steps %+v", steps)
	}
	msgs := gitOut(t, wt, "log", "-2", "--format=%s", "--reverse")
	if msgs != saveName+": review 1 started\n"+saveName+": review 1 passed" {
		t.Errorf("messages %q", msgs)
	}
	if n := gitOut(t, wt, "rev-list", "--count", "HEAD"); n == before {
		t.Error("no commits")
	}
	started := gitOut(t, wt, "show", "--name-only", "--format=", "HEAD~1")
	if started != "docs/slices/"+saveName+"/.events.jsonl\ndocs/slices/"+saveName+"/reviews/spec-1.md" {
		t.Errorf("started files %q", started)
	}
	if got := gitOut(t, wt, "show", "HEAD~1:docs/slices/"+saveName+"/reviews/spec-1.md"); !strings.HasSuffix(got, "---") {
		t.Errorf("started file %q", got)
	}
	if got := readFile(t, reviewFilePath(wt, "1")); !strings.Contains(got, "## Needs human decision\n\nNone.") {
		t.Errorf("file %q", got)
	}
	lines := strings.Split(strings.TrimRight(readFile(t, saveLogPath(wt)), "\n"), "\n")
	l := lines[len(lines)-2]
	if !strings.Contains(l, `"event":"review.started"`) || !strings.Contains(l, `"commit":"`+submitCommit+`"`) ||
		!strings.Contains(l, `"file":"reviews/spec-1.md"`) || !strings.Contains(l, `"round":1`) {
		t.Errorf("started line %s", l)
	}
	if l := lines[len(lines)-1]; !strings.HasSuffix(l, `"data":{"doc":"spec.md","round":1,"result":"passed"}}`) {
		t.Errorf("finished line %s", l)
	}
}

func TestReviewStates(t *testing.T) {
	ctx := context.Background()
	at := fixedNow
	submitted := eventlog.New(2, at, "spec.submitted", eventlog.Submitted{Result: "accepted"})
	started := eventlog.New(3, at, "review.started", eventlog.ReviewStarted{Doc: "spec.md", Round: 1, Reviewer: "placeholder", File: "reviews/spec-1.md", Commit: "3f2a1c9e5b7d4a6f8c0e2b4d6f8a0c2e4b6d8f0a"})
	finished := eventlog.New(4, at, "review.finished", eventlog.ReviewFinished{Doc: "spec.md", Round: 1, Result: "passed"})
	cases := []struct {
		name string
		evs  []eventlog.Event
		raw  string
		want string
	}{
		{"draft", nil, "", "spec.md is a draft; submit it first: sliceflow spec submit. Nothing was changed."},
		{"in_review", []eventlog.Event{submitted, started}, "", "spec.md is in_review: review 1 has started, and has not finished. Nothing was changed."},
		{"review_passed", []eventlog.Event{submitted, started, finished}, "", "spec.md passed review 1; there is nothing to review. Nothing was changed."},
		{"no created", nil, "not json\n", "slice login-google has no slice.created event, so its state is unknown. Nothing was changed."},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r, wt := saveSlice(t)
			if c.raw != "" {
				commitRawLog(t, wt, c.raw)
			} else if c.evs != nil {
				commitEvents(t, wt, c.evs...)
			}
			before := commitCount(t, wt)
			log := readFile(t, saveLogPath(wt))
			called := false
			err := ReviewSpec(ctx, r, saveName, func(Step) { called = true })
			wantErrMsg(t, err, c.want)
			if called || commitCount(t, wt) != before || readFile(t, saveLogPath(wt)) != log {
				t.Error("something changed")
			}
		})
	}
}

func TestReviewFileExists(t *testing.T) {
	r, wt, _ := submittedSlice(t)
	p := reviewFilePath(wt, "1")
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte("mine\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	before := commitCount(t, wt)
	err := ReviewSpec(context.Background(), r, saveName, func(Step) {})
	wantErrMsg(t, err, "docs/slices/login-google/reviews/spec-1.md already exists, but review 1 has not started. Move it away, then review again. Nothing was changed.")
	if commitCount(t, wt) != before || readFile(t, p) != "mine\n" {
		t.Error("something changed")
	}
}

func TestReviewRound2(t *testing.T) {
	ctx := context.Background()
	r, wt, _ := submittedSlice(t)
	if err := ReviewSpec(ctx, r, saveName, func(Step) {}); err != nil {
		t.Fatal(err)
	}
	b := readFile(t, specPath(wt))
	edited := strings.Replace(b, "title: Slice and events", "title: Slice and events, edited", 1)
	if edited == b {
		t.Fatal("no edit")
	}
	if err := os.WriteFile(specPath(wt), []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}
	if sub, err := Submit(ctx, r, saveName); err != nil || sub.Refused != nil {
		t.Fatalf("submit: %v %+v", err, sub.Refused)
	}
	submit2 := gitOut(t, wt, "rev-parse", "HEAD")
	var steps []Step
	if err := ReviewSpec(ctx, r, saveName, func(s Step) { steps = append(steps, s) }); err != nil {
		t.Fatal(err)
	}
	if len(steps) != 2 || steps[0].Round != 2 || steps[0].File != "reviews/spec-2.md" {
		t.Fatalf("steps %+v", steps)
	}
	if _, err := os.Stat(reviewFilePath(wt, "2")); err != nil {
		t.Error(err)
	}
	if got := gitOut(t, wt, "show", "HEAD~1:docs/slices/"+saveName+"/.events.jsonl"); !strings.Contains(lastLine(got), `"round":2`) ||
		!strings.Contains(lastLine(got), `"commit":"`+submit2+`"`) {
		t.Errorf("started line %s", lastLine(got))
	}
}

func TestReviewNoName(t *testing.T) {
	r, _, _ := submittedSlice(t)
	err := ReviewSpec(context.Background(), r, "", func(Step) {})
	se, ok := err.(*Error)
	if !ok {
		t.Fatalf("want *Error, got %T: %v", err, err)
	}
	lines := strings.Split(se.Message, "\n")
	if len(lines) != 3 || lines[1] != "Give a name: sliceflow spec review <name>" {
		t.Errorf("message %q", se.Message)
	}
}

func TestReviewNoNameInterrupted(t *testing.T) {
	r, wt, _ := submittedSlice(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	before := commitCount(t, wt)
	err := ReviewSpec(ctx, r, "", func(Step) {})
	wantErrMsg(t, err, InterruptedSave().(*Error).Message)
	if commitCount(t, wt) != before {
		t.Error("commit made")
	}
}

func TestReviewFinishFails(t *testing.T) {
	r, wt, _ := submittedSlice(t)
	writeSaveHook(t, r, "commit-msg", `if grep -q passed "$1"; then exit 1; fi`)
	var steps []Step
	err := ReviewSpec(context.Background(), r, saveName, func(s Step) { steps = append(steps, s) })
	if err == nil || !strings.HasSuffix(err.Error(), "The steps before it were saved; spec.md is in_review.") {
		t.Fatalf("err = %v", err)
	}
	if len(steps) != 1 || steps[0].Kind != "started" {
		t.Errorf("steps %+v", steps)
	}
	if got := gitOut(t, wt, "log", "-1", "--format=%s"); got != saveName+": review 1 started" {
		t.Errorf("head %q", got)
	}
	if !strings.Contains(lastLine(readFile(t, saveLogPath(wt))), `"event":"review.started"`) {
		t.Error("log not put back")
	}
	err = ReviewSpec(context.Background(), r, saveName, func(Step) {})
	wantErrMsg(t, err, "spec.md is in_review: review 1 has started, and has not finished. Nothing was changed.")
}

func TestReviewMovedFirst(t *testing.T) {
	r, wt, _ := submittedSlice(t)
	var steps []Step
	err := ReviewSpec(context.Background(), r, saveName, func(s Step) {
		steps = append(steps, s)
		if s.Kind == "started" {
			commitEvents(t, wt, eventlog.New(4, fixedNow, "review.finished",
				eventlog.ReviewFinished{Doc: "spec.md", Round: 1, Result: "changes"}))
		}
	})
	wantErrMsg(t, err, "another sliceflow command changed slice login-google first; spec.md is now draft. The steps before it were saved.")
	if len(steps) != 1 {
		t.Errorf("steps %+v", steps)
	}
	if got := gitOut(t, wt, "log", "-1", "--format=%s"); strings.Contains(got, "passed") {
		t.Errorf("head %q", got)
	}
}
