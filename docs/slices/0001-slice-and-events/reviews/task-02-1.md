---
review: task-02-1
slice: 0001-slice-and-events
kind: task
target: task 02
round: 1
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

The writer and reader work as R19, R25, R26, and R27 say, with the exact
texts, and the tests are real: mutations I tried made them fail. One
`ASSUMED` choice goes beyond R27: the first good line must have `seq` 1, or
it gets a seq warning (Q1). One low edge case: a `seq` too big for an `int`
gets a reason R26 does not list (F1).

## Checked

- Commits: be69ba0
- Files read: internal/eventlog/event.go, internal/eventlog/parse.go,
  internal/eventlog/event_test.go, internal/eventlog/parse_test.go, the
  brief, the commit message (`ASSUMED` notes), status.md (no deferred
  findings for this task)
- Ran:
  go test -count=1 ./internal/eventlog/ (pass)
  go vet ./internal/eventlog/ (pass); gofmt -l internal (prints nothing)
  probe of edge cases on a scratch copy of the package (huge seq, seq 0 and -1, version -0, BOM, empty title, nested duplicate in an array, fractional seconds in at, a non-object line): results as expected, except F1 and Q1
  mutations on a scratch copy (no UTC in NewCreated, seq not kept after a gap, duplicate path without its prefix, no commit check, Phase always unknown): each made a task test fail
- Requirements: R19, R25, R26, R27
- Interfaces: Event, Created, From (fields and JSON names), NewCreated,
  Encode, Parse, Phase match the task; after Parse, Data is a Created

## Tests

- `TestEncodeCreated`: ok
- `TestEncodeLocalTimeIsUTC`: ok
- `TestParseGoodLog`: ok
- `TestParseBadLines`: ok
- `TestParseFirstFaultWins`: ok
- `TestParseSeqGap`: ok
- `TestPhaseUnknown`: ok

## Previous findings

None.

## Findings

### F1. A seq too big for an int gets a wrong value reason R26 does not list

Level: low
Where: internal/eventlog/parse.go:266-269
Breaks: nothing
Why: a line with `"seq":99999999999999999999` passes the whole-number
  check, then `strconv.Atoi` fails, and the line gets
  `field "seq" has a wrong value; skipped`. R26's list of wrong values
  names only `at`, `commit`, `branch`, and `title`. I found it with a probe
  on a scratch copy. No real log has such a number.
Fix: leave as is, or report it as `has the wrong type`, the reason R26
  already uses for a `seq` that is not a usable whole number.

## Requirements check

None.

## Needs human decision

- **Q1.** R27 compares a line's `seq` with the last good line's `seq`.
  It does not say what happens when there is no last good line yet. The
  worker chose (`ASSUMED`) to expect `seq` 1 then, so a log whose first
  good line has `seq` 3, or whose line 1 is bad and line 2 has `seq` 2,
  gets `.events.jsonl line <n>: seq <found>, expected 1`. Read literally,
  R27 gives no seq warning for the first good line. R19 does say the first
  event has `seq` 1, which supports the worker's choice.
  Guess: keep the worker's choice (expect 1 when there is no last good
  line), or give no seq warning for the first good line.
