---
review: task-05-1
slice: 0002-spec-and-state
kind: task
target: task 05
round: 1
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

The task covers R16, R17, and R18's skip rules, and every message matches
R16's table and R17 word for word. One open point: an empty Open questions
section gets one warning, where R16 read literally gives two (Q1). One low
finding: `TestCheckOrder` does not really prove the error-before-warning
order at one line (F1).

## Checked

- Commits: ebb1ee0
- Files read: internal/spec/warnings.go, internal/spec/check.go,
  internal/spec/warnings_test.go, internal/spec/doc.go,
  internal/spec/parse.go, internal/spec/errors.go (helpers),
  internal/spec/errors_test.go (goodSpec, lineOf, TestProblemsSorted),
  internal/spec/testdata/0001-spec.md (R11, R40, R43 lines),
  internal/template/spec.md (Open questions), the brief, spec.md R16 to
  R18 and S6
- Ran:
  go test ./internal/spec/ (pass)
  a throwaway probe test, deleted after: an Open questions section with
  only a comment gives `"## Open questions" is empty; ...` and
  `template comment left`, but no `is not None.` warning; the fresh
  template gives placeholder and comment warnings, `1 open question left
  (Q1)`, and `no scenario shows R1`
- Requirements: R16, R17, R18
- Interfaces: `Check(d Doc, slice string) []Problem` unchanged; `Problem`
  and `Doc` used as they are

## Tests

- `TestWarningRows`: ok
- `TestNoteNoScenario`: ok
- `TestWarnings0001`: ok
- `TestCheckOrder`: weak, see F1

## Previous findings

None.

## Findings

### F1. TestCheckOrder passes even with no level order at one line

Level: low
Where: internal/spec/warnings_test.go:118-134
Breaks: nothing
Why: `Check` appends the errors first, then the warnings, then the notes,
  so problems at one line are already in level order before the sort.
  If `sortProblems` compared only lines, `TestCheckOrder` would still
  pass. `TestProblemsSorted` checks only the line order too. So no test
  proves that the error comes before the warning at one line. Today the
  output is right; a later change to the append order could break it
  without a test failing.
Fix: add a case that calls `sortProblems` on problems at one line in
  the order note, warning, error, and checks the result.

## Requirements check

None.

## Needs human decision

- **Q1.** An Open questions section that holds nothing, or only a
  comment, is empty, and it does not say `None.`. R16's first row and
  last row both apply, so R16 read literally gives two warnings at the
  heading: `"## Open questions" is empty; write None. if there is nothing
  to say` and `"## Open questions" is not None.`. The worker assumed only
  the first one (ASSUMED 2), so the user sees one warning. Is one
  warning right, or should both be shown?
  Guess: keep one warning (the "is empty" one, which also tells the user
  to write None.), or report both, as R16 reads.
