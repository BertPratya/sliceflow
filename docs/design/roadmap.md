# Roadmap

The order of the slices, and what comes after. The design behind each
slice is in `sliceflow-design.md`. The Design column names its sections.

## How the slices are cut (decided, the human)

1. Each slice gives something to try: a real command, run on a test repo
   in the preview.
2. The slices follow the order of the workflow: each one uses the one
   before it.
3. Each slice stays under about 25-30 requirements.
4. **The state machine grows slice by slice** (design 1.4). 0002 builds
   the engine once. Each later slice adds only the moves of its own
   step, never the whole loop ahead of time. The marks (blocked,
   cancelled) come with the slices that need them.
5. **Every step sliceflow runs can be stopped and run again** (design
   1.4). This starts with 0003, the first slice where sliceflow runs a
   step by itself. Each slice that adds a step proves it in its tests.
6. **A slice that grows too big is split** (2026-11-25, the human). The
   new slice takes a letter after the number it follows, e.g. 0009a, as
   0001a did, so no other number changes: finished specs name later
   slices by number. 0001 and 0002 stay as they are.

## The slices

The roadmap has two parts. **Part A is required**: what sliceflow needs
to be a complete product. **Part B is optional**: for later, when users
ask for it.

Why this order:

- **Agents come before the plan.** So the spec is reviewed by an agent
  before the plan phase starts, and each review is built once, in the
  order of the workflow (rule 2).
- **The dashboard and web actions come right after the run.** Web
  control is what sets sliceflow apart, and a dashboard needs a live run
  to show.
- **Claude Code is the first agent tool.** Codex comes in 0008. The
  first release needs both.
- **One release, at the end of Part A** (decided 2026-11-11, the human).
  sliceflow is released once, when Part A is done, so the first thing
  people try is the complete product. Renumbered the same day: the
  release moved from 0009 to 0014, and finish (was 0010) moved to
  Part B. The date is set in Timeline, below (2026-11-24).

### Part A: required

