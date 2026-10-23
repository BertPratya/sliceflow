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
   `flow-doc-reviewer`, model `opus`. Its prompt: read
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

The backlog is `docs/slices/backlog.md`, on `main`. Add an item at once,
so every slice can see it:

1. In the main checkout, check it is on `main`. If not, stop and ask the
   human: it is their checkout.
2. Add one row at the end of the table, as `docs/templates/backlog.md`
   says: status `open`, From the spec and question, e.g.
   `docs/slices/<name>/spec.md Q4`.
3. Commit only that file: `backlog: <the item, short>`.

## Part 2: the plan

To be written.
