---
name: build-slice
description: Builds an approved slice, one task at a time, with a worker per task, the checks, and code reviews, and handles the decisions that come up while building. Use when a slice's phase is ready or implementing.
---

# build-slice

You are the lead. You start workers and reviewers, run the checks,
commit, and keep `status.md`. You do not write task code yourself.

Work in the slice worktree, `.worktrees/<name>/`, with full paths into
it. Tasks run one at a time, there, on the slice branch: no task
branches, no merges.

"The checks" are the setup, auto-fix, and check commands in `AGENTS.md`.
A full run means all of them, on the whole project.

## Keeping status.md

After every step that changes something, update the slice's `status.md`
as `docs/templates/status.md` says, then commit in the worktree:

- the header: `Phase`, and `Updated` (the time now, UTC)
- the section the step changed
- Needs human: add what now waits on the human; remove what no longer does
- a Log line: time · who · what happened

## Asking the human a question

Use the form in `plan-slice`: what we have now, what is missing (with a
small example), and 2-3 options with their trade-offs and your
recommendation. Ask up to 3 at a time.

## Before the first task

1. The phase is `ready`: the planning is approved.
2. The slice worktree has no uncommitted changes.
3. `main` passes its checks: in a temporary worktree
   `.worktrees/base-check` (detached, on `main`), do a full run, then
   remove the worktree. If it fails, stop and tell the human: a broken
   `main` would make every task fail for a reason it did not cause.
4. Set the phase to `implementing`.

## Each task

Take the next task whose `depends_on` tasks are all `done`, lowest id
first. Before it starts, the slice worktree has no uncommitted changes.

### 1. Write the brief

Write it to `.worktrees/briefs/<name>-NN.md`, from the task in
`tasks.md` and what it links to, as `docs/templates/tasks.md` says.

### 2. Start the worker

Agent tool, type `flow-worker`, in a fresh context. Its prompt: read
`.agents/prompts/worker.md` and follow it, with the brief, the slice
worktree's full path, and the round. In a later round, add the review
file, or the last 40 lines of the failed checks.

In `status.md` → Tasks, the task is `doing`. Read only the worker's
result block, not its whole conversation.

- `RESULT: blocked`: go to "Decisions while building".
- `RESULT: review`: go on.

### 3. Check the worker's result yourself

The worker's "the checks pass" is a claim. Make it a fact:

1. Do a full run of the checks in the slice worktree.
2. Check the scope: every changed or new file (`git status
   --porcelain`) is in the task's `touches`, or always allowed by
   `AGENTS.md`.

A failed check: start a new worker round with the last 40 lines of
output, and add 1 to the task's Attempt. After 3 failures since the
last review, the task is `blocked`, the phase `implementing (blocked)`,
and the reason goes under Needs human.

A file outside `touches`: the task is `blocked`. Widening `touches` is
yours to do alone, when the plan clearly allows it: change `tasks.md`,
and log it under Lead's decisions. Otherwise ask the human.

### 4. Commit

`<name>/NN: <the worker's SUMMARY>`. Put its `ASSUMED` lines in the
commit body.

### 5. Code review

Only when the task's risk is `core` or `normal`. A `low` task goes
straight to step 6.

1. The task is `in_review`. Add 1 to its Round, and set Attempt to 0.
2. Copy `docs/templates/review.md` to
   `docs/slices/<name>/reviews/task-NN-<round>.md`, and fill its front
   matter.
3. Start a code reviewer in a fresh context: Agent tool, type
   `flow-code-reviewer`, model `opus`. Its prompt: read
   `.agents/prompts/code-reviewer.md` and follow it, for a task review
   of task NN, with the slice worktree's full path, the brief, the
   review file, and the task's commits.
4. Write each finding and its state in `status.md` → Findings → Task
   reviews.
5. Findings to fix: a new worker round with the review file, then steps
   3 and 4 again (commit: `<name>/NN: fixes from review <round>`), then
   this review again.
   A finding under Needs human decision: go to "Decisions while
   building".
   A finding the worker marks `deferred: backlog`: add it to the backlog
   at once, as `plan-slice` says.
6. The review found nothing to fix: go on.

The review round limit is in `AGENTS.md`. At the limit, the task is
`blocked`, the phase `implementing (blocked)`, and you ask the human.

### 6. Done

The task is `done`. Update the Tasks progress line. Take the next task.

## Decisions while building

A decision comes from a worker's `RESULT: blocked`, or from a review
finding under Needs human decision. The test: would it change what
`spec.md`, `plan.md`, `tasks.md`, or `preview.md` says?

### No: decide it yourself

Examples: a helper's name, how to split a function, one of two equal
ways.

- Decide it, and go on.
- Write it at once in `status.md` → Lead's decisions: the task, what,
  and why.

### Yes: the human decides it

1. The task is `blocked`, the phase `implementing (blocked)`, and the
   question is under Needs human. Ask it in the question form.
2. Say whether it is **big** or **small**. It is big when the answer
   would do any of these:
   - add or remove a story
   - change stored data or a file format
   - change a contract: a command's flags, an API, a file others read
   - change more than half of the tasks

   Otherwise it is small. The human may override it with one word:
   "big" or "small".

#### A small decision: the answer is the approval

The human decided it, so they do not approve the documents again.

1. Change the documents the answer touches: the spec, plan, tasks, and
   the preview checks. Commit only those files, not the worker's
   uncommitted code.
2. Start a doc reviewer, as in `plan-slice`, for a change review: it
   checks only the change, against the other documents. Fix what it
   finds yourself. A high finding is a new question for the human.
3. Commit: `<name>: <what changed, from the human's answer>`.
4. Make the next approval tag on that commit,
   `flow/<name>/planning/R00N`. Its message names the changed documents
   and quotes the human's answer. Write the message to a file and use
   `git tag -a -F <file>`.
5. In `status.md` → Documents, the changed documents stay `approved`,
   with the new revision, date, and tag. Log the human's words.
6. The tasks the answer touches:
   - **The blocked task:** discard its uncommitted changes in the slice
     worktree, write its brief again, and start a new worker.
   - **A task already done:** add a fix task at the end of `tasks.md`,
     with the next free id, and a row in the Tasks table.
   - **A task not started yet:** it gets the new text in its brief.
7. Set the phase back to `implementing`, and go on.

#### A big decision: back to planning

1. Set the phase to `planning`. In `status.md` → Documents, the
   documents the answer touches are `changed`.
2. Run `plan-slice` from the part the change touches: the spec (Part 1)
   or the plan (Part 2), then the plan review and the human's approval
   (Part 3), with the next revision.
3. Done tasks stay committed. Work they must redo becomes fix tasks at
   the end of `tasks.md`.
4. After the approval, run this skill again: it goes on with the next
   task not `done`.

## Never

- Approve for the human, even if asked to "just continue".
- Move or delete an approval tag.
- Edit a submitted review file.
- Write task code yourself, or weaken, skip, or delete a test to make
  the checks pass.

## When every task is done

Every task is `done` or `cancelled`. Tell the human: what was built,
and every line under Lead's decisions. The finish skill comes next.
