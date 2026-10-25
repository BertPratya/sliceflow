---
review: task-01-1
slice: 0002-spec-and-state
kind: task
target: task 01
round: 1
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

The code reads and writes the four new events as R5 says, with the R26
reasons in the right order, and the names match Interfaces. One medium
problem: the tests leave out the S15 line, so the `accepted` result of
`spec.submitted` is never encoded or parsed in a test (F1).

## Checked

- Commits: ffc681e (diff against ffc681e~1)
- Files read: internal/eventlog/event.go, internal/eventlog/parse.go,
  internal/eventlog/event_test.go, internal/eventlog/parse_test.go,
  docs/slices/0002-spec-and-state/spec.md (S14, S15, S20, S30),
  docs/slices/0001-slice-and-events/spec.md (R26, R27), the brief
- Ran:
  go test ./internal/eventlog/ -run NewEvents -v (pass)
  go test ./... (pass)
- Requirements: R5
- Interfaces: Event.Line with `json:"-"`, Titled, Submitted,
  ReviewStarted, ReviewFinished (JSON names and order), New(seq, at,
  event, data) match the task; Parse, Encode, Phase unchanged

## Tests

- `TestEncodeNewEvents`: weak, see F1
- `TestParseNewEvents`: weak, see F1
- `TestParseNewEventsBad`: ok

## Previous findings

None.

## Findings

### F1. The tests leave out the S15 line

Level: medium
Where: internal/eventlog/event_test.go:36-48, internal/eventlog/parse_test.go:104-124
Breaks: task 01 (Done when: the exact JSON lines of S14, S15, S20, and S30)
Why: `newEventCases` has four lines: S30, S14, and the two lines of S20.
  The S15 line,
  `{"version":1,"seq":3,"event":"spec.submitted","at":"2026-10-24T10:00:00Z","data":{"result":"accepted","errors":0,"warnings":0}}`,
  is missing. So no test encodes or parses `"result":"accepted"`. A typo
  in `oneOf("accepted", "refused")` would still pass every test. The
  ASSUMED note says the lines have no consistent `seq`; with S15 they do:
  S14, S15, S20 give `seq` 2, 3, 4, 5.
Fix: add the S15 line. Then the parse test can use S14, S15, S20 (seq 2
  to 5) and check the one exact warning `.events.jsonl line 1: seq 2,
  expected 1`, with S30's line checked on its own.

### F2. TestParseNewEvents checks only that some warning exists

Level: low
Where: internal/eventlog/parse_test.go:120-123
Breaks: nothing
Why: `len(warns) == 0` is the only check on the warnings. A wrong or
  extra warning, e.g. a seq warning with the wrong line number, would
  still pass.
Fix: compare `warns` with the exact list, as F1's fix says.

## Requirements check

None.

## Needs human decision

None.
