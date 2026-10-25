---
review: plan-5
slice: 0002-spec-and-state
kind: change
target: docs/slices/0002-spec-and-state/spec.md
round: 5
reviewer: doc-reviewer (claude opus)
---

# Review

## Summary

The new `spec review` command is in all four documents. Every
requirement and scenario is covered, each of S1 to S36 has an e2e test
with the right number, and every `follows` name exists. All five
plan-4 findings are fixed. One medium gap: since S24 went away, no
brief or test covers R10's `The steps before it were saved` ending for
a failed `review.finished` commit (F1). The rest are leftover old
wording and stale task numbers (F2 to F4). None of them needs the
human.

## Checked

- Files read: docs/slices/0002-spec-and-state/spec.md (all),
  plan.md (all), tasks.md (tasks 01, 07, 08, 09, 11, 13, 14, 15, 16,
  17, and the yaml blocks of all tasks), preview.md (all), status.md,
  reviews/plan-4.md; docs/templates/review.md (via the review file's
  comments); .agents/prompts/doc-reviewer.md
- Ran:
  git diff 7bca891 29ead4e -- docs/slices/0002-spec-and-state/ (spec.md, plan.md, tasks.md, preview.md, status.md changed)
  git diff --stat 29ead4e f418bc1 (only status.md changed after it)
  grep for every `Scenario:` line and yaml `covers`, `depends_on`, and `follows` in tasks.md (S1 to S36 each have one test; every `follows` name is a heading in plan.md)
  grep for `third error`, `Tail`, `steps before`, `to start review`, `task 1[4-7]`, `Task 1[4-7]`, `Only`, and `submit and title` in the four files (see F3 and F4)
- Requirements: R6, R10, R11, R12, R19, R20, R23, R24, R25, R29, R36, R37

## Tests

None.

## Previous findings

- plan-4/F1: verified (fixed)
- plan-4/F2: verified (fixed)
- plan-4/F3: verified (fixed)
- plan-4/F4: verified (fixed)
- plan-4/F5: verified (fixed)

## Findings

### F1. No brief or test covers a failed `review.finished` commit

Level: medium
Where: tasks.md → Task 15 → Interfaces and Done when, tasks.md →
  Task 17 → Done when, spec.md → Story 4
Breaks: R10, R37
Why: R10 says that when an earlier event of the same command was
  already saved, the error ends with
  `The steps before it were saved; spec.md is <state>.` Since this
  change, `spec review` is the only command that saves two events. The
  old S24 tested this path: its hook let `spec submitted` through and
  refused the review. The new S23 refuses every commit with `review`,
  so the first one, `review 1 started`, already fails, and the error
  ends with `Nothing was changed.` Task 15's Interfaces never mention
  `saveOp.Tail`. No test in task 15 or task 17 makes the second save
  fail. Input: a `commit-msg` hook that refuses only messages with
  `passed`, then `sliceflow spec review`. `review.started` is
  committed, and `review.finished` fails. A worker who follows task 15
  as written leaves `Tail` `""`. The user then reads
  `... Nothing was changed.`, although a commit was made and
  `spec.md` is `in_review`. Nothing checks either that
  `Review 1 started: ...` is still on stdout when the command exits
  with 1 (R37). `TestSaveTail` in task 11 tests only the mechanism.
Fix: in task 15's Interfaces, say that the `review.finished` save sets
  `Tail` to `The steps before it were saved; spec.md is in_review.`, and
  add a test for it, e.g. `TestReviewFinishCommitFails`. Add an e2e test
  for it to task 17, with a hook that refuses `passed`: the stdout line,
  the stderr line, and exit 1.

### F2. Three spec lines still describe the old submit

Level: low
Where: spec.md → Out of scope (the 0004 and 0015 items), spec.md →
  SD7 → Not chosen
Breaks: nothing
Why: Out of scope still says "what `spec submit` prints when a review
  does not pass". A review's result is now printed by `spec review`.
  The 0015 item lists `--json` for `spec validate`, `spec submit`,
  `slice title`, and `template show`. It leaves out `spec review`,
  which R36 now names too. SD7 rejects "no review step until 0004,
  because submit would then end in a state that 0004 has to change
  again". But submit now ends at `submitted` either way, so that reason
  no longer holds.
Fix: write `spec review` in the 0004 item, and add it to the 0015
  item. In SD7, give a reason that still holds, e.g. "because a
  submitted spec would then have no step that moves it on".

### F3. The plan still leaves `spec review` out of the lock and Flow 8, and names task 16 for Ctrl-C

Level: low
Where: plan.md → Shared rules → The lock → step 4, plan.md → Flow 8 →
  step 3, plan.md → Approach (the `internal/slice/` row), plan.md →
  Testing ("the e2e test (task 16)")
Breaks: nothing
Why: The lock's step 4 says "Only `spec submit` and `slice title` take
  the lock". Flow 8's step 3 names `spec validate`, `spec submit`, and
  `slice title`, with "(submit and title only)". Both disagree with R11
  and with Flow 5, where `spec review` takes the lock. The Approach row
  names submit, title, validate, and status, but not the review.
  `TestSubmitInterrupted` is now in task 17, not task 16. Flow 5 and R11
  say it right, so a worker would probably not go wrong.
Fix: add `spec review` in the three places, and write task 17 in
  Testing.

### F4. Stale task numbers and error index in tasks.md

Level: low
Where: tasks.md → Task 09 → Interfaces (`printProblems`), Task 11 →
  Interfaces (`saveEvent`) and Out of scope, Task 13 → Interfaces
  (`Reviewer`) and Out of scope, Task 14 → Done when
  (`TestSubmitStates`), Task 15 → Done when (`TestReviewStates`)
Breaks: nothing
Why: Task 09 says "Task 15 uses it" for `printProblems`; it is task 16
  now. Task 11 says "Tasks 12 and 14 use it", and its Out of scope says
  "task 12, task 14". Task 15 calls `saveEvent` too. Task 13 says
  "Task 14 uses it" for `Reviewer`, and "Saving the review events. Done
  in: task 14". Both are task 15 now. Tasks 14 and 15 name "R20's third
  error" for no good `slice.created`. R20 has a new first bullet,
  `submitted`, so the third bullet is now the `in_review` error, and
  the `slice.created` error is the fourth.
Fix: change the numbers, and write "R20's error for no good
  `slice.created`" in place of "R20's third error".

## Requirements check

None.

## Needs human decision

None.
