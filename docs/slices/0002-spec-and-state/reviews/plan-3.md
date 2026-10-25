---
review: plan-3
slice: 0002-spec-and-state
kind: change
target: docs/slices/0002-spec-and-state/plan.md
round: 3
reviewer: doc-reviewer (claude opus)
---

# Review

## Summary

The fixes since plan review 2 work. `Prepare` now returns the commit
message with the data and the files, so a submit picks its R21 message
under the lock, after R20's state check. Tasks 12 and 14 agree with the
new `saveOp` and `prepared` types. All five findings are verified. One
new low finding (F1): the fix to task 07's Out of scope line breaks the
template's `Done in:` form. Nothing needs the human.

## Checked

- Files read: docs/slices/0002-spec-and-state/plan.md (Approach, Flows
  1 and 3 to 6, Shared rules, Core API, Testing), tasks.md (tasks 06,
  07, and 11 to 16), spec.md (R21, R27, R30), status.md,
  reviews/plan-1.md (F1, F3, F7, F10), reviews/plan-2.md;
  docs/templates/tasks.md (the brief size, the markers, Out of scope);
  docs/templates/review.md; .agents/prompts/doc-reviewer.md
- Ran:
  git diff 3b879ba c052ea9 on plan.md, tasks.md, preview.md, spec.md (plan.md and tasks.md changed; preview.md and spec.md did not)
  grep for `Write`, `Message`, `saveOp.`, and `Data any` in plan.md and tasks.md (only the new `prepared` type and `spec.Problem`; no leftover field)
  grep for every `Done in:` line in tasks.md (one line not in the template's form, see F1)
  grep for `Done in` in internal/ and docs/design/ (no reader yet)
- Requirements: R6, R10, R19, R20, R21, R24, R27, R30, R32, R33
- Interfaces: `prepared`, `saveOp`, `saveEvent` (task 11) against
  their callers `SetTitle` (task 12) and `Submit` (task 14), with Flow 3
  step 4, Flow 4 steps 3, 4, and 7, Flow 5, and Flow 6 step 4. Each
  save's message is fixed by its `Prepare`: `title set`, the two R21
  messages, `review <n> started`, and R27's three. Flow 5's skip of
  `spec.submitted` fits too: a `Prepare` that finds `submitted` can
  return an error that `Submit` knows, and `TestSavePrepareFails` says
  `saveEvent` returns it as it is. `specChanged`'s `(bool, error)`
  against the new `TestSpecChanged` case.

## Tests

None.

## Previous findings

- plan-1/F1: verified (fixed)
- plan-1/F3: verified (fixed)
- plan-1/F7: verified (fixed)
- plan-1/F10: verified (fixed)
- plan-2/F1: verified (fixed)

## Findings

### F1. Task 07's Done in: line is not in the template's form

Level: low
Where: tasks.md → Task 07 → Out of scope
Breaks: nothing
Why: The line is now `Done in: task 15 (S13 to S17), task 16 (S23)`.
  docs/templates/tasks.md says sliceflow reads the `Done in:` lines,
  and the value is one or more tasks, e.g. `task 02, task 04`. The
  scenario numbers in brackets are not in that form. Every other
  `Done in:` line in tasks.md keeps it.
Fix: move the scenarios to the text, as before:
  `The status after a real submit (S13 to S17, S23). Done in: task 15, task 16`

## Requirements check

None.

## Needs human decision

None.
