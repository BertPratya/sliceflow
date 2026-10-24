---
review: task-07-1
slice: 0001-slice-and-events
kind: task
target: task 07
round: 1
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

The task works: `slice create` prints R9's five lines and R10's JSON,
wrong use and help follow R38 and R39, and the tests are real. The
worker's ASSUMED choice holds on every system, because `Created.Worktree`
is built from the same `Repo.Root` string it is cut from. One low
problem: `TestCreateSpecSubmit` does not check the event log line that
its Done when names (F1).

## Checked

- Commits: 2840bb8
- Files read: internal/cli/create.go, internal/e2e/create_command_test.go,
  internal/e2e/helpers_test.go, internal/cli/cli.go, internal/cli/flags.go,
  internal/cli/output.go, internal/slice/repo.go, internal/slice/create.go,
  internal/slice/create_test.go (event line check), internal/testrepo/testrepo.go,
  docs/slices/0001-slice-and-events/tasks.md (tasks 08, 09, 11),
  docs/slices/0001-slice-and-events/spec.md (R42, S29)
- Ran:
  gofmt -l cmd internal (nothing)
  go vet ./... (pass)
  go test ./... (pass)
  go build ./... (pass)
  built binary, temp repo, `slice create a1` from the subfolder `sub/deep` (S1's five lines, relative paths, exit 0)
  built binary, same repo, `slice create a2 --json` from `sub/deep` (one JSON object, full path with `/`, exit 0)
- Requirements: R1, R3, R5, R9, R10, R38, R39, R41
- Interfaces: `FindRepo`, `Create`, `CreateOptions`, `Created`, `register`,
  `parseFlags`, `usageError`, `fail`, `run` are used as the task says

## Tests

- `TestCreateSpecSubmit`: weak, see F1
- `TestCreateJSON`: ok
- `TestCreateFromFlag`: ok
- `TestCreateInsideWorktree`: ok
- `TestCreateHelp`: ok
- `TestCreateWrongUse`: ok

## Previous findings

None.

## Findings

### F1. TestCreateSpecSubmit does not check the event log line

Level: low
Where: internal/e2e/create_command_test.go:23-48
Breaks: nothing
Why: Done when says "the log is one line like S1's", and S1 says
  `.events.jsonl` holds one `slice.created` line. The test checks only
  the commit subject (`git log --format=%s`), never the content of
  `.events.jsonl`. The unit test in internal/slice/create_test.go:77-78
  already checks the exact line, so R19 and R20 stay proven; only this
  e2e test checks less than its Done when.
Fix: read `slice/spec-submit:docs/slices/spec-submit/.events.jsonl` with
  `git show`, and check it is one line with `"event":"slice.created"` and
  `"from":{"branch":"main","commit":"<full commit>"}`.

## Requirements check

None.

## Needs human decision

None.
