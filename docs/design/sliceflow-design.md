# sliceflow design

> **Status: current design.** Last updated 2026-10-29.
>
> This document holds the principles: who decides what, and how the parts
> fit together. The deep details (exact fields, counters, timers, flags,
> messages) are settled in the spec of the slice that builds each part.
> The order of the slices is in `roadmap.md`. Section 4 lists every
> decision with its date, including the ones that were replaced.
>
> Started 2026-10-18 as a brainstorm. Rewritten on 2026-10-29 to match
> slice 0001 as built, the template decisions of 2026-10-28, the workflow
> decisions of 2026-10-28, and the roadmap of 2026-10-29.

---

## 0. How we work on this document

- **One topic at a time.** The human says what they need. The lead
  researches, gives 2-3 real options with trade-offs, and recommends one.
  Questions come only for points that are really open, one at a time.
- **Principles here, details in the slice** (the human, 2026-10-19). When
  a point is mostly detail, this document says "settled in the slice".
- **A decision says who made it and when**, e.g. (decided, the human,
  2026-10-28). Everything else is the lead's proposal.
- A replaced decision is not deleted from section 4. It is marked
  "replaced", with the date and what replaced it.

---

## 1. The big picture

### 1.1 What sliceflow is

`sliceflow` is a command-line tool that runs a spec-driven workflow with AI
coding agents. It is written in Go, and runs on Windows, macOS, and Linux.
The command is `sliceflow`, one word; prose may write SliceFlow.

**The pitch** (the human, 2026-10-28):

> **sliceflow: a CI pipeline for agent work, with a control panel.**
> Approve the plan. Walk away. Come back to a reviewed, tested branch.
> Local only. Works with Claude Code and Codex.

What sets it apart is the combination, not any one part: an approved spec
and plan; a run that follows fixed rules in code; review gates; recovery
after a crash; a "Needs you" list; and all of it controlled from a local
web dashboard. A web board alone is not new (vibe-kanban has one).

A **slice** is one feature, built on its own branch, from spec to merge.
The workflow runs from creating the slice to finishing it: a local merge
into `main`, or a pull request (2.5).

### 1.2 Rules in code, judgment in agents

Two ways to run the same workflow:

| | An AI lead orchestrates | Hybrid: `sliceflow` runs the rules, agents do the judgment |
|---|---|---|
| Same input, same steps? | No. A model can skip a step or bend a rule. | Yes. The rules are code, and tests prove them. |
| Cost | Tokens on every routine step | Tokens only on real work and on stops |
| Long runs | The lead's context fills up | Each agent starts fresh |
| State | Kept by hand; it drifts | Written by `sliceflow` from what really happened |
| Unusual cases | Handled well | `sliceflow` stops and asks a lead |

**Decided: the hybrid** (the human, 2026-10-18). The sources agree:
Anthropic separates fixed workflows from open-ended agents and advises the
simplest design that works; the MAST study of 1,600+ multi-agent runs
found most failures come from design and coordination, not from the model.
A warning we keep: a fallback that is the same kind of model can repeat
the same mistake, so a fallback gets a narrow job and clear limits.

### 1.3 Three kinds of work

| Kind | Who | Examples |
|---|---|---|
| Rules | `sliceflow` | pick the next task, start an agent, run the checks, count failures, merge, record events |
| Judgment | agents (lead, workers, reviewers) | write the spec and plan, write code, review, answer a question whose answer is in an approved document |
| Decisions | the human | approve, choose between options, accept a risk |

`sliceflow` calls a lead only when a step needs judgment and no worker or
reviewer owns it. Agents act through `sliceflow` commands, so the state
stays correct.

### 1.4 Phases and lifecycle

```
spec → plan → implement → preview → done
```

- **Phases** say what kind of work is going on. **States inside a phase**
  (draft, in review, awaiting approval, approved) are separate.
- **No "ready" and no "running" state.** "Ready" means the spec and plan
  are approved; it is worked out, not stored. Whether a run is going shows
  in the tasks of the implement phase.
- **Blocked and cancelled are marks on top of a phase**, e.g. "implement
  (blocked: ...)". The phase stays, so "where it was" is never lost.
- **The state machine grows slice by slice** (the human, 2026-10-24).
  0002 builds the engine once: how a move is described, replaying the log
  through the moves, refusing a move that is not allowed, and showing what
  comes next. Each later slice adds only its own moves.
- **Every step can be stopped and run again** (the human, 2026-10-24). A
  step `sliceflow` runs by itself records started, then done, failed, or
  stopped, and is safe to run twice. Retrying is one more move.

### 1.5 One autonomous run

After the human approves the plan, `sliceflow` runs by itself. How far it
goes is a setting: **to preview** (build, review, and merge every task, run
the integration review, then wait for the human), or **preview prepared**
(the same, and the preview is set up for the human). There is no "build
only": the run always goes at least to preview. The default is settled in
the slice.

### 1.6 The escalation ladder

```
sliceflow  ──can't──▶  lead  ──can't──▶  human
```

**Decided (the human, 2026-10-19):**

1. **The test for every stop: can it be undone, and does it stay inside
   what the human approved?** Yes: the lowest level that can do it
   decides. No: the human. That covers changing an approved document,
   accepting a risk (a high finding left unfixed), anything new (a
   dependency, a secret, a CI change), and merging or pushing.
2. **Route by kind, do not always climb.** A stop goes straight to the
   lowest level that has the authority.
3. **Each level tries once.** No loops (1.7).
4. **`sliceflow` enforces what a lead may do.** Only the human may approve,
   reject, accept a high finding, or cancel.
5. **Every case has the same shape:** what stopped, the evidence, what was
   tried, the options with a recommendation, and what happens meanwhile.
