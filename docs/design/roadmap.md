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

## The slices

| # | Slice | What you can do after it | Design |
|---|---|---|---|
| 0001 | slice and events | create a slice (`sliceflow slice create`), see it (`sliceflow slice status`, `sliceflow slice list`); the event log with its first event, `slice.created` | 1.13, 2.1 |
| 0002 | spec | the state machine's engine, and the spec's moves: write a spec and check it; see where a slice is and what comes next; set a slice's title | 1.4, 2.2 |
| 0003 | plan | write the plan, tasks, and preview, review them, and approve the planning once; the slice is ready to build | 2.3 |
| 0004 | agents | `sliceflow` starts agents itself: the Codex adapter, the permission layers, settings per role; first use: the spec check and the plan review; who caused each event (`by`) | 1.9, 1.12c |
| 0005 | the run | `sliceflow run` builds an approved plan by itself, one task at a time: worktree, worker, checks, task review, fix rounds, merge; `sliceflow stop`, and `sliceflow run` continues where it stopped | 2.4 |
| 0006 | parallel tasks | independent tasks run at the same time: worker limit, `touches` overlap, merge queue, conflicts | 2.4 |
| 0007 | reliability | a run survives a killed `sliceflow`, a lost network, a stuck agent: heartbeats, time limits, retries, checking the log against reality on restart; `sliceflow retry` for a failed or stopped step; cleaning up a slice half made by a crash | 1.7, 2.4 |
| 0008 | cases and the lead | each stop goes to the right level: the chat lead, the fallback lead, or the human | 1.6, 1.12c |
| 0009 | finish | the full path to `main`: integration review, preview, result approval, update from `main`, merge, backlog | 2.5 |
| 0010 | stacking and cancel | design a slice ahead, on top of another; cancel and reopen | 2.6, 2.7 |
| 0011 | Claude Code | a second agent tool; proves the adapter shape | 1.9 |
| 0012 | configurable steps | pick and order the built-in steps | 1.10 |

- **0009 is the first release.** After it, `sliceflow` runs the whole workflow,
  and this repo switches from the skills to `sliceflow` itself.
- **0002 and 0003** were first cut with separate spec and plan approvals.
  Since 2026-10-21 there is one planning phase with one approval, so 0002
  ends at a checked spec, and 0003 ends at the approval. Settled when
  0002's planning starts.
- **Codex first:** 0004 builds the Codex adapter; Claude Code comes in
  0011.

## Where the backlog items go (decided, the human, 2026-10-24)

| Backlog item | Goes to |
|---|---|
| Set or change a slice's title | 0002 |
| Who caused an event (`by`) | 0004 |
| You stop a run, and continue it | 0005 |
| You retry a failed or stopped step | 0007 |
| Clean up a slice half made by a crash | 0007 |
| Cancel a slice | 0010 |
| User hooks before and after events | after 0012 |
| Errors as JSON, with codes | after 0012: output other programs can rely on |
| A local server and browser app; stop, continue, and retry buttons | after 0012: the web dashboard |
| Teammates' slices in `slice list` | after the first release, with real use |
| A `status.md` written from the log | decided after the dashboard |

## Time (the lead's estimate, 2026-10-23)

The human's hours: talking through the spec and plan, answering
questions, the preview, and fixing what goes wrong. The agents write the
code. At 35 hours a week:

| # | Slice | Hours |
|---|---|---|
| - | before 0001: the skills and `AGENTS.md` in place | 3-6 |
| 0001 | slice and events | 12-20 |
| 0002 | spec | 12-20 |
| 0003 | plan | 15-25 |
| 0004 | agents | 25-40 |
| 0005 | the run | 25-40 |
| 0006 | parallel tasks | 20-30 |
| 0007 | reliability | 25-40 |
| 0008 | cases and the lead | 15-25 |
| 0009 | finish | 20-30 |
| 0010 | stacking and cancel | 15-25 |
| 0011 | Claude Code | 15-25 |
| 0012 | configurable steps | 15-25 |

- **First release (up to 0009):** about 170-275 hours, so 5-8 weeks.
- **Up to 0012:** about 215-350 hours, so 6-10 weeks.
- The biggest unknowns are 0004 (running agents headless, the sandbox on
  Windows) and 0007 (processes and restarts on three systems). If the
  0004 experiment goes badly, add a week.

## After 0012 (the lead's proposal)

In this order:

1. **Use it for real, and make it easy to try.** Run real slices with
   `sliceflow` on this repo and one other project, and fix what hurts. Then a
   README, a short demo, and a write-up with real numbers: time, tokens,
   how often the lead was needed.
2. **Spec Kit import** (decided, the human, 2026-10-23): turn a Spec Kit
   feature into a slice once; the human fills each task's `touches`.
3. **Read Spec Kit files directly** (decided, the human, 2026-10-23).
4. **Configurable templates** (decided, the human, 2026-10-23): a project
   changes the template text and the reading rules (`sliceflow-design.md`
   1.13c).
5. **Output other programs can rely on:** the event log and
   `sliceflow status --json`, documented and versioned.
6. **From the backlog, when needed:** pull requests instead of a local
   merge, a cost limit, the web dashboard.
