---
review: task-01-2
slice: 0001-slice-and-events
kind: task
target: task 01
round: 2
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

Both findings from round 1 are handled. F2 is fixed: the no-network test
now fails against a broken `Run`, even when the outer shell sets the two
variables. F1 is a wontfix with a doc comment, which is fine for a low
finding. The fix commit adds no new problems.

## Checked

- Commits: ab4fce5, d95a9de (diff since round 1: ab4fce5..d95a9de)
- Files read: internal/git/git.go, internal/git/git_test.go, the brief,
  reviews/task-01-1.md, the d95a9de commit message
- Ran:
  gofmt -l internal (prints nothing)
  go vet ./... (pass)
  go test -count=1 ./... (pass; no GIT_DIR, GIT_WORK_TREE, or GIT_INDEX_FILE set)
  go build ./... (pass)
  a scratch copy with the two variables removed from `Run`, run with GIT_TERMINAL_PROMPT=0 and GIT_NO_LAZY_FETCH=1 in the shell: TestRunSetsNoNetworkEnv fails (good; in round 1 it passed)
- Requirements: R40
- Interfaces: Run, Error, RefExists, Show, Setup, New, Git unchanged
  since round 1, and still match the task

## Tests

- `TestRunReturnsOutput`: ok
- `TestRunErrorHasStderr`: ok
- `TestRunSetsNoNetworkEnv`: ok
- `TestRunCancelled`: ok
- `TestRefExists`: ok
- `TestShowMissingFile`: ok

## Previous findings

- task-01-1/F1: verified (wontfix)
- task-01-1/F2: verified (fixed)

## Findings

None.

## Requirements check

None.

## Needs human decision

None.
