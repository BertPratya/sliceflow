---
review: task-14-1
slice: 0002-spec-and-state
kind: task
target: task 14
round: 1
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

`Submit` does R19 to R23 right. The checks run in R20's order, and refused
and accepted submits save the right data, files, and message. Two medium
problems: a Ctrl-C while the current branch is read (no name) returns
`context canceled`, not R10's error (F1), and no test catches a wrong order
of R20's state check and R19 (F2).

## Checked

- Commits: a1be3eb
- Files read: internal/slice/submit.go, internal/slice/submit_test.go,
  internal/slice/save.go, internal/slice/open.go, internal/slice/title.go,
  internal/slice/validate.go, internal/slice/next.go (specChanged),
  internal/slice/status.go (currentSliceName), internal/git/git.go (Run),
  internal/state/state.go, internal/state/moves.go, internal/cli/title.go,
  docs/templates/tasks.md, docs/slices/0002-spec-and-state/spec.md
  (R10, S6, S7), the brief
- Ran:
  go test ./internal/slice/ -count=1 (pass)
  go vet ./internal/slice/ (pass)
  mutations in a scratch copy, go test -run TestSubmit for each:
  R19 read moved before the state check (all pass: not caught, see F2);
  refused saved with result accepted (TestSubmitRefused fails);
  Before always draft (TestSubmitStates fails);
  no spec.md in Keep (TestSubmitRefused, TestSubmitAccepted fail);
  notes counted as warnings (four tests fail);
  in_review let through (TestSubmitStates fails);
  changed check turned around (TestSubmitStates fails);
  no slice.created check removed (TestSubmitStates fails)
  a probe test in the scratch copy: Submit with a cancelled ctx and no
  name returns `*errors.errorString context canceled` (see F1)
- Requirements: R12, R19, R20, R21, R22, R23
- Interfaces: Submission{Problems, Refused, Before} and
  Submit(ctx, r, name) match the task; uses openSlice, saveEvent,
  specChanged, spec.Parse, spec.Check, currentSliceName with
  `spec submit`

## Tests

- `TestSubmitRefused`: weak, see F3
- `TestSubmitAccepted`: ok
- `TestSubmitStates`: ok
- `TestSubmitNoName`: ok
- `TestSubmitNoSpec`: ok

## Previous findings

None.

## Findings

### F1. Ctrl-C while reading the current branch gives "context canceled"

Level: medium
Where: internal/slice/submit.go:29-35
Breaks: R10
Why: with no name, Submit calls currentSliceName first. On a cancel,
  git.Run returns ctx.Err(), and currentSliceName passes it on as it is.
  Submit returns it without the ctx.Err() check that the openSlice call
  has. A probe test with a cancelled ctx and name `""` got
  `context canceled`, not
  `could not save the event: interrupted. Nothing was changed.` The
  lead's rule for this task: a Ctrl-C before any event is saved gives
  InterruptedSave(), not git's error. The window is small (one git
  command), but `spec submit` with no name is the common case.
Fix: after currentSliceName fails, return InterruptedSave() when
  ctx.Err() != nil, as after openSlice. Add a unit test with a cancelled
  ctx and no name.

### F2. No test catches a wrong order of R20's state check and R19

Level: medium
Where: internal/slice/submit_test.go:117-214, internal/slice/submit_test.go:244-256
Breaks: AGENTS.md, Code rules: every change that adds behavior adds a
  test (R20's last sentence: the state check, then R19)
Why: the code has the right order. But in a scratch copy I moved the
  spec.md read to the top of Prepare, before the state check. All five
  TestSubmit tests still passed. Each state test writes spec.md, and
  TestSubmitNoSpec uses a draft slice, so no test has a wrong state and
  a missing spec.md at once. With that broken order, a submitted slice
  with no spec.md gives R19's error instead of S16's, and no test fails.
Fix: in TestSubmitStates, add a case: state submitted and spec.md
  deleted gives S16's error, not R19's.

### F3. TestSubmitRefused does not use S14's spec

Level: low
Where: internal/slice/submit_test.go:56, internal/slice/submit_test.go:205
Breaks: task 14 (Done when, TestSubmitRefused: "S14's spec")
Why: Done when names S14's spec, whose one warning is the open questions
  Q1 and Q2 (S6). The test uses 0001's spec plus a `<!-- todo -->`
  comment, so its one warning is a template comment. The test has no
  Scenario: line, so the template's exact-values rule for e2e tests does
  not bind it, and Submit only counts levels, so the outputs it checks
  (0 errors, 1 warning, Before draft, the commit, the log line) prove
  R21 and R22 as well. But it misses one point S14 shows: two open
  questions are 1 warning, not 2. The "refused keeps review" case uses
  the same comment.
Fix: in the test, list Q1 and Q2 under Open questions in place of the
  comment, as S6 says.

## Requirements check

None.

## Needs human decision

None.
