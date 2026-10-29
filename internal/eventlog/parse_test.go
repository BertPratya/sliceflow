package eventlog

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

// hex64 is a good 64-character commit id.
const hex64 = "9dd7ebf4c2a1e8d3b5f6a7c8d9e0f1a2b3c4d5e69dd7ebf4c2a1e8d3b5f6a7c8"

// goodLine returns a good slice.created line with the given seq.
func goodLine(seq string) string {
	return `{"version":1,"seq":` + seq + `,"event":"slice.created","at":"2026-10-24T07:42:00Z","data":{"from":{"branch":"main","commit":"` + commit + `"}}}`
}

// TestParseOneBadLine checks the reason for each kind of bad line (R26).
func TestParseOneBadLine(t *testing.T) {
	at := `"at":"2026-10-24T07:42:00Z"`
	from := `"from":{"branch":"main","commit":"` + commit + `"}`
	mk := func(top string, data string) string { return `{` + top + `,"data":` + data + `}` }
	head := `"version":1,"seq":1,"event":"slice.created",` + at

	tests := []struct {
		name   string
		line   string
		reason string // "" means the line is good
	}{
		{"blank empty", "", "blank line"},
		{"blank spaces", "   ", "blank line"},
		{"not json", "{nope", "not valid JSON"},
		{"json not object", "[1,2]", "not valid JSON"},
		{"json number", "7", "not valid JSON"},
		{"dup top level", `{"version":1,"seq":1,"seq":2}`, `field "seq" appears twice`},
		{"dup in data", mk(head, `{`+from+`,"title":"a","title":"b"}`), `field "data.title" appears twice`},
		{"dup in data.from", mk(head, `{"from":{"branch":"a","branch":"b","commit":"`+commit+`"}}`), `field "data.from.branch" appears twice`},
		{"version missing", `{"seq":1}`, `missing field "version"`},
		{"version string", `{"version":"1"}`, `field "version" has the wrong type`},
		{"version float", `{"version":1.5}`, `field "version" has the wrong type`},
		{"version 2", `{"version":2}`, "unknown version 2"},
		{"event missing", `{"version":1}`, `missing field "event"`},
		{"event number", `{"version":1,"event":5}`, `field "event" has the wrong type`},
		{"event unknown", `{"version":1,"event":"x.y"}`, `unknown event "x.y"`},
		{"unknown top field", `{"version":1,"event":"slice.created","extra":1}`, `unknown field "extra"`},
		{"seq missing", `{"version":1,"event":"slice.created",` + at + `,"data":{}}`, `missing field "seq"`},
		{"at missing", `{"version":1,"seq":1,"event":"slice.created","data":{}}`, `missing field "at"`},
		{"data missing", `{"version":1,"seq":1,"event":"slice.created",` + at + `}`, `missing field "data"`},
		{"seq string", `{"version":1,"seq":"1","event":"slice.created",` + at + `,"data":{}}`, `field "seq" has the wrong type`},
		{"at number", `{"version":1,"seq":1,"event":"slice.created","at":5,"data":{}}`, `field "at" has the wrong type`},
		{"data array", `{"version":1,"seq":1,"event":"slice.created",` + at + `,"data":[]}`, `field "data" has the wrong type`},
		{"at wrong shape", `{"version":1,"seq":1,"event":"slice.created","at":"2026-10-24 07:42:00",` + `"data":{}}`, `field "at" has a wrong value`},
		{"at offset not Z", `{"version":1,"seq":1,"event":"slice.created","at":"2026-10-24T07:42:00+00:00","data":{}}`, `field "at" has a wrong value`},
		{"at not a real time", `{"version":1,"seq":1,"event":"slice.created","at":"2026-13-40T99:00:00Z","data":{}}`, `field "at" has a wrong value`},
		{"unknown data field", mk(head, `{`+from+`,"x":1}`), `unknown field "data.x"`},
		{"unknown data.from field", mk(head, `{"from":{"branch":"a","commit":"`+commit+`","z":1}}`), `unknown field "data.from.z"`},
		{"data.from missing", mk(head, `{}`), `missing field "data.from"`},
		{"branch missing", mk(head, `{"from":{"commit":"`+commit+`"}}`), `missing field "data.from.branch"`},
		{"commit missing", mk(head, `{"from":{"branch":"main"}}`), `missing field "data.from.commit"`},
		{"from string", mk(head, `{"from":"main"}`), `field "data.from" has the wrong type`},
		{"branch number", mk(head, `{"from":{"branch":1,"commit":"`+commit+`"}}`), `field "data.from.branch" has the wrong type`},
		{"commit null", mk(head, `{"from":{"branch":"a","commit":null}}`), `field "data.from.commit" has the wrong type`},
		{"title number", mk(head, `{`+from+`,"title":3}`), `field "data.title" has the wrong type`},
		{"branch empty", mk(head, `{"from":{"branch":"","commit":"`+commit+`"}}`), `field "data.from.branch" has a wrong value`},
		{"commit short", mk(head, `{"from":{"branch":"a","commit":"abc123"}}`), `field "data.from.commit" has a wrong value`},
		{"commit uppercase", mk(head, `{"from":{"branch":"a","commit":"`+strings.ToUpper(commit)+`"}}`), `field "data.from.commit" has a wrong value`},
		{"title empty", mk(head, `{`+from+`,"title":""}`), `field "data.title" has a wrong value`},
		{"good with title", mk(head, `{`+from+`,"title":"T"}`), ""},
		{"good 64-char commit", mk(head, `{"from":{"branch":"a","commit":"`+hex64+`"}}`), ""},
		{"good other order", `{"data":{"title":"T",` + from + `},"at":"2026-10-24T07:42:00Z","event":"slice.created","seq":1,"version":1}`, ""},
		{"good CRLF", goodLine("1") + "\r", ""},
		{"seq too big", `{"version":1,"seq":99999999999999999999,"event":"slice.created",` + at + `,"data":{}}`, `field "seq" has a wrong value`},
		{"seq 1.0", `{"version":1,"seq":1.0,"event":"slice.created",` + at + `,"data":{}}`, `field "seq" has the wrong type`},
		{"seq 1e0", `{"version":1,"seq":1e0,"event":"slice.created",` + at + `,"data":{}}`, `field "seq" has the wrong type`},
		{"data null", `{"version":1,"seq":1,"event":"slice.created",` + at + `,"data":null}`, `field "data" has the wrong type`},
		{"title null", mk(head, `{`+from+`,"title":null}`), `field "data.title" has the wrong type`},
		{"dup written with escape", `{"version":1,"seq":1,"\u0073eq":2}`, `field "seq" appears twice`},
		// one reason only, in our order
		{"version before event", `{"version":2,"event":"nope"}`, "unknown version 2"},
		{"dup before version", `{"version":2,"version":1}`, `field "version" appears twice`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			l := Parse([]byte(tc.line + "\n"))
			if tc.reason == "" {
				if len(l.Events) != 1 || len(l.Warnings) != 0 {
					t.Fatalf("want one good event, got %d events, warnings %q", len(l.Events), l.Warnings)
				}
				return
			}
			want := ".events.jsonl line 1: " + tc.reason + "; skipped"
			if len(l.Events) != 0 || !reflect.DeepEqual(l.Warnings, []string{want}) {
				t.Fatalf("got events %d, warnings %q, want %q", len(l.Events), l.Warnings, want)
			}
		})
	}
}

