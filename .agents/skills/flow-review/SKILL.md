---
name: flow-review
description: Starts a review in a fresh context and reports the result: spec check, plan review, task review, or integration review. Use when a review is due, or when the human asks to review a spec, plan, task, or slice.
---

# flow-review

You start the review. A fresh reviewer does it. Follow these steps in order.
If `flow` does not exist yet, use `bootstrap.md` in this folder wherever a
step names a `flow` command.

## 1. Pick the kind

| Kind | When | Reviewer role | Target |
|---|---|---|---|
| `spec` | the spec draft is ready, before approval | plan-reviewer | the spec revision |
| `plan` | plan, contracts, and tasks are written | plan-reviewer | the plan revision |
| `task` | a task passed its checks and needs review (see Tracks in `AGENTS.md`) | code-reviewer | the task and its commits |
| `integration` | every task is done | code-reviewer | the whole slice |

Done when: you know the kind and the target.

## 2. Create the review file

Run `flow review start <kind> <slice>/<target>`. It creates
`reviews/<kind>-<target>-<round>.md` from the template and fills the header.

Done when: the review file exists with its header filled.

## 3. Start the reviewer in a fresh context

Never review in your own context. You wrote or ran the work, so you
share its blind spots.

Use the tool the project sets for the role (stack profile in `AGENTS.md`).
If it is your own tool, start a subagent. If it is another tool, start it
through that tool's command line (e.g. `codex exec "…"`, `claude -p "…"`).
If the project sets none, use your own subagent.

The prompt is always:

    You are the <role>. Read workflow/roles/<role>.md and follow it.
    Run `flow brief <role> <slice>/<review>`.

Done when: the reviewer has ended and the review is submitted.
If it ended without submitting, start it once more. If that fails too,
tell the human.

## 4. Act on the verdict

| Verdict | Next |
|---|---|
| `pass` | The work moves on: spec or plan to approval, task to done, slice to preview. |
| `changes`, task | The worker gets another round (`flow-implement`). |
| `changes`, plan | You fix the findings (`flow-plan`, step 9). |
| `changes`, integration | `flow` turns the findings into fix tasks (`flow-implement`). |
| `needs_human` | Ask the human each question, in plain words. |

For a spec check, hold the questions until the human has read the spec
(`flow-spec`, step 6).

Report to the human in one or two lines: the verdict, and the number of
findings by level.

Done when: the next step has started, or the human has the questions.
