---
review: spec-2
slice: 0002-spec-and-state
kind: spec
target: docs/slices/0002-spec-and-state/spec.md
round: 2
reviewer: doc-reviewer (claude opus)
---

# Review

## Summary

All 18 findings of spec-1 are fixed, or left for an acceptable reason.
What 0002 says about 0001 matches 0001's merged spec and code. Two fixes
caused medium gaps: R15 does not count the template's own indented
`ideas` list as a list (F1), and R11 makes `spec validate` take the
lock, which R6 keeps for writing an event (F2).

## Checked

- Files read: docs/slices/0002-spec-and-state/spec.md,
  docs/slices/0002-spec-and-state/status.md,
  docs/slices/0002-spec-and-state/reviews/spec-1.md,
  docs/slices/0001-slice-and-events/spec.md, docs/templates/review.md,
  docs/templates/spec.md, docs/design/roadmap.md, docs/backlog.md,
  AGENTS.md, .agents/prompts/doc-reviewer.md, internal/template/spec.md,
  internal/template/template.go, internal/eventlog/parse.go,
  internal/eventlog/event.go, internal/slice/status.go,
  internal/slice/repo.go, internal/slice/check.go,
  internal/slice/errors.go, internal/cli/cli.go, internal/cli/flags.go,
  internal/cli/output.go, internal/cli/status.go, internal/cli/list.go
- Ran:
  git diff 0068c7b HEAD -- docs/slices/0002-spec-and-state/spec.md (the fixes since spec-1)
  grep for review.submitted in docs/design/ (sliceflow-design.md lines 761 and 843 still say review.submitted)
  a test repo with a pre-commit hook that reads .git/COMMIT_EDITMSG (it sees the message of the commit before, not its own)
  a test repo with a commit-msg hook that refuses messages with review (refuses only those; git prints nothing, exit code 1)
- Requirements: R1 to R37, S1 to S31, SD1 to SD15, and every reference
  to 0001-slice-and-events (R7, R8, R12 to R14, R18, R24 to R33, R38,
  R40, SD2, SD4, SD5) against 0001's merged spec and code

## Tests

None.

## Previous findings

- spec-1/Q1: verified (fixed)
- spec-1/F1: verified (fixed)
- spec-1/F2: verified (fixed)
- spec-1/F3: verified (fixed)
- spec-1/F4: verified (fixed)
- spec-1/F5: verified (fixed)
- spec-1/F6: verified (fixed)
- spec-1/F7: verified (fixed)
- spec-1/F8: verified (fixed)
- spec-1/F9: verified (fixed)
- spec-1/F10: verified (fixed)
- spec-1/F11: verified (fixed)
- spec-1/F12: verified (fixed)
- spec-1/F13: verified (fixed)
- spec-1/F14: verified (fixed)
- spec-1/F15: verified (wontfix)
- spec-1/F16: verified (fixed)
- spec-1/F17: verified (fixed)

## Findings

### F1. The template's own `ideas` list is not a list by R15

Level: medium
Where: spec.md → Requirements → R15
Breaks: nothing
Why: the F13 fix says `ideas` is a list when it "is followed by lines
  that start with `- `". The built-in template (internal/template/spec.md
  line 4, and S29) writes its item indented by two spaces, then `- `.
  Input: a spec whose front matter is `ideas:` and then
  `  - docs/ideas/2026-10-23-due-dates.md`, the form the template shows.
  Read as written, the line starts with two spaces, not `- `, so
  `spec validate` reports
  `front matter ideas is not a list, e.g. ideas: []`. A worker must
  guess whether indented items count.
Fix: "followed by lines that start with `- ` after any spaces".

### F2. R11 makes `spec validate` take the lock

Level: medium
Where: spec.md → Requirements → R11, R6, R13
Breaks: R6, R13
Why: the F7 fix gives one order for `spec validate`, `spec submit`, and
  `slice title`: ... this requirement, "then the lock (R6)", then the
  command's own checks. R6 takes the lock only "before a command writes
  an event", and R13 says `spec validate` writes nothing. Input:
  `spec validate` while `.sliceflow/locks/login-google.lock` exists. By
  R11, it waits 10 seconds and fails with R7's busy error. By R6 and
  R13, it takes no lock and prints the problems.
Fix: in R11, "then, for `spec submit` and `slice title`, the lock (R6)".

### F3. R9's "read without error" is a new term for 0001's "last good line"

Level: low
Where: spec.md → Requirements → R9
Breaks: nothing
Why: the F9 fix sets the new `seq` from "the last line read without
  error". 0001's R27 says a line with a `seq` warning is kept and "counts
  as the last good line", and 0001's code (internal/eventlog/parse.go)
  does so. A line with a `seq` warning is read with a warning, so a
  worker may doubt whether it counts. Only the 0001 reading gives a log
  with no new `seq` warning.
Fix: "1 more than the `seq` of the last good line, as R27 of
  0001-slice-and-events says, also when R4 skipped that line".

### F4. The third assumption is now checked

Level: low
Where: spec.md → Assumptions
Breaks: nothing
Why: it assumes 0001 skips an unknown event with a warning, and says it
  is noticed "after 0001 merges". 0001 is merged. Its parser returns
  `unknown event "<e>"` for every event but `slice.created`
  (internal/eventlog/parse.go), so an older sliceflow skips 0002's
  events, as 0001's R26 says. It is a fact now, not an assumption.
Fix: remove it, or say it is checked.

## Requirements check

None.

## Needs human decision

None.
