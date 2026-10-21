---
name: flow-implementation
description: Builds an approved slice task by task, with the lead as orchestrator, a worker per task, and reviewers, and handles the decisions that come up while building. Use when a slice's phase is ready or implementing.
---

# flow-implementation

The lead orchestrates. It starts workers and reviewers, runs the checks,
merges, and keeps `status.md`. It does not write task code itself.

## Before the first task

- The phase is `ready`: the planning is approved.
- The slice worktree has no uncommitted changes.
- The base branch passes its checks: in a temporary worktree
  `.worktrees/base-check` (detached, on the base branch), run the setup
  commands and every check in `AGENTS.md`, then remove it. Never use the
  main checkout: it is the human's.

Set the phase to `implementing` when the first task starts.

## The pipeline, per task

A task can start when every task in its `depends_on` is merged into the
slice branch, and its `touches` overlap no unmerged task. Several tasks
may run at once.

1. **Worktree:** branch `task/<name>-NN` from `slice/<name>`, worktree
   `.worktrees/task-<name>-NN/`. Run the setup commands there.
2. **Brief:** write it to `.worktrees/briefs/<name>-NN.md`, from the
   task in `tasks.md` and what it links to, as `docs/templates/tasks.md`
   says.
3. **Worker:** start one in a fresh context (`AGENTS.md` → Tools per
   role). Its prompt: read `.agents/prompts/worker.md` and follow it,
   with the brief, the worktree's full path, and the round. After
   a review or failed checks, add the findings or the last 40 lines of
   output. Read only its result, not its whole conversation.
4. **Checks:** the auto-fix commands, then the task checks, in the task
   worktree. A failure starts a new worker round. After 3 failures since
   the last review, the task is `blocked`.
5. **Scope:** every changed file is in the task's `touches`, or always
   allowed by `AGENTS.md`. If not, the task is `blocked`.
6. **Commit** on the task branch: `<name>/NN: <summary>`.
7. **Review** when the task's risk is `core` or `normal` (not `low`):
   a code reviewer, told to read `.agents/prompts/code-reviewer.md`.
   Changes asked: a new worker round. Pass: merge. After the review
   round limit in `AGENTS.md`, the task is `blocked`.
8. **Merge**, one task at a time: merge the slice branch into the task
   branch, run the task checks again, then `git merge --ff-only
   task/<name>-NN` in the slice worktree. Remove the task worktree and
   delete the task branch. The task is `done`.

Update `status.md` with a Log line after every step, and commit it.

## Decisions while building

Every decision that comes up is one of two kinds. The test: would it
change what `spec.md`, `plan.md`, or `tasks.md` says?

### 1. No: you decide it

Examples: how to write code inside a task, a helper's name, which of two
equal ways to use.

- Decide it. Do not stop the run.
- Write it at once in `status.md`: what you decided, why, and for which
  task.
- Tell the human in a short summary when they next check in. The full
  list goes to the preview.

### 2. Yes: the human decides it

Anything that changes the approved documents: behavior the user sees,
scope, interfaces, data, security.

- Only the tasks it touches wait. Other tasks go on.
- Ask up to 3 at a time. For each: what we have, the problem with a small
  example, the options, and your recommendation.
- When the human answers:
  1. Change the spec, plan, and tasks the answer touches.
  2. Start a doc reviewer (`.agents/prompts/doc-reviewer.md`) for a
     change review, to check the documents still agree with each other. Sort its findings as in planning: high
     back to the human, medium as a `(lead's call)`, low fixed or left.
  3. Commit: `<name>: <what changed, from the human's answer>`.
  4. Make the next approval tag on that commit,
     `flow/<name>/planning/R00N`. Its message quotes the human's answer.
     The answer is the approval: do not ask the human to approve again.
     The documents stay `approved`, with the new revision in `status.md`.
  5. A running task the answer changes: stop it and start it again from
     the changed task. A task already done: add a fix task.

If the answer is big, for example it adds a story or changes most of the
tasks, do not patch it in. Tell the human, and suggest going back to
planning.

## Stops

Stop, mark the task `blocked`, and tell the human in plain words when:

- its checks failed 3 times since its last review
- a review still does not pass after the round limit
- it is unclear, or the spec, plan, or a contract looks wrong
- it needs a new dependency, a storage format change, secrets, or CI
  changes

## When every task is done

Every task is `done` or `cancelled`: the slice moves on to integration.
