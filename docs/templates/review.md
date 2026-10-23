---
review: {{review}}
slice: {{slice}}
kind: {{kind}}
target: {{target}}
round: {{round}}
reviewer: {{reviewer}}
---

<!-- HOW TO FILL THIS TEMPLATE

     A review checks one thing. A doc review (kind: spec, plan, or
     change) checks a spec, a plan with its tasks.md and preview.md, or
     the documents changed during building. A code review (kind: task or
     integration) checks one task, or the whole slice. sliceflow
     creates this file and fills
     the front matter. You write the sections below, then hand the review
     in through sliceflow. After that the file is locked: never edit it again.

     sliceflow reads this file and sets the verdict from it.
     Code review:
     - Needs human decision is not None.            → needs_human
     - else any high or medium finding, new or
       reopened                                     → changes
     - else                                         → pass
     (low findings alone pass)
     Doc review:
     - Needs human decision is not None.            → needs_human
     - else                                         → pass
     (the lead fixes medium findings, and fixes or leaves low ones)

     sliceflow reads this file. Keep every marker exactly as shown: the
     headings, the finding headings (### F1.), the labeled lines (Level:,
     Where:, Breaks:, Why:, Fix:, Guess:), the numbers (**Q1.**), and the
     line forms in Tests, Previous findings, and Requirements check.
     Don't rename or reorder them. A
     labeled line may go on over more lines: indent the extra lines 2
     spaces more than the label.

     Report real problems only: things that break a requirement, a
     decision, an interface, or a project rule, or that are clearly wrong.
     Don't invent issues to look thorough. If everything is fine, say so,
     and write None. under Findings.

     A spec review is short: fill Summary, Checked, Findings (medium and
     low), and Needs human decision (every high gap, most important
     first). Write None. in the other sections.

     - Replace every {{...}} with real text.
     - A section that doesn't apply gets one line: None.
     - When you are done, delete every comment and check that no {{ is
       left. Every placeholder of this template counts, even inside
       code. A {{ or a comment marker you write yourself inside a code
       block or `inline code` is real content. To name them in plain
       text, use `inline code`. -->

# Review

## Summary

<!-- 1-3 sentences. The overall result, and the main problem if there is
     one.

     Example:
     The task covers R10 and the tests are real. One high problem: a done
     task that is overdue still shows OVERDUE (F1). -->

{{1-3 sentences: the overall result, and the main problem if there is one.}}

## Checked

<!-- What this review covered: what you read, and what you ran. This
     shows the review was real, so a "None." under Findings can be
     trusted. List only what you actually did.

     - Commits: the commits reviewed (task and integration reviews).
     - Files read: the files you read, including ones the change uses.
     - Ran: each command you ran, and its result. More than one: one
       indented line per command under Ran:.
     - Requirements: the requirements you checked.
     - Interfaces: the names you checked against the task, if any.
     Leave out a line that doesn't apply.

     Example:
     - Commits: a1b2c3d..e4f5a6b
     - Files read: internal/list/format.go, internal/list/format_test.go,
       internal/task/overdue.go
     - Ran: go test ./internal/list/... (pass)
     - Requirements: R10
     - Interfaces: Format(tasks, today) matches the task -->

- Files read: {{files}}
- Ran:
  {{command (result), one per line}}
- Requirements: {{R1, R2}}

## Tests

<!-- Task reviews only. Other kinds: write None.
     One line per test named in the task's Done when. Read each test,
     and say whether it really proves what it claims:
     - ok: it exists, checks the stated input and expected result, and
       would fail if the code were broken.
     - weak, see F2: it exists but would pass with broken code, or
       checks less than Done when says. Add a finding.
     - missing, see F3: it doesn't exist. Add a finding.

     Example:
     - `TestParseDueRejectsBadDate`: ok
     - `TestDueUnknownTask`: weak, see F2
     - `TestListSortsByDue`: missing, see F3 -->

- `{{TestName}}`: {{ok, or: weak, see F2, or: missing, see F3}}

## Previous findings

<!-- Round 2 and later. Every finding from the last round that was marked
     fixed, wontfix, or deferred. One line each:
     - <review>/F1: verified (fixed)
     - <review>/F2: verified (wontfix)
     - <review>/F3: verified (deferred)
     - <review>/F4: reopened. why

     verified means: a fixed finding is really fixed, or the reason given
     for wontfix or deferred is acceptable. In brackets, say which one.
     Otherwise: reopened.

     Example:
     - task-04-1/F1: verified (fixed)
     - task-04-1/F2: reopened. A done task due today still shows OVERDUE.

     Round 1: write None. -->

- {{task-02-1/F1}}: {{verified (fixed, wontfix, or deferred), or: reopened. why}}

## Findings

<!-- New findings only. Number them F1, F2, ... in this file. Outside
     this file they are called <review>/F1, e.g. task-04-1/F1. List them
     high first, then medium, then low.

     Each finding has these lines:
     - Level: high, medium, or low. In a doc review the level says who
       deals with it:
       - high: needs the human's decision. Never a finding: write it as
         a question under Needs human decision, with a Guess: line.
       - medium: the lead can decide and fix it alone: one fix is
         clearly right and changes nothing the human chose.
       - low: a small thing, e.g. a typo or a missing example. The lead
         fixes it or leaves it.
       In a code review the level says how bad it is:
       - high: wrong behavior, breaks the spec, plan, or contract,
         security, or data loss. Blocks. Only a human can mark it
         wontfix, or defer it out of the slice.
       - medium: a real problem that doesn't break anything yet. Must be
         fixed, or a reason given.
       - low: nice to have. Optional.
     - Where: file:line, or the document and section. A range:
       file:12-18. More than one place: separate them with commas.
     - Breaks: what it breaks: a requirement (R5), a decision (SD1,
       TD2), a task's Interfaces (task 02), or a project rule (name it).
       A decision without a number: name the document and section,
       e.g. docs/sliceflow-design.md 2.3.
       The work adds behavior nothing asks for? Write: scope, and what
       it adds, e.g. Breaks: scope: adds --color, which no requirement
       asks for. A scope finding is at least medium: the plan never adds
       scope.
       Breaks nothing approved, e.g. a missing rule or a gap? Write:
       nothing. Choose the level by its effect, as above. With Breaks:
       nothing, a finding is medium only if Why: names an input and the
       wrong result it gives. Otherwise it is low.
     - Why: what is wrong, and how you found it: the test you ran, or
       the line you read.
     - Fix: optional, one or two lines. No fix to suggest? Remove the
       whole Fix: line.

     Example:
     ### F1. A done task that is overdue still shows OVERDUE

     Level: high
     Where: internal/list/format.go:31
     Breaks: R12
     Why: format.go checks only the due date, not whether the task is
       done. I added a test with a done task due yesterday; it printed
       OVERDUE.
     Fix: call task.IsOverdue() instead of comparing the dates here. -->

### F1. {{title}}

Level: {{high, medium, or low}}
Where: {{file:line, or the document and section}}
Breaks: {{R5, TD2, task 02, a project rule, scope: what it adds, or nothing}}
Why: {{what is wrong, and how you found it}}
Fix: {{optional: one or two lines}}

## Requirements check

<!-- Integration reviews only. Other kinds: write None.
     Every requirement in the spec, checked against the finished slice.
     (A review whose requirements were given with the request, not by a
     spec, checks those instead.)
     One line each: ok, fails (with the finding), or manual (goes to the
     preview). A fails line names the high or medium findings that make
     it fail; lows may follow, e.g. fails, see F3; also F7. Only low
     findings against it? It is ok: write ok, see F7.

     Example:
     - R1: ok
     - R3: fails, see F2
     - R14: manual -->

- {{R1}}: {{ok, or: fails, see F2, or: manual}}

## Needs human decision

<!-- Questions only the human can answer, numbered Q1, Q2, ... Each one
     says what is unclear, and has a Guess: line: what a worker would
     have to guess without the answer. In a doc review, every high gap
     goes here. Anything here makes the verdict needs_human.

     Example:
     - **Q1.** What happens when a task's due date is changed after it
       was marked done? R2 and R12 don't say.
       Guess: keep it done and change the date, or reopen it. -->

- **Q1.** {{What is unclear.}}
  Guess: {{what a worker would have to guess without the answer}}