| # | Slice | What you can do after it | Design |
|---|---|---|---|
| 0001 | slice and events | create a slice (`sliceflow slice create`), with its branch, its worktree outside the repo (another folder can be chosen with the `worktree.dir` setting), and its spec from the built-in template; see it (`sliceflow slice status`, `sliceflow slice list`); the event log with its first event, `slice.created`, written under a file lock (on Windows too, with `LockFileEx`) | 1.13, 1.15, 2.1 |
| 0001a | plan on base | plan a slice on its base branch, in your own checkout: `slice create` writes and commits only the slice's folder there, with no branch and no worktree; the base comes from `--from`, then the `sliceflow.base` setting, then `main`; `create` refuses when the checkout is not on the base; `slice status` and `slice list` read only your checkout, and warn about a log from another branch. Changes 0001; added 2026-11-19, with no renumbering | 1.13, 1.15, 2.1 |
| 0002 | spec | the state machine's engine; check a spec like a compiler (`sliceflow spec validate`); print its template (`sliceflow spec template`); see a slice's state and what comes next; set or change a slice's title and description (`sliceflow slice title` is the first command; the one for the description is settled in 0002), the first command that adds an event to an existing log. No approval yet: the spec, plan, and tasks are approved once, together, in 0004 | 1.4, 2.2 |
| 0003 | agents | `sliceflow` starts agents itself, and keeps track of them: start Claude Code on a slice in the background (`sliceflow agent start`), list a slice's agent runs with their state, model, time, and usage (`sliceflow agent list`), read a run's conversation, newest page first (`sliceflow agent show`), and stop a run and its child processes (`sliceflow agent stop`); one function starts every run, whoever asks; the Claude Code adapter; a reader per agent tool that reads the tool's own session files, the single source of truth, so `sliceflow` saves no copy of a conversation; supervision of each run (alive, silent too long, over its time limit); the run's session, tool, model, role, and usage saved as events in the slice's log; the permission list for a run; who caused each event (`by`); `SLICEFLOW_SLICE` for agents sliceflow starts. Settings per role moved to 0005, and binding the lead to 0009a; the spec review to 0003a, and the live page to 0003b (2026-11-25, the human) | 1.7, 1.8, 1.9 |
| 0003a | spec review | `sliceflow spec submit`: an agent reviews the spec, read-only; its answer is JSON, checked against a schema, and saved as an event; `sliceflow` sets the verdict from the findings; `slice status` shows the result and what comes next. Split from 0003 on 2026-11-25, with no renumbering | 1.9, 2.2 |
| 0003b | live view | a page in the browser, on 127.0.0.1 with a token, that streams a running agent's events live, and opens an old run a page at a time, newest first, through the same readers as `sliceflow agent show`. Split from 0003 on 2026-11-25, with no renumbering | 1.8 |
| 0004 | plan | write the plan and tasks; an agent reviews them; approve the spec, plan, and tasks once, together, with the terminal prompt (the approval records `via`); the slice is ready to build; the preview checklist is built from the spec; `spec validate` reports a requirement or scenario number reused for something new after approval; the rules for changing an approved document | 1.6, 2.3 |
| 0005 | the run | `sliceflow run` builds an approved plan, one task at a time: it makes the slice's branch and worktree the first time it starts (until then a slice has none, 0001a); worktrees reused, with setup; a commit per round, a `--no-ff` merge per task; checks, task review, fix rounds; `sliceflow stop`, continue, and retry of a failed task; "Needs you"; a reached rate limit pauses the run; `sliceflow diff` by state (the whole slice, one task, one review round, since an approval); settings per role in `sliceflow.yaml` (personal overrides in git config), moved from 0003 | 1.8, 1.15, 2.4 |
| 0006 | dashboard | `sliceflow serve`, grown from the live page of 0003b: every slice live in the browser, on 127.0.0.1 with a token: tasks, agent activity grouped by slice, usage, "Needs you", and the diff view (changed files, a colored diff, the comparison picker, from the same core as `sliceflow diff`) | 1.8 |
| 0007 | web actions | stop, continue, and retry from the browser; approve or reject in the browser (`via: browser`), with the document and what changed; browser notifications; open a worktree in the editor (`sliceflow open`) | 1.6, 1.8, 2.4 |
| 0008 | Codex | a second agent tool, through the Codex app server; proves the adapter shape; one tool can review what the other wrote; binding and the sandbox for Codex | 1.9 |
| 0009 | cases and the lead | each stop goes to the right level: the case shape, routing by kind, the chat lead through `sliceflow wait`, the fallback lead (always on), and "blocked: needs the human" | 1.6, 1.7 |
| 0009a | chats | each slice has its own chats, so you never search for the right one: a chat bound to a slice (`bind` and `unbind` for a lead in a terminal, one function that picks the slice, the slice name optional in every command), moved from 0003; the slice's page in the dashboard lists its chats; open a slice's chat in the agent's own app: Claude Code Remote Control on claude.ai/code, the Codex app, and Claude Desktop if it can open a given session (to check). Every way to open a chat is optional. Added 2026-11-25, with no renumbering | 1.7 |
| 0010 | reliability | a run survives a killed `sliceflow`, a lost network, a stuck agent: heartbeats, time limits per agent run, retries, checking the log against reality on restart; `sliceflow retry` for any failed or stopped step; taking over a lock left by a crashed `sliceflow` | 1.8, 2.4 |
| 0011 | parallel tasks | independent tasks run at the same time: worker slots (one reused worktree each), `touches` overlap, merge queue, conflicts | 2.4 |
| 0011a | stacking | plan a group of slices, where each one needs the code of the slice before it, e.g. one feature split into slices; review and approve the group once; `sliceflow` builds the slices one after another, by itself. Open, decided when 0011a starts: how a slice builds on the one below it before that one is merged (from its finished branch, with updates merged in, never rebased), or waits for the merge. Moved from Part B on 2026-11-25, with no renumbering | 2.6 |
| 0012 | configurable steps | pick and order the built-in steps | 1.10 |
| 0013 | budgets, cancel, cleanup | budget caps in tokens per task and per slice; cancel and reopen a slice; clean up a slice half made by a crash | 1.9, 2.7 |
| 0014 | first release | `sliceflow init`, the lead's guide as a skill, `--version`, release binaries for Windows, macOS, and Linux, a README with the pitch, a demo GIF and video, and a website for sliceflow | 1.1, 1.7 |

### Timeline (decided 2026-11-24, the human)

All of Part A is built in 12 weeks from 2026-11-24, so the first public
release, 0014, is around **2027-02-16**. A month of feedback and fixes
from the first users follows, until around mid-March 2027.

