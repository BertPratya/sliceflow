---
review: task-06-2
slice: 0002-spec-and-state
kind: task
target: task 06
round: 2
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

The fixes for task-06-1/F1 to F3 work, and the new F1 tests fail on the
round 1 code, as they should. F4's wontfix reason holds. One new low
finding: specChanged now also hides a cancelled context, so its error
result is never set (F1).

## Checked

- Commits: 2e9534f, 0716b44 (`git diff 2e9534f 0716b44`, and the whole
  task `git diff 2e9534f~1 0716b44`), and the ASSUMED lines of both
- Files read: internal/slice/worktree.go, internal/slice/worktree_test.go,
  internal/slice/next.go, internal/slice/next_test.go,
  internal/slice/status.go (readSlice), internal/slice/status_test.go
  (TestStatusWorktreeDeletedByHand), internal/git/git.go (Run),
  internal/cli/create.go (the only signal.NotifyContext), the brief,
  reviews/task-06-1.md
- Ran:
  go vet ./... (pass)
  go test -count=1 ./... (pass)
  a copy of 0716b44 in the scratchpad, with worktree.go, next.go, and status.go from 2e9534f: go test -run 'TestFindWorktree|TestSpecChanged|TestStatusWorktreeDeletedByHand|TestSpecNext' (TestFindWorktreeDeletedByHand, TestSpecChangedNoFolder, and TestStatusWorktreeDeletedByHand fail with the round 1 code: `deleted: true <nil>`, and `chdir ...: no such file or directory` twice; the others pass)
- Requirements: R1, R3, R29, R30, R34
- Interfaces: findWorktree and specChanged keep the task's signatures

## Tests

- `TestSpecNext`: ok
- `TestStatusState`: ok
- `TestSpecChanged`: ok
- `TestFindWorktree`: ok

## Previous findings

- task-06-1/F1: verified (fixed)
- task-06-1/F2: verified (fixed)
- task-06-1/F3: verified (fixed)
- task-06-1/F4: verified (wontfix)

## Findings

### F1. specChanged hides a cancelled context too

Level: low
Where: internal/slice/next.go:30-31, internal/slice/status.go:177
Breaks: nothing
Why: the fix for task-06-1/F1 made specChanged return false, nil for
  every error that is not a `*git.Error`. git.Run returns `ctx.Err()`
  when the context is done, so a cancel during the `git diff` now reads
  as "not changed", with no error. specChanged's error result is never
  set any more, so the `_` in readSlice hides nothing today. Nothing
  breaks for `slice status` or `slice list`: they do not cancel their
  context (only `slice create` uses signal.NotifyContext), and `slice
  list` would fail on its next git command anyway. But task 14 uses
  specChanged for R20, inside a command that writes. There a cancel
  should stop the command, not count as "not changed". Flow 1 step 3's
  "any other failure" is about git failing, not the user stopping the
  command.
Fix: in specChanged, return `ctx.Err()` when it is set, and false, nil
  for the other errors. readSlice then returns that error instead of `_`.

## Requirements check

None.

## Needs human decision

None.
