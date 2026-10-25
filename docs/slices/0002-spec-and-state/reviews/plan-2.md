---
review: plan-2
slice: 0002-spec-and-state
kind: plan
target: docs/slices/0002-spec-and-state/plan.md
round: 2
reviewer: doc-reviewer (claude opus)
---

# Review

## Summary

Most round-1 findings are fixed. The Prepare step puts submit's checks
under the lock in R20's order, and R35 now says what the human chose.
Four are reopened. The main one is F1: `saveOp.Message` is still fixed
before the lock, but a submit's commit message depends on the result
that Prepare finds. The other three are small: F3, F7, and F10. One new
low finding (F1) lists small leftovers of the fixes. Nothing needs the
human.

## Checked

- Files read: docs/slices/0002-spec-and-state/spec.md, plan.md,
  tasks.md, preview.md, status.md, reviews/plan-1.md; the diff
  3ccb8d3..3b879ba of docs/slices/0002-spec-and-state/;
  docs/templates/tasks.md, plan.md (touches and optional sections),
  review.md; AGENTS.md; .agents/prompts/doc-reviewer.md;
  internal/cli/cli.go (`register`, `topUsage`, `Run`), the imports of
  every file in internal/cli; the file lists of internal/cli,
  internal/slice, internal/e2e
- Ran:
  a Python script that estimates each task's brief from its links (14: about 3,000 words, 15: 3,050, 16: 3,050; 06: 2,500, 07: 2,050, 09: 2,350, 11: 2,900; the others 1,150 to 1,800; every follows name found)
  grep for the stale names `Write` and `S23). Done in` in tasks.md (one hit each, see F1)
- Requirements: R1 to R37, with R35's new sentence, and each task's
  covers against its Proves: and Scenario: lines (tasks 14, 15, 16 in
  full)
- Interfaces: `saveOp`, `saveEvent`, `state.Allowed`, `specChanged`,
  `currentSliceName`, `openSlice`, `Submit`, `review.Request`,
  `template.Raw`, `cli.register`, `topUsage`; the status.md task table
  against tasks.md (ids, risk, covers, depends_on all match)

## Tests

None.

## Previous findings

- plan-1/Q1: verified (fixed)
- plan-1/F1: reopened. Prepare fixes the order and the errors, but
  `saveOp.Message` is still set before the lock (tasks.md → Task 11 →
  Interfaces). For `spec.submitted`, the message is
  `<name>: spec submitted` or `<name>: spec submit refused` (R21). Which
  one depends on the strict check, and that check now runs inside
  Prepare, after R20's state check (R20: "then checks this, then R19").
  Prepare returns only `(any, map[string][]byte, error)`, so task 14
  cannot give the right message. Its only way out is to check `spec.md`
  before `saveEvent`. That is round 1's problem again: with a busy lock
  and a deleted `spec.md`, the user gets R19's error at once, not R7's
  after 10 s. Fix: let Prepare also return the message, e.g. as a small
  struct `{Data any; Files map[string][]byte; Message string}`, and keep
  `Message` in `saveOp` only as the default.
- plan-1/F2: verified (fixed)
- plan-1/F3: reopened. Task 15 was split, but tasks 15 and 16 are each
  about 3,000 words by my estimate, and neither gives a reason. Task 16
  covers R20 and R29, which bring in 9 more scenarios. Task 14's line
  still says 2,700 words. With R12 and R19 added (S9, S12), it is now
  about 3,000. Fix: add the reason line to 15 and 16, as for the
  others, and update 14's number. Or drop R29 from task 16's covers and
  from `TestSubmitStopsHalfway`'s Proves: S23 still reaches the brief
  through R10. That takes out S1, S4, and S16.
- plan-1/F4: verified (fixed)
- plan-1/F5: verified (fixed)
- plan-1/F6: verified (fixed)
- plan-1/F7: reopened. The `seq` pair and the e2e file names now agree.
  But plan.md → Testing says the e2e Ctrl-C test is in task 15. It is
  `TestSubmitInterrupted`, in task 16.
- plan-1/F8: verified (fixed)
- plan-1/F9: verified (fixed)
- plan-1/F10: reopened. "Only `slice`, `git`, and a reviewer" and
  `Request.File` are fixed. The dependency line still leaves out `cli`
  → `template`: task 08's `internal/cli/template.go` prints
  `template.Raw()`.
- plan-1/F11: verified (fixed)
- plan-1/F12: verified (fixed)

## Findings

### F1. Small leftovers of the round-1 fixes

Level: low
Where: tasks.md → Task 11 → Done when (`TestSaveHookFails`), tasks.md →
  Task 07 → Out of scope, tasks.md → Task 06 → Done when
  (`TestSpecChanged`)
Breaks: nothing
Why: `TestSaveHookFails` says "a file in `Write` is gone", but `saveOp`
  has no `Write` field now: the files come from `Prepare`. Task 07's
  Out of scope says S23 is done in task 15; it is now in task 16
  (`TestSubmitStopsHalfway`). Flow 1 step 3's new rule (any other git
  failure, e.g. exit code 128 for a missing reviewed commit, means not
  compared) has no test in task 06. AGENTS.md says every change that
  adds behavior adds a test.
Fix: write "a file that Prepare returned"; write "Done in: task 15,
  task 16"; add a case to `TestSpecChanged`: a commit id that is not in
  the repo gives false and no error.

## Requirements check

None.

## Needs human decision

None.
