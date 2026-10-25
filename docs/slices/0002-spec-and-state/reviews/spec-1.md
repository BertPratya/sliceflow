---
review: spec-1
slice: 0002-spec-and-state
kind: spec
target: docs/slices/0002-spec-and-state/spec.md
round: 1
reviewer: doc-reviewer (claude opus)
---

# Review

## Summary

The spec is complete and mostly exact, and its scenarios match R29's
table. One open point needs the human: what happens to a committed log
whose last line is half written (Q1). The main medium problems: a
review file left behind by a failed commit blocks the retry that S23
shows (F3), and `spec validate` on a fresh spec checks the example items
inside the template's comments (F1).

## Checked

- Files read: docs/slices/0002-spec-and-state/spec.md,
  docs/slices/0002-spec-and-state/status.md, docs/templates/spec.md,
  docs/templates/review.md, AGENTS.md, .agents/prompts/doc-reviewer.md,
  docs/design/roadmap.md, docs/design/sliceflow-design.md (1.4, 1.13,
  1.13b, 2.2), docs/backlog.md, the 0001 spec at tag
  sliceflow/0001-slice-and-events/planning/R004, and
  internal/template/spec.md and internal/template/template.go at tag
  sliceflow/0001-slice-and-events/result/R001
- Ran:
  git show sliceflow/0001-slice-and-events/planning/R003, R004 and result/R001 :docs/slices/0001-slice-and-events/spec.md (all three the same)
  git show sliceflow/0001-slice-and-events/result/R001:internal/template/spec.md (starts with `ideas:` and a `{{` item; comments hold list items `**S1.**`, `**S2.**`, `**SC1.**`, `**R3.**`, `**R9.**`, `**SD2.**`)
  git grep ideas on the same tag (`slice create` fills `ideas: []`)
- Requirements: R1 to R37, S1 to S30, SD1 to SD14

## Tests

None.

## Previous findings

None.

## Findings

### F1. A fresh spec's template comments are checked as content

Level: medium
Where: spec.md → Requirements → R15, R16, R18
Breaks: nothing
Why: R18 skips fenced code blocks and `inline code`, but not the text
  inside a template comment. The built-in template (0001's
  internal/template/spec.md) has example list items inside its comments:
  `- **S1.**`, `- **S2.**`, `- **SC1.**`, `- **R3.**`, `- **R9.**`,
  `- **SD2.**`, and `{{` in its rule text. Input: `spec validate` on a
  spec just made by `slice create`, the state of S1. Read as written, it
  reports `S1 is used twice`, `S2 is used twice`, `SC1 is used twice`,
  `S1 shows R2, which does not exist`, and placeholder warnings for the
  rule text, on top of one `template comment left` per comment. A worker
  must guess whether comments are skipped.
Fix: in R18, also skip everything inside a template comment. R16 still
  reports each comment once, at its `<!--`.

### F2. S28 shows front matter the built-in template does not have

Level: medium
Where: spec.md → User stories → S28, spec.md → Requirements → R35
Breaks: R35
Why: R35 prints 0001's built-in template "before its front matter is
  filled". That file starts with `ideas:` and then the item
  `  - {{an idea file this slice builds on, ...}}`. `ideas: []` is what
  `slice create` writes when it fills the front matter. S28 shows
  `ideas: []`. A test written from S28 fails against R35, or the worker
  changes the template to match S28.
Fix: show the template's real first lines in S28.

### F3. A review file left by a failed commit blocks the retry in S23

Level: medium
Where: spec.md → Requirements → R10, R24, R25, spec.md → User stories → S22, S23
Breaks: R24, S23
Why: R25's commit holds the new file `reviews/spec-<round>.md`, so the
  file is written before the commit. In S22 that commit fails. R10 puts
  back only the log, so `reviews/spec-1.md` stays on disk. In S23, R24
  then finds `reviews/spec-1.md` with no review started, and refuses
  with `Error: docs/slices/login-google/reviews/spec-1.md already exists, but review 1 has not started. ...`.
  S23 says review 1 starts. In the same way, a failed `review.finished`
  commit leaves the filled review file, and files sliceflow staged for a
  failed commit stay staged, while the message says "Nothing was
  changed."
