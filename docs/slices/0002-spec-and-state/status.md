# Status: 0002-spec-and-state

- Phase: planning
- Branch: slice/0002-spec-and-state
- After: 0001-slice-and-events
- Updated: 2026-10-24 21:20 UTC

## Needs human

- Review as its own node, or inside `spec submit`: brainstorm with the human

## Documents

| Doc | Status | Round | Revision | Approved |
|---|---|---|---|---|
| spec.md | draft | 3 | - | - |
| plan.md | draft | 4 | - | - |
| preview.md | draft | with plan.md | - | - |
| tasks.md | draft | with plan.md | - | - |

## Tasks

0 of 16 done · 16 todo

| # | Name | Risk | Status | Attempt | Round | Depends on | Covers |
|---|---|---|---|---|---|---|---|
| 01 | eventlog-events | core | todo | 0 | 0 | - | R5 |
| 02 | state-machine | core | todo | 0 | 0 | 01 | R1, R2, R3, R4 |
| 03 | spec-reader | core | todo | 0 | 0 | - | R18 |
| 04 | spec-errors | core | todo | 0 | 0 | 03 | R15, R18 |
| 05 | spec-warnings | core | todo | 0 | 0 | 04 | R16, R17, R18 |
| 06 | status-core | core | todo | 0 | 0 | 02 | R1, R3, R29, R30, R34 |
| 07 | status-spec | normal | todo | 0 | 0 | 06 | R4, R29, R31 |
| 08 | cli-groups-and-template | core | todo | 0 | 0 | - | R35, R36 |
| 09 | validate-command | core | todo | 0 | 0 | 05, 06, 08 | R11, R12, R13, R14, R18, R19, R36, R37 |
| 10 | lock | core | todo | 0 | 0 | - | R6, R7 |
| 11 | save-event | core | todo | 0 | 0 | 01, 02, 06, 10 | R2, R6, R8, R9, R10 |
| 12 | title-command | normal | todo | 0 | 0 | 08, 09, 11 | R32, R33, R34, R36 |
| 13 | reviewer | core | todo | 0 | 0 | - | R25, R26, R27 |
| 14 | submit-core | core | todo | 0 | 0 | 05, 09, 11, 13 | R12, R19, R20, R21, R22, R23, R24, R25, R27, R28 |
| 15 | submit-command | normal | todo | 0 | 0 | 07, 12, 14 | R20, R21, R22, R23, R25, R26, R27, R28, R29, R30, R37 |
| 16 | submit-failures | normal | todo | 0 | 0 | 15 | R6, R7, R8, R9, R10, R20, R24, R29 |

## Findings

Open: 0 · Waiting for verify: 5 · Closed: 37

### Doc reviews

