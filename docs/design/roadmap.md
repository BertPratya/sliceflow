# Roadmap

The order of the slices, and what comes after. The design behind each
slice is in `sliceflow-design.md` (the Design column names its sections).

## How the slices are cut (decided, the human, 2026-10-19)

1. Each slice gives something to try: a real command, run on a test repo
   in the preview.
2. The slices follow the order of the workflow: each one uses the one
   before it.
3. Each slice stays under about 25-30 requirements.
4. **The state machine grows slice by slice** (decided, the human,
   2026-10-24). 0002 builds the engine once: how a move is described
   (an event, when it is allowed, the state it moves to), replaying the
   log through the moves, refusing a move that is not allowed, and
   showing where a slice is and what comes next. Each slice then adds
   only the moves of its own step, never the whole loop ahead of time.
   The engine lets a mark (blocked, cancelled) sit on top of any phase
   (design 1.4); the marks themselves come with the slices that need
   them.
5. **Every step sliceflow runs can be stopped and run again** (decided,
   the human, 2026-10-24). From 0004 on, each step sliceflow runs by
   itself has the events started, then done, failed, or stopped, and is
   safe to run twice. Retrying a failed or stopped step is then one
   more move of the state machine. Each slice that adds a step proves
   this in its tests.

## The slices (replanned, the human, 2026-10-29)

The roadmap has two parts. **Part A is required**: what sliceflow needs
to be a complete product. **Part B is optional**: for later, when users
ask for it. The order changed after a product review (2026-10-28): the
dashboard and web actions come right after the run, because web control
is what sets sliceflow apart, and a dashboard needs a live run to show.
The decisions behind the new contents are rows in `docs/backlog.md`
(2026-10-28).

### Part A: required

| # | Slice | What you can do after it | Design |
|---|---|---|---|
| 0001 | slice and events | create a slice (`sliceflow slice create`), see it (`sliceflow slice status`, `sliceflow slice list`); the event log with its first event, `slice.created` | 1.13, 2.1 |
| 0002 | spec | the state machine's engine, and the spec's moves: write a spec and check it; see where a slice is and what comes next; set a slice's title; approve with the terminal prompt (an approval records `via`) | 1.4, 2.2 |
| 0003 | plan | write the plan and tasks (the 2026-10-28 templates), review them, and approve the planning once; the slice is ready to build; the preview checklist is built from the spec; `spec validate` reports a requirement or scenario number reused for something new after approval | 2.3 |
| 0004 | agents | `sliceflow` starts agents itself: the Codex adapter, the permission layers, settings per role in `sliceflow.yaml` (personal overrides in git config); reviews answer in JSON and are saved as events; every agent run records its usage; first use: the spec check and the plan review; who caused each event (`by`); each session sliceflow starts is bound to its slice | 1.9, 1.12c |
| 0005 | the run | `sliceflow run` builds an approved plan, one task at a time: worktrees outside the repo, reused, with setup; a commit per round, a `--no-ff` merge per task; checks, task review, fix rounds; `sliceflow stop`, continue, and retry of a failed task; "Needs you"; a reached rate limit pauses the run; `sliceflow diff` by state (the whole slice, one task, one review round, since an approval) | 2.4 |
| 0006 | dashboard | `sliceflow serve`: every slice live in the browser, on 127.0.0.1 with a token: tasks, agent activity, usage, "Needs you", and the diff view (changed files, a colored diff, the comparison picker, from the same core as `sliceflow diff`) | 1.7 |
| 0007 | web actions | stop, continue, and retry from the browser; approve or reject in the browser, with the document and what changed; browser notifications; open a worktree in the editor (`sliceflow open`) | 1.7, 2.4 |
| 0008 | first release | `sliceflow init`, the lead's guide as a skill, `--version`, release binaries for Windows, macOS, and Linux, a README with the pitch, and a demo | - |
| 0009 | finish | the full path to `main`: integration review, the preview, result approval, update from `main`, then a local `--no-ff` merge or a pull request (`finish: pr`, the push approved by the human), backlog | 2.5 |
| 0010 | cases and the lead | each stop goes to the right level: the case shape, routing by kind, the chat lead through `sliceflow wait`, the fallback lead (always on), and "blocked: needs the human" | 1.6, 1.12c |
| 0011 | Claude Code | a second agent tool; proves the adapter shape; one tool can review what the other wrote | 1.9 |
| 0012 | reliability | a run survives a killed `sliceflow`, a lost network, a stuck agent: heartbeats, time limits, retries, checking the log against reality on restart; `sliceflow retry` for any failed or stopped step; taking over a lock left by a crashed `sliceflow` | 1.7, 2.4 |
| 0013 | parallel tasks | independent tasks run at the same time: worker slots (one reused worktree each), `touches` overlap, merge queue, conflicts | 2.4 |
| 0014 | configurable steps | pick and order the built-in steps | 1.10 |
| 0015 | budgets, cancel, cleanup | budget caps in tokens per task and per slice; cancel and reopen a slice; clean up a slice half made by a crash | 2.7 |

### Part B: optional

Numbered when they are picked. No fixed order.