Fix: in R10, put back every file the event wrote, as it was committed
  (remove a file the event made), and leave git's index as it was.

### F4. `slice title` "in any phase" conflicts with R2 when the phase is unknown

Level: medium
Where: spec.md → Requirements → R32, R2, R33
Breaks: R2
Why: R32 says `slice title` works in any phase. R2 allows `slice.titled`
  only after a good `slice.created`. Input: `slice title old-idea "Old idea"`
  for a slice with no log (0001 S18, phase `unknown`). If it saves the
  event, replay skips it with `comes before slice.created`, and the
  title never shows. If it refuses, no message is named.
Fix: refuse with R20's message for this case,
  `Error: slice <name> has no slice.created event, so its state is unknown. Nothing was changed.`,
  and say "any phase but unknown".

### F5. R37 does not cover the step lines printed before a later failure

Level: medium
Where: spec.md → Requirements → R37, R10, spec.md → User stories → S22
Breaks: R38 of 0001-slice-and-events
Why: S22 prints `Submitted spec.md.` on stdout, and then fails with exit
  code 1. A failed `review.finished` commit would likewise follow
  `Review 1 started: ...`. R37 names only the problem lines, the summary
  line, and the `Submit refused:` line as the exception to 0001's R38
  (nothing on stdout when a command fails). A worker who follows R37
  must drop those lines, and then S22 fails.
Fix: add to R37 the step lines (`Submitted spec.md.`,
  `Review <n> started: ...`) printed before a later step fails.

### F6. No message when a later event's move is refused under the lock

Level: medium
Where: spec.md → Requirements → R6, R20, R23, R25, R27
Breaks: nothing
Why: R6 takes the lock for one event, and checks the move again by R2
  each time. The lock is free between the events of one command. Input:
  two `spec submit` at once. A saves an accepted `spec.submitted` and
  frees the lock. B takes the lock, sees `submitted`, and saves
  `review.started` for round 1, as R20 allows. A then takes the lock for
  its own `review.started`, and R2 refuses the move (`in_review`). R20's
  messages cover only the first check. The spec names no message or
  exit code for A. A worker must guess.
Fix: name it, e.g. R20's `in_review` error, with R10's last sentence
  `The steps before it were saved; spec.md is in_review.`

### F7. The order of the first checks for the new commands is not named

Level: medium
Where: spec.md → Requirements → R11, R12, R20, R32, R33
Breaks: nothing
Why: R20 starts its order at R11. 0001's earlier checks (not in a git
  repo, no commits, invalid name, no slice with that name: 0001 R12,
  R13, R14, R33) are named only for `slice create`, `slice status`, and
  `slice list`. Input: `sliceflow spec validate no-such-slice`: R11's
  `has no worktree` message, or 0001 R33's
  `Error: no slice named no-such-slice. See all slices: sliceflow slice list`?
  Input: `sliceflow slice title "My Slice" x`: 0001 R14's message, or
  R11's? `spec validate` and `slice title` have no order at all, e.g.
  R33's empty title before or after R11 and the lock.
Fix: one line for these commands: 0001 R12, R13, R14 (on the name), and
  R33 first, then R11, then the rest.

### F8. "The first good event only" for `slice.created` can be read two ways

Level: medium
Where: spec.md → Requirements → R2 (table), R4
Breaks: nothing
Why: Input: line 1 is a good `slice.titled`, line 2 a `slice.created`.
  Read as "only when it is the first good line", `slice.created` is
  skipped and the phase is `unknown`. Read as "only the first good
  `slice.created`", line 1 is skipped with R4's
  `comes before slice.created`, and the phase is `planning`. R4's second
  message suggests the second reading, but a worker must guess.
Fix: write "the first good `slice.created` only; a later one is not
  allowed".

### F9. Which lines count for the new seq, the seq check, and the next round

Level: medium
Where: spec.md → Requirements → R9, R25, R4
Breaks: nothing
Why: R9 adds "the one new line", but does not say its `seq`. 0001's R19
  says 1 more each time, but a committed log can end with a bad line or
  a line that R4 skipped. Input: lines with `seq` 1 and 2, then a line
  that is not valid JSON: is the new `seq` 3 or 4? Also unclear: whether
  a line R4 skipped counts as "the last good line" for 0001's R27, and
  whether R25's "highest round of spec.md in the log" counts a
  `review.started` that R4 skipped.
