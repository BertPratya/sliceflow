# flow-spec: bootstrap mode

Use this while `flow` does not exist yet. Where a step in `SKILL.md` or
`change.md` names a `flow` command, do this by hand instead. After each
one, add a dated Log line to the slice's `status.md`
(format: `workflow/formats/status.md`).

Set the phase in `status.md` when its event happens
(`workflow/formats/lifecycle.md` → Slice phases): `spec` when you write the
first spec draft, `plan` when the human approves the spec.

## `flow slice new <name>`

1. Find the highest slice number in `slices/` on the base branch and in the
   names of `slice/*` branches. The new slice is the next one, 4 digits: `0007-<name>`.
2. Create the branch and worktree the git policy names (`AGENTS.md` section 10):
   `git worktree add <worktree folder> -b <slice branch> <start>`.
   `<start>` is the base branch, or the earlier slice's branch when stacked.
   Do every following step, and all later work on the slice, in that worktree.
3. Create `slices/NNNN-<name>/` with `tasks/`, `reviews/`, and `history/`.
4. Copy `workflow/templates/spec.md` to `spec.md`. Fill `slice`, `title`,
   and `created` (today, UTC).
5. Create `status.md` with the header (phase `spec`, and `After: NNNN`
   when stacked) and every other section "(none)".

## `flow slice move <slice> --after <NNNN | base>`

Only in phase spec or plan. In the slice worktree:
`git rebase --onto <new start> <old start>`. Update `After:` in `status.md`.

## `flow validate`

Check by hand:
- every requirement is `- **R<n>.** …` with a `Check:` line (unit, api, e2e, manual)
- numbers are unique, and none from the last approved revision in
  `history/spec/` is missing
- Open questions is empty

## `flow doc submit spec`

1. Submitted again after answering the spec check, with no new requirements
   since that check, or with `--wording-only`: set `status: awaiting_approval`
   and stop here.
2. Otherwise set the spec header `status: in_review`, and start the spec check
   with `flow-review` (kind `spec`, target `R<next>`). For a change, or after
   answers added requirements, tell the reviewer which requirements to check.

## `flow approve spec`

Only when the spec is `awaiting_approval`, and the human has just said they approve:
1. Set the spec header: `status: approved`, `revision: R<next>`,
   `approved_at` (now, UTC), `approved_by` (the human's name).
   `R<next>` is the next revision, 3 digits: R001, R002, …
2. Copy the spec to `history/spec/R<next>.md`.
3. Commit: `NNNN: spec R<next> approved`.

## `flow reject spec`

Set the spec header `status: draft`, and add a Log line with the reason.

## `flow status`

List the slices not merged yet: `git branch --list 'slice/*' --no-merged <base branch>`.
Read each one's `status.md` in its worktree.

## `flow diff spec`

`git diff` between `history/spec/<last revision>.md` and `spec.md`.

## Editing an approved spec

When you start editing an approved spec (`change.md`), set its header
`status: changed`. Submitting it again sets `in_review`.
