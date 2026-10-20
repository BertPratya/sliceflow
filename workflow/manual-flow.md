# Manual flow

flow-cli is built with its own workflow. Until flow's first version exists,
the lead does by hand every step flow will do later: it keeps the state,
checks the documents, builds each worker's brief, and creates the approval
tags. The documents are already in flow's format (`template/`), so when
flow is ready it takes over the same files.

This guide is for the lead. The other roles have their own files in
`workflow/roles/`. The design behind every rule here is in
`docs/flow-design.md` (the notes marked "decided, the human").

## A slice

A slice is one feature, on its own branch, in its own folder.

- **Name:** the slice name only, e.g. `due-dates`. Lowercase letters,
  digits, and hyphens. A name is never used again. (This repo's slices made
  before 2026-10-20, like `0001-slice-and-events`, keep their numbers.)
- **Branch:** `slice/<name>`, from the base branch.
- **Worktree:** `<repo root>/.worktrees/<name>/`. `<repo root>` is the main
  checkout's folder: `git rev-parse --path-format=absolute --git-common-dir`,
  without the last `/.git`. Always use full paths.
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

**Templates:** `template/` holds one template per document. Each template's
comments hold its rules. Before you write or change a document, read its
template again: a filled document no longer has the comments.

## Status and commits

- `status.md` follows `template/status.md`. Only the lead changes it. It is
  the only place where status lives: never in a document's header.
- After every step that changes something: update `status.md`, add a Log
  line (time in UTC), and commit.
- Commit messages: `<name>: <what happened>`. Task work: `<name>/NN: <summary>`.

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
   In that worktree: `slices/<name>/spec.md` from `template/spec.md` (fill
   `slice`, `title`, `created`), and `status.md` from `template/status.md`
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
   2. Tag that commit, annotated, with the approval record:
      `git tag -a flow/<name>/spec/R<next> -m "spec R<next> of <name> approved by <who>, <date>. Their words: \"...\""`.
   3. Create `preview.md` from `template/preview.md`, with its fixed checks:
      one for the first scenario of each story that is not removed, one per
      `manual` requirement, one per success criterion. Set the phase to
      `plan`. Commit.

Once per clone: `git config push.followTags true`, so a plain push also
sends the approval tags.

## 3. Plan

1. **Learn the code** the slice touches. Settle hard decisions with the
   human, one at a time.
2. **Write `plan.md`** by its template, until its Open questions say None.
3. **Write `tasks.md`** by its template, and fill `preview.md`'s steps.
4. **Self-check**, as flow's validator later will:
   - every requirement is in a task's `covers`, and in a flow or shared rule
   - every scenario that is not removed has an e2e test with its `Scenario:` line
   - every SD has a `Rule:` line naming a real requirement, or `none`
   - every TD is in some task's `follows`, or has `Plan only:`
   - every `follows` name exists in the plan
   - a task that uses a name another task creates depends on it
   - tasks that can run at the same time touch no file in common
   - each task's brief (see Briefs) is under about 1,500 words, or the
     reason is noted for the reviewer
   - no `{{` and no template comment is left
5. **Plan review:** a review of kind `plan`. You fix its findings; take
   decisions to the human. Record each finding's state in `status.md`.
   Submit again until it passes. After 3 rounds without a pass, ask the
   human.
6. **Approval**, only right after the human says so: `plan.md`, `tasks.md`,
   and `preview.md` approved together, revision `R<next>`. Commit, then tag:
   `flow/<name>/plan/R<next>`, annotated, as for the spec. Phase `ready`.

## 4. Implement

**Ready:** the spec and plan are approved, the worktree is clean, a stacked
slice's earlier slice is merged, and the base checks pass in a temporary
worktree of the base branch (never the main checkout).

**Run the tasks.** A task can start when every task in its `depends_on` is
merged into the slice branch, and its `touches` overlap no task that is not
merged yet. Several tasks may run at once. For each task:

1. **Worktree:** `git worktree add <repo root>/.worktrees/task-<name>-NN -b task/<name>-NN slice/<name>`.
2. **Brief:** write the task's brief to a file under `<repo root>/.worktrees/`
   (see Briefs).
3. **Worker:** start it in a fresh context, as `AGENTS.md` section 9 says,
   with this prompt:

       You are the worker. Read workflow/roles/worker.md and follow it.
       Work only in <task worktree, full path>. All paths are relative to it.
       Your brief: <brief file, full path> (round <reviews so far + 1>).
       <after a review: Fix the findings in <review file> first.>
       <checks failed: The checks failed. Last lines: <up to 40 lines>.>
       <after the first worker: Your previous attempt: <summary>; Assumed: <note>; changed files: <list>.>
       <unblocked: Note from the lead: <answer, and where it is written>.>

   Read only its `RESULT` block, never its whole conversation.
