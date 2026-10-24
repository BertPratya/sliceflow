package slice

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BertPratya/sliceflow/internal/eventlog"
	"github.com/BertPratya/sliceflow/internal/testrepo"
)

func specPath(wt string) string {
	return filepath.Join(wt, "docs", "slices", saveName, "spec.md")
}

// writeSpec writes the finished spec (0 errors, 0 warnings) for saveName,
// plus extra, to the worktree's spec.md.
func writeSpec(t *testing.T, wt, extra string) {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "spec", "testdata", "0001-spec.md"))
	if err != nil {
		t.Fatal(err)
	}
	text := strings.Replace(string(b), "slice: 0001-slice-and-events", "slice: "+saveName, 1)
	if err := os.WriteFile(specPath(wt), []byte(text+extra), 0o644); err != nil {
		t.Fatal(err)
	}
}

// commitEvents appends events to the log and commits it with plain git.
func commitEvents(t *testing.T, wt string, evs ...eventlog.Event) {
	t.Helper()
	log := readFile(t, saveLogPath(wt))
	for _, e := range evs {
		line, err := eventlog.Encode(e)
		if err != nil {
			t.Fatal(err)
		}
		log += string(line)
	}
	commitRawLog(t, wt, log)
}

func submitFiles(t *testing.T, wt string) string {
	return strings.TrimSpace(gitOut(t, wt, "show", "--name-only", "--format=", "HEAD"))
}

func wantSubmitFiles() string {
	return "docs/slices/" + saveName + "/.events.jsonl\ndocs/slices/" + saveName + "/spec.md"
}

func TestSubmitRefused(t *testing.T) {
	r, wt := saveSlice(t)
	writeSpec(t, wt, "\n<!-- todo -->\n")
	before := commitCount(t, wt)
	sub, err := Submit(context.Background(), r, saveName)
	if err != nil {
		t.Fatal(err)
	}
	if sub.Refused == nil || sub.Refused.Errors != 0 || sub.Refused.Warnings != 1 {
		t.Fatalf("refused %+v", sub.Refused)
	}
	if sub.Before != "draft" {
		t.Errorf("before %q", sub.Before)
	}
	if len(sub.Problems) == 0 {
		t.Error("no problems")
	}
	if commitCount(t, wt) == before {
		t.Fatal("no commit")
	}
	if got := gitOut(t, wt, "log", "-1", "--format=%s"); got != saveName+": spec submit refused" {
		t.Errorf("message %q", got)
	}
	if got := submitFiles(t, wt); got != wantSubmitFiles() {
		t.Errorf("files %q", got)
	}
	got := lastLine(readFile(t, saveLogPath(wt)))
	if !strings.Contains(got, `"seq":2,"event":"spec.submitted"`) || !strings.HasSuffix(got, `"data":{"result":"refused","errors":0,"warnings":1}}`) {
		t.Errorf("last line %s", got)
	}
}

func TestSubmitAccepted(t *testing.T) {
	r, wt := saveSlice(t)
	writeSpec(t, wt, "")
	before := commitCount(t, wt)
	sub, err := Submit(context.Background(), r, saveName)
	if err != nil {
		t.Fatal(err)
	}
	if sub.Refused != nil {
		t.Fatalf("refused %+v", sub.Refused)
	}
	if commitCount(t, wt) == before {
		t.Fatal("no commit")
	}
	if got := gitOut(t, wt, "log", "-1", "--format=%s"); got != saveName+": spec submitted" {
		t.Errorf("message %q", got)
	}
	if got := submitFiles(t, wt); got != wantSubmitFiles() {
		t.Errorf("files %q", got)
	}
	if got := lastLine(readFile(t, saveLogPath(wt))); !strings.HasSuffix(got, `"data":{"result":"accepted","errors":0,"warnings":0}}`) {
		t.Errorf("last line %s", got)
	}
	if _, err := os.Stat(filepath.Join(wt, "docs", "slices", saveName, "reviews")); err == nil {
		t.Error("reviews folder exists")
	}
	// The state is submitted: a second submit is refused with S16's error.
	_, err = Submit(context.Background(), r, saveName)
	wantErrMsg(t, err, "spec.md is submitted; review it with: sliceflow spec review. Nothing was changed.")
}

