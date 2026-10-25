---
review: task-02-1
slice: 0002-spec-and-state
kind: task
target: task 02
round: 1
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

The task covers R1 to R4 correctly: every move of R2 is in place, R4's three
warning forms come out exactly as the spec words them, and the tests are
real. Tasks 06 and 11 get what they need from `Replay` and `Allowed`. One
low finding: no test checks the review's reviewer, file, and commit, or the
title from `slice.created`, which task 06 relies on (F1).

## Checked

- Commits: 804587e
- Files read: internal/state/state.go, internal/state/moves.go,
  internal/state/replay.go, internal/state/state_test.go,
  internal/eventlog/event.go, internal/eventlog/parse.go (the data types
  `Parse` builds, which are value types, so the type switches in
  moves.go match), docs/slices/0002-spec-and-state/spec.md (R1 to R5, R29
  to R31, R34), docs/slices/0002-spec-and-state/tasks.md (tasks 06 and 11),
  the brief
- Ran:
  go test -count=1 -v ./internal/state/ (pass, 4 tests)
  go vet ./internal/state/ (pass)
  gofmt -l internal/state (prints nothing)
- Requirements: R1, R2, R3, R4
- Interfaces: `State`, `Doc`, `Review`, `Counts`, `Replay(events []eventlog.Event) (State, []string)`,
  and `Allowed(s State, e eventlog.Event) string` match the task exactly

## Tests

- `TestMoves`: ok
- `TestReplayCreated`: ok
- `TestReplayNotAllowed`: ok
- `TestReplayRounds`: ok

## Previous findings

None.

## Findings

### F1. No test checks the review's reviewer, file, and commit, or the title from slice.created

Level: low
Where: internal/state/state_test.go:24-27, internal/state/state_test.go:113-122
Breaks: nothing
Why: `apply` copies `Reviewer`, `File`, and `Commit` from `review.started`
  into `Spec.Review`, and takes `Title` from `slice.created` (an ASSUMED
  choice, which agrees with R34). The code is right, but no test reads these
  fields: grep finds no check of `.Reviewer`, `.File`, `.Commit`, or a title
  from `slice.created`. If `apply` dropped them, all tests would still
  pass. Task 06 needs `Commit` for R30's `git diff`, and `Reviewer` and the
  title for R29 and R34's texts.
Fix: in `TestReplayRounds`, also check `Review.Reviewer`, `File`, and
  `Commit`; in `TestReplayCreated`, add a `slice.created` with a title.

## Requirements check

None.

## Needs human decision

None.
