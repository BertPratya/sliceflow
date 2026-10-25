---
review: task-11-3
slice: 0002-spec-and-state
kind: task
target: task 11
round: 3
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

task-11-2/F1 is fixed: `staleIndexLock` now returns false at once when
`ctx` was not cancelled, so a plain commit failure never removes another
git's `index.lock`. The window and the 2 s margin stay. The new
`TestStaleIndexLock` covers all four cases, and each case fails when its
check is removed. No new findings.

## Checked

- Commits: d994f92, 5892c61, 7439232 (the diff since round 2:
  7439232)
- Files read: internal/slice/save.go (the undo, `staleIndexLock`,
  `failSave`), internal/slice/save_test.go (`TestStaleIndexLock`),
  reviews/task-11-1.md, reviews/task-11-2.md, status.md (Lead's
  decisions on task-11-1/Q1 and F2), the brief, the commit message of
  7439232 (its ASSUMED line)
- Ran:
  go test ./internal/slice/ -count=1 (pass)
  a copy of the repo, outside the worktree, with `fi.ModTime().After(stopped)` removed: go test -run TestStaleIndexLock (fails, case "newer than stopped": returned true, lock removed)
  the same copy with `fi.ModTime().Before(started.Add(-lockMargin))` removed: go test -run TestStaleIndexLock (fails, case "older than started less margin": returned true, lock removed)
- Requirements: R6, R8, R10
- Interfaces: prepared, saveOp, and saveEvent(ctx, r, name, worktree, op)
  still match the task

## Tests

- `TestSaveCommit`: ok
- `TestSaveOnlyOwnFiles`: ok
- `TestSaveNoNewline`: ok
- `TestSavePrepareFails`: ok
- `TestSaveNotAllowed`: ok
- `TestSaveHookFails`: ok
- `TestSaveTail`: ok
- `TestSaveCancelled`: ok
- `TestSaveTwoAtOnce`: ok
- `TestSaveSeqAfterSkipped`: ok
- `TestSaveLeftoverTemp`: ok

## Previous findings

- task-11-2/F1: verified (fixed)

## Findings

None.

## Requirements check

None.

## Needs human decision

None.
