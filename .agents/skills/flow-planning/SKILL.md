---
name: flow-planning
description: Runs a slice's planning phase with the human, starting with its spec. Use when a slice is in its planning phase, or when the human wants to write or change a slice's spec.
---

# flow-planning

The slice was created by `flow-slice`. Work in its worktree, on
`docs/slices/<name>/`. Read `docs/templates/spec.md` before writing the
spec: its comments hold the spec's rules.

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
  Q1, Q2, ...
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
2. Start a plan reviewer for a spec check (`AGENTS.md` → Tools per role).
   It looks for conflicts and unresolved points in `spec.md`.
3. The reviewer gives each finding a level:
   - **high:** needs the human's decision. Add it to Open questions.
   - **medium:** you can decide it. Fix it with the mark `(lead's call)`,
     and tell the human.
   - **low:** a small finding. Fix it or leave it, as you see fit. It
     never goes to Open questions.
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

Commit after every step that changes a file: `<name>: <what happened>`.
When `spec.md` is `review_passed`, the plan can start.
