# Status: 0002-spec-and-state

- Phase: implementing
- Branch: slice/0002-spec-and-state
- After: 0001-slice-and-events
- Updated: 2026-10-24 17:12 UTC

## Needs human

None.

## Documents

| Doc | Status | Round | Revision | Approved |
|---|---|---|---|---|
| spec.md | approved | 3 | R001 | 2026-10-24 · human · sliceflow/0002-spec-and-state/planning/R001 |
| plan.md | approved | 5 | R001 | 2026-10-24 · human · sliceflow/0002-spec-and-state/planning/R001 |
| preview.md | approved | with plan.md | R001 | 2026-10-24 · human · sliceflow/0002-spec-and-state/planning/R001 |
| tasks.md | approved | with plan.md | R001 | 2026-10-24 · human · sliceflow/0002-spec-and-state/planning/R001 |

## Tasks

3 of 17 done · 1 doing · 13 todo

| # | Name | Risk | Status | Attempt | Round | Depends on | Covers |
|---|---|---|---|---|---|---|---|
| 01 | eventlog-events | core | done | 0 | 2 | - | R5 |
| 02 | state-machine | core | done | 0 | 2 | 01 | R1, R2, R3, R4 |
| 03 | spec-reader | core | done | 0 | 3 | - | R18 |
| 04 | spec-errors | core | doing | 0 | 0 | 03 | R15, R18 |
| 05 | spec-warnings | core | todo | 0 | 0 | 04 | R16, R17, R18 |
| 06 | status-core | core | todo | 0 | 0 | 02 | R1, R3, R29, R30, R34 |
| 07 | status-spec | normal | todo | 0 | 0 | 06 | R4, R29, R31 |
| 08 | cli-groups-and-template | core | todo | 0 | 0 | - | R35, R36 |
| 09 | validate-command | core | todo | 0 | 0 | 05, 06, 08 | R11, R12, R13, R14, R18, R19, R36, R37 |
| 10 | lock | core | todo | 0 | 0 | - | R6, R7 |
| 11 | save-event | core | todo | 0 | 0 | 01, 02, 06, 10 | R2, R6, R8, R9, R10 |
| 12 | title-command | normal | todo | 0 | 0 | 08, 09, 11 | R32, R33, R34, R36 |
| 13 | reviewer | core | todo | 0 | 0 | - | R25, R26, R27 |
| 14 | submit-core | core | todo | 0 | 0 | 05, 09, 11 | R12, R19, R20, R21, R22, R23 |
| 15 | review-core | core | todo | 0 | 0 | 09, 11, 13 | R10, R12, R24, R25, R27, R28 |
| 16 | submit-review-commands | normal | todo | 0 | 0 | 07, 12, 14, 15 | R20, R21, R22, R23, R24, R25, R26, R27, R28, R29, R30, R37 |
| 17 | submit-review-failures | normal | todo | 0 | 0 | 16 | R6, R7, R8, R9, R10, R24, R25, R29, R37 |

## Findings

Open: 0 · Waiting for verify: 4 · Closed: 51

### Doc reviews

| ID | Level | Title | Status |
|---|---|---|---|
| plan-5/F1 | medium | A failed review.finished commit has no brief and no test | fixed (not re-reviewed: the human approved) |
| plan-5/F2 | low | Three leftover lines in spec.md | fixed |
| plan-5/F3 | low | plan.md leaves out spec review; wrong task for Ctrl-C | fixed |
| plan-5/F4 | low | Stale task numbers; R20's error index | fixed |
| plan-4/F1 | low | R31: next for a merged slice | verified (fixed) |
| plan-4/F2 | low | R29: refused mark on a merged slice | verified (fixed) |
| plan-4/F3 | low | Task 09: how openSlice finds a merged slice | verified (fixed) |
| plan-4/F4 | low | The Make one: path works only from the root | verified (fixed) |
| plan-4/F5 | low | Task 07's brief reason names S1 to S4 | verified (fixed) |
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

