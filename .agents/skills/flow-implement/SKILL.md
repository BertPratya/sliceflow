---
name: flow-implement
description: Runs an approved slice to merged: runs the tasks with workers and reviewers, handles every stop, runs the integration review, the human preview, and the merge. Use when a slice's plan is approved, or the slice is implementing, in integration, or in preview.
---

# flow-implement

You are the lead, and you run the loop. Follow these steps in order. Each
step ends when its "Done when" is true. After every step that changes
something, update `status.md`, add a Log line, and
commit (`workflow/roles/lead.md` → Keeping the status).
Also follow the project rules for the lead and for Everyone
(`AGENTS.md` section 10 → Other rules).

Phases (`workflow/formats/lifecycle.md` → Slice phases): `implementing`
when the first worker starts, `integration` when all tasks are done,
`preview` when the integration review passes, back to `implementing` for
any fix task, and `done` after the merge.

## 1. Check the slice is ready

Start right after the human approves the plan, without asking. Wait only
if the slice is stacked and the earlier slice is not merged, or the human
said to hold.

- The spec and the plan are `approved`, the plan is based on the current
  spec revision, and the contract changes match `history/contract/`: write
  the same diff as at plan approval (`flow-plan` step 10) with `--output` to
  a file under `<repo root>/.worktrees/`, then compare:
  `git diff --no-index --quiet history/contract/R<n>.patch <that file>`.
- The slice worktree has no uncommitted changes.
- A stacked slice: the earlier slice is merged, and this slice has been
  updated from the base branch. If not, tell the human it is waiting, and stop.
- The base checks pass: in a temporary worktree of the base branch
  (`git worktree add --detach <repo root>/.worktrees/base-check <base branch>`),
  run the setup commands, then the task checks from `AGENTS.md` section 9,
  without auto-fix. Then remove it. Never use the main checkout: it is the
  human's working copy. If the checks already fail, tell the human and stop.

Done when: all of these are true.

## 2. Run the tasks

Tasks run in parallel, each in its own worktree. Start every task that can
start, and handle each result as it comes back.

A task can start when every task in its `depends_on` is merged into the
slice branch, and its `touches` do not overlap any task whose task branch
is not merged yet (running, in review, waiting on the human, blocked, or
interrupted). There is no other limit. The overlap rule keeps most merges
free of conflicts.

Where things live: you edit and commit `status.md` only in the slice
worktree, never on a task branch. A task's review files live on its task
branch: create and commit them in the task worktree, and they reach the
slice branch when the task merges. So task branches never touch `status.md`.

`<repo root>` below is the main checkout's folder, not the slice worktree:
the output of `git rev-parse --path-format=absolute --git-common-dir`,
without the last `/.git`.

1. **Create its worktree:** `git worktree add <repo root>/.worktrees/task-NNNN-NN -b task/NNNN-NN <slice branch>`,
   from the newest slice branch, so it has every task merged so far. Always
   use the full path: you work in the slice worktree, and a relative path
   would nest the folder inside it. If `task/NNNN-NN` already exists (an
   interrupted task), reuse its worktree; if the folder is gone,
   `git worktree prune`, then `git worktree add <path> task/NNNN-NN`. If a
   part it touches has a setup command (stack profile, e.g. `npm ci`), run it there.
2. **Start a worker** the way the stack profile's "How to start" column says
   (`AGENTS.md` section 9). Whatever the way: a fresh context, the prompt
   below, it works only in the folder the prompt names, and only its
   `RESULT` block comes back to you. Never read its whole conversation.
   For a command line, write the prompt to a file under
   `<repo root>/.worktrees/` and pass the file or pipe it in, never as one
   quoted argument: test output holds quotes and `$` that the shell changes.
   Every round starts a new worker, with this prompt:

       You are the worker. Read workflow/roles/worker.md and follow it.
       Work only in <absolute path of the task worktree>. All paths are relative to it.
       Your task: slices/<slice>/tasks/NN-<name>.md (round <reviews so far + 1>).
       <after a review: Fix the findings in <path of the last review file> first.>
       <checks failed: The checks failed. Last lines: <up to 40 lines>.>
       <every worker after the first: Your previous attempt: <summary>; Assumed: <note>; changed files: <list>.>
       <merge conflict: Resolve the conflicts in <files>, keeping the changes from both sides.
        Lock files and generated code: run the install or generate command instead.>
       <unblocked: Note from the lead: <answer, and where it is written>.>

