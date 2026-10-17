# Lifecycles

How slice phases, documents, tasks, findings, and checks change status, and how tasks are merged. `flow` enforces these rules. Read this before changing any status.

## Slice phases

```
spec → plan → ready → implementing → integration → preview → done
                       any phase → blocked | cancelled
```

The phase changes only when its event happens. `flow` sets it (bootstrap mode: the lead, at the same moment).

Brainstorming is not a phase. Exploring a vague idea, or the whole system, happens before a slice exists, outside `flow`. A slice is created when the idea is one clear feature.

| Phase | Starts when | What happens in it |
|---|---|---|
| spec | `flow slice new` | The interview, the spec, the spec check, the human's approval. |
| plan | the human approves the spec | The plan and tasks, the plan review, the human's approval. |
| ready | the human approves the plan | Waiting for `flow run` (a stacked slice also waits until the earlier slice is merged). |
| implementing | the first worker starts | Workers build, reviewers check, `flow` merges tasks. |
| integration | all tasks are done | Update from the base branch, slice checks, integration review. |
| preview | the integration review passes | The human tries the slice. |
| done | the slice is merged into the base branch | Nothing. |

- A fix task (from an integration finding, a conflict with the base branch, or the preview) sends the phase back to implementing. When the tasks are done, it moves on again.
- blocked: the whole slice cannot go on, e.g. the plan review fails 3 rounds. When it is solved, the slice goes back to the phase it was in. A blocked task does not block the slice.
- cancelled: only when the human says so.

## Documents

Applies to `spec.md`, `plan.md`, and a slice's change to a contract or other project file.

```
draft → in_review → awaiting_approval → approved
```

The status shows who has the ball: `draft` the lead, `in_review` the reviewer, `awaiting_approval` the human.

- `flow` sets the status when the review is submitted: no questions or findings → awaiting_approval; questions or findings → back to draft.
- Spec, submitted again after answering its spec check: no new requirements → awaiting_approval directly (no new reviewer); new requirements → in_review, and the check covers only the new ones.
- Plan, submitted again after fixing findings: always in_review (the next round verifies the fixes).
- Only an `awaiting_approval` document can be approved. Rejected: back to draft.
- Edited after approval: the status becomes `changed`, then it is submitted again as above. Every change needs a new approval.
- A contract counts as edited when the slice's change to the file no longer matches the last approved patch in `history/contract/`.
- The plan, its tasks, and the slice's contract changes are approved together, with one approval. `flow` still records and snapshots each one.
- A fix task added after the plan is approved (from the integration review, a failing slice check, a conflict, or the preview) needs no new approval when it covers only requirements in the approved spec and changes no plan, contract, or decision. The lead tells the human about each one in one line. Anything else is a plan change (`flow-plan` → `change.md`).

## Tasks

```
todo → in_progress → in_review → done
```

- Checks fail: attempt +1, stays in_progress.
- Sent to review: round +1. The round is the number of reviews started, so it matches the review file name (`task-02-2` is round 2).
- Review asks for changes: attempt back to 0, back to in_progress.
- Simple track, non-core task: in_progress → done when checks pass.
- done means merged into the slice branch (see "Branches and worktrees").
- The merge fails (a conflict, or checks fail after the rebase): back to in_progress, for the same worker. Not a failure: it does not count toward attempt or the stop rules.
- A stop rule hits: blocked. The lead resolves it when the answer is already in an approved document, or, after failed checks, when it can see the cause; otherwise the human does. Then back to todo or in_progress. Unblocking sets attempt to 0.
- Interrupted: back to todo. The work stays on its task branch (bootstrap mode: on a backup branch). An interruption is not a failure.
- Removed by a re-plan: cancelled.
- done is final. A problem found later becomes a new fix task.
- "Waiting" (a dependency is not done) and "stale" (based on an older plan revision) are calculated, never set.

## Branches and worktrees

Branch names and the worktree folder are in the git policy (`AGENTS.md` section 10).

