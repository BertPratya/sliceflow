---
review: task-11-2
slice: 0002-spec-and-state
kind: task
target: task 11
round: 2
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

F1, F2, F4, and Q1 are fixed as the review and the lead's decisions say,
and a Ctrl-C before, during, and after `git commit` now gives the right
result. One new medium problem, caused by F2's fix: the lock rule no
longer needs a stopped commit, so after a plain commit failure that races
with another git, the undo can delete that git's live `index.lock`
(F1). The 2 s margin is what lets that lock through.

## Checked

- Commits: d994f92, 5892c61 (the whole task: d994f92~1..5892c61)
- Files read: internal/slice/save.go, internal/slice/save_test.go (the
  diff since round 1, and the round 1 tests), internal/git/git.go (Run,
  Error), reviews/task-11-1.md, status.md (Lead's decisions on
  task-11-1/Q1 and F2), the brief, the commit messages (ASSUMED lines
  and the Known gap)
- Ran:
  go test ./internal/slice/ -run TestSave -count=1 -v (pass, 17 tests)
  a scratch test in a copy of the repo, outside the worktree: an
    `index.lock` made now, `staleIndexLock` called with `started` 1 s
    later, `stopped` 50 ms after that, and a reset error
    `Unable to create '...index.lock': File exists.` (it returns true,
    and the lock is gone)
- Requirements: R2, R6, R8, R9, R10
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

- task-11-1/F1: verified (fixed)
- task-11-1/F2: verified (fixed)
- task-11-1/F3: verified (deferred)
- task-11-1/F4: verified (fixed)

## Findings

### F1. After a plain commit failure, the undo can delete another git's index.lock

Level: medium
Where: internal/slice/save.go:160-171, internal/slice/save.go:263,
  internal/slice/save.go:269-284
Breaks: R8
Why: Round 1 removed the lock only when `ctx` was cancelled. The fix
  dropped that guard (ASSUMED: "does not also need a cancelled ctx"). But
  only a git that was killed leaves its lock behind. A git commit that
  fails in the normal way (a hook exits 1, or the lock is taken) removes
  its own lock. So when the commit was not stopped, a lock found at reset
  time always belongs to another process. This case can happen: `git add`
  ends, then `rev-parse HEAD` runs (a few ms), then `git commit` starts.
  If an editor's `git status` takes `index.lock` in that gap, `git commit`
  fails with `index.lock': File exists`, and so does the reset. The lock's
  time is then a few ms before `started`. That is inside the 2 s margin,
  so `staleIndexLock` deletes a live lock, and two processes write the
  index at once. My scratch test shows the helper does this for a lock
  1 s older than `started`. Ctrl-C is not needed. The lead's decision
  speaks of "the stopped `git commit`", and R8 says other staged changes
  stay as they were.
  On the lead's points. The lower bound can be reached only through this
  race. A lock made before `git add` ran would have made `add` fail, and
  `started` and `stopped` are then still zero, so nothing is removed;
  `TestSaveOldIndexLockKept` tests that zero window, not the lower bound.
  The margin itself is still useful: on a file system with coarse times,
  the killed git's own lock can carry a time just before `started`. It is
  safe once the rule also needs a stopped commit. A lock made in the 2 s
  before the start, while our commit was then killed, needs another git
  to take the lock in the gap of a few ms and keep it until the kill. Our
  commit would then fail at once on that lock, before the Ctrl-C comes.
  That is too rare to matter against plan "Interrupts" and R8.
  F1 at each moment works. Ctrl-C before `git commit`: `add` or
  `rev-parse` returns `ctx.Err()`, HEAD has not moved, so the undo runs
  with the reason `interrupted`. During `pre-commit`: git is killed, HEAD
  has not moved, so the same result (`TestSaveCancelled`). During
  `post-commit`: HEAD moved, so no undo, and the result is success
  (`TestSaveCancelAfterCommit`). After git returns: the last `rev-parse`
  uses `context.WithoutCancel`, so the result is success.
Fix: remove the lock only when `ctx.Err() != nil`, as round 1 did, and
  keep the window and the 2 s margin. Cover the bounds with a direct unit
  test of `staleIndexLock` that passes in given times and a crafted
  `*git.Error`: a lock inside the window (removed), one older than
  `started` less the margin (kept), and one newer than `stopped` (kept).
  This is deterministic, and is the right way to close the commit's
  "Known gap", because the end-to-end path cannot reach that bound on
  purpose.

## Requirements check

None.

## Needs human decision

None.
