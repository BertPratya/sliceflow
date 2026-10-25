---
review: plan-1
slice: 0002-spec-and-state
kind: plan
target: docs/slices/0002-spec-and-state/plan.md
round: 1
reviewer: doc-reviewer (claude opus)
---

# Review

## Summary

The plan, tasks, and preview cover every requirement and every scenario,
and the names they use match 0001's real code. The R18 change reads
cleanly and conflicts with nothing. The main problem is that
`saveEvent`'s interface does not fit the flows that call it. It takes the
lock itself, and its data is fixed before the lock, so `spec submit`
cannot run R20's state check, R19, and the strict check under one lock.
It also cannot give R20's, R32's, or R6's errors (F1). One question for
the human: `template show` outside a git repo (Q1).

## Checked

- Files read: docs/slices/0002-spec-and-state/spec.md, plan.md,
  tasks.md, preview.md, status.md, reviews/spec-1.md to spec-3.md (by
  their findings in status.md); docs/templates/plan.md, tasks.md,
  preview.md, review.md; docs/slices/0001-slice-and-events/spec.md
  (R6 to R42), plan.md (Testing, follows), tasks.md (follows, Task 15);
  AGENTS.md; .agents/prompts/doc-reviewer.md;
  .agents/skills/build-slice/SKILL.md; docs/design/sliceflow-design.md
  (risk rules); internal/eventlog/event.go, parse.go;
  internal/slice/status.go, list.go, repo.go, errors.go, and the function
  list of every file in internal/slice; internal/cli/cli.go, status.go,
  list.go, create.go, output.go; internal/git/git.go;
  internal/template/template.go, spec.md; internal/testrepo/testrepo.go;
  internal/e2e/helpers_test.go, and the helper names of every e2e file