4. **Checks**, in the task worktree: the auto-fix commands, then the
   checks (`AGENTS.md` section 9). On a failure, add 1 to attempt, and
   start a new worker with the last 40 lines. After 3 failures since the
   last review, the task is blocked.
5. **Scope:** the changed files are inside `touches`, and every test named
   in Done when exists in them. If not, the task is blocked.
6. **Commit** on the task branch: `<name>/NN: <summary>`, with the worker's
   "Assumed" note and finding lines in the body.
7. **Review**, when the task's risk is `core` or `normal`: a review of kind
   `task`. `changes`: a new worker round. `needs_human`: ask the human.
   `pass`: merge.
8. **Merge** into the slice branch, one task at a time: in the task
   worktree `git merge slice/<name>` (a conflict: a worker resolves it),
   run the checks again, then in the slice worktree
   `git merge --ff-only task/<name>-NN`. Remove the task worktree and
   branch. The task is `done`.

**Stops:**

| What happened | Do |
|---|---|
| A worker asks; the answer is in an approved document | Unblock it with the answer and where it is written. |
| A worker asks something new | Ask the human. Only they unblock it. |
| A worker needs a file outside `touches` | Widen it only if it stays inside the project, overlaps no unmerged task, and adds no behavior; tell the human. Otherwise it is a plan change. |
| Checks failed 3 times | If you see the cause, unblock with a hint. If not, ask the human. |
| 3 review rounds without a pass | Ask the human. |

**Interrupted:** commit what is there on the task branch as
`<name>/NN: interrupted`; the task goes back to `todo`.

## 5. Integration

1. **Update from the base branch** in an update worktree
   (`update/<name>` from the slice branch), run the checks, then
   `git merge --ff-only update/<name>` in the slice worktree.
2. **Integration review:** a review of kind `integration`. Each high or
   medium finding becomes a fix task.

**Fix tasks:** add a new `## Task NN` section at the end of `tasks.md`
(the next free id), with a Log line. It needs no new approval when it
covers only approved requirements. New behavior is a spec change, not a
fix task.

## 6. Preview

Run the setup commands in the slice worktree. Give the human: what the
slice does now, how to run it (`AGENTS.md` section 9, Run locally), and the
checks in `preview.md`. Write each result in `status.md` → Preview. A
failed check becomes a fix task. Something new is a spec change.

## 7. Finish

Only after the human approves the result:

1. Update from the base branch again, as in step 5.1.
2. Copy into `backlog.md`, by `template/backlog.md`: the plan's Parked
   items, the spec's Out of scope items marked `Goes to: the backlog`, and
   every finding deferred out of the slice (its title). Each line ends with
   its source and the date.
3. Phase `done`. Commit.
4. In the main checkout (on the base branch, clean): merge with a merge
   commit, `git merge --no-ff slice/<name>`. Remove the slice worktree.
   Push as the git policy says (`push.followTags` sends the tags).

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

## Reviews

| Kind | When | Role |
|---|---|---|
| `spec` | the spec draft is ready | plan reviewer |
| `plan` | plan, tasks, and preview are written | plan reviewer |
| `task` | a `core` or `normal` task passed its checks | code reviewer |
| `integration` | every task is done | code reviewer |

1. **The file:** `reviews/<kind>-<target>-<round>.md` from
   `template/review.md` (integration: `reviews/integration-<round>.md`).
   Fill the front matter. A task review lives on the task branch; the
   others on the slice branch.
2. **The reviewer:** a fresh context, never your own, as `AGENTS.md`
   section 9 says:

       You are the <role>. Read workflow/roles/<role>.md and follow it.
       Work only in <worktree, full path>. All paths are relative to it.
       This is a <kind> review. Your review file: <path>.
       <task review: Review commits <hashes> of task NN. The task's brief: <brief file>.>
       <integration review: Review the diff of slice/<name> against <base branch>.>
       <a changed spec: Only R<n>, R<n> changed. Check them and what they affect.>

3. **The verdict**, from the review file: anything under Needs human
   decision → `needs_human`; else any high or medium finding, new or
   reopened → `changes`; else `pass`. Record it, and each finding's state,
   in `status.md`. Commit: `<name>: review <review> (<verdict>)`.
4. A submitted review is never edited.

## Changing an approved document

Never change an approved spec or plan without the human. When they agree:
edit it (read its template again first), set it to `changed` in
`status.md`, review it again (only what changed), and approve it again with
the next revision and a new tag. Requirement and scenario numbers are never
reused. A plan approved before the spec changed becomes `changed` too.
