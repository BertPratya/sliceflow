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
  Part B.

### Part A: required

| # | Slice | What you can do after it | Design |
|---|---|---|---|
| 0001 | slice and events | create a slice (`sliceflow slice create`), with its branch, its worktree outside the repo (another folder can be chosen with the `worktree.dir` setting), and its spec from the built-in template; see it (`sliceflow slice status`, `sliceflow slice list`); the event log with its first event, `slice.created`, written under a file lock (on Windows too, with `LockFileEx`) | 1.13, 1.15, 2.1 |
| 0002 | spec | the state machine's engine; check a spec like a compiler (`sliceflow spec validate`); print its template (`sliceflow spec template`); see a slice's state and what comes next; set or change a slice's title and description (`sliceflow slice title` is the first command; the one for the description is settled in 0002), the first command that adds an event to an existing log. No approval yet: the spec, plan, and tasks are approved once, together, in 0004 | 1.4, 2.2 |
| 0003 | agents | `sliceflow` starts agents itself: the Claude Code adapter, the permission layers, settings per role in `sliceflow.yaml` (personal overrides in git config); reviews answer in JSON and are saved as events; every agent run records its usage; who caused each event (`by`); sessions bound to a slice (`SLICEFLOW_SLICE` for agents sliceflow starts, `bind` and `unbind` for the lead, one function that picks the slice, the slice name optional in every command); first use: `sliceflow spec submit` and the spec review | 1.7, 1.9, 1.14 |
| 0004 | plan | write the plan and tasks; an agent reviews them; approve the spec, plan, and tasks once, together, with the terminal prompt (the approval records `via`); the slice is ready to build; the preview checklist is built from the spec; `spec validate` reports a requirement or scenario number reused for something new after approval; the rules for changing an approved document | 1.6, 2.3 |
| 0005 | the run | `sliceflow run` builds an approved plan, one task at a time: worktrees reused, with setup; a commit per round, a `--no-ff` merge per task; checks, task review, fix rounds; `sliceflow stop`, continue, and retry of a failed task; "Needs you"; a reached rate limit pauses the run; `sliceflow diff` by state (the whole slice, one task, one review round, since an approval) | 1.8, 1.15, 2.4 |
| 0006 | dashboard | `sliceflow serve`: every slice live in the browser, on 127.0.0.1 with a token: tasks, agent activity grouped by slice, usage, "Needs you", and the diff view (changed files, a colored diff, the comparison picker, from the same core as `sliceflow diff`) | 1.8 |
| 0007 | web actions | stop, continue, and retry from the browser; approve or reject in the browser (`via: browser`), with the document and what changed; browser notifications; open a worktree in the editor (`sliceflow open`) | 1.6, 1.8, 2.4 |
| 0008 | Codex | a second agent tool, through the Codex app server; proves the adapter shape; one tool can review what the other wrote; binding and the sandbox for Codex | 1.9 |
| 0009 | cases and the lead | each stop goes to the right level: the case shape, routing by kind, the chat lead through `sliceflow wait`, the fallback lead (always on), and "blocked: needs the human" | 1.6, 1.7 |
| 0010 | reliability | a run survives a killed `sliceflow`, a lost network, a stuck agent: heartbeats, time limits per agent run, retries, checking the log against reality on restart; `sliceflow retry` for any failed or stopped step; taking over a lock left by a crashed `sliceflow` | 1.8, 2.4 |
| 0011 | parallel tasks | independent tasks run at the same time: worker slots (one reused worktree each), `touches` overlap, merge queue, conflicts | 2.4 |
| 0012 | configurable steps | pick and order the built-in steps | 1.10 |
| 0013 | budgets, cancel, cleanup | budget caps in tokens per task and per slice; cancel and reopen a slice; clean up a slice half made by a crash | 1.9, 2.7 |
| 0014 | first release | `sliceflow init`, the lead's guide as a skill, `--version`, release binaries for Windows, macOS, and Linux, a README with the pitch, a demo GIF and video, and a website for sliceflow | 1.1, 1.7 |

### Part B: optional

Numbered when they are picked. No fixed order.

| Slice | What you can do after it | Design |
|---|---|---|
| finish | the full path to `main`: integration review, the preview, result approval, update from `main`, then a local `--no-ff` merge or a pull request (`finish: pr`, the push approved by the human); deferred findings go to the backlog. Moved from Part A on 2026-11-11; whether it comes back is decided later | 2.4, 2.5 |
| stacking | design a slice ahead, on top of another | 2.6 |
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
  at a checked spec. 0003 adds the spec review by an agent. 0004 ends at
  the approval of the spec, plan, and tasks together.
- **Windows before Codex.** Native Windows has no sandbox for Claude
  Code (design 1.9). So until 0008, a run on Windows has only layers 1
  and 3 of the permissions.
- **The biggest unknowns** are 0003 (running agents headless, and the
  permissions without a sandbox on Windows) and 0010 (processes and
  restarts on three systems). A small prototype before each of them
  lowers the risk.
- The numbers, order, and contents of the later slices may still change
  when we get there.

## After the slices

- **Use it for real.** After the first release, run real slices with
  `sliceflow` on this repo and one other project, and fix what hurts.
  Then a write-up with real numbers: time, tokens, and how often the lead
  or the human was needed. This is an activity, not a slice.
- **A portfolio website** for the human, made with AI. It is not part of
  sliceflow, so it is an activity, not a slice.