6. **The human is not paged.** Cases for the human wait in one "Needs you"
   list, shown by `sliceflow slice status` and the dashboard, with a
   browser notification. The run goes on with everything that does not
   depend on them.

**Who handles what** (a first version; each slice settles its rows):

| Stop | sliceflow | lead | human |
|---|---|---|---|
| Passing failure (network, the tool crashed, a rate limit) | waits and retries; a rate limit waits until its reset time | — | told if it lasts |
| Checks fail (a counter) | new attempt | cause visible: a hint, one more try | cause unclear, or the hint did not help |
| Worker question | — | answer in an approved document: answers | a new decision |
| Review round limit | fix rounds | prepares the case | decides |
| Merge fails (a counter) | merges again with a worker | widens `touches` inside the rules, or finds a missing dependency | anything that changes the plan |
| Agent over its time limit | — | continue, restart, or stop | if it is stopped for good |
| Reviewer: "needs the human" | — | — | decides |
| Budget cap reached | stops starting agents for it | — | decides |
| New dependency, secret, CI change | — | — | decides |
| `sliceflow` itself broken (a bad log, an internal error) | stops safely | must not repair the log by hand | decides |

**Approvals record how they were given** (the human, 2026-10-28). An agent
runs as the same user on the same computer, so no local check is a perfect
lock; the sandbox is the wall (1.9). The goal is that self-approval is hard
and always visible. Each approval event stores `via`:

- `terminal`: the approve command asks the human to type the slice name
  in a real terminal. An agent's command without a terminal is refused.
- `browser`: an Approve button in the dashboard. Its token is stored
  outside the repo, where an agent's sandbox cannot read it.
- `lead`: the chat lead runs the command with the human's words
  (`--human "<words>"`). Allowed, but marked.

Later, a project setting can refuse `via: lead` (a team-only policy, 1.14).

### 1.7 The lead

**Who runs `sliceflow`.** Mostly the lead: the chat agent the human talks
to. So every output is read first by an agent: it must be exact and
complete (what happened, what is next, why it stopped), and still plain
enough for the human. Any session can be the lead, because the state is in
files, not in a chat.

**Three kinds of commands** (the human, 2026-10-19):

| Kind | Examples | How the lead uses it |
|---|---|---|
| Quick | `slice create`, `spec validate`, `slice status` | runs it, reads the result |
| Starts work | `spec submit`, `plan submit`, `run` | `sliceflow` starts the work in its own process and returns at once |
| The doorbell | `sliceflow wait` | runs it in the background; it ends when something needs the lead or the human, and the agent tool wakes the lead |

- **The work runs in its own process**, so closing a chat never stops a
  run.
- **The mailbox is the event log and "Needs you".** Every case is written
  there first. The doorbell only wakes; it can fail, so nothing depends on
  it.
- **The listening mark.** While `sliceflow wait` runs, it leaves a mark in
  `.sliceflow/`, so `sliceflow` knows a chat lead is there.
- The lead side needs no adapter: the lead only runs commands. Claude Code
  wakes the lead when a background command ends. Cursor has no reliable
  wake-up yet, so a Cursor lead does not listen during a run.
- **One guide for every lead tool**, shipped as a skill (first release,
  0008).

**The fallback lead is required and always on** (the human, 2026-10-28).
When a case needs a lead and no chat lead is listening, `sliceflow` starts
a fresh lead for that one case, with its own time limit and narrow
permissions (1.9). It wakes only when something is wrong: no polling, a
fresh context with just that case. Without it, "walk away" breaks.

- **A lead may answer "blocked: needs the human"**, with a reason. The case
  then goes to "Needs you", and `sliceflow` never wakes a lead for it
  again.
- **Open:** what a "solved" answer looks like: an answer `sliceflow` acts
  on, or the lead simply continues the flow through commands. Settled in
  the slice that builds it (0010).
- An off switch may come later, with configurable steps.

### 1.8 Monitoring, reliability, and the dashboard

One choice carries all of these: **`sliceflow` writes down every step
before and after it does it** (the event log, 1.13). The dashboard reads
it. Recovery reads it.

- **Check reality on restart.** `sliceflow` compares the log with what is
  really there (branches, worktrees, commits, processes) and resumes,
  restarts a task (its work is kept on its branch), or reports.
- **Heartbeats.** While an agent runs, `sliceflow` records that it is alive
  and when it last did something. Dead, or silent too long: `sliceflow`
  knows, and says so.
- **Two kinds of failure.** *Passing* (network, a crashed tool, a rate
  limit): `sliceflow` waits and retries; it does not count as the task
  failing. *Real* (checks or review fail): the counters and the ladder.
- **Two kinds of state.** *History* (what happened to the slice) is in the
  event log, in git. *Live* (which process runs, the last heartbeat) is
  local in `.sliceflow/`, not in git, and never a source of truth.
- **Time limits are per agent run, per role** (the human, 2026-10-18): a
  worker, a reviewer, and a lead each have their own; each fresh worker
  starts its own clock. Over a limit, a lead decides: continue, restart,
  or stop.
- **Parallel slices.** Each slice has its own run and lock. Merges into
  `main` happen one at a time.
- **The agents view.** `sliceflow` starts every agent, so it knows who runs
  where: role, slice, task, tool, model, started, last activity.

**The dashboard** (0006, read-only first; actions in 0007). `sliceflow
serve` on 127.0.0.1, with a token and Host and Origin checks. It shows
every slice live: its tasks, agent activity, usage, and "Needs you". It
calls the same core as the commands (0001 SD13), never the commands
themselves. In 0007 it gets actions: stop, continue, retry, approve or
reject (with the document and what changed), browser notifications, and
"Open in editor".