// TestParseS17 is S17: a bad line 2 and a line 3 with seq 4.
func TestParseS17(t *testing.T) {
	content := goodLine("1") + "\n{bad\n" + goodLine("4") + "\n"
	l := Parse([]byte(content))
	want := []string{
		".events.jsonl line 2: not valid JSON; skipped",
		".events.jsonl line 3: seq 4, expected 2",
	}
	if !reflect.DeepEqual(l.Warnings, want) {
		t.Fatalf("warnings = %q, want %q", l.Warnings, want)
	}
	if len(l.Events) != 2 {
		t.Fatalf("events = %d, want 2", len(l.Events))
	}
}

// TestParseSeq checks the seq rule of R27.
func TestParseSeq(t *testing.T) {
	join := func(lines ...string) []byte { return []byte(strings.Join(lines, "\n") + "\n") }
	tests := []struct {
		name    string
		content []byte
		want    []string
		events  int
	}{
		{"1 2 3", join(goodLine("1"), goodLine("2"), goodLine("3")), nil, 3},
		{"1 2 4 5", join(goodLine("1"), goodLine("2"), goodLine("4"), goodLine("5")),
			[]string{".events.jsonl line 3: seq 4, expected 3"}, 4},
		{"first is 3", join(goodLine("3")),
			[]string{".events.jsonl line 1: seq 3, expected 1"}, 1},
		{"skipped bad line does not count", join(goodLine("1"), `{"seq":2}`, goodLine("2")),
			[]string{".events.jsonl line 2: missing field \"version\"; skipped"}, 2},
		{"no final newline", []byte(goodLine("1") + "\n" + goodLine("2")), nil, 2},
		{"empty file", nil, nil, 0},
		{"blank line in the middle", join(goodLine("1"), "", goodLine("2")),
			[]string{".events.jsonl line 2: blank line; skipped"}, 2},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			l := Parse(tc.content)
			if !reflect.DeepEqual(l.Warnings, tc.want) {
				t.Errorf("warnings = %q, want %q", l.Warnings, tc.want)
			}
			if len(l.Events) != tc.events {
				t.Errorf("events = %d, want %d", len(l.Events), tc.events)
			}
		})
	}
}

