---
review: plan-2
slice: 0001-slice-and-events
kind: plan
target: docs/slices/0001-slice-and-events/plan.md
round: 2
reviewer: doc-reviewer (claude opus)
---

# Review

## Summary

Fifteen of the 17 plan-1 findings are really fixed, and the spec changes
for Q1 and Q2 are correct and match the plan and tasks. Two are
reopened: most briefs are still well over 1,500 words (F10), and the
check order for a line with two faults disagrees with the step that
finds a field twice (F15). One new medium: task 08 builds a git path but
does not follow Cross-platform paths (F1). Nothing needs the human.

## Checked

- Files read: docs/slices/0001-slice-and-events/spec.md, plan.md,
  tasks.md, preview.md, status.md, reviews/plan-1.md,
  docs/templates/plan.md, tasks.md, review.md, spec.md (front matter),
  AGENTS.md, .agents/prompts/doc-reviewer.md
- Ran:
  git diff 23a6e13 HEAD -- docs/slices/0001-slice-and-events/spec.md (S12, S29, Key concepts, R18, R26, R28, R36, R37, R38, R42, R43, Out of scope changed)
  git diff 43e6509 HEAD --stat (plan.md, spec.md, tasks.md, status.md changed since plan-1)
  ls cmd internal (no Go code yet; every task makes new files)
  a Python script that adds up each task's brief: its section, Key concepts, its covers, the scenarios they bring, and its follows (01: 1812, 02: 2073, 03: 1913, 04: 2168, 05: 987, 06: 2012, 07: 1847, 08: 2157, 09: about 2130, 10: 1784, 11: 2004, 12: 2356, 13: 822 words)
- Requirements: R1 to R43, S1 to S29, TD1 to TD8, every shared rule,
  every task's depends_on, touches, follows, and Interfaces

## Tests

None.

## Previous findings

- plan-1/F1: verified (fixed)
- plan-1/F2: verified (fixed)
- plan-1/F3: verified (fixed)
- plan-1/F4: verified (fixed)
- plan-1/F5: verified (fixed)
- plan-1/F6: verified (fixed)
- plan-1/F7: verified (fixed)
- plan-1/F8: verified (fixed)
- plan-1/F9: verified (fixed)
- plan-1/F10: reopened. The split helped, but 11 of 13 briefs are still about 1,800 to 2,350 words, and only task 08 says why. The largest: task 12 (about 2,350), task 04 (about 2,170), task 09 (about 2,130), task 02 (about 2,070). Much of it comes from R38 in covers, which pulls S5, S12, S15, and S27 into tasks 03, 07, 08, 09, 11, and 12, and from Interrupts in task 01. Trim follows and covers where a task builds no part of them, or add one line of reason to each.
- plan-1/F11: verified (fixed)
- plan-1/F12: verified (fixed)
- plan-1/F13: verified (fixed)
- plan-1/F14: verified (fixed)
- plan-1/F15: reopened. Reading a line now says the checks run in R26's order of reasons, where `field "<f>" appears twice` comes after `unknown version`, `unknown event`, and `unknown field`. But the same section, and task 02 step 4, find a name twice first, with `json.Decoder.Token`, before the map is read. For `{"version":2,"seq":1,"seq":1,...}` the order rule gives `unknown version 2`, and the steps give `field "seq" appears twice`. Say which one wins, and fix the other text.
- plan-1/F16: verified (fixed)
- plan-1/F17: verified (fixed)

## Findings

### F1. Task 08 builds a git path, but does not follow Cross-platform paths

Level: medium
Where: tasks.md → Task 08 → follows, plan.md → Flow 2, step 5
Breaks: AGENTS.md (paths must work on Windows, macOS, and Linux)
Why: The split put the R15 folder check in task 08. It runs
  `git cat-file -e <commit>:docs/slices/<name>`, and checks
  `.sliceflow/worktrees/<name>` on disk. Task 08's follows has Flow 2,
  Git arguments, Tests, and Testing, but not Cross-platform paths, which
  says git paths use `path`. A worker who writes `filepath.Join` gets
  `<commit>:docs\slices\x` on Windows. git finds no such path, so a
  slice merged into the base branch is not found. `slice create x` then
  goes on instead of printing
  `Error: slice x already exists: folder docs/slices/x on main. Nothing was changed.`
Fix: add Cross-platform paths to task 08's follows.

### F2. TestSpecFrontMatter passes when the template's ideas line is left in

Level: low
Where: tasks.md → Task 05 → Done when → `TestSpecFrontMatter`,
  `TestSpecNoPlaceholdersInFrontMatter`
Breaks: nothing
Why: `docs/templates/spec.md` has `ideas:` and, under it, a line
  `- {{an idea file this slice builds on, ...}}`. The test checks only
  that the output starts with `...ideas: []\n`. If step 3 writes
  `ideas: []` and keeps the next line, both tests still pass, because
  the second test looks only for `{{slice}}`, `{{title}}`, and
  `{{date}}`.
Fix: check that the output starts with
  `---\nslice: spec-submit\ncreated: 2026-10-25\nideas: []\n---\n`.

### F3. Ctrl-C while FindRepo runs prints a Go error

Level: low
Where: plan.md → Shared rules → Interrupts, tasks.md → Task 09 → Steps,
  step 4
Breaks: nothing
Why: The fix for plan-1/F16 moved `FindRepo` into the CLI, before
  `Create`. Task 09 makes the signal `ctx` in `internal/cli/create.go`.
  If Ctrl-C comes while `FindRepo` runs, `git.Run` returns `ctx.Err()`.
  That is not an `*Error`, so the CLI prints `Error: context canceled`
  as a bug, not R18's `interrupted` error. Interrupts covers the checks
  inside `Create` only.
Fix: say that the CLI prints R18's `interrupted` error when the `ctx`
  is done before `Create` returns, or that it makes the signal `ctx`
  after `FindRepo`.

### F4. R18: the name of the git command, and a step that is not git

Level: low
Where: spec.md → R18, plan.md → Flow 3, step 3
Breaks: nothing
Why: R18 gives `git <command> failed (exit code <n>)` with the example
  `git commit`. It does not say whether `<command>` for
  `git worktree add` is `worktree` or `worktree add`. Steps 5 and 7 of
  Flow 1 write files with OS calls, not git, and R18 does not say what
  the reason is when one of them fails. A worker would guess Go's error
  text, e.g. `open ...: permission denied`.
Fix: in Flow 3, say `<command>` is the git subcommand and its own
  subcommand (`worktree add`), and that an OS error gives its own text
  as the reason. (lead's call)

### F5. R38's list of wrong use in brackets is shorter than its list of messages

Level: low
Where: spec.md → R38
Breaks: nothing
Why: The first sentence says wrong use is "an unknown command or flag,
  or a missing value". The new text adds `missing slice name` and
  `too many arguments` as wrong use too, with exit code 2. The two lists
  now disagree.
Fix: write "(an unknown command or flag, a missing value, a missing
  slice name, or too many arguments)", or drop the brackets.

## Requirements check

None.

## Needs human decision

None.
