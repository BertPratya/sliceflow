# flow-plan: bootstrap mode

Use this while `flow` does not exist yet. Where a step in `SKILL.md` or
`change.md` names a `flow` command, do this by hand instead. After each
one, add a dated Log line to the slice's `status.md`
(format: `workflow/formats/status.md`).

## Creating plan.md and task files

Copy `workflow/templates/plan.md` and `workflow/templates/task.md`. Fill
the `{{placeholders}}`. `based_on` is the approved spec revision for the
plan, and the plan revision being written for tasks.

## `flow validate`

Check by hand:
- every `Covers:` in the plan and `covers:` in a task names a real requirement
- every requirement is covered by a flow or shared rule, and by a task
- each task's requirements match the spec word for word
- `depends_on` has no loops and names real tasks
- each task's `touches` is inside the code folders, and never includes a file in the plan's `touches`
- warning: another unmerged slice changes a file in the plan's `touches`, and
  neither slice is stacked on the other (read the plans in the other slice worktrees)
- Open questions in the plan is empty

## `flow doc submit plan`

1. Set the plan header `status: in_review`.
2. Start the plan review with `flow-review` (kind `plan`, target `R<next>`).

## `flow finding fixed | deferred | wontfix`

Record the new status and your note in the Findings table of `status.md`.

## `flow approve plan`

Only after the human has just said they approve:
1. Set the plan header: `status: approved`, `revision: R<next>`,
   `approved_at` (now, UTC), `approved_by` (the human's name).
2. Copy the plan to `history/plan/R<next>.md`.
3. Save the slice's contract changes: `git diff <base branch> -- <files in touches>`
   into `history/contract/R<next>.patch`.
4. Commit: `NNNN: plan R<next> approved`.
5. Set the phase in `status.md` to `ready`.

## `flow diff plan`

`git diff` between `history/plan/<last revision>.md` and `plan.md`.

## `flow task cancel`

Set the task's `status: cancelled` and record it in `status.md`.

## Editing an approved plan

When you start editing an approved plan (`change.md`), set its header
`status: changed`. Submitting it again sets `in_review`.
