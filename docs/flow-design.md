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

**Lead's proposal, to talk about** (not agreed):

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

### 1.7 Not clear yet

To come back to when the processes are designed:

- The full list of lifecycle states, and what moves a slice between them.
- When and how each level of the ladder hands on (1.6).
- What the lead receives when it is called, and what it may do.
- Which run choice is the default (1.5).

---

## Sources

- Anthropic, Building effective agents: https://www.anthropic.com/research/building-effective-agents
- Cemri et al., Why Do Multi-Agent LLM Systems Fail? (MAST): https://arxiv.org/abs/2503.13657
- Compiled AI: Deterministic Code Generation for LLM-Based Workflow Automation: https://arxiv.org/pdf/2604.05150
- Bernstein docs, Deterministic vs. LLM-Based Orchestration: https://bernstein.readthedocs.io/en/latest/architecture/orchestration-approaches/
- Agentic fail-safes need deterministic modes: https://dev.to/omnithium/the-t-mobile-outage-lesson-why-agentic-fail-safes-need-sos-mode-determinism-1chm
