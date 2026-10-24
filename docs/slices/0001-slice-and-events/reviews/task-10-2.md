---
review: task-10-2
slice: 0001-slice-and-events
kind: task
target: task 10
round: 2
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

task-10-1/F1 is fixed. `baseWarning` now checks `refs/heads/<base>`
with `show-ref --verify --quiet`, as `checkCreate` does. `readSlice`
reads the base folder and log through `refs/heads/<base>`. The three
new tests fail on the old code and pass on the fix. There are no new
findings.

## Checked

- Commits: 2979d45, 15564ff (diff since round 1: 2979d45..15564ff)
- Files read: internal/slice/status.go, internal/slice/status_test.go
  (the diff, and the six Done when tests), internal/slice/base.go,
  internal/slice/check.go (base handling), internal/git/git.go,
  reviews/task-10-1.md, the brief for task 10
- Ran:
  gofmt -l cmd internal (prints nothing)
  go vet ./... (pass)
  go test -count=1 ./... (pass)
  go build ./... (pass)
  go test -count=1 -run TestStatus -v ./internal/slice/ (10 tests pass)
  the three new tests against the status.go of 2979d45, in a throwaway
    copy of the tree in the scratchpad, now deleted (all three fail: no
    warning for a tag `main`, `no slice named spec-submit` for an old
    tag `main`, no warning for `sliceflow.base` set to `HEAD`)
  git show-ref --verify --quiet refs/heads/<b> in a throwaway repo, for
    b = nope, main~1, a b, main:x, a..b, empty, -x, HEAD (exit code 1
    for each, so a strange `sliceflow.base` gives R23's warning, not an
    error)
- Requirements: R21, R23, R24, R25, R26, R33
- Interfaces: `SliceStatus`, `Status`, `readSlice`, and `baseWarning`
  still match task 10 exactly

## Tests

- `TestStatusReadsBranch`: ok
- `TestStatusReadsCommitted`: ok
- `TestStatusMergedFromBase`: ok
- `TestStatusNoLog`: ok
- `TestStatusBadName`: ok
- `TestStatusBaseMissing`: ok

## Previous findings

- task-10-1/F1: verified (fixed)

## Findings

None.

## Requirements check

None.

## Needs human decision

None.
