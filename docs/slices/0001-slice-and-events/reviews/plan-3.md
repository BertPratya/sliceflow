---
review: plan-3
slice: 0001-slice-and-events
kind: plan
target: docs/slices/0001-slice-and-events/plan.md
round: 3
reviewer: doc-reviewer (claude opus)
---

# Review

## Summary

Six of the seven findings marked fixed are really fixed. plan-2/F2 is
reopened: the new check counts lines that start with `ideas:`, but the
template's leftover line under it starts with `  - {{`, so the test
still passes with it left in. The spec's new R18 text for a step that
is not git did not reach Flow 3 or task 09, and Go's file errors would
print a Windows path with `\` (F1, medium). Nothing needs the human.

## Checked

- Files read: docs/slices/0001-slice-and-events/spec.md, plan.md,
  tasks.md, preview.md, status.md, reviews/plan-1.md (F10, F15),
  reviews/plan-2.md, docs/templates/review.md, docs/templates/spec.md
  (front matter), AGENTS.md, .agents/prompts/doc-reviewer.md
- Ran:
  git diff a94da0f HEAD --stat (plan.md, spec.md, tasks.md, status.md changed since plan-2)
  git diff a94da0f HEAD -- plan.md tasks.md spec.md preview.md (the plan-2 fixes)
  git diff 23a6e13 HEAD --stat -- spec.md (spec changed since its review: R18 and R38 are the new parts since plan-2)
  a Python script that adds up each task's brief: its section, Key concepts, its covers, their scenarios, and its follows (01: 1560, 02: 2119, 03: 1929, 04: 2191, 05: 995, 06: 2025, 07: 1870, 08: 1975, 09: 2058, 10: 1796, 11: 1772, 12: 2381, 13: 819 words)
- Requirements: R18, R26, R38, R41, R7, and the covers and follows of
  every task touched by the fixes (01, 02, 03, 05, 07, 08, 09, 11, 12)

## Tests

None.

## Previous findings

- plan-2/F1: verified (fixed)
- plan-2/F2: reopened. The template's front matter has `ideas:` and,
  under it, `  - {{an idea file this slice builds on, ...}}`. If `Spec`
  writes `ideas: []` and keeps that next line, the front matter still
  has exactly one line that starts with `ideas:`, and
  `TestSpecNoPlaceholdersInFrontMatter` looks only for `{{slice}}`,
  `{{title}}`, and `{{date}}`. Both tests pass. Check that the output
  starts with `---\nslice: spec-submit\ncreated: 2026-10-25\nideas: []\n---\n`,
  or that the front matter has no `{{`.
- plan-2/F3: verified (fixed)
- plan-2/F4: verified (fixed)
- plan-2/F5: verified (fixed)
- plan-1/F10: verified (fixed)
- plan-1/F15: verified (fixed)

## Findings

### F1. A failed file write prints a Windows path with `\`

Level: medium
Where: plan.md → Flow 3, step 3, tasks.md → Task 09 → Steps, step 3
Breaks: R41, R18
Why: The plan-2/F4 fix added to R18: for a step that is not a git
  command, `<reason>` is the system's error message. Flow 3 and task 09
  step 3 were not changed: they give the reason only from
  `git.Error.Stderr`, and no test has a step that is not git. The
  simple code is `err.Error()` of Go's `*fs.PathError`. On Windows,
  when writing `spec.md` fails, that prints
  `Error: could not create slice x: open C:\Users\me\p\.sliceflow\worktrees\x\docs\slices\x\spec.md: Access is denied. ...`,
  a path with `\`, which R41 does not allow. Cross-platform paths puts
  only the CLI's own paths through `filepath.ToSlash`, not a path inside
  an error text.
Fix: in Flow 3 and task 09 step 3, say the reason of a failed OS call
  is its error text with any path in it written with `/` (for a
  `*fs.PathError`: `Op`, `filepath.ToSlash(Path)`, and `Err`), and add a
  unit test with a step that is not git.

### F2. Task 09 does not say how to get `<command>` from the git arguments

Level: low
Where: tasks.md → Task 09 → Steps, step 3, plan.md → Flow 3, step 3
Breaks: nothing
Why: R18 now says `<command>` is the git command without its options
  and values, e.g. `git worktree add`. Task 09 builds it from
  `git.Error.Args`, which for that step is
  `worktree add -b slice/<name> <worktree> <commit>`. Only the example
  shows that the two paths after the options are left out too; a worker
  who drops only words starting with `-` and their values would print
  `git worktree add .sliceflow/worktrees/x 9dd7ebf... failed`. The only
  e2e test (S12) is `git commit`.
Fix: in task 09 step 3, say each step names its own command, e.g.
  `worktree add`, `add`, `commit`, or that `<command>` is the leading
  arguments up to the first one that starts with `-`.

## Requirements check

None.

## Needs human decision

None.
