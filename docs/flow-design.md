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
- What the lead receives when it is called, and what it may do.
- Which run choice is the default (1.5).
- Is the agents view needed (1.7)?
- A web dashboard and the rule "flow makes no network calls": a local page
  is not an outside call, but the rule should say so.
- How agents get permissions when nobody is there to click "allow" (1.9).
- How configurable the workflow is: levels 1-2 or more (1.10).
- Exit codes: the sketch says two different things (step 18 and step 27).

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
- **`flow task new` creates a task** with what `flow` can know or check:
  the next number, the file name, the header (covers, depends on, touches,
  risk), and the headings. It refuses at once if a covered requirement or a
  task it depends on does not exist, or if it makes a loop.
- **The lead writes only what needs judgment:** what to build, and "Done
  when" (the tests that prove it).

**The order (lead's proposal):**

1. The lead writes `plan.md`. `flow plan validate` checks it.
2. When the plan is finished, the lead creates the tasks with
   `flow task new`, and fills each one.
3. `flow plan submit` validates everything together: plan, tasks, contract
   changes, conflicts with other slices. Then the reviewer (if on).
4. One review and one approval for the plan and its tasks together. The
   reviewer checks that the tasks really match the plan. Two separate
   approvals would cost the human twice and split that check.

**Different from the spec:** the reviewer gives findings to fix (high,
medium, low), not only questions. The lead fixes them and submits again, up
to the round limit.
