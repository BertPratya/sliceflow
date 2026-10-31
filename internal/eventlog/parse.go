package eventlog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

// NotFound is the warning for a slice whose log file does not exist (R26).
// The caller adds the slice name in front of it.
const NotFound = ".events.jsonl not found"

// Log is what Parse read from a log: the good events in file order, and
// the warnings, in the order they were found. A warning has no slice name;
// the caller adds "Warning: <name>: " in front of it.
type Log struct {
	Events   []Event
	Warnings []string
}

var (
	// integerRe matches a JSON integer as written.
	integerRe = regexp.MustCompile(`^-?(0|[1-9][0-9]*)$`)
	// atRe is the shape of the time in R19. A match is not enough: the
	// time must also be a real one.
	atRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`)
	// commitRe is a full commit id: 40 or 64 lowercase hex characters.
	commitRe = regexp.MustCompile(`^([0-9a-f]{40}|[0-9a-f]{64})$`)
)

// member is one key of a JSON object, with its value as written.
type member struct {
	key string
	raw json.RawMessage
}

// Parse reads the content of a log (R19, R26, R27). Lines are split on \n.
// The empty piece after the last \n is not a line. Line numbers start at 1.
// A bad line is skipped and gets a warning. A line with the wrong seq is
// kept and gets a warning.
//
// Assumed: a \r at the end of a line is removed before the line is checked.
// Assumed: valid JSON that is not an object counts as "not valid JSON".
// A leading UTF-8 BOM makes line 1 "not valid JSON". A seq of 0 or less is
// kept and gets the R27 warning; the spec lists no other reason for it.
//
// A line with several faults gets one reason only. The order of the checks
// is ours; the spec does not fix it. It is: blank line; not valid JSON;
// a field that appears twice; version; event; the other top-level fields
// (unknown, missing, wrong type, wrong value); then the fields of data
// and of data.from the same way. Within a step, fields are checked in
// file order for unknown ones, and in the order of R19 for the others.
func Parse(content []byte) Log {
	var l Log
	pieces := strings.Split(string(content), "\n")
	if pieces[len(pieces)-1] == "" {
		pieces = pieces[:len(pieces)-1]
	}
	last := 0 // seq of the last good line
	for i, piece := range pieces {
		n := i + 1
		piece = strings.TrimSuffix(piece, "\r")
		e, reason := parseLine(piece)
		if reason != "" {
			l.Warnings = append(l.Warnings,
				fmt.Sprintf(".events.jsonl line %d: %s; skipped", n, reason))
			continue
		}
		if e.Seq != last+1 {
			l.Warnings = append(l.Warnings,
				fmt.Sprintf(".events.jsonl line %d: seq %d, expected %d", n, e.Seq, last+1))
		}
		last = e.Seq
		e.Line = n
		l.Events = append(l.Events, e)
	}
	return l
}

// Created returns the first good slice.created event and its data. ok is
// false when there is none. Assumed: if there are more, the first counts.
func (l Log) Created() (e Event, data SliceCreated, ok bool) {
	for _, ev := range l.Events {
		if ev.Event != EventSliceCreated {
			continue
		}
		if err := json.Unmarshal(ev.Data, &data); err != nil {
			continue
		}
		return ev, data, true
	}
	return Event{}, SliceCreated{}, false
}

// parseLine checks one line. It returns the event, or a reason (the text
// of R26) when the line is bad.
func parseLine(text string) (Event, string) {
	if strings.TrimSpace(text) == "" {
		return Event{}, "blank line"
	}
	raw := []byte(text)
	if !json.Valid(raw) || firstByte(raw) != '{' {
		return Event{}, "not valid JSON"
	}
	if name := firstDuplicate(raw, ""); name != "" {
		return Event{}, fmt.Sprintf("field %q appears twice", name)
	}
	top, _ := members(raw)
	get := func(key string) (json.RawMessage, bool) { return find(top, key) }

	// version
	rawVersion, ok := get("version")
	if !ok {
		return Event{}, missing("version")
	}
	if !integerRe.Match(rawVersion) {
		return Event{}, wrongType("version")
	}
	if string(rawVersion) != strconv.Itoa(Version) {
		return Event{}, fmt.Sprintf("unknown version %s", rawVersion)
	}

	// event
	rawEvent, ok := get("event")
	if !ok {
		return Event{}, missing("event")
	}
	name, ok := asString(rawEvent)
	if !ok {
		return Event{}, wrongType("event")
	}
	checkData, ok := dataChecks[name]
	if !ok {
		return Event{}, fmt.Sprintf("unknown event %q", name)
	}

	// the other top-level fields
	if r := unknownFields(top, "", "version", "seq", "event", "at", "data"); r != "" {
		return Event{}, r
	}
	rawSeq, okSeq := get("seq")
	rawAt, okAt := get("at")
	rawData, okData := get("data")
	for _, f := range []struct {
		name string
		ok   bool
	}{{"seq", okSeq}, {"at", okAt}, {"data", okData}} {
		if !f.ok {
			return Event{}, missing(f.name)
		}
	}
	if !integerRe.Match(rawSeq) {
		return Event{}, wrongType("seq")
	}
	atText, ok := asString(rawAt)
	if !ok {
		return Event{}, wrongType("at")
	}
	if firstByte(rawData) != '{' {
		return Event{}, wrongType("data")
	}
	seq, err := strconv.Atoi(string(rawSeq))
	if err != nil { // an integer too big for int
		return Event{}, wrongValue("seq")
	}
	at, err := time.Parse(atLayout, atText)
	if !atRe.MatchString(atText) || err != nil {
		return Event{}, wrongValue("at")
	}

	// the event's own data
	if r := checkData(rawData); r != "" {
		return Event{}, r
	}
	return Event{
		Version: Version,
		Seq:     seq,
		Event:   name,
		At:      at.UTC(),
		Data:    rawData,
	}, ""
}

// dataChecks maps each known event name to the check for its data. A
// check returns a reason, or "" when the data is good. A new event needs
// one line here.
var dataChecks = map[string]func(json.RawMessage) string{
	EventSliceCreated: checkCreatedData,
	EventSliceTitled:  checkTitledData,
}

// checkCreatedData checks the data of a slice.created event (R20). It
// returns a reason, or "" when the data is good.
func checkCreatedData(rawData json.RawMessage) string {
	data, _ := members(rawData)
	if r := unknownFields(data, "data.", "from", "title"); r != "" {
		return r
	}
	rawFrom, ok := find(data, "from")
	if !ok {
		return missing("data.from")
	}
	rawTitle, hasTitle := find(data, "title")
	if firstByte(rawFrom) != '{' {
		return wrongType("data.from")
	}
	var title string
	if hasTitle {
		if title, ok = asString(rawTitle); !ok {
			return wrongType("data.title")
		}
	}

	from, _ := members(rawFrom)
	if r := unknownFields(from, "data.from.", "branch", "commit"); r != "" {
		return r
	}
	rawBranch, okBranch := find(from, "branch")
	rawCommit, okCommit := find(from, "commit")
	if !okBranch {
		return missing("data.from.branch")
	}
	if !okCommit {
		return missing("data.from.commit")
	}
	branch, ok := asString(rawBranch)
	if !ok {
		return wrongType("data.from.branch")
	}
	commit, ok := asString(rawCommit)
	if !ok {
		return wrongType("data.from.commit")
	}

	if branch == "" {
		return wrongValue("data.from.branch")
	}
	if !commitRe.MatchString(commit) {
		return wrongValue("data.from.commit")
	}
	if hasTitle && title == "" {
		return wrongValue("data.title")
	}
	return ""
}

// checkTitledData checks the data of a slice.titled event (R29): only
// title, a string that is not empty.
func checkTitledData(rawData json.RawMessage) string {
	data, _ := members(rawData)
	if r := unknownFields(data, "data.", "title"); r != "" {
		return r
	}
	rawTitle, ok := find(data, "title")
	if !ok {
		return missing("data.title")
	}
	title, ok := asString(rawTitle)
	if !ok {
		return wrongType("data.title")
	}
	if title == "" {
		return wrongValue("data.title")
	}
	return ""
}

// missing returns the reason for a field that is not there.
func missing(name string) string { return fmt.Sprintf("missing field %q", name) }

// wrongType returns the reason for a field of the wrong JSON type.
func wrongType(name string) string { return fmt.Sprintf("field %q has the wrong type", name) }

// wrongValue returns the reason for a field with a value that is not allowed.
func wrongValue(name string) string { return fmt.Sprintf("field %q has a wrong value", name) }

// unknownFields returns the reason for the first member (in file order)
// whose key is not in allowed, or "" when all are known. prefix is put in
// front of the key, e.g. "data.from.".
func unknownFields(ms []member, prefix string, allowed ...string) string {
	for _, m := range ms {
		if !slices.Contains(allowed, m.key) {
			return fmt.Sprintf("unknown field %q", prefix+m.key)
		}
	}
	return ""
}

// find returns the value of a key. The keys are unique here, because
// firstDuplicate ran before.
func find(ms []member, key string) (json.RawMessage, bool) {
	for _, m := range ms {
		if m.key == key {
			return m.raw, true
		}
	}
	return nil, false
}

// asString reads a JSON string. ok is false when raw is not a string.
func asString(raw json.RawMessage) (string, bool) {
	if firstByte(raw) != '"' {
		return "", false
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return "", false
	}
	return s, true
}

// firstByte returns the first byte of raw that is not white space, or 0.
func firstByte(raw []byte) byte {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return 0
	}
	return raw[0]
}

// members reads one JSON object into its keys and values, in file order.
// Repeated keys are all kept. ok is false when raw is not an object.
// encoding/json would keep only the last of a repeated key, so we read
// tokens ourselves.
func members(raw []byte) (ms []member, ok bool) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	if t, err := dec.Token(); err != nil || t != json.Delim('{') {
		return nil, false
	}
	for dec.More() {
		t, err := dec.Token()
		if err != nil {
			return ms, false
		}
		key, isString := t.(string)
		if !isString {
			return ms, false
		}
		var v json.RawMessage
		if err := dec.Decode(&v); err != nil {
			return ms, false
		}
		ms = append(ms, member{key: key, raw: v})
	}
	return ms, true
}

// firstDuplicate looks for a key that appears twice in the same object. It
// looks in the line, in data, and in data.from only, the objects whose
// fields we know, so the work stays linear however deep the JSON is. It
// returns the full name of the first one found, such as "seq" or
// "data.from.branch", or "" if there is none. prefix is the name of the
// object that raw is the value of, with a dot at the end.
func firstDuplicate(raw []byte, prefix string) string {
	ms, ok := members(raw)
	if !ok {
		return ""
	}
	seen := map[string]bool{}
	for _, m := range ms {
		if seen[m.key] {
			return prefix + m.key
		}
		seen[m.key] = true
		full := prefix + m.key
		if (full == "data" || full == "data.from") && firstByte(m.raw) == '{' {
			if name := firstDuplicate(m.raw, full+"."); name != "" {
				return name
			}
		}
	}
	return ""
}
