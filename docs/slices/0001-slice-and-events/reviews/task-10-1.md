---
review: task-10-1
slice: 0001-slice-and-events
kind: task
target: task 10
round: 1
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

The core follows Flow 4. The names in Interfaces match, and the six tests
in Done when are real. One high problem: the base branch goes to git as a
bare name, not as `refs/heads/<base>`. So a tag named `main` hides a
merged slice, and it also stops R23's warning (F1). On the lead's point
(2): `Status` returns R32's text as one `*Error` with three lines. Task 11
can print S15 exactly if it prints a `*slice.Error` from `Status` as
`Error: ` plus the message, and does not pass it to `fail`.

## Checked

- Commits: 2979d45
- Files read: internal/slice/status.go, internal/slice/status_test.go,
  internal/slice/base.go, internal/slice/check.go, internal/slice/errors.go,
  internal/slice/create.go (base handling), internal/slice/create_test.go
  (helpers), internal/git/git.go, internal/eventlog/event.go,
  internal/eventlog/parse.go, internal/cli/output.go, plan.md (Flow 4,
  Flow 5, Git arguments), tasks.md (tasks 11 and 12), spec.md (S15, R31 to
  R33), reviews/task-03-1.md (F2)
- Ran:
  gofmt -l cmd internal (prints nothing)
  go vet ./... (pass)
  go test ./... (pass)
  go build ./... (pass)
  three probe tests in a throwaway copy of the tree in the scratchpad,
    now deleted: a tag `main` with no branch `main`; a tag `main` on the
    first commit plus a merged slice on branch `main`; `sliceflow.base`
    set to `HEAD` (results in F1)
- Requirements: R21, R23, R24, R25, R26, R33
- Interfaces: `SliceStatus`, `Status`, `readSlice`, and `baseWarning`
  match task 10 exactly

## Tests

- `TestStatusReadsBranch`: ok
- `TestStatusReadsCommitted`: ok
- `TestStatusMergedFromBase`: ok
- `TestStatusNoLog`: ok
- `TestStatusBadName`: ok
- `TestStatusBaseMissing`: ok

## Previous findings

None.

## Findings

### F1. The base branch goes to git as a bare name, not `refs/heads/<base>`

Level: high
Where: internal/slice/status.go:88, internal/slice/status.go:112,
  internal/slice/status.go:116
Breaks: plan.md Git arguments, R23, R24
Why: `baseWarning` runs `git rev-parse --verify --quiet <base>^{commit}`,
  and `readSlice` runs `git cat-file -e <base>:docs/slices/<name>`, then
  `git show <base>:...` through `git.Show`. Git looks a bare name up in
  `refs/tags/` before `refs/heads/`, and it also takes any revision. I
  ran probe tests in a throwaway copy:
  (a) There is a tag `main` and no branch `main`. `baseWarning` returns
  exists=true and no warning, so R23's warning is missing.
  (b) There is a tag `main` on the first commit, and the merged slice
  `spec-submit` is on branch `main`. `Status(ctx, r, "spec-submit")`
  returns `no slice named spec-submit. See all slices: sliceflow slice list`,
  where R24 says to read it from the base branch.
  (c) `sliceflow.base` is set to `HEAD`. It is taken as an existing base,
  though no branch has that name.
  The plan says a branch from `sliceflow.base` is passed as
  `refs/heads/<branch>`. `create.go` and `check.go` already do that.
  Task 12 builds on `baseWarning` and `readSlice`, so it would get the
  same bug.
Fix: in `baseWarning`, check `refs/heads/`+base. In `readSlice`, set
  `ref` to `refs/heads/`+base for both the `cat-file` and the `show`.
  `base` itself stays the bare name for the warning text. Add a test
  where a tag `main` exists.

## Requirements check

None.

## Needs human decision

None.
