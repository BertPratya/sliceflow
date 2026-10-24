---
review: task-03-1
slice: 0001-slice-and-events
kind: task
target: task 03
round: 1
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

The skeleton is small and follows the brief and TD2, and all six Done
when tests are real. One high problem: any flag right after `slice`,
e.g. `sliceflow slice --colour`, prints the usage and exits with 0
instead of R38's wrong-use error with 2 (F1).

## Checked

- Commits: b887065
- Files read: cmd/sliceflow/main.go, internal/cli/cli.go,
  internal/cli/flags.go, internal/cli/output.go,
  internal/cli/cli_test.go, internal/cli/flags_test.go,
  internal/e2e/helpers_test.go, internal/e2e/cli_skeleton_test.go,
  internal/testrepo/testrepo.go, internal/slice/errors.go, the brief,
  spec.md (R18, R28, R32, R38, R39, S15), plan.md (Flow 3), tasks.md
  (tasks 07, 09, 11, 12, 13)
- Ran:
  gofmt -l cmd internal (prints nothing)
  go vet ./... (pass)
  go test ./... (pass, no GIT_DIR, GIT_WORK_TREE, or GIT_INDEX_FILE set)
  go build ./... (pass)
  the built binary with `slice --colour`, `slice --json`, `slice -x` (each: usage on stdout, exit 0)
  the built binary with `--colour`, `slice nope --help`, `create`, `--help extra` (as expected)
- Requirements: R28, R38, R39
- Interfaces: Run, command, register, parseFlags, errHelp, usageError,
  fail, printWarnings, run match the task exactly; slice.Warning and
  testrepo.Setup, testrepo.New used as task 01 and 04 made them

## Tests

- `TestParseFlagsAfterName`: ok
- `TestParseFlagsWrongUse`: ok
- `TestPrintWarnings`: ok
- `TestFailIndentsLines`: ok
- `TestHelpTop`: ok
- `TestUnknownCommand`: ok

## Previous findings

None.

## Findings

### F1. A flag after `slice` prints the usage and exits 0

Level: high
Where: internal/cli/cli.go:40-44
Breaks: R38
Why: The condition `args[0] == "slice" && len(args) > 1 &&
  strings.HasPrefix(args[1], "-")` treats every flag after `slice` as
  help, not only `-h` and `--help`. I ran the built binary:
  `sliceflow slice --colour`, `sliceflow slice --json`, and
  `sliceflow slice -x` each print the top usage on stdout and exit 0.
  R38 says an unknown flag is wrong use: `Error: unknown flag --colour.
  Run sliceflow --help.` on stderr, nothing on stdout, exit 2. The
  ASSUMED note names only `slice --help` and `slice -h`, so the code
  does more than the worker decided. No test covers it.
Fix: show the usage only when `args[1]` is `-h` or `--help` (use
  `isHelp`); give any other flag there `unknown flag <flag>` with exit 2,
  and add a case to `TestRunUnknown`.

### F2. `fail` indents lines that S15 prints without indent

Level: low
Where: internal/cli/output.go:18-25
Breaks: nothing
Why: `fail` indents each line after the first by 2 spaces, as the brief
  says, which fits R18's hook lines (task 09). But S15 (R32, task 11)
  prints `Give a name: ...` and `See all slices: ...` with no indent.
  Task 11's Interfaces list `fail`; if its worker passes S15's three
  lines to `fail`, the output breaks S15. The code is right for this
  task; the note is for the lead, so task 11's brief says S15 must not
  go through `fail` as one multi-line error.

### F3. A bad flag value gives a text R38 does not name

Level: low
Where: internal/cli/flags.go:58-60
Breaks: nothing
Why: `--json=yes` or `--all=x` makes `fs.Set` fail, and `parseFlags`
  returns `bad value for --json: ...`. R38 lists the only five texts for
  wrong use, and this is not one of them. The ASSUMED note says so. Only
  the `=value` form of a bool flag can reach it, so it is rare.
Fix: leave it, or have the lead decide which R38 text it maps to.

## Requirements check

None.

## Needs human decision

None.
