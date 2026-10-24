---
review: plan-1
slice: 0001-slice-and-events
kind: plan
target: docs/slices/0001-slice-and-events/plan.md
round: 1
reviewer: doc-reviewer (claude opus)
---

# Review

## Summary

The plan, tasks, and preview cover every requirement and every scenario,
and the spec changes since 23a6e13 are consistent with each other. Two
points need the human: what bare and incomplete commands do, and R18's
reason when git prints nothing (Q1, Q2). The main fixes for the lead:
the Core API names `Status` and `List` twice and will not compile (F1),
tasks 02 and 05 cannot run their tests before task 01 (F2), and the
interrupt path has gaps that can hang or fail the undo (F3).

## Checked

- Files read: docs/slices/0001-slice-and-events/spec.md, plan.md,
  tasks.md, preview.md, status.md, docs/templates/plan.md, tasks.md,
  preview.md, review.md, spec.md, AGENTS.md, .gitattributes,
  .github/workflows (checks), .agents/prompts/doc-reviewer.md
- Ran:
  git diff 23a6e13 HEAD -- docs/slices/0001-slice-and-events/spec.md (R18, R26, R28, R36, R37, R42, R43, S29, Out of scope changed)
  ls cmd internal (no Go code yet: plan touches only new files and go.mod)
  git commit with a pre-commit hook that only runs exit 1, in a scratch repo (exit 1, stderr empty: 0 bytes)
  wc -w on spec.md, plan.md, tasks.md and the task and plan sections (to estimate brief sizes)
- Requirements: R1 to R43, S1 to S29, TD1 to TD8, every shared rule

## Tests

None.

## Previous findings

None.

## Findings

### F1. `Status` and `List` are each a type and a function in one package

Level: medium
Where: plan.md → Interfaces → Core API, tasks.md → Task 09 → Interfaces,
  tasks.md → Task 10 → Interfaces
Breaks: TD7, task 09, task 10
Why: The Core API declares `func Status(...) (Status, error)` and
  `func List(...) (List, error)`, and tasks 09 and 10 create
  `type Status struct` and `type List struct` in `internal/slice`. Go
  does not allow a type and a function with the same name in one
  package: `Status redeclared in this block`. Task 09's worker cannot
  build its own Interfaces as written.
Fix: rename the result types, e.g. `SliceStatus` and `SliceList`, in
  the plan and in tasks 09 and 10. The three plain functions stay.

### F2. Tasks 02 and 05 cannot run their tests before go.mod exists

Level: medium
Where: tasks.md → Task 02 → depends_on, tasks.md → Task 05 → depends_on
Breaks: tasks.md rules (dependencies are complete)
Why: Task 01 creates `go.mod`. Tasks 02 and 05 have `depends_on: []`,
  so they may run first or at the same time. Their step 1 says to run
  the tests and see them fail; without a module, `go test
  ./internal/eventlog/` fails with `go: cannot find main module`, and
  `go.mod` is outside their touches. Their imports also need the module
  path of TD1.
Fix: give tasks 02 and 05 `depends_on: ["01"]`.

### F3. Interrupts: undo with a cancelled ctx, a hook that keeps running, and the reason

Level: medium
Where: plan.md → Shared rules → Interrupts, plan.md → Shared rules →
  Undo on failure, tasks.md → Task 01 → Steps, tasks.md → Task 08 →
  Steps
Breaks: R18
Why: Four gaps a worker would have to guess:
  1. The undo runs `git worktree remove` and `git branch -D` through
  `git.Run`, which uses `exec.CommandContext`. With the cancelled `ctx`
  of the signal, each undo command is killed at once, so after Ctrl-C
  the result is `Could not remove: ...` instead of the clean removal of
  R18. The plan must say the undo runs with a context that is not
  cancelled (e.g. `context.WithoutCancel`).
  2. When `ctx` is cancelled, Go kills `git`, but the hook git started
  (`sh` running `sleep 30`) keeps the stderr pipe open. `cmd.Wait` then
  waits until the hook ends, so `TestCreateInterrupt` prints its error
  only after 30 seconds. Task 01's `Run` needs `cmd.WaitDelay`.
  3. A real Ctrl-C reaches git and the hook too. The failed `git commit`
  may return before sliceflow sees the signal, and its stderr can be
  empty. Task 08 should check whether a signal arrived before it picks
  git's message as the reason.
  4. Ctrl-C during the checks (before anything is made) is not
  described. If `RefExists` treats a killed `git rev-parse` as "not
  found", R16 prints `base branch main does not exist`, which is wrong.
Fix: add the four points to Interrupts, and step 3 of task 01 and step
  3 of task 08.

### F4. The remote branch check matches a branch that only ends in slice/<name>

