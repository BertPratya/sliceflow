---
review: task-17-1
slice: 0002-spec-and-state
kind: task
target: task 17
round: 1
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

The tests of task 17 are real: each one checks the exact output of its
scenario, and the ones I broke on purpose failed. I found no high or medium
problem. One low gap: `TestSubmitBusy` does not check that the other
command's lock file is still there afterwards (F1).

## Checked

- Commits: 2236606
- Files read: internal/e2e/submit_review_failures_test.go,
  internal/e2e/no_network_check_test.go, internal/e2e/helpers_test.go,
  internal/e2e/submit_review_commands_test.go (the helpers it uses),
  internal/e2e/create_undo_test.go (`writeUndoHook`, `TestCreateInterrupt`),
  internal/e2e/status_spec_test.go (`sliceWT`), internal/testrepo/testrepo.go,
  internal/slice/save.go, internal/slice/lock.go, internal/git/git.go
- Ran:
  go test ./internal/e2e/ -run (the 9 tests of Done when, and TestNoNetwork) -count=1 -v (pass; TestSubmitBusy 10.2 s)
  in a copy of the tree, the undo does not write the log back (TestSubmitHookFails and TestSubmitInterrupted fail)
  in a copy, the undo does not remove the files it created (TestReviewAgain fails, with S22's error)
  in a copy, the undo does not unstage (TestSubmitHookFails fails)
  in a copy, git commit takes the whole index instead of `--only` the event's files (TestSubmitOwnFiles fails)
  in a copy, `freeLock` does nothing (TestSubmitInterrupted fails: lock file left)
- Requirements: R6, R7, R8, R9, R10, R24, R25, R29, R37, and R40 of
  0001-slice-and-events

Notes on the lead's points:

- `TestSubmitBusy` compares stderr with S17's exact line, stdout with "",
  the exit code with 1, the wait with at least 9 s, and HEAD before and
  after. The lock file is made in the main checkout, in the form
  `takeLock` writes (`{"pid":...,"since":...}`).
- `TestSubmitInterrupted` skips on Windows as `TestCreateInterrupt` does,
  waits for a marker from the sleeping `pre-commit` hook before it sends
  `os.Interrupt`, checks R10's exact `interrupted` line, the log on disk
  against HEAD, and that the lock file is gone.
- `TestNoNetworkNewCommands`: the guard is sound. `run` calls
  `testrepo.Setup`, whose `clearLocalEnv` runs `git rev-parse` with
  `GIT_TRACE` set, so the test itself does write lines into the trace; but
  no test code runs `git commit` after `GIT_TRACE` is set, so a `commit`
  line can only come from the binary. Every git command of the binary goes
  through `git.Run`, the only `exec.Command` outside tests, which passes
  `os.Environ()` on. So a trace that works for `spec submit` works the same
  way for `spec validate` and `template show`.
- The `commit-msg` helper writes `#!/bin/sh` and a `grep -q ... "$1"` line
  through `writeUndoHook`, with `\n` line endings. Git for Windows runs
  hooks with its own `sh` and ships `grep`, so it works there too. Hooks
  live in the common git folder, so the worktree's commits run them.

## Tests

- `TestSubmitBusy`: ok
- `TestSubmitHookFails`: ok
- `TestSubmitOwnFiles`: ok
- `TestReviewFileExistsCmd`: ok
- `TestReviewCommitRefused`: ok
- `TestReviewAgain`: ok
- `TestReviewStopsHalfway`: ok
- `TestSubmitInterrupted`: ok
- `TestNoNetworkNewCommands`: ok

## Previous findings

None.

## Findings

### F1. TestSubmitBusy does not check that the other command's lock stays

Level: low
Where: internal/e2e/submit_review_failures_test.go:151-172
Breaks: nothing
Why: R6 says a command removes its own lock. A busy command never took
  the lock, so it must leave the other command's file alone. The test
  checks the error, the wait, and that no commit was made, but not that
  `.sliceflow/locks/login-google.lock` still exists afterwards. A change
  that deleted the file on the busy error would still pass. The code is
  right today (`takeLock` returns `busyError` without removing anything).
Fix: after the run, check that the lock file still exists with the same
  content.

## Requirements check

None.

## Needs human decision

None.
