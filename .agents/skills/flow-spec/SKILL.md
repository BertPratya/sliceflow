---
name: flow-spec
description: Turns an idea into an approved spec: interview the human, write spec.md, run the spec check, get approval. Use when starting a new feature or slice, when a slice is in brainstorm or spec, or when an approved spec must change.
---

# flow-spec

You are the lead. Follow these steps in order. Each step ends when its
"Done when" is true. If `flow` does not exist yet, use `bootstrap.md` in
this folder wherever a step names a `flow` command.
To change a spec that is already approved, use `change.md` instead.

## 1. Is it a slice?

Check the idea against `AGENTS.md` section 1. If it is a small change,
say so and stop. If it looks bigger than one slice (more than about
25-30 requirements), propose how to split it.

Before creating it, run `flow status` to list the slices not merged yet.
If the new slice needs something from one of them (its code, a contract,
the data model, or a decision), propose stacking it on that slice: it can
be planned now, but built only after that slice is merged. Otherwise it
starts from the base branch. The human confirms.
(`workflow/formats/lifecycle.md` → Branches and worktrees.)

Run `flow slice new <name>`, with `--after <NNNN>` when stacked. It creates
the slice branch, its worktree, the folder, and `spec.md` from the template.
From now on, work in that worktree.

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
- **They have a problem or a vague idea (C):** talk about the problem first:
  who has it, what they do today, what hurts. Offer 2-3 genuinely different
  approaches with their trade-offs. Help them pick, then continue as B.
  "Not worth building" and "split into several slices" are fine results.

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
The ones most often missed:
- What, never how. No endpoints, tables, libraries, or file names.
- The walkthrough shows real output, not "an error is shown".
- Every requirement is testable, in one of the EARS forms, with a `Check:` line.
- Every out-of-scope item says where it goes.
- Every decision says why. One that replaces an earlier slice's decision names it.

Done when: every section has content or "(none)", every walkthrough step
is covered by at least one requirement, and Open questions is empty.

## 5. Self-check

Run `flow validate`. Fix every error it reports.
Then read the spec once as a worker would: only what is written.
Anything you would have to guess, fix now.

Done when: `flow validate` passes and you found nothing to guess.

## 6. Spec check

Run `flow doc submit spec`. This starts the spec check: a fresh
plan reviewer writes up to 5 questions.

When the questions are back, do not show them yet. Ask the human to read
the spec first. Show the questions after they say they have read it.
A list of questions shown first makes people stop reading closely.

Answer the questions with the human, as one more interview round.
Update the spec, and run `flow validate` again.

Done when: every question is answered and the spec is updated.

## 7. Approval

Show the human a short summary: what the slice does, the number of
requirements, what is out of scope, and the decisions.
Show the full spec when asked.
The human approves with `flow approve spec`. Run it only when they
have just told you to.

Done when: the spec is approved.
