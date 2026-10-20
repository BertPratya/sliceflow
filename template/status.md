<!-- HOW TO FILL THIS FILE

     status.md is the overview and the only state of one slice. Every
     status lives here, and nowhere else: not in document headers. The
     lead keeps it up to date after every step, with one Log line for each
     change, and commits it. Everyone else only reads it.

     (In manual flow, before flow's first version, the lead writes this
     file by hand. Later, flow writes it from the event log.)

     The sections always appear in this order. A section with nothing in
     it gets one line: None. Replace every {{...}}, and delete every
     comment when the file is first filled. To update it later, read this
     template again.

     Status words, and what they mean:

     Phase (the slice):
     - spec: the spec is being written or reviewed.
     - plan: the plan, tasks, and preview are being written or reviewed.
     - ready: the plan is approved; building has not started.
     - implementing: workers are building tasks.
     - integration: all tasks are done; the whole slice is being reviewed.
     - preview: the human is trying the slice.
     - done: merged into the base branch.
     - blocked: stopped until a person decides something.
     - cancelled: stopped for good.

     Documents:
     - draft: being written.
     - in review: a reviewer is checking it.
     - waiting for approval: the review passed; the human has not
       approved yet.
     - approved: the human approved it; the Approved column names the tag.
     - changed: it was edited after approval and must be approved again.

     Tasks:
     - todo: not started. In brackets, why it cannot start yet, e.g.
       todo (waiting: 04, 06).
     - doing: a worker is building it.
     - in review: a reviewer is checking it.
     - done: reviewed (or not needed, for low risk) and merged into the
       slice branch.
     - blocked: stopped; the reason is under Needs human.

     Findings:
     - open: not fixed yet.
     - fixed (waiting for verify): the worker says it is fixed; the next
       review checks it.
     - verified: the reviewer confirmed the fix, or accepted the reason.
     - reopened: the reviewer says it is not fixed.
     - wontfix: why: not fixed, with the reason.
     - deferred: backlog: moved out of the slice; copied to backlog.md
       when the slice is done.

     Preview checks:
     - not tried: the human has not done it yet.
     - pass: it worked as expected.
     - fail: why: it did not; a fix task follows. -->

# Status: {{slice}}

- Phase: {{phase}}
- Branch: slice/{{slice}}
- After: {{the slice this one waits for, or delete this line}}
- Updated: {{YYYY-MM-DD HH:MM}} UTC

## Needs human

<!-- Everything waiting on a person, one line each: documents waiting for
     approval, review questions, blocked tasks with the reason, high
     findings a worker is blocked on, failed preview checks.

     Example:
     - plan.md is waiting for approval (plan review passed).
     - Task 04 is blocked: checks failed 3 times. Last error: the date
       parser rejects 2026-02-28. -->

- {{What is waiting, and on whom.}}

## Documents

<!-- spec.md, plan.md, tasks.md, preview.md, and every project file in the
     plan's touches. Revision is the approved revision (R001, ...), or -
     before the first approval. Approved names the date, who, and the git
     tag.

     Example:
     | Doc | Status | Revision | Approved |
     |---|---|---|---|
     | spec.md | approved | R001 | 2026-10-20 by the human, tag flow/due-dates/spec/R001 |
     | plan.md | draft | - | - | -->

| Doc | Status | Revision | Approved |
|---|---|---|---|
| spec.md | {{status}} | {{R001 or -}} | {{date, who, tag, or -}} |

## Tasks

<!-- A progress line, then one row per task from tasks.md, in id order.
     Title is the task's name. Risk, Depends on, and Covers come from the
     task's YAML block. Attempt counts check runs since the last review;
     Round counts review rounds.

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

<!-- A count line, then one table per review kind: Plan review, Task
     reviews, Integration review. A kind with no reviews yet: None.
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

### Plan review

None.

### Task reviews

None.

### Integration review

None.

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
     - 2026-10-20 14:02 · the human · spec R001 approved: "approved" -->

Newest first.

- {{YYYY-MM-DD HH:MM}} · {{who}} · {{what happened}}
