---
review: change-6
slice: 0001-slice-and-events
kind: change
target: docs/slices/0001-slice-and-events/plan.md, docs/slices/0001-slice-and-events/tasks.md
round: 1
reviewer: doc-reviewer (claude opus)
---

# Review

## Summary

The new paragraph follows the human's answer (kill git at once, then
clean up what is left), agrees with R18's message, Flow 3, and
Interrupts, and handles a branch left with no worktree, because each
bullet stands alone. The main problem: when two `create` runs use the
same name, its ownership test also matches the other run's worktree, so
it can remove that worktree and then its branch (F1). Also, the paragraph
does not say how the folder is removed (F2), or where the cleanup runs
in the undo order (F3).

## Checked

- Files read: the diff of plan.md and tasks.md,
  docs/slices/0001-slice-and-events/reviews/task-09-1.md, spec.md
  (R11, R15, R17, R18, S9, S12), plan.md (Flow 1, Flow 2, Flow 3,
  Cross-platform paths, Undo on failure, Interrupts, Git arguments),
  tasks.md (Task 09), docs/templates/review.md,
  .agents/prompts/doc-reviewer.md, internal/git/git.go
- Ran:
  git 2.43.0 in a temporary repo, `git worktree add -b` with a `post-checkout` hook that runs `git worktree list --porcelain` (the hook runs when the worktree is fully registered, on `refs/heads/slice/a`, and no longer locked)
  `git worktree remove --force` on a registered worktree whose folder was deleted (exit 0, entry gone), then `git branch -D slice/a` (deleted)
- Requirements: R15, R17, R18

## Tests

None.

## Previous findings

None.

## Findings

### F1. Another run's worktree matches the ownership test, and removing it leaves its branch open to removal

Level: medium
Where: plan.md → Shared rules → Undo on failure → the new exception
  paragraph
Breaks: R18; plan.md → Undo on failure ("the one that loses never
  removes the other's branch")
Why: The paragraph says what is "left" points at this run because "R15
  proved it did not exist before" and "R17 proved it was not there
  before". Those checks ran at the start. The paragraph just above
  plans for two `create` runs with the same name at once, so between
  the checks and `git worktree add`, run B can make `slice/<name>` and
  `.sliceflow/worktrees/<name>`. Input: B's `worktree add` wins and is
  in its `post-checkout` hook, or B is done. A's `worktree add -b` then
  fails at once ("a branch named ... already exists"), and a Ctrl-C
  reaches A in that moment. `create` sees `ctx.Err()`, so it takes the
  exception. Bullet 1 finds B's worktree at exactly
  `.sliceflow/worktrees/<name>` and runs `git worktree remove --force`
  on it. Bullet 3's protection ("`git branch -D` refuses a branch that
  another worktree has checked out") is gone now, because A just removed
  that worktree. If B has not committed yet, its branch still points at
  the start commit and is deleted too. B's worktree, branch, and
  folder all look the same as A's own leftovers (I checked the state
  during the hook: registered, unlocked, on `refs/heads/slice/<name>`
  at the start commit). So no check on the worktree list alone can tell
  them apart.
Fix: make ownership something only this run can have. One way, with no
  new git behavior: split step 6 into `git branch slice/<name> <commit>`
  and then `git worktree add <worktree> slice/<name>`. The branch goes on
  the list when `git branch` succeeds. Git refuses a second checkout of
  that branch, and B's `git branch` fails, so a worktree that
  `git worktree list --porcelain` shows on `refs/heads/slice/<name>`,
  and the folder, can only be this run's. R18's reasons stay
  (`git worktree add failed (exit code 128)` is still possible). Another
  way: `git worktree add --lock --reason <a token for this run>`, remove
  only an entry locked with that token, and run `git worktree unlock`
  after it succeeds.

### F2. The paragraph does not say how the folder is removed, and the rule above says "a folder only if it is empty"

Level: medium
Where: plan.md → Shared rules → Undo on failure → the new exception
  paragraph, second bullet
Breaks: R18
Why: The list rule above says files and folders are removed "with an OS
  call, a folder only if it is empty". The new bullet says only "the
  folder ..., if it still exists". After git is killed in the middle of
  the checkout, that folder holds part of the checkout and a `.git`
  file, and the worktree may not be registered yet (git writes the
  folder and `.git/worktrees/<id>` in steps). A worker who follows the
  rule above leaves it. The message then ends with
  `Could not remove: .sliceflow/worktrees/<name>.`, and the next
  `create` with that name fails with R17. A worker who uses
  `os.RemoveAll` follows a rule the plan does not write.
Fix: say it: the folder is removed with `os.RemoveAll`, after the
  worktree bullet, and only when F1's ownership test holds. Then run
  `git worktree prune`, or say why an entry left in `.git/worktrees/`
  is harmless.

### F3. Where the cleanup runs in the undo order, and `.sliceflow/worktrees/`, are not said

Level: medium
Where: plan.md → Shared rules → Undo on failure → the new exception
  paragraph, and item 2 of the list
Breaks: R18
Why: The list is removed "newest first", and `.sliceflow/` is removed
  only if it is empty. The paragraph does not say whether the cleanup
  runs before or after the list's items. Input: a new repo with no
  `.sliceflow/`, and a Ctrl-C during `worktree add`. If the list runs
  first, `.sliceflow/` is not empty yet, so it stays, and the message
  wrongly ends with `Could not remove: .sliceflow.`. Also, item 2 adds
  `.sliceflow/worktrees/` to the list "after step 6 when it did not"
  exist. If step 6 is interrupted, it is not clear if that folder goes
  on the list. If not, an empty `.sliceflow/worktrees/` (and so
  `.sliceflow/`) stays, while the message says
  `Everything sliceflow made was removed.`
Fix: say the cleanup runs first, as the newest item, and that
  `.sliceflow/worktrees/` goes on the list when it did not exist before
  step 6, whether step 6 succeeded or not.

### F4. The lock reason `initializing` is in the user's language

Level: low
Where: plan.md → Shared rules → Undo on failure → the new exception
  paragraph, first bullet
Breaks: nothing
Why: Git writes that lock reason through its translations
  (`_("initializing")` in git's `builtin/worktree.c`), and `git.Run`
  does not set `LC_ALL` (internal/git/git.go:51). In a German setup,
  the test for `initializing` fails, the single `--force` is refused,
  and the worktree lands in `Could not remove:`. I could not run this,
  because no git translations are installed here. Once a worktree is
  known to be this run's (F1), the user cannot have locked it, so the
  condition adds nothing.
Fix: always use `git worktree remove --force --force` for this
  worktree, and drop the lock reason test.

### F5. The new test reaches only one of the states git can leave

Level: low
Where: tasks.md → Task 09 → Done when → `TestUndoInterruptedWorktreeAdd`
Breaks: nothing
Why: The test is real and can reach the case: I checked that git runs
  `post-checkout` from `git worktree add`, after the worktree is
  registered, so killing git there leaves a branch, a registered
  unlocked worktree, and the folder, and `worktree add` has not
  succeeded. It does not reach a worktree locked `initializing`, a
  folder with no registered worktree, or a branch alone, which the
  bullets also handle. Those need a kill during the checkout, which
  cannot be timed in a test, so this is acceptable. Also, it checks
  only "the `interrupted` error", not the last sentence. If the code
  removes things but still lists one in `Could not remove:`, the test
  passes.
Fix: check the whole text,
  `could not create slice spec-submit: interrupted. Everything sliceflow made was removed.`
  For F1's case, a unit test where `slice/spec-submit` and its worktree
  exist from another run would prove they stay.

## Requirements check

None.

## Needs human decision

None.
