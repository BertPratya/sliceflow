---
name: finish-slice
description: Takes a built slice to main: update from main, the integration review, the preview with the human, the human's approval of the result, the merge, and the push. Use when every task of a slice is done or cancelled, or the phase is integration or preview.
---

# finish-slice

You are the lead. Every task is `done` or `cancelled`. Set the phase to
`integration`.

Each task passed on its own. Three things are still open: do the tasks
work together, does the slice still work with the latest `main`, and is
it what the human meant. This skill answers them, in that order, then
merges.

Work in the slice worktree, `.worktrees/<name>/`, with full paths into
it, until the merge.

## Keeping status.md

After every step that changes something, update the slice's `status.md`
as `docs/templates/status.md` says, then commit in the worktree:

- the header: `Phase`, and `Updated` (the time now, UTC)
- the section the step changed
- Needs human: add what now waits on the human; remove what no longer does
- a Log line: time · who · what happened

## Fix tasks

A finding or a failed check that needs code becomes a fix task: add it at
the end of `tasks.md` with the next free id, and a row in the Tasks
table. Set the phase to `implementing`, build it with `build-slice`, then
come back to Part 1, step 1.

## Part 1: integration

### 1. Update from main

Two changes can merge cleanly as text and still not work together. So
test the merged result.

1. In the slice worktree: `git merge main`.
   - Nothing came in: go on to step 2.
   - A conflict: resolve it, keeping the changes from both sides. Tell
     the human what you chose. A conflict you cannot resolve without
     changing behavior: `git merge --abort`, and ask the human.
2. Do a full run of the checks (`AGENTS.md`).
   - A failure: a fix task.

### 2. The integration review

1. Copy `docs/templates/review.md` to
   `docs/slices/<name>/reviews/integration-<round>.md`, and fill its
   front matter. Round counts integration reviews, from 1.
2. Start a code reviewer in a fresh context: Agent tool, type
   `sliceflow-code-reviewer`, model `opus`. Its prompt: read
   `.agents/prompts/code-reviewer.md` and follow it, for an integration
   review, with the slice worktree's full path and the review file.
3. Write each finding and its state in `status.md` → Findings →
   Integration review.
   - **high or medium:** a fix task.
   - **low:** fix it as a fix task, or defer it to the backlog, and list
     what you did in one line.
   - **Needs human decision:** ask the human, as in `build-slice` →
     Decisions while building.
4. Nothing to fix: set the phase to `preview`. Fill `status.md` →
   Preview: one line per check in `preview.md`, `not tried`.

The review round limit is in `AGENTS.md`. At the limit, the slice is
`integration (blocked)`: ask the human what to do.

## Part 2: preview

The tests prove the code does what the tasks say. The preview lets the
human see it does what they meant.

### 1. Hand it to the human

In plain words:

1. what the slice does now, in two or three sentences
2. how to run it (`AGENTS.md` → Commands), in the slice worktree, with
   full paths
3. the checks in `preview.md`, one by one
4. one or two ideas to try freely, e.g. "try a name with spaces":
   scripted checks miss what a real user does
5. every line under `status.md` → Lead's decisions, so the human can
   disagree now
6. what went to the backlog from this slice

### 2. Record the results

Write each result in `status.md` → Preview: `pass`, or `fail: why`. A
failed check also goes under Needs human until its fix task is done.

- **A failed check:** a fix task. After Part 1 passes again, the human
  tries the failed checks again.
- **The human disagrees with a lead's decision:** handle it as a
  decision in `build-slice` → Decisions while building.
- **Something new the human wants:** offer the backlog or a later slice.
  If it must be in this slice, it changes the spec: handle it as a
  decision in `build-slice`.

## Part 3: approval and merge

### 1. Check that done is done

Every one of these is true. If one is false, the slice is not done: fix
it first.

- Every task is `done` or `cancelled`.
- The last integration review passed, and every requirement in its
  Requirements check is ok.
- No high finding is open. Every deferred finding names the task that
  fixed it, or is in the backlog.
- Every preview check is `pass`.
- Since the latest approval tag `sliceflow/<name>/planning/R00N`, `tasks.md`
  only gained fix tasks and wider `touches`, and no other document
  changed.
- The checks passed on the slice branch after the last update from
  `main`.

### 2. The human approves the result

In `status.md` → Needs human: `The result is waiting for approval`. Ask
the human to approve it. Say that their approval also means: merge into
`main`, and push. They may say "approve, but don't merge yet".

Only the human approves. Never approve for them, even if asked to "just
continue". Only a clear "approve" counts.

When they approve:

1. In `status.md`: remove the approval from Needs human, and add a Log
   line with their words.
2. Commit, and make an annotated tag on that commit,
   `sliceflow/<name>/result/R001`. Its message quotes the human's approval.
   Write the message to a file and use `git tag -a -F <file>`.

### 3. Update from main once more

`main` may have moved since Part 1. Repeat Part 1, step 1.

- **Nothing came in:** go on.
- **Something came in, and the checks pass with no fix:** tell the human
  what came in, and go on.
- **A fix was needed:** build it as a fix task, and run the integration
  review on it. If the fix changes what users see, the human tries the
  preview checks it touches and approves again: the next tag, `R002`.
  Otherwise tell the human what changed, and go on.

### 4. Merge and push

1. In `status.md`: set the phase to `done`, and add a Log line. Commit.
2. In the main checkout: check it is on `main` with no uncommitted
   changes. If not, stop and ask the human: it is their checkout.
3. `git merge --no-ff slice/<name>`, with a message naming the slice and
   what it does.
4. The backlog, `docs/backlog.md`: every open item this slice
   built gets status `closed`, Closed by `slice <name>`. Check that every
   Parked item, every Out of scope item marked `Goes to: the backlog`,
   and every finding deferred to the backlog has a row; add any missing.
   Commit only that file: `backlog: after slice <name>`.
5. Push `main`. The tags go with it.
6. Remove the slice worktree, and delete the branch `slice/<name>`: it is
   merged, and the tags keep its approvals.

Tell the human the slice is merged and pushed: the merge commit, and
what is in the backlog from it.

Slices merge into `main` one at a time. Another open slice updates from
`main` in its own Part 1.
