---
review: task-10-2
slice: 0002-spec-and-state
kind: task
target: task 10
round: 2
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

Both findings of round 1 are fixed, and each new test fails with its fix
undone. R7's message and S17 are unchanged. One low point from the F2
fix: on Linux and macOS, a lasting permission problem now waits the full
10 s before its error (F1).

## Checked

- Commits: 76aa1c8, 2e8f5ea (diff 76aa1c8..2e8f5ea for this round)
- Files read: internal/slice/lock.go, internal/slice/lock_test.go,
  internal/cli/output.go (how a plain error is printed), the brief,
  plan.md (The lock, TD2, Risks), task-10-1.md, the ASSUMED lines of
  2e8f5ea
- Ran:
  go test -count=1 -race -run TestLock ./internal/slice/ in a copy of HEAD (pass)
  the same copy, F2 fix undone (no ErrPermission exception): TestLockCreateRetriesPermission fails (permission denied)
  the same copy, F1 fix undone (freeTries = 1): TestLockFreeRetriesRemove fails (lock file still there)
  a throwaway test in the copy, read-only .sliceflow/locks/, LockWait 300 ms: raw "open ...: permission denied" after 302 ms
- Requirements: R6 (the free), R7
- Interfaces: `Repo.LockWait` and `takeLock` unchanged, match the task

## Tests

- `TestLockFile`: ok
- `TestLockBusy`: ok
- `TestLockUnreadable`: ok
- `TestLockWaits`: ok

## Previous findings

- task-10-1/F1: verified (fixed)
- task-10-1/F2: verified (fixed)

## Findings

### F1. A lasting permission error waits the full 10 s on every system

Level: low
Where: internal/slice/lock.go:96-106
Breaks: nothing
Why: the F2 fix treats every `fs.ErrPermission` from the create as busy.
  That is right for Windows' "delete pending" state, but on Linux and
  macOS a permission error from `open` is never short-lived. Input: a
  read-only `.sliceflow/locks/` folder. Result: the command waits the
  full `LockWait`, then prints `Error: open <full path>: permission
  denied` (I ran it with 300 ms: the error came after 302 ms). This does
  not break R7: R7 is about a lock file that exists, and here none does,
  so the raw error, not R7's busy message, is the honest one (the
  ASSUMED line agrees). S17 is unchanged. The plan's The lock says to
  retry "when the file exists"; retrying on a permission error goes a
  little past its words but is what F2 asked for. The only cost is the
  wait, on a rare setup problem.
Fix: retry on `fs.ErrPermission` only when `runtime.GOOS == "windows"`;
  elsewhere return it at once.

## Requirements check

None.

## Needs human decision

None.
