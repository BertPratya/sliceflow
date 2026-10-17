# flow-setup: bootstrap mode

Use this while `flow` does not exist yet. Where a step in `SKILL.md`
names a `flow` command, do this by hand instead.

## `flow doctor`

Go through every item in `readiness.md`, in order. For each one, look,
and write one line: ok, missing, out of date, or skipped (listed under
"Known gaps" in `AGENTS.md` section 9).

To compare versions, read `workflow/VERSION` in this repo and in the source.

## Workflow files (`flow init`)

Ask the human for the source: a path or URL of the agent-flow repo, at a
version or tag. From it, copy:
- the text between `<!-- agent-flow:start … -->` and `<!-- agent-flow:end -->`
  into this repo's `AGENTS.md`. If the markers exist, replace only what is
  between them. If not, add the block at the top and sections 8-10 below it.
- `workflow/` (templates, roles, formats, `VERSION`)
- `.agents/skills/`

Never copy the source's sections 8-10. They describe the source project.
