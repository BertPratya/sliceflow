---
name: flow-slice
description: Creates a new slice (branch, worktree, and slice folder) and starts its planning phase. Use when the human asks to create, start, or open a slice.
---

# flow-slice

## The name

- The human gives the name. Use it as given: `slice-and-events` and
  `0001-slice-and-events` are both fine.
- If the human gave no name, propose one and wait for their yes.
- The name goes into a branch, a folder, and tags, so it may hold only
  lowercase letters, digits, and hyphens.
- Reserved, never a slice name: `base-check`, `briefs`, `messages`, and
  names that start with `task-` or `update-`.
- If the branch `slice/<name>` or the folder `docs/slices/<name>/` already
  exists, stop and tell the human.

When the name is fine, create the slice. Do not ask again.

## Create it

1. Create the branch and its worktree from `main`:
   `git worktree add -b slice/<name> .worktrees/<name> main`
2. In the worktree, create `docs/slices/<name>/`.
3. Copy `docs/templates/spec.md` and `docs/templates/status.md` into it.
4. Fill the spec's front matter: `slice` (the name), `title` (a short
   title, from the human's words), `created` (today).
5. Fill `status.md`, as its template says:
   - the header: `# Status: <name>`, Phase `planning`, Branch
     `slice/<name>`, Updated (the time now, UTC). Delete the After: line,
     unless the slice waits for another one.
   - Documents: `spec.md` as `draft`, Revision `-`, Approved `-`.
   - Every other section: `None.`
   - Log: `<time> · lead (<tool>) · slice created`.
   - Delete the template's comments.
6. Commit in the worktree: `<name>: slice created`.

Work in the worktree from now on, with full paths into it.

## Then

Tell the human the slice is created: its branch, its worktree, and its
folder. This skill ends here. The planning phase is the
`flow-planning` skill.