Level: medium
Where: plan.md → Flow 2, step 5
Breaks: R15
Why: Flow 2 matches any ref under `refs/remotes/` that ends in
  `/slice/<name>`. A teammate's branch `user/slice/spec-submit` gives
  `refs/remotes/origin/user/slice/spec-submit`, which matches, so
  `slice create spec-submit` fails with `already exists` although no
  branch `slice/spec-submit` exists. The `<remote>` in the message is
  also unclear for such a ref.
Fix: read the remote names with `git remote` (it reads only the config,
  no network), and check `refs/remotes/<remote>/slice/<name>` for each.
  `TestCheckExistsOrder` then adds the remote `team/origin` to the
  config.

### F5. git add fails when the project ignores .jsonl or docs files

Level: medium
Where: plan.md → Flow 1, step 8
Breaks: R7, R8
Why: Flow 1 runs `git add -- <the two files>`. In a project whose
  `.gitignore` has `*.jsonl` or `.*`, or a user with such a global
  excludes file, git refuses with `The following paths are ignored`.
  Then every `slice create` in that repo fails with R18's error, and R7
  can never hold there.
Fix: use `git add -f -- <the two files>`; the two paths are fixed and
  sliceflow's own.

### F6. A branch slice/My_Thing exists, but status says no slice named My_Thing

Level: medium
Where: spec.md → Key concepts → A slice exists, plan.md → Flow 4,
  step 4
