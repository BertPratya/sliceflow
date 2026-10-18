# flow design (brainstorm)

> **Status: brainstorm, not approved.** This is where we design the
> orchestration workflow that the `flow` command runs, one process at a time.
> When it is settled, it is cut into slices, and each slice gets its spec.
>
> It replaces `docs/pipeline-sketch.md` as the main design document.
> The sketch stays as input: we take from it what still fits.
>
> Started: 2026-10-18

---

## 0. How we work on this document

We go through the workflow one process at a time. For each process:

1. **The human talks.** What they need from that part, in their own words.
2. **The lead consults.** How it works today (the sketch, the workflow
   files), 2-3 real options with trade-offs, what other tools do, risks.
3. **We talk** until it feels right.
4. **The lead writes a short section** here: what `flow` does, its commands,
   its events, what it shows. The human reads it.
5. **Questions come only now**, for what is still unclear. One at a time,
   each explained.

When every process is designed, we cut the workflow into slices.

---

## 1. The big picture

### 1.1 The vision (from the human, 2026-10-18)

flow-cli is a tool for development automation.

The workflow has **4 phases: spec, plan, implement, preview.**

Today the **lead** (the main chat agent) orchestrates the pipeline: it reads
the status, decides the next step, starts workers and reviewers, runs the
checks, updates the status, and commits.

**The workflow runs from creating the slice branch to merging it into
`main`** (the human, 2026-10-18). What comes after, such as pushing to a
remote, is outside it.

The new idea: **the pipeline orchestrates itself.** `flow` holds the
control flow. The lead is a **fallback**, called only when `flow` cannot go
on by itself. The goal is a workflow that is efficient and automatic.

### 1.2 Lead-orchestrated or hybrid? (research, 2026-10-18)

Two ways to run the same workflow:

- **A. Lead-orchestrated (today).** An AI agent reads the rules and decides
  each step.
- **B. Hybrid.** `flow` runs the steps as fixed code. An AI agent is called
  only when `flow` stops.

| | A. Lead orchestrates | B. Hybrid: `flow` + lead as fallback |
|---|---|---|
| Same input, same steps? | No. A model can skip a step or bend a rule, even at temperature 0. | Yes. The rules are code, and tests prove them. |
| Cost | Tokens on every routine step (start a worker, run checks, update status). | Tokens only on real work (spec, code, reviews) and on stops. |
| Long runs | The lead's context fills. Today the lead must suggest a fresh session after about 5 tasks. | `flow` has no context to fill. Each agent starts fresh. |
| State | Kept by hand in `status.md`. It can drift from what really happened. | Written by `flow` at each step, from what really happened. |
| Unusual cases | Handled well: the lead can think. | `flow` stops and calls the lead. Only as good as the stop rules. |
| Work to build | Little: Markdown rules. | More: every rule must be designed and coded. |

**What the sources say:**

- Anthropic separates *workflows* (fixed code paths) from *agents* (the model
  chooses its own steps). Workflows fit well-defined tasks and give
  predictability. Agents fit open-ended tasks. The advice is to choose the
  simplest design that works.
- A study of over 1,600 multi-agent runs (MAST) found that most failures come
  from system design and coordination between agents, not from what the
  model can do. Coordination is exactly what a lead does in design A.
- Practitioners make the same point: an AI orchestrator puts something that
  is not deterministic on the control path of a system whose value is
  reliability. The common answer is a state machine in code, with AI used for
  bounded steps.
- A warning for B: when the fallback is the same kind of model, it can repeat
  the same mistake. The fallback needs a narrow job and clear limits.

**Lead's recommendation: B, hybrid.** Our workflow is well-defined. Its rules
are already written (`workflow/formats/lifecycle.md`, the stop rules in
`AGENTS.md`). The parts that need judgment stay with agents: writing a spec,
writing code, reviewing. The parts that follow rules move into `flow`.

### 1.3 Three kinds of work

This follows from 1.2:

| Kind | Who | Examples |
|---|---|---|
| Rules | `flow` | pick the next task, start a worker, run the checks, count failures, merge, update status |
| Judgment | agents (lead, workers, reviewers) | write the spec, write code, review, answer a blocked task when the answer is in an approved document |
| Decisions | the human | approve, choose between options, accept a risk |

The fallback rule: `flow` calls the lead only when a step needs judgment and
no worker or reviewer owns it. The lead acts through `flow` commands, so the
state stays correct.

### 1.4 Phases and lifecycle (from the human, 2026-10-18)

Two different things, not to be mixed:

- **Phases** are the big picture: what kind of work is going on.
  There are 4: **spec, plan, implement, preview.**
- **Lifecycle states** say where a slice is. They are not phases.
  - **ready** means "ready to implement": the plan is approved, nothing has
    started yet.
  - **done** means the slice is finished and merged.
  - Others (for example blocked, cancelled) are states too.
- **Integration** is a step inside the implement phase, not a phase of its own.

The workflow files today list 7 phases (spec, plan, ready, implementing,
integration, preview, done). They mix phases and states. The new design
keeps them apart.

**The slice lifecycle. Decided (the human, 2026-10-18):**

```
spec → plan → implement → preview → done
```

- **No "running" state.** Whether a run is going is shown by the status of
  each task in the implement phase.
- **No "ready" state.** "Ready" means the spec and the plan are both
  approved. It is worked out from that, not stored.
- **A blocked slice remembers where it was,** so it can go back there. It
  must be recoverable (1.7).

Lead's notes, to talk about (not decided):

