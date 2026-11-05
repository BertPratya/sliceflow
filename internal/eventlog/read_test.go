package eventlog_test

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/BertPratya/sliceflow/internal/eventlog"
)

// writeLog writes content to a new .events.jsonl in a temporary folder,
// and returns its path.
func writeLog(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), ".events.jsonl")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// createdData is the data of a good slice.created event.
const createdData = `{"name":"login","title":"Log in","description":"","base":"main","base_commit":"3f9a1c2e8b7d4a6f0c1e5d9b2a7f3c8e4d6b1a09"}`

// createdLine is a good slice.created line, without its line ending.
const createdLine = `{"version":1,"seq":1,"event":"slice.created","at":"2026-11-02T09:15:30Z","data":` + createdData + `}`

// eventLine is a good line of any other event, with empty data.
func eventLine(seq int, event string) string {
	return fmt.Sprintf(`{"version":1,"seq":%d,"event":%q,"at":"2026-11-02T10:00:00Z","data":{}}`, seq, event)
}

// readLog writes content to a log and reads it with Read. It fails the
// test if Read returns an error.
func readLog(t *testing.T, content string) ([]eventlog.Event, []eventlog.BadLine) {
	t.Helper()
	events, bad, err := eventlog.Read(writeLog(t, content))
	if err != nil {
		t.Fatalf("Read returned the error %v, want no error", err)
	}
	return events, bad
}

// badLineNumbers returns the line numbers of bad, for messages and
// comparisons.
func badLineNumbers(bad []eventlog.BadLine) []int {
	nums := make([]int, len(bad))
	for i, b := range bad {
		nums[i] = b.Line
	}
	return nums
}

// Test 1: one good slice.created line is read, with every field.
func TestReadOneCreatedLine(t *testing.T) {
	events, bad := readLog(t, createdLine+"\n")

	if len(bad) != 0 {
		t.Errorf("bad lines: %v, want none", badLineNumbers(bad))
	}
	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}
	e := events[0]
	if e.Version != 1 || e.Seq != 1 || e.Event != "slice.created" || e.At != "2026-11-02T09:15:30Z" {
		t.Errorf("got version %d, seq %d, event %q, at %q; want 1, 1, \"slice.created\", \"2026-11-02T09:15:30Z\"",
			e.Version, e.Seq, e.Event, e.At)
	}
	if string(e.Data) != createdData {
		t.Errorf("data is %s, want %s", e.Data, createdData)
	}
}

// Test 2: three good lines come back as three events, in file order.
func TestReadThreeLinesInOrder(t *testing.T) {
	content := createdLine + "\n" + eventLine(2, "title.changed") + "\n" + eventLine(3, "spec.approved") + "\n"
	events, bad := readLog(t, content)

	if len(bad) != 0 {
		t.Errorf("bad lines: %v, want none", badLineNumbers(bad))
	}
	var seqs []int
	for _, e := range events {
		seqs = append(seqs, e.Seq)
	}
	if !slices.Equal(seqs, []int{1, 2, 3}) {
		t.Errorf("got seqs %v, want [1 2 3]", seqs)
	}
}

// Test 3: a bad line in the middle is skipped and reported, and the
// lines after it are still read.
func TestReadSkipsABadLineInTheMiddle(t *testing.T) {
	content := createdLine + "\n" + "hello" + "\n" + eventLine(3, "x") + "\n"
	events, bad := readLog(t, content)

	if len(events) != 2 {
		t.Errorf("got %d events, want 2", len(events))
	}
	if got := badLineNumbers(bad); !slices.Equal(got, []int{2}) {
		t.Errorf("bad lines: %v, want [2]", got)
	}
}

// Test 4: a cut-off last line, with no "\n", is a line, and a bad one.
func TestReadCutOffLastLine(t *testing.T) {
	content := createdLine + "\n" + `{"version":1,"se`
	events, bad := readLog(t, content)

	if len(events) != 1 {
		t.Errorf("got %d events, want 1", len(events))
	}
	if got := badLineNumbers(bad); !slices.Equal(got, []int{2}) {
		t.Errorf("bad lines: %v, want [2]", got)
	}
}

