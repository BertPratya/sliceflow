# Manual flow

flow-cli is built with its own workflow. Until flow's first version exists,
the lead does by hand every step flow will do later: it keeps the state,
checks the documents, builds each worker's brief, and creates the approval
tags. The documents are already in flow's format (`templates/`), so when
flow is ready it takes over the same files.

This guide is for the lead. The other roles have their own files in
`workflow/roles/`. The design behind every rule here is in
`docs/flow-design.md` (the notes marked "decided, the human").

## Once per clone

- Git 2.36 or newer (`git --version`). Older git prints wrong paths for
  `git rev-parse --path-format=absolute` without an error.
- `git config push.followTags true`, so a plain push also sends the
  approval tags.
- `git config core.hooksPath .githooks`, to turn on the pre-commit checks.

## A slice

A slice is one feature, on its own branch, in its own folder.

- **Name:** the slice name only, e.g. `due-dates`. Lowercase letters,
  digits, and hyphens. A name is never used again, even after a slice is
  cancelled. Not allowed, because helper folders use them: `base-check`,
  `briefs`, `messages`, and names starting with `task-` or `update-`.
- **Branch:** `slice/<name>`, from the base branch (also when stacked).
- **Worktree:** `<repo root>/.worktrees/<name>/`. `<repo root>` is the main
  checkout's folder: `git rev-parse --path-format=absolute --git-common-dir`,
  without the last `/.git`. Always use full paths, so a worktree never ends
  up inside another one.
- **Folder:** `slices/<name>/`:

```
slices/<name>/
  spec.md       what the feature does            the human approves
  plan.md       how it will be built             the human approves, with tasks.md and preview.md
  tasks.md      every task: structure and content
  preview.md    how the human tries it
  status.md     the only state of the slice      the lead keeps it
  reviews/      review files
  notes.md      optional: technical ideas from the spec talk; not approved
```

**Templates:** `templates/` holds one template per document. Each
template's comments hold its rules. Before you write or change a document,
read its template again: a filled document no longer has the comments.

## Status and commits

- `status.md` follows `templates/status.md`. Only the lead changes it, and
  only in the slice worktree, on the slice branch. It is the only place
  where status lives: never in a document's header.
- After every step that changes something: update `status.md`, add a Log
  line (time in UTC), and commit.
- Commit messages: `<name>: <what happened>`. Task work: `<name>/NN: <summary>`.
- A commit message that holds quoted words (the human's, a worker's notes)
  is written to a file under `<repo root>/.worktrees/messages/` and
  committed with `git commit -F <file>`: quotes and `$` would change on the
  command line.
- **Phase changes**, each with a Log line:

| Phase | Set when |
|---|---|
| `spec` | the slice is created |
| `plan` | the spec is approved |
| `ready` | the plan is approved |
| `implementing` | the first worker starts, and again for any fix task |
| `integration` | every task is done or cancelled |
| `preview` | the integration review passes |
| `done` | the human approved the result; set in the last commit before the merge |

  **Marks** go in brackets after the phase, and keep it. The slice gets
  `(blocked)` only when a plan or integration review fails 3 rounds, or the
  human stops the whole slice, e.g. `plan (blocked)`. A single blocked task
  is marked in the Tasks table, and other tasks go on. The slice gets
  `(cancelled)` when the human cancels it, e.g. `implementing (cancelled)`.

## 1. Start a slice

1. **Is it a slice?** Check `AGENTS.md` section 1. A small change is not a
   slice. An idea bigger than about 25-30 requirements is more than one.
   A vague idea is explored first; ideas that are not built go to
   `backlog.md`.
2. **Stacked?** List the slices not merged yet
   (`git branch --list "slice/*" --no-merged <base branch>`). If the new
   slice needs one of them, it waits for it: `status.md` gets an `After:`
   line. The human confirms.
3. **Create it:** `git worktree add <repo root>/.worktrees/<name> -b slice/<name> <base branch>`.
   In that worktree: `slices/<name>/spec.md` from `templates/spec.md` (fill
   `slice`, `title`, `created`), and `status.md` from `templates/status.md`
   (phase `spec`). Commit: `<name>: new slice`.

## 2. Spec

1. **Learn the facts yourself:** the code, earlier specs and their
   decisions, `backlog.md`. Never ask the human what the code can answer.