- So **blocked is a mark on top of the phase**, not a phase of its own: for
  example "implement (blocked: plan review failed 3 rounds)". When it is
  solved, the mark goes away, and the phase is still there.
- Worked-out states ("ready", "running") are computed from the event log
  when shown. This keeps one source of truth (1.13).
- Cancelled: not talked about yet.

### 1.5 One autonomous run (from the human, 2026-10-18)

After the human approves the plan, `flow` runs by itself. How far it goes is
**configurable**, with two choices:

- **To preview:** build every task, review it, merge it, run the
  integration review, then wait for the human at preview.
- **Preview prepared:** the same, and `flow` also prepares the preview for
  the human (for example a test repo and the list of things to try).

There is no "build only" choice. The run is autonomous, so it always goes at
least as far as preview.

### 1.6 The escalation ladder (from the human, 2026-10-18)

When something goes wrong during a run, it climbs 3 levels:

```
flow  ──can't──▶  lead  ──can't──▶  human
```

1. **flow:** if `flow` can handle it, it does.
2. **lead:** if not, `flow` sends it to the lead. If the lead can handle
   it, it does.
3. **human:** if the lead can't, or it needs a person, it goes to the human.

When each level hands it on, and how, is not defined yet.

**Lead's proposal, to talk about** (not decided: the limits, the case, and
the table below are all open):

Each level has a clear job and a limit:

| Level | Handles | Limit before it climbs |
|---|---|---|
| flow | anything with a fixed rule: retry, a new worker for a fix round, merge again after a conflict | the counters (for example 3 failed checks, 3 review rounds, 2 failed merges) |
| lead | a question whose answer is in an approved document; a failure whose cause it can see | one try per stop (to be decided) |
| human | anything that changes an approved document, accepts a risk, or needs something new (a dependency, a secret, a CI change) | none: the human decides |

Rules for every climb:

- It carries a **case**: what stopped, what was tried, and the options. The
  next level never starts from nothing.
- No work is lost. The task's work stays committed on its branch.
- While one task waits, the run goes on with tasks that do not depend on it.
- The lead acts only through `flow` commands, so the state stays correct.

The four examples, under this proposal:

| What happens | Who handles it |
|---|---|
| A task's checks fail 3 times | flow stops the task → lead looks at the log: cause visible → hint and retry; else → human |
| A worker asks "the spec doesn't say what happens on an empty list" | lead: answer in the spec or plan → it answers; else → human (it is a spec decision) |
| A reviewer finds a serious problem | flow: a new worker fixes it (fix round). After 3 rounds without a pass → lead → human |
| Two tasks conflict at merge | flow: a new worker merges again. After 2 failed merges → lead (usually wrong `touches` or `depends_on`) |

### 1.7 Monitoring, reliability, parallel slices (from the human, 2026-10-18)

**Monitoring.** A dashboard, read-only first. It shows the progress of each
slice: what is happening now, which step, failures, what is blocked, and why.

**Agents.** Maybe also every AI agent that is running: who is working now,
on what. (The human is not sure yet if this is needed.)

**Reliability.** `flow` must handle stopping unexpectedly: the program is
killed, the computer restarts, or an agent stops (for example the internet
connection drops while a worker is working). The human must see that it
stopped, and why. If possible, the system recovers by itself.

**Parallel slices.** Slices that can run at the same time should run at the
same time.

**Lead's proposal, to talk about** (not decided):

One design choice carries all four: **`flow` writes down every step before
and after it does it** (an event log). The dashboard reads that log. Recovery
reads that log. Nothing else is needed for either.

1. **Event log.** `task.started` is written before the worker starts;
   `task.handed_in` after. On restart, anything started but not finished was
   interrupted. (Durable execution systems like Temporal work this way.)
2. **Check reality on restart.** `flow` compares the log with what is
   really there: branches, worktrees, commits, running processes. It repairs
   the difference: resume, restart the task (its work is kept on the branch),
   or report it. (Kubernetes works this way: "what should be" against "what is".)
3. **Heartbeat.** While an agent runs, `flow` records that it is alive and
   when it last did something. Dead or silent too long → `flow` knows, and
   says so.
4. **Two kinds of failure.**
   - *Passing* (network down, rate limit, the agent tool crashed): `flow`
     waits and retries by itself. It does not count as the task failing.
   - *Real* (checks fail, review fails): the counters and the ladder (1.6).
5. **Two kinds of state.**
   - *History* (what happened to the slice): kept with the slice, in git.
   - *Live* (which process runs, last heartbeat): local only, not in git.
6. **Agents view.** `flow` starts every agent itself, so it already knows
   who runs where. Recording it costs little: role, slice, task, tool and
   model, started, last activity.
7. **Parallel slices.** Each slice has its own run and its own lock.
   Merges into `main` still happen one at a time.
8. **Dashboard.** First `flow status` in the terminal (with a live mode).
   A web view later, reading the same data.

What is easy, and what is hard:

| Part | Effort | Why |
|---|---|---|
| Event log, `flow status` | easy | append a line, read the lines |
| Agents view | easy | `flow` starts every agent anyway |
| Detect a stop, show the reason | medium | the reason depends on what the agent tool reports (exit code, last output) |
| Recover after `flow` itself is killed | medium | every step must be safe to run twice; check reality on restart |
| Silent agent (alive but stuck) | medium | needs a time limit and an activity check |
| Parallel slices | medium | locks, and merging into `main` one at a time |
| Processes on Windows, macOS, Linux | medium | starting, watching, and stopping processes differs per system |
| Recover every case by itself | hard | start with: detect, report, and one command to resume; add automatic recovery case by case |

