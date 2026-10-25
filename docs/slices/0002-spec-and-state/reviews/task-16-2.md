---
review: task-16-2
slice: 0002-spec-and-state
kind: task
target: task 16
round: 2
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

Both earlier findings are fixed. `Submit` now returns the kept review's
round in `Submission.BeforeRound`, read under the lock, and the CLI prints
R22's last sentence from it with no second core call. A new e2e test and
the unit test cover that line. Nothing else in task 16 changed. No new
findings.

## Checked

- Commits: b6fd199..1506b15 (task code: 1506b15)
- Files read: internal/slice/submit.go, internal/slice/submit_test.go
  (TestSubmitStates), internal/cli/submit.go,
  internal/e2e/submit_review_commands_test.go, reviews/task-16-1.md,
  tasks.md (task 16, R003: Interfaces, Done when, touches), spec.md (R22),
  the brief
- Ran:
  go test ./internal/slice/ ./internal/e2e/ -run 'TestSubmit|TestReview' -count=1 (pass)
  git diff --stat b6fd199 1506b15 (code changes only in the 4 files task 16 touches; the rest are review and status documents)
  grep for slice.Status and BeforeRound in internal/cli/submit.go and internal/cli/review.go (the second Status call is gone; BeforeRound is used once)
- Requirements: R22
- Interfaces: `Submission{Problems, Refused, Before, BeforeRound}` matches
  task 16's Interfaces (R003)

## Tests

- `TestSubmitRefusedCmd`: ok
- `TestSubmitAcceptedCmd`: ok
- `TestSubmitTwice`: ok
- `TestReviewCmd`: ok
- `TestReviewDraft`: ok
- `TestReviewInReview`: ok
- `TestReviewNothing`: ok
- `TestSubmitUnchanged`: ok
- `TestSubmitChangedStatus`: ok
- `TestReviewRound2Cmd`: ok
- `TestSubmitRefusedReviewed`: ok

## Previous findings

- task-16-1/F1: verified (fixed). `Submit`'s Prepare sets `BeforeRound`
  from `st.Spec.Review.Round` only when `st.Spec.Status` is
  `review_passed`, so it is the passed review that the refused submit
  keeps, and 0 for a draft. internal/cli/submit.go prints from it; the
  `slice.Status` call after the save is gone.
- task-16-1/F2: verified (fixed). `TestSubmitRefusedReviewed` checks exit
  1, empty stderr, and that stdout ends with R22's full line
  `... (0 errors, 1 warning). It keeps the result of review 1.`, as the
  approved Done when says ("as the last line of stdout"). It fails if the
  review_passed branch is lost (`It stays a draft.`) or the round is not
  set (`review 0`). The problem lines before it are R21's and are checked
  in full by `TestSubmitRefusedCmd`. `TestSubmitStates` also checks
  `BeforeRound` 1.
- task-16-1/F3: verified (deferred). R28 and spec.md Out of scope leave a
  review that does not pass to 0004; it cannot happen in this slice.

## Findings

None.

## Requirements check

None.

## Needs human decision

None.