**The diff by state** (the human, 2026-10-29). `sliceflow` knows the whole
story from the event log, so a user picks a comparison by name, not by
commit: the whole slice against where it started, one task, one review
round against the round before, everything since an approval, and
everything since the user last looked. `sliceflow` finds the commits behind
the name from the events; git computes the diff. The core and `sliceflow
diff` come in 0005; the browser view in 0006; side-by-side with comments
on lines that go to the agent is Part B.

### 1.9 Starting agents, and permissions

**Adapters.** `sliceflow` starts every worker, reviewer, and fallback lead
itself, through the agent tool's headless mode. One small adapter per tool,
all the same shape: start, events, result, resume, stop. **Codex first**
(the human, 2026-10-19), through the Codex app server; Claude Code (`claude
-p` with streaming JSON) comes in 0011. Both give a stream of events and a
session that can be resumed.

**Permissions: three layers** (the human, 2026-10-19). Nobody is there to
click "allow".

1. **A fixed list per role**, set before each agent starts. Everything
   else is refused, not asked.

   | Role | May | May not |
   |---|---|---|
   | Worker | read the repo, edit files in its worktree, run the setup and check commands | commit, merge, push: `sliceflow` does all git work |
   | Reviewer | read everything | change any file; its answer is JSON (below) |
   | Fallback lead | read, run `sliceflow` commands | edit code, git; approve, reject, cancel |

2. **A sandbox on top, where the system has one:** writes only inside the
   worktree, network only to hosts the project lists. Native Windows has
   no sandbox for Claude Code: the docs say so plainly.
3. **`sliceflow` checks the result after the agent ends**, on every system:
   only files in the task's `touches` changed, and no approved document or
   event log was touched. A failed check is a case.

- An agent that needs something refused says so in its result ("blocked:
  needs network for `go mod download`"). That is a case for the ladder.
- "Skip all prompts" only as a setting the human turns on, inside a
  container or virtual machine (the container profile is Part B).
- Rules alone have holes (a full path, `sh -c`); only a sandbox closes
  them. A real case: an autonomous setup pushed to `main` by accident.

**Reviews answer in JSON** (the human, 2026-10-28). A reviewer's final
message must match a JSON schema: Codex's `outputSchema` on `turn/start`
(checked in codex-cli 0.160.1), Claude Code's `--json-schema` (read from
`structured_output`). The tool enforces the shape, so there are no format
errors and no Markdown parser for reviews. A schema guarantees the shape,
not the quality. Each review is saved as one event in the slice's log
(1.13). `sliceflow` sets the verdict (pass, changes, needs the human) from
the findings; the agent never does. The rules of the old review template
become the reviewer's instructions: high, medium, and low levels; `Breaks:`
with the scope rule (behavior nothing asks for is at least medium); each
test marked ok, weak, or missing; previous findings verified or reopened;
the requirements check; questions with a `Guess:`.

**Usage** (the human, 2026-10-28). Every agent run records its usage in
the event that marks it done: tokens (input, cached, output, reasoning),
time, and dollars only when the tool reports them. Codex reports tokens
and rate-limit windows; Claude Code reports a dollar cost. `sliceflow`
never guesses prices. `slice status` and the dashboard show usage per
task, per review, and per slice, and Codex users see their rate-limit
window. Optional budget caps in tokens, per task and per slice, with no cap
by default (0015).

**Known facts for 0004** (from the human's other project, 2026-10-19): the
Codex sandbox makes git's own files read-only (fine: `sliceflow` does git
work); tool caches outside the worktree are read-only (Go needed `GOCACHE`
per worktree); the sandbox blocks even `localhost`, so test services start
outside it; "model at capacity" is a passing failure; Codex helper
processes can outlive a worktree, so stop them before removing it.

### 1.10 A configurable workflow

Three levels of "configurable":

| Level | Example | Effort |
|---|---|---|
| 1. Settings | limits, run reach, review on or off | easy |
| 2. Steps from a fixed set | per phase, pick and order built-in steps | medium |
| 3. Any workflow | your own steps and states | hard: a general engine |

Levels 1 and 2 (configurable steps is required, slice 0014). Level 3
breaks the reliability promise: `sliceflow` can recover a step safely only
if it knows what the step does. Each built-in step has the same shape
(what it reads, what it writes, its events, safe to run twice), so level 3
stays possible later.

### 1.11 Build or reuse?

No existing tool matches: spec kits (Spec Kit) have no orchestration;
agent runners (Claude Squad, vibe-kanban, Conductor) have no specs,
approvals, or recovery; workflow engines (Temporal) need a server and know
nothing about agents. **Decided (the human, 2026-10-18):** build `sliceflow`
as a specific engine for this workflow, and reuse the agent tools, git,
and the ideas (durable steps, task state machines, spec files). A general
engine is a later, optional project.

### 1.12 No data out

**Decided (the human, 2026-10-18):** `sliceflow` itself sends no data to any
outside service: no telemetry, no uploads, no calls home. It reads and
writes repo files and runs local programs: git, the checks, the agent
tools. A dashboard on the human's own computer is fine.

- Pushing a branch and opening a pull request send code to the user's own
  remote. That is never a default: it happens only with `finish: pr`, and
  each push needs a human approval (2.5).
- `sliceflow` never fetches prices or anything else from the internet.

### 1.13 Events and files

**The event log is the only source of truth** (the human, 2026-10-18;
0001 SD2). Each slice has one log, `docs/slices/<name>/.events.jsonl`,
one event per line.

- Only `sliceflow` writes it. Not the human, not a lead, not an agent.
- It is committed in git with the slice. It holds the steps that matter,
  not heartbeats.
