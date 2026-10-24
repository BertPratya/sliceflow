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
		{"cr only", "\r", "blank line"},
		{"version", `{"version":2,"seq":2,"event":"slice.created","at":"2026-10-24T07:42:00Z","data":{"from":{"branch":"main","commit":"` + c + `"}}}`, "unknown version 2"},
		{"event", `{"version":1,"seq":2,"event":"slice.moved","at":"2026-10-24T07:42:00Z","data":{"from":{"branch":"main","commit":"` + c + `"}}}`, `unknown event "slice.moved"`},
		{"unknown", `{"version":1,"seq":2,"event":"slice.created","at":"2026-10-24T07:42:00Z","by":"x","data":{"from":{"branch":"main","commit":"` + c + `"}}}`, `unknown field "by"`},
		{"twice", `{"version":1,"seq":2,"seq":2,"event":"slice.created","at":"2026-10-24T07:42:00Z","data":{"from":{"branch":"main","commit":"` + c + `"}}}`, `field "seq" appears twice`},
		{"twice nested", `{"version":1,"seq":2,"event":"slice.created","at":"2026-10-24T07:42:00Z","data":{"from":{"branch":"main","branch":"x","commit":"` + c + `"}}}`, `field "data.from.branch" appears twice`},
		{"missing", `{"version":1,"seq":2,"event":"slice.created","data":{"from":{"branch":"main","commit":"` + c + `"}}}`, `missing field "at"`},
		{"type string", `{"version":1,"seq":"2","event":"slice.created","at":"2026-10-24T07:42:00Z","data":{"from":{"branch":"main","commit":"` + c + `"}}}`, `field "seq" has the wrong type`},
		{"type float", `{"version":1,"seq":2.0,"event":"slice.created","at":"2026-10-24T07:42:00Z","data":{"from":{"branch":"main","commit":"` + c + `"}}}`, `field "seq" has the wrong type`},
		{"type huge", `{"version":1,"seq":99999999999999999999,"event":"slice.created","at":"2026-10-24T07:42:00Z","data":{"from":{"branch":"main","commit":"` + c + `"}}}`, `field "seq" has the wrong type`},
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

func TestParseSeqFirstGoodLine(t *testing.T) {
	three := strings.Replace(goodLine, `"seq":1`, `"seq":3`, 1)
	events, warns := Parse([]byte(three + "\n"))
	if len(events) != 1 || len(warns) != 1 || warns[0] != ".events.jsonl line 1: seq 3, expected 1" {
		t.Errorf("events %d warns %q", len(events), warns)
	}

	two := strings.Replace(goodLine, `"seq":1`, `"seq":2`, 1)
	events, warns = Parse([]byte("{\n" + two + "\n"))
	want := []string{
		".events.jsonl line 1: not valid JSON; skipped",
		".events.jsonl line 2: seq 2, expected 1",
	}
	if len(events) != 1 || !reflect.DeepEqual(warns, want) {
		t.Errorf("events %d warns %q", len(events), warns)
	}
}

func TestPhaseUnknown(t *testing.T) {
	if Phase(nil) != "unknown" {
		t.Error("want unknown")
	}
}

func TestParseNewEvents(t *testing.T) {
	var log string
	for _, c := range newEventCases {
		log += c.line + "\n"
	}
	events, warns := Parse([]byte(log))
	if len(events) != 4 {
		t.Fatalf("events %v warns %v", events, warns)
	}
	for i, c := range newEventCases {
		want := c.ev
		want.Line = i + 1
		if !reflect.DeepEqual(events[i], want) {
			t.Errorf("line %d: got %+v want %+v", i+1, events[i], want)
		}
	}
	// seq gaps are warned, not skipped; here seq 2,2,4,5 gives warnings
	if len(warns) == 0 {
		t.Error("expected seq warnings")
	}
}

func TestParseNewEventsBad(t *testing.T) {
	c := goodCommit
	ln := func(ev, data string) string {
		return `{"version":1,"seq":2,"event":"` + ev + `","at":"2026-10-24T07:50:00Z","data":` + data + `}`
	}
	started := func(round, doc, commit string) string {
		return ln("review.started", `{"doc":`+doc+`,"round":`+round+`,"reviewer":"r","file":"f","commit":`+commit+`}`)
	}
	q := `"` + c + `"`
	rows := []struct{ name, bad, reason string }{
		{"title empty", ln("slice.titled", `{"title":""}`), `field "data.title" has a wrong value`},
		{"title type", ln("slice.titled", `{"title":3}`), `field "data.title" has the wrong type`},
		{"title extra", ln("slice.titled", `{"title":"a","x":1}`), `unknown field "data.x"`},
		{"title missing", ln("slice.titled", `{}`), `missing field "data.title"`},
		{"data type", ln("slice.titled", `3`), `field "data" has the wrong type`},
		{"result", ln("spec.submitted", `{"result":"maybe","errors":0,"warnings":0}`), `field "data.result" has a wrong value`},
		{"errors", ln("spec.submitted", `{"result":"accepted","errors":-1,"warnings":0}`), `field "data.errors" has a wrong value`},
		{"warnings type", ln("spec.submitted", `{"result":"accepted","errors":0,"warnings":1.5}`), `field "data.warnings" has the wrong type`},
		{"round 0", started("0", `"spec.md"`, q), `field "data.round" has a wrong value`},
		{"commit", started("1", `"spec.md"`, `"abc"`), `field "data.commit" has a wrong value`},
		{"doc", started("1", `"plan.md"`, q), `field "data.doc" has a wrong value`},
		{"round missing", ln("review.started", `{"doc":"spec.md","reviewer":"r","file":"f","commit":`+q+`}`), `missing field "data.round"`},
		{"finished result", ln("review.finished", `{"doc":"spec.md","round":1,"result":"ok"}`), `field "data.result" has a wrong value`},
		{"finished round missing", ln("review.finished", `{"doc":"spec.md","result":"passed"}`), `missing field "data.round"`},
		{"finished extra", ln("review.finished", `{"doc":"spec.md","round":1,"result":"passed","x":1}`), `unknown field "data.x"`},
		{"top unknown", `{"version":1,"seq":2,"event":"slice.titled","at":"2026-10-24T07:50:00Z","by":"x","data":{"title":"a"}}`, `unknown field "by"`},
		{"at", `{"version":1,"seq":2,"event":"slice.titled","at":"yesterday","data":{"title":"a"}}`, `field "at" has a wrong value`},
	}
	for _, r := range rows {
		events, warns := Parse([]byte(r.bad + "\n"))
		want := ".events.jsonl line 1: " + r.reason + "; skipped"
		if len(events) != 0 || len(warns) != 1 || warns[0] != want {
			t.Errorf("%s: events %v warns %q, want %q", r.name, events, warns, want)
		}
	}
}
