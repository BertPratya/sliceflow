# flow-implement: bootstrap mode

Use this while `flow` does not exist yet. It replaces steps 2 and 4 of
`SKILL.md`. All other steps stay the same; where they name a `flow`
command, do the same thing by hand and record it in `status.md`.

You are the loop. You also do what `flow` would record: after every step
below, update the task file's header (status, attempt, review_round, updated)
and the slice's `status.md` (status, attempt, round, findings), and add a
dated Log line. Format: `workflow/formats/status.md`.

Set the phase in `status.md` when its event happens
(`workflow/formats/lifecycle.md` → Slice phases): `implementing` when the
first worker starts, `integration` when all tasks are done, `preview` when
the integration review passes, back to `implementing` for any fix task,
and `done` after the merge.

## Step 2 by hand: run the tasks

Take the next task whose dependencies are done, in task order. Bootstrap
mode has no task branches: run one task at a time, in the slice worktree. Then:

1. **Start a worker** with the tool the stack profile sets for it: your own
   subagent if it is your tool, otherwise that tool's command line
   (e.g. `codex exec "…"`). Use this prompt:
   *"You are the worker. Read `workflow/roles/worker.md` and follow it.
   Your task: `slices/<slice>/tasks/NN-*.md`."*
   In round 2 and later, add: *"Fix the findings in `reviews/task-NN-<round>.md` first."*
   If the task was unblocked, add the unblock note to the prompt.
2. **Read its `RESULT:` line.**
   - `blocked`: handle it as in step 3 of `SKILL.md`.
   - `review`: record the worker's finding lines (`<id>: fixed | deferred | wontfix`)
     in `status.md`, then go on.
3. **Run the project checks yourself** (`AGENTS.md` section 9), for the parts the task touched.
   If they fail, add 1 to attempt and send the worker back with the errors.
   After 3 failures since the last review, the task is blocked.
4. **Check the changed files** are inside the task's `touches`, plus the
   files the stack profile always allows. If not, the task is blocked.
5. **Commit**: `NNNN/NN: <what changed>`, with any "Assumed" note in the body.
   One commit per task, plus one per fix round.
6. **Review**, if the task needs one (see Tracks in `AGENTS.md`): start it
   with `flow-review` (kind `task`, target the task and its commits).
   - `changes`: send the worker back (attempt back to 0; the next review is the next round).
   - `needs_human`: ask the human.
   - `pass`: the task is done.
   If the task needs no review, it is done once the checks pass.

Repeat until every task is done or waiting on the human.

## Step 4 by hand: integration review

1. Update the slice branch: `git merge <base branch>`. On a conflict, leave
   the merge in progress. Write a fix task in `tasks/` from the task template:
   layer `integration`, `touches` = the conflicting files, goal "resolve the
   conflicts, keeping the changes from both sides". Run it as in step 2; its
   commit completes the merge. A conflict in a file from the plan's `touches`
   is yours to resolve, as `SKILL.md` step 4 says.
2. Run the checks for every part the slice touched, then the slice checks
   (tier `slice` in `AGENTS.md` section 9), if there are any.
   A failure becomes a fix task: back to step 2.
3. Start it with `flow-review` (kind `integration`, target the whole slice).
4. For each high or medium finding, write a new fix task in `tasks/` from
   the task template, and go back to step 2.

The update in step 6 of `SKILL.md` works the same way.
