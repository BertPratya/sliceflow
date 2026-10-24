---
review: task-12-2
slice: 0001-slice-and-events
kind: task
target: task 12
round: 2
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

task-12-1/F1 is fixed: `git ls-tree` now runs with `--full-tree`, so
`slice list` from a subfolder gives the same rows as from the root. The new
test fails without the fix and passes with it. The fix changes nothing else,
and there are no new findings.

## Checked

- Commits: 9a94155, e3cbe38 (the diff since round 1)
- Files read: internal/slice/list.go, internal/e2e/slice_list_test.go
  (s20Repo, mergeAndRemove, TestListFromSubfolder), the brief,
  reviews/task-12-1.md
- Ran:
  gofmt -l cmd internal (prints nothing)
  go vet ./... (pass)
  go test -count=1 ./... (pass)
  go build ./... (pass)
  go test -run TestListFromSubfolder ./internal/e2e/ on a scratch copy without `--full-tree` (fails in both `--all` cases, as expected)
  the built binary in a temporary repo, first-slice merged with its branch kept, spec-submit in progress: `slice list` in `sub/` (only spec-submit, correct), `slice list --all` in `sub/` (both, correct), `slice list --all --json` in the slice worktree's `docs/slices/spec-submit/` (both, exit code 0, correct)
- Requirements: R35, R34
- Interfaces: `List(ctx, r, all)` unchanged

## Tests

- `TestListInProgress`: ok
- `TestListAll`: ok
- `TestListBadLine`: ok
- `TestListJSON`: ok
- `TestListEmpty`: ok
- `TestListBaseMissing`: ok
- `TestListUnknownFlag`: ok
- `TestListNotARepo`: ok
- `TestListErrorJSON`: ok
- `TestListBadBranchName`: ok
- `TestListOrder`: ok

## Previous findings

- task-12-1/F1: verified (fixed)

## Findings

None.

## Requirements check

None.

## Needs human decision

None.
