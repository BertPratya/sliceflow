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

### 1.5 Not clear yet

To come back to when the processes are designed:

- The full list of lifecycle states, and what moves a slice between them.
- When `flow` calls the lead: the exact list of stops.
- What the lead receives when it is called, and what it may do.

---

## Sources

- Anthropic, Building effective agents: https://www.anthropic.com/research/building-effective-agents
- Cemri et al., Why Do Multi-Agent LLM Systems Fail? (MAST): https://arxiv.org/abs/2503.13657
- Compiled AI: Deterministic Code Generation for LLM-Based Workflow Automation: https://arxiv.org/pdf/2604.05150
- Bernstein docs, Deterministic vs. LLM-Based Orchestration: https://bernstein.readthedocs.io/en/latest/architecture/orchestration-approaches/
- Agentic fail-safes need deterministic modes: https://dev.to/omnithium/the-t-mobile-outage-lesson-why-agentic-fail-safes-need-sos-mode-determinism-1chm
