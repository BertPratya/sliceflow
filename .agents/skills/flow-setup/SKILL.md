---
name: flow-setup
description: Checks whether a repo is ready for the agent-flow workflow, then sets up what is missing, with the human choosing what to fix or skip. Use when adding the workflow to a repo, when updating it to a newer version, or when the human asks if the repo is ready.
---

# flow-setup

You are the lead. Follow these steps in order. Each step ends when its
"Done when" is true. If `flow` does not exist yet, use `bootstrap.md` in
this folder wherever a step names a `flow` command.

Setup adds workflow files, settings, and hooks. It never changes the
project's code. A gap that needs code changes goes to `backlog.md`, to be
done later as a slice.

## 1. Check

Run `flow doctor`. It checks every item in `readiness.md` and reports each
one as ok, missing, out of date, or skipped (a known gap).

Done when: you have the report.

## 2. Report and choose

Give the human the report. List the required items first.
For every item that is not ok, show:

1. **Found:** what is wrong, in plain words.
2. **Why it matters:** what goes wrong because of it.
3. **Options**, numbered, with your recommendation first.
   Recommended and optional items always include "skip".
4. **If skipped:** what it costs, and how the workflow works around it.

`readiness.md` lists the usual options for each item. Adapt them to what
you found. Ask in rounds of at most about 5 items.

Required items have no skip option. If one cannot be fixed now, setup
stops there: "Ready for slices: no".

Done when: the human has chosen an option for every item that is not ok.

## 3. Learn the project

Facts are your job. Before filling the project sections, find:
- the parts: build files (`go.mod`, `package.json`, …) and folders
- contracts (OpenAPI, AsyncAPI, protobuf, GraphQL) and a data model doc
- existing checks: Makefile targets, package scripts, lint configs, CI files
- test dependencies: what the tests connect to, from code, configs,
  environment variables, Docker Compose, and CI files
- existing rules: `CONTRIBUTING.md`, code style configs

Then ask the human, in rounds with recommendations, only what the repo
cannot tell you: the git policy (base branch, personal branches, who
merges), the tools for each role, and project rules.

Done when: you can fill `AGENTS.md` sections 8-10 without guessing.

## 4. Set up

Do what the human chose, in this order. Skip what is already ok.

1. **Workflow files**, from the source the human gives (`bootstrap.md`)
   or `flow init`: the `AGENTS.md` block between the `agent-flow` markers,
   `workflow/`, and `.agents/skills/`. Never touch sections 8-10 when updating.
2. **`AGENTS.md` sections 8-10:** project, stack profile (parts, connections,
   tools per role, always allowed files, auto-fix commands, checks with tiers, test dependencies,
   known gaps), and project rules with the git policy.
3. **Tool adapters** for the tools the team uses, as `workflow/adapters/README.md` says.
4. **Hooks and CI:** pre-commit runs `flow sync` and `flow validate`; Stop hooks
   run `flow check`; a CI job runs the task checks and `flow validate`.
   A CI change needs the human's OK.
5. **Repo basics:** `slices/`, `backlog.md`, `.gitattributes`
   (`slices/*/events.jsonl merge=union`, `*.md text eol=lf`), and `.gitignore`
   with `.flow/`.
6. **Known gaps:** every skipped item, with its effect on the workflow,
   who chose it, and the date.
7. **Gaps that need code changes:** for each one the human chose to fix,
   add one line to `backlog.md`: the goal, "from setup", and the date.
   It becomes a slice when someone starts it with `flow-spec`.

Commit on its own branch, as the git policy says, so the human can review it.

Done when: every chosen item is done or is in `backlog.md`.

## 5. Check again

Run `flow doctor` again.

Done when: every required item is ok, and every other item is ok, skipped
as a known gap, or is in `backlog.md`. Tell the human "Ready for slices: yes",
and which backlog item to start first.