// Test 5: lines ending in "\r\n" are read the same as lines ending in
// "\n" (R33).
func TestReadWindowsLineEndings(t *testing.T) {
	content := createdLine + "\r\n" + eventLine(2, "x") + "\r\n"
	events, bad := readLog(t, content)

	if len(bad) != 0 {
		t.Errorf("bad lines: %v, want none", badLineNumbers(bad))
	}
	if len(events) != 2 {
		t.Fatalf("got %d events, want 2", len(events))
	}
	if string(events[0].Data) != createdData {
		t.Errorf("data is %s, want %s", events[0].Data, createdData)
	}
}

// Test 6: an empty line in the middle counts as a line, and a bad one.
func TestReadEmptyLineInTheMiddle(t *testing.T) {
	content := createdLine + "\n" + "\n" + eventLine(3, "x") + "\n"
	events, bad := readLog(t, content)

	if len(events) != 2 {
		t.Errorf("got %d events, want 2", len(events))
	}
	if got := badLineNumbers(bad); !slices.Equal(got, []int{2}) {
		t.Errorf("bad lines: %v, want [2]", got)
	}
}

// Test 7: an empty file has no events and no bad lines, and is not an
// error.
func TestReadEmptyFile(t *testing.T) {
	events, bad := readLog(t, "")

	if len(events) != 0 || len(bad) != 0 {
		t.Errorf("got %d events and bad lines %v, want none of either", len(events), badLineNumbers(bad))
	}
}

// Test 8: a missing file is an error.
func TestReadMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".events.jsonl")
	if _, _, err := eventlog.Read(path); err == nil {
		t.Error("Read of a missing file returned no error")
	}
}

// Test 9: an event name sliceflow does not know is still readable.
func TestReadUnknownEvent(t *testing.T) {
	events, bad := readLog(t, eventLine(1, "spec.approved")+"\n")

	if len(bad) != 0 {
		t.Errorf("bad lines: %v, want none", badLineNumbers(bad))
	}
	if len(events) != 1 || events[0].Event != "spec.approved" {
		t.Errorf("got %v, want one spec.approved event", events)
	}
}

// longCreatedLine is a good slice.created line whose description is n
// bytes long.
func longCreatedLine(n int) string {
	return `{"version":1,"seq":1,"event":"slice.created","at":"2026-11-02T09:15:30Z","data":` +
		`{"name":"login","title":"Log in","description":"` + strings.Repeat("a", n) +
		`","base":"main","base_commit":"3f9a1c2e8b7d4a6f0c1e5d9b2a7f3c8e4d6b1a09"}}`
}

// Test 10: a line of 1 MiB, under the 4 MiB limit, is read normally.
func TestReadLongLineUnderTheLimit(t *testing.T) {
	events, bad := readLog(t, longCreatedLine(1<<20)+"\n"+eventLine(2, "x")+"\n")

	if len(bad) != 0 {
		t.Errorf("bad lines: %v, want none", badLineNumbers(bad))
	}
	if len(events) != 2 {
		t.Errorf("got %d events, want 2", len(events))
	}
}

// Test 11: a line over 4 MiB stops Read with an error, because the lines
// after it cannot be read.
func TestReadLineOverTheLimit(t *testing.T) {
	path := writeLog(t, longCreatedLine(5<<20)+"\n"+eventLine(2, "x")+"\n")
	if _, _, err := eventlog.Read(path); err == nil {
		t.Error("Read of a line over 4 MiB returned no error")
	}
}

