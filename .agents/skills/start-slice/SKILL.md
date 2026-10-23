---
name: start-slice
description: Creates a new slice (branch, worktree, slice folder, spec.md, and status.md) and starts its planning phase. Use when the human asks to start or create a slice.
---

# start-slice

## 1. Settle the name

- The human gives the name. Use it as given: `slice-and-events` and
  `0001-slice-and-events` are both fine.
- If the human gave no name, propose one and wait for their yes.
- The name goes into a branch, a folder, and tags, so it holds only
  lowercase letters, digits, and hyphens.
- The name must be new. Stop and tell the human if any of these exists:
  - the branch `slice/<name>`
  - the folder `docs/slices/<name>/` on `main`
  - the worktree `.worktrees/<name>/`
  - a tag that starts with `flow/<name>/`

When the name is fine, go on. Do not ask again.

## 2. Find the ideas it builds on

Read `docs/ideas/index.md`, if it exists. If one or more ideas fit this
slice, propose them in one line, e.g. "This builds on due-dates and
reminders. Use both?", and wait for the human's yes. The slice names only
ideas the human confirmed. None fit, or the human says none: the slice
has no ideas.

If two ideas disagree, name both. The conflict is settled in planning,
while writing the spec.

## 3. Create the branch and worktree

From `main`:

```
git worktree add -b slice/<name> .worktrees/<name> main
```

Work in the worktree from now on, with full paths into it. Install
nothing: planning holds only Markdown. Setup commands run later, where
code runs.

## 4. Create the slice folder

In the worktree, create `docs/slices/<name>/`. Copy
`docs/templates/spec.md` and `docs/templates/status.md` into it.

## 5. Fill the spec's front matter

- `slice`: the name
- `title`: a short title, from the human's words
- `created`: today, YYYY-MM-DD
- `ideas`: the confirmed idea files, one per line, e.g.
  `- docs/ideas/2026-10-23-due-dates.md`. No ideas: `ideas: []`.

Leave the rest of the spec as the template has it: planning writes it.

## 6. Fill status.md

As `docs/templates/status.md` says:

- the header: `# Status: <name>`, Phase `planning`, Branch
  `slice/<name>`, Updated (the time now, UTC)
- the `After:` line: keep it only when the human said this slice waits
  for another slice, with that slice's name. Otherwise delete it.
- Documents: one row, `spec.md` as `draft`, Round `0`, Revision `-`,
  Approved `-`
- every other section: `None.` Findings keeps its count line and its
  three sub-headings, each with `None.`
- Log: `<time> · lead (<tool>) · slice created`
- delete the template's comments

## 7. Commit in the worktree

`<name>: slice created`

## 8. Mark the ideas

Only when the slice has ideas. In the main checkout:

1. Check it is on `main`. If not, stop and ask the human: it is their
   checkout.
2. For each idea: add the slice name to its row in `docs/ideas/index.md`
   (the Slices column) and to its file's `slices:` list, and set both
   statuses to `in slices`.
3. Commit only those files: `ideas: <idea topics> in slice <name>`.

## 9. Tell the human

The slice is ready: its branch, its worktree, its folder, and the ideas
it builds on. This skill ends here. Planning comes next, with its own
skill.
