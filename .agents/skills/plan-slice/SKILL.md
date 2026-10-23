---
name: plan-slice
description: Runs a slice's planning phase with the human, from the spec to the one approval. Use when a slice is in its planning phase, or when the human wants to write or change a slice's spec.
---

# plan-slice

The slice was created by `start-slice`. Work in its worktree, on
`docs/slices/<name>/`. Before you write the spec, read
`docs/templates/spec.md`: its comments hold the spec's rules. If the
spec's front matter names ideas, read them first.

## Keeping status.md

After every step that changes something, update the slice's `status.md`
as `docs/templates/status.md` says, then commit in the worktree:

- the header: `Phase`, and `Updated` (the time now, UTC)
- the section the step changed
- Needs human: add what now waits on the human; remove what no longer does
- a Log line: time · who · what happened

## Asking the human a question

Every question to the human, in any part of this skill, has this form:

- **What we have now:** what the spec already says, in plain words.
- **What is missing:** the gap, with a small example of the case.
- **Options:** 2-3, each with its trade-offs, and your recommendation.

The human may choose an option, or propose their own. When they propose
their own, analyze it honestly: say whether it works, where it is weaker
or stronger, and compare it with the other options. Then let them decide.

## Part 1: the spec

### 1. Brainstorm with the human

- Invite the human to describe what the system should look like, in
  their own words. Let them talk first.
- Talk naturally, like brainstorming. Point out gaps and trade-offs as
  they come up, with your recommendation.
- Find the cases nobody covered yet, and ask about them one at a time as
  "what if" questions, e.g. "What if the name is already taken?"

The step is done when the direction is clear: the human has no open
point they want to settle now.

### 2. Play it back, then write it

Summarize the workflow in the chat, in plain words: what the user does,
and what the system does in return. Then ask both: "Does this match what
you're thinking? Shall I write it into `spec.md`?"

If they correct it, fix it and play it back again. Write `spec.md` only
after their yes, in the template's form (stories, scenarios,
requirements). Recite the plain version to the human, never the formal
form.

### 3. Check the spec yourself

Reread the whole `spec.md` as a worker would: only what is written. Flag
every point that is unresolved, or that conflicts with another part.
Write each one into Open questions, numbered Q1, Q2, ... While questions
wait on the human, list them under Needs human in `status.md`.

### 4. Ask the open questions, 3 at a time

- Ask 3 at a time, the most important first: questions whose answer
  changes other questions, or the shape of the feature. Use the form in
  "Asking the human a question".
- The next 3 come only after the human has dealt with all 3 of this
  batch.
- An answered question: write the answer into the spec (as a
  requirement, a decision, or out of scope), and remove it from Open
  questions.
- A question the human wants to answer later stays in Open questions,
  and the batch waits for it.
- A question the human wants to defer: first check whether it blocks the
  build. It blocks when a worker would have to guess what the user sees,
  the data, or security without its answer.
  - **It blocks:** tell the human why, with an example. It stays an open
    question.
  - **It does not block:** move it out of the slice. Add it to Out of
    scope with `Goes to: the backlog`, remove it from Open questions, and
    add it to the backlog (below).

The step is done when Open questions says `None.`

### 5. The spec review

1. In `status.md`, set `spec.md` to `in_review`, and add 1 to its Round.
2. Copy `docs/templates/review.md` to
   `docs/slices/<name>/reviews/spec-<round>.md`, and fill its front
   matter.
3. Start a doc reviewer in a fresh context: Agent tool, type
   `sliceflow-doc-reviewer`, model `opus`. Its prompt: read
   `.agents/prompts/doc-reviewer.md` and follow it, for a spec check of
   `docs/slices/<name>/spec.md`, with the worktree's full path and the
   review file. It looks for conflicts and unresolved points.
4. Set `spec.md` back to `draft` while you deal with the findings, and
   write each finding and its state in `status.md` → Findings → Doc
   reviews:
   - **A finding you can decide** (medium): show the human, in one list,
     what you would decide and what you would change for each. One yes
     covers the list; the human may still say "not number 2". Change the
     spec only after their yes.
   - **A low finding** (a typo, a missing example): fix it or leave it,
     and list what you did in one line.
   - **A finding that needs the human** (high, under Needs human
     decision): write it into Open questions, and ask it as in step 4.
5. If the review found nothing to change: set `spec.md` to
   `review_passed`. Otherwise, run the review again from 1.

The review round limit is in `AGENTS.md`. When the spec still does not
pass at the limit, mark the slice `planning (blocked)` and ask the human
what to do.

The spec part is done when `status.md` shows `spec.md` as
`review_passed`. Part 2 starts only then.

### The backlog

The backlog is `docs/backlog.md`, on `main`. Add an item at once,
so every slice can see it:

1. In the main checkout, check it is on `main`. If not, stop and ask the
   human: it is their checkout.
