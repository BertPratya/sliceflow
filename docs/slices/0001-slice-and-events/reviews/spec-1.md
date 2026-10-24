---
review: spec-1
slice: 0001-slice-and-events
kind: spec
target: docs/slices/0001-slice-and-events/spec.md
round: 1
reviewer: doc-reviewer (claude opus)
---

# Review

## Summary

The spec is clear, its scenarios match the requirements, and every
decision has a Rule: line. No gap needs the human. Ten medium gaps would
make a worker guess: mainly the order of the checks before creating (F1),
the text of the built-in template (F2), and what counts as a slice when a
branch has no log (F4).

## Checked

- Files read: docs/slices/0001-slice-and-events/spec.md,
  docs/slices/0001-slice-and-events/status.md, docs/templates/spec.md,
  docs/templates/review.md, .agents/prompts/doc-reviewer.md, AGENTS.md
- Ran:
  grep for leftover `{{` and `<!--` in spec.md (none found)
  cat -A on the S20 table, to count the column widths (gaps of 3, 2, 2)
  git show a4ba6d1 and 2c4b9d3, to see the human's answers to Q1-Q4
- Requirements: R1-R36, S1-S24, SD1-SD13

## Tests

None.

## Previous findings

None.

## Findings

### F1. The order of the checks before creating is not given

Level: medium
Where: spec.md → Requirements → Checks before creating (R11-R16), R14
Breaks: nothing
Why: Each check prints its own message, but the spec does not say which
  runs first. Input: `sliceflow slice create "My Slice"` in a folder that
  is not a git repo. One worker prints R13's message, another prints R11's.
  The same goes for R12 against R15: in a repo with no commits, `main`
  does not exist either. And when R14 finds both a local and a remote
  branch, the spec does not say which `<where>` is printed.
Fix: say "in this order": R11, R12, R13, R15, R14, R16. For R14, print
  the first hit in the order the requirement lists them.

### F2. The text of the built-in spec template is not clear

Level: medium
Where: spec.md → Requirements → R7, Assumptions (second), Decisions → SD8, SD12
Breaks: nothing
Why: R7 says the template is built into sliceflow, and the assumption
  says it is this repo's template text. But docs/templates/spec.md has
  `title: {{title}}`, a list placeholder for `ideas:`, and comments that
  say "sliceflow fills slice, title, and date itself" and "The title of
  the slice lives only in this front matter". All of these disagree with
  R7 and SD12. A worker must guess whether to change the comment text,
  and how. R7 also does not give the form of `created`. Input: create at
  2026-10-24 01:00 in UTC+7. A local date gives `2026-10-24`, a UTC date
  gives `2026-10-23`.
Fix: say the built-in template is docs/templates/spec.md with no `title`
  line, `ideas: []`, and its comments changed to match SD12. Say
  `created` is the UTC date as YYYY-MM-DD, the same day as the event's `at`.

### F3. The relative paths in the create output do not say what they are relative to

Level: medium
Where: spec.md → Story 1 → S1, S4; Requirements → R9
Breaks: nothing
Why: S1 prints `Worktree: .sliceflow/worktrees/spec-submit` and a
  relative `Next:` path. Input: the command runs in the subfolder `cmd/`,
  or inside another worktree as in S4. Relative to the folder you are in,
  that path is wrong. One worker prints it relative to the root of the
  main checkout. Another prints it relative to the current folder.
Fix: state the rule in R9. For example: relative to the current folder
  when that folder is the root of the main checkout, else the full path.

### F4. What counts as a slice is not defined when a branch has no log

