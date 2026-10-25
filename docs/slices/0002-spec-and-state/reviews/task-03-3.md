---
review: task-03-3
slice: 0002-spec-and-state
kind: task
target: task 03
round: 3
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

The fix for task-03-2/F1 works, and its new test proves it: a stray
backtick in one item no longer hides the next item. Inline code over a
line break inside one item still works. No new problems.

## Checked

- Commits: 2b91ce9 (the diff 31a6ab8..2b91ce9 in full); the whole task
  5b38b6f~1..2b91ce9 as read in rounds 1 and 2
- Files read: internal/spec/parse.go (scanLines, closingRun, spansLines,
  startsBlock, fenceOpen), internal/spec/parse_test.go, the brief,
  docs/slices/0002-spec-and-state/reviews/task-03-2.md, status.md
  (Lead's decisions), the commit message of 2b91ce9 and its ASSUMED line
- Ran:
  go test ./internal/spec/ (pass)
  the new test file on the round 2 parse.go from 31a6ab8, in a scratch
  module (TestParseStrayBacktickKeepsNextItem fails, as it should)
  the new parse.go in the scratch module on 3 more inputs: inline code
  over a line break inside one item, followed by a `{{x}}` and another
  item (the code is skipped, `{{x}}` is a placeholder, both items read);
  a stray backtick in the last item of a section, followed by a
  `## Scenarios` heading (both sections and their items read); a stray
  backtick followed by a fenced block and another item (the fence is
  skipped, R2 is read with its label) (all as expected)
- Requirements: R18
- Interfaces: Parse, Doc, Field, Section, Item, Label, and Mark are
  unchanged and still match the brief

## Tests

- `TestParseSections`: ok
- `TestParseSkips`: ok
- `TestParseCRLF`: ok
- `TestParseRemoved`: ok
- `TestParseFrontMatter`: ok

## Previous findings

- task-03-2/F1: verified (fixed)

## Findings

None.

## Requirements check

None.

## Needs human decision

None.
