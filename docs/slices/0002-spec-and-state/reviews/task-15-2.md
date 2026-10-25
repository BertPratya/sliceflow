---
review: task-15-2
slice: 0002-spec-and-state
kind: task
target: task 15
round: 2
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

Both findings of round 1 are fixed, and the new tests fail when the
code is broken. The lead's extra check holds: `TestReviewPasses` fails
when the `review 1 started` commit holds the filled file, more than the
front matter, or no review file. One low problem: a comment in a new
test says the wrong thing (F1).

## Checked

- Commits: db7f630, 46ddee6 (diff db7f630..46ddee6)
- Files read: internal/slice/review.go, internal/slice/review_test.go,
  internal/review/placeholder.go, internal/review/review.go
  (FrontMatter, ReadResult), internal/slice/create_test.go (gitOut),
  internal/slice/save_test.go (commitCount, readFile); the brief;
  reviews/task-15-1.md
- Ran:
  go test ./internal/slice/ -count=1 (pass)
  in a copy in the scratchpad, go test ./internal/slice/ -run TestReview, once per change to review.go:
  started commit holds the front matter, then extra text (TestReviewPasses fails at review_test.go:64, the started file check)
  started commit holds the front matter and the placeholder text (TestReviewPasses fails)
  review file written to disk but not in the started commit (TestReviewPasses fails at review_test.go:61 and 63)
  started commit with no review file (TestReviewPasses, TestReviewRound2, TestReviewMovedFirst fail)
  front matter with a wrong reviewer (TestReviewPasses fails at review_test.go:76)
  finished commit with no review file (TestReviewPasses fails at review_test.go:71 and 76)
  a lock file made while the reviewer runs (TestReviewNoLockWhileReviewing fails)
  no ctx check after the reviewer (TestReviewInterruptedWhileReviewing fails)
- Requirements: R10, R12, R24, R25, R27, R28; also R6
- Interfaces: Step, ReviewSpec, reviewer unchanged since round 1

## Tests

- `TestReviewPasses`: ok
- `TestReviewStates`: ok
- `TestReviewFileExists`: ok
- `TestReviewRound2`: ok
- `TestReviewNoName`: ok
- `TestReviewFinishFails`: ok
- `TestReviewMovedFirst`: ok

## Previous findings

- task-15-1/F1: verified (fixed)
- task-15-1/F2: verified (fixed)

## Findings

### F1. A comment in TestReviewNoLockWhileReviewing says the wrong thing

Level: low
Where: internal/slice/review_test.go:274
Breaks: nothing
Why: The comment says "The reviewer wrote no file, so finishing fails".
  But the review file already holds R25's front matter from the started
  commit. `ReadResult` gives `needs_human`, and `review.finished` is
  saved with `review 1 needs the human`. The test ignores the error,
  so it is still right; only the comment is wrong.
Fix: say the review finishes with needs the human, and that only the
  lock matters here.

## Requirements check

None.

## Needs human decision

None.
