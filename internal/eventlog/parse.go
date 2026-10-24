package eventlog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	wholeNumber = regexp.MustCompile(`^-?(0|[1-9][0-9]*)$`)
	commitRe    = regexp.MustCompile(`^([0-9a-f]{40}|[0-9a-f]{64})$`)
)

// Parse reads a whole log. It returns the good events and the warnings.
func Parse(log []byte) ([]Event, []string) {
	var events []Event
	var warnings []string
	if len(log) == 0 {
		return nil, nil
	}
	lines := strings.Split(string(log), "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	last := 0
	for i, line := range lines {
		n := i + 1
		ev, reason := parseLine(line)
		if reason != "" {
			warnings = append(warnings, fmt.Sprintf(".events.jsonl line %d: %s; skipped", n, reason))
			continue
		}
		if ev.Seq != last+1 {
			warnings = append(warnings, fmt.Sprintf(".events.jsonl line %d: seq %d, expected %d", n, ev.Seq, last+1))
		}
		last = ev.Seq
		events = append(events, ev)
	}
	return events, warnings
}

// Phase returns "planning" when there is a good slice.created event,
// else "unknown".
func Phase(events []Event) string {
	for _, e := range events {
		if e.Event == "slice.created" {
			return "planning"
		}
	}
	return "unknown"
}

// findDuplicate walks the tokens and returns the path of the first name
// that appears twice in one object, or "".
func findDuplicate(dec *json.Decoder, prefix string) (string, error) {
	tok, err := dec.Token()
	if err != nil {
		return "", err
	}
	d, ok := tok.(json.Delim)
	if !ok {
		return "", nil
	}
	switch d {
	case '{':
		seen := map[string]bool{}
		for dec.More() {
			kt, err := dec.Token()
			if err != nil {
				return "", err
			}
			key, _ := kt.(string)
			if seen[key] {
				return prefix + key, nil
			}
			seen[key] = true
			if dup, err := findDuplicate(dec, prefix+key+"."); err != nil || dup != "" {
				return dup, err
			}
		}
	case '[':
		for dec.More() {
			if dup, err := findDuplicate(dec, prefix); err != nil || dup != "" {
				return dup, err
			}
		}
	}
	_, err = dec.Token()
	return "", err
}

func isWhole(raw json.RawMessage) bool {
	return wholeNumber.Match(raw)
}

func asString(raw json.RawMessage) (string, bool) {
	if len(raw) == 0 || raw[0] != '"' {
		return "", false
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return "", false
	}
	return s, true
}

func asObject(raw json.RawMessage) (map[string]json.RawMessage, bool) {
	if len(raw) == 0 || raw[0] != '{' {
		return nil, false
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, false
	}
	return m, true
}

func unknownIn(m map[string]json.RawMessage, allowed []string, prefix string) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		known := false
		for _, a := range allowed {
			if a == k {
				known = true
			}
		}
		if !known {
			return fmt.Sprintf("unknown field %q", prefix+k)
		}
	}
	return ""
}

// parseLine returns the event, or a reason when the line is bad.
func parseLine(line string) (Event, string) {
	if strings.Trim(line, " \t\r") == "" {
		return Event{}, "blank line"
	}
	b := []byte(line)
	if !json.Valid(b) || strings.TrimLeft(line, " \t\r")[0] != '{' {
		return Event{}, "not valid JSON"
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	if dup, err := findDuplicate(dec, ""); err != nil {
		return Event{}, "not valid JSON"
	} else if dup != "" {
		return Event{}, fmt.Sprintf("field %q appears twice", dup)
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal(b, &top); err != nil {
		return Event{}, "not valid JSON"
	}

	// unknown version, unknown event
	if raw, ok := top["version"]; ok && isWhole(raw) && string(raw) != "1" {
		return Event{}, "unknown version " + string(raw)
	}
	if raw, ok := top["event"]; ok {
		if s, ok := asString(raw); ok && s != "slice.created" {
			return Event{}, fmt.Sprintf("unknown event %q", s)
		}
	}

	// unknown fields
	if r := unknownIn(top, []string{"version", "seq", "event", "at", "data"}, ""); r != "" {
		return Event{}, r
	}
	data, dataOK := asObject(top["data"])
	var from map[string]json.RawMessage
	fromOK := false
	if dataOK {
		if r := unknownIn(data, []string{"from", "title"}, "data."); r != "" {
			return Event{}, r
		}
		from, fromOK = asObject(data["from"])
		if fromOK {
			if r := unknownIn(from, []string{"branch", "commit"}, "data.from."); r != "" {
				return Event{}, r
			}
		}
	}

	// missing fields
	for _, f := range []string{"version", "seq", "event", "at", "data"} {
		if _, ok := top[f]; !ok {
			return Event{}, fmt.Sprintf("missing field %q", f)
		}
	}
	if dataOK {
		if _, ok := data["from"]; !ok {
			return Event{}, `missing field "data.from"`
		}
		if fromOK {
			for _, f := range []string{"branch", "commit"} {
				if _, ok := from[f]; !ok {
					return Event{}, fmt.Sprintf("missing field %q", "data.from."+f)
				}
			}
		}
	}

	// wrong types
	wrong := func(f string) (Event, string) {
		return Event{}, fmt.Sprintf("field %q has the wrong type", f)
	}
	if !isWhole(top["version"]) {
		return wrong("version")
	}
	seq, err := strconv.Atoi(string(top["seq"]))
	if !isWhole(top["seq"]) || err != nil {
		return wrong("seq")
	}
	if _, ok := asString(top["event"]); !ok {
		return wrong("event")
	}
	at, ok := asString(top["at"])
	if !ok {
		return wrong("at")
	}
	if !dataOK {
		return wrong("data")
	}
	if !fromOK {
		return wrong("data.from")
	}
	branch, ok := asString(from["branch"])
	if !ok {
		return wrong("data.from.branch")
	}
	commit, ok := asString(from["commit"])
	if !ok {
		return wrong("data.from.commit")
	}
	title := ""
	if raw, has := data["title"]; has {
		if title, ok = asString(raw); !ok {
			return wrong("data.title")
		}
	}

	// wrong values
	bad := func(f string) (Event, string) {
		return Event{}, fmt.Sprintf("field %q has a wrong value", f)
	}
	if t, err := time.Parse(TimeLayout, at); err != nil || t.Format(TimeLayout) != at {
		return bad("at")
	}
	if !commitRe.MatchString(commit) {
		return bad("data.from.commit")
	}
	if branch == "" {
		return bad("data.from.branch")
	}
	if _, has := data["title"]; has && title == "" {
		return bad("data.title")
	}

	return Event{
		Version: 1,
		Seq:     seq,
		Event:   "slice.created",
		At:      at,
		Data:    Created{From: From{Branch: branch, Commit: commit}, Title: title},
	}, ""
}