2. **Talk with the human:** one question at a time. Each one says what is
   happening, why it matters, the options, and your recommendation.
   Cover: who it is for and why; the main path; each failure; limits;
   empty states; what changes for existing users; security; what is out of
   scope. Technical ideas go to `notes.md`. Update `spec.md` as you go.
3. **Write the spec** by the template's rules. The ones most often missed:
   exact values in requirements; a scenario for every requirement a user
   can see; a `Rule:` line on every decision; assumptions are facts, not
   choices.
4. **Self-check:** the rules in the template's comments, then read the spec
   as a worker would: only what is written.
5. **Spec check:** a review of kind `spec` (see Reviews). Hold its questions
   until the human has read the spec. Answer facts yourself in one line;
   bring decisions to the human, one at a time. Write the answers in, and
   check again. If answers added requirements, the spec check runs again on
   those only.
6. **Approval**, only right after the human says so:
   1. `status.md`: spec `approved`, revision `R<next>` (R001, R002, ...).
      Commit: `<name>: spec R<next> approved ("<the human's words>")`.
   2. Tag that commit (see Approval tags): `flow/<name>/spec/R<next>`.
   3. Create `preview.md` from `templates/preview.md`, with its fixed
      checks: one for the first scenario of each story that is not removed,
      one per `manual` requirement, one per success criterion. Phase
      `plan`. Commit.

## 3. Plan

1. **Learn the code** the slice touches. For a stacked slice, also read the
   earlier slice's approved spec and plan. Settle hard decisions with the
   human, one at a time.
2. **Write `plan.md`** by its template, until its Open questions say None.
   Project files in the plan's `touches` (contracts, the data model) are
   edited on the slice branch now, and listed in `status.md` → Documents.
3. **Write `tasks.md`** by its template, and fill `preview.md`'s steps.
4. **Self-check**, as flow's validator later will:
   - every requirement is in a task's `covers`, and in a flow or shared rule
   - every covered requirement is in a `Proves:` line, except one whose
     Check is only `manual`
   - every scenario that is not removed has an e2e test with a `Scenario:`
     line, and that scenario is in the task's brief
   - every success criterion has a line in Approach
   - every SD has a `Rule:` line naming a real requirement, or `none`
   - every TD is in some task's `follows`, or has `Plan only:`
   - every `follows` name exists in the plan
   - no dependency loops; a task that uses a name another task creates
     depends on it
   - tasks that can run at the same time touch no file in common
   - `preview.md` has every fixed check
   - each task's brief (see Briefs) is under about 1,500 words; if not,
     the reason goes in the plan review prompt
   - no `{{` and no template comment is left
5. **Plan review:** a review of kind `plan`. You fix its findings; take
   decisions to the human. Record each finding's state in `status.md`.
   Submit again until it passes. After 3 rounds without a pass, the slice
   is blocked: ask the human.
6. **Approval**, only right after the human says so: `plan.md`, `tasks.md`,
   `preview.md`, and the project files in `touches` approved together,
   revision `R<next>`. Commit, then tag: `flow/<name>/plan/R<next>`. Phase
   `ready`.

## 4. Implement

**Ready:** the spec and plan are approved; the slice worktree has no
uncommitted changes; a stacked slice's earlier slice is merged, and this
slice is updated from the base branch (as in 5.1); and the base checks
pass: `git worktree add --detach <repo root>/.worktrees/base-check <base branch>`,
run the setup commands and all checks there, task and slice tier (`AGENTS.md` section 9), then
`git worktree remove` it. Never use the main checkout: it is the human's.

**Run the tasks.** A task can start when every task in its `depends_on` is
merged into the slice branch, and its `touches` overlap no task that is not
merged yet. Several tasks may run at once. Set the phase to `implementing`
when the first one starts. For each task:

1. **Worktree:** `git worktree add <repo root>/.worktrees/task-<name>-NN -b task/<name>-NN slice/<name>`.
   If `task/<name>-NN` already exists (an interrupted task), reuse its
   worktree; if the folder is gone, run `git worktree prune`, then
   `git worktree add <repo root>/.worktrees/task-<name>-NN task/<name>-NN`.
   Run the setup commands there (`AGENTS.md` section 9).
2. **Brief:** write the task's brief to `<repo root>/.worktrees/briefs/<name>-NN.md`
   (see Briefs).
