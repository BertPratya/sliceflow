# Status: 0001-slice-and-events

- Phase: planning (blocked)
- Branch: slice/0001-slice-and-events
- Updated: 2026-10-24 04:59 UTC

## Needs human

- spec review 3 is at the round limit with 1 medium and 2 low findings: the human decides how to go on.

## Documents

| Doc | Status | Round | Revision | Approved |
|---|---|---|---|---|
| spec.md | draft | 3 | - | - |

## Tasks

None.

## Findings

Open: 3 · Waiting for verify: 0 · Closed: 18

### Doc reviews

| ID | Level | Title | Status |
|---|---|---|---|
| spec-3/F1 | medium | Warning items: drop Warning: and <name>: , the R23 item | open |
| spec-3/F2 | low | R37: with no slices, only slices is [] | open |
| spec-3/F3 | low | R26: a bad at, by the form of R19 | open |
| spec-2/F1 | medium | R9 vs S1 and R10: what the spec path is relative to | verified |
| spec-2/F2 | medium | The base-branch warning: its form, in --json, and the ! mark | verified |
| spec-2/F3 | medium | slice list --json with no slices | verified |
| spec-1/F1 | medium | Order of the checks, and which place R14 names | verified |
| spec-1/F2 | medium | Built-in template vs R7 and SD12; created date form | verified |
| spec-1/F3 | medium | Relative paths in the output: relative to what | verified |
| spec-1/F4 | medium | When a slice exists, in progress, merged | verified |
| spec-1/F5 | medium | Base branch and errors for status and list | verified |
| spec-1/F6 | medium | Exact warning texts; wrong types, bad data | verified |
| spec-1/F7 | medium | Does a kept line with a wrong seq count | verified |
| spec-1/F8 | medium | Message for wrong use; --help | verified |
| spec-1/F9 | medium | Git hooks and signing on the slice commit | verified |
| spec-1/F10 | low | List column widths | verified |
| spec-1/F11 | low | Short commit length | verified |
| spec-1/F12 | low | SD13 is about how it is built | wontfix: why: the human's decision, and Rule: none already says it changes no behavior |
| spec-1/F13 | low | Committed log is never used | verified |
| spec-1/F14 | low | S12 may not make git fail | verified |
| spec-1/F15 | low | No .gitignore when .sliceflow/ exists without one | verified |

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

- 2026-10-24 04:59 · lead (claude) · spec at the review round limit: planning (blocked)
- 2026-10-24 04:59 · doc reviewer · spec review 3 handed in: 5 verified, 1 new medium, 2 new low
- 2026-10-24 04:57 · lead (claude) · spec review round 3 started
- 2026-10-24 04:57 · lead (claude) · spec fixed for spec-1 F6, F10 and spec-2 F1-F3; slice list --json is now an object
- 2026-10-24 04:57 · human · yes to the lead's fixes for spec review 2
- 2026-10-24 04:56 · doc reviewer · spec review 2 handed in: 12 verified, 2 reopened (F6, F10), 3 new medium
- 2026-10-24 04:53 · lead (claude) · spec review round 2 started
- 2026-10-24 04:53 · lead (claude) · spec fixed for spec-1 F1-F11, F13-F15; F12 left as is; requirements renumbered (now R1-R41)
- 2026-10-24 04:53 · human · yes to the lead's fixes for spec review 1
- 2026-10-24 04:50 · doc reviewer · spec review 1 handed in: high 0, medium 9, low 6
- 2026-10-24 04:47 · lead (claude) · spec review round 1 started
- 2026-10-24 04:47 · human · Q4: paths with / on every system; open questions: none
- 2026-10-24 04:45 · human · Q1: remove the title from the spec, a title command later; Q2: keep one slice; Q3: merged slices from the base branch
- 2026-10-24 04:38 · lead (claude) · spec.md written from the brainstorm; open questions Q1-Q4
- 2026-10-23 07:42 · lead (claude) · slice created
