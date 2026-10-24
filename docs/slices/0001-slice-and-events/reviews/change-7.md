---
review: change-7
slice: 0001-slice-and-events
kind: change
target: docs/slices/0001-slice-and-events/plan.md, docs/slices/0001-slice-and-events/tasks.md
round: 2
reviewer: doc-reviewer (claude opus)
---

# Review

## Summary

All five change-6 findings are fixed. The split into `git branch`, then
`git worktree add`, makes ownership something only this run can have.
The change still follows the human's option 2, and agrees with R4, R5,
R15, R17, R18, Flow 1, Flow 3, and Interrupts. One medium gap: the
cleanup matches the worktree by its branch, but git first writes an
empty HEAD. A stop in that moment leaves git's locked entry, and the
next `create` with that name fails (F1).

## Checked

- Files read: the diff of plan.md and tasks.md,
  docs/slices/0001-slice-and-events/reviews/change-6.md, status.md
  (Findings → Doc reviews), spec.md (R4, R5, R11, R15, R17, R18, S12),
  plan.md (Approach, Flow 1, Flow 2, Flow 3, Undo on failure,
  Interrupts, Git arguments), tasks.md (Task 06, Task 09), preview.md
  (P5), docs/templates/review.md, .agents/prompts/doc-reviewer.md,
  internal/slice/create.go
- Ran:
  git 2.43.0 in a temporary repo: `git branch slice/a <commit>` twice (exit 0, then exit 128 with `fatal: a branch named 'slice/a' already exists`)
  `git worktree add .sliceflow/worktrees/a slice/a` with a `post-checkout` hook (the hook runs, and the worktree is on `refs/heads/slice/a`)
  a registered worktree with HEAD set to the null id and `locked initializing`, as git leaves it before its `symbolic-ref` step: `git worktree list --porcelain` shows `detached` and `locked initializing`, no `branch` line
  `rm -rf` of that folder, then `git worktree prune` (the locked entry stays), `git branch -D` (works), then `git branch` and `git worktree add` at the same path again (exit 128: `is a missing but locked worktree`)
  `git worktree remove --force --force` on that missing, locked entry (exit 0, entry gone)
  `git worktree add <path> refs/heads/slice/c` (makes a worktree with a detached HEAD, not on the branch)
- Requirements: R4, R5, R15, R17, R18

## Tests

None.

## Previous findings

- change-6/F1: verified (fixed)
- change-6/F2: verified (fixed)
- change-6/F3: verified (fixed)
- change-6/F4: verified (fixed)
- change-6/F5: verified (fixed)

## Findings

### F1. The cleanup misses git's entry when the stop comes before git set the worktree's HEAD

Level: medium
Where: plan.md → Shared rules → Undo on failure → the exception
  paragraph, first bullet
Breaks: R18
Why: The first bullet removes a worktree at `.sliceflow/worktrees/<name>`
  only when `git worktree list --porcelain` shows it "on the branch
  `slice/<name>`". `git worktree add` first locks its entry
  (`initializing`), writes the null id as HEAD, and only then runs
  `symbolic-ref` to put HEAD on the branch. I made that state by hand:
  porcelain shows `HEAD 0000...`, `detached`, `locked initializing`, with
  no `branch` line. So a stop in that moment skips the first bullet. The
  second bullet removes the folder, and the list removes the branch.
  git's locked entry stays, and `git worktree prune` does not clear a
  locked entry. Input: Ctrl-C in that moment, then `create` again with
  the same name. Result: the first run prints
  `Everything sliceflow made was removed.`, and the second fails with
  `fatal: '.sliceflow/worktrees/<name>' is a missing but locked worktree`
  instead of creating the slice. The branch test adds no safety. The
  second bullet already removes the folder on the grounds that only this
  run's `git worktree add` can have made it, and the same holds for
  git's entry at that path.
Fix: match the worktree by its path alone, whatever its HEAD. I checked
  that `git worktree remove --force --force` clears a missing, locked
  entry, so the order of the two bullets can stay.

### F2. "After step 6" puts `.sliceflow/worktrees/` after the worktree in the list

Level: low
Where: plan.md → Shared rules → Undo on failure → list item 2
Breaks: nothing
Why: Item 2 adds `.sliceflow/worktrees/` "after step 6". Step 6 now adds
  the branch and then the worktree, so read word for word, the folder
  comes after the worktree. Newest first, the undo would then try the
  folder while the worktree is still in it, and a folder is removed
  "only if it is empty". create.go does it the right way today (its
  comment: the folder "goes on the list before the worktree"), but task
  09 rewrites this part for the split.
Fix: say the folder goes on the list after `git worktree add` returns,
  before the worktree's entry.

### F3. `git worktree add` needs the short branch name

Level: low
Where: plan.md → Flow 1 → step 6, plan.md → Shared rules → Git arguments
Breaks: nothing
Why: Git arguments says base branches are passed as
  `refs/heads/<branch>`. A worker could do the same for the slice branch
  in `git worktree add`. I ran `git worktree add <path> refs/heads/slice/c`:
  git makes a detached worktree, so the commit in step 8 would not land
  on `slice/<name>`. Step 6 writes `slice/<name>`, which is right, but it
  does not say the short name is needed. R2 already keeps the name from
  starting with `-`.
Fix: one line in step 6 or Git arguments: pass `slice/<name>`, not
  `refs/heads/slice/<name>`, to `git worktree add`, or git detaches HEAD.

### F4. A stop during `git branch` itself is not covered

Level: low
Where: plan.md → Shared rules → Undo on failure → the last paragraph of
  the exception
Breaks: nothing
Why: The paragraph covers a `git branch` that fails, e.g. because
  another run made the branch first. When the signal stops `git branch`,
  git may already have written the branch, or left
  `refs/heads/slice/<name>.lock`. The branch is not on the list, and this
  run cannot tell it from another run's branch. So it stays, while the
  message says `Everything sliceflow made was removed.` The window is a
  few microseconds, so this is rare. Before the split, `worktree add -b`
  had the same window.
Fix: name the case and accept it, as task 09's Out of scope does for a
  hard kill.

## Requirements check

None.

## Needs human decision

None.
