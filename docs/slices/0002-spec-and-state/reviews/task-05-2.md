---
review: task-05-2
slice: 0002-spec-and-state
kind: task
target: task 05
round: 2
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

The task covers R16, R17, and R18, and both points from round 1 are
fixed. An empty Open questions section now gets both of R16's warnings,
`is not None.` and then `is empty; ...`, as the lead decided. The new
sort test fails when the sort compares only lines. No new findings.

## Checked

- Commits: ebb1ee0, adfc0fd (the diff since round 1)
- Files read: internal/spec/warnings.go, internal/spec/check.go,
  internal/spec/warnings_test.go, the brief, reviews/task-05-1.md,
  status.md (Lead's decisions, task-05-1/Q1)
- Ran:
  go test ./internal/spec/ (pass)
  a copy of the worktree in the scratchpad, with `sortProblems` changed
  to compare only lines (the level compare made `return false`): go test
  ./internal/spec/ fails, only `TestSortProblemsLevelAtOneLine`. Deleted
  after.
- Requirements: R16, R17, R18

## Tests

- `TestWarningRows`: ok
- `TestNoteNoScenario`: ok
- `TestWarnings0001`: ok
- `TestCheckOrder`: ok

## Previous findings

- task-05-1/F1: verified (fixed)

## Findings

None.

## Requirements check

None.

## Needs human decision

None.