### 1.8 Who is the lead (from the human, 2026-10-18)

The human works in two ways. The brainstorm, spec, plan, and tasks are
written in one chat session. Then the implement phase runs either:

- in **the same session**, or
- in **a new session**, which gets the current status and context first.

This does not need its own design. `flow` keeps the state in files, so any
session can be the lead. We design `flow`, not the session.

### 1.9 How `flow` starts and ends agents (from the human, 2026-10-18)

To consider: how `flow` starts the workers and reviewers, and cleans up
after them.

Today the lead starts them as subagents inside its own chat. In the hybrid,
`flow` starts each one as a separate program, using the agent tool's
non-interactive mode (for example `claude -p`, `codex exec`).

**Lead's notes, to talk about** (not decided):

- **Permissions.** Nobody is there to click "allow". What each agent may do
  must be set before it starts (sketch H4).
- **Result and reason.** The tool's exit code and last output tell `flow`
  whether the agent finished, failed, or hit a limit. JSON output also gives
  cost and turns.
- **Per role.** Tool, model, and effort per role come from the settings.
- **Ending.** Stop the whole process (with its children) on every system,
  save the work on the task branch, remove the worktree when the task is done.

### 1.10 A configurable workflow (from the human, 2026-10-18)

The workflow should be configurable. For example, a project that does not
need reviews turns them off, or changes the order of steps.

**Lead's proposal, to talk about** (not decided). Three levels of "configurable":

| Level | Example | Effort |
|---|---|---|
| 1. Settings | limits, run reach, review on or off | easy |
| 2. Steps from a fixed set | per phase, pick and order built-in steps, each with a condition | medium |
| 3. Any workflow | your own steps, states, and scripts | hard: this is building a general workflow engine |

Proposal: levels 1 and 2. Level 3 breaks the reliability promise (1.7):
`flow` can only recover a step safely if it knows what the step does.

