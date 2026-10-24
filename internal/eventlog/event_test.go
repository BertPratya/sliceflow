package eventlog

import (
	"testing"
	"time"
)

const goodCommit = "9dd7ebf4c2a1e8d3b5f6a7c8d9e0f1a2b3c4d5e6"
const goodLine = `{"version":1,"seq":1,"event":"slice.created","at":"2026-10-24T07:42:00Z","data":{"from":{"branch":"main","commit":"` + goodCommit + `"}}}`

func TestEncodeCreated(t *testing.T) {
	at := time.Date(2026, 10, 24, 7, 42, 0, 0, time.UTC)
	from := From{Branch: "main", Commit: goodCommit}
	got, err := Encode(NewCreated(at, from, ""))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != goodLine+"\n" {
		t.Errorf("got %q", got)
	}
	got, _ = Encode(NewCreated(at, from, "Login with Google"))
	want := goodLine[:len(goodLine)-2] + `,"title":"Login with Google"}}` + "\n"
	if string(got) != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestEncodeLocalTimeIsUTC(t *testing.T) {
	at := time.Date(2026, 10, 24, 9, 42, 0, 0, time.FixedZone("x", 2*3600))
	got, _ := Encode(NewCreated(at, From{Branch: "main", Commit: goodCommit}, ""))
	if string(got) != goodLine+"\n" {
		t.Errorf("got %q", got)
	}
}

var newEventCases = []struct {
	ev   Event
	line string
}{
	{New(2, time.Date(2026, 10, 24, 7, 50, 0, 0, time.UTC), "slice.titled", Titled{Title: "Sign in with Google"}),
		`{"version":1,"seq":2,"event":"slice.titled","at":"2026-10-24T07:50:00Z","data":{"title":"Sign in with Google"}}`},
	{New(2, time.Date(2026, 10, 24, 9, 10, 0, 0, time.UTC), "spec.submitted", Submitted{Result: "refused", Errors: 0, Warnings: 1}),
		`{"version":1,"seq":2,"event":"spec.submitted","at":"2026-10-24T09:10:00Z","data":{"result":"refused","errors":0,"warnings":1}}`},
	{New(4, time.Date(2026, 10, 24, 10, 5, 0, 0, time.UTC), "review.started", ReviewStarted{Doc: "spec.md", Round: 1, Reviewer: "placeholder", File: "reviews/spec-1.md", Commit: "3f2a1c9e5b7d4a6f8c0e2b4d6f8a0c2e4b6d8f0a"}),
		`{"version":1,"seq":4,"event":"review.started","at":"2026-10-24T10:05:00Z","data":{"doc":"spec.md","round":1,"reviewer":"placeholder","file":"reviews/spec-1.md","commit":"3f2a1c9e5b7d4a6f8c0e2b4d6f8a0c2e4b6d8f0a"}}`},
	{New(5, time.Date(2026, 10, 24, 10, 5, 0, 0, time.UTC), "review.finished", ReviewFinished{Doc: "spec.md", Round: 1, Result: "passed"}),
		`{"version":1,"seq":5,"event":"review.finished","at":"2026-10-24T10:05:00Z","data":{"doc":"spec.md","round":1,"result":"passed"}}`},
}

func TestEncodeNewEvents(t *testing.T) {
	for _, c := range newEventCases {
		got, err := Encode(c.ev)
		if err != nil || string(got) != c.line+"\n" {
			t.Errorf("%s: got %q err %v", c.ev.Event, got, err)
		}
	}
}
