---
review: task-15-1
slice: 0002-spec-and-state
kind: task
target: task 15
round: 1
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

`ReviewSpec` follows Flow 5 and gives the spec's exact texts for R24,
R25, R27, R10's `Tail`, and R6. The reviewed commit, the lock handling,
and each Ctrl-C point are right. One medium problem: `TestReviewPasses`
does not check that the `review 1 passed` commit holds the filled
review file, so dropping it from that commit still passes (F1).

## Checked

- Commits: db7f630
- Files read: internal/slice/review.go, internal/slice/review_test.go,
  internal/slice/save.go, internal/slice/submit.go,
  internal/slice/save_test.go (helpers), internal/slice/submit_test.go
  (helpers), internal/slice/validate_test.go (wantErrMsg),
  internal/review/review.go, internal/review/placeholder.go,
  internal/state/state.go, internal/state/moves.go,
  internal/state/replay.go; spec.md R6, R10, R20, R24 to R28, Out of
  scope, SD10; plan.md Flow 5, Review files, Interfaces, Interrupts;
  reviews/plan-5.md F1; the brief
- Ran:
  go test ./internal/slice/ -run TestReview -count=1 (pass)
  go vet ./internal/slice/ (pass)
  in a copy in the scratchpad: review.finished's Prepare with Files nil, then go test ./internal/slice/ -run TestReview (pass, see F1)
- Requirements: R10, R12, R24, R25, R27, R28; also R6's error
- Interfaces: Step, ReviewSpec, var reviewer = review.Placeholder(),
  saveOp.Tail, review.FrontMatter, review.ReadResult match the task

## Tests

- `TestReviewPasses`: weak, see F1
- `TestReviewStates`: ok
- `TestReviewFileExists`: ok
- `TestReviewRound2`: ok
- `TestReviewNoName`: ok
- `TestReviewFinishFails`: ok
- `TestReviewMovedFirst`: ok

## Previous findings

None.

## Findings

### F1. No test checks that the review.finished commit holds the review file

Level: medium
Where: internal/slice/review_test.go:48-73
Breaks: R27
Why: R27 and S20 say the `review 1 passed` commit holds the filled
  `reviews/spec-1.md` and the log. `TestReviewPasses` checks the files
  of the `review 1 started` commit only. For the second commit it reads
  the review file from disk, where the reviewer wrote it. In a copy of
  the code I set `Files` of the `review.finished` prepare step to nil:
  the commit then holds only the log, and every `TestReview` test still
  passes.
Fix: check `git show --name-only --format= HEAD` is the log and
  `reviews/spec-1.md`, and that `HEAD:` of the review file holds the
  placeholder text. Also check the commit count is exactly 2 more.

### F2. No test shows the lock is free while the reviewer runs

Level: low
Where: internal/slice/review_test.go
Breaks: nothing
Why: The task's goal is "the reviewer with no lock held" (R6: never
  while a reviewer runs). The code is right: `step` and the reviewer run
  after the first `saveEvent` returned, and its deferred `free` ran. But
  `TestReviewMovedFirst` commits with plain git, which takes no lock, so
  it would also pass if the reviewer ran under the lock. A Ctrl-C while
  the reviewer runs (plan Interrupts) is also not tested here. Both can
  be tested by swapping the package's `reviewer` for a test reviewer
  that saves a `slice.titled` through `saveEvent`, or that cancels the
  ctx.

## Requirements check

None.

## Needs human decision

None.
