---
review: plan-4
slice: 0002-spec-and-state
kind: change
target: docs/slices/0002-spec-and-state/spec.md
round: 4
reviewer: doc-reviewer (claude opus)
---

# Review

## Summary

Both changes are in place in all four documents. Every S reference
in spec.md, plan.md, tasks.md, and preview.md points at the right
scenario, except one stale range in task 07's Goal (F5). plan-3/F1 is
fixed. There are four small gaps around the new merged case (F1 to F4).
None of them needs the human.

## Checked

- Files read: docs/slices/0002-spec-and-state/spec.md (S1 to S32,
  R11, R29, R30, R31), plan.md (Flows 1, 2, and 8, Shared rules, Core
  API, Testing, TD1), tasks.md (tasks 06, 07, 08, 09, 13, 14, 15, 16,
  and the S references in tasks 02 and 05), preview.md, status.md,
  reviews/plan-3.md; docs/slices/0001-slice-and-events/spec.md (R24,
  R29, R30, R33, R38); internal/slice/status.go (`readSlice`,
  `currentSliceName`), internal/slice/errors.go, internal/cli/output.go,
  internal/cli/create.go; docs/templates/review.md;
  .agents/prompts/doc-reviewer.md
- Ran:
  git diff 2a9784f c7c3670 -- docs/slices/0002-spec-and-state/ (spec.md, plan.md, tasks.md, preview.md, and status.md changed)
  grep for every scenario heading in spec.md (S1 to S32, in order, none missing)
  grep for every S reference in the four files, each compared by hand with the scenario it names (all right, except task 07's Goal, see F5)
  grep for `specNext`, `Merged`, `SliceStatus`, `openSlice`, and `findWorktree` in plan.md and tasks.md
  grep for `worktrees` and `Error: ` in internal/ (0001 makes worktrees at `.sliceflow/worktrees/<name>` from the main checkout's root; an `*Error` may hold more than one line)
- Requirements: R11, R29, R30, R31

## Tests

None.

## Previous findings

- plan-3/F1: verified (fixed)

## Findings

### F1. R31 does not say what `next` is in the JSON for a merged slice

Level: low
Where: spec.md → R31, tasks.md → Task 07 → Done when
Breaks: nothing
Why: R29's new row gives `Next:` `-` for a merged slice. R31 says
  `next` is "the text of R29's `Next:`", and is left out only with no
  good `slice.created`. Read word for word, a merged slice gives
  `"next":"-"`. But the other place where `Next:` is `-` leaves `next`
  out, and 0001's R30 leaves out the values that print as `-`. A
  worker has to guess which one. No test covers it:
  `TestStatusJSONReview` has no merged case.
Fix: say it in R31, e.g. "For a merged slice, `next` is left out", and
  add the case to `TestStatusJSONReview`.

### F2. R29 does not say which rule wins for a merged slice with a refused submit

Level: low
Where: spec.md → R29 (the merged row, and the paragraph under the table)
Breaks: nothing
Why: The merged row says `Next:` is `-`. The paragraph under the table
  says that when the last spec event is a refused `spec.submitted` and
  `spec.md` is `draft`, `Next:` shows
  `fix spec.md, then: sliceflow spec submit`. A slice merged with a
  refused submit as its last spec event fits both. Nothing in this
  slice stops such a merge. `specNext(st, changed, merged)` has to pick
  an order, and `TestSpecNext` does not say which.
Fix: in the paragraph, say "and the slice is not merged". `Spec:` keeps
  its refused mark either way.

### F3. Task 09 does not say how `openSlice` finds a merged slice, and the merged error has no e2e test

Level: low
Where: tasks.md → Task 09 → Interfaces and Done when, plan.md → Flow 2
  step 2
Breaks: nothing
Why: The merged error needs to know two things: that the log was read
  from the base branch, and the base branch's name. Task 06 adds
  `SliceStatus.Merged`, but task 09's `uses` line names only
  `currentSliceName`, `findWorktree`, `spec.Parse`, and `spec.Check`.
  It does not name `Merged`, or where `<base branch>` comes from (the
  base of R21 of 0001, not `From`). In Flow 2 step 2, the sentence just
  before is about worktree entries, so "a slice with no such branch"
  can be read as "no worktree entry". That reading gives S12's case the
  merged error. R11 is `Check: e2e`, but its merged error appears only
  in `TestOpenSliceOrder`. That test is written without `Error: `, so
  it reads as a unit test, and there is no scenario.
Fix: in task 09's Interfaces, add `SliceStatus.Merged` (task 06) and
  the base branch name as used; in Flow 2, write "a slice whose branch
  `slice/<name>` does not exist (its log is read from the base
  branch)"; in `TestValidateNoWorktree` or a new e2e test, add the
  merged case with its full stderr line.

### F4. The `Make one:` command works only from the main checkout's root

Level: low
Where: spec.md → R11, S12
Breaks: nothing
Why: `git worktree add <path>` reads a relative path from the current
  folder. The hint's `.sliceflow/worktrees/<name>` is relative to the
  main checkout's root, as in 0001's other messages and in S13. If
  someone copies the command and runs it from another slice's worktree,
  e.g. `.sliceflow/worktrees/other`, git makes a worktree at
  `.sliceflow/worktrees/other/.sliceflow/worktrees/login-google`,
  inside the other worktree. The same thing happens in a subfolder of
  the main checkout, e.g. `docs/`. S12 runs on `main` at the root, so it
  is right as written. The human chose the line's text, so any change
  to that text is theirs to make.
Fix: leave the text, or ask the human about e.g.
  `Make one, in the repo's root: git worktree add ...`.

### F5. Task 07's Goal still names S1 to S4

Level: low
Where: tasks.md → Task 07 → Goal
Breaks: nothing
Why: The Goal says the brief is long "because S1 to S4 hold the exact
  output". Task 07 now also has S5 (`TestStatusMerged`).
Fix: `S1 to S5`.

## Requirements check

None.

## Needs human decision

None.