| Week | From | Slices |
|---|---|---|
| 1 | 2026-11-24 | 0002 spec |
| 2-3 | 2026-12-01 | 0003 agents |
| 4 | 2026-12-15 | 0004 plan |
| 5-6 | 2026-12-22 | 0005 the run |
| 7 | 2027-01-05 | 0006 dashboard |
| 8 | 2027-01-12 | 0007 web actions, 0012 configurable steps |
| 9 | 2027-01-19 | 0008 Codex, 0009 cases and the lead |
| 10 | 2027-01-26 | 0010 reliability, 0013 budgets, cancel, cleanup |
| 11 | 2027-02-02 | 0011 parallel tasks |
| 12 | 2027-02-09 | 0014 first release |

- 0003a, 0003b, 0009a, and 0011a, added 2026-11-25, are not in this
  table yet. The checkpoint at the end of week 4 places them, and the
  chat page if it joins Part A.
- **How:** agents build at night, from the day's spec, plan, and tasks;
  the human decides in the evening and reviews by risk in the morning
  (AGENTS.md). The human's hours, not the agents', are the limit.
- **Estimate when it was set:** about 550-840 hours left (likely about
  700), more than 12 weeks of the human's own hours; the night runs
  close the gap.
- **Checkpoints:** at the end of week 4 (0002 to 0004 done) and week 6
  (the run works end to end), the human compares the progress with this
  table, and decides what to do if it is behind: move the date, or move
  hours. 0010 (processes and restarts) and 0011 (parallel tasks) are the
  slices most likely to run long.

### Part B: optional

Numbered when they are picked. No fixed order.

| Slice | What you can do after it | Design |
|---|---|---|
| finish | the full path to `main`: integration review, the preview, result approval, update from `main`, then a local `--no-ff` merge or a pull request (`finish: pr`, the push approved by the human); deferred findings go to the backlog. Moved from Part A on 2026-11-11; whether it comes back is decided later | 2.4, 2.5 |
| stable output | errors as JSON with codes; the event log and `--json` documented and versioned; a parsed document as JSON (e.g. `sliceflow spec show --json`) | 1.13 |
| user hooks | your own hooks before and after events, with a trust check before a repo's hooks run | - |
| team view | teammates' slices from remote branches, and whose slice it is | - |
| configurable templates | a project changes the template text and the reading rules | 1.13 |
| Spec Kit import and files | turn a Spec Kit feature into a slice once, or read Spec Kit files directly | 1.13 |
| GitHub issue to slice | `sliceflow slice create --from-issue 42`, through the local `gh` | - |
| MCP server | agent tools see sliceflow's commands as tools | - |
| container profile | a run inside a container, for full autonomy with a safe wall | 1.9 |
| analytics | rounds, time, and tokens per phase, over many slices | - |
| more agent tools | e.g. Gemini CLI, OpenCode | 1.9 |
| diff view, step 2 | side-by-side, syntax colors, and comments on lines that go to the agent as review findings | 1.8 |
| doctor | `sliceflow doctor`: checks a repo is ready, and cleans up worktrees left behind | 1.15 |

### Notes

- **0014 is the first public release.** It comes after all of Part A.
  So the first thing people try is a slice built and controlled from the
  browser, with Claude Code or Codex, that survives a crash, and runs
  tasks in parallel. Without finish (Part B), merging into `main` is
  done by hand with git.
- **0002 to 0004 make one planning phase with one approval.** 0002 ends
  at a checked spec. 0003a adds the spec review by an agent. 0004 ends at
  the approval of the spec, plan, and tasks together.
- **Windows before Codex.** Native Windows has no sandbox for Claude
  Code (design 1.9). So until 0008, a run on Windows has only layers 1
  and 3 of the permissions.
- **The biggest unknowns** are 0003 (running agents headless, and the
  permissions without a sandbox on Windows) and 0010 (processes and
  restarts on three systems). A small prototype before each of them
  lowers the risk.
- **sliceflow's own chat page** (2026-11-25, the human): a chat with
  an agent inside the dashboard, drawn from its headless stream, so a
  user needs nothing but a browser. Optional, like every way to open a
  chat (0009a). Whether it goes in Part A, after 0009a, or in Part B is
  decided after the 0003 prototype shows whether sending messages to a
  running Claude Code works well. Codex's app server is two-way by
  design.
- The numbers, order, and contents of the later slices may still change
  when we get there.

## After the slices

- **Feedback month.** After the first release, about a month for what
  the first users find (see Timeline).
- **Use it for real.** After the first release, run real slices with
  `sliceflow` on this repo and one other project, and fix what hurts.
  Then a write-up with real numbers: time, tokens, and how often the lead
  or the human was needed. This is an activity, not a slice.
- **A portfolio website** for the human, made with AI. It is not part of
  sliceflow, so it is an activity, not a slice.
