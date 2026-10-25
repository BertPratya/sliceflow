package state

import (
	"fmt"
	"testing"
	"time"

	"github.com/BertPratya/sliceflow/internal/eventlog"
)

var t0 = time.Date(2026, 10, 24, 7, 42, 0, 0, time.UTC)

func created() eventlog.Event {
	return eventlog.NewCreated(t0, eventlog.From{Branch: "main", Commit: "9dd7ebf"}, "")
}

func titled(title string) eventlog.Event {
	return eventlog.New(2, t0, "slice.titled", eventlog.Titled{Title: title})
}

func submitted(result string, errs, warns int) eventlog.Event {
	return eventlog.New(2, t0, "spec.submitted", eventlog.Submitted{Result: result, Errors: errs, Warnings: warns})
}

func started(round int) eventlog.Event {
	return eventlog.New(2, t0, "review.started", eventlog.ReviewStarted{
		Doc: "spec.md", Round: round, Reviewer: fmt.Sprintf("r%d", round),
		File: fmt.Sprintf("reviews/spec-%d.md", round), Commit: fmt.Sprintf("abc123%d", round)})
}

func finished(round int, result string) eventlog.Event {
	return eventlog.New(2, t0, "review.finished", eventlog.ReviewFinished{Doc: "spec.md", Round: round, Result: result})
}

// numbered sets Line and Seq as a log would.
func numbered(evs ...eventlog.Event) []eventlog.Event {
	for i := range evs {
		evs[i].Line = i + 1
		evs[i].Seq = i + 1
	}
	return evs
}

func TestMoves(t *testing.T) {
	all := []string{"draft", "submitted", "in_review", "review_passed"}
	rows := []struct {
		name  string
		ev    eventlog.Event
		from  []string
		to    string // "" means as it was
		round int    // round the state is in, for review.finished
	}{
		{"submit accepted", submitted("accepted", 0, 1), []string{"draft", "review_passed"}, "submitted", 0},
		{"submit refused", submitted("refused", 2, 1), []string{"draft", "review_passed"}, "", 0},
		{"review started", started(1), []string{"submitted"}, "in_review", 0},
		{"finished passed", finished(1, "passed"), []string{"in_review"}, "review_passed", 1},
		{"finished changes", finished(1, "changes"), []string{"in_review"}, "draft", 1},
		{"finished needs_human", finished(1, "needs_human"), []string{"in_review"}, "draft", 1},
		{"titled", titled("New"), all, "", 0},
		{"created again", created(), nil, "", 0},
	}
	for _, r := range rows {
		for _, from := range all {
			ok := false
			for _, f := range r.from {
				if f == from {
					ok = true
				}
			}
			s := State{Phase: "planning", Spec: Doc{Status: from, Review: &Review{Round: 1}, MaxRound: 1, Refused: &Counts{1, 1}}}
			reason := Allowed(s, r.ev)
			if ok != (reason == "") {
				t.Errorf("%s from %s: allowed=%v, reason %q", r.name, from, reason == "", reason)
				continue
			}
			if !ok {
				continue
			}
			next := apply(s, r.ev)
			want := r.to
			if want == "" {
				want = from
			}
			if next.Spec.Status != want {
				t.Errorf("%s from %s: status %q, want %q", r.name, from, next.Spec.Status, want)
			}
		}
	}

	// slice.titled sets Title in every state.
	for _, from := range all {
		s := apply(State{Phase: "planning", Title: "old", Spec: Doc{Status: from}}, titled("New"))
		if s.Title != "New" || s.Spec.Status != from {
			t.Errorf("titled from %s: %+v", from, s)
		}
	}

	// Refused is set by a refused submit, cleared by an accepted one or a review start.
	s := State{Phase: "planning", Spec: Doc{Status: "draft"}}
	s = apply(s, submitted("refused", 2, 3))
	if s.Spec.Status != "draft" || s.Spec.Refused == nil || *s.Spec.Refused != (Counts{2, 3}) {
		t.Fatalf("refused: %+v", s.Spec)
	}
	s = apply(s, submitted("accepted", 0, 0))
	if s.Spec.Refused != nil || s.Spec.Status != "submitted" {
		t.Fatalf("accepted: %+v", s.Spec)
	}
	s.Spec.Refused = &Counts{1, 1}
	s = apply(s, started(1))
	if s.Spec.Refused != nil || s.Spec.Status != "in_review" {
		t.Fatalf("started: %+v", s.Spec)
	}
}

