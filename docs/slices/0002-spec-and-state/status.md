# Status: 0002-spec-and-state

- Phase: planning
- Branch: slice/0002-spec-and-state
- After: 0001-slice-and-events
- Updated: 2026-10-24 15:00 UTC

## Needs human

None.

## Documents

| Doc | Status | Round | Revision | Approved |
|---|---|---|---|---|
| spec.md | review_passed | 3 | - | - |
| plan.md | draft | 0 | - | - |
| preview.md | draft | with plan.md | - | - |

## Tasks

None.

## Findings

Open: 0 · Waiting for verify: 0 · Closed: 22

### Doc reviews

| ID | Level | Title | Status |
|---|---|---|---|
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