Later, level 3 stays possible (lead's note): build the steps inside `flow` as
separate units, each with the same shape (what it reads, what it writes, its
events, safe to run twice). Then a general engine is a later step: open that
shape to steps a project writes itself. Specific first, general later.

A level 2 example:

```yaml
implement:
  steps: [build, check, review, merge]
  review: core          # always | core | never
preview:
  prepare: true
```

### 1.11 Build or reuse? (research, 2026-10-18)

No existing tool matches this workflow. The closest ones, by kind:

| Kind | Example | What it has | What it lacks |
|---|---|---|---|
| Spec-driven kits | GitHub Spec Kit | the same phases (specify, plan, tasks, implement), Markdown per feature | no orchestration: the chat agent runs each step, like today |
| Agent orchestrators | ORCH, Orca, and similar | parallel agents in git worktrees, a task state machine, review | no spec or plan approvals, no escalation ladder |
| Workflow engines | Temporal, GitHub Actions | durable steps, retries, recovery | a server or the cloud; nothing about agents, specs, or approvals |

**Lead's recommendation:** build `flow`, but reuse what exists:
- the agent tools themselves (we never build an agent),
- git for branches, worktrees, and merges,
- the ideas: durable steps (Temporal), task state machines (ORCH),
  spec files (Spec Kit).

Before the plan, look closer at ORCH-like tools, to borrow ideas.

**Decided (the human, 2026-10-18):** build `flow` first, as a specific engine
for this workflow. A general engine is a later, optional project.

### 1.12 Who runs `flow` (from the human, 2026-10-18)

The **lead** runs the `flow` commands, not the human. The human works with
the lead in the chat. `flow` is designed for the lead to use.

Lead's note: so every `flow` output is read first by an AI agent. It must
be exact and complete (what happened, what is next, why it stopped), and
still plain enough for the human to read when they look.

### 1.12b No data out (decided, the human, 2026-10-18)

The old rule "`flow` makes no network calls" is removed. The new rule:
**`flow` itself sends no data to any outside service** (no telemetry, no
uploads, no calls home). It runs local programs: git, the check commands,
the agent tools. What those programs send is theirs, not `flow`'s. A
dashboard page served on the human's own computer is fine.

### 1.13 Where and how events are kept (decided, the human, 2026-10-18)

**The source of truth.** Each slice has one event log:
`slices/NNNN-<name>/.events.jsonl`, one event per line. It is the only
source of truth for the slice.

- Only `flow` writes it. Not the human, not the lead, not any agent.
- It is committed in git with the rest of the slice.
- It holds only the steps that matter: for example `slice.created`,
  `plan.approved`. No high-frequency signals such as heartbeats.

**`status.md` is a view, not a second truth.** `flow` writes it from
`.events.jsonl` so people can read the status.

- `flow` never reads `status.md` to decide anything.
- If it is wrong or deleted, `flow` writes it again from the log.
- Its first line says it is written by `flow` and must not be edited.
- `flow` rewrites it only when the status really changes.

**Live state is not history.** Heartbeats, retries, and which process runs
where stay local in `.flow/`, in a small file `flow` overwrites. It is
not in git and is not a source of truth for the slice.

**Lead's notes** (technical, for the plan):

- Several `flow` processes may write at once (two workers hand in at the
  same time): `flow` locks the file while it adds a line.
- Task worktrees have their own copy of the slice folder. `flow` always
  writes the copy in the slice worktree. A task never changes it.
- Crash safety: "starting X" before X, "X done" after. A half-written last
  line is skipped when reading, and reported.
- If a log ever grows large, keep the computed status in `.flow/` and apply
  only the new events. Not needed at the start.

### 1.14 Not clear yet

To come back to when the processes are designed:

- The full list of lifecycle states, and what moves a slice between them.
- When and how each level of the ladder hands on (1.6).
- What the lead receives when it is called, and what it may do. Also what
  a fallback lead (started by `flow`, 2.4) may do alone.
- Which run choice is the default (1.5).
- Is the agents view needed (1.7)?
- How agents get permissions when nobody is there to click "allow" (1.9).
- How configurable the workflow is: levels 1-2 or more (1.10).
- Exit codes: the sketch says two different things (step 18 and step 27).
- The full list of events and what each one holds; what `status.md` and the
  dashboard show (after the overview pass).

---

## Sources

- Anthropic, Building effective agents: https://www.anthropic.com/research/building-effective-agents
- Cemri et al., Why Do Multi-Agent LLM Systems Fail? (MAST): https://arxiv.org/abs/2503.13657
- Compiled AI: Deterministic Code Generation for LLM-Based Workflow Automation: https://arxiv.org/pdf/2604.05150
- Bernstein docs, Deterministic vs. LLM-Based Orchestration: https://bernstein.readthedocs.io/en/latest/architecture/orchestration-approaches/
- Agentic fail-safes need deterministic modes: https://dev.to/omnithium/the-t-mobile-outage-lesson-why-agentic-fail-safes-need-sos-mode-determinism-1chm

---

## 2. The processes

One section per process, in order. Each says what `flow` does, its commands,
its events, and what it shows.

### 2.1 Create a slice

**From the human (2026-10-18):**

1. The lead runs `flow slice new` with the slice name. `flow` gives it the
   next number: `NNNN-<name>`.
2. `flow` creates the slice's branch and its worktree in `.flow/worktrees/`.
3. `flow` copies the spec template into the slice.
4. `flow` creates the files that record the slice's status.

**Lead's view, to talk about** (not decided):

- **Command form. Decided (the human, 2026-10-18):** `flow <noun> <verb>`,
  for every command: `flow slice new | list | cancel | done`, `flow task …`,
  `flow review …`. Like `gh pr create` and `git worktree add`.
- **`.flow/` for all local state.** Worktrees, locks, logs, and live agent
  state all go under `.flow/`: one folder, one line in `.gitignore`.
  The workflow files today say `.worktrees/`; they change with this design.
- **The status files. Decided (the human, 2026-10-18):** an event log
  is the truth (where and how it is kept: open, see 1.13). After every step, `flow` writes a
  `status.md` from it, readable on GitHub and in a pull request, never edited
  by hand. (Rejected: the event log only, which nobody can read without
  `flow`; a hand-kept `status.md`, which can drift.)
- **Checks before it creates anything.** A technical detail: the lead
  proposes the list in this slice's spec (the sketch has a first list).
  If one fails, nothing is created.
- **All or nothing.** If a step fails halfway, `flow` undoes what it did and
  says why.
- **Output.** The slice ID, branch, worktree, the files it created, and the
  next step.
- **Commit.** On the slice branch: `NNNN: new slice`.
- **Event:** `slice.created` { slice, title, after, by, at }.
- **The same number on two machines** (sketch H11): `flow` sees only the
  branches it knows. Two people offline can get the same number. Open.

### 2.2 Write and approve the spec

**From the human (2026-10-18):**

1. The human and the lead write the spec together. Its status is **draft**
   until it is finished.
2. `flow spec validate` checks the format, like a compiler: the headings
   are there, the format follows the template. If something is wrong, it
   says what. (The exact format rules are refined later, in a slice.)
3. When the draft is finished, the lead runs `flow spec submit`
   (decided, the human, 2026-10-18).
4. Then, depending on the settings:
   - **With a review:** `flow` starts a reviewer. It looks for unclear
     points and conflicts. If it finds any, they are written down, and the
     lead is told.
   - **Without a review:** the lead reports to the human.
5. No open points → status **awaiting approval**. It waits for the human.
6. The human may read the spec or not. They tell the lead "approved".
7. The lead runs the approve command. The slice enters the **plan** phase.
8. Every approval keeps a copy: `.history/spec/`. History is read-only.

**Lead's view, to talk about** (not decided):

- **Spec statuses.** draft → in review → awaiting approval → approved.
  Without a review: draft → awaiting approval. The reviewer finds open
  points: back to draft. The status lives in the event log, not in the
  spec's header.
- **"The draft is finished": `flow spec submit`** (name decided). It runs
  `flow spec validate` first and refuses if the format is broken, so the
  check can never be skipped. Then it starts the reviewer, or goes straight
  to awaiting approval.
- **Validate like a compiler.** Each problem on one line:
  `spec.md:42: error: R7 has no Check: line`. Errors (broken) block
  submit. Warnings (not finished yet, for example open questions) are fine
  while drafting, but block submit too.
- **Where the reviewer writes.** A review file in the slice's `reviews/`
  folder. `flow` reads it, records `review.submitted`, and tells the lead:
  "3 open points". The lead answers facts itself and brings decisions to
  the human, updates the spec, and submits again.
- **Approve.** `flow spec approve`, run by the lead only right after the
  human says so. The event records who approved and the human's words.
  `flow spec reject` sends it back to draft, with the reason.
- **Read-only history.** File permissions do not travel in git and work
  differently on Windows, so "read-only" must be a check: the approval
  event stores a fingerprint (hash) of the copy, and `flow` reports any
  history file that no longer matches.
- **A pattern:** files that only `flow` writes start with a dot
  (`.events.jsonl`, `.history/`). Files people and agents write do not
  (`spec.md`, `reviews/`).
- **Changed after approval.** `flow` sees `spec.md` differ from the last
  copy in `.history/`: status **changed**. It is submitted and approved again.
- **Setting:** `spec.review: on | off`.
- **Events:** `spec.submitted`, `review.started`, `review.submitted`,
  `spec.approved`, `spec.rejected`.

### 2.3 Write and approve the plan

**From the human (2026-10-18):**

- It works like the spec (2.2): draft, validate, submit, review (if on),
  awaiting approval, approve, a copy in `.history/`.
- The difference: more files, most of all the task files.
- Open: how task files are made. Today the lead writes them by hand. Idea:
  the plan is settled first, then a command creates the task files, the lead
  fills them, submits, and the reviewer reviews.

**Research: three ways to make task files (2026-10-18)**

| | A. The agent writes each file from nothing | B. `flow` copies a blank template, the agent fills it | C. The agent names the task, `flow` fills the fixed parts, the agent writes only the rest |
|---|---|---|---|
| Agent's work per task | 1 write, but it types everything: header, headings, copied text | read the template, then several edits | 1 command, then 1 write of the free text |
| Wasted output | headers, headings, and requirement text copied from the spec | little | none |
| Format errors | likely: wrong number, wrong header, a requirement copied with a word changed | fewer | almost none: `flow` writes the fixed parts |
| Checks needed | many (T1-T6 today) | some | few: `flow` refuses a bad task when it is created |

What the sources say:

- Output (what the agent types) is the slow and costly part. Typing text
  that is always the same is waste.
- Models are bad at copying text exactly, even when it is right in front of
  them. Over a long text, a small error rate per word adds up. The advice:
  let a tool do the copying.
- Editing a file in many small places is also error-prone, and every edit is
  one more round trip. So B is often not faster than A.

**Lead's recommendation: C, and do not copy the requirements at all.**

- **Do not copy what already exists.** Today each task file copies its
  requirements word for word from the spec, and a check (T5) makes sure the
  copy is exact. Instead, the task file only names them (`covers: R3, R5`).
  When a worker starts, `flow brief` shows the text from the approved spec.
  One source of truth, no copy that can drift, and check T5 goes away.
- **The lead writes only what needs judgment:** what to build, and "Done
  when" (the tests that prove it).

**Agreed (the human, 2026-10-18):** option C, and requirements are named,
not copied.

**All tasks at once, from the plan (from the human, 2026-10-18):**

The lead writes the plan with a **task list**: every task, what it depends
on, and which requirements it covers. Every requirement must appear. Then
one `flow` command creates **all** the task files from that list.

Research (2026-10-18):

- **Plan everything first, then build, is as good as step by step.** A 2026
  study of tool-using agents found no accuracy gain from planning step by
  step instead of planning the whole thing first. Another found step by step
  only 3 points better, at almost 6 times the cost. A full plan first is
  also what the human approves, so it must exist anyway.
- **Other tools do the same.** Task Master keeps one list of all tasks with
  their dependencies (`tasks.json`), and one command generates the task
  files from it. GitHub Spec Kit writes all tasks in one list too.
- **Seeing the whole list at once allows checks no single file can do:**
  every requirement is covered, no dependency loops, no two tasks that run
  at the same time touch the same files. `flow` checks them before any file
  exists.

**Lead's view, to talk about** (not decided):

- **One truth for the structure. Decided (the human, 2026-10-18).** If the
  list is in `plan.md` and also in each task file's header, there are two
  copies that can drift. So the structure (number, name, covers, depends on, touches, risk) lives
  only in the plan's task list. A task file holds only its content (what
  to build, "Done when"). `flow brief` puts the two together for the worker.