// Test 12: each rule of R32, broken alone, makes the line bad. The line
// after it is good, so the test also checks that Read goes on. When the
// reason is one of the design's own, the test checks its text. When the
// line is rejected by json.Unmarshal, the reason is Unmarshal's error, so
// the test only checks that there is one.
func TestReadEachRuleBroken(t *testing.T) {
	tests := []struct {
		name   string
		line   string
		reason string // "" means: any reason, from json.Unmarshal
	}{
		{"not JSON", `hello`, ""},
		{"null", `null`, "version is missing"},
		{"no version", `{"seq":1,"event":"x","at":"2026-11-02T09:15:30Z","data":{}}`, "version is missing"},
		{"version 2", `{"version":2,"seq":1,"event":"x","at":"2026-11-02T09:15:30Z","data":{}}`, "version is not 1"},
		{"no event", `{"version":1,"seq":1,"at":"2026-11-02T09:15:30Z","data":{}}`, "event is missing"},
		{"event as a number", `{"version":1,"seq":1,"event":5,"at":"2026-11-02T09:15:30Z","data":{}}`, ""},
		{"no seq", `{"version":1,"event":"x","at":"2026-11-02T09:15:30Z","data":{}}`, "seq is missing"},
		{"seq as text", `{"version":1,"seq":"3","event":"x","at":"2026-11-02T09:15:30Z","data":{}}`, ""},
		{"seq as 2.5", `{"version":1,"seq":2.5,"event":"x","at":"2026-11-02T09:15:30Z","data":{}}`, ""},
		{"no at", `{"version":1,"seq":1,"event":"x","data":{}}`, "at is not a valid time"},
		{"bad at", `{"version":1,"seq":1,"event":"x","at":"yesterday","data":{}}`, "at is not a valid time"},
		{"at not in UTC", `{"version":1,"seq":1,"event":"x","at":"2026-11-02T16:15:30+07:00","data":{}}`, "at is not a valid time"},
		{"at with fractions", `{"version":1,"seq":1,"event":"x","at":"2026-11-02T09:15:30.123Z","data":{}}`, "at is not a valid time"},
		{"no data", `{"version":1,"seq":1,"event":"x","at":"2026-11-02T09:15:30Z"}`, "data is not an object"},
		{"data as a list", `{"version":1,"seq":1,"event":"x","at":"2026-11-02T09:15:30Z","data":[]}`, "data is not an object"},
		{"data as null", `{"version":1,"seq":1,"event":"x","at":"2026-11-02T09:15:30Z","data":null}`, "data is not an object"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events, bad := readLog(t, tt.line+"\n"+eventLine(2, "x")+"\n")

			if len(events) != 1 {
				t.Errorf("got %d events, want 1 (only line 2)", len(events))
			}
			if len(bad) != 1 || bad[0].Line != 1 {
				t.Fatalf("bad lines: %v, want [1]", badLineNumbers(bad))
			}
			if bad[0].Err == nil {
				t.Fatal("the bad line has no reason")
			}
			if tt.reason != "" && !strings.Contains(bad[0].Err.Error(), tt.reason) {
				t.Errorf("reason is %q, want it to say %q", bad[0].Err, tt.reason)
			}
		})
	}
}

// Test 13: a slice.created line is bad when its data misses any one of
// its five fields, and the reason names the field.
func TestReadCreatedMissingAField(t *testing.T) {
	fields := []string{"name", "title", "description", "base", "base_commit"}
	values := map[string]string{
		"name":        `"login"`,
		"title":       `"Log in"`,
		"description": `""`,
		"base":        `"main"`,
		"base_commit": `"3f9a1c2e8b7d4a6f0c1e5d9b2a7f3c8e4d6b1a09"`,
	}
	for _, missing := range fields {
		t.Run(missing, func(t *testing.T) {
			var parts []string
			for _, f := range fields {
				if f != missing {
					parts = append(parts, fmt.Sprintf("%q:%s", f, values[f]))
				}
			}
			line := `{"version":1,"seq":1,"event":"slice.created","at":"2026-11-02T09:15:30Z","data":{` +
				strings.Join(parts, ",") + `}}`

			events, bad := readLog(t, line+"\n")

			if len(events) != 0 {
				t.Errorf("got %d events, want 0", len(events))
			}
			if len(bad) != 1 || bad[0].Err == nil {
				t.Fatalf("bad lines: %v, want [1] with a reason", badLineNumbers(bad))
			}
			// HasSuffix, not Contains: "missing base" is also inside
			// "missing base_commit", so Contains could pass for the
			// wrong field.
			want := "slice.created data is missing " + missing
			if !strings.HasSuffix(bad[0].Err.Error(), want) {
				t.Errorf("reason is %q, want it to end with %q", bad[0].Err, want)
			}
		})
	}
}