Fix: name the rule for each, in R9 and R25.

### F10. The lock: an unreadable lock file, and removing it after a failure

Level: low
Where: spec.md → Requirements → R6, R7, R10
Breaks: nothing
Why: R7 prints the pid and time read from the lock file. A lock file
  left empty by a crash right after it was made gives no values to
  print. R6 says the lock is removed after the event is saved. It does
  not say so after a failed commit, a refused move, or Ctrl-C (R10).
  SD10's cost names only a crash.
Fix: say the lock is removed on every way out but a hard kill, and what
  R7 prints when the file cannot be read.

### F11. R9 says how the log is written

Level: low
Where: spec.md → Requirements → R9
Breaks: nothing
Why: "a temporary file in the same folder, then renames that file over
  the log" is how it is built. The template says that goes in the plan.
  What the user gets is that the log is never half written.
Fix: keep the rule (the new log is the committed log plus one line, and
  never half written), and move the method to the plan.

### F12. Scenario coverage: R19 has none, and S8 names R12

Level: low
Where: spec.md → Requirements → R19, spec.md → User stories → S8
Breaks: nothing
Why: R19 is an error a user sees (`Error: <path> not found. ...`), and no
  scenario shows it. The template says every requirement a user can see
  is in a scenario. S8 gives the name, but its Shows: names R12, which
  is about running with no name.
Fix: add a scenario for R19, and drop R12 from S8's Shows:.

### F13. R15 and R16 leave three inputs open

Level: low
Where: spec.md → Requirements → R15, R16
Breaks: nothing
Why: R15 does not say which forms count as a list for `ideas`, e.g.
  `ideas: []`, `ideas:` with `  - x` lines, or `ideas: [a, b]`. R15 has a
  message for a Shows: line that names a removed requirement, but none
  for a Rule: line that does. R16's message holds the text "from `{{`
  to `}}`", but a `{{` may have no `}}` on its line.

### F14. Out of scope misses two items given to 0002

Level: low
Where: spec.md → Out of scope
Breaks: nothing
Why: Design 2.2 ("History in git") says the names of the git wrapper
  commands (show the spec as approved, the difference between two
  revisions) settle in slice 0002. The backlog row for the title says
  "and maybe showing the title in `spec.md`", and the roadmap gives that
  row to 0002. The spec neither does them nor says where they go.
Fix: add both to Out of scope, with their place.

### F15. `review.finished` replaces the design's `review.submitted` without a decision

Level: low
Where: spec.md → Requirements → R2, R5, R27, spec.md → Decisions
Breaks: docs/design/sliceflow-design.md 2.2
Why: design 2.2 ("History in git") lists the spec's events as
  `spec.submitted`, `review.started`, `review.submitted`, ... The spec
  uses `review.finished`, a stored event name, and no decision says
  why.
Fix: add a short decision, so the human sees the change at approval.

### F16. R25's `commit` in the resume path

Level: low
Where: spec.md → Requirements → R25, R20, spec.md → User stories → S23
Breaks: nothing
Why: R25's `commit` is "the full id of the commit R21 made". When
  `spec.md` is `submitted` (S23), R21 does not run in that command.
Fix: "the commit that saved the last accepted `spec.submitted`".

### F17. SD6's Not chosen: gives no reason

Level: low
Where: spec.md → Decisions → SD6
Breaks: nothing
Why: the template asks for the other option "and why not". The reason
  is in Why:, but Not chosen: does not say it.

## Requirements check

None.

## Needs human decision

- **Q1.** What does sliceflow do when the committed log's last line is
  half written, or has no `\n`? 0001's Out of scope gives "Adding an
  event to an existing log: the lock, and cutting off a half-written
  last line" to the first slice that adds a second event, which is this
  one. Design 1.13 says such a line is skipped when reading, and
  reported. This spec has the lock (R6), but neither handles that line
  nor moves it to Out of scope. R9 adds one new line to the committed
  log.
  Guess: cut off the last line when it has no `\n`, or keep it and add a
  `\n` first (it is then skipped as `not valid JSON`), or join the new
  line to it; and which warning, if any, the command prints.