- `flow slice new` creates the slice branch and its worktree. The slice's documents, state, and finished tasks live there.
- A new slice starts from the base branch. `flow slice new <name> --after <NNNN>` starts it from that slice's branch instead (stacking), when it needs something from that unmerged slice: its code, a contract, the data model, or a decision. The lead proposes it; the human confirms. A slice stacks on at most one other slice. If it needs two, write its spec now and plan it after one of them is merged.
- A stacked slice can be specced and planned, but it waits ("waiting for NNNN", calculated) until the earlier slice is merged. Then it updates from the base branch, and `flow run` can start it.
- If the earlier slice's plan or contract changes after the stacked slice's plan is approved, the stacked plan is stale. The lead checks what is affected; a needed change is a plan change (`flow-plan` → `change.md`).
- If the earlier slice is cancelled, the stacked slice moves onto the base branch and goes back to phase plan.
- `flow slice move <slice> --after <NNNN | base>` moves a slice that holds only documents (phase spec or plan). After that, no moving.
- `flow validate` warns when two unmerged slices change the same file from their plans' `touches` and neither is stacked on the other.
- When a task starts, `flow` creates its task branch from the newest slice branch, in its own worktree. The worker works only there.
- When the worker finishes (`flow task review`), `flow` commits its work on the task branch: `NNNN/NN: <what changed>`, with the "Assumed" note in the body.
- A task starts when every task in its `depends_on` is done, its `touches` do not overlap a running task's, and a worker slot is free (a `flow run` setting, default 1). Tasks whose `touches` overlap never run at the same time: the lower number goes first.
- When a task passes its review (or its checks, if it needs no review), `flow run` merges it. One task at a time:
  1. Rebase the task branch onto the newest slice branch.
  2. Run the task checks.
  3. Move the slice branch to the task, only if the slice branch has not moved since step 1.

  Then the task is done, and its branch and worktree are removed.
- If the rebase conflicts or the checks fail, the slice branch stays as it was. The task goes back to its own worker, with the conflicting files. The next review round checks the fix.
- Only `flow run` merges. Every other `flow` command only records an event, under a short file lock.
- More than one worker needs checks that are safe to run at the same time (`flow-setup` checks this).
- Bootstrap mode has no task branches. Tasks run one at a time, in the slice worktree.

## Findings

```
open → fixed | wontfix | deferred (fixer) → verified | reopened (reviewer)
```

The fixer is the worker for a task review, the lead for a plan review, and the fix task's worker for an integration review. Every change is recorded by a `flow` command; the Findings list in `status.md` is generated from it, never edited.

- Closed means verified: the reviewer accepted the fix, the wontfix, or the deferral.
- A deferred finding stays visible ("deferred → task 05") until that task is done and its review, or the integration review, verifies it.
- Low findings: marking them is optional. When the slice is done, unmarked lows close as "left as is" and are listed once in the slice's final summary. Lows titled "later" go to `backlog.md`.

- Only a human can mark a high finding wontfix.
- A high finding can be deferred only to a later task in the same slice that is not done yet (`--task <NN>`). Deferring it out of the slice (backlog, a later slice) needs the human, like wontfix.
- A task cannot be done while it has an open or reopened high finding.
- Integration findings become new fix tasks.

## Checks

Checks are automatic commands that prove the code is in good shape. Each project lists its own commands per part in the stack profile (`AGENTS.md` section 9). The usual kinds:

| Kind | Proves | Go example | Web example |
|---|---|---|---|
| format | code is formatted the standard way | gofmt | prettier |
| lint | no common mistakes | go vet | eslint |
| types | types are correct | (part of build) | tsc |
| test | the tests pass | go test | vitest |
| build | the code builds | go build | vite build |

Checks have two tiers, set per check in the stack profile:

| Tier | Runs | Typical checks |
|---|---|---|
| task | every `flow check`, after each task | format, lint, types, unit tests, database tests |
| slice | once all tasks are done, before the integration review | end-to-end and browser tests |

Checks are self-contained:

- They use only tools installed on the machine. Never a shared service: no dev database, no shared cache, no real external API, not even its sandbox.
- Each test dependency gets a local stand-in that starts fresh and is thrown away: a temporary container, an embedded version, or a fake. The stack profile lists them under "Test dependencies".
- Tests refuse to run against an address that is not local or a temporary container.
- Only test keys and fake credentials. Workers never get real secrets.

Running checks:

- `flow check` first runs the auto-fix commands (stack profile) for the parts a task touched, then the task checks, and records the result. `flow check --slice` runs the task checks for every part the slice touched, then the slice checks. `flow check --base` runs the task checks on the base branch, without auto-fix.
- A task moves to review or done only after `flow check` passes. While working you may run single tests. Only `flow check` counts.
- If the checks already fail on the base branch, report it. It does not count against the task.
- Every part a task touches must have checks before coding starts. If a slice adds a new part, its first task sets up the checks for it.
- A part whose folders (stack profile) do not exist yet has nothing to check: its checks are skipped. The task that creates the part makes its checks pass.