3. **Worker:** start it in a fresh context, as `AGENTS.md` section 9 says,
   with this prompt (write it to a file under `<repo root>/.worktrees/briefs/`, never as one quoted argument:
   test output holds quotes and `$`):

       You are the worker. Read workflow/roles/worker.md and follow it.
       Work only in <task worktree, full path>. All paths are relative to it.
       Your brief: <brief file, full path> (round <reviews so far + 1>).
       <after a review: Fix the findings in <review file> first.>
       <checks failed: The checks failed. Last lines: <up to 40 lines>.>
       <after the first worker: Your previous attempt: <summary>; Assumed: <note>; changed files: <list>.>
       <a finding deferred here: Also fix <review>/F<n>, deferred to this task.>
       <merge conflict: Resolve the conflicts in <files>, keeping the changes from both sides.
        Lock files and generated code: run the install or generate command instead.>
       <unblocked: Note from the lead: <answer, and where it is written>.>

   Read only its `RESULT` block, never its whole conversation. A high
   finding marked wontfix, or deferred without naming a later task, blocks
   the task until the human decides.
4. **Checks**, in the task worktree: the auto-fix commands, then the
   task-tier checks (`AGENTS.md` section 9). Slice-tier checks run at
   integration. Each failed run adds 1 to attempt;
   start a new worker with the last 40 lines. After 3 failures since the
   last review, the task is blocked. Attempt goes back to 0 after a review
   or an unblock.
5. **Scope:** the changed files are inside `touches`, plus the files the
   stack profile always allows (`AGENTS.md` section 9), and every test named
   in Done when exists in them. If not, the task is blocked.
6. **Commit** on the task branch: `<name>/NN: <summary>`, with the worker's
   "Assumed" note and finding lines in the body.
7. **Review**, when the task's risk is `core` or `normal`: a review of kind
   `task`. `changes`: a new worker round. `needs_human`: ask the human;
   other tasks go on. If the answer means the code must change, start a
   worker round with the answer in the prompt; if it means the code is
   fine, write "accepted by the human ("<their words>")" in the Log and
   merge. `pass`: merge.
8. **Merge** into the slice branch, one task at a time:
   1. In the task worktree: `git merge slice/<name>`, to bring in the tasks
      merged meanwhile. On a conflict, leave the merge in progress and
      start a worker with the merge-conflict line; then `git add` the files
      and commit, and add a Log line naming the merge commit (the
      integration review checks it). A conflict in a file from the plan's
      `touches` is yours: resolve it, and the human approves that change.
   2. Run the task-tier checks again in the task worktree. On a failure, a worker
      fixes it with the checks-failed line (within `touches`); commit, and
      run the checks again.
   3. In the slice worktree: `git merge --ff-only task/<name>-NN`.
   4. Stop anything running in the task worktree and leave the folder (on
      Windows, an open process locks it). Then, in the slice worktree:
      `git worktree remove <repo root>/.worktrees/task-<name>-NN` and
      `git branch -d task/<name>-NN` (it only counts as merged from the
      slice branch). The task is `done`.

   While 8.1 to 8.3 run, you may edit `status.md`, but don't commit it;
   commit it after 8.3. If `--ff-only` is refused because the slice branch
   moved, repeat 8.1 and 8.2. After 2 failed merges in a row, the task is
   blocked: its `touches` or `depends_on` was likely wrong.

**Stops:**

| What happened | Do |
|---|---|
| A worker asks; the answer is in an approved document | Unblock it with the answer and where it is written. |
| A worker asks something new | Ask the human. Only they unblock it. |
| A worker needs a file outside `touches` | See "Changes to an approved tasks.md". |
| Checks failed 3 times | If you see the cause, unblock with a hint. If not, ask the human. |
| 3 review rounds without a pass | Ask the human. |

**Interrupted:** commit what is there on the task branch as
`<name>/NN: interrupted`; the task goes back to `todo`. If the pre-commit
hook refuses the unfinished code, leave the work uncommitted in the task
worktree instead. Never skip the hook.

## 5. Integration

When every task is done or cancelled, set the phase to `integration`.