- **A form `flow` can read exactly.** A Markdown table is easy to break by
  hand. A small fenced block in `plan.md` (for example YAML) is exact and
  still readable.
- **Safe to run again.** Plans change. The command can run any time: it
  creates the files that are missing, never overwrites content the lead
  wrote, and reports a task file whose task left the list (it does not
  delete it). A name for that: `flow task sync`.

**The whole plan process, step by step (lead's proposal, 2026-10-18):**

| # | Who | Command | What happens | Plan status |
|---|---|---|---|---|
| 1 | lead | (writes) | `plan.md` with its task list, and the contract changes | draft |
| 2 | lead | `flow plan validate` | checks the plan and its task list; repeat until clean | draft |
| 3 | lead | `flow task sync` | creates the task files from the task list | draft |
| 4 | lead | (writes) | fills each task file: what to build, "Done when" | draft |
| 5 | lead | `flow plan validate` | checks plan, task list, task files, and contract changes together | draft |
| 6 | lead | `flow plan submit` | validates again (refuses if broken), then starts the reviewer, or skips it when review is off | in review |
| 7 | reviewer | (writes, then hands in) | one review file with findings: high, medium, low, and questions for the human | in review |
| 8 | `flow` | — | reads the review: no high or medium findings, no questions → pass; else → back to the lead | awaiting approval, or draft |
| 9 | lead | (fixes) + marks each finding | fixed, wontfix, or deferred, each with a note; decisions go to the human | draft |
| 10 | lead | `flow plan submit` | next round: the reviewer checks the fixes (up to the round limit, then the human decides) | in review |
| 11 | human | says "approved" | | awaiting approval |
| 12 | lead | `flow plan approve` | copies the plan, task files, and contract changes to `.history/plan/`; the slice is **ready** | approved |

- **One submit for the plan and its tasks:** `flow plan submit`, not a
  separate `flow task submit`. They are reviewed and approved together.
- **`flow task submit` is kept for the implement phase:** a worker hands
  in its finished task with it.

**Different from the spec:** the reviewer gives findings to fix (high,
medium, low), not only questions. The lead fixes them and submits again, up
to the round limit.

**Approving the plan, and what goes into history (research, 2026-10-18):**

The command is `flow plan approve`: the same thing that was submitted is
approved. There is no `flow task approve`: tasks are approved with the plan.

Should the task files be copied into history too? What other tools do:

- **GitHub** (branch protection, "dismiss stale approvals"): an approval
  belongs to the exact content that was approved. When new code is pushed,
  the approval is dropped, and someone must approve again.
- **Terraform** (saved plans): the reviewed plan is saved as a file, and
  `apply` carries out exactly that file. A guide warns that if the source
  changes after the plan was saved, nobody may notice, so it should be
  checked.

Both say the same thing: an approval must be tied to the exact content, all
of it. The tasks are part of what the human approved. If a task file could
change after approval without anyone noticing, the approval would mean
nothing for that task.

**Decided (the human, 2026-10-18):** `flow plan approve` saves the plan, the
task files, and the contract changes together, one folder per revision:

  ```
  .history/plan/R001/
    plan.md
    tasks/01-event-log.md
    tasks/02-slice-new.md
    contracts.patch        (the slice's contract changes)
  ```

Details (lead's notes):

- The `plan.approved` event stores a fingerprint (hash) of each file, so a
  changed copy, or a changed task after approval, is found at once.
- Why copy, when git already has the history? A copy can be read without
  git commands, shows on GitHub, and survives a squash merge (which removes
  the commit where the approval happened).
- Copies are small Markdown files; the cost is low.
- A fix task added after approval (from the integration review or the
  preview) is recorded as an event, not as a new plan revision. It needs no
  new approval when it only covers approved requirements (the rule in the
  workflow files today).

### 2.4 Implement: the autonomous run

**From the human (2026-10-18):**

1. **Pick a task** whose dependencies are all done.
2. **Start a worker** with its prompt. Not through a desktop app: through a
   command-line agent, for example the Claude Code CLI or the Codex app
   server. Building on those should make tracking and recovery easier.
3. **Each worker has its own worktree.** Example: `A → B → D`, `C → D`.
   A and C have no dependencies, so they start together (up to the worker
   limit in `flow.yaml`). A starts from the slice branch. When A is done, it
   is merged into the slice branch. B then starts from the slice branch with
   A in it. C may start from the slice branch before or after A or B is
   merged; both are fine.
4. **Only one task merges into the slice branch at a time** (no race).
5. **Failed review:** back to a worker. The number of review rounds has a
   limit. Over it: `flow` reports to the lead, and the lead tells the human
   what they need to decide.
6. **Task review is configurable** in `flow.yaml`. A simple project can
   turn it off, to save tokens.
7. **Integration:** when all tasks are done, a final review of the whole
   slice. When it passes, the slice enters the preview phase.

**Lead's view, to talk about** (not decided):

**Starting agents (2).** Yes, a command-line agent is the right base.

| | Claude Code `claude -p` | Codex `codex exec --json` | Codex app server |
|---|---|---|---|
| How it runs | one program run per job, then it exits | one program run per job, then it exits | a program that stays running; `flow` talks to it (JSON-RPC) |
| What `flow` sees | a stream of events (`--output-format stream-json`): start, each message and tool use, a final result with cost | a stream of events: thread, turn, item (command, file change), failed | the same events, plus approval requests, and `flow` can interrupt |
| Recovery | a session ID; `--resume <id>` continues it | the thread can be resumed | `thread/resume` |

Both tools give a stream of events and a session to resume. That is what
tracking and recovery need. Proposal: one small **adapter** per tool in
`flow`, all with the same shape (start, events, result, resume, stop).
Start with the one-run-per-job mode of each tool (simplest); the Codex app
server can come later through the same adapter, if its extra control
(interrupt, approvals) is needed.

**Worktrees (3).** Yes, one per running task. Two rules make the example
work:

- **Start from the newest slice branch.** When a task starts, its branch is
  made from the slice branch as it is at that moment. B starts after A is
  merged, so it has A.
- **Bring it up to date at merge, and check again.** C may have started
  before A was merged. At merge, `flow` puts C's work on top of the newest
  slice branch (with A, maybe B), runs the checks again, and only then
  merges. A conflict, or failing checks, goes back to a worker.
- Also: two tasks that change the same files (`touches`) do not run at the
  same time. Fewer conflicts.

**One merge at a time (4).** Yes: a merge queue with a lock. This is the
"not rocket science rule" (bors, GitHub merge queue): only merge what passed
the checks on the exact result of the merge.

**Failed review (5).** Proposal: a **fresh worker** for each fix round (new
context, as the sketch settled), with the findings and what the last worker
did. The round limit hits: the lead does not try to fix it; it prepares a
case for the human: what was found, what was tried, the options.

**Configurable reviews (6):**

```yaml
implement:
  task_review: never      # always | core | never
  integration_review: on
```

**Decided (the human, 2026-10-18):** the final review of the whole slice
runs even when task reviews are off. On by default, with its own setting.

**Integration (7).** From the workflow files today, before the final
review: update the slice from `main`, then run the slice checks. Then the
final review. Fix tasks from it run like any task.

**A merge conflict (from the human, 2026-10-18).** Example: `A → B → D`,
`C → D`. C merges first. Then B merges, and there is a conflict. Who solves
it: the lead, or B's worker?

**Lead's view, to talk about** (not decided): **B's worker solves it.**

- The task being merged owns the conflict. B is the one that does not fit
  the newest slice branch yet. Merge queues work the same way: a change
  that no longer fits goes back to its author, not to the people who merged
  before it.
- Not the lead: the lead does not write code, and it does not know B's work.
  It stays the fallback.

What `flow` does:

1. It tries to put B's work on top of the newest slice branch (with A and C).
   Conflict.
2. It stops the attempt. The slice branch and B's branch stay as they were.
3. It starts a worker on B, with a conflict brief: the files in conflict,
   what C changed and why (C's task), and the rule "keep what both tasks
   meant to do". Resuming B's last session is a good choice here: that
   worker already knows B's work.
4. The worker solves it, runs the checks, and hands in. Then the merge is
   tried again.

Limits and the ladder:

- A failed merge does not count against the worker's check failures.
- After the merge limit (for example 2 failed merges in a row): the lead.
  A conflict between two tasks that ran at the same time means their
  `touches` overlapped without anyone knowing, or a dependency is missing
  (B should have waited for C). That is a plan problem, so it can go on to
  the human.

**Decided (the human, 2026-10-18):** after a conflict is solved, passing
checks are enough. The final review sees the whole slice anyway. Exception:
if solving it changed more than the fit, the task review runs again (when
task reviews are on).

How `flow` knows (lead's proposal, not decided): **it measures, it does not
ask.**

- `flow` compares B's change before and after the conflict was solved (git
  can compare two versions of the same change: `git range-diff`).
- Only the files that were in conflict changed → it was the fit. Checks are
  enough.
- Any other file changed too → more than the fit. Review again.
- The worker's hand-in note goes along as information for the reviewer, but
  it does not decide. An agent judging its own change is not reliable, and
  this is a fact code can measure (the hybrid rule, 1.3).

**Conflicts are visible, and do not block (from the human, 2026-10-18).**
A conflict is shown to the human (dashboard, `flow status`), but the run
keeps going.

Lead's notes:

- A conflict is a step that matters, so it is an event in the slice's log:
  `task.merge_conflict` { task, with, files, attempt }. Solved:
  `task.merged` as usual.
- `status.md` and the dashboard show it while it lasts, for example:
  `B: merge conflict with C (internal/slice/new.go), worker solving it, try 1 of 2`.
- The merge queue does not wait for B: the next task that is ready merges.
  Other tasks keep running.

**Watching a long run (from the human, 2026-10-18).** The implement phase
runs the longest, so it must be reliable:

- `flow` checks often that the run is still healthy.
- There is a time limit. Over it, the lead looks at what happened and
  decides what to do.

**Lead's view, to talk about** (not decided):

Who checks what:

| What can go wrong | Who notices | How |
|---|---|---|
| An agent is dead (its program ended without a result) | the `flow` run | the program's exit |
| An agent is alive but silent (no output for a long time) | the `flow` run | the time since its last event |
| An agent is busy, but takes too long | the `flow` run | the time since it started |
| The `flow` run itself died (killed, computer restarted) | the next `flow` command (`flow status`, the dashboard, `flow run`) | the run writes "alive at <time>" into `.flow/` every few seconds; too old, or its process is gone → "the run stopped at 14:02, last step: task B review". `flow run` continues from the log. |

**Decided (the human, 2026-10-18):** a time limit is **per agent run**, not
per task or per slice. Each role has its own: a worker and a reviewer have
separate limits. Each fresh worker in a fix round starts its own clock.

Lead's proposal: two kinds of limit per role (in `flow.yaml`):

```yaml
limits:
  worker:
    minutes: 60           # still working, but long
    silent_minutes: 15    # no output at all: probably stuck
  reviewer:
    minutes: 20
    silent_minutes: 10
  lead:                   # a fallback lead that flow starts itself
    minutes: 20
    silent_minutes: 10
```

- Over a limit, `flow` does not stop the agent by itself. It tells the lead,
  with a case: which task, how long, the last activity, the last lines of
  output, the cost so far.
- The lead decides: let it continue, stop it and start a fresh worker, or
  stop it and ask the human.
- An idea for later: a cost limit per task, so a stuck agent cannot burn
  tokens without end.

**When no lead session is there. Decided (the human, 2026-10-18):** `flow`
starts a fresh lead agent itself, the same way it starts a worker. It gets
the case, handles it, and ends. If it cannot handle it, the case goes on to
the human, as the ladder says (1.6): the task is blocked and waits, and the
run goes on with other tasks.

So there are two kinds of lead:

- **The chat lead:** the session the human talks to (1.8).
- **A fallback lead:** started by `flow` for one case, with its own time
  limit (`limits.lead`).

Open: what a fallback lead may do, and what it must leave to the human.

### 2.5 Preview and finish

**From the human (2026-10-18):** as the workflow files describe it today:

1. **Show the result.** What the feature does now, how to run it, a list of
   things to try by hand, and what was parked for later. With the run
   choice "preview prepared" (1.5), `flow` prepares this before the human
   looks: for example a test repo and the list of things to try.
2. **The human tries it.** A problem → fix tasks, and the run starts again
   (implement). Something new → a spec change, not a fix.
3. **The human approves the result.**
4. **Finish.** Update the slice from `main` once more (checks again; a
   conflict → a fix task, then the final review and preview again). Merge
   into `main`. Parked items go to `backlog.md`. The slice is **done**, and
   its worktree is removed.

**Lead's view, to talk about** (not decided):

- **Commands**, following the same pattern as spec and plan:
  `flow slice approve` records the human's approval of the result (only
  right after they say so, with their words). `flow slice done` does the
  finish: update from `main`, checks, merge into `main` (one slice at a
  time, like the task merge queue), backlog, remove the worktree.
- **The things to try** come from the requirements checked `manual` in the
  spec, so the list is made by `flow`, not written by hand.
- **Push. Decided (the human, 2026-10-18):** not part of the workflow.
  The workflow ends with the merge into `main`; pushing is outside it.
- **Events:** `preview.started`, `slice.approved`, `slice.done`.

### 2.6 Design ahead: stacked slices

**From the human (2026-10-18):** while 0002 is in implement, the human
wants to design the spec of 0003, or even 0004. They do not want to wait
until 0002 is finished and merged.

**Research (2026-10-18):**

- **Dual-track agile** (Jeff Patton and others): design ("discovery") runs
  ahead of building ("delivery"), at the same time, in one team. Design does
  not have to be finished for everything before building starts. The common
  trouble: design cannot keep up with building. With AI workers that build
  fast, that trouble is even bigger, so designing ahead is the right goal.
- **Stacked pull requests** (Graphite, GitHub stacks, gh-stack): code built
  on top of an unmerged branch. When the lower branch changes, every branch
  above it must be moved on top again ("restack"). Done by hand it is
  "tedious and error-prone"; tools do it automatically. Advice: keep stacks
  short. After the lower branch is merged, move only the upper branch's own
  commits onto the new base.

**Decided (the human, 2026-10-18):** points 1-4 below. Point 5 is for later,
only if waiting turns out to be too slow.

Two different things wait on 0002, and only one of them has to:

| | Design (spec, plan) | Build (implement) |
|---|---|---|
| What it changes | only files in the slice's own folder | code |
| Needs 0002's code? | no: it needs 0002's approved spec and plan | yes |
| Can run while 0002 is in implement? | yes, always | A: no, waits for the merge. B: yes, on top of 0002's branch |

1. **Design ahead, always.** 0003 and 0004 can be created, specified, and
   planned at any time. A stacked slice records what it waits for:
   `after: 0002`. Its branch starts from `main`: it only holds its own
   documents until it builds, so it never needs to be moved while designing.
2. **The plan reviewer reads 0002's approved spec and plan** (and contract
   changes) as the base 0003 builds on.
3. **If 0002's plan changes after 0003's plan was approved**, `flow` marks
   0003's plan "stale: 0002 changed", and it is checked and approved again
   (check X3 in the workflow files today).
4. **Build after the merge (A), automatically.** When 0002 is merged into
   `main`, `flow` updates 0003's branch from the local `main` (no remote is
   involved). 0003 holds only its own documents, so this cannot conflict.
   The checks run once, then `flow` starts 0003's run by itself (if its
   plan is approved). 0004 follows 0003 the same way.