3. **Read its `RESULT` block.**
   - `blocked`: see step 3, Handle a stop.
   - `review`: record its finding lines in `status.md` → Findings. A high
     finding marked wontfix, or deferred without naming a later task:
     treat the task as blocked.
4. **Run the checks yourself, in the task worktree:** the auto-fix commands,
   then the task checks, for the parts the task touched (`AGENTS.md` section 9).
   Read only whether each one passed. For a failure, read only the last 40
   lines of its output, and pass those to the next worker; test output can
   be very long. If they fail, add 1 to attempt and start a new worker with
   the errors. After 3 failures since the last review, the task is blocked.
   If Known gaps say the checks are not safe to run at the same time, run
   them in one task worktree at a time.
5. **Check the changed files** are inside the task's `touches`, plus the
   files the stack profile always allows. If not, the task is blocked.
   For a merge conflict fix, check only that the worker changed nothing
   beyond the conflicting files.
6. **Commit on the task branch, before the review:** `NNNN/NN: <summary>`,
   with the "Assumed" note and the worker's FINDINGS lines in the body, so
   the reviewer reads them from the commits. The reviewer reviews exact commits;
   each fix round adds one more commit on top.
7. **Review, only if the task needs one** (see Tracks in `AGENTS.md`):
   `flow-review` (kind `task`, target the task and its commits).
   - `changes`: a new worker, the next round (attempt back to 0).
   - `needs_human`: ask the human. Other tasks go on meanwhile. Then:
     if the answer means the code must change, start a new worker round
     with the answer in the prompt (a normal fix round, then review). If
     it means the code is fine as it is, write "accepted by the human
     ("<their words>")" in the Log, and merge it. If the answer changes
     what the slice should do, it is a spec change. If that review also
     has high or medium findings, it is a fix round either way, with the
     answer in the prompt.
   - `pass`: merge it (step 8).

   A task that needs no review goes straight to the merge once its checks pass.
8. **Merge it into the slice branch**, one task at a time:
   1. In the task worktree: `git merge <slice branch>`, to bring in the
      tasks merged meanwhile. On a conflict, leave the merge in progress
      and start a new worker for this task to resolve it (prompt above);
      then `git add` the resolved files and commit. A conflict fix gets no task review: the checks run
      again in 8.2, and the integration review checks every merge. A conflict in a file from the plan's `touches` (a
      contract) is yours: resolve it, and the human approves the contract
      change again (`flow-plan` → `change.md`).
   2. Run the checks again in the task worktree. This catches two tasks
      that each work alone but break together. On a failure, start a new
      worker for this task with the "checks failed" prompt line. It may
      change files within the task's `touches`. Like a conflict fix, it
      gets no task review (the integration review checks every merge).
      Then commit, and run 8.2 again.
   3. In the slice worktree: `git merge --ff-only task/NNNN-NN`. The slice
      branch only ever moves forward, to code that passed its checks.
   4. Remove the task worktree and branch, from the slice worktree. Stop
      anything still running there and leave the folder first: on Windows,
      a running process or an open terminal locks it.
      `git worktree remove <repo root>/.worktrees/task-NNNN-NN`, `git branch -d task/NNNN-NN`.
      The task is `done`.

   While 8.1 to 8.3 run, edit `status.md` but do not commit it; commit it
   after 8.3. This is safe: task branches never change `status.md`, so
   `--ff-only` works with it edited. If `--ff-only` is still refused
   because the slice branch moved (another merge), repeat 8.1 and 8.2.

   Merge problems do not count against the worker. After 2 failed merges
   in a row, the task is blocked: usually its `touches` or `depends_on` was
   wrong in the plan.

Repeat until every task is done or waiting on the human.

Done when: every task is done, or what is left waits on the human.

## 3. Handle a stop

