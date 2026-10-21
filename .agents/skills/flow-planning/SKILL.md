---
name: flow-planning
description: Runs a slice's planning phase with the human (the spec, the plan, tasks, and preview, their reviews, and the one approval). Use when a slice is in its planning phase, or when the human wants to write or change a slice's spec.
---

# flow-planning

The slice was created by `flow-slice`. Work in its worktree, on
`docs/slices/<name>/`. Read `docs/templates/spec.md` before writing the
spec: its comments hold the spec's rules.

## Keeping status.md

After every step that changes something, update the slice's `status.md`
as `docs/templates/status.md` says, then commit:

- the header: `Phase`, and `Updated` (the time now, UTC)
- the section the step changed
- Needs human: add what now waits on the human; remove what no longer does
- a Log line: time · who · what happened

## Part 1: the spec

### 1. Let the human propose

- The human talks first. Do not open with numbered questions.
- Talk naturally, like brainstorming. Talk about real cases: what the user
  wants to do, and how the system behaves when they do it.
- Ask one open question at a time while the idea is still forming.

### 2. Say it back before you write it

Before you write anything into `spec.md`, say it back in plain, natural
words, and wait for the human to agree. For example: "So when someone runs
`flow slice new` without a name, it stops and asks for one, right?"

Only after they agree, write it in the template's form (stories,
scenarios, requirements). Do not recite the formal form to the human.

### 3. Find what is missing

When the human has finished proposing, look for what the spec still does
not say. Sort each missing point:

- **You can decide it** when one option is clearly best, it does not
  change what the user sees, and it is easy to change later. Example: the
  order of output lines. Decide it, write it into the spec with the mark
  `(lead's call)`, and tell the human what you decided.
- **The human decides it** when it is about scope, behavior the user
  sees, data, security, or cost, or when you are not sure. Add it to Open
  questions.

### 4. Ask the open questions, 3 at a time

- Put every open question in the Open questions section first, numbered
  Q1, Q2, ... While a batch waits on the human, list it under Needs human
  in `status.md`.
- Ask them 3 at a time, the most important first: questions whose answer
  changes other questions, or the shape of the feature.
- The next 3 come only after the human has dealt with all 3 of the batch.
- For each question, say what is happening, why it matters, the options,
  and your recommendation.
- An answered question: say back what you will write, write it into the
  spec (as a requirement, a decision, or out of scope), and remove it from
  Open questions.
- A deferred question counts as dealt with. Do not ask it again. It stays
  in Open questions, marked `Deferred.`

### 5. The spec check

When every open question is answered or deferred:

1. Set `spec.md` to `in_review` in `status.md`.
2. Start a doc reviewer for a spec check (`AGENTS.md` → Agents).
   Its prompt: read `.agents/prompts/doc-reviewer.md` and follow it. Its
   review file: `reviews/spec-<round>.md`, from `docs/templates/review.md`
   with the front matter filled.
   It looks for conflicts and unresolved points in `spec.md`.
3. The reviewer gives each finding a level:
   - **high:** needs the human's decision. Add it to Open questions.
   - **medium:** you can decide it. Fix it with the mark `(lead's call)`,
     and tell the human.
   - **low:** a small finding. Fix it or leave it, as you see fit. It
     never goes to Open questions.
   Write each finding and its state in `status.md` → Findings → Doc
   reviews.
4. If there are high findings: set `spec.md` back to `draft`, ask the new
   questions as in step 4, then run the spec check again.
5. If there are none: set `spec.md` to `review_passed`.

The review round limit is in `AGENTS.md`. When it is reached and the spec
still has high findings, mark the slice `blocked` and ask the human what
to do.

### 6. When the spec changes later

If writing the plan shows a gap in the spec, set `spec.md` back to
`draft`, settle the gap as in steps 2-4, and run the spec check again. It
counts as a round.

When `spec.md` is `review_passed`, the plan can start.

## Part 2: the plan

Read `docs/templates/plan.md`, `tasks.md`, and `preview.md` before you
write each one: their comments hold their rules.

### 1. Start the plan

1. Copy `docs/templates/plan.md` and `docs/templates/preview.md` into the
   slice folder. Fill their front matter: `slice`, and for the plan
   `created` (today) and `touches`.
2. In `preview.md`, write the fixed checks the preview template lists for
   the spec.
3. In `status.md`, add `plan.md` and `preview.md` to the Documents table
   as `draft`, and add a Log line.

### 2. Draft the plan with the human

The plan says how the spec gets built, which you know best. So here you
propose first, and the human reacts.

- Propose each part in plain words: the approach, the flows, the
  interfaces, the data, the tests, the risks.
- Say it back before you write it, as in the spec, and wait for the
  human's yes.
- Sort missing points the same way: a `(lead's call)` you decide and
  report, or an open question for the human.
- Write `preview.md` along with the plan: its setup, steps, and Expect:
  lines.

### 3. Ask the open questions, 3 at a time

Exactly as in the spec: all questions listed in the plan's Open
questions, asked 3 at a time, the next 3 only after the human dealt with
all 3, and a deferred question stays listed, marked `Deferred.`

If a question shows a gap in the spec, change the spec as in Part 1,
step 6.

### 4. Write the tasks

When every open question is answered or deferred:

1. Copy `docs/templates/tasks.md` into the slice folder, fill its front
   matter (`slice`), and write every task.
2. In `status.md`, add `tasks.md` to the Documents table as `draft`, and
   fill the Tasks table: one row per task, status `todo`, Attempt and
   Round `0`.

### 5. The plan review

1. Set `plan.md`, `tasks.md`, and `preview.md` to `in_review`.
2. Start a doc reviewer for a plan review (`AGENTS.md` → Agents), told
   to read `.agents/prompts/doc-reviewer.md`. Its review file:
   `reviews/plan-<round>.md`. It checks the plan, tasks, and preview
   together, against the spec.
3. Sort its findings by level, as in the spec check: high to Open
   questions, medium as a `(lead's call)`, low fixed or left. Write them
   in `status.md` → Findings → Doc reviews.
4. If there are high findings: set the documents they touch back to
   `draft`, ask the questions 3 at a time, and run the plan review again.
5. If there are none: set the three documents to `review_passed`.

The review round limit is in `AGENTS.md`, as for the spec check.

## Part 3: approval

### 1. Close the deferred questions

Before you ask for approval, list every question still marked
`Deferred.` in the spec and the plan. Each must be answered, or moved to
Out of scope with `Goes to: the backlog`. Ask them 3 at a time. The human
cannot approve while one is still open.

### 2. Ask the human to approve

In `status.md` → Needs human: `The planning is waiting for approval`.
Tell the human the planning is ready: the four documents, what each
says in a few lines, and every `(lead's call)` you made. Ask them to read
and approve.

Only the human approves. Never approve for them, even if asked to "just
continue". Only a clear "approve" counts.

### 3. If the human asks for changes

Set the documents you change back to `draft`, make the change with the
human, and run the plan review again on the change. It counts as a
review round. Then ask for approval again.

### 4. When the human approves

1. In `status.md`: set the phase to `ready` and Updated, and set
   `spec.md`, `plan.md`, `tasks.md`, and `preview.md` to `approved`, with
   Revision `R001` and the date, the human, and the tag in Approved.
   Remove the approval from Needs human. Add a Log line with the human's
   words.
2. Commit: `<name>: planning approved`.
3. Make one annotated tag on that commit, `flow/<name>/planning/R001`.
   Its message names the four documents and quotes the human's approval.
   Write the message to a file and use `git tag -a -F <file>`.

A later approval after a change gets the next revision: `R002`, and so
on. Never move or delete an approval tag.
