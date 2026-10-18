---
name: flow-review
description: Starts a review in a fresh context and reports the result: spec check, plan review, task review, or integration review. Use when a review is due, or when the human asks to review a spec, plan, task, or slice.
---

# flow-review

You start the review. A fresh reviewer does it. Follow these steps in order.
After every step that changes something, update `status.md`, add a Log
line, and commit (`workflow/roles/lead.md` → Keeping the status).
Also follow the project rules for the lead and for Everyone
(`AGENTS.md` section 10 → Other rules).

## 1. Pick the kind

| Kind | When | Reviewer role | Target |
|---|---|---|---|
| `spec` | the spec draft is ready, before approval | plan-reviewer | the spec revision |
| `plan` | plan, contracts, and tasks are written | plan-reviewer | the plan revision |
| `task` | a task passed its checks and needs review (see Tracks in `AGENTS.md`) | code-reviewer | the task and its commits |
| `integration` | every task is done | code-reviewer | the whole slice |

Done when: you know the kind and the target.

## 2. Create the review file

1. **Round:** 1 + the number of earlier reviews of the same kind and target in `reviews/`.
   A task review file is created and committed in the task worktree, on the
   task branch. All other review files go in the slice worktree.
2. Copy `workflow/templates/review.md` to `reviews/<kind>-<target>-<round>.md`
   (integration: `reviews/integration-<round>.md`).
3. Fill the header: review, slice, kind, target, round, reviewer (role and
   tool), and the title. For a task, put the commits to review in `target`.

Done when: the review file exists with its header filled.

## 3. Start the reviewer in a fresh context

Never review in your own context. You wrote or ran the work, so you
share its blind spots.

Start the reviewer the way the stack profile's "How to start" column says
(`AGENTS.md` section 9). Whatever the way: a fresh context, the prompt below, it works only in the
folder the prompt names, and only its `RESULT` block comes back to you.
Never read its whole conversation.

The prompt:

    You are the <role>. Read workflow/roles/<role>.md and follow it.
    Work only in <absolute path of the worktree: the task's for a task review, otherwise the slice's>.
    All paths are relative to it.
    This is a <kind> review. Your review file: <path>.
    <task review: Review commits <hashes> of task <NN>.>
    <integration review: Review the diff of <slice branch> against <base branch>.
     Merges with conflicts: <commits, from the Log>.>
    <plan review: The track is <track>.>
    <spec check of a change: Only R<n>, R<n> changed. Check them and what they affect.>

Done when: the reviewer has ended with `RESULT: submitted`.
If it ended without it, start it once more. If that fails too, tell the human.

## 4. Read the review and set the verdict

1. Check the lines you read have their exact form (the note at the top of
   the review template). If one cannot be read, start the reviewer again
   with the line to fix. After two failed tries, tell the human.
2. Set the verdict, by the rule in the template: anything under "Needs human decision" → `needs_human`;
   else any high or medium finding, new or reopened → `changes`; else `pass`.
3. Record the verdict (with the time, UTC) in a Log line, and every finding in `status.md` → Findings: new ones
   `open`, earlier ones `verified` or `reopened`.
4. Spec or plan review: set the document's status, as `workflow/formats/lifecycle.md`
   → Documents says: `pass` → `awaiting_approval`; `changes` or `needs_human` → `draft`.
5. Commit: `NNNN: review <review> (<verdict>)`. The review file goes on the
   branch where it was created; `status.md` only ever on the slice branch.

Done when: the verdict is set, recorded, and committed.

## 5. Act on the verdict

| Verdict | Next |
|---|---|
| `pass` | The work moves on: spec or plan to approval, task to the merge, slice to preview. |
| `changes`, task | The worker gets another round (`flow-implement`). |
| `changes`, plan | You fix the findings (`flow-plan`, step 9). |
| `changes`, integration | Each high or medium finding becomes a fix task (`flow-implement`). |
| `needs_human` | Ask the human each question, in plain words. |

For a spec check, hold the questions until the human has read the spec
(`flow-spec`, step 6).

Report to the human in one or two lines: the verdict, and the number of
findings by level.

Done when: the next step has started, or the human has the questions.