1. **Update from the base branch:**
   1. `git worktree add <repo root>/.worktrees/update-<name> -b update/<name> slice/<name>`.
   2. There: `git merge <base branch>`. On a conflict, leave the merge in
      progress and start a worker in the update worktree with only the
      merge-conflict line (there is no brief); then `git add` and commit,
      with a Log line naming the merge commit. A conflict in a file from
      the plan's `touches` is yours, and the human approves that change.
   3. There: run the setup commands and all checks, task and slice tier. On a failure: if the
      merge brought changes in, a worker fixes it with the checks-failed
      line, then commit and check again; if it brought nothing in, the
      slice itself is broken: remove the update worktree and its branch
      (`git branch -D update/<name>`), and add a fix task.
   4. In the slice worktree: `git merge --ff-only update/<name>`. Then,
      still there, remove the update worktree and branch. As in 4.8, don't commit `status.md`
      during 5.1.1 to 5.1.4.
2. **Integration review:** a review of kind `integration`. Each high or
   medium finding becomes a fix task (see "Changes to an approved
   tasks.md"). Then phase `implementing`, run the fix tasks as in step 4,
   and when every task is done or cancelled, come back to step 5. When the review
   passes, set the phase to `preview`.

## 6. Preview

Run the setup commands in the slice worktree. Give the human:

1. what the slice does now, in two or three sentences
2. how to run it (`AGENTS.md` section 9 → Run locally), in the slice worktree
3. the checks in `preview.md`
4. what was parked for later (the plan's Parked items)

Write each result in `status.md` → Preview. A failed check becomes a fix
task: back to step 4, then step 5 again. Something new is a spec change.

## 7. Finish

Only after the human approves the result:

1. Record the approval: a Log line with their words.
2. Update from the base branch again, as in 5.1. If anything came in, the
   integration review runs again (step 5.2), and the human previews again.
3. Copy into `backlog.md`, by `templates/backlog.md`: the plan's Parked
   items, the spec's Out of scope items marked `Goes to: the backlog`, and
   every finding deferred out of the slice (its title). Each item is one
   line, ending with its source and the date.
4. Phase `done`. Commit.
5. In the main checkout: check it is on the base branch with no uncommitted
   changes (if not, ask the human). Merge with a merge commit:
   `git merge --no-ff slice/<name>`. Leave the slice worktree and remove it.
   Push as the git policy says.

## Briefs

A worker sees only its brief. Build it from the links the documents hold,
never from memory:

1. the task's section of `tasks.md`, from its `## Task NN:` heading to the
   next one
2. the exact text of each requirement in its `covers`, from the approved
   spec
3. each scenario whose `Shows:` names one of those requirements
4. the spec's Key concepts
5. the exact text of each plan item in its `follows`: a TD, a shared rule
   or an interface by its heading, or the Testing section
6. a last line: the full spec and plan are at `slices/<name>/`, to open if
   something is unclear

## Approval tags

Every approval is an annotated tag on the commit that records it. The tag's
message is the approval record: what was approved, the revision, the date,
who, and the human's own words. Write the message to a file under
`<repo root>/.worktrees/messages/`, never as a quoted argument (quotes and `$` in
the words would change), then:

    git tag -a flow/<name>/<doc>/R<next> -F <message file>

`<doc>` is `spec` or `plan`. A tag is never moved or deleted.

## Reviews

| Kind | When | Role | Target |
|---|---|---|---|
| `spec` | the spec draft is ready | plan reviewer | the spec revision it would become, e.g. `R001` |
| `plan` | plan, tasks, and preview are written | plan reviewer | the plan revision it would become |
| `task` | a `core` or `normal` task passed its checks | code reviewer | the task id and its commits |
| `integration` | every task is done or cancelled | code reviewer | the slice |

1. **The file:** in `slices/<name>/reviews/`, from `templates/review.md`:
   `spec-R001-<round>.md`, `plan-R001-<round>.md`, `task-NN-<round>.md`, or
   `integration-<round>.md`. A task review's commits go only in the front
   matter's `target`.
   `<round>` is 1 plus the number of earlier reviews of the same kind and
   target. Fill the front matter. A task review file is created and
   committed in the task worktree, on the task branch; the others in the
   slice worktree.
2. **The reviewer:** a fresh context, never your own, as `AGENTS.md`
   section 9 says, with this prompt:

       You are the <role>. Read workflow/roles/<role>.md and follow it.
       Work only in <worktree, full path>. All paths are relative to it.
       This is a <kind> review. Your review file: <path>.
       <task review: Review commits <hashes> of task NN. The task's brief: <brief file>.
        Round 2 and later: the diff since the last round is <last reviewed commit>..HEAD.
        Also check <review>/F<n>, deferred to this task.>
       <plan review: Briefs over 1,500 words, with the reason: <list, or none>.
        Stacked on: <earlier slice>; read its approved spec and plan.>
       <integration review: Review the diff of slice/<name> against <base branch>.
        Merges with conflicts: <commits, from the Log>.
        Check that git diff flow/<name>/plan/R<n> -- slices/<name>/tasks.md holds only
        fix tasks at the end and wider touches.>
       <a changed document: Only <R…, S…, TD…> changed. Check them and what they affect.>

   If it ends without `RESULT: submitted`, start it once more. If that
   fails too, tell the human.
3. **The verdict**, from the review file: anything under Needs human
   decision → `needs_human`; else any high or medium finding, new or
   reopened → `changes`; else `pass`. Record it, and each finding's state,
   in `status.md` (on the slice branch). Commit the review file where it
   was created, and `status.md` on the slice branch:
   `<name>: review <review> (<verdict>)`. For a task review, don't commit
   `status.md` while 4.8.1 to 4.8.3 run.
4. A submitted review is never edited.

## Changes to an approved tasks.md

After approval, `tasks.md` may change in only two ways without a new
approval (decided, the human, 2026-10-20):

1. **Fix tasks**, added at the end with the next free id, covering only
   approved requirements.
2. **A wider `touches`**, when a worker needs one more file, and the path
   stays inside the project's parts, is not in the plan's `touches`,
   overlaps no task that is not merged yet, and adds no behavior.

Each such change gets a Log line, and you tell the human in one line.
`tasks.md` stays `approved`. `git diff flow/<name>/plan/R<n> -- slices/<name>/tasks.md`
shows every change since the approval, and the integration review checks
that it holds only these two kinds. Any other change is a plan change (see
below).

## Changing an approved spec or plan

Only when the human agrees. Then:

1. **Write the reason** in one or two sentences. Set the document to
   `changed` in `status.md`.
2. **Change only what must change**, after reading its template again.
   Numbers are never reused: a removed requirement or scenario stays with
   "(removed in R…: why)"; a replaced decision stays, marked replaced.
3. **The plan follows the spec:** if the spec changed, the plan becomes
   `changed` too. New scenarios, `manual` requirements, or success criteria
   get new preview checks (`templates/preview.md`). For tasks:
   - `done` tasks stay done; a problem in their code becomes a fix task
   - a task no longer needed: `cancelled` in `status.md`, with a Log line;
     remove its worktree, and keep its branch (as Cancel does)
   - a new task takes the next free id
   - a task not started yet: update it to the new plan
   - a task with a branch: update it; after the approval, run
     `git merge slice/<name>` in its worktree, then a new worker round
4. **Review again**, only what changed (the prompt names it). A
   wording-only change (same meaning, no requirement, scenario, or
   decision added, changed, or removed) skips the review.
5. **Show the human:** the reason, the diff against the last approval
   (spec: `git diff flow/<name>/spec/R<n> -- slices/<name>/spec.md`;
   plan: `git diff flow/<name>/plan/R<n> -- slices/<name>/plan.md
   slices/<name>/tasks.md slices/<name>/preview.md <touches files>`), what was
   added, changed, and removed, and the tasks it affects.
6. **Approve again**, only right after the human says so: the next revision
   and a new tag.

## Cancel and reopen

**Cancel**, only when the human says so:

1. Stop every worker. Commit each task's work as `<name>/NN: interrupted`
   (as in step 4, Interrupted).
2. Add the mark: e.g. `implementing (cancelled)`, with a Log line holding
   the human's reason, in their words. Commit.
3. Remove the worktrees: each task's, then the slice's. Keep a task
   worktree that holds uncommitted work (never `--force`), with a Log line
   saying so; reopen reuses it. Keep everything else: the slice branch, the
   task branches, the tags, the reviews.
4. A slice stacked on this one: tell the human it waits on a cancelled
   slice. They choose: move it onto the base branch (maybe with a spec or
   plan change), cancel it too, or leave it waiting.

The name is never used again.

**Reopen**, only when the human says so: make the slice worktree again
(`git worktree add <repo root>/.worktrees/<name> slice/<name>`), remove the
mark, and go on from the phase. Interrupted tasks are `todo`, and their
work is still on their branches.
