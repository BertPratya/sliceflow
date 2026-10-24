---
review: task-04-3
slice: 0002-spec-and-state
kind: task
target: task 04
round: 3
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

The fix for task-04-2/F1 works: a requirement above the first section now
counts for `Shows:` and `Rule:`, as one in any other wrong section does.
The new test fails when the fix is undone. No new findings.

## Checked

- Commits: c25ef94, e164aa8, 7810cce (round 3 diff: e164aa8..7810cce)
- Files read: internal/spec/errors.go (itemErrors, requirementErrors,
  scenarioErrors), internal/spec/errors_test.go
  (TestErrorShowsRequirementBeforeFirstSection), the brief, reviews
  task-04-1 and task-04-2, the commit messages of e164aa8 and 7810cce
  with their ASSUMED lines
- Ran:
  go test ./internal/spec/ -count=1 (pass)
  go vet ./internal/spec/ (pass)
  go test ./... -count=1 (pass)
  on a copy in my scratchpad, errors.go with the requirements map built from `d.Sections` only again: TestErrorShowsRequirementBeforeFirstSection fails (got the extra `S1 shows R9, which does not exist`)
- Requirements: R15 (the `Shows:` and `Rule:` rows, and the marker outside its section)
- Interfaces: `Problem`, `Check(d Doc, slice string) []Problem`, and `Doc` unchanged since round 2

## Tests

- `TestErrorRows`: ok
- `TestErrorPlaceholderValue`: ok
- `TestErrorsFresh`: ok
- `TestErrors0001`: ok
- `TestErrorMarkerBeforeFirstSection`: ok

## Previous findings

- task-04-2/F1: verified (fixed)

## Findings

None.

## Requirements check

None.

## Needs human decision

None.
