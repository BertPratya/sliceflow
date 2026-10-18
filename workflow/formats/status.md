# Format: slice status.md

`slices/NNNN-<name>/status.md` is the overview and the only state of one slice: every status lives here, and nowhere else. The lead keeps it up to date after every step, with a dated Log line for each change. Everyone else only reads it.

## Where the data comes from

| Data | Source |
|---|---|
| Task title, layer, risk, depends on, covers | task file headers |
| Finding titles and levels | submitted review files |
| Phase, status, attempt, round, approvals, finding status, log | recorded by the lead at each step |
| Waiting, stale | worked out from the task headers when the lead updates the file |

## Sections

The sections always appear in this order. A section with nothing in it says "(none)".

1. **Header.** Slice ID and title, then phase, track, branch, the slice it is stacked on (if any, e.g. `After: 0001 (waiting until merged)`), and last update time (UTC).
2. **Needs human.** Everything waiting on a person: blocked tasks with the reason, reviews with verdict `needs_human`, documents waiting for approval, and high findings a worker blocked on, asking for wontfix or to defer them out of the slice. One line each.
3. **Documents.** `spec.md`, `plan.md`, and every project file in the plan's `touches`. Status, revision, and who approved it and when.
4. **Tasks.** A progress line, then one row per task, in task order.
   - Status shows calculated states in brackets: `todo (waiting: 04, 06)`, `todo (stale: plan R001 → R002)`.
5. **Findings.** A count line, then one table per review kind: Plan review, Task reviews, Integration review. A kind with no reviews yet says "(none)".
   - Open findings first, then closed ones. Closed findings are kept.
   - Plan findings show the document they are about.
   - Task findings show the task.
   - Integration findings show the fix task created for them. They close when that fix task is done and the finding is verified. Columns: `ID | Level | Fix task | Title | Status`, e.g. `integration-1/F1 | high | 13 | Worker and server disagree on expiry | open (task 13 todo)`.
6. **Parked.** Ideas for later: low findings with "later" in the title, and items the lead deferred. One line each, with where it came from. Moved to the backlog when the slice is done.
7. **Log.** Every event as one readable line, newest first: `time · actor · what happened`.

## Example

```markdown
<!-- Kept by the lead. Updated after every step. -->

# 0001-pairing: Device pairing

- Phase: implementing
- Track: complex
- Branch: slice/0001-pairing
- Updated: 2026-10-17 14:32 UTC

## Needs human
- Task 04 is blocked: checks failed 3 times. Last error: migration 0002 fails on empty database.
- Review task-06-3 needs a decision: should an expired code be retried or rejected?

## Documents
| Doc | Status | Revision | Approved |
|---|---|---|---|
| spec.md | approved | R002 | 2026-10-12 by alex |
| plan.md | approved | R001 | 2026-10-13 by alex |
| api/openapi.yaml | approved | R001 | 2026-10-13 by alex |

## Tasks
3 of 12 done · 1 in progress · 1 in review · 1 blocked · 6 todo

| # | Title | Layer | Risk | Status | Attempt | Round | Depends on | Covers |
|---|---|---|---|---|---|---|---|---|
| 01 | Server foundation | backend | core | done | 1 | 1 | — | R1, R2 |
| 02 | Start, poll, approval | backend | core | done | 2 | 2 | 01 | R3–R9 |
| 03 | Approval page | frontend | normal | done | 1 | 1 | 02 | R10, R11 |
| 04 | Code and credential | backend | core | blocked | 3 | 0 | 02 | R12–R15 |
| 05 | Worker local storage | backend | normal | in_progress | 1 | 0 | 01 | R16 |
| 06 | Worker start and poll | backend | normal | in_review | 1 | 3 | 05 | R17, R18 |
| 07 | Worker code and confirm | backend | normal | todo (waiting: 04, 06) | 0 | 0 | 04, 06 | R19 |

## Findings
Open: 2 · Waiting for verify: 1 · Closed: 3

### Plan review
| ID | Level | About | Title | Status |
|---|---|---|---|---|
| plan-R001-1/F3 | medium | plan.md | Retry rule unclear | wontfix: covered by TD4 |

### Task reviews
| ID | Level | Task | Title | Status |
|---|---|---|---|---|
| task-06-2/F1 | high | 06 | Poll does not stop after expiry | reopened |
| task-06-3/F1 | medium | 06 | No backoff on 429 | open |
| task-06-2/F2 | low | 06 | Log message too long | fixed (waiting for verify) |
| task-02-1/F1 | high | 02 | Missing 404 for unknown key | verified |
| task-02-1/F2 | low | 02 | Unused helper | deferred: backlog |

### Integration review
(none)

## Parked
- Show remaining pairing time in the worker's output (from task-06-2/F3)

## Log
Newest first.

- 2026-10-17 14:32 · code-reviewer (codex) · review task-06-3 submitted: needs_human
- 2026-10-17 14:05 · worker (claude) · task 06 sent to review (round 3)
- 2026-10-17 13:40 · lead (claude) · task 04 blocked: checks failed 3 times
- 2026-10-17 13:12 · worker (claude) · task 04 checks failed (attempt 3)
- 2026-10-12 10:03 · alex · spec R002 approved: "added R19 for confirm step"
```