| ID | Level | Task | Title | Status |
|---|---|---|---|---|
| task-02-1/F1 | low | 02 | No test checks the review's reviewer, file, commit, or the created title | verified |
| task-03-2/F1 | medium | 03 | A lone backtick in one item hides the next item | verified |
| task-03-1/F1 | high | 03 | An ideas list item with no space before - is dropped | verified |
| task-03-1/F2 | high | 03 | Markers like **R 7.** are neither items nor bad markers | verified |
| task-03-1/F3 | medium | 03 | An unclosed front matter cannot give the no front matter error | verified |
| task-03-1/F4 | medium | 03 | A section with only code counts as empty | verified |
| task-03-1/F5 | low | 03 | Inline code is paired within one line only | verified |
| task-01-1/F1 | medium | 01 | The S15 line (accepted) is not tested | verified |
| task-01-1/F2 | low | 01 | TestParseNewEvents checks the warnings too loosely | verified |

### Integration review

None.

## Lead's decisions

- Task 03, task-03-1/F3: a front matter with no closing `---` gives `HasFront` false, so task 04 reports R15's `no front matter` error. Why: R15 asks for front matter between two `---` lines, and `Doc` keeps the fields of tasks.md (no new field).
- Task 02, task-02-1/F1 (low): fixed in a worker round, not left. Why: tasks 06, 11, and 15 rely on the review's reviewer, file, and commit, and on the title from `slice.created`; no other task's tests check them in `state`.

## Preview

None.

## Log

Newest first.

