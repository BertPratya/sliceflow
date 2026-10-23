<!-- HOW TO FILL THIS FILE

     status.md is the overview and the only state of one slice. Every
     status lives here, and nowhere else: not in document headers. The
     lead keeps it up to date after every step, with one Log line for each
     change, and commits it. Everyone else only reads it.

     The sections always appear in this order. A section with nothing in
     it gets one line: None. Replace every {{...}}, and delete every
     comment when the file is first filled. To update it later, read this
     template again.

     Status words, and what they mean:

     Phase (the slice):
     - planning: the spec, plan, tasks, and preview are being written or
       reviewed.
     - ready: the human approved the planning; building has not started.
     - implementing: workers are building tasks.
     - integration: all tasks are done or cancelled; the whole slice is
       being reviewed.
     - preview: the human is trying the slice.
     - done: the human approved the result; set in the last commit
       before the merge into the base branch.
     Marks, in brackets after the phase, which stays as it was:
     - (blocked): stopped until a person decides something, e.g.
       implementing (blocked).
     - (cancelled): the human cancelled the slice; it can be reopened
       from the same phase.

     Documents:
     - draft: being written.
     - in_review: a reviewer is checking it.
     - review_passed: the review passed; the human has not approved yet.
     - approved: the human approved it; the Approved column names the tag.
     - changed: it was edited after approval and must be approved again.
     Round (spec.md and plan.md): the review rounds so far, 0 before the
     first review. It never resets, so review file names stay unique.
     tasks.md and preview.md are reviewed with the plan: their Round is
     with plan.md.

     Tasks:
     - todo: not started. In brackets, why it cannot start yet, e.g.
       todo (waiting: 04, 06).
     - doing: a worker is building it.
     - in_review: a reviewer is checking it.
     - done: reviewed (or not needed, for low risk) and merged into the
       slice branch.
     - blocked: stopped; the reason is under Needs human.
     - cancelled: no longer needed after a plan change; its branch is
       kept.

     Findings:
     - open: not fixed yet.
     - fixed (waiting for verify): the worker says it is fixed; the next
       review checks it.
     - verified: the reviewer confirmed the fix, or accepted the reason.
     - reopened: the reviewer says it is not fixed.
     - wontfix: why: not fixed, with the reason.
     - deferred: task NN: a later task in this slice fixes it; that
       task's review checks it.
     - deferred: backlog: moved out of the slice; copied to backlog.md
       when the slice is done.

     Preview checks:
     - not tried: the human has not done it yet.
     - pass: it worked as expected.
     - fail: why: it did not; a fix task follows. -->

# Status: {{slice}}

- Phase: {{phase, with a mark in brackets if any}}
- Branch: slice/{{slice}}
- After: {{the slice this one waits for, or delete this line}}
- Updated: {{YYYY-MM-DD HH:MM}} UTC

## Needs human

<!-- Everything waiting on a person, one line each: documents waiting for
     approval, review questions, blocked tasks with the reason, high
     findings a worker is blocked on, failed preview checks.

     Example:
     - The planning is waiting for approval (plan review passed).
     - Task 04 is blocked: checks failed 3 times. Last error: the date
       parser rejects 2026-02-28. -->

- {{What is waiting, and on whom.}}

## Documents

<!-- spec.md, plan.md, tasks.md, preview.md, and every project file in the
     plan's touches. Round counts review rounds (see the status words
     above). Revision is the approved revision (R001, ...), or - before
     the first approval. Approved names the date, who, and the git tag.

     Example:
     | Doc | Status | Round | Revision | Approved |
     |---|---|---|---|---|
     | spec.md | review_passed | 2 | - | - |
     | plan.md | in_review | 1 | - | - |
     | tasks.md | in_review | with plan.md | - | - |
     | preview.md | in_review | with plan.md | - | - | -->

| Doc | Status | Round | Revision | Approved |
|---|---|---|---|---|
| spec.md | {{status}} | {{0}} | {{R001 or -}} | {{date, who, tag, or -}} |

## Tasks

<!-- A progress line, then one row per task from tasks.md, in id order.
     Title is the task's name. Risk, Depends on, and Covers come from the
     task's YAML block. Attempt counts failed check runs since the last
     review or unblock; Round counts review rounds.

     Example:
     2 of 5 done · 1 doing · 2 todo

     | # | Name | Risk | Status | Attempt | Round | Depends on | Covers |
     |---|---|---|---|---|---|---|---|
     | 01 | due-date-model | core | done | 1 | 1 | - | R1, R13 |
     | 03 | sort-by-due | normal | todo (waiting: 01) | 0 | 0 | 01 | R6, R7 | -->

{{N}} of {{N}} done · {{counts by status}}

| # | Name | Risk | Status | Attempt | Round | Depends on | Covers |
|---|---|---|---|---|---|---|---|
| {{01}} | {{name}} | {{risk}} | {{status}} | {{0}} | {{0}} | {{-}} | {{R1}} |

## Findings

<!-- A count line, then one table per review kind: Doc reviews (spec
     check, plan review, change reviews), Task reviews, Integration review. A kind with no reviews yet: None.
     Open findings first, then closed ones; closed findings are kept.
     The finding's text stays in its review file; only its state is here.

     Example:
     Open: 1 · Waiting for verify: 0 · Closed: 2

     ### Task reviews
     | ID | Level | Task | Title | Status |
     |---|---|---|---|---|
     | task-03-1/F1 | medium | 03 | Tasks without a date sort first | open |
     | task-01-1/F2 | low | 01 | Unused helper | deferred: backlog | -->

Open: {{N}} · Waiting for verify: {{N}} · Closed: {{N}}

### Doc reviews

None.

### Task reviews

None.

### Integration review

None.

## Lead's decisions

<!-- Every decision the lead made alone while building: it changes no
     approved document. One line each: the task, what, and why. Shown to
     the human at the preview.

     Example:
     - 03: the sort keeps tasks with the same date in input order. Why:
       the spec gives no order, and input order is what users see today. -->

- {{task}}: {{what}}. Why: {{why}}

## Preview

<!-- One line per check in preview.md, with its result. The human reports
     results; the lead writes them here.

     Example:
     - P1 (S1): pass
     - P2 (SC1): fail: the lead was called after task 03. Fix task 06.
     - P3 (R14): not tried -->

- {{P1}} ({{what it covers}}): {{not tried, pass, or fail: why}}

## Log

<!-- Every change as one line, newest first: time · who · what happened.
     Who is a role and its tool, e.g. lead (claude), or the human.

     Example:
     - 2026-10-20 15:04 · lead (claude) · plan submitted for review
     - 2026-10-20 14:02 · reviewer (claude) · spec check passed -->

Newest first.

- {{YYYY-MM-DD HH:MM}} · {{who}} · {{what happened}}
