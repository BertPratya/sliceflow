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
| 0002 | spec | the state machine's engine, and the spec's moves: write a spec and check it; see where a slice is and what comes next; set a slice's title; the spec's draft state comes from `slice.created`, with no extra event | 1.4, 2.2 |
| 0003 | plan | write the plan, tasks, and preview, review them, and approve the planning once; the slice is ready to build; `spec validate` reports a requirement or scenario number reused for something new after approval, by comparing with the approved version | 2.3 |
| 0004 | agents | `sliceflow` starts agents itself: the Codex adapter, the permission layers, settings per role; first use: the spec check and the plan review; who caused each event (`by`); each session sliceflow starts is bound to its slice by its session ID, and commands on other slices are refused | 1.9, 1.12c |
| 0005 | the run | `sliceflow run` builds an approved plan by itself, one task at a time: worktree, worker, checks, task review, fix rounds, merge; `sliceflow stop`, and `sliceflow run` continues where it stopped | 2.4 |
| 0006 | parallel tasks | independent tasks run at the same time: worker limit, `touches` overlap, merge queue, conflicts | 2.4 |
| 0007 | reliability | a run survives a killed `sliceflow`, a lost network, a stuck agent: heartbeats, time limits, retries, checking the log against reality on restart; `sliceflow retry` for a failed or stopped step; cleaning up a slice half made by a crash; taking over a lock left by a crashed `sliceflow`, after checking that its process is gone | 1.7, 2.4 |
| 0008 | cases and the lead | each stop goes to the right level: the chat lead, the fallback lead, or the human | 1.6, 1.12c |
| 0009 | finish | the full path to `main`: integration review, preview, result approval, update from `main`, merge, backlog | 2.5 |
| 0010 | stacking and cancel | design a slice ahead, on top of another; cancel and reopen | 2.6, 2.7 |
| 0011 | Claude Code | a second agent tool; proves the adapter shape | 1.9 |
| 0012 | configurable steps | pick and order the built-in steps | 1.10 |
| 0013 | local server and dashboard | `sliceflow serve`: see every slice live in the browser, with its sessions grouped under it, on 127.0.0.1 with a token | 1.7 |
| 0014 | web actions | stop, continue, and retry from the browser; start agents | 1.7, 2.4 |
| 0015 | stable output | errors as JSON with codes; the event log and `--json` documented and versioned; a parsed document as JSON (e.g. `sliceflow spec show --json`) | 1.13b |
| 0016 | user hooks | your own hooks before and after events, with a trust check before a repo's hooks run | - |
| 0017 | team view | teammates' slices from remote branches, and whose slice it is | - |
| 0018 | Spec Kit import | turn a Spec Kit feature into a slice once; the human fills each task's `touches` | 1.13c |
| 0019 | Spec Kit files | read Spec Kit files directly | 1.13c |
| 0020 | configurable templates | a project changes the template text and the reading rules | 1.13c |

- **0009 is the first release.** After it, `sliceflow` runs the whole workflow,
  and this repo switches from the skills to `sliceflow` itself.
- **0002 and 0003** were first cut with separate spec and plan approvals.
  Since 2026-10-21 there is one planning phase with one approval, so 0002
  ends at a checked spec, and 0003 ends at the approval. Settled when
  0002's planning starts.
- **Codex first:** 0004 builds the Codex adapter; Claude Code comes in
  0011.
- **After 0012** (decided, the human, 2026-10-24): the slices are
  numbered so the whole order is visible, but their numbers, order, and
  contents may still change when we get there. The web app comes first
  (0013, 0014): it calls the core directly, so it does not wait for
  stable output (0015).
- **0018 to 0020** were decided by the human on 2026-10-23, then as
  "after 0012" items.

## Where the backlog items go (decided, the human, 2026-10-24)

| Backlog item | Goes to |
|---|---|
| Set or change a slice's title | 0002 |
| Who caused an event (`by`) | 0004 |
| Bind a session to a slice by its session ID; refuse commands on other slices | 0004 for the sessions sliceflow starts; 0011 for Claude Code; the lead's binding (bound or not, the human's choice) decided later, by 0008 |
| Sessions grouped by slice in the browser | 0013 |
| You stop a run, and continue it | 0005 |
| You retry a failed or stopped step | 0007 |
| Clean up a slice half made by a crash | 0007 |
| Remove a lock left by a crash | 0007 |
| Cancel a slice | 0010 |
| User hooks before and after events | 0016 |
| Errors as JSON, with codes | 0015 |
| A local server and browser app; stop, continue, and retry buttons | 0013, 0014 |
| Teammates' slices in `slice list` | 0017 |
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

## Between and after the slices

- **Use it for real, and make it easy to try.** After the first release,
  and again after 0012, run real slices with `sliceflow` on this repo
  and one other project, and fix what hurts. Then a README, a short
  demo, and a write-up with real numbers: time, tokens, how often the
  lead was needed. This is an activity, not a slice.
- **From the backlog, when needed:** pull requests instead of a local
  merge, a cost limit.
