---
name: flow-spec
description: Turns an idea into an approved spec: interview the human, write spec.md, run the spec check, get approval. Use when starting a new feature or slice, when a slice is in spec, or when an approved spec must change.
---

# flow-spec

You are the lead. Follow these steps in order. Each step ends when its
"Done when" is true. After every step that changes something, update the
slice's `status.md`, add a Log line, and commit
(`workflow/roles/lead.md` → Keeping the status).
Also follow the project rules for the lead and for Everyone
(`AGENTS.md` section 10 → Other rules).
To change a spec that is already approved, use `change.md` instead.

## 1. Is it a slice?

If the human has a problem or a vague idea, not yet a feature, explore it
first, before creating anything (case C in step 3). Brainstorming happens
before a slice exists. "Not worth building" and "several slices" are fine results:
then nothing is created, and ideas go to `backlog.md`.

Check the idea against `AGENTS.md` section 1. If it is a small change,
say so and stop. If it looks bigger than one slice (more than about
25-30 requirements), propose how to split it.

Before creating it, list the slices not merged yet:
`git branch --list "slice/*" --no-merged <base branch>`, and read each one's `status.md`.
If the new slice needs something from one of them (its code, a contract,
the data model, or a decision), propose stacking it on that slice: it can
be planned now, but built only after that slice is merged. Otherwise it
starts from the base branch. The human confirms.
(`workflow/formats/lifecycle.md` → Branches and worktrees.)

Create the slice:
1. **Number:** the highest slice number in `slices/` on the base branch and
   in the names of `slice/*` branches, plus one. 4 digits: `0007-<name>`.
2. **Branch and worktree**, as the git policy names them (`AGENTS.md` section 10):
   `git worktree add <repo root>/.worktrees/NNNN-<name> -b <slice branch> <start>`.
   `<repo root>` is the main checkout's folder: the output of
   `git rev-parse --path-format=absolute --git-common-dir`, without the
   last `/.git`. Always use full paths, so the folder never ends up inside
   another worktree.
   `<start>` is the base branch, or the earlier slice's branch when stacked.
   From now on, work in that worktree.
3. **Files:** copy `workflow/templates/spec.md` to `slices/NNNN-<name>/spec.md`,
   and fill `slice`, `title`, and `created` (today, UTC). Create `status.md`
   (`workflow/formats/status.md`) with phase `spec`, `After: NNNN` when
   stacked, and every other section "(none)".
4. **Commit:** `NNNN: new slice`.

Done when: the human agrees it is one slice, and the slice folder exists.

## 2. Learn the facts yourself

Before asking anything, find what you can look up:
- the code and current behavior this feature touches
- earlier slices' specs and their Decisions (SD…), and `backlog.md`
- the project's contracts and data model
- any document the request came with: a ticket, a requirements doc, a design, a bug report

Facts are your job. Decisions are the human's. Never ask the human
something the code or the docs can answer.

Done when: you can describe, in a few sentences, how things work today
in the area this feature touches.

## 3. Interview the human

### Where to start

Start from where the human is. Read their first message:

- **They describe a flow or steps (A):** play it back as a walkthrough and
  let them correct it. List any conflicts with the code. Then interview
  only the branches their flow does not cover yet.
- **They have a goal but no steps (B):** propose a main path from the code
  and the goal. At each real choice, show the options and your
  recommendation. Then interview.
- **They have a problem or a vague idea (C):** this happens in step 1,
  before the slice exists. Talk about the problem first: who has it, what
  they do today, what hurts. Offer 2-3 genuinely different approaches with
  their trade-offs. Help them pick, then continue as B.

If unclear, ask once: "Do you already have a flow in mind, or should I
propose one?" People may switch for one part of the feature.

Also check:
- **A document?** If the request comes with a ticket, doc, design, or bug
  report, interview only what it does not answer.
- **Who decides?** If the human is not the one who decides the product
  questions, mark those questions "needs <who>". Do not let them guess.
  The spec cannot be approved until they are answered.
- **A change to an existing feature?** Start from that slice's spec and the
  current code. Ask what changes and what must stay the same.
