---
review: spec-3
slice: 0001-slice-and-events
kind: spec
target: docs/slices/0001-slice-and-events/spec.md
round: 3
reviewer: doc-reviewer (claude opus)
---

# Review

## Summary

All five fixes hold, and no gap needs the human. The F2 fix of spec-2
left one point open: whether the base-branch warning keeps its
`Warning: ` prefix as an item of `warnings` in `--json` (F1). Two small
points: the fixed empty-list output of R37, and the loose "like" in R26.

## Checked

- Files read: docs/slices/0001-slice-and-events/spec.md,
  docs/slices/0001-slice-and-events/status.md,
  docs/slices/0001-slice-and-events/reviews/spec-1.md,
  docs/slices/0001-slice-and-events/reviews/spec-2.md,
  docs/templates/spec.md, docs/templates/review.md,
  .agents/prompts/doc-reviewer.md, AGENTS.md
- Ran:
  git show ec7a45f, to read the fix diff of spec.md
  grep for leftover placeholders and comments in spec.md (none found)
  counted the S20 column widths against R34 (all match), and the length
  of the S1 commit id (40)
- Requirements: R1-R41, S1-S28, SD1-SD13; closely R9, R10, R23, R26-R30,
  R34-R37, S16, S20-S25

## Tests

None.

## Previous findings

- spec-1/F6: verified (fixed)
- spec-1/F10: verified (fixed)
- spec-2/F1: verified (fixed)
- spec-2/F2: verified (fixed)
- spec-2/F3: verified (fixed)

## Findings

### F1. The form of the base-branch warning as an item of warnings is not given

Level: medium
Where: spec.md → Requirements → R23, R30, R37; Story 3 → S25
Breaks: nothing
Why: R30 says an item of `warnings` is the text "without
  `Warning: <name>: `". The warning of R23 has no `<name>`; its text, as
  R23 and S25 write it, starts with `Warning: `. Input: S25's repo, then
  `sliceflow slice list --json`. One worker writes
  `"warnings":["Warning: base branch main does not exist (the default); merged slices are not shown"]`,
  another drops the `Warning: ` and writes `"base branch main does not
  exist ..."`. The same goes for `slice status --json` in R30.
Fix: say every item is the text without its `Warning: ` prefix (and
  `<name>: ` when there is one), and give the R23 item as an example.

### F2. R37's output for no slices ignores the base-branch warning

Level: low
Where: spec.md → Requirements → R37
Breaks: nothing
Why: R37 says "With no slices, it prints `{"slices":[],"warnings":[]}`".
  In a repo whose base branch does not exist and with no slices, R37's
  own `warnings` rule puts the R23 warning in the list, so the output is
  not that fixed text. The intent is clear, but the sentence reads as
  one exact output.
Fix: say "With no slices, `slices` is `[]`."

### F3. R26's wrong `at` is given only by an example

Level: low
Where: spec.md → Requirements → R26
Breaks: nothing
Why: "an `at` that is not a UTC time like `2026-10-24T07:42:00Z`" does
  not say if `2026-10-24T07:42:00.5Z` or `2026-10-24T07:42:00+00:00` is
  good. Both are UTC times in RFC 3339, but not the form R19 writes.
Fix: say "an `at` not in the form of R19".

## Requirements check

None.

## Needs human decision

None.