5. **Build on top (B) later, as a setting**, if the waiting turns out to be
   too slow. It costs a restack every time 0002 changes, and rework when
   0002's preview finds a problem.

Why A first: the human's time (design, approval, preview) is the slow part,
and designing ahead keeps it busy. Building is fast with AI workers, so
waiting for the merge costs little. B adds restacks, rework, and harder
recovery.

### 2.7 Cancel a slice

**From the human (2026-10-18):**

- Cancelling stops everything that is running, and remembers where the slice
  was. The history stays, so it is **recoverable**.
- The slice's status is **cancelled**.
- An agent that was running gets the status **interrupted**.
- Everything else stays as it is. With no agent running, there is nothing
  more to do.

**Decided (the human, 2026-10-18):** the points below, and the stacked case.
(The slice number point was explained on request; see below.)

- **Cancelled is a mark on top of the phase, like blocked** (1.4):
  "implement (cancelled)". The phase never changes, so "where it was" is
  never lost. `flow slice reopen` removes the mark, and the slice goes on
  from there.
- **Commands:** `flow slice cancel --reason "…"` (only when the human says
  so, with their words in the event) and `flow slice reopen`.
- **Before an agent is stopped, its work is saved** on its task branch.
  The task is marked interrupted. On reopen, an interrupted task goes back
  to todo, and its work is still on its branch (the rule in the workflow
  files today: an interruption is not a failure).
- **Kept:** the slice branch, every task branch, the event log, `.history/`,
  reviews. **Removed:** the worktrees in `.flow/`, to free disk space; they
  can be made again from the branches on reopen.
- **The slice number is never used again.**
- **Stopping without cancelling is a different thing:** `flow run stop`
  stops the run (agents interrupted, work saved) but the slice is not
  cancelled. For "pause for now", not "we do not want this".
- **Events:** `slice.cancelled` { by, reason, at }, `slice.reopened`,
  `task.interrupted` { task, at }.

**A slice stacked on a cancelled one** (0003 waits for 0002, and 0002 is
cancelled): `flow` does not change 0003 by itself. It shows
"0003 waiting: 0002 was cancelled", and the human chooses:

- **Move 0003 onto `main`:** the dependency is removed (maybe with a spec or
  plan change first, if 0003 needed something from 0002).
- **Cancel 0003 too.**
- **Leave it waiting,** if 0002 may be reopened.