- **No end user?** For technical work, the "user" is a developer or the
  system. The walkthrough shows commands, logs, or numbers.

### Rounds

Build a **design tree**: each decision opens the decisions that depend on it.
Ask in **rounds** of at most about 5 questions. A round asks the open
questions whose parent is already decided. A question that depends on
another one in the same round waits for the next round.

Number each question and give your recommendation:

    Q3. What happens when the code expires during approval?
    Recommended: show "Code expired, start again" and cancel the pairing.

Word it so that "yes" accepts your recommendation.

During the interview:
- **Sharpen words.** When a word could mean two things, ask which one.
  Use one word per concept from then on.
- **Test with scenarios.** Invent concrete edge cases: two clicks at
  once, an empty list, a slow network, a user without permission.
- **Check against the code.** If the human describes behavior the code
  does not have, say so.
- **Keep the how for later.** If a technical idea comes up, write it in
  the slice's `notes.md` and move on. The spec stays about what.
- **Write as you go.** After each round, update the `spec.md` draft.
  If the session ends, the next one continues from the file.

### Branches to cover

The tree must cover every branch below. Each one ends answered, or
marked "not relevant" by the human:

- who it is for, and why now
- the main path, step by step, with real screens or output
- each failure: invalid input, not found, expired, conflict, no permission, network or service down
- limits: size, count, time, rate
- empty and first-time states
- what existing behavior changes, and who it affects
- security and privacy
- what is out of scope, and where each item goes

Done when: every branch is answered or marked not relevant, you have
played back a short summary, and the human says it matches what they mean.

## 4. Write the spec

Fill `spec.md` from the template, following the rules in its comments.
When you fill a section, delete its guidance comment: every reader pays for
it, and the template keeps the guidance. The ones most often missed:
- What, never how. No endpoints, tables, libraries, or file names.
- The walkthrough shows real output, not "an error is shown".
- Every requirement is testable, in one of the EARS forms, with a `Check:` line.
- Every out-of-scope item says where it goes.
- Every decision says why. One that replaces an earlier slice's decision names it.

Done when: every section has content or "(none)", every walkthrough step
is covered by at least one requirement, and Open questions is empty.

## 5. Self-check

Run the spec checks in `workflow/formats/document-checks.md`. Fix every
error. Then read the spec once as a worker would: only what is written.
Anything you would have to guess, fix now.

Done when: no errors and no warnings are left, and you found nothing to guess.

## 6. Spec check

Submit the spec: set its status to `in_review` in `status.md`, commit
`NNNN: spec submitted for review`, and start the spec check with
`flow-review` (kind `spec`, target `R<next>`, the next revision: R001, R002, …).
A fresh plan reviewer lists every real gap as a question, most important first.
When the review ends: questions (`needs_human`) → spec back to `draft`; none (`pass`) → `awaiting_approval`.

When the questions are back, do not show them yet. Ask the human to read
the spec first. Show the questions after they say they have read it.
A list of questions shown first makes people stop reading closely.

Sort the questions. A fact the code or docs can answer: answer it yourself
and tell the human in one line. A decision: ask the human, in rounds of at
most about 5, most important first, each with your recommendation.
Write the answers into the spec after each round, and run the spec checks again.
Then submit again: if the answers added no requirements, set
its status to `awaiting_approval` directly, with no new reviewer. If they added
requirements, the spec check runs again, on the new requirements only.

Done when: every question is answered and the spec is updated.

## 7. Approval

Show the human a short summary: what the slice does, the number of
requirements, what is out of scope, and the decisions.
Show the full spec when asked.

Only when the spec is `awaiting_approval` and the human has just said they approve:
1. In `status.md` → Documents: status `approved`, revision `R<next>`,
   approved at (now, UTC) and by (the human's name).
2. Copy the spec to `history/spec/R<next>.md`.
3. Set the phase to `plan`.
4. Commit: `NNNN: spec R<next> approved ("<the human's words>")`.

If the human rejects it: back to `draft`, with their reason in the Log.

Done when: the spec is approved.
