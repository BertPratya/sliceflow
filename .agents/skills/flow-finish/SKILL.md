---
name: flow-finish
description: Takes a built slice to main (update from main, integration review, preview with the human, the human's approval of the result, and the merge). Use when every task of a slice is done or cancelled, or the phase is integration or preview.
---

# flow-finish

Every task is `done` or `cancelled`. Set the phase to `integration`.

## Part 1: integration

### 1. Update from main

Each task passed on its own. The slice can still break once it meets what
others merged into `main` meanwhile: two changes can merge cleanly as text
and still not work together. So test the merged result, not the branch
alone.

1. Create branch `update/<name>` from `slice/<name>`, worktree
   `.worktrees/update-<name>/`.
2. There, run `git merge main`. On a conflict, start a worker
   (`.agents/prompts/worker.md`) to resolve it, keeping the changes from
   both sides. A conflict in a project file from the plan's `touches` is
   yours: resolve it, and tell the human what you chose.
3. There, run the setup commands and every check in `AGENTS.md`, task and
   slice checks. On a failure:
   - the merge brought changes in: a worker fixes it, then check again.
   - the merge brought nothing in: the slice itself is broken. Remove the
     update worktree and branch, and add a fix task (`flow-implementation`).
4. In the slice worktree: `git merge --ff-only update/<name>`. Then remove
   the update worktree and branch.

### 2. The integration review

Start a code reviewer (`.agents/prompts/code-reviewer.md`) for an
integration review, in a fresh context. It reads the spec, plan, tasks,
the whole slice diff against `main`, and every deferred finding.

- **high or medium finding:** becomes a fix task at the end of
  `tasks.md`. Set the phase back to `implementing`, build the fix tasks
  with `flow-implementation`, then come back to step 1.
- **Needs human decision:** ask the human, as in `flow-implementation`
  → Decisions.
- **pass:** set the phase to `preview`.

After the review round limit in `AGENTS.md`, the slice is `blocked`: ask
the human what to do.

## Part 2: preview

The tests prove the code does what the tasks say. The preview lets the
human see it does what they meant.

### 1. Hand it to the human

Run the setup commands in the slice worktree. Then tell the human, in
plain words:

1. what the slice does now, in two or three sentences
2. how to run it (`AGENTS.md` → Run locally), in the slice worktree, with
   full paths
3. the checks in `preview.md`, one by one
4. a few minutes to try it freely: one or two ideas of what to poke at,
   e.g. "try a name with spaces". Scripted checks miss what a real user
   does.
5. every decision you made alone while building (from `status.md`), so
   the human can disagree now
6. what was parked for later: the plan's Parked items, and findings
   deferred out of the slice

### 2. Record the results

Write each check's result in `status.md` → Preview: `pass`, or `fail`
with why.

- **A failed check:** a fix task. Back to `flow-implementation`, then
  Part 1 again, then the failed checks again.
- **A decision the human now disagrees with:** handle it as in
  `flow-implementation` → Decisions, kind 2.
- **Something new the human wants:** not in this slice. Offer the backlog
  or a later slice. If it must be in this slice, it is a spec change: go
  back to planning.

## Part 3: finish

### 1. Done means all of this is true

Check each one before you ask for approval. One that is false means the
slice is not done.

- Every task is `done` or `cancelled`.
- The integration review passed, and every requirement in its
  Requirements check is ok.
- No high finding is open. Every deferred finding names a later task that
  fixed it, or is deferred out of the slice.
- Every preview check is `pass`.
- The documents match the latest approval tag `flow/<name>/planning/R00N`,
  apart from fix tasks and wider `touches` in `tasks.md`.
- The checks passed on the updated slice branch.

### 2. The human approves the result

Ask the human to approve the result. Only the human approves. Never
approve for them, even if asked to "just continue". Only a clear
"approve" counts.

When they approve:

1. Add a Log line in `status.md` with their words.
2. Commit, and make an annotated tag on that commit,
   `flow/<name>/result/R001`. Its message quotes the human's approval.
   Write the message to a file and use `git tag -a -F <file>`.

### 3. Update from main again

`main` may have moved since the integration. Repeat Part 1, step 1.

- Nothing came in: go on.
- Something came in, and the checks pass: tell the human what came in.
  If it touched the same files as the slice, run the integration review
  again on the update, and ask the human whether they want to preview
  again. A new result approval gets the next tag, `R002`.

### 4. The backlog

Copy into `backlog.md` at the project root, in the form
`docs/templates/backlog.md` gives:

- the plan's Parked items
- the spec's Out of scope items marked `Goes to: the backlog`
- every finding deferred out of the slice, by its title

### 5. Merge into main

1. In `status.md`, set the phase to `done`. Commit.
2. In the main checkout: check it is on `main` with no uncommitted
   changes. If not, stop and ask the human: it is their checkout.
3. Merge with a merge commit: `git merge --no-ff slice/<name>`, with a
   message naming the slice and what it does.
4. Push as the git policy in `AGENTS.md` says. The tags go with it.
5. Remove the slice worktree, and delete the branch `slice/<name>`: it is
   merged, and the tags keep its approvals.

Slices merge into `main` one at a time. Another slice that is open gets
updated from `main` before its own merge (step 3).

Tell the human the slice is merged: the merge commit, and what went to
the backlog.
