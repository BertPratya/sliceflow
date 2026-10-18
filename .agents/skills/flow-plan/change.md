# flow-plan: changing an approved plan

Use this when the plan is already approved and something must change:
the spec changed, a worker or reviewer found the plan wrong, the human
changed a decision, or the slice this one is stacked on changed its plan
or contract. A fix task that only fixes existing requirements does
not need this (`workflow/formats/lifecycle.md` → Documents).

1. **Write down the reason** in one or two sentences. It goes with the approval.
   Set the plan's status to `changed` in `status.md`.
2. **Change only what must change.** Settle new hard decisions with the
   human, as in step 3 of `SKILL.md`.
3. **Edit the plan and the contracts.** Keep TD numbers stable: a replaced
   decision stays, marked "(replaced by TD9)".
4. **Update the tasks:**
   - `done` tasks stay done. A problem in their code becomes a new fix task.
   - A task that is no longer needed: set it to `cancelled` in `status.md`.
   - A new task takes the next free number.
   - A stale task that is not started yet: update it to the new plan.
   - A stale task that already has a branch: update its task file; after the
     approval, run `git merge <slice branch>` in its worktree, then a new
     worker round.
   - A cancelled task that has a branch: `git worktree remove` its worktree,
     `git branch -D` its branch, and add a Log line.
5. **Run the document checks** (`workflow/formats/document-checks.md`). Fix every error.
6. **Plan review**, as in step 9 of `SKILL.md`. On a small change, the
   reviewer checks only what changed and what it affects.
7. **Show the human:** the reason, the diff (`git diff --no-index history/plan/<last revision>.md plan.md`), the tasks
   added, changed, and cancelled, and the contract changes.
8. **The human approves**, as in step 10 of `SKILL.md`. Only when they
   have just told you to.

Done when: the new revision is approved and every task agrees with it.