- There is no `status.md`. `sliceflow slice status` and the dashboard are
  views of the log. Later, a pull request's description can carry a short
  status.
- Each event has `version`, `seq`, `event`, `at`, and `data` (0001 SD3). A
  bad line is skipped with a warning, and never stops a command (SD5).
- Several `sliceflow` processes may write at once: a lock while a line is
  added. Crash safety: "started" before a step, "done" after it.
- **When the log is committed** (the human, 2026-10-28): each event is
  written to the file at once, and `sliceflow` commits the file when a step
  ends (an approval, a task merged, a review done, the run stopped or
  finished), not after every event. On restart it commits what is left. An
  approval's tag points at a commit that holds the approval event. One lock
  covers log commits and merges.
- **Reviews** are events too (1.9). A finding's state (fixed, wontfix,
  deferred, verified, reopened) is more events. There is no review file and
  no findings file.

**File formats** (the human, 2026-10-19):

- **Files people read, approve, or edit:** Markdown, with a small YAML
  header, and a YAML block for exact fields (one per task in `tasks.md`).
- **Files only `sliceflow` writes:** JSON Lines (`.events.jsonl`).
- **Settings:** YAML, read strictly (`sliceflow.yaml`, 1.14).
- Why: models understand the common formats about equally; Markdown costs
  the fewest tokens; strict formats make free text worse. So accuracy comes
  from rules, not the format: each fact has one copy (requirements are
  named, not copied); every file is read strictly, with unknown keys
  refused; state changes only through `sliceflow` commands.
- Exception: a reviewer answers in JSON, because the tool enforces the
  schema and no human writes it (1.9).

**The documents** (the human, 2026-10-28). Every slice has the same three
documents, in `docs/slices/<name>/`: `spec.md`, `plan.md`, and `tasks.md`.
There are no optional sections: a part is always there, or there because
of a fact (e.g. Interfaces when tasks share a name). "None" is allowed where
a part can truly be empty. 2.2 and 2.3 say what each holds.

- **Templates are built into `sliceflow`** (the human, 2026-10-20), so
  they always match its checks. A repo needs no template files. A filled
  document no longer holds its rule comments, so `sliceflow` prints a
  template on request (named in 0002).
- **All Markdown reading lives in one Go package** (the human,
  2026-10-23). It turns a document into plain data; the rest of `sliceflow`
  never looks at headings. This keeps Spec Kit support and templates a
  project can change (both Part B) cheap.
- **The backlog** is `docs/backlog.md` on `main`: one table (Item, From,
  Added, Closed by), a short paragraph per item, rows never deleted. It
  stays Markdown, not events: it is project-wide while the log is per
  slice, it is its own truth, people edit it by hand, and it is read on
  GitHub (the human, 2026-10-28).

### 1.14 Settings

**Two layers, plus defaults** (the human, 2026-10-28):

| Layer | Where | For |
|---|---|---|
| Team | `sliceflow.yaml` at the repo root, committed | rules everyone shares: worktree setup, reviews, budget caps, `finish:` |
| Personal | git config, `sliceflow.<key>` | one person's choices: `sliceflow.base`, `worktree.dir`, the editor |
| Default | built in | e.g. base: `main` |

- Personal beats team, and team beats the default.
- **Policy settings are team-only**, e.g. "approvals only from the browser
  or the terminal". Git config cannot change them.
- An unknown key in `sliceflow.yaml` is an error that names the key.
- 0001's `git config sliceflow.base` keeps working as a personal override.
- No extra dependency: the YAML reader is needed for front matter and
  task blocks anyway. Settled before 0004 adds per-role settings.

### 1.15 Worktrees

**Outside the repo** (the human, 2026-10-28). Default:
`~/.sliceflow/worktrees/<repo>/<slice>/`, with task worktrees under it. A
worktree inside the repo can break ESLint (it reads the parent repo's
config too) and Jest (duplicate files). Windsurf and similar tools keep
theirs outside too.

- A setting, `worktree.dir`, chooses another folder (Cursor users asked
  for this).
- **Reading the code** stays easy: `git diff main...slice/<name>` or the
  editor's git view from the normal checkout (all worktrees share one
  repo); `sliceflow open <slice>` opens the worktree in the user's editor;
  the dashboard has an "Open in editor" button; outputs print full,
  clickable paths.
- Cleanup of worktrees left behind by a deleted repo: `git worktree prune`,
  later a `doctor` command. Watch path length on Windows.
- 0001 built `.sliceflow/worktrees/<name>/` (SD7). It moves before the first
  public release, at the latest with 0005. `.sliceflow/` stays, for live
  state.

**Reused, with setup** (the human, 2026-10-28). A new worktree has only
tracked files: no `.env`, no installed packages, and ports clash.

- Settings, all optional and empty by default: `worktree.copy` (small files
  from the main checkout, e.g. `.env`; not `node_modules`),
  `worktree.setup` (commands run in a worktree, each as its own process,
  like Cursor's `worktrees.json` or Conductor's `conductor.json`),
  `worktree.setup_when_changed` (e.g. `yarn.lock`: setup runs again only
  when these changed since the last run in that worktree), and
  `worktree.port` (a free port per worktree, as `SLICEFLOW_PORT`).
- **One worktree per worker slot**, made once per slice. For each task,
  `sliceflow` switches it to the task's branch and cleans tracked and
  untracked files, but keeps ignored install folders. So a slice pays at
  most one install per worker slot, not one per task.
- Copied `.env` files are secrets kept under the worktree folder, outside
  git. The docs say so.

### 1.16 Still open

Principles: none. Details, settled in the slice that builds them:

