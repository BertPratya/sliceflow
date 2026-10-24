---
review: task-12-1
slice: 0001-slice-and-events
kind: task
target: task 12
round: 1
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

The task covers its requirements from the repo root, the tests are real, and
task-11-1/F1 is fixed. One high problem: `slice list` run from a subfolder
does not find the merged slices, so a merged slice shows as in progress, or
is missing from `--all` (F1).

## Checked

- Commits: 9a94155
- Files read: internal/slice/list.go, internal/slice/list_test.go,
  internal/cli/list.go, internal/cli/status.go, internal/cli/output.go,
  internal/slice/status.go, internal/slice/repo.go, internal/git/git.go,
  internal/e2e/slice_list_test.go, internal/e2e/helpers_test.go, the
  brief, spec.md (R2, R9, R24, R29, R30, Key concepts)
- Ran:
  gofmt -l cmd internal (prints nothing)
  go vet ./... (pass)
  go test -count=1 ./... (pass)
  go build ./... (pass)
  the built binary in a temporary repo: first-slice merged with its branch kept, spec-submit in progress; `slice list` in the root (only spec-submit, correct) and in a subfolder `sub/` (first-slice listed too, wrong)
  the same repo after removing first-slice's worktree and branch: `slice list --all` in `docs/` and in the slice worktree's `docs/` (first-slice missing, wrong)
  git ls-tree --full-tree -d --name-only -z refs/heads/main docs/slices/ in a subfolder (lists docs/slices/first-slice)
- Requirements: R21, R22, R23, R28, R34, R35, R36, R37, R38, R42, R43
- Interfaces: `SliceList` and `List(ctx, r, all)` match the task; it uses
  `readSlice`, `baseWarning`, `SliceStatus`, `FindRepo`, `printWarnings`,
  and `run`

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

- task-11-1/F1: verified (fixed)

## Findings

### F1. In a subfolder, slice list does not see the merged slices

Level: high
Where: internal/slice/list.go:52
Breaks: R35
Why: `git ls-tree` reads its path relative to the current folder, and `List`
  runs it in `r.Dir`, the folder the command runs in, not the top of the
  repo. In `docs/`, git looks for `docs/docs/slices/`. So the list of merged
  slices is empty. I ran the built binary in a temporary repo with
  first-slice merged into `main`, its branch kept. In the root, `slice list`
  printed only spec-submit. In a subfolder `sub/`, it also printed
  first-slice as a row, but first-slice is merged. Then I removed
  first-slice's branch. `slice list --all` in `docs/` and in the slice
  worktree's `docs/` left out first-slice. The spec cares about subfolders
  (R9). A user often runs sliceflow inside the slice worktree's
  `docs/slices/<name>/`. The tests run only in the repo root, so they pass.
  `readSlice` is fine: its `<ref>:<path>` form is always from the top.
Fix: add `--full-tree` to the `ls-tree` command (I checked: in a subfolder
  it lists `docs/slices/first-slice`). Add a test that runs `slice list` and
  `--all` from a subfolder.

## Requirements check

None.

## Needs human decision

None.