| Slice | What you can do after it | Design |
|---|---|---|
| stacking | design a slice ahead, on top of another | 2.6 |
| stable output | errors as JSON with codes; the event log and `--json` documented and versioned; a parsed document as JSON (e.g. `sliceflow spec show --json`) | 1.13b |
| user hooks | your own hooks before and after events, with a trust check before a repo's hooks run | - |
| team view | teammates' slices from remote branches, and whose slice it is | - |
| configurable templates | a project changes the template text and the reading rules | 1.13c |
| Spec Kit import and files | turn a Spec Kit feature into a slice once, or read Spec Kit files directly | 1.13c |
| GitHub issue to slice | `sliceflow slice create --from-issue 42`, through the local `gh` | - |
| MCP server | agent tools see sliceflow's commands as tools | - |
| container profile | a run inside a container, for full autonomy with a safe wall | 1.9 |
| analytics | rounds, time, and tokens per phase, over many slices | - |
| more agent tools | e.g. Gemini CLI, OpenCode | 1.9 |
| diff view, step 2 | side-by-side, syntax colors, and comments on lines that go to the agent as review findings | - |

### Notes

- **0008 is the first public release.** It comes after web control, so
  the first thing people try is a slice built and controlled from the
  browser. Until 0009, merging into `main` is done by hand with git;
  until 0011, it runs on Codex only.
- **0002 and 0003:** since 2026-10-21 there is one planning phase with
  one approval, so 0002 ends at a checked spec, and 0003 ends at the
  approval.
- **Codex first:** 0004 builds the Codex adapter; Claude Code comes in
  0011.
- **0001 changes before the first release:** worktrees move outside the
  repo (`~/.sliceflow/worktrees/<repo>/<slice>/`), at the latest with
  0005.
- The numbers, order, and contents of the later slices may still change
  when we get there.

## Where the backlog items go (updated 2026-10-29)

| Backlog item | Goes to |
|---|---|
| Set or change a slice's title | 0002 |
| Approvals record `via`; the terminal prompt | 0002, 0003; the browser in 0007 |
| Reviews as JSON events | 0004 |
| Usage recorded per agent run | 0004; shown in 0006; caps in 0015 |
| Settings in `sliceflow.yaml` and git config | 0004 |
| Who caused an event (`by`) | 0004 |
| Bind a session to a slice by its session ID | 0004 for the sessions sliceflow starts; 0011 for Claude Code; the lead's binding decided by 0010 |
| Commit rules during a run; worktrees reused, outside the repo, with setup | 0005 |
| You stop a run, continue it, retry a failed task | 0005 |
| `sliceflow diff` by state | 0005; the browser view in 0006 |
| A local server and browser app; stop, continue, and retry buttons | 0006, 0007 |
| Sessions grouped by slice in the browser | 0006 |
| Finish with a pull request | 0009 |
| The fallback lead, "blocked" | 0010 |
| You retry any failed or stopped step; remove a lock left by a crash | 0012 |
| Cancel a slice; clean up a slice half made by a crash; budget caps | 0015 |
| User hooks, errors as JSON, teammates' slices | Part B |

## Time (the lead's estimate, 2026-10-29)

The human's hours, at about 22 hours a week on sliceflow. 0002 and 0003
are built in fast mode (agents write the code). From 0004 on, slices are
built in pair mode: the human writes pieces and learns Go, and writes a
small prototype before each hard slice, so those estimates are about
twice the 2026-10-23 ones.

| # | Slice | Hours |
|---|---|---|
| 0002 | spec | 20-25 |
| 0003 | plan | 25-30 |
| 0004 | agents | 55-75 |
| 0005 | the run | 60-80 |
| 0006 | dashboard | 45-60 |
| 0007 | web actions | 30-45 |
| 0008 | first release | 15-30 |
| 0009 | finish | 35-50 |
| 0010 | cases and the lead | 35-50 |
| 0011 | Claude Code | 25-40 |
| 0012 | reliability | 45-65 |
| 0013 | parallel tasks | 40-55 |
| 0014 | configurable steps | 30-40 |
| 0015 | budgets, cancel, cleanup | 25-35 |

- **Up to web control (0002-0007):** about 235-315 hours.
- **Up to the first release (0008):** about 250-345 hours.
- **All of Part A:** about 485-680 hours.
- **The budget until February 2027** (when job applications start) is
  about 285 hours. The path to the first release does not fit safely.
  Open: what to shrink, about 30-50 hours. Candidates from the review:
  in 0004, one review type run by a real agent first, and no Windows
  sandbox work (document it); in 0005, retry only a failed task, and
  no time limits yet; in 0006-0007, agents build the browser pages, and
  the diff is plain; in 0002-0003, only what 0005 needs; in 0008, the
  minimum (`init`, binaries, README, demo). Each slice's spec settles
  its own cuts.
- The biggest unknowns are 0004 (running agents headless, the sandbox on
  Windows) and 0012 (processes and restarts on three systems).

## Between and after the slices

- **Use it for real.** After the first release, run real slices with
  `sliceflow` on this repo and one other project, and fix what hurts.
  Then a write-up with real numbers: time, tokens, how often the lead or
  the human was needed. This is an activity, not a slice.
