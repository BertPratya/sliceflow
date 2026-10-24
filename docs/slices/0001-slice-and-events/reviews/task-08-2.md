---
review: task-08-2
slice: 0001-slice-and-events
kind: task
target: task 08
round: 2
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

task-08-1/F1 is fixed. R16 now checks the exact ref
`refs/heads/<base>`, so `main~1`, `main@{0}`, `main^`, and `HEAD` give
R16's error, and new tests prove it. The lead's wontfix of
task-08-1/F2 is acceptable. No new findings.

## Checked

- Commits: 9cdc0eb, f73cd3b (the diff since round 1)
- Files read: internal/slice/check.go, internal/slice/check_test.go,
  internal/e2e/create_checks_test.go, internal/slice/create.go (the
  commit lookup after the checks), internal/git/git.go (Error,
  ExitCode, RefExists), docs/slices/0001-slice-and-events/status.md
  (Findings, Lead's decisions), reviews/task-08-1.md
- Ran:
  gofmt -l cmd internal (prints nothing)
  go vet ./... (pass)
  go test ./... (pass)
  go build ./... (pass)
  `git show-ref --verify --quiet refs/heads/<b>` in a temporary repo:
    exit 0 for `main` and `feature/x`; exit 1 for `main~1`,
    `main@{0}`, `HEAD`, `-x`, `a..b`, `foo bar`, `main.lock`, and an
    empty name. So a bad ref format also gives R16's error, not a
    system error.
- Requirements: R2, R11, R12, R13, R14, R15, R16, R17, R42
- Interfaces: `checkCreate(ctx, r, opts, base, source) error` is
  unchanged; it uses `CheckName`, `git.RefExists`, `git.Run`

## Tests

- `TestCreateBadName`: ok
- `TestCreateReservedName`: ok
- `TestCreateRemoteExists`: ok
- `TestCreateBaseMissing`: ok
- `TestCreateFolderExists`: ok
- `TestCreateNotARepo`: ok
- `TestCreateNoCommits`: ok
- `TestCreateBadNameJSON`: ok
- `TestCheckExistsOrder`: ok
- `TestCheckOtherRemoteBranch`: ok
- `TestCheckOrder`: ok

## Previous findings

- task-08-1/F1: verified (fixed)
- task-08-1/F2: verified (wontfix)

## Findings

None.

## Requirements check

None.

## Needs human decision

None.
