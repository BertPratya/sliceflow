# Status: 0001-slice-and-events

- Phase: planning
- Branch: slice/0001-slice-and-events
- Updated: 2026-10-24 07:28 UTC

## Needs human

- change-3/F1 (medium): the lead's fix waits for the human's yes

## Documents

| Doc | Status | Round | Revision | Approved |
|---|---|---|---|---|
| spec.md | draft | 3 | - | - |
| plan.md | draft | 3 | - | - |
| preview.md | draft | with plan.md | - | - |
| tasks.md | draft | with plan.md | - | - |

## Tasks

0 of 13 done · 13 todo

| # | Name | Risk | Status | Attempt | Round | Depends on | Covers |
|---|---|---|---|---|---|---|---|
| 01 | git-runner | core | todo | 0 | 0 | - | R40 |
| 02 | event-log | core | todo | 0 | 0 | 01 | R19, R25, R26, R27 |
| 03 | cli-skeleton | core | todo | 0 | 0 | 01, 04 | R28, R38, R39 |
| 04 | repo-and-names | core | todo | 0 | 0 | 01 | R2, R3, R12, R13, R14, R21, R22 |
| 05 | spec-template | normal | todo | 0 | 0 | 01 | R7 |
| 06 | create-core | core | todo | 0 | 0 | 02, 04, 05 | R1, R4, R5, R6, R7, R8, R19, R20 |
| 07 | create-command | normal | todo | 0 | 0 | 03, 06 | R1, R3, R5, R9, R10, R38, R39, R41 |
| 08 | create-checks | core | todo | 0 | 0 | 07 | R2, R11, R12, R13, R14, R15, R16, R17, R42 |
| 09 | create-undo | core | todo | 0 | 0 | 08 | R8, R18 |
| 10 | status-core | core | todo | 0 | 0 | 02, 04, 06 | R21, R23, R24, R25, R26, R33 |
| 11 | status-command | normal | todo | 0 | 0 | 03, 07, 10 | R22, R26, R28, R29, R30, R31, R32, R33, R42 |
| 12 | slice-list | normal | todo | 0 | 0 | 11 | R21, R22, R23, R28, R34, R35, R36, R37, R38, R42, R43 |
| 13 | no-network-check | normal | todo | 0 | 0 | 09, 12 | R40 |

## Findings

Open: 2 · Waiting for verify: 0 · Closed: 45

### Doc reviews

| ID | Level | Title | Status |
|---|---|---|---|
| change-3/F1 | medium | TestReasonFileError fails on Linux and macOS (ToSlash only converts the own separator) | open |
| change-3/F2 | low | Two periods when the reason ends with . | open |
| plan-3/F1 | medium | R18's system error reason: not in Flow 3 or task 09, path with \ on Windows | verified |
| plan-3/F2 | low | Task 09: how to get <command> from git.Error.Args | verified |
| plan-2/F1 | medium | Task 08's folder check misses Cross-platform paths | verified |
| plan-2/F2 | low | TestSpecFrontMatter passes with the ideas: example left in | verified |
| plan-2/F3 | low | Ctrl-C during FindRepo prints context canceled | verified |
| plan-2/F4 | low | R18: <command> for worktree add; a step that is not git | verified |
| plan-2/F5 | low | R38's short list leaves out the new wrong uses | verified |
| plan-1/F1 | medium | Status and List are each a type and a function | verified |
| plan-1/F2 | medium | Tasks 02 and 05 need go.mod from task 01 | verified |
| plan-1/F3 | medium | Interrupts: undo ctx, WaitDelay, the reason, Ctrl-C during checks | verified |
| plan-1/F4 | medium | Remote check matches any ref ending in /slice/<name> | verified |
| plan-1/F5 | medium | git add fails when the project ignores the files | verified |
| plan-1/F6 | medium | Key concept A slice exists vs R43 and Flow 4 | verified |
| plan-1/F7 | medium | TestSpecNoTitle: wrong placeholder, impossible body | verified |
| plan-1/F8 | medium | Cross-platform paths misses tasks 08-10; Tests rule reaches no task | verified |
| plan-1/F9 | medium | R42 proved only for create | verified |
| plan-1/F10 | medium | Briefs of tasks 06, 07, 09 over 1,500 words | verified |
| plan-1/F11 | low | TD3 both followed and Plan only | verified |
| plan-1/F12 | low | follows names flows and Data subsections | verified |
| plan-1/F13 | low | Some Proves: lines name a requirement not proved | verified |
| plan-1/F14 | low | e2e/ outside the gofmt check | verified |
| plan-1/F15 | low | R26: which reason wins with two faults | verified |
| plan-1/F16 | low | Flow 1: Create finds the repo, but gets a Repo | verified |
| plan-1/F17 | low | git makes .sliceflow/worktrees/, undo list misses it | verified |
| spec-3/F1 | medium | Warning items: drop Warning: and <name>: , the R23 item | verified |
| spec-3/F2 | low | R37: with no slices, only slices is [] | verified |
| spec-3/F3 | low | R26: a bad at, by the form of R19 | verified |
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