- What a lead's "solved" answer looks like (1.7; 0010).
- Which run reach is the default (1.5).
- Each event's fields beyond the shared ones.
- About 30-50 hours to shrink on the path to the first release
  (`roadmap.md`, Time); each slice's spec settles its own cuts.
- Changes to an approved document, in three kinds (backlog, 2026-10-24;
  with 0003).

---

## 2. The processes

Each process at the level of principles. The commands, events, and
messages are settled in its slice.

### 2.1 Create a slice (built in 0001)

`sliceflow slice create <name> [--title "..."] [--from <branch>]`

- **The name is the slice ID**, with no number (the human, 2026-10-19;
  SD1): branch `slice/<name>`, folder `docs/slices/<name>/`, commits
  `<name>: ...`. A number picked by counting slices clashes when two
  people work offline. A name is never used again, even after cancel.
- **Commands are `sliceflow <noun> <verb>`** (the human, 2026-10-18), like
  `gh pr create`.
- **Check first, then change.** Every check runs before anything is
  created. If a step fails halfway, `sliceflow` undoes what it did (an undo
  list), and says "Nothing was changed".
- It adds exactly two files on the slice branch, in one commit: the
  spec, from the built-in template, and `.events.jsonl`, with
  `slice.created`.
- The base branch comes from `--from`, else `git config sliceflow.base`,
  else `main` (SD6). The title lives only in the event log (SD12).
- `sliceflow slice status` shows where a slice is and what comes next;
  `sliceflow slice list` lists local slices. Both only format what a shared
  core returns (SD13), so the dashboard calls the same core.

### 2.2 Write and approve the spec

1. The lead and the human write `spec.md`. Its state is **draft**.
2. `sliceflow spec validate` checks it like a compiler, one problem per
   line, e.g. `spec.md:42: error: R7 has no Check: line`. Every error
   carries a hint on how to fix it. Errors block submit; so do warnings
   such as open questions.
3. `sliceflow spec submit` validates again, then starts the spec review
   (if on), or goes straight to awaiting approval.
4. The reviewer answers in JSON (1.9). Open points go back to draft; the
   lead answers facts itself and brings decisions to the human.
5. The human approves (1.6: `via` terminal, browser, or lead), or
   rejects with a reason.

**What a spec holds** (the human, 2026-10-28). A spec says WHAT, never HOW.

- **Summary:** two or three sentences, what it lets someone do, and why.
  The why helps a worker guess right when a requirement leaves something
  open.
- **User stories:** every scenario sits under a story; each story always
  has a short title and one sentence (who wants to do what). Scenarios
  are numbered S1, S2, ... across the spec, with Given / When / Then, the
  real output, and a `Shows:` line naming their requirements. Every
  requirement a user can see appears in at least one scenario.
- **Requirements:** R1, R2, ..., grouped under headings. EARS forms
  without "shall", in the present tense: "When <trigger>, sliceflow
  prints ...", "If <bad case>, sliceflow ...", "While", "Where", or a plain
  "sliceflow ...". "never" for what it must not do. Exact values (the
  message, the limit, the exit code). One behavior each. A term that needs
  one fixed meaning is defined where it first appears. Each has a `Check:`
  line: unit, api, e2e, or manual.
- **Out of scope:** one line per item. What should be done later goes into
  the backlog directly.
- **Decisions:** SD1, SD2, ..., each with `Why:`, `Not chosen:`, `Cost:`
  (required; "None" allowed), and `Rule:` (the requirement that states it,
  or none). `Replaces:` when it replaces an earlier slice's decision.
- **Open questions:** must say None before the review starts.

Numbers are renumbered freely only before the first approval. After it, a
number is never reused, and a removed item stays, marked removed.

