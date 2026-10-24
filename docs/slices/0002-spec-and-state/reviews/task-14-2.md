---
review: task-14-2
slice: 0002-spec-and-state
kind: task
target: task 14
round: 2
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

All three findings of round 1 are fixed, and the new tests catch the
bugs they were about. The task does R12, R19 to R23 as the spec says. No
new findings.

## Checked

- Commits: a1be3eb, e5536c4 (the diff since round 1: a1be3eb..e5536c4)
- Files read: internal/slice/submit.go, internal/slice/submit_test.go,
  internal/slice/status.go (currentSliceName), the brief,
  reviews/task-14-1.md, the commit message of e5536c4
- Ran:
  go test ./internal/slice/ -run TestSubmit -count=1 -v (pass, 6 tests)
  mutation in a scratch copy, the F1 fix removed (TestSubmitNoNameInterrupted fails)
  mutation in a scratch copy, the R19 read moved before the state check (TestSubmitStates/submitted,_no_spec.md fails)
  mutation in a scratch copy, notes counted as warnings (TestSubmitRefused, TestSubmitAccepted, TestSubmitStates, TestSubmitNoName fail)
- Requirements: R12, R19, R20, R21, R22, R23
- Interfaces: Submission{Problems, Refused, Before} and
  Submit(ctx, r, name) match the task

## Tests

- `TestSubmitRefused`: ok
- `TestSubmitAccepted`: ok
- `TestSubmitStates`: ok
- `TestSubmitNoName`: ok
- `TestSubmitNoSpec`: ok

## Previous findings

- task-14-1/F1: verified (fixed)
- task-14-1/F2: verified (fixed)
- task-14-1/F3: verified (fixed)

## Findings

None.

## Requirements check

None.

## Needs human decision

None.
