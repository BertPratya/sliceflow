---
review: task-04-2
slice: 0002-spec-and-state
kind: task
target: task 04
round: 2
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

The fixes for task-04-1 F1 to F3 work, and each new test fails when its
fix is undone. The messages still match R15's table word for word. One
new low finding: an R marker above the first section does not count as a
requirement for `Shows:` and `Rule:`, but one in another wrong section
does (F1).

## Checked

- Commits: c25ef94, e164aa8 (round 2 diff: c25ef94..e164aa8)
- Files read: internal/spec/parse.go, internal/spec/doc.go,
  internal/spec/errors.go (the whole diff, and itemErrors in full),
  internal/spec/errors_test.go (the three new tests),
  internal/spec/parse_test.go (TestParsePreamble),
  internal/spec/testdata/fresh.md (top), the brief, reviews task-04-1,
  the commit message with its ASSUMED lines,
  docs/slices/0002-spec-and-state/spec.md (R15 table, R18)
- Ran:
  go test ./internal/spec/ -count=1 (pass)
  grep for `%q` in internal/spec (only in test failure output; none left in messages)
  on a copy in my scratchpad, parse.go with `Mark.Text` from `clean` again: TestErrorBadMarkerKeepsInlineCode fails (got `"**R2 uses     here"`)
  on a copy, errors.go with the Goes to: message back on `%q`: TestErrorMessagesKeepUserText fails (got `Goes to: "\"x\\y\""`)
  probe test on a copy: Check on 0001's and 0002's spec.md (0 errors each); an unclosed front matter with a commented and an inline-code marker above the first heading (only the expected no front matter and missing section errors); R9 with a Check: line above the first heading (only `R9 is outside "## Requirements"`); a CRLF file with R9 above the first heading (outside, and no Check: line, at line 47)
- Requirements: R15, R18 (placeholder values, and skipping code and comments above the first heading)
- Interfaces: `Doc` with `Preamble Section` in the order the brief gives; `Preamble` has `Name` "" and `Line` 0 and is not in `Sections`; `Problem` and `Check(d Doc, slice string) []Problem` unchanged; task 03's other names unchanged

## Tests

- `TestErrorRows`: ok
- `TestErrorPlaceholderValue`: ok
- `TestErrorsFresh`: ok
- `TestErrors0001`: ok
- `TestErrorMarkerBeforeFirstSection`: ok

## Previous findings

- task-04-1/F1: verified (fixed)
- task-04-1/F2: verified (fixed)
- task-04-1/F3: verified (fixed)
- task-04-1/F4: verified (wontfix)

## Findings

### F1. An R marker above the first section is not a requirement for Shows: and Rule:

Level: low
Where: internal/spec/errors.go:183-192
Breaks: nothing
Why: itemErrors builds the map of requirements from `d.Sections` only,
  but checks the markers of `d.Preamble` too. So `- **R9.** x` in
  `## Summary` counts as a requirement: `Shows: R9` gives only
  `R9 is outside "## Requirements"`. The same R9 above the first `##`
  heading also gives `S1 shows R9, which does not exist`, though R9 is
  in the file. Both are wrong placements, but they get different
  errors. Found by reading the loop; the fix for task-04-1/F1 caused it.
Fix: build the map from the same list the loop below uses:
  `append([]Section{d.Preamble}, d.Sections...)`.

## Requirements check

None.

## Needs human decision

None.