2. Add one row at the end of the table, as `docs/templates/backlog.md`
   says: status `open`, From the spec and question, e.g.
   `docs/slices/<name>/spec.md Q4`.
3. Commit only that file: `backlog: <the item, short>`.

## Part 2: the plan

Before you write each document, read its template in `docs/templates/`
(`plan.md`, `tasks.md`, `preview.md`): its comments hold its rules.

### 1. Start the plan

1. Copy `docs/templates/plan.md` and `docs/templates/preview.md` into the
   slice folder. Fill their front matter: `slice`, and for the plan
   `created` (today). `touches` comes with the draft.
2. In `preview.md`, write the fixed checks its template lists for the
   spec.
3. In `status.md` → Documents, add `plan.md` (Round `0`) and `preview.md`
   (Round `with plan.md`), both `draft`.

### 2. Say how deep to go

Tell the human whether the slice looks simple or complex, and why, in one
or two lines. The human may change it with one word.

- **Complex** when it has any of: security, stored data or a migration,
  several programs or parts that must agree, concurrency, recovery after
  a crash, a new contract (an API, a file format), or more than about 6
  tasks.
- **Simple** otherwise, e.g. a CRUD feature on existing parts.

### 3. Propose the plan, and settle the decisions

You propose first; the human reacts.

- **Simple:** propose the whole draft at once, in plain words.
- **Complex:** go part by part, each settled before the next, because
  each one shapes the next:
  1. the approach and the package layout
  2. the data
  3. the interfaces and contracts
  4. the risky mechanisms: security, concurrency, recovery
  5. the flows and the testing

Some choices are always the human's, whatever the size: a new dependency,
the package layout, stored data and migrations, a contract, security, and
anything costly to change after it is built. Ask them in the form in
"Asking the human a question".

Decide everything else yourself, mark it `(lead's call)` in the plan, and
list it for the human.

Say each part back before you write it, and wait for the human's yes.
Write `preview.md` along with the plan: its setup, steps, and Expect:
lines.

### 4. Flag what is still open

When the draft is written, reread `plan.md` and `preview.md` against the
spec. Write every unresolved or unclear point into the plan's Open
questions, and ask them as in Part 1, step 4: 3 at a time, the same rules
for later answers and for deferring. A deferred plan question that does
not block goes to Parked and to the backlog.

If a point shows a problem in the spec itself, set `spec.md` back to
`draft` and settle it as in Part 1, steps 2-4. The plan review then
checks the spec change too.

The step is done when the plan's Open questions says `None.`

### 5. Write the tasks

1. Copy `docs/templates/tasks.md` into the slice folder, fill its front
   matter (`slice`), and write every task.
2. In `status.md`: add `tasks.md` to Documents (`draft`, Round `with
   plan.md`), and fill the Tasks table: one row per task, `todo`,
   Attempt and Round `0`.

### 6. The plan review

1. In `status.md`, set `plan.md`, `tasks.md`, and `preview.md` (and
   `spec.md`, if it changed in this part) to `in_review`, and add 1 to
   the plan's Round.
2. Copy `docs/templates/review.md` to
   `docs/slices/<name>/reviews/plan-<round>.md`, and fill its front
   matter.
3. Start a doc reviewer in a fresh context, as in Part 1, step 5, for a
   plan review: it checks the spec, plan, tasks, and preview together.
4. Deal with the findings exactly as in Part 1, step 5: the documents
   they touch go back to `draft`; findings you can decide go to the human
   in one list; lows fixed and listed; findings that need the human go
   to Open questions.
5. If the review found nothing to change: set the documents to
   `review_passed`. Otherwise, run the review again from 1.

The review round limit is in `AGENTS.md`, as for the spec.

## Part 3: approval

### 1. Ask the human to approve

Only when the review passed, and Open questions in both the spec and the
plan say `None.`

In `status.md` → Needs human: `The planning is waiting for approval`.
Tell the human the planning is ready: the four documents and what each
says in a few lines, every `(lead's call)`, and what went to the backlog.
Ask them to read and approve.

Only the human approves. Never approve for them, even if asked to "just
continue". Only a clear "approve" counts.

### 2. If the human asks for changes

Set the documents you change back to `draft`, make the change with the
human, and run the plan review again (Part 2, step 6). Then ask for
approval again.

### 3. When the human approves

1. In `status.md`: set the phase to `ready`, and set `spec.md`,
   `plan.md`, `tasks.md`, and `preview.md` to `approved`, with Revision
   `R001`, and in Approved the date, the human, and the tag. Remove the
   approval from Needs human. Add a Log line with the human's words.
2. Commit: `<name>: planning approved`.
3. Make one annotated tag on that commit, `sliceflow/<name>/planning/R001`.
   Its message names the four documents and quotes the human's
   approval. Write the message to a file and use `git tag -a -F <file>`.
4. Ask the human whether to start the implementation phase now.

A later approval after a change gets the next revision: `R002`, and so
on. Never move or delete an approval tag.
