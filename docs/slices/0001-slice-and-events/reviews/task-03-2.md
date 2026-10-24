---
review: task-03-2
slice: 0001-slice-and-events
kind: task
target: task 03
round: 2
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

The fixes from round 1 work. A flag after `slice` other than `-h` or
`--help` now gives R38's unknown-flag error with exit 2, and a bad flag
value gives the text the lead decided. No new findings.

## Checked

- Commits: b887065, 622be4e (diff since round 1)
- Files read: internal/cli/cli.go, internal/cli/flags.go,
  internal/cli/output.go, internal/cli/cli_test.go,
  internal/cli/flags_test.go, internal/e2e/cli_skeleton_test.go, the
  brief, task-03-1, the body of 622be4e
- Ran:
  gofmt -l cmd internal (prints nothing)
  go vet ./... (pass)
  go test -count=1 ./... (pass, no GIT_DIR, GIT_WORK_TREE, or GIT_INDEX_FILE set)
  go build ./... (pass)
  the built binary with `slice --colour`, `slice --json`, `slice -x`, `slice --colour=x`, `--colour` (each: R38's unknown-flag error on stderr, exit 2)
  the built binary with `slice --help`, `slice -h` (exit 0)
- Requirements: R28, R38, R39
- Interfaces: Run, command, register, parseFlags, errHelp, usageError,
  fail, printWarnings, run unchanged since round 1, and match the task

## Tests

- `TestParseFlagsAfterName`: ok
- `TestParseFlagsWrongUse`: ok
- `TestPrintWarnings`: ok
- `TestFailIndentsLines`: ok
- `TestHelpTop`: ok
- `TestUnknownCommand`: ok

## Previous findings

- task-03-1/F1: verified (fixed)
- task-03-1/F2: verified (deferred)
- task-03-1/F3: verified (fixed)

## Findings

None.

## Requirements check

None.

## Needs human decision

None.