- 2026-10-24 07:28 · doc reviewer (claude opus) · change-3: the three fixes verified; new from the F1 fix: 1 medium, 1 low; nothing for the human to decide
- 2026-10-24 07:27 · lead (claude) · plan-3 fixed (F1, F2, plan-2/F2); change review started (reviews/change-3.md), it checks only these fixes
- 2026-10-24 07:27 · human · after the round limit: fix the three, then a change review of only the fixes (option 1); block removed
- 2026-10-24 07:23 · lead (claude) · round limit reached; slice marked planning (blocked); asking the human
- 2026-10-24 07:23 · doc reviewer (claude opus) · plan-3: 6 of 7 verified, plan-2/F2 reopened; new: 1 medium, 1 low; nothing for the human to decide
- 2026-10-24 07:21 · lead (claude) · plan review round 3 started (reviews/plan-3.md), the last round of the limit
- 2026-10-24 07:20 · lead (claude) · plan-2 fixed: task 08 follows Cross-platform paths; R38 only in tasks 03, 07, 12; size reasons in long briefs; lows F2-F5 and plan-1/F15
- 2026-10-24 07:20 · human · plan-2 fixes: yes
- 2026-10-24 07:15 · doc reviewer (claude opus) · plan-2: 15 of plan-1 verified, F10 and F15 reopened; new: 1 medium, 4 low; nothing for the human to decide; documents back to draft
- 2026-10-24 07:11 · lead (claude) · plan review round 2 started (reviews/plan-2.md)
- 2026-10-24 07:11 · lead (claude) · plan-1 fixed: spec R18 reason, R38 wrong use, key concept; plan F1-F17; tasks now 13 (create and status split), e2e in internal/e2e/
- 2026-10-24 07:11 · human · plan-1 Q1: bare commands show usage, a missing or extra name is wrong use (R38); Q2: first line as reason, the rest indented (R18); the lead's fixes: ok
- 2026-10-24 06:58 · doc reviewer (claude opus) · plan-1: 10 medium, 7 low, 2 questions for the human; documents back to draft
- 2026-10-24 06:52 · lead (claude) · plan review round 1 started (reviews/plan-1.md); spec.md, plan.md, tasks.md, preview.md in_review
- 2026-10-24 06:52 · lead (claude) · tasks.md written: 11 tasks; plan adds internal/testrepo for test repos
- 2026-10-24 06:48 · human · plan Q1: a field twice gets its own warning (R26); Q2: a slice/ branch with a bad name is skipped with a repo warning (R43); plan open questions: none
- 2026-10-24 06:46 · lead (claude) · plan part 5 (flows, testing) and preview written after the human's yes; reread found plan Q1-Q2
- 2026-10-24 06:46 · human · plan part 5, flows and testing: yes
- 2026-10-24 06:43 · human · plan part 4: undo, safety, and crash notes ok; Ctrl-C and SIGTERM undo like a failed step (R18); a hard kill is cleaned up later, with cancel in the backlog
- 2026-10-24 06:38 · human · plan part 3, interfaces: core API option 1, three plain functions
- 2026-10-24 06:35 · human · plan part 2, the data: yes
- 2026-10-24 06:32 · human · errors stay text on stderr with --json (R42, S29); JSON errors with codes go to the backlog; spec.md back to draft, the plan review checks it
- 2026-10-24 06:30 · human · plan part 1: package layout yes; module path github.com/BertPratya/sliceflow; flag package plus a small helper
- 2026-10-24 05:06 · lead (claude) · plan started: plan.md and preview.md copied; preview checks P1-P4 from the spec; depth: complex
- 2026-10-24 05:04 · lead (claude) · spec.md review_passed
- 2026-10-24 05:04 · doc reviewer · change review 2 handed in: spec-3/F1 verified; no findings
- 2026-10-24 05:03 · lead (claude) · change review 2 started (reviews/change-2.md)
- 2026-10-24 05:03 · human · "2": one more short check of the last fix
- 2026-10-24 05:01 · lead (claude) · spec-3/F1 finished: R37's warnings items written as in R30
- 2026-10-24 05:01 · doc reviewer · change review 1 handed in: F2, F3 verified; F1 reopened (R37's top-level warnings)
- 2026-10-24 05:00 · lead (claude) · spec fixed for spec-3 F1-F3; change review of only these fixes started (reviews/change-1.md)
- 2026-10-24 05:00 · human · "2": fix spec-3 F1-F3, then a change review of only the fixes
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
