---
review: change-8
slice: 0001-slice-and-events
kind: change
target: docs/slices/0001-slice-and-events/plan.md, docs/slices/0001-slice-and-events/tasks.md
round: 1
reviewer: doc-reviewer (claude opus)
---

# Review

## Summary

The change follows the human's answers: clean up after every failed
`git worktree add` (task-09-2/Q1), and add `--quiet` (task-09-2/F2). It
agrees with R18, Flow 1, Flow 3, Interrupts, and the rest of Undo on
failure, and the new test is real and can reach the case. One medium
gap: the cleanup matches a worktree entry by its path alone, so a stale
entry that another worktree left at that path now gets removed every
time, because that entry is exactly what makes `git worktree add` fail
(F1).

## Checked

- Files read: the diff of plan.md and tasks.md, reviews/task-09-2.md,
  reviews/change-6.md, reviews/change-7.md, spec.md (R8, R11, R15, R17,
  R18, S12), plan.md (Approach, Flow 1, Flow 2, Flow 3, Undo on failure,
  Interrupts, Git arguments), tasks.md (Task 09),
  docs/templates/tasks.md (Done when), docs/templates/review.md,
  .agents/prompts/doc-reviewer.md
- Ran:
  git 2.43.0 in a temporary repo with a `post-checkout` hook that only runs `exit 3`: `git branch slice/spec-submit main`, then `git worktree add --quiet .sliceflow/worktrees/spec-submit slice/spec-submit` (exit 3, stderr empty, the worktree stays registered on `refs/heads/slice/spec-submit`)
  then `git worktree remove --force --force` on it (exit 0), `git branch -D slice/spec-submit` (deleted), `.sliceflow/` holds only an empty `worktrees/`
  `git worktree add -q --detach .sliceflow/worktrees/x main`, then `rm -rf` of that folder (porcelain shows the entry, `detached`, `prunable`)
  then `git branch slice/x main` (exit 0) and `git worktree add --quiet .sliceflow/worktrees/x slice/x` (exit 128: `fatal: '.sliceflow/worktrees/x' is a missing but already registered worktree;`)
  a worktree on `slice/y` whose folder was deleted, then `git branch -D slice/y` (refused: used by worktree)
- Requirements: R8, R17, R18

## Tests

None.

## Previous findings

None.

## Findings

### F1. A failed `git worktree add` now always removes a stale entry another worktree left at the path

Level: medium
Where: plan.md → Shared rules → Undo on failure → the exception
  paragraph, first bullet
Breaks: R18 ("remove what it made"); plan.md → Undo on failure ("It
  never removes a thing that is not on the list")
Why: The first bullet removes any entry that
  `git worktree list --porcelain` shows at `.sliceflow/worktrees/<name>`,
  with `--force --force`, and "matches the path alone" (from
  change-7/F1). Being the only owner of the branch proves that a
  worktree on the branch is this run's. It does not prove that an entry
  at the path is this run's. R17 checks only the folder (`os.Lstat`), not
  git's entries. Input: a worktree at `.sliceflow/worktrees/x` on a
  detached HEAD, e.g. after the user ran `git checkout --detach` in it
  and later deleted its folder by hand. Its entry stays, and no branch
  holds it, so R15 and R17 pass. I ran this with git 2.43:
  `git worktree add --quiet .sliceflow/worktrees/x slice/x` fails with
  exit 128, `is a missing but already registered worktree`. Before this
  change, only a signal in that moment took the exception. Now the
  failure itself takes it, so every `create x` removes that entry. That
  deletes git's folder for that worktree, with its HEAD and HEAD
  reflog, so a commit that only that detached HEAD reached is no longer
  reachable. `--force --force` also removes the entry if the user locked
  it, which is what a lock is meant to stop. The message still says
  `Everything sliceflow made was removed.`. (A missing worktree on a
  branch is no risk: git refuses `git branch -D` on its branch, so the
  branch still exists and R15 stops first.)
Fix: before step 6, note whether `git worktree list --porcelain` already
  shows an entry at that path. If it did, skip the first bullet. Then
  git's error is the reason, its hint follows indented, and the stale
  entry stays for the user to prune. A unit test could make that
  stale, detached entry and check that it is still listed after the
  error.

## Requirements check

None.

## Needs human decision

None.
