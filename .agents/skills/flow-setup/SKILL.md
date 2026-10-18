---
name: flow-setup
description: Checks whether a repo is ready for the agent-flow workflow, then sets up what is missing, with the human choosing what to fix or skip. Use when adding the workflow to a repo, when updating it to a newer version, or when the human asks if the repo is ready.
---

# flow-setup

You are the lead. Follow these steps in order. Each step ends when its
"Done when" is true.

Setup adds workflow files, settings, and hooks. It never changes the
project's code. A gap that needs code changes goes to `backlog.md`, to be
done later as a slice.

## 1. Check

Go through every item in `readiness.md`, in order. For each one, look, and
write one line: ok, missing, out of date, or skipped (listed under "Known
gaps" in `AGENTS.md` section 9). To compare versions, read `workflow/VERSION`
in this repo and in the source.

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

For each role, find how to start it from the lead's tool: a subagent of
your own tool, or another tool's command line (e.g. `codex exec`). Each
tool does it differently. Try it once with a prompt from a file. For the
worker, the prompt asks it to run the task checks of one part: a sandbox
may block the network, caches, or Docker, and a tiny prompt would not show
it. Write the way that worked, with any sandbox options or cache settings
it needed, into the "How to start" column of the stack profile. If the
human prefers, leave a row empty: the lead fills it the first time it
needs that role.

Ask the human how work should be done in this project, beyond the
workflow: e.g. how to split tasks, how to build UI screens, how to test,
code style. Propose rules from what you found in the repo. Write each one
as "<role>, <when>: <rule>" (e.g. "Worker, UI tasks: use only the colors
in `web/src/styles/tokens.css`"), so each role knows which rules are its
own. A rule that needs a command (a screenshot script, a lint rule) is
tried once, like "How to start". If the command writes files (screenshots,
reports), add their folder to `.gitignore`, and say in the rule where the
reviewer finds them.
For task splitting, show the defaults in `flow-plan` step 7 and ask which to
change; write only the changes.

Done when: you can fill `AGENTS.md` sections 8-10 without guessing.

## 4. Set up

Do what the human chose, in this order. Skip what is already ok.

1. **Workflow files.** Ask the human for the source: a path or URL of the
   agent-flow repo, at a version or tag. From it, copy:
   - the text between `<!-- agent-flow:start … -->` and `<!-- agent-flow:end -->`
     into this repo's `AGENTS.md`. If the markers exist, replace only what is
     between them. If not, add the block at the top and sections 8-10 below it.
   - `workflow/` (templates, roles, formats, adapters, `VERSION`)
   - `.agents/skills/`

   Never copy the source's sections 8-10, and never touch this repo's
   sections 8-10 when updating.
2. **`AGENTS.md` sections 8-10**, from `workflow/templates/project.md`:
   project, stack profile (parts with their folders, connections, tools per
   role with how to start each one, setup commands, always allowed files,
   auto-fix commands, checks with tiers, run locally, test dependencies, known gaps), and
   project rules: the git policy, and the other rules as "<role>, <when>: <rule>",
   or "(none)".
3. **Tool adapters** for the tools the team uses, as `workflow/adapters/README.md` says.
4. **Hooks and CI:** a pre-commit hook runs the auto-fix commands and the
   task checks; Stop hooks run the task checks, where the tool supports them;
   a CI job runs the task checks. A CI change needs the human's OK.
5. **Repo basics:** `slices/`, `backlog.md`, `.gitattributes`
   (`*.md text eol=lf` and `*.patch text eol=lf`), and `.gitignore` with
   `.worktrees/`. On Windows: `git config core.longpaths true`, since
   folders like `node_modules` inside `.worktrees/` pass the 260-character limit.
6. **Known gaps:** every skipped item, with its effect on the workflow,
   who chose it, and the date.
7. **Gaps that need code changes:** for each one the human chose to fix,
   add one line to `backlog.md`: the goal, "from setup", and the date.
   It becomes a slice when someone starts it with `flow-spec`.

Commit on its own branch, as the git policy says, so the human can review it.

Done when: every chosen item is done or is in `backlog.md`.

## 5. Check again

Go through `readiness.md` again.

Done when: every required item is ok, and every other item is ok, skipped
as a known gap, or is in `backlog.md`. Tell the human "Ready for slices: yes",
and which backlog item to start first.