// TestParseReadsBackLine checks that a line written by Line reads back as
// one good event, and that Phase and Created see it.
func TestParseReadsBackLine(t *testing.T) {
	at := time.Date(2026, 10, 24, 7, 42, 0, 0, time.UTC)
	want := SliceCreated{From: Ref{Branch: "main", Commit: commit}, Title: "Login"}
	data, err := Data(want)
	if err != nil {
		t.Fatal(err)
	}
	b, err := Line(Event{Version: Version, Seq: 1, Event: EventSliceCreated, At: at, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	l := Parse(b)
	if len(l.Warnings) != 0 || len(l.Events) != 1 {
		t.Fatalf("events %d, warnings %q", len(l.Events), l.Warnings)
	}
	e := l.Events[0]
	if e.Version != 1 || e.Seq != 1 || e.Event != EventSliceCreated || !e.At.Equal(at) || string(e.Data) != string(data) {
		t.Errorf("event = %+v", e)
	}
	ce, cd, ok := l.Created()
	if !ok || ce.Seq != 1 || cd != want {
		t.Errorf("Created = %+v, %+v, %v", ce, cd, ok)
	}
}

// TestCreatedMissing checks that a log without a good slice.created has none.
func TestCreatedMissing(t *testing.T) {
	for _, content := range []string{"", "{bad\n"} {
		l := Parse([]byte(content))
		if _, _, ok := l.Created(); ok {
			t.Errorf("Created(%q) ok = true", content)
		}
	}
}

// TestParseCRLF checks a file with CRLF line endings and several good lines.
func TestParseCRLF(t *testing.T) {
	content := goodLine("1") + "\r\n" + goodLine("2") + "\r\n" + goodLine("3") + "\r\n"
	l := Parse([]byte(content))
	if len(l.Events) != 3 || len(l.Warnings) != 0 {
		t.Fatalf("events %d, warnings %q", len(l.Events), l.Warnings)
	}
}

// TestParseDeepIsFast checks that a line nested 5,000 levels deep is read
// as a bad line, and quickly.
func TestParseDeepIsFast(t *testing.T) {
	const depth = 5000
	deep := `{"version":1,"seq":1,"event":"slice.created","at":"2026-10-24T07:42:00Z","data":{"x":` +
		strings.Repeat(`{"a":`, depth) + "1" + strings.Repeat("}", depth) + "}}\n"
	start := time.Now()
	l := Parse([]byte(deep))
	if d := time.Since(start); d > 200*time.Millisecond {
		t.Errorf("Parse took %v, want under 200ms", d)
	}
	if len(l.Events) != 0 || len(l.Warnings) != 1 {
		t.Errorf("events %d, warnings %q, want one bad line", len(l.Events), l.Warnings)
	}
}

// FuzzParse checks that Parse never panics, and that every event it
// returns is a good slice.created event.
func FuzzParse(f *testing.F) {
	f.Add([]byte(goodLine("1") + "\n"))
	f.Add([]byte("{bad\n\n[1]\n"))
	f.Add([]byte(`{"version":1,"seq":1,"seq":2}` + "\r\n"))
	f.Add([]byte(`{"version":2}`))
	f.Fuzz(func(t *testing.T, content []byte) {
		l := Parse(content)
		for _, e := range l.Events {
			if e.Version != Version || e.Event != EventSliceCreated {
				t.Fatalf("bad event returned: %+v", e)
			}
		}
		if len(l.Events) > 0 {
			if _, _, ok := l.Created(); !ok {
				t.Fatal("Created not ok with events")
			}
		}
	})
}
