# Lifecycles

How slice phases, documents, tasks, findings, and checks change status. Every status lives only in the slice's `status.md`, kept by the lead, with a dated Log line for each change. File headers hold no status. Read this before changing any status.

## Slice phases

```
spec → plan → ready → implementing → integration → preview → done
                       any phase → blocked | cancelled
```

The phase changes only when its event happens. The lead sets it at that moment.

Brainstorming is not a phase. Exploring a vague idea, or the whole system, happens before a slice exists. A slice is created when the idea is one clear feature.

| Phase | Starts when | What happens in it |
|---|---|---|
| spec | the slice is created | The interview, the spec, the spec check, the human's approval. |
| plan | the human approves the spec | The plan and tasks, the plan review, the human's approval. |
| ready | the human approves the plan | The lead starts the tasks right away (a stacked slice first waits until the earlier slice is merged). |
| implementing | the first worker starts | Workers build, reviewers check, the lead commits each finished task. |
| integration | all tasks are done | Update from the base branch, slice checks, integration review. |
| preview | the integration review passes | The human tries the slice. |
| done | set in the last commit before the merge into the base branch | Nothing. |

- A fix task (from an integration finding, a slice check that fails on the slice's own code, or the preview) sends the phase back to implementing. When the tasks are done, it moves on again.
- blocked: the whole slice cannot go on, e.g. the plan review fails 3 rounds. When it is solved, the slice goes back to the phase it was in. A blocked task does not block the slice.
- cancelled: only when the human says so.

## Documents

Applies to `spec.md`, `plan.md`, and a slice's change to a contract or other project file.

```
draft → in_review → awaiting_approval → approved
```

The status shows who has the ball: `draft` the lead, `in_review` the reviewer, `awaiting_approval` the human.

- When a review ends, the document status follows the verdict: `pass` → awaiting_approval; `changes` or `needs_human` → back to draft. Low findings alone pass, so they never send a document back. The lead may still fix them before the approval, without a new review.
- Spec, submitted again after answering its spec check: no new requirements → awaiting_approval directly (no new reviewer); new requirements → in_review, and the check covers only the new ones.
- Plan, submitted again after fixing findings: always in_review (the next round verifies the fixes).
- Only an `awaiting_approval` document can be approved, and only right after the human says so. The approval commit quotes the human's words. Rejected: back to draft.
- Edited after approval: the status becomes `changed`, then it is submitted again as above. Every change needs a new approval.
- A contract counts as edited when the slice's change to the file no longer matches the last approved patch in `history/contract/`.
- The plan, its tasks, and the slice's contract changes are approved together, with one approval. Each one is still copied to `history/`.
- A fix task added after the plan is approved (from the integration review, a failing slice check, or the preview) needs no new approval when it covers only requirements in the approved spec and changes no plan, contract, or decision. The lead tells the human about each one in one line. Anything else is a plan change (`flow-plan` → `change.md`).

## Tasks

```
todo → in_progress → in_review → done
```

Tasks run in parallel, each on its own task branch and worktree (see "Branches and worktrees").

- Checks fail: attempt +1, stays in_progress.
- Sent to review: round +1. The round is the number of reviews started, so it matches the review file name (`task-02-2` is round 2).
- Review asks for changes: attempt back to 0, back to in_progress.
- Review needs the human: the human answers. Code must change → back to in_progress, a new round. Code is fine → accepted by the human (Log line, with their words), and the task goes to the merge.
- Simple track, non-core task: goes to the merge without review once its checks pass.
- done means the task is merged into the slice branch.
- A merge fails (a conflict, or checks fail after bringing in the slice branch): back to in_progress, with a new worker. It does not count toward attempt or the stop rules. After 2 failed merges in a row: blocked.
- A stop rule hits: blocked. The lead resolves it when the answer is already in an approved document, or, after failed checks, when it can see the cause; otherwise the human does. Then back to todo or in_progress. Unblocking sets attempt to 0.
- Interrupted: back to todo. The work is committed on its task branch. An interruption is not a failure.
- Removed by a re-plan: cancelled.
- done is final. A problem found later becomes a new fix task.
- "Waiting" (a dependency is not done) and "stale" (based on an older plan revision) are worked out from the task headers (`depends_on`, `based_on`) when the lead updates `status.md`, never set.

## Branches and worktrees

Branch names and the worktree folder are in the git policy (`AGENTS.md` section 10).

- Creating a slice creates its branch and worktree. The slice's documents, status, and finished tasks live there.
- A new slice starts from the base branch, or from another unmerged slice's branch (stacking) when it needs something from that slice: its code, a contract, the data model, or a decision. The lead proposes it; the human confirms. A slice stacks on at most one other slice. If it needs two, write its spec now and plan it after one of them is merged.
- A stacked slice can be specced and planned, but it waits until the earlier slice is merged. Then it updates from the base branch, and its tasks can start.
- If the earlier slice's plan or contract changes after the stacked slice's plan is approved, the stacked plan is stale. The lead checks what is affected; a needed change is a plan change (`flow-plan` → `change.md`).
- If the earlier slice is cancelled, the stacked slice moves onto the base branch and goes back to phase plan.
- A slice that holds only documents (phase spec, plan, or ready) can be moved onto another base: `git rebase --onto <new start> <old start>` in its worktree. After that, no moving.
- When two unmerged slices change the same file from their plans' `touches`, and neither is stacked on the other, the lead tells the human before approval.
- A task starts when every task in its `depends_on` is merged into the slice branch, and its `touches` do not overlap any task whose branch is not merged yet. There is no other limit.
- Each task gets a branch `task/NNNN-NN` from the newest slice branch, in its own worktree. Its checks, commits, and review files happen there. The slice branch is not touched until the task is merged.
- `status.md` is edited and committed only on the slice branch, never on a task branch. So merging a task never conflicts in `status.md`.
- The lead merges one task at a time: bring the slice branch into the task branch (`git merge`), run the checks there, then move the slice branch forward (`git merge --ff-only`). The slice branch only ever moves forward, to code that passed its checks.
- A merge conflict goes to a new worker for that task (the task merged second fixes it). A conflict in a contract file goes to the lead, and the human approves it again. A conflict with the base branch is solved the same way, by a worker in an update worktree (`flow-implement` step 4.1), not by a fix task.

## Findings

```
open → fixed | wontfix | deferred (fixer) → verified | reopened (reviewer)
```

The fixer is the worker for a task review, the lead for a plan review, and the fix task's worker for an integration review. The lead records every change in the Findings list of `status.md`.

- Closed means verified: the reviewer accepted the fix, the wontfix, or the deferral.
- A deferred finding stays visible ("deferred → task 05") until that task is done and its review, or the integration review, verifies it.
- Low findings: marking them is optional. When the slice is done, unmarked lows close as "left as is" and are listed once in the slice's final summary. Lows titled "later" go to `backlog.md`.

- Only a human can mark a high finding wontfix.
- A high finding can be deferred only to a later task in the same slice that is not done yet. Deferring it out of the slice (backlog, a later slice) needs the human, like wontfix.
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
| task | after each task | format, lint, types, unit tests, database tests |
| slice | once all tasks are done, before the integration review | end-to-end and browser tests |

Checks are self-contained:

- They use only tools installed on the machine. Never a shared service: no dev database, no shared cache, no real external API, not even its sandbox.
- Each test dependency gets a local stand-in that starts fresh and is thrown away: a temporary container, an embedded version, or a fake. The stack profile lists them under "Test dependencies".
- Tests refuse to run against an address that is not local or a temporary container.
- Only test keys and fake credentials. Workers never get real secrets.

Running checks:

- "Run the checks" means: the auto-fix commands for the parts a task touched, then the task checks for those parts (stack profile). Slice checks: the task checks for every part the slice touched, then the slice-tier checks. Base checks: the task checks on the base branch, without auto-fix.
- A task moves to review or done only after the checks pass. While working, single tests are fine. Only a full run of the checks counts.
- If the checks already fail on the base branch, report it. It does not count against the task.
- Every part a task touches must have checks before coding starts. If a slice adds a new part, its first task sets up the checks for it.
- A part whose folders (stack profile) do not exist yet has nothing to check: its checks are skipped. The task that creates the part makes its checks pass.