- 2026-10-24 17:12 · lead (claude) · task 04 started: worker round 1
- 2026-10-24 17:12 · lead (claude) · task 02 done
- 2026-10-24 17:12 · code-reviewer (claude opus) · task review 2 of task 02 handed in: task-02-1/F1 verified; no new findings
- 2026-10-24 17:08 · lead (claude) · task review 2 of task 02 started
- 2026-10-24 17:08 · lead (claude) · task 02 fixes committed (e309867); the checks pass
- 2026-10-24 17:08 · worker (claude sonnet) · task 02 round 2 handed in: F1 fixed (the lead asked once for values that differ per round)
- 2026-10-24 17:01 · lead (claude) · task 02 worker round 2 started, to fix the low task-02-1/F1
- 2026-10-24 17:01 · code-reviewer (claude opus) · task review 1 of task 02 handed in: 1 low
- 2026-10-24 16:56 · lead (claude) · task review 1 of task 02 started
- 2026-10-24 16:56 · lead (claude) · task 02 committed (804587e); the checks pass
- 2026-10-24 16:56 · worker (claude sonnet) · task 02 round 1 handed in
- 2026-10-24 16:51 · lead (claude) · task 02 started: worker round 1
- 2026-10-24 16:51 · lead (claude) · task 03 done
- 2026-10-24 16:51 · code-reviewer (claude opus) · task review 3 of task 03 handed in: task-03-2/F1 verified; no new findings
- 2026-10-24 16:46 · lead (claude) · task review 3 of task 03 started (the last round allowed)
- 2026-10-24 16:46 · lead (claude) · task 03 fixes committed (2b91ce9); the checks pass
- 2026-10-24 16:46 · worker (claude sonnet) · task 03 round 3 handed in: task-03-2/F1 fixed
- 2026-10-24 16:42 · lead (claude) · task 03 worker round 3 started, with review task-03-2
- 2026-10-24 16:42 · code-reviewer (claude opus) · task review 2 of task 03 handed in: task-03-1 F1-F5 verified; 1 new medium
- 2026-10-24 16:36 · lead (claude) · task review 2 of task 03 started
- 2026-10-24 16:36 · lead (claude) · task 03 fixes committed (31a6ab8); the checks pass
- 2026-10-24 16:36 · worker (claude sonnet) · task 03 round 2 handed in: F1-F5 fixed
- 2026-10-24 16:31 · lead (claude) · task 03 worker round 2 started, with review task-03-1
- 2026-10-24 16:31 · lead (claude) · decided task-03-1/F3: an unclosed front matter gives HasFront false
- 2026-10-24 16:31 · code-reviewer (claude opus) · task review 1 of task 03 handed in: 2 high, 2 medium, 1 low
- 2026-10-24 16:24 · lead (claude) · task review 1 of task 03 started
- 2026-10-24 16:24 · lead (claude) · task 03 committed (5b38b6f); the checks pass
- 2026-10-24 16:24 · worker (claude sonnet) · task 03 round 1 handed in
- 2026-10-24 16:17 · lead (claude) · task 03 started: worker round 1
- 2026-10-24 16:17 · lead (claude) · task 01 done
- 2026-10-24 16:17 · code-reviewer (claude opus) · task review 2 of task 01 handed in: task-01-1 F1, F2 verified; no new findings
- 2026-10-24 16:12 · lead (claude) · task review 2 of task 01 started
- 2026-10-24 16:12 · lead (claude) · task 01 fixes committed (706baa5); the checks pass
- 2026-10-24 16:12 · worker (claude sonnet) · task 01 round 2 handed in: F1, F2 fixed
- 2026-10-24 16:08 · lead (claude) · task 01 worker round 2 started, with review task-01-1
- 2026-10-24 16:08 · code-reviewer (claude opus) · task review 1 of task 01 handed in: 1 medium, 1 low
- 2026-10-24 16:00 · lead (claude) · task review 1 of task 01 started
- 2026-10-24 16:00 · lead (claude) · task 01 committed (ffc681e); the checks pass
- 2026-10-24 16:00 · worker (claude sonnet) · task 01 round 1 handed in
- 2026-10-24 15:55 · lead (claude) · task 01 started: worker round 1
- 2026-10-24 15:49 · lead (claude) · main's checks pass (base-check at be47b4a); phase implementing
- 2026-10-24 15:46 · human · planning approved, R001: "yes, okay, and I approve" (the plan-5/F1 fix went in with this approval, with no separate review)
- 2026-10-24 15:46 · lead (claude) · plan-5/F1 fixed: task 15 sets R10's last sentence for a failed review.finished, with tests in tasks 15 and 17
- 2026-10-24 15:40 · lead (claude) · plan-5 lows fixed; documents back to draft
- 2026-10-24 15:40 · doc-reviewer (claude opus) · change review 5 handed in: plan-4 all verified; 1 medium, 3 low
- 2026-10-24 15:37 · lead (claude) · change review 5 started: the review as its own command
- 2026-10-24 15:37 · lead (claude) · human: the review is its own command, `sliceflow spec review`; `spec submit` ends at submitted; a general `sliceflow run` goes to the backlog. Spec (Stories 3 and 4, R20, R23, R24, SD8, SD13), plan (Flows 3 to 5), tasks (14 to 17), preview (P3, P4) changed
- 2026-10-24 14:50 · lead (claude) · human: the Make one: line prints the full path (R11, S12); the review question goes to a brainstorm first
- 2026-10-24 14:48 · lead (claude) · plan-4 lows F1, F2, F3, F5 fixed; F4 and the human's question (review as its own node) are open; documents back to draft
- 2026-10-24 14:48 · doc-reviewer (claude opus) · change review 4 handed in: 5 low
- 2026-10-24 14:44 · lead (claude) · change review 4 started: only the human's two changes before approval
- 2026-10-24 14:44 · lead (claude) · human, before approval: R11 gets a 'Make one:' line and a merged-slice error; a merged slice shows Next: - (R29, new S5; S5-S31 renumbered to S6-S32). Documents back to draft
- 2026-10-24 13:53 · lead (claude) · plan-3 low fixed; spec, plan, tasks, preview review_passed; waiting for the human's approval
- 2026-10-24 13:53 · doc-reviewer (claude opus) · change review 3 handed in: all reopened findings verified; 1 new low
- 2026-10-24 13:51 · lead (claude) · change review 3 of the plan started: only the fixes since round 2 (the last round)
- 2026-10-24 13:51 · lead (claude) · human: "yes": Prepare also returns the commit message; reasons for the long briefs of tasks 14, 15, 16
- 2026-10-24 13:32 · lead (claude) · plan-2 lows fixed (plan-1/F7, F10, plan-2/F1); documents back to draft
- 2026-10-24 13:32 · doc-reviewer (claude opus) · plan review 2 handed in: nothing for the human; plan-1/F1, F3 reopened (medium), F7, F10 reopened (low); 1 new low
- 2026-10-24 13:28 · lead (claude) · plan review 2 started
- 2026-10-24 13:28 · lead (claude) · human: "yes" to the plan-1 F1-F4 fixes: a Prepare step under the lock; R12 and R19 in task 14; task 15 split into 15 and 16; Allowed returns a string
- 2026-10-24 13:22 · lead (claude) · human: plan-1 Q1 option 1, template show works outside a git repo (spec R35)
- 2026-10-24 13:15 · lead (claude) · plan-1 lows fixed (F5-F12, two parts wait for F3); documents back to draft
- 2026-10-24 13:15 · doc-reviewer (claude opus) · plan review 1 handed in: 1 high, 4 medium, 8 low
- 2026-10-24 13:03 · lead (claude) · plan review 1 started: spec, plan, tasks, and preview together
- 2026-10-24 13:03 · lead (claude) · tasks written: 15 tasks; plan Open questions: None.
- 2026-10-24 12:59 · lead (claude) · human: plan Q1 option 1, a value with `{{` is not checked, only warned (spec R18); the plan review checks the spec change
- 2026-10-24 12:57 · lead (claude) · plan flows, testing, and preview written; plan Q1 shows a gap in the spec, so spec.md is back to draft
- 2026-10-24 12:54 · lead (claude) · human: keep the 10 s lock wait (a slow hook is a listed risk). Shared rules for the lock, saving, undo, and crash leftovers; TD1, TD2
- 2026-10-24 12:51 · lead (claude) · human: the reviewer writes the review file, and sliceflow reads the result from it (option 1). Interfaces written
- 2026-10-24 12:47 · lead (claude) · human: lock file is one JSON line (pid, since); a leftover lock is removed by hand until 0007 (roadmap); reads take no lock. Data written
- 2026-10-24 12:39 · lead (claude) · human: complex; package layout option 1 (three new packages: state, spec, review); Approach written
- 2026-10-24 12:34 · lead (claude) · plan started: plan.md and preview.md copied; preview's fixed checks P1-P5
- 2026-10-24 12:34 · lead (claude) · spec.md review_passed
- 2026-10-24 12:34 · doc-reviewer (claude opus) · change review 3 handed in: spec-2 all verified; no new findings
- 2026-10-24 12:32 · lead (claude) · spec review 3 started: a change review of the spec-2 fixes only, as for 0001
- 2026-10-24 12:32 · lead (claude) · human: "yes" to the spec-2 F1, F2 fixes
- 2026-10-24 12:30 · lead (claude) · spec-2 lows fixed (F3, F4), and S23 uses a commit-msg hook (the reviewer's note); spec.md back to draft
- 2026-10-24 12:30 · doc-reviewer (claude opus) · spec review 2 handed in: spec-1 all verified; 0 high, 2 medium, 2 low
- 2026-10-24 12:24 · lead (claude) · spec review 2 started, also against 0001's merged code
- 2026-10-24 12:24 · lead (claude) · human: "yes" to the F1-F9 fixes; Q1 option 1, keep a bad last line (SD15)
- 2026-10-24 12:22 · lead (claude) · spec-1 lows fixed (F10-F14, F16, F17), F15 left; spec.md back to draft
- 2026-10-24 12:22 · lead (claude) · updated from main after 0001 merged
- 2026-10-24 12:22 · doc-reviewer (claude opus) · spec review 1 handed in: 1 high, 9 medium, 8 low
- 2026-10-24 12:13 · lead (claude) · spec review 1 started
- 2026-10-24 12:13 · lead (claude) · human: problems on stdout (SD14); Open questions: None.
- 2026-10-24 12:12 · lead (claude) · human: option 2, spec submit goes on from submitted (R10, R20, R24, R29, SD13, S22-S23)
- 2026-10-24 12:08 · lead (claude) · human: keep one slice (old Q1); review file with all 7 template sections (old Q3, R26); renumbered: Q1 stuck submit, Q2 stdout
- 2026-10-24 11:41 · lead (claude) · spec written after the human's yes to the playback; Q1-Q4 open
- 2026-10-24 08:12 · lead (claude) · slice created