**History is in git** (the human, 2026-10-20). Each approval makes one
annotated git tag on the approved commit, e.g.
`sliceflow/<name>/spec/R001`; its message holds the approval record (what,
the revision, the date, the human's words, `via`). The approval event also
stores a hash of the approved files' contents, so the approval can be
checked even if a commit is lost after a squash merge (the human,
2026-10-28). `sliceflow` turns on `push.followTags` in clones it sets up.
Others read the history with plain git (`git tag -l "sliceflow/<name>/*"`,
`git show <tag>:<path>`).

### 2.3 Write and approve the plan and tasks

It works like the spec: draft, validate, submit, review, awaiting approval,
approve. **One planning phase, one approval** (the human, 2026-10-21): the
plan and its tasks are approved together, after the spec. The plan review
always runs.

**What a plan holds** (the human, 2026-10-28). The plan says HOW. It never
adds scope: what the user sees comes from the spec, and a gap there is an
Open question, so the spec changes.

- **Approach:** the idea in a few sentences, a table of the parts that
  change, and new dependencies (each needs the human's approval).
- **Flows:** at least one; the steps in order; a Mermaid diagram for
  complex flows. **Shared rules** under it, including each project rule
  that applies (a `Project rule:` line saying where it comes from); a
  missing one is a plan review finding.
- **Interfaces, Data:** "None" when there is nothing.
- **Testing:** how each kind of check runs, so every task uses the same
  helpers; and **Trying it by hand**: Setup steps, and steps for each
  `Check: manual` requirement, each with an `Expect:` line.
- **Decisions:** TD1, TD2, ..., like spec decisions, with `Cost:` required.
  Each is named in at least one task's `follows`, or has a `Plan only:`
  line saying why.
- **Risks and rollout:** always there; "None" allowed.
- **Open questions:** must say None before the tasks are written.

**What `tasks.md` holds** (the human, 2026-10-20 and 2026-10-28). All tasks
in one file; the lead writes all of it, and `sliceflow plan validate`
checks every fixed part strictly. Each task:

- `## Task 01: <name>`, then a YAML block: `covers` (requirements),
  `depends_on`, `touches` (the files and folders it may change; `sliceflow`
  rejects a change outside it), `follows` (plan items), `risk` (core,
  normal, or low).
- **Goal**; **Steps** in order: tests first and seen to fail, then the
  code, then the tests until they pass. Together the steps name every file
  the task creates or changes, test files included. When similar code
  exists, a step names the file to copy the style from. A rule lives in
  one place only: the spec if the user sees it, a plan shared rule if
  several tasks need it, otherwise the step that needs it.
- **Interfaces**, exactly when the task shares a name with another task
  (creates, changes, or uses). A task that uses a name another task
  creates must depend on it; `plan validate` checks this.
- **Done when:** each test with real input and expected result, a
  `Proves:` line, and a `Scenario:` line for an e2e test of a spec
  scenario. Every scenario gets at least one e2e test. Each task puts its
  e2e tests in its own file.
- **Out of scope**, with `Done in:` (another task, or not in this slice).

**Risk** (the human, 2026-10-20): `core` (others build on it, stored data,
a file format, security, a contract or shared interface), `normal` (other
code), `low` (no change in behavior). Core and normal tasks get a code
review; low tasks do not. There is no simple/complex track.

**The brief** (the human, 2026-10-20). `sliceflow brief` gives each worker
only what its task links to, and never guesses: its own section, the text
of the requirements in `covers`, the scenarios whose `Shows:` names one of
them, and the exact text of the plan items in `follows`. It ends by saying
where the full spec and plan are. `plan validate` warns when a brief is
over a set size (about 1,500 words), for the plan reviewer to judge.

**After approval**, `tasks.md` may change without a new approval only by
fix tasks added at the end (covering approved requirements only) and a
wider `touches` by the rules. Anything else is a plan change.

### 2.4 The run

`sliceflow run` builds an approved plan by itself.

1. **Pick a task** whose dependencies are done. In 0005, one at a time;
   in 0013, up to the worker limit, never two tasks with overlapping
   `touches` at the same time.
2. **Start a worker** in a worktree (1.15), with its brief. A fresh worker
   for each fix round, with the findings and what the last worker did.
3. **Checks**, then the **task review** (by risk). A failed review goes
   back to a worker, up to the round limit; then the lead prepares a case
   for the human.
4. **Merge into the slice branch, one at a time** (a merge queue with a
   lock). Only what passed the checks on the exact result of the merge is
   merged (the "not rocket science rule").
5. When all tasks are merged: update from `main`, run the checks, and the
   **integration review** of the whole slice. It runs even when task
   reviews are off (the human, 2026-10-18). Its fix tasks run like any
   task. Then the preview.

**Commits during a run** (the human, 2026-10-28):

- The worker commits every round on the task branch (the first build,
  then each fix), so a reviewer sees what changed since the last round,
  and every commit an event names exists.
- A finished task joins the slice branch with `git merge --no-ff`: one
  merge commit per task. `git log --first-parent` shows one line per task;
  the rounds stay inside. Never squashed, and commits an event names are
  never rewritten.
- The event log is committed when a step ends (1.13).

**Merge conflicts** (the human, 2026-10-18). The task being merged owns the
conflict: a worker on that task solves it, with a brief of the files in
conflict and what the other task changed. A failed merge does not count
against the worker's check failures; after the merge limit, it is a case
(often a plan problem: overlapping `touches` or a missing dependency).
After a conflict is solved, passing checks are enough, unless the solution
changed more than the fit: `sliceflow` measures this (only the conflicted
files changed, or not), it does not ask the worker. Conflicts are shown,
and the run keeps going with other tasks.

**Stop, continue, retry.** `sliceflow stop` stops the run (agents
interrupted, work saved on their branches); `sliceflow run` continues from
the log. A failed task can be retried (0005); any failed or stopped step
from 0012.

### 2.5 Preview and finish

**The preview checks intent** (the human, 2026-10-20). Tests prove the code
does what the tasks say, not that the tasks say what the human meant: in a
study by Wang, Pradel, and Liu, 29.6% of patches that pass the tests
behave differently from the correct fix. So the human tries the result
before it leaves the slice.

- **The checklist is built by `sliceflow` from the spec** (the human,
  2026-10-28): the first scenario of each story, with its exact command
  and output, and each `Check: manual` requirement. The plan's "Trying it
  by hand" adds the setup and the manual steps. There is no `preview.md`.
- A problem: fix tasks, and the run goes on. Something new: a spec change,
  not a fix.
- The human approves the result (`via`, 1.6).

**Finish** (the human, 2026-10-28). A team setting, `finish:`:

- **`merge` (the default):** update the slice from `main` once more, run
  the checks (a conflict becomes a fix task, then review and preview
  again), and merge into `main` locally with `--no-ff`, one slice at a
  time. Nothing leaves the machine.
- **`pr` (opt-in):** `sliceflow` pushes the slice branch and opens a pull
  request with `gh`. Its description is written from the slice: the spec's
  summary, the tasks, the review results, and usage. The push needs a
  human approval (`via` terminal or browser), because it sends code out
  and is hard to undo. When the pull request is merged on GitHub,
  `sliceflow` sees it on the next `git fetch` and marks the slice done.
  Approvals survive a squash merge through their content hash (2.2).

**Findings deferred out of the slice** go to the backlog at the end, with
their source (e.g. "slice due-dates, review task-04-1, F3"). Low findings
never fixed are not copied: they were optional. The slice's worktrees are
removed when it is done.

### 2.6 Design ahead: stacked slices (Part B)

The human wants to design the next slice while one is being built
(dual-track agile). **Decided (the human, 2026-10-18):**

1. **Design ahead, always.** A slice can be created, specified, and planned
   at any time. A stacked slice records what it waits for (`after:`). Its
   branch starts from `main` and holds only its own documents until it
   builds.
2. Its plan reviewer reads the approved spec and plan it builds on.
3. If that plan changes after this plan was approved, this plan is marked
   stale, and approved again.
4. **Build after the merge.** When the slice below is merged, `sliceflow`
   updates this slice's branch from `main`. Starting its run then asks the
   human first (the human, 2026-10-28), instead of starting by itself.
5. Building on top of an unmerged slice (restacking) only if waiting turns
   out to be too slow.

### 2.7 Cancel a slice (0015)

**Decided (the human, 2026-10-18):**

- **Cancelled is a mark on top of the phase**, like blocked.
  `sliceflow slice reopen` removes it, and the slice goes on from there.
- `sliceflow slice cancel --reason "..."` is a human-only action (1.6).
- Before an agent is stopped, its work is saved on its task branch, and
  the task is marked interrupted; on reopen it goes back to todo. An
  interruption is not a failure.
- **Kept:** the branches, the event log, the approval tags. **Removed:** the
  worktrees, to free disk space; they can be made again on reopen.
- **The name is never used again.**
- `sliceflow stop` is different: it pauses a run; the slice is not
  cancelled.
- A slice stacked on a cancelled one is not changed by itself: it shows
  "waiting: <slice> was cancelled", and the human chooses to move it onto
  `main`, cancel it too, or leave it waiting.
- Cancel also cleans up a slice half made by a crash during `slice create`
  (a branch and worktree with no committed log).

---

## 3. The slices

The order, contents, and time estimates of the slices are in `roadmap.md`
(replanned 2026-10-29): Part A (required, 0001-0015) and Part B
(optional). The rules for cutting slices are there too.

---

## 4. Decisions log

One line per decision: the date, who decided, and the section. "The
human" means the human decided; the lead proposed. Replaced decisions stay,
marked.

**2026-10-18**
- Hybrid: `sliceflow` runs the rules, agents do judgment (the human; 1.2).
- Build a specific engine first; a general engine later, optional (the
  human; 1.11).
- Lifecycle spec → plan → implement → preview → done; no "ready" or
  "running" state; blocked and cancelled are marks (the human; 1.4).
- The run's reach is configurable: to preview, or preview prepared (the
  human; 1.5).
- No data out (the human; 1.12).
- The event log is the only source of truth (the human; 1.13).
- `status.md` written from the log. **Replaced** 2026-10-24 by 0001 SD2
  (no `status.md`); the template was deleted 2026-10-28.
- Commands are `sliceflow <noun> <verb>` (the human; 2.1).
- The workflow ends at the merge into `main`; pushing is outside it.
  **Replaced** for `finish: pr` on 2026-10-28 (2.5).
- Approval copies in `.history/`. **Replaced** 2026-10-20 by history in
  git.
- Tasks: requirements are named, not copied (the human; 2.3).
- Time limits per agent run, per role (the human; 1.8).
- When no chat lead is there, `sliceflow` starts a fallback lead (the
  human; 1.7). Confirmed as required and always on, 2026-10-28.
- The integration review runs even when task reviews are off (the human;
  2.4).
- After a solved conflict, checks are enough unless more than the fit
  changed; conflicts are visible and do not block (the human; 2.4).
- Design ahead; build after the merge (the human; 2.6). Starting the run
  automatically after the merge: **replaced** 2026-10-28 by asking first.
- Cancel is a recoverable mark; names are never reused (the human; 2.7).

**2026-10-19**
- Principles in this document, details in each slice (the human; 0).
- The escalation ladder's six points (the human; 1.6).
- Permissions in three layers (the human; 1.9).
- Three kinds of commands; the doorbell `sliceflow wait`; the listening
  mark (the human; 1.7).
- File formats: Markdown with YAML, JSON Lines, YAML settings (the human;
  1.13).
- Slice ID is the name, no number (the human; 2.1).
- Codex first; Claude Code later (the human; 1.9).
- How slices are cut (the human; `roadmap.md`).

**2026-10-20**
- Templates built into `sliceflow` (the human; 1.13).
- History in git: annotated tags, `push.followTags` (the human; 2.2).
- Tasks in one `tasks.md`; the lead writes all of it (the human; 2.3).
- What the brief holds (the human; 2.3). Key concepts in the brief:
  **replaced** 2026-10-28 (Key concepts removed).
- Stable scenario IDs; every scenario gets an e2e test (the human; 2.3).
- Every decision reaches a worker: `Rule:`, `follows`, `Plan only:` (the
  human; 2.3).
- A task waits for the code it uses; tests first; the plan handles project
  rules; a warning for long briefs; limited changes to an approved
  `tasks.md` (the human; 2.3).
- Test files belong to the task, listed under Files. Files: **replaced**
  2026-10-28 (the steps name every file).
- No simple/complex track; what each risk means (the human; 2.3).
- Reviews check tests (ok, weak, missing) and scope (the human; 1.9).
- The preview checks intent (the human; 2.5). `preview.md`: **replaced**
  2026-10-28 by a checklist built from the spec plus the plan's "Trying it
  by hand".
- Assumptions are facts, not choices. **Replaced** 2026-10-28: the
  Assumptions section was removed; guesses become open questions.
- Full EARS with "shall". **Replaced** 2026-10-28 by EARS without "shall",
  present tense.
- An optional `Cost:` line. **Replaced** 2026-10-28: `Cost:` is required.
- The plan says how each success criterion is reached. **Replaced**
  2026-10-28: success criteria were removed.
- Parked items and `Goes to: the backlog` copied by `sliceflow`.
  **Replaced** 2026-10-28: both removed; items go to the backlog directly.
- No findings file; finding text in the review file. Review files:
  **replaced** 2026-10-28 by reviews as JSON events.

**2026-10-21**
- One planning phase, one approval for the plan and tasks (the human;
  2.3).

**2026-10-23**
- The name `sliceflow` (the human; 1.1).
- All Markdown reading in one package; `sliceflow` keeps its own format;
  Spec Kit later (the human; 1.13). Spec Kit is now Part B.

**2026-10-24**
- The state machine grows slice by slice; every step can be stopped and
  run again (the human; 1.4).
- 0001: no `status.md` (SD2); events have `version`, `seq`, `event`, `at`,
  `data` (SD3); bad lines are skipped (SD5); base from `--from`, git
  config, or `main` (SD6); the title only in the log (SD12); commands
  format a shared core (SD13).
- 0001 SD7, worktrees in `.sliceflow/worktrees/<name>/`. **Replaced**
  2026-10-28: worktrees live outside the repo.

**2026-10-28**
- The templates: spec, plan, and tasks lighter, with no optional sections;
  Why folded into Summary; Success criteria, Key concepts, Assumptions,
  `Goes to:`, plan `Covers:`, Parked, Files, Copy the style of, and Rules
  removed; Risks always; Interfaces by fact; story titles and sentences
  always (the human; 2.2, 2.3).
- `preview.md`, `status.md`, `review.md`, and `idea.md` deleted; the
  `ideas:` front matter removed, with 0001's R7 (the human).
- The backlog stays Markdown; no Status column; a short paragraph per item
  (the human; 1.13).
- Reviews answer in JSON and are saved as events (the human; 1.9).
- Approvals record `via`: terminal, browser, or lead (the human; 1.6).
- The fallback lead is required and always on; "blocked: needs the human"
  (the human; 1.7).
- Commits during a run: the log at the end of each step; a task commit per
  round; `--no-ff` merges per task; never squashed (the human; 2.4).
- Worktrees outside the repo, reused per worker slot, with copy, setup,
  `setup_when_changed`, and port (the human; 1.15).
- Usage recorded and shown; rate limits pause; optional token caps (the
  human; 1.9).
- Settings: `sliceflow.yaml` for the team, git config for personal;
  policies team-only (the human; 1.14).
- Finish by a local merge (default) or a pull request (opt-in, the push
  approved by the human); approvals store a content hash (the human; 2.5).
- The pitch (the human; 1.1).

**2026-10-29**
- The roadmap: Part A and Part B; web control right after the run; the
  first public release after it (the human; `roadmap.md`).
- The diff by state: the core and `sliceflow diff` in 0005, the browser
  view in 0006, side-by-side and comments in Part B (the human; 1.8).

---

## Sources

The big picture (1.2):
- Anthropic, Building effective agents: https://www.anthropic.com/research/building-effective-agents
- Cemri et al., Why Do Multi-Agent LLM Systems Fail? (MAST): https://arxiv.org/abs/2503.13657
- Bernstein docs, Deterministic vs. LLM-Based Orchestration: https://bernstein.readthedocs.io/en/latest/architecture/orchestration-approaches/

The lead and the doorbell (1.7):
- Claude Code, Channels reference: https://code.claude.com/docs/en/channels-reference.md
- Gas Town, AGENTS.md (nudge and mail): https://github.com/gastownhall/gastown/blob/main/AGENTS.md
- Background task wake-up not happening (bug report): https://claudeissues.com/issue/21048-bug-background-task-completion-notification-not-appearing-activating-claude
- OpenAI, Unlocking the Codex harness: https://openai.com/index/unlocking-the-codex-harness

Agents and reviews (1.9):
- OpenAI, Codex non-interactive mode: https://developers.openai.com/codex/noninteractive
- Claude Code, headless mode: https://docs.claude.com/en/docs/claude-code/headless

File formats (1.13):
- He et al., Does Prompt Formatting Have Any Impact on LLM Performance?: https://arxiv.org/abs/2411.10541
- Tam et al., Let Me Speak Freely? Format restrictions and LLM performance: https://arxiv.org/abs/2408.02442
- Aider, LLMs are bad at returning code in JSON: https://aider.chat/2024/08/14/code-in-json.html
- StrictYAML, Why implicit typing was removed: https://hitchdev.com/strictyaml/why/implicit-typing-removed/

Worktrees (1.15):
- Zenn, ESLint error with a worktree inside a repo (Japanese): https://zenn.dev/aiya000/articles/d527fbfe288c3f
- Windsurf, worktrees: https://docs.devin.ai/desktop/cascade/worktrees
- Cursor forum, configuring the worktree location: https://forum.cursor.com/t/configure-worktree-location/149222
- pnpm, Git worktrees: https://pnpm.io/git-worktrees
- Cursor, worktrees: https://cursor.com/docs/configuration/worktrees

The documents (2.2, 2.3):
- EARS (Mavin et al.): https://alistairmavin.com/ears/
- INCOSE 2024, LLMs and EARS in automotive requirements: https://www.incose.org/wp-content/uploads/legacy/presentation/is2024-pdf/paper_505.pdf
- plainlanguage.gov, "shall" and "must": https://www.plainlanguage.gov/guidelines/conversational/shall-and-must/
- Nygard, Documenting Architecture Decisions (ADRs, "all consequences")
- SHERLOC, Structured Diagnostic Localization for Code Repair Agents: https://arxiv.org/pdf/2606.24820
- GitHub Spec Kit: https://github.com/github/spec-kit
- Kiro, Specs: https://kiro.dev/docs/specs/index

The preview (2.5):
- Wang, Pradel, Liu, Are "Solved Issues" in SWE-bench Really Solved Correctly?
