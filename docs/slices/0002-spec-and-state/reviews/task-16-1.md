---
review: task-16-1
slice: 0002-spec-and-state
kind: task
target: task 16
round: 1
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

Both commands print what S14 to S16, S20, S21, and S25 to S29 show,
character for character, on the right stream and with the right exit
code, and the ten tests are real. Two medium problems: the CLI finds
R22's review round with a second call to `slice.Status` after the event
is saved, which puts core logic in the CLI and can lose the
`Submit refused:` line (F1); and that branch has no test (F2). The silent
exit 0 after a review that does not pass is allowed by Out of scope, and
cannot happen in this slice (F3, low).

## Checked

- Commits: b6fd199
- Files read: internal/cli/submit.go, internal/cli/review.go,
  internal/e2e/submit_review_commands_test.go, internal/cli/validate.go,
  internal/cli/title.go, internal/cli/problems.go, internal/cli/status.go,
  internal/slice/submit.go, internal/slice/review.go,
  internal/slice/status.go, internal/state/moves.go,
  internal/e2e/helpers_test.go, the e2e helpers in
  status_spec_test.go, status_command_test.go, validate_command_test.go,
  title_command_test.go, plan.md (Flows 3 to 5, CLI, Interrupts, Core
  API), spec.md (R20 to R30, R37, Out of scope), 0001's R38, status.md
  (Findings)
- Ran:
  go test ./... (pass)
  go test ./internal/e2e/ -run 'TestSubmit|TestReview' -count=1 (pass)
  extra tests in a copy of the tree, outside the worktree: a refused submit from `review_passed` prints `... (0 errors, 1 warning). It keeps the result of review 1.`, exit 1, stderr empty; a spec with an error prints `(1 error, 1 warning). It stays a draft.`; `spec submit login-google` with a name, and `spec review login-google` from the main checkout, give the same lines (pass)
- Requirements: R20, R21, R22, R23, R24, R25, R26, R27, R28, R29, R30, R37
- Interfaces: `Submit`, `Submission`, `ReviewSpec`, `Step`, `printProblems`
  are used with the names and fields of tasks 09, 14, and 15

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

## Previous findings

None.

## Findings

### F1. R22's review round is read by a second core call after the save

Level: medium
Where: internal/cli/submit.go:69-77
Breaks: plan.md Core API (the CLI calls the core, and a later server
  calls the same functions); AGENTS.md Code rules (shared Go logic goes
  in internal/)
Why: after `slice.Submit` has saved and committed the refused event, the
  CLI calls `slice.Status` to find the round for
  `It keeps the result of review <n>.`. The round is already known inside
  `Submit`'s Prepare step, under the lock (`st.Spec.Review.Round`), but
  `Submission` does not return it. So a later server would have to repeat
  this second read. The second read also has two real effects. (1) When
  `Status` fails, the CLI prints an `Error:` line on stderr and exits 1,
  but the event was saved, and the `Submit refused:` line is never
  printed; the problem lines are already on stdout. (2) It reads the log
  again without the lock, so a command run in between can change the
  state. The lead asked about this point; the output is right in the
  normal case (I ran it: `It keeps the result of review 1.`).
Fix: have `Submit` return the round of the kept review with `Before`
  (e.g. a `BeforeRound int` in `Submission`), and print from it. This
  needs internal/slice/submit.go in task 16's touches.

### F2. R22's review_passed sentence has no test

Level: medium
Where: internal/e2e/submit_review_commands_test.go, internal/cli/submit.go:69-77
Breaks: AGENTS.md Code rules (every change that adds behavior adds a test)
Why: task 16 covers R22. `TestSubmitRefusedCmd` checks only
  `It stays a draft.`. No test runs a refused submit from
  `review_passed`, so breaking the branch at submit.go:70 or the round at
  submit.go:76 fails nothing. The worker says this in its ASSUMED line.
  I checked it by hand in a copy of the tree (see Checked): the output is
  right today. The slice unit test (internal/slice/submit_test.go:248)
  checks `Before`, not the printed line.
Fix: add an e2e test: `reviewedSlice`, write a changed spec that has an
  open question, run `spec submit`, and check the full stdout ending
  `It keeps the result of review 1.`, exit 1, empty stderr.

### F3. A review that does not pass prints only its start, and exits 0

Level: low
Where: internal/cli/review.go:61-66
Breaks: nothing
Why: when `Step.Result` is not `passed`, the CLI prints only
  `Review <n> started: ...` and exits 0. R28 names only the passed case,
  and spec.md Out of scope sends "what `spec review` prints when a review
  does not pass" to 0004. Exit 0 also fits 0001's R38: the command did
  not fail; it saved both events. The placeholder always passes, so this
  cannot happen in this slice. It does not need the human now; 0004 must
  decide the lines and the exit code.
Fix: none in this slice; keep it for 0004 (backlog, with task-13-1/F1).

## Requirements check

None.

## Needs human decision

None.
