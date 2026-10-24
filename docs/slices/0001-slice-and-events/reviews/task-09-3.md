---
review: task-09-3
slice: 0001-slice-and-events
kind: task
target: task 09
round: 3
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

Both findings of round 2 are fixed as R004's plan now says. The cleanup
runs after every failed `git worktree add`, an entry git listed before
step 6 stays, and `git worktree add` runs with `--quiet`. The two new
tests are real: each fails when I break the code it covers. No new
findings. The stale-entry reason ends with `;.`, e.g.
`... is a missing but already registered worktree;. Everything sliceflow made was removed.`
This is not a finding. It follows R18 (git's first line) and Flow 3
step 5 (drop only a final `.`) exactly. Changing it would change the
plan, and it only reads a little oddly in a rare case.

## Checked

- Commits: b99ebcb, 62a4d16, 52b9a1a (the diff since round 2: 52b9a1a)
- Files read: internal/slice/create.go, internal/slice/undo.go,
  internal/slice/undo_test.go, the brief (R004, with Flow 1 step 6),
  reviews/task-09-1.md, reviews/task-09-2.md, the body of 52b9a1a and
  its `ASSUMED` note
- Ran:
  gofmt -l cmd internal (prints nothing)
  go vet ./... (pass)
  go build ./... (pass)
  go test -count=1 ./... (pass)
  go test -count=3 -run 'TestUndo|TestReason' ./internal/slice/ (pass)
  go test -count=2 -run 'TestCreateHook|TestCreateInterrupt' ./internal/e2e/ (pass)
  a mutation in a scratch copy: listedBefore forced to false (TestUndoKeepsStaleEntry fails: "the stale entry was removed")
  a mutation in a scratch copy: cleanFailedAdd only when ctx is done (TestUndoFailedWorktreeAdd fails: "Could not remove: branch slice/spec-submit")
  the built binary in a temporary repo with a stale detached worktree entry at `.sliceflow/worktrees/x` (exit 1, the entry stays, no branch, `.sliceflow/.gitignore` made by the run removed, `git status` clean, git's second line indented by 2 spaces)
- Requirements: R8, R18
- Interfaces: Create returns R18's *Error after the checks; git.Error used as in task 01. cleanStopped renamed to cleanFailedAdd, which is internal.

## Tests

- `TestCreateHookRefuses`: ok
- `TestCreateHookOutput`: ok
- `TestCreateInterrupt`: ok
- `TestUndoCancelledContext`: ok
- `TestUndoCancelledDuringChecks`: ok
- `TestUndoInterruptedWorktreeAdd`: ok
- `TestUndoFailedWorktreeAdd`: ok
- `TestUndoKeepsStaleEntry`: ok
- `TestUndoKeepsOthers`: ok
- `TestUndoCouldNotRemove`: ok
- `TestReasonFileError`: ok
- `TestReasonGitCommand`: ok

## Previous findings

- task-09-2/F1: verified (fixed)
- task-09-2/F2: verified (fixed)

## Findings

None.

## Requirements check

None.

## Needs human decision

None.
