package eventlog

import (
	"bytes"
	"testing"
	"time"
)

const commit = "9dd7ebf4c2a1e8d3b5f6a7c8d9e0f1a2b3c4d5e6"

// makeLine builds the line of a slice.created event for the tests.
func makeLine(t *testing.T, at time.Time, title string) string {
	t.Helper()
	data, err := Data(SliceCreated{From: Ref{Branch: "main", Commit: commit}, Title: title})
	if err != nil {
		t.Fatalf("Data: %v", err)
	}
	b, err := Line(Event{Version: Version, Seq: 1, Event: EventSliceCreated, At: at, Data: data})
	if err != nil {
		t.Fatalf("Line: %v", err)
	}
	return string(b)
}

// TestLine checks the exact text of the line for several inputs.
func TestLine(t *testing.T) {
	utc := time.Date(2026, 10, 24, 7, 42, 0, 0, time.UTC)
	plus7 := time.Date(2026, 10, 24, 14, 42, 0, 0, time.FixedZone("x", 7*3600))
	nanos := time.Date(2026, 10, 24, 7, 42, 0, 999999999, time.UTC)
	head := `{"version":1,"seq":1,"event":"slice.created","at":"2026-10-24T07:42:00Z","data":{"from":{"branch":"main","commit":"` + commit + `"}`

	tests := []struct {
		name  string
		at    time.Time
		title string
		want  string
	}{
		{"S1 no title", utc, "", head + "}}\n"},
		{"S2 title after from", utc, "Login with Google", head + `,"title":"Login with Google"}}` + "\n"},
		{"html not escaped", utc, "a & b <c>", head + `,"title":"a & b <c>"}}` + "\n"},
		{"other time zone becomes UTC", plus7, "", head + "}}\n"},
		{"nanoseconds are cut", nanos, "", head + "}}\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeLine(t, tt.at, tt.title); got != tt.want {
				t.Errorf("got  %q\nwant %q", got, tt.want)
			}
		})
	}
}

// TestLineErrors checks that Line refuses what would break R19.
func TestLineErrors(t *testing.T) {
	at := time.Date(2026, 10, 24, 7, 42, 0, 0, time.UTC)
	tests := []struct {
		name string
		e    Event
	}{
		{"version 0", Event{Version: 0, Seq: 1, Event: EventSliceCreated, At: at, Data: []byte("{}")}},
		{"version 2", Event{Version: 2, Seq: 1, Event: EventSliceCreated, At: at, Data: []byte("{}")}},
		{"empty data", Event{Version: Version, Seq: 1, Event: EventSliceCreated, At: at}},
		{"data is an array", Event{Version: Version, Seq: 1, Event: EventSliceCreated, At: at, Data: []byte("[1]")}},
		{"data is a string", Event{Version: Version, Seq: 1, Event: EventSliceCreated, At: at, Data: []byte(`"x"`)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := Line(tt.e); err == nil {
				t.Errorf("want an error, got %q", got)
			}
		})
	}
}

// TestLineShape checks one newline at the end (even when the title has
// newlines), no BOM, and that non-ASCII text is written as it is.
func TestLineShape(t *testing.T) {
	at := time.Date(2026, 10, 24, 7, 42, 0, 0, time.UTC)

	got := []byte(makeLine(t, at, "a\nb\r\n"))
	if !bytes.HasSuffix(got, []byte("\n")) || bytes.Count(got, []byte("\n")) != 1 {
		t.Errorf("want exactly one newline at the end, got %q", got)
	}
	if bytes.HasPrefix(got, []byte("\xef\xbb\xbf")) {
		t.Error("line starts with a BOM")
	}

	got = []byte(makeLine(t, at, "Ünï 日本"))
	if !bytes.Contains(got, []byte(`"title":"Ünï 日本"`)) {
		t.Errorf("non-ASCII title was changed: %q", got)
	}
}
