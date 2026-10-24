---
review: task-01-1
slice: 0001-slice-and-events
kind: task
target: task 01
round: 1
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

The task works as the brief says: `Run`, `Error`, `RefExists`, `Show`,
and the test helpers match Interfaces, and every Done when test is real.
Two low findings only: `Show` hides a bad ref or a non-repo folder as
"file not there" (F1), and the no-network test passes against a broken
`Run` when the outer environment already sets the two variables (F2).

## Checked

- Commits: ab4fce5
- Files read: go.mod, internal/git/git.go, internal/git/git_test.go,
  internal/testrepo/testrepo.go, internal/testrepo/testrepo_test.go,
  the brief, plan.md (Flow 4, the Ctrl-C notes, TD6), tasks.md (later
  uses of `Show`, `RefExists`, `git.Error`, testrepo)
- Ran:
  gofmt -l internal (prints nothing)
  go vet ./... (pass)
  go test -count=1 ./... (pass)
  go build ./... (pass)
  go test -count=1 ./... with GIT_DIR, GIT_INDEX_FILE, and GIT_WORK_TREE pointing at a scratch decoy repo, as inside a hook (pass; the decoy still has no commit and no new config)
  a scratch copy with the two variables removed from `Run`: TestRunSetsNoNetworkEnv fails (good), but passes when the outer env sets GIT_TERMINAL_PROMPT=0 and GIT_NO_LAZY_FETCH=1 (see F2)
  a scratch probe: Show in a folder that is no repo, and Show with ref nosuchref, both return nil, false, nil (see F1)
- Requirements: R40
- Interfaces: Run, Error, RefExists, Show, Setup, New, Git match the
  task. go.mod has module github.com/BertPratya/sliceflow and go 1.25,
  no require lines (TD1, TD3). git is found once with exec.LookPath,
  run with exec.CommandContext and cmd.Dir (TD8), WaitDelay 2 seconds.

## Tests

- `TestRunReturnsOutput`: ok
- `TestRunErrorHasStderr`: ok
- `TestRunSetsNoNetworkEnv`: ok
- `TestRunCancelled`: ok
- `TestRefExists`: ok
- `TestShowMissingFile`: ok

## Previous findings

None.

## Findings

### F1. Show reports a bad ref or a non-repo folder as a missing file

Level: low
Where: internal/git/git.go:91-97
Breaks: nothing
Why: `Show` returns `nil, false, nil` for any `*git.Error` from
  `cat-file -e`. A scratch probe showed `Show(ctx, <not a repo>, "main",
  "a.txt")` and `Show(ctx, repo, "nosuchref", "a.txt")` both return
  `nil, false, nil`, not an error. Interfaces says the bool is false
  "when the file is not in that commit"; a missing commit or repo is a
  different case. In this slice the callers check the ref first (Flow 4
  step 5) and the repo (task 04), so no user-visible wrong result yet.
  The worker's ASSUMED note names this choice.
Fix: optionally check the ref with `RefExists` (or `rev-parse --verify`)
  first, and pass its error on; or leave it, and note it in the doc
  comment.

### F2. The no-network test can pass against a broken Run

Level: low
Where: internal/git/git_test.go:44-62
Breaks: nothing
Why: the test reads the variables from the child's environment, which
  also inherits the test process's own. With the two variables removed
  from `Run` in a scratch copy, the test failed, but passed when the
  shell already had `GIT_TERMINAL_PROMPT=0` and `GIT_NO_LAZY_FETCH=1`.
  A CI runner or developer shell that sets them hides a regression.
Fix: `t.Setenv("GIT_TERMINAL_PROMPT", "1")` and
  `t.Setenv("GIT_NO_LAZY_FETCH", "0")` before the call; Go keeps the
  last value of a duplicate key, so the test then proves `Run` sets them.

## Requirements check

None.

## Needs human decision

None.