func TestSubmitStates(t *testing.T) {
	ctx := context.Background()
	at := fixedNow
	submitted := eventlog.New(2, at, "spec.submitted", eventlog.Submitted{Result: "accepted"})

	t.Run("submitted", func(t *testing.T) {
		r, wt := saveSlice(t)
		writeSpec(t, wt, "")
		commitEvents(t, wt, submitted)
		before := commitCount(t, wt)
		_, err := Submit(ctx, r, saveName)
		wantErrMsg(t, err, "spec.md is submitted; review it with: sliceflow spec review. Nothing was changed.")
		if commitCount(t, wt) != before {
			t.Error("commit made")
		}
	})

	// reviewed makes a review_passed spec: spec.md committed, then the log.
	reviewed := func(t *testing.T, finish bool) (Repo, string) {
		r, wt := saveSlice(t)
		writeSpec(t, wt, "")
		testrepo.Git(t, wt, "add", "--", specPath(wt))
		testrepo.Git(t, wt, "commit", "-m", "spec")
		head := gitOut(t, wt, "rev-parse", "HEAD")
		evs := []eventlog.Event{submitted,
			eventlog.New(3, at, "review.started", eventlog.ReviewStarted{Doc: "spec.md", Round: 1, Reviewer: "placeholder", File: "reviews/spec-1.md", Commit: head})}
		if finish {
			evs = append(evs, eventlog.New(4, at, "review.finished", eventlog.ReviewFinished{Doc: "spec.md", Round: 1, Result: "passed"}))
		}
		commitEvents(t, wt, evs...)
		return r, wt
	}

	t.Run("review_passed unchanged", func(t *testing.T) {
		r, wt := reviewed(t, true)
		before := commitCount(t, wt)
		_, err := Submit(ctx, r, saveName)
		wantErrMsg(t, err, "spec.md has not changed since review 1, which passed. Nothing was changed.")
		if commitCount(t, wt) != before {
			t.Error("commit made")
		}
	})

	t.Run("in_review", func(t *testing.T) {
		r, wt := reviewed(t, false)
		before := commitCount(t, wt)
		_, err := Submit(ctx, r, saveName)
		wantErrMsg(t, err, "spec.md is in_review: review 1 has started, and has not finished. Nothing was changed.")
		if commitCount(t, wt) != before {
			t.Error("commit made")
		}
	})

	t.Run("no slice.created", func(t *testing.T) {
		r, wt := saveSlice(t)
		commitRawLog(t, wt, "not json\n")
		before := commitCount(t, wt)
		_, err := Submit(ctx, r, saveName)
		wantErrMsg(t, err, "slice login-google has no slice.created event, so its state is unknown. Nothing was changed.")
		if commitCount(t, wt) != before {
			t.Error("commit made")
		}
	})

	t.Run("review_passed changed", func(t *testing.T) {
		r, wt := reviewed(t, true)
		b := readFile(t, specPath(wt))
		edited := strings.Replace(b, "title: Slice and events", "title: Slice and events, edited", 1)
		if edited == b {
			t.Fatal("no edit made")
		}
		if err := os.WriteFile(specPath(wt), []byte(edited), 0o644); err != nil {
			t.Fatal(err)
		}
		sub, err := Submit(ctx, r, saveName)
		if err != nil {
			t.Fatal(err)
		}
		if sub.Refused != nil {
			t.Fatalf("refused %+v", sub.Refused)
		}
		if got := gitOut(t, wt, "log", "-1", "--format=%s"); got != saveName+": spec submitted" {
			t.Errorf("message %q", got)
		}
	})

	t.Run("refused keeps review", func(t *testing.T) {
		r, wt := reviewed(t, true)
		writeSpec(t, wt, "\n<!-- todo -->\n")
		sub, err := Submit(ctx, r, saveName)
		if err != nil {
			t.Fatal(err)
		}
		if sub.Refused == nil || sub.Before != "review_passed" {
			t.Errorf("got %+v", sub)
		}
	})
}

func TestSubmitNoName(t *testing.T) {
	ctx := context.Background()
	r, wt := saveSlice(t)
	writeSpec(t, wt, "")
	_, err := Submit(ctx, r, "")
	se, ok := err.(*Error)
	if !ok {
		t.Fatalf("want *Error, got %T: %v", err, err)
	}
	lines := strings.Split(se.Message, "\n")
	if len(lines) != 3 || lines[1] != "Give a name: sliceflow spec submit <name>" {
		t.Errorf("message %q", se.Message)
	}
	// In the worktree: the worktree's slice.
	wr := r
	wr.Dir = wt
	sub, err := Submit(ctx, wr, "")
	if err != nil {
		t.Fatal(err)
	}
	if sub.Refused != nil {
		t.Errorf("refused %+v", sub.Refused)
	}
	if got := gitOut(t, wt, "log", "-1", "--format=%s"); got != saveName+": spec submitted" {
		t.Errorf("message %q", got)
	}
}

func TestSubmitNoSpec(t *testing.T) {
	r, wt := saveSlice(t)
	if err := os.Remove(specPath(wt)); err != nil {
		t.Fatal(err)
	}
	before := commitCount(t, wt)
	log := readFile(t, saveLogPath(wt))
	_, err := Submit(context.Background(), r, saveName)
	wantErrMsg(t, err, ".sliceflow/worktrees/login-google/docs/slices/login-google/spec.md not found. Nothing was changed.")
	if commitCount(t, wt) != before || readFile(t, saveLogPath(wt)) != log {
		t.Error("something changed")
	}
}