- Ran:
  a Python script that estimates each task's brief from its links (06: about 2,450 words, 07: 2,000, 09: 2,300, 11: 2,700, 14: 2,700, 15: 3,900; the others 1,150 to 1,750; every follows name found)
  a Python script that checks 0001's and 0002's specs against R15's marker and labeled-line rules (no problems, so task 04's 0-error tests can pass)
  grep for the requirements that no Shows: line names in 0001's spec (R11, R40, R43, as task 05 and P2 expect)
  grep for a leftover placeholder or comment in plan.md, tasks.md, preview.md (only inside code, which is allowed)
  grep for the uses of `eventlog.Phase`, `currentSliceName`, and `readSlice` (only in internal/slice, and eventlog's own test)
- Requirements: R1 to R37, and R18's added sentence
- Interfaces: `eventlog.Parse`, `Encode`, `NewCreated`, `TimeLayout`,
  `Phase`; `slice.Repo`, `SliceStatus`, `Status`, `List`,
  `currentSliceName`, `readSlice`, `baseWarning`, `FindRepo`, `Error`,
  `Warning`, `failureReason`; `git.Run`, `git.Show`, `git.RefExists`;
  `template.Spec`; `cli.register`, `topUsage`, `Run`, `usageError`,
  `fail`, `newSliceJSON`; `testrepo.New`, `Git`, `Setup`. All the names
  the tasks use exist with the signatures the tasks give, except
  `state.Allowed` (F4).

## Tests

None.

## Previous findings

None.

## Findings

### F1. saveEvent takes the lock itself, so the submit and title flows cannot keep their order or their errors

Level: medium
Where: tasks.md → Task 11 → Interfaces, Done when (`TestSaveNotAllowed`,
  `TestSaveTwoAtOnce`, `TestSaveCancelled`), plan.md → Flows → Shared
  rules → Saving an event (step 2), Flow 3 (steps 1 to 4), Flow 4 (step
  2), Flow 6 (step 3), tasks.md → Task 12, Task 14
Breaks: R6, R11, R20, R32, task 11 (Interfaces)
Why: Flow 3 takes the lock, replays, checks the state (R20), reads
  `spec.md` (R19), checks it, checks R24, and then saves. That is R20's
  order: "R11 first, then takes the lock (R6), then checks this, then
  R19". But `saveEvent` takes the lock inside: `TestSaveTwoAtOnce` and
  `TestSaveCancelled` ("frees the lock") need that, and task 14 does not
  use `takeLock`. Its `saveOp.Data` is fixed before the call. So task 14
  must replay, choose Flow 3 or Flow 5, and read and check `spec.md`
  before the lock.
  Input: two `spec submit` runs at once on a `draft` spec. The second
  reads `draft` with no lock, checks the spec, and calls `saveEvent`.
  `state.Allowed` then refuses `spec.submitted` in `submitted`, and
  `TestSaveNotAllowed` says it returns R4's reason. The user sees
  `spec.submitted is not allowed when spec.md is submitted`. That text is
  not in the spec. The spec's order gives Flow 5, or R6's error.
  Input: a busy lock and a deleted `spec.md`. R20's order gives R7's busy
  error after 10 s. This gives R19's error at once.
  Also, `Check` runs only after `state.Allowed`. So `slice title` on a
  log with no good `slice.created` gets R4's reason
  (`slice.titled comes before slice.created`), not R32's error. And
  Saving an event step 2 says "return the spec's error, e.g. R20's or
  R6's", but `saveEvent` cannot know which one.
Fix: run the caller's check first, under the lock, and let it return the
  spec's error and build `Data` from the replayed state. For example:
  `Prepare func(state.State) (data any, err error)`, run before
  `state.Allowed`. Or have the callers take the lock with `takeLock`
  across their checks, and have `saveEvent` assume it is held. Change
  `TestSaveNotAllowed` and the Saving an event text to match.

### F2. R12 and R19 for spec submit are in no task

Level: medium
Where: tasks.md → Task 14 → covers, Interfaces, Done when; Task 15 →
  covers, Done when; plan.md → Flow 3, steps 1 and 3
Breaks: R12, R19
Why: R12 and R19 apply to `spec validate` and `spec submit`. Only task
  09 covers them, and its tests run only `spec validate`. Task 14 does
  not say whether `Submit` takes `""` for the current branch's slice.
  Task 09 says that only for `Validate`, and `openSlice` does not
  resolve `""`. Task 14 does not list `currentSliceName` in uses. R19 is
  not in task 14's covers, so its message is not in the brief.
  Input: `sliceflow spec submit` on `main`. R12 gives the three lines
  with `Give a name: sliceflow spec submit <name>`. A worker who passes
  `""` to `openSlice` gives R14's `invalid slice name ""` error.
  Input: `spec submit` with `spec.md` deleted. R19 gives
  `Error: <path> not found. Nothing was changed.`, but a worker may
  return the raw system error. No test catches either case.
Fix: add R12 and R19 to task 14's covers, say that `Submit` with `""`
  uses `currentSliceName(ctx, r, "spec submit")`, and add a unit test
  for each.

### F3. Six task briefs are well over 1,500 words, with no reason given

Level: medium
Where: tasks.md → Task 06, Task 07, Task 09, Task 11, Task 14, Task 15
Breaks: docs/templates/tasks.md (keep each brief to about 1,500 words;
  over that, split the task or trim what it follows)
Why: my estimate of each brief is the task section, its covered
  requirements, the scenarios that show them, Key concepts, and its
  follows items. It gives about 2,450 words for 06, 2,000 for 07, 2,300
  for 09, 2,700 for 11, 2,700 for 14, and 3,900 for 15. None of them
  says why it is over, as 0001's task 03 did. Task 15 covers 15
  requirements and 12 scenarios, and follows three flows.
Fix: split task 15, e.g. the refused and accepted paths (S13 to S18,
  S24) and the failure paths (S19 to S23). For the others, trim follows
  (e.g. task 15 does not need all three flows), or add one line that
  says why the brief is longer.

### F4. The plan and task 02 give state.Allowed two return types

Level: medium
Where: plan.md → Interfaces → State machine; tasks.md → Task 02 →
  Interfaces
Breaks: task 02 (Interfaces)
Why: the plan says `func Allowed(s State, e eventlog.Event) error`.
  Task 02 says it returns `string`, `""` when allowed. Task 02 follows
  "State machine", so its worker gets both texts and must pick one.
  Task 11 uses whichever one is built.
Fix: use one signature in both places. With F1's fix, say what the
  caller does with the reason.

### F5. Flow 1 does not say what happens when git cannot compare against the reviewed commit

Level: low
Where: plan.md → Flow 1, step 3; tasks.md → Task 06 → `specChanged`
Breaks: nothing
Why: step 3 says only that exit code 1 means changed. A `review.started`
  whose `commit` is well formed (R5) but not in the repo, e.g. a log
  written by hand, or a branch that was rebased, makes
  `git diff --quiet` exit with 128. If `specChanged` returns that as an
  error, as 0001's code does with git errors, then `slice status` fails,
  and `slice list` fails for every slice, because of one line. A worker
  has to guess. If the answer is a new warning text, the spec must say
  it.
Fix: say in Flow 1 what any other exit code does. The simplest choice
  that adds no new output is to return git's error, as for other git
  failures. If the lead wants a warning instead, ask the human.

### F6. Some Done when lines leave a rule untested or give no expected text

Level: low
Where: tasks.md → Task 02 (`TestReplayNotAllowed`), Task 04
  (`TestErrorPlaceholderValue`), Task 07 (`TestStatusJSONReview`), Task
  11 (Done when), Task 15 (covers, `TestSubmitAcceptedCmd`)
Breaks: nothing
Why: `TestReplayNotAllowed` says "a second `slice.created` is not
  allowed" with no warning text. By R4 it is e.g.
  `.events.jsonl line 2: slice.created is not allowed when spec.md is draft; skipped`.
  `TestErrorPlaceholderValue` has no front matter case, e.g.
  `slice: {{slice}}` and `created: {{date}}`, and no `Rule:` case.
  Those are the values the new R18 sentence is mostly for, and the
  fresh test file has its front matter filled. No task tests R9's "a
  line that R4 skipped counts" for `seq`. Input: `slice.created` seq 1,
  then a skipped `review.finished` seq 2; the new event must get seq 3.
  R31's `changed_after_review`, a running review with no `result`, and
  the two new fields in `slice list --json` have no test. R26 and R27
  have `Check: e2e`, but only unit tests prove them. S14's e2e test
  checks their output but does not name them, and task 15 does not
  cover them.
Fix: add these cases and texts. Add R26 and R27 to task 15's covers and
  to `TestSubmitAcceptedCmd`'s Proves.

### F7. Testing and the tasks disagree in three small places

Level: low
Where: plan.md → Testing; tasks.md → Task 07 → Files, Task 08 → Files,
  Task 11 → `TestSaveTwoAtOnce`
Breaks: nothing
Why: Testing says "Ctrl-C: as in 0001, the e2e test skips on Windows",
  but no task has an e2e Ctrl-C test. Only task 11's unit
  `TestSaveCancelled` exists. Testing says two writes at once save "with
  `seq` 6 and 7"; task 11 says 2 and 3. Testing says one e2e file per
  task, `internal/e2e/<task_name>_test.go`, but task 07 uses
  `status_spec_test.go` and task 08 uses `template_show_test.go`. (07's
  own name is taken by 0001's file.)
Fix: drop the e2e Ctrl-C line, or add the test to task 15. Use one
  `seq` pair. Allow the two file names in Testing, or rename the files.

### F8. Three tasks that later tasks build on are marked normal

Level: low
Where: tasks.md → Task 05, Task 08, Task 09 → risk
Breaks: docs/templates/tasks.md (core: other tasks build on it)
Why: task 05 changes `spec.Check`, which tasks 09 and 14 use. Task 08
  adds the command groups and `template.Raw`, which tasks 09, 12, and 15
  need. Task 09 creates `openSlice`, `Validation`, and `printProblems`,
  which tasks 12, 14, and 15 use. Both levels get a code review, so
  nothing changes now. But the plan reviewer is asked to check risk
  against these rules.
Fix: mark 05, 08, and 09 `core`.

### F9. Task 06 can be read as removing SliceStatus.Phase and Title

Level: low
Where: tasks.md → Task 06 → Interfaces (`SliceStatus`)
Breaks: nothing
Why: "`Phase` and `Title` come from `State`" can mean the fields are
  removed. internal/cli/status.go and list.go read `st.Phase` and
  `st.Title`, and they are not in task 06's touches. So removing the
  fields breaks `go build ./...` in a file the task may not change.
Fix: say the fields stay, and are filled from `State`.

### F10. Plan wording: the Approach and the Reviewer interface

Level: low
Where: plan.md → Approach (the two lines under the table); plan.md →
  Interfaces → Reviewer (`Request.File`); tasks.md → Task 13 →
  Interfaces
Breaks: nothing
Why: Approach says "Only `slice` and `git` touch files or run git", but
  the reviewer writes the review file (Interfaces → Reviewer). The
  dependency line also leaves out `cli` → `spec` (`printProblems` takes
  `[]spec.Problem`), `cli` → `template` (`template show`), and `slice` →
  `eventlog` (`saveEvent` uses `eventlog.New`). The plan calls
  `Request.File` "the review file's path in the worktree", but task 13
  says "the review file's full path".
Fix: say "only `slice`, `git`, and the reviewer touch files", complete
  the dependency line, and write "full path" in the plan.

### F11. Plan template form: touches and Parked

Level: low
Where: plan.md → front matter `touches`, Parked
Breaks: docs/templates/plan.md (touches; optional sections)
Why: the template says `touches` holds "the project-level files this
  slice changes, e.g. a contract or the data model doc". 0001's plan
  used `go.mod`. This plan lists eight code folders instead; the tasks'
  touches already hold those. Parked is optional and says "None.", but
  the template says to delete an optional section that does not apply.
Fix: `touches: []`, and delete the Parked section.

### F12. Preview P2 and P3 do not use S5's and S13's values

Level: low
Where: preview.md → P2, P3
Breaks: docs/templates/preview.md (for a scenario, the steps and Expect:
  lines follow the scenario's exact command and output)
Why: P2 builds a different spec from 0001's (one open question, three
  notes), so its output is not S5's. P2's step 3 Expect: also gives no
  exact line for the error, only "an error for R7's missing Check:
  line". The human cannot compare it word for word. The approach works,
  and the summary lines are exact.
Fix: give the exact lines, e.g.
  `spec.md:<R7's line>: error: R7 has no Check: line`, and say in one
  line that P2 uses 0001's spec in place of S5's.

## Requirements check

None.

## Needs human decision

- **Q1.** Does `sliceflow template show spec` work outside a git repo?
  Flow 7 step 2 says it needs no repo and runs no git command, marked
  (lead's call). R35 does not say. In 0001, every command checks for a
  repo first (R12 of 0001-slice-and-events), so the user sees either the
  template or `Error: not inside a git repository. Nothing was changed.`
  That is something the user sees, so the spec should say it. If the
  answer is yes, R35 gets one sentence, and task 08's
  `TestTemplateShow` already checks it.
  Guess: it works outside a repo, as the plan says. The other choice is
  that it fails with 0001's R12 error, as the other commands do.
