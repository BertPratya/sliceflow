package eventlog

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseGoodLog(t *testing.T) {
	for name, log := range map[string]string{
		"lf":   goodLine + "\n",
		"crlf": goodLine + "\r\n",
		"none": goodLine,
	} {
		events, warns := Parse([]byte(log))
		if len(events) != 1 || len(warns) != 0 {
			t.Fatalf("%s: events %v warns %v", name, events, warns)
		}
		want := Created{From: From{Branch: "main", Commit: goodCommit}}
		if !reflect.DeepEqual(events[0].Data, want) {
			t.Errorf("%s: data %v", name, events[0].Data)
		}
		if Phase(events) != "planning" {
			t.Errorf("%s: phase", name)
		}
	}
}

func TestParseBadLines(t *testing.T) {
	c := goodCommit
	line := func(s string) string { return s }
	rows := []struct{ name, bad, reason string }{
		{"blank", "   ", "blank line"},
		{"json", "{", "not valid JSON"},
		{"version", `{"version":2,"seq":2,"event":"slice.created","at":"2026-10-24T07:42:00Z","data":{"from":{"branch":"main","commit":"` + c + `"}}}`, "unknown version 2"},
		{"event", `{"version":1,"seq":2,"event":"slice.moved","at":"2026-10-24T07:42:00Z","data":{"from":{"branch":"main","commit":"` + c + `"}}}`, `unknown event "slice.moved"`},
		{"unknown", `{"version":1,"seq":2,"event":"slice.created","at":"2026-10-24T07:42:00Z","by":"x","data":{"from":{"branch":"main","commit":"` + c + `"}}}`, `unknown field "by"`},
		{"twice", `{"version":1,"seq":2,"seq":2,"event":"slice.created","at":"2026-10-24T07:42:00Z","data":{"from":{"branch":"main","commit":"` + c + `"}}}`, `field "seq" appears twice`},
		{"twice nested", `{"version":1,"seq":2,"event":"slice.created","at":"2026-10-24T07:42:00Z","data":{"from":{"branch":"main","branch":"x","commit":"` + c + `"}}}`, `field "data.from.branch" appears twice`},
		{"missing", `{"version":1,"seq":2,"event":"slice.created","data":{"from":{"branch":"main","commit":"` + c + `"}}}`, `missing field "at"`},
		{"type string", `{"version":1,"seq":"2","event":"slice.created","at":"2026-10-24T07:42:00Z","data":{"from":{"branch":"main","commit":"` + c + `"}}}`, `field "seq" has the wrong type`},
		{"type float", `{"version":1,"seq":2.0,"event":"slice.created","at":"2026-10-24T07:42:00Z","data":{"from":{"branch":"main","commit":"` + c + `"}}}`, `field "seq" has the wrong type`},
		{"value at", `{"version":1,"seq":2,"event":"slice.created","at":"2026-10-24 07:42","data":{"from":{"branch":"main","commit":"` + c + `"}}}`, `field "at" has a wrong value`},
		{"value commit", `{"version":1,"seq":2,"event":"slice.created","at":"2026-10-24T07:42:00Z","data":{"from":{"branch":"main","commit":"9dd7ebf"}}}`, `field "data.from.commit" has a wrong value`},
		{"value branch", `{"version":1,"seq":2,"event":"slice.created","at":"2026-10-24T07:42:00Z","data":{"from":{"branch":"","commit":"` + c + `"}}}`, `field "data.from.branch" has a wrong value`},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			events, warns := Parse([]byte(goodLine + "\n" + line(r.bad) + "\n"))
			want := ".events.jsonl line 2: " + r.reason + "; skipped"
			if len(events) != 1 || len(warns) != 1 || warns[0] != want {
				t.Errorf("events %d warns %q want %q", len(events), warns, want)
			}
		})
	}
}

func TestParseFirstFaultWins(t *testing.T) {
	bad := `{"version":1,"seq":2,"event":"slice.created","by":"x","data":{"from":{"branch":"main","commit":"` + goodCommit + `"}}}`
	_, warns := Parse([]byte(goodLine + "\n" + bad + "\n"))
	want := `.events.jsonl line 2: unknown field "by"; skipped`
	if len(warns) != 1 || warns[0] != want {
		t.Errorf("warns %q", warns)
	}
}

func TestParseSeqGap(t *testing.T) {
	var sb strings.Builder
	for _, s := range []string{"1", "2", "4", "5"} {
		sb.WriteString(strings.Replace(goodLine, `"seq":1`, `"seq":`+s, 1) + "\n")
	}
	events, warns := Parse([]byte(sb.String()))
	if len(events) != 4 || len(warns) != 1 || warns[0] != ".events.jsonl line 3: seq 4, expected 3" {
		t.Errorf("events %d warns %q", len(events), warns)
	}
}

func TestPhaseUnknown(t *testing.T) {
	if Phase(nil) != "unknown" {
		t.Error("want unknown")
	}
}