Level: medium
Where: spec.md → Key concepts (In progress, Merged); Requirements → R29, R30; Story 2 → S18
Breaks: nothing
Why: In S18, `slice/old-idea` has no log and is still a slice. R30
  sorts "slices with no created time" last. But the Key concepts define
  In progress and Merged by "a slice whose log is on" a branch. By those
  words, `old-idea` is neither, so it is never listed. R29 ("no slice has
  the given name") also does not say what makes a slice exist. Input: a
  local branch `slice/old-idea` with no log, then `sliceflow slice list`.
  One worker lists it as `unknown !`, another leaves it out.
Fix: define a slice as a local branch `slice/<name>`, or a folder
  `docs/slices/<name>/` on the base branch. In progress: it has the local
  branch, and the base branch has no such folder. Merged: the base branch
  has the folder.

### F5. status and list do not say which base branch they use, or how they fail

Level: medium
Where: spec.md → Requirements → R20, R31; Checks before creating
Breaks: nothing
Why: R3 gives the base branch only "when creating". `status` and `list`
  read the base branch (R20, R31), but they have no `--from`, and the spec
  does not say which branch they read. Their failure paths are missing
  too. R11 and R12 are under "Checks before creating", and "Nothing was
  changed" fits only create. Input: `sliceflow slice list` outside a git
  repo, in a repo with no commits, or with `sliceflow.base develop` and no
  `develop`. The spec gives no message and no exit code.
Fix: say the read commands use `git config sliceflow.base`, else `main`.
  Give each failure its message and exit code 1, without
  "Nothing was changed."

### F6. R22 does not name its reasons, and misses some bad lines

Level: medium
Where: spec.md → Requirements → R22
Breaks: nothing
Why: Only `not valid JSON` is a named reason (S17). The reasons for a
  blank line, a wrong `version`, an unknown event, and an unknown or
  missing field are not given, so each worker writes its own words. Some
  lines are not covered at all. Input: `"seq":"1"`, an `at` that is not
  RFC 3339, a `data` that is not an object, a `slice.created` with no
  `data.from`, or the fields in another order than R18. The spec does not
  say whether to skip the line or keep it.
Fix: list each reason's exact text. Add "a field has the wrong type or
  form" as a reason to skip. Say whether field order matters when reading.

### F7. R23 does not say whether a kept line counts as the last good line

Level: medium
Where: spec.md → Requirements → R23
Breaks: nothing
Why: R23 keeps a line whose `seq` is wrong. Input: lines with `seq` 1, 4,
  5. If the kept line 2 counts as good, line 3 expects 5 and gives no
  warning. If it does not count, line 3 gets the warning `seq 5, expected 2`.
Fix: say a kept line counts as the last good line.

### F8. The message for wrong use is not named

Level: medium
Where: spec.md → Requirements → R34
Breaks: nothing
Why: R34 gives exit code 2 for an unknown command or flag, or a missing
  value. It gives no message. Input: `sliceflow slice create` with no
  name, `--from` with no value, or `--colour`. The worker must make up
  the text. It also must guess whether `--help` exists.
Fix: give one message form, e.g.
  `Error: <problem>. Usage: sliceflow slice create <name> [--from <branch>] [--title <text>] [--json]`,
  and say if `--help` is in scope.

### F9. The spec does not say whether git hooks and signing run on the slice commit

Level: medium
Where: spec.md → Requirements → R8, R17
Breaks: nothing
Why: The project may have a pre-commit hook, or commit signing that
  asks for a passphrase. Input: a repo whose pre-commit hook fails on
  every commit. One worker skips the hooks and the commit works. Another
  runs them, and the create is rolled back by R17. The user sees two
  different results.
Fix: say the commit runs as a plain `git commit` would, with the user's
  hooks and settings. A failure is a failed step under R17.

### F10. The list layout has no rule, and S20's gaps differ

Level: low
Where: spec.md → Story 3 → S20; Requirements → R30
Breaks: nothing
Why: In S20 the NAME column is the longest name plus 3 spaces. TITLE and
  PHASE are the longest value plus 2 spaces. R30 gives no rule for
  column widths or long titles, or for two slices with the same created
  time.
Fix: say each column is as wide as its longest value plus 2 spaces, and
  that two slices with the same time sort by name. Fix S20 to match.

### F11. "Short commit" has no length

Level: low
Where: spec.md → Requirements → R9, R25
Breaks: nothing
Why: S1 shows 7 characters. git's own short form can be longer in a big
  repo. The spec does not say a fixed 7, or git's short form.

### F12. SD13 is a "how" decision

Level: low
Where: spec.md → Decisions → SD13
Breaks: nothing
Why: A shared core, and how a later server calls it, is about how the
  code is built. The template keeps that out of the spec. It has Rule:
  none, and it fits the plan as a TD.

### F13. "Committed log" is defined but never used

Level: low
Where: spec.md → Key concepts
Breaks: nothing
Why: No requirement or scenario uses the term. R20 says the same thing in
  other words.

### F14. S12's Given may not make the commit fail

Level: low
Where: spec.md → Story 1 → S12
Breaks: nothing
Why: With no user name or email set, git can guess a name from the
  system and commit anyway. A test of S12 needs git to refuse the commit,
  e.g. with `user.useConfigOnly`. The Given could name that.

### F15. R6 skips the case where `.sliceflow/` exists with no `.gitignore`

Level: low
Where: spec.md → Requirements → R6
Breaks: nothing
Why: R6 writes `.gitignore` only when sliceflow creates the folder. If the
  folder is there without the file, the new worktree shows up in
  `git status` of the main checkout.
Fix: write the file whenever it is missing.

## Requirements check

None.

## Needs human decision

None.
