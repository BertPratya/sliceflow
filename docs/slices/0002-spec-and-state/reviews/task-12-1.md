---
review: task-12-1
slice: 0002-spec-and-state
kind: task
target: task 12
round: 1
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

The task covers R32, R33, R34, and R36, and the output matches S30 to S33
character for character. The tests are real. Two low findings: a Ctrl-C
during the first read-only checks prints a git error instead of R10's
`interrupted` (F1), and the ASSUMED reading of "the title the slice has
now" has no test (F2).

## Checked

- Commits: c6081ab
- Files read: internal/slice/title.go, internal/slice/title_test.go,
  internal/cli/title.go, internal/e2e/title_command_test.go,
  internal/slice/save.go, internal/slice/open.go, internal/cli/create.go,
  internal/cli/output.go, internal/cli/status.go (statusFail),
  internal/cli/flags.go, internal/state/moves.go, internal/state/state.go,
  spec.md (R11, R32 to R34, R36, S30 to S33, SD11, the event table),
  the brief with its plan items (Flow 6, Core API, CLI, Interrupts, Git
  arguments, Tests, Testing)
- Ran:
  go test ./internal/slice/ ./internal/e2e/ -run Title -count=1 (pass)
  go build of the binary, then in a temporary repo with `slice create demo`:
  `slice title demo demo` (exit 1, R33's "already" error),
  `slice title demo -- -x` (exit 0, title `-x`),
  `slice title demo Hello extra` (exit 2, `too many arguments`),
  `slice title demo` (exit 2, `missing title`),
  `slice title` (exit 2, `missing slice name`),
  `slice title demo --json t` (exit 2, `unknown flag --json`),
  `slice title demo 'a "q" b'` (exit 0, printed as typed; `slice status`
  shows it; commit `demo: title set`)
- Requirements: R32, R33, R34, R36
- Interfaces: `SetTitle(ctx, r, name, title string) error` matches the
  task; uses `openSlice` and `saveEvent`; registered with `register`

Notes on the lead's points:
- The ASSUMED reading of R33 ("the title the slice has now" is the shown
  title: the newest saved title, else the slice name) agrees with R34 and
  R29 of 0001 (`Title:` shows the name when no title was saved), and with
  SD11's reason ("the name already shows when no title was set"). Saving
  the name as the title would change nothing the user sees, so the
  "already" error is a fair reading. It changes no document. Its other
  ASSUMED lines (statusFail, literal quotes, a title that starts with `-`
  needs `--`) also change no document.
- Order of the checks: `openSlice` runs R12 to R14 and R33 of 0001, then
  R11; `saveEvent` takes the lock; then `Prepare` runs no `slice.created`,
  empty, same, in Flow 6 step 3's order. This is R11's order.
- The event is written only through `saveEvent`, so the lock, the undo,
  and R10's errors are those of task 11. The signal `ctx` is set up as in
  `slice create`, before any git command. Plain git: the title goes only
  into the JSON line and the output; the commit message is
  `<name>: title set`.

## Tests

- `TestTitleSet`: ok
- `TestTitleSame`: ok
- `TestTitleEmpty`: ok
- `TestTitleMissing`: ok
- `TestTitleUnknownState`: ok

## Previous findings

None.

## Findings

### F1. A Ctrl-C before the save prints a git error, not R10's interrupted

Level: low
Where: internal/cli/title.go:50-57, internal/slice/title.go:14-17
Breaks: nothing
Why: The signal `ctx` is made before `FindRepo` and `openSlice`. A signal
  while their git commands run stops git, and the CLI prints that git
  error, e.g. a killed `git rev-parse`. `saveEvent` checks `ctx.Err()`
  only when it starts. Nothing is changed in this window, and it is very
  short. `slice create` handles the same case after `FindRepo` with
  `slice.Interrupted`. Found by reading the code; not run.
Fix: when `ctx.Err() != nil` after `FindRepo` or `openSlice` fails,
  return R10's error with the reason `interrupted`, as `saveEvent` does.

### F2. No test for an untitled slice given its own name as the title

Level: low
Where: internal/slice/title_test.go, internal/slice/title.go:28-31
Breaks: nothing
Why: The ASSUMED reading (no saved title: the slice name is the title
  now) is code that no test reaches. `TestSetTitleSameAndEmpty` uses a
  slice created with a title. I ran it by hand: `slice title demo demo`
  on an untitled slice gives R33's "already" error, as the worker meant.
Fix: add a case to `TestSetTitleSameAndEmpty` with a slice that has no
  title, and its name as the title.

## Requirements check

None.

## Needs human decision

None.
