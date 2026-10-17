---
name: flow-implement
description: Runs an approved slice to merged: starts the loop, handles every stop, runs the integration review, the human preview, and the merge. Use when a slice's plan is approved, or the slice is implementing, in integration, or in preview.
---

# flow-implement

You are the lead. Follow these steps in order. Each step ends when its
"Done when" is true. If `flow` does not exist yet, use `bootstrap.md` in
this folder for steps 2 and 4.

## 1. Check the slice is ready

- The plan is approved, with its tasks. `flow validate` passes.
- The slice branch exists, as the git policy says (`AGENTS.md` section 10).
- A stacked slice: the earlier slice is merged, and this slice has been
  updated from the base branch. If not, tell the human it is waiting, and stop.
- `flow check --base` passes. If the base branch already fails, tell the human and stop.

Done when: all of these are true.

## 2. Run the loop

- **Claude Code:** run `flow run` as a background command. You are woken when it exits.
- **Cursor, or any tool without background commands:** run `flow run --detach`,
  then `flow wait`. If it says "still running", run `flow wait` again.
  While waiting, only wait.

Done when: `flow run` has exited and you have its report.

## 3. Handle the stop

| Report says | Do |
|---|---|
| All tasks done | Go to step 4. |
| A worker asks a question, and the answer is in an approved document | `flow task unblock <slice>/<task> --note "<answer, and where it is written>"`. Back to step 2. |
| A worker asks a question that needs a new decision | Ask the human. Only they unblock it. |
| Checks failed 3 times | Read the errors. If you can see the cause, unblock with a hint for the next worker. If not, ask the human. |
| A review asks for a human decision | Ask the human. |
| Interrupted | Back to step 2. |
| `flow` error | Tell the human. Wait for them. |

Give the next worker a better note. The fix itself is the worker's job.

Done when: every stopped item is unblocked, answered by the human, or waiting on the human.

## 4. Integration review

When all tasks are done, `flow run`:

1. Updates the slice branch from the base branch, so the review and the
   preview see the code that will be merged. On a conflict, it leaves the
   update in progress and adds a fix task that names the conflicting files.
   A worker resolves them, keeping the changes from both sides.
   A conflict in a file from the plan's `touches` (a contract) is yours,
   not a worker's: resolve it, then the human approves the contract change
   again (`flow-plan` → `change.md`).
2. Runs the slice checks (`flow check --slice`). A failure becomes a fix task.
3. Starts the integration review. Each high or medium finding becomes a fix task.

Any fix task: tell the human in one line what it fixes and which
requirement it serves, then back to step 2. A fix that needs new or changed
behavior, a plan decision, or a contract is not a fix task: it is a spec or
plan change (`change.md` in `flow-spec` or `flow-plan`).

Done when: the integration review passes.

## 5. Preview with the human

Give the human, in this order:
1. what the slice does now, in two or three sentences
2. how to run it on their machine
3. every requirement marked `Check: manual`, as a list to try
4. what was parked for later

A problem the human finds becomes a fix task, as in step 4: back to step 2.
If they ask for something new, it is a spec change, not a fix task.

Done when: the human says the result is approved.

## 6. Merge

1. Update the slice branch from the base branch again. It may have changed
   during the preview.
2. If anything came in, `flow check --slice` must pass. A conflict or a
   failing check becomes a fix task, as in step 4. Then the integration
   review runs again, and the human previews again (step 5).
3. Merge and push as the git policy says.
4. `flow slice done`: parked items move to `backlog.md` at the project root.

Done when: the slice is merged and its phase is `done`.
