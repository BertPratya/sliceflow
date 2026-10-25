---
review: change-2
slice: 0002-spec-and-state
kind: change
target: docs/slices/0002-spec-and-state/tasks.md
round: 1
reviewer: doc-reviewer (claude opus)
---

# Review

## Summary

The change to task 16 is sound. `BeforeRound`, set in `Submit`'s
`Prepare` under the lock, gives R22's `review <n>` with no second read,
which fixes task-16-1/F1. The new `TestSubmitRefusedReviewed` covers
task-16-1/F2. No other task can run at the same time and touch the same
files. Only low findings: the risk label (F1), the unit test is not
named (F2), and the Steps do not name the core file (F3).

## Checked

- Files read: docs/slices/0002-spec-and-state/tasks.md (the diff from
  sliceflow/0002-spec-and-state/planning/R002, task 16 part only; tasks
  14, 16, and 17 in full; the touches and depends_on of every task),
  spec.md (R20, R21, R22, R23), plan.md (Flow 3, Core API),
  status.md (Tasks, Findings, the task 16 question),
  reviews/task-16-1.md, reviews/change-1.md, docs/templates/tasks.md
  (risk rules), docs/templates/review.md, .agents/skills/build-slice/SKILL.md
  (what risk changes), AGENTS.md, internal/slice/submit.go,
  internal/slice/submit_test.go (TestSubmitStates, "refused keeps
  review"), internal/cli/submit.go:55-85
- Ran:
  git diff sliceflow/0002-spec-and-state/planning/R002 -- docs/slices/0002-spec-and-state/tasks.md (task 06 and task 16 hunks; only task 16 reviewed)
  git log --oneline -12 (e41a589 is the change)
- Requirements: R20, R21, R22
- Interfaces: `Submission` with `BeforeRound int` against task 14's
  `Submission` and the code in internal/slice/submit.go; `round` is
  already computed in `Prepare`, so the field can be set there with no
  new read

## Tests

None.

## Previous findings

None.

## Findings

### F1. Task 16 now changes a shared interface, but its risk stays normal

Level: low
Where: tasks.md → Task 16 → yaml (risk), Interfaces
Breaks: nothing
Why: docs/templates/tasks.md says a task that "changes a contract or a
  shared interface" is `core`. Task 16 now changes `slice.Submission`,
  part of the Core API that task 14 (risk core) created. build-slice
  uses risk only to decide whether a task is code-reviewed, and both
  `core` and `normal` are, so the label changes nothing in practice.
  The change is small: one added field, used only by `spec submit`.
Fix: set `risk: core`, or leave it.

### F2. The unit check of BeforeRound has no test name

Level: low
Where: tasks.md → Task 16 → Done when (`TestSubmitRefusedReviewed`)
Breaks: nothing
Why: the line says "In `internal/slice/submit_test.go`, the refused
  submit from `review_passed` also checks `BeforeRound` 1", but names no
  test. The check fits in the existing `TestSubmitStates` subtest
  "refused keeps review" (submit_test.go:240-251). A code reviewer lists
  Done when's tests by name, so an unnamed check is easy to miss.
Fix: name it: `TestSubmitStates` ("refused keeps review") also checks
  `BeforeRound` 1.

### F3. The Steps do not name the change to internal/slice/submit.go

Level: low
Where: tasks.md → Task 16 → Steps
Breaks: nothing
Why: Steps 2 and 3 name only `internal/cli/submit.go` and
  `internal/cli/review.go`. The `BeforeRound` change is only in Files and
  Interfaces. A worker can still find it, since Step 1 writes the tests
  first and Files names the file.
Fix: add to Step 2: "`internal/slice/submit.go`: set `BeforeRound` in
  `Prepare`; then print it, with no `slice.Status` call."

## Requirements check

None.

## Needs human decision

None.