| ID | Level | Title | Status |
|---|---|---|---|
| plan-4/F1 | low | R31: next for a merged slice | fixed |
| plan-4/F2 | low | R29: refused mark on a merged slice | fixed |
| plan-4/F3 | low | Task 09: how openSlice finds a merged slice | fixed |
| plan-4/F4 | low | The Make one: path works only from the root | fixed: full path (the human's choice) |
| plan-4/F5 | low | Task 07's brief reason names S1 to S4 | fixed |
| plan-3/F1 | low | Task 07's Done in: line form | verified (fixed) |
| plan-2/F1 | low | Leftovers of the round-1 fixes | verified (fixed) |
| plan-1/Q1 | high | template show outside a git repo | verified (fixed) |
| plan-1/F1 | medium | saveEvent can't run the command's checks under its lock | verified (fixed) |
| plan-1/F2 | medium | R12 and R19 for spec submit are in no task | verified (fixed) |
| plan-1/F3 | medium | Six briefs over 1,500 words; task 15 too big | verified (fixed) |
| plan-1/F4 | medium | state.Allowed: error in the plan, string in task 02 | verified (fixed) |
| plan-1/F5 | low | git diff exit 128 for a missing commit | verified (fixed) |
| plan-1/F6 | low | Gaps in Done when | verified (fixed) |
| plan-1/F7 | low | Testing and the tasks disagree | verified (fixed) |
| plan-1/F8 | low | Tasks 05, 08, 09 should be core | verified (fixed) |
| plan-1/F9 | low | Task 06's Phase and Title line | verified (fixed) |
| plan-1/F10 | low | Approach's file and dependency lines; Request.File | verified (fixed) |
| plan-1/F11 | low | Plan touches; Parked | verified (fixed) |
| plan-1/F12 | low | Preview P2's exact line | verified (fixed) |
| spec-2/F1 | medium | An indented ideas item is not a list | verified (fixed) |
| spec-2/F2 | medium | R11's order makes validate take the lock | verified (fixed) |
| spec-2/F3 | low | "last line read without error" is a new term | verified (fixed) |
| spec-2/F4 | low | The third assumption is now confirmed | verified (fixed) |
| spec-1/Q1 | high | A half-written last line of the log | verified (fixed) |
| spec-1/F1 | medium | Template comments are not skipped | verified (fixed) |
| spec-1/F2 | medium | S28's template start is not the real one | verified (fixed) |
| spec-1/F3 | medium | A failed commit leaves the review file behind | verified (fixed) |
| spec-1/F4 | medium | slice title in phase unknown | verified (fixed) |
| spec-1/F5 | medium | Step lines before a failure on stdout | verified (fixed) |
| spec-1/F6 | medium | Two submits at once: no message | verified (fixed) |
| spec-1/F7 | medium | Order of the first checks | verified (fixed) |
| spec-1/F8 | medium | "the first good event only" reads two ways | verified (fixed) |
| spec-1/F9 | medium | seq and round after skipped lines | verified (fixed) |
| spec-1/F10 | low | Lock: unreadable file, removed after failure | verified (fixed) |
| spec-1/F11 | low | R9 says how, not what | verified (fixed) |
| spec-1/F12 | low | R19 has no scenario; S8's Shows | verified (fixed) |
| spec-1/F13 | low | Open inputs in R15 and R16 | verified (fixed) |
| spec-1/F14 | low | Two Out of scope items missing | verified (fixed) |
| spec-1/F15 | low | review.finished vs design 2.2 | verified (wontfix) |
| spec-1/F16 | low | Resume path: which commit | verified (fixed) |
| spec-1/F17 | low | SD6 Not chosen has no reason | verified (fixed) |

### Task reviews

None.

### Integration review

None.

## Lead's decisions

None.

## Preview

None.

## Log

Newest first.

- 2026-10-24 21:30 · lead (claude) · human: the Make one: line prints the full path (R11, S12); the review question goes to a brainstorm first
- 2026-10-24 21:20 · lead (claude) · plan-4 lows F1, F2, F3, F5 fixed; F4 and the human's question (review as its own node) are open; documents back to draft
- 2026-10-24 21:10 · doc-reviewer (claude opus) · change review 4 handed in: 5 low
- 2026-10-24 20:50 · lead (claude) · change review 4 started: only the human's two changes before approval
- 2026-10-24 20:45 · lead (claude) · human, before approval: R11 gets a 'Make one:' line and a merged-slice error; a merged slice shows Next: - (R29, new S5; S5-S31 renumbered to S6-S32). Documents back to draft
- 2026-10-24 20:15 · lead (claude) · plan-3 low fixed; spec, plan, tasks, preview review_passed; waiting for the human's approval
- 2026-10-24 20:10 · doc-reviewer (claude opus) · change review 3 handed in: all reopened findings verified; 1 new low
- 2026-10-24 19:55 · lead (claude) · change review 3 of the plan started: only the fixes since round 2 (the last round)
- 2026-10-24 19:50 · lead (claude) · human: "yes": Prepare also returns the commit message; reasons for the long briefs of tasks 14, 15, 16
- 2026-10-24 19:40 · lead (claude) · plan-2 lows fixed (plan-1/F7, F10, plan-2/F1); documents back to draft
- 2026-10-24 19:35 · doc-reviewer (claude opus) · plan review 2 handed in: nothing for the human; plan-1/F1, F3 reopened (medium), F7, F10 reopened (low); 1 new low
- 2026-10-24 19:10 · lead (claude) · plan review 2 started
- 2026-10-24 19:05 · lead (claude) · human: "yes" to the plan-1 F1-F4 fixes: a Prepare step under the lock; R12 and R19 in task 14; task 15 split into 15 and 16; Allowed returns a string
- 2026-10-24 18:50 · lead (claude) · human: plan-1 Q1 option 1, template show works outside a git repo (spec R35)
- 2026-10-24 18:40 · lead (claude) · plan-1 lows fixed (F5-F12, two parts wait for F3); documents back to draft
- 2026-10-24 18:30 · doc-reviewer (claude opus) · plan review 1 handed in: 1 high, 4 medium, 8 low
- 2026-10-24 18:00 · lead (claude) · plan review 1 started: spec, plan, tasks, and preview together
- 2026-10-24 17:55 · lead (claude) · tasks written: 15 tasks; plan Open questions: None.
- 2026-10-24 17:25 · lead (claude) · human: plan Q1 option 1, a value with `{{` is not checked, only warned (spec R18); the plan review checks the spec change
- 2026-10-24 17:10 · lead (claude) · plan flows, testing, and preview written; plan Q1 shows a gap in the spec, so spec.md is back to draft
- 2026-10-24 16:30 · lead (claude) · human: keep the 10 s lock wait (a slow hook is a listed risk). Shared rules for the lock, saving, undo, and crash leftovers; TD1, TD2
- 2026-10-24 16:10 · lead (claude) · human: the reviewer writes the review file, and sliceflow reads the result from it (option 1). Interfaces written
- 2026-10-24 15:50 · lead (claude) · human: lock file is one JSON line (pid, since); a leftover lock is removed by hand until 0007 (roadmap); reads take no lock. Data written
- 2026-10-24 15:20 · lead (claude) · human: complex; package layout option 1 (three new packages: state, spec, review); Approach written
- 2026-10-24 15:00 · lead (claude) · plan started: plan.md and preview.md copied; preview's fixed checks P1-P5
- 2026-10-24 14:55 · lead (claude) · spec.md review_passed
- 2026-10-24 14:50 · doc-reviewer (claude opus) · change review 3 handed in: spec-2 all verified; no new findings
- 2026-10-24 14:40 · lead (claude) · spec review 3 started: a change review of the spec-2 fixes only, as for 0001
- 2026-10-24 14:35 · lead (claude) · human: "yes" to the spec-2 F1, F2 fixes
- 2026-10-24 14:20 · lead (claude) · spec-2 lows fixed (F3, F4), and S23 uses a commit-msg hook (the reviewer's note); spec.md back to draft
- 2026-10-24 14:15 · doc-reviewer (claude opus) · spec review 2 handed in: spec-1 all verified; 0 high, 2 medium, 2 low
- 2026-10-24 13:40 · lead (claude) · spec review 2 started, also against 0001's merged code
- 2026-10-24 13:35 · lead (claude) · human: "yes" to the F1-F9 fixes; Q1 option 1, keep a bad last line (SD15)
- 2026-10-24 13:10 · lead (claude) · spec-1 lows fixed (F10-F14, F16, F17), F15 left; spec.md back to draft
- 2026-10-24 13:05 · lead (claude) · updated from main after 0001 merged
- 2026-10-24 13:00 · doc-reviewer (claude opus) · spec review 1 handed in: 1 high, 9 medium, 8 low
- 2026-10-24 12:30 · lead (claude) · spec review 1 started
- 2026-10-24 12:30 · lead (claude) · human: problems on stdout (SD14); Open questions: None.
- 2026-10-24 12:20 · lead (claude) · human: option 2, spec submit goes on from submitted (R10, R20, R24, R29, SD13, S22-S23)
- 2026-10-24 12:05 · lead (claude) · human: keep one slice (old Q1); review file with all 7 template sections (old Q3, R26); renumbered: Q1 stuck submit, Q2 stdout
- 2026-10-24 11:45 · lead (claude) · spec written after the human's yes to the playback; Q1-Q4 open
- 2026-10-24 08:12 · lead (claude) · slice created
