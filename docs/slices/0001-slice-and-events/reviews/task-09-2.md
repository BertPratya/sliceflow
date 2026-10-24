---
review: task-09-2
slice: 0001-slice-and-events
kind: task
target: task 09
round: 2
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

All four findings of round 1 are handled. The interrupt during
`git worktree add` now cleans up fully, as the plan's Undo on failure
(R003) says, and I could not make it leave anything. One new high
problem, which round 1 already had: when `git worktree add` fails
without a signal, e.g. because the user's `post-checkout` hook fails, git
keeps the full worktree. `create` leaves it, does not name it, and
removes `.sliceflow/.gitignore` (F1). The plan does not say what to do
then (Q1).

## Checked

- Commits: b99ebcb, 62a4d16 (the diff since round 1: 62a4d16)
- Files read: internal/slice/create.go, internal/slice/undo.go,
  internal/slice/undo_test.go, internal/cli/create.go,
  internal/git/git.go, plan.md (Flow 1 step 6, Flow 3, Undo on failure,
  Interrupts), the brief, reviews/task-09-1.md, the body of 62a4d16
- Ran:
  gofmt -l cmd internal (prints nothing)
  go vet ./... (pass)
  go build ./... (pass)
  go test -count=1 ./... (pass)
  go test -count=3 for TestUndo*, TestCreateInterrupt, TestCreateHook*, TestReason* (pass)
  the built binary in a temporary repo with 20,000 files, slice create, SIGINT after 0.10 to 0.6 s, 10 runs (each: interrupted, exit 1, no branch, no worktree, no folder, no .git/worktrees entry, no .sliceflow/ made by the run). At 0.15 s, git's child `git reset --hard` was running; it was gone after, and nothing came back.
  the built binary in a temporary repo with a post-checkout hook that prints to stderr and exits with 3 (F1, F2)
- Requirements: R8, R18
- Interfaces: Create returns R18's *Error after the checks; git.Error used as in task 01. slice.Interrupted kept (task-09-1/F3, wontfix).

## Tests

- `TestCreateHookRefuses`: ok
- `TestCreateHookOutput`: ok
- `TestCreateInterrupt`: ok
- `TestUndoCancelledContext`: ok
- `TestUndoCancelledDuringChecks`: ok
- `TestUndoInterruptedWorktreeAdd`: ok
- `TestUndoKeepsOthers`: ok
- `TestUndoCouldNotRemove`: ok
- `TestReasonFileError`: ok
- `TestReasonGitCommand`: ok

## Previous findings

- task-09-1/F1: verified (fixed)
- task-09-1/F2: verified (fixed)
- task-09-1/F3: verified (wontfix)
- task-09-1/F4: verified (fixed)

## Findings

### F1. A failed git worktree add without a signal leaves the worktree, unnamed, and removes .sliceflow/.gitignore

Level: high
Where: internal/slice/create.go:149-156
Breaks: R18
Why: git keeps a worktree when only its `post-checkout` hook fails, and
  then exits with the hook's code (git's builtin/worktree.c: "Hook
  failure does not warrant worktree deletion"). R8 says the user's hooks
  run, so this is a real case. `create` runs `cleanStopped` only when
  `ctx` is done, and adds the worktree to the list only on success. So the
  undo skips the worktree, `git branch -D` refuses the branch the
  worktree holds, and `.sliceflow/.gitignore` is removed while the
  worktree stays. I ran the built binary in a fresh repo with a
  `post-checkout` hook that prints `post-checkout says no` and exits with
  3. It printed
  `Error: could not create slice pc: Preparing worktree (checking out 'slice/pc'). Could not remove: branch slice/pc.`
  and exit code 1, but `.sliceflow/worktrees/pc` stayed as a registered
  worktree on `slice/pc`, and is not in the list. `.sliceflow/` now holds
  only `worktrees/`, with no `.gitignore`, so the main checkout's
  `git status` shows `.sliceflow/` as untracked. Round 1 had the same
  gap (its message even said everything was removed). The code follows
  the plan, which handles only the stopped case, so see Q1.
Fix: see Q1. One way: run `cleanStopped` after every failed
  `git worktree add`, not only after a signal. The plan's reason holds
  in both cases: R17 proved the folder was not there, and only this run
  owns the branch. Add a unit test with a `post-checkout` hook that runs
  `exit 3`.

### F2. The reason for a failed git worktree add is git's progress line

Level: low
Where: internal/slice/create.go:155
Breaks: nothing
Why: `git worktree add` first prints `Preparing worktree (checking out 'slice/<name>')`
  on stderr. So any failure after that line, e.g. the hook in F1 or a
  failed checkout, gives the reason
  `Preparing worktree (checking out 'slice/pc')`, and git's real error
  moves to the indented lines. This is what R18 says ("the first line of
  git's message that is not empty"), so nothing approved breaks. It only
  reads badly. `git worktree add --quiet` would drop that line; that is a
  change to Flow 1 step 6, for the lead or the human.

## Requirements check

None.

## Needs human decision

- **Q1.** The plan's Undo on failure adds the worktree to the list only
  "after `git worktree add` succeeds", and cleans up what the command
  left only "when the signal stops `git worktree add`". It does not say
  what to do when `git worktree add` fails on its own and git keeps the
  worktree, as it does when the `post-checkout` hook fails (F1). Should
  `create` also clean up at `.sliceflow/worktrees/<name>` then, the same
  way as after a signal?
  Guess: yes. Run the same cleanup after every failed
  `git worktree add`, since the same reason (R17, and this run owns the
  branch) proves the worktree is this run's.