func TestReplayCreated(t *testing.T) {
	s, w := Replay(numbered(created()))
	if s.Phase != "planning" || s.Spec.Status != "draft" || len(w) != 0 {
		t.Errorf("created: %+v %v", s, w)
	}
	s, w = Replay(numbered(eventlog.NewCreated(t0, eventlog.From{Branch: "main", Commit: "9dd7ebf"}, "First title")))
	if s.Title != "First title" || len(w) != 0 {
		t.Errorf("created with title: %+v %v", s, w)
	}
	s, w = Replay(nil)
	if s.Phase != "unknown" || s.Spec.Status != "" || len(w) != 0 {
		t.Errorf("empty: %+v %v", s, w)
	}
}

func TestReplayNotAllowed(t *testing.T) {
	// S3: review.finished at line 4 with no review started.
	s, w := Replay(numbered(created(), titled("A"), submitted("refused", 1, 0), finished(1, "passed")))
	want := ".events.jsonl line 4: review.finished is not allowed when spec.md is draft; skipped"
	if len(w) != 1 || w[0] != want || s.Spec.Status != "draft" {
		t.Errorf("S3: %v %+v", w, s)
	}

	_, w = Replay(numbered(titled("A")))
	want = ".events.jsonl line 1: slice.titled comes before slice.created; skipped"
	if len(w) != 1 || w[0] != want {
		t.Errorf("before created: %v", w)
	}

	_, w = Replay(numbered(created(), created()))
	want = ".events.jsonl line 2: slice.created is not allowed when spec.md is draft; skipped"
	if len(w) != 1 || w[0] != want {
		t.Errorf("second created: %v", w)
	}

	_, w = Replay(numbered(created(), submitted("accepted", 0, 0), started(1), finished(2, "passed")))
	want = ".events.jsonl line 4: review.finished is for round 2, but review 1 is running; skipped"
	if len(w) != 1 || w[0] != want {
		t.Errorf("round mismatch: %v", w)
	}
	evs := numbered(created(), submitted("accepted", 0, 0), started(1), submitted("accepted", 0, 0), finished(2, "passed"))
	_, w = Replay(evs)
	want = ".events.jsonl line 5: review.finished is for round 2, but review 1 is running; skipped"
	if len(w) != 2 || w[1] != want {
		t.Errorf("round mismatch 2: %v", w)
	}
}

func TestReplayRounds(t *testing.T) {
	evs := numbered(created(),
		submitted("accepted", 0, 0), started(1), finished(1, "passed"),
		started(2), // skipped: spec is review_passed
		submitted("accepted", 0, 0), started(2), finished(2, "passed"))
	s, w := Replay(evs)
	if len(w) != 1 {
		t.Fatalf("warnings: %v", w)
	}
	if s.Spec.MaxRound != 2 || s.Spec.Review == nil || s.Spec.Review.Round != 2 || s.Spec.Review.Result != "passed" {
		t.Errorf("state: %+v %+v", s.Spec, s.Spec.Review)
	}
	if r := s.Spec.Review; r.Reviewer != "r2" || r.File != "reviews/spec-2.md" || r.Commit != "abc1232" {
		t.Errorf("review fields: %+v", r)
	}
	if s.Spec.SubmitSeq != 6 {
		t.Errorf("SubmitSeq %d", s.Spec.SubmitSeq)
	}

	// A skipped review.started does not count.
	s, _ = Replay(numbered(created(), started(1)))
	if s.Spec.MaxRound != 0 || s.Spec.Review != nil {
		t.Errorf("skipped start counted: %+v", s.Spec)
	}
}