| What happened | Do |
|---|---|
| A worker asks a question, and the answer is in an approved document | Unblock it with the answer and where it is written. |
| A worker asks a question that needs a new decision | Ask the human. Only they unblock it. |
| A worker needs a file outside its `touches` | Widen `touches` if the path stays inside the project's parts, is not in the plan's `touches`, does not overlap any task whose branch is not merged yet, and adds no behavior; tell the human in one line. Otherwise it is a plan change: ask the human. |
| Checks failed 3 times | Read the errors. If you can see the cause, unblock with a hint. If not, ask the human. |
| A review needs a human decision | Ask the human. |
| 3 review rounds without a pass | Ask the human: accept it as it is, change the plan, or allow one more round. |

Unblocking: set the task to `todo` (or `in_progress` if work was kept),
attempt to 0, and put your note in the next worker's prompt. When several
things wait on the human, ask them in one round, most important first.

If the work is interrupted: the task's work stays in its worktree. Commit
what is there on its task branch as `NNNN/NN: interrupted`, and set the task
back to `todo`. The next worker continues from it.

Give the next worker a better note. The fix itself is the worker's job.

Done when: every stopped item is unblocked, answered by the human, or waiting on the human.

## 4. Integration

When all tasks are done:

1. **Update from the base branch**, so the review and the preview see the
   code that will be merged. Never merge in the slice worktree itself; use
   the same pattern as a task merge (step 2.8):
   1. Create a branch and worktree from the newest slice branch:
      `git worktree add <repo root>/.worktrees/update-NNNN -b update/NNNN <slice branch>`.
   2. There: `git merge <base branch>`. On a conflict, handle it like a
      task merge conflict: leave the merge in progress, start a worker in
      this update worktree with only the merge-conflict prompt line (there
      is no task file), add a Log line, then `git add` the resolved files
      and commit. It gets no review.
      A conflict in a file from the plan's `touches` (a contract) is yours,
      not a worker's: resolve it, then the human approves the contract
      change again (`flow-plan` → `change.md`).
   3. There: run the setup commands, then the slice checks: the task checks
      for every part the slice touched, then the slice-tier checks
      (`AGENTS.md` section 9). On a failure: if the merge brought in changes
      from the base branch, start a worker here with the "checks failed"
      prompt line, then commit and run 4.1.3 again. If it brought in
      nothing, the slice itself is broken: remove the update worktree, and
      write a fix task.
   4. In the slice worktree: `git merge --ff-only update/NNNN`. Then remove
      the update worktree and branch (`git worktree remove`, `git branch -d`).

   As in step 2.8: while 4.1.1 to 4.1.4 run, do not commit `status.md`.
   If `--ff-only` is refused because the slice branch moved, run
   `git merge <slice branch>` in the update worktree and repeat 4.1.3.
2. **Integration review:** `flow-review` (kind `integration`). Each high or
   medium finding becomes a fix task whose Goal names the finding ID and
   quotes it (e.g. `integration-1/F2`).

Any fix task: tell the human in one line what it fixes and which
requirement it serves, then back to step 2. A fix that needs new or changed
behavior, a plan decision, or a contract is not a fix task: it is a spec or
plan change (`change.md` in `flow-spec` or `flow-plan`).

Done when: the integration review passes.

## 5. Preview with the human

First run the setup commands in the slice worktree, for the parts the
slice touched. Then give the human, in this order:
1. what the slice does now, in two or three sentences
2. how to run it: the "Run locally" commands (`AGENTS.md` section 9), in the slice worktree
3. every requirement marked `Check: manual`, as a list to try
4. what was parked for later

A problem the human finds becomes a fix task, as in step 4: back to step 2.
If they ask for something new, it is a spec change, not a fix task.

Done when: the human says the result is approved.

## 6. Merge

1. Update the slice branch from the base branch again, as in step 4.1. It
   may have changed during the preview. If nothing came in, skip it. A
   conflict or a failing check is handled as in step 4.1; if anything
   changed, the integration review runs again, and the human previews again (step 5).
2. Move the parked items to `backlog.md` at the project root. Close
   unmarked low findings as "left as is". Set the phase to `done`. Commit.
3. In the main checkout: check with `git status` that it is on the base
   branch and has no uncommitted changes; if not, ask the human. Then
   `git merge slice/NNNN-<name>`, as the git policy says. Stop anything
   running in the slice worktree, leave it, and remove it:
   `git worktree remove <repo root>/.worktrees/NNNN-<name>`.

Done when: the slice is merged and its phase is `done`.
