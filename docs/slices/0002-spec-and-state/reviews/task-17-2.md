---
review: task-17-2
slice: 0002-spec-and-state
kind: task
target: task 17
round: 2
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

F1 is fixed. `TestSubmitBusy` now checks that the other command's lock
file is still there, with the same content. I broke the code on purpose in
two throwaway copies, and the new check failed both times. Nothing else
changed in the code. No new findings.

## Checked

- Commits: 2236606..a10ec7e
- Files read: internal/e2e/submit_review_failures_test.go (the diff since
  round 1), internal/slice/lock.go (`takeLock`, `busyError`),
  docs/slices/0002-spec-and-state/reviews/task-17-1.md
- Ran:
  git diff --stat 2236606 a10ec7e -- internal cmd (one file: internal/e2e/submit_review_failures_test.go, +5 -1)
  in a throwaway copy of HEAD under the scratchpad, the busy error removes the lock file after it builds its message: TestSubmitBusy fails at line 81, "lock file changed: \"\", ... no such file or directory"
  in a second throwaway copy, the busy error writes "x\n" into the lock file: TestSubmitBusy fails at line 81, "lock file changed: \"x\\n\", <nil>"
  go test ./internal/e2e/ -run (the 9 tests of Done when, and TestNoNetwork) -count=1 in the worktree (pass); git status is clean afterwards
- Requirements: R6, R7

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

- task-17-1/F1: verified (fixed)

## Findings

None.

## Requirements check

None.

## Needs human decision

None.