Breaks: nothing
Why: Key concepts says a slice exists when a local branch
  `slice/<name>` exists. With a branch `slice/My_Thing`, Flow 4 step 4
  (lead's call) prints R33's `no slice named My_Thing`, and R43 leaves
  it out of the list. Both follow the human's choice in R43, but the
  key concept still says the slice exists. A worker reading the key
  concept would show its status.
Fix: in Key concepts, say a slice exists only when its name passes R2,
  and name R43 there.

### F7. TestSpecNoTitle checks the wrong placeholder and an impossible body

Level: medium
Where: tasks.md → Task 05 → Done when → `TestSpecNoTitle`
Breaks: R7
Why: The spec template's front matter uses `{{date}}`, not
  `{{created}}`, so the test passes when `{{date}}` is left in. The
  test also says the body is the same as `docs/templates/spec.md` after
  its front matter, but step 2 removes the rules about the title from
  that body (the template's comment lines 28 and 32 name the title). A
  test written first as stated cannot pass after step 2.
Fix: check that no `{{slice}}`, `{{title}}`, or `{{date}}` is left in
  the front matter, and compare the body with the embedded
  `internal/template/spec.md`, or list the lines that may differ.

### F8. Cross-platform paths does not reach tasks 08 to 10, and the Tests rule reaches no task

Level: medium
Where: tasks.md → Task 08, 09, 10 → follows, plan.md → Shared rules →
  Tests
Breaks: AGENTS.md (paths must work on Windows, macOS, and Linux)
Why: Task 09 builds `<ref>:docs/slices/<name>/.events.jsonl` and task
  10 reads `git ls-tree` paths, but neither follows Cross-platform
  paths, which says git paths use `path`, not `filepath`. A worker who
  writes `filepath.Join` gets `docs\slices\x\.events.jsonl` on Windows,
  `git show` finds no file, and every slice shows `.events.jsonl not
  found`. Task 08 removes files and folders with OS calls, too. The
  shared rule Tests (a project rule) is in no task's follows; the tasks
  follow Testing instead.
Fix: add Cross-platform paths to tasks 08, 09, and 10, and either add
  Tests to the follows or merge it into Testing.

### F9. R42 is proved only for create

Level: medium
Where: tasks.md → Task 09, Task 10 → covers and Done when
Breaks: R42
Why: R42 is about every command with `--json`. Only task 07 covers it,
  with S29. Tasks 09 and 10 do not get R42 in their brief, so a worker
  may print a JSON error for `slice status no-such-slice --json` or
  `slice list --json` outside a repo, and no test would catch it.
Fix: add R42 to tasks 09 and 10, with one e2e case each, e.g. S19 with
  `--json` and S28 with `--json`: stdout empty, the same stderr, exit 1.

### F10. Most briefs are well over 1,500 words

Level: medium
Where: tasks.md → Task 06, Task 07, Task 09
Breaks: the brief size rule of docs/templates/tasks.md
Why: Task 06's own section is about 530 words. Its brief adds Key
  concepts (about 160), Testing (about 270), Flow 1 (about 240), six
  more plan items, 13 requirements, and the scenarios S1 to S4, S12, and
  S26: about 2,700 words. Tasks 07 and 09 are of the same size (task 09
  brings R26 to R33 and S13 to S19, S25, S28). No reason is given.
Fix: trim follows where the item adds nothing to that task (e.g. TD7
  and Core API next to Flow 1 in task 06), or split task 09 (core
  reading, then the command), or write one line in each saying why the
  size is needed.

### F11. TD3 is both followed and Plan only

Level: low
Where: plan.md → Decisions → TD3, tasks.md → Task 01 → follows
Breaks: nothing
Why: The plan template says sliceflow warns about a decision that is
  both followed and plan only. Task 01 follows TD3 and writes `go 1.25`.
Fix: remove the Plan only: line of TD3.

### F12. follows names flows and Data subsections

Level: low
Where: tasks.md → Task 02, 06, 07, 08, 09, 10 → follows
Breaks: nothing
Why: The tasks template allows in follows a decision, a shared rule or
  an interface by its heading, or Testing. Tasks name `Flow 1` to
  `Flow 5` (the headings are e.g. `Flow 1: Create a slice`) and the Data
  subsections `Events in the code` and `Reading a line`. If sliceflow
  matches only those kinds, these names are not found and their text
  does not reach the worker.
Fix: confirm the lead's brief builder takes these, or name them by
  their full headings.

### F13. Some Proves: lines name a requirement the test does not prove

Level: low
Where: tasks.md → Task 01 → Done when, tasks.md → Task 03 → Done when →
  `TestPrintWarnings`
Breaks: nothing
Why: `TestRunReturnsOutput`, `TestRunErrorHasStderr`, `TestRefExists`,
  and `TestShowMissingFile` prove no part of R40; only
  `TestRunSetsNoNetworkEnv` does. `TestPrintWarnings` checks R28's form,
  not R38.
Fix: keep R40 on the env test only, and name what the others check;
  `TestPrintWarnings` proves R28 (add R28 to task 03's covers).

### F14. e2e/ is outside the gofmt check

Level: low
Where: plan.md → Approach → table, plan.md → Testing
Breaks: nothing
Why: The checks run `gofmt -l cmd internal`, and CI does the same. The
  Go files in `e2e/` are never format-checked.
Fix: put the e2e tests in `internal/e2e/`, or ask the human to widen the
  check.

### F15. R26 does not say which reason wins when a line has two faults

Level: low
Where: spec.md → R26, plan.md → Data → Reading a line
Breaks: nothing
Why: A line with an unknown field and a missing `at` gives one warning,
  but which one is not said. Each test row has one fault, so any order
  passes; later slices may then disagree.
Fix: in Reading a line, say the checks run in R26's order of reasons,
  and the first fault found is the warning.

### F16. Flow 1 says Create finds the repo, but Create gets a Repo

Level: low
Where: plan.md → Flow 1, step 2, tasks.md → Task 06 → Steps, step 2
Breaks: nothing
Why: `Create` takes `r Repo`, and `FindRepo` (task 04) makes it, so the
  CLI finds the repo before `Create`. Task 06's step 2 tells the worker
  to do Flow 1 step 2 inside `Create`.
Fix: say the CLI calls `FindRepo` and passes the Repo, in Flow 1 and
  task 06.

### F17. git makes .sliceflow/worktrees/, so the undo list misses it

Level: low
Where: plan.md → Shared rules → Undo on failure, item 1
Breaks: nothing
Why: Flow 1 step 5 makes `.sliceflow/` and the `.gitignore`. The folder
  `.sliceflow/worktrees/` is made by `git worktree add` in step 6, but
  the undo list says each thing is added right after sliceflow makes
  it. `TestCreateHookRefuses` expects no `.sliceflow/` left, so the
  worker must find this alone.
Fix: say `create` notes before step 6 whether `.sliceflow/worktrees/`
  exists, and adds it to the list after step 6 if it did not.

## Requirements check

None.

## Needs human decision

- **Q1.** What do `sliceflow` with no command, `sliceflow slice` with no
  subcommand, `slice create` with no name, and `slice create a b` (two
  names) do? The spec's R38 lists an unknown command or flag and a
  missing value, but not these. Task 03 decides that bare `sliceflow`
  prints the usage and exits with 0, and picks the texts
  `unknown command "slice nope"` and `flag needs a value: --title`,
  with no (lead's call). The user sees all of these.
  Guess: bare `sliceflow` and `sliceflow slice` print the usage on
  stdout and exit 0; a missing name is `Error: missing slice name. Run
  sliceflow --help.` with exit 2; an extra name is
  `Error: too many arguments. Run sliceflow --help.` with exit 2.
- **Q2.** What is R18's `<reason>` when git prints nothing, or more than
  one line? A `pre-commit` hook that only runs `exit 1`, as in S12,
  makes `git commit` exit 1 with an empty stderr (I ran it). The
  message is then
  `Error: could not create slice spec-submit: . Everything sliceflow made was removed.`
  Git's messages also often have several lines (`error:` and `hint:`).
  Guess: with no message, the reason is `git commit failed (exit code 1)`;
  with several lines, only the first line that is not empty is used.
