---
review: task-11-1
slice: 0001-slice-and-events
kind: task
target: task 11
round: 1
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

The task covers R22, R26, R28, R29, R30, R31, R32, R33, and R42. Each
test in Done when checks the exact output, stderr, and exit code. The
deferred task-03-1/F2 is fixed: S15's three lines print with no indent.
One low note: the R30 object is built inline in `runStatus`, and task 12
needs the same object but cannot touch `internal/cli/status.go` (F1).

## Checked

- Commits: 1abe237
- Files read: internal/cli/status.go, internal/e2e/status_command_test.go,
  internal/cli/output.go, internal/cli/cli.go, internal/cli/flags.go,
  internal/cli/create.go (JSON part), internal/slice/status.go,
  internal/slice/errors.go, internal/slice/repo.go (errors),
  internal/eventlog/event.go (From), internal/e2e/helpers_test.go,
  the brief, tasks.md task 12, spec.md R37
- Ran:
  gofmt -l cmd internal (prints nothing)
  go vet ./... (pass)
  go build ./... (pass)
  go test ./... (pass)
  go test -count=1 -run TestStatus -v ./internal/e2e/ (all 12 pass)
  built binary by hand in temp repos with a temp GIT_CONFIG_GLOBAL: repo with no commits gives R13's error, exit 1; TZ=Asia/Bangkok still prints `Created:` in UTC; `--json` before the name works (TD2); `-h` prints the usage; `--jsn` gives `unknown flag --jsn`, exit 2
- Requirements: R22, R26, R28, R29, R30, R31, R32, R33, R42
- Interfaces: `FindRepo`, `Status`, `SliceStatus`, `register`,
  `parseFlags`, `usageError`, `fail`, `printWarnings`, `run` are used as
  the task says

## Tests

- `TestStatusSpecSubmit`: ok
- `TestStatusCurrentBranch`: ok
- `TestStatusNotOnSlice`: ok
- `TestStatusJSON`: ok
- `TestStatusBadLines`: ok
- `TestStatusNoLog`: ok
- `TestStatusNoSuchSlice`: ok
- `TestStatusErrorJSON`: ok
- `TestStatusNotARepo`: ok

## Previous findings

- task-03-1/F2: verified (fixed)

## Findings

### F1. The R30 object is built inline, and task 12 cannot reuse it

Level: low
Where: internal/cli/status.go:58-87
Breaks: nothing
Why: R37 says each slice in `slice list --json` is an object "in the
  form of R30". The struct for that object, the `from` type, and the
  `created` format are local to `runStatus`. Task 12's touches list
  `internal/cli/list.go` but not `internal/cli/status.go`, so task 12
  must copy this code, and the two copies can drift apart.
Fix: widen task 12's touches to `internal/cli/status.go`, so its
  worker can move the R30 object into one helper that both commands call.

## Requirements check

None.

## Needs human decision

None.
