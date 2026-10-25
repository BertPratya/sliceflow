---
review: task-18-2
slice: 0002-spec-and-state
kind: task
target: task 18
round: 2
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

task-18-1/F1 is fixed. The undo now touches the index only when
`git add` may have run, so a stop before it, e.g. while `spec submit`
checks the spec, leaves a staged `spec.md` as it was. The new test fails
without the fix. F2 is in the backlog. No new findings.

## Checked

- Commits: 7aea506..715ebe3
- Files read: internal/slice/save.go (saveEvent, its undo,
  staleIndexLock, failSave), internal/slice/save_test.go (titleOp, the
  new test), internal/slice/submit.go (the op of `spec submit`),
  internal/git/git.go (`Run`), docs/backlog.md on main (a7f51a1),
  docs/slices/0002-spec-and-state/reviews/task-18-1.md, the brief of
  task 18
- Ran:
  go test ./... -count=1 (pass)
  go test ./internal/slice/ -run TestSaveUndoKeepsStaged -count=1 -v, in a copy of 715ebe3 outside the worktree with the change to save.go reverted (TestSaveUndoKeepsStagedBeforeAdd fails: the staged spec.md is the committed template again; TestSaveUndoKeepsStaged passes)
- Requirements: R8, R10
- Interfaces: `saveEvent` and `saveOp` unchanged

## Tests

- `TestSaveUndoKeepsStaged`: ok
- `TestSubmitStagedSpecKept`: ok

## Previous findings

- task-18-1/F1: verified (fixed)
- task-18-1/F2: verified (deferred)

## Findings

None.

## Requirements check

None.

## Needs human decision

None.
