---
review: task-06-1
slice: 0002-spec-and-state
kind: task
target: task 06
round: 1
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

The task covers R1, R3, R29, R30, and R34, and the Done when tests are
real. R29's table and refused mark match the spec, and the ASSUMED lines
change nothing. One high problem: a slice whose worktree folder was
deleted without `git worktree prune` makes `slice status` and `slice list`
fail with a `chdir` error (F1).

## Checked

- Commits: 2e9534f (`git diff 2e9534f~1 2e9534f`), and its ASSUMED lines
- Files read: internal/slice/next.go, internal/slice/next_test.go,
  internal/slice/status.go, internal/slice/status_test.go,
  internal/slice/worktree.go, internal/slice/worktree_test.go,
  internal/slice/list.go (the readSlice call), internal/eventlog/parse.go,
  internal/eventlog/parse_test.go, internal/e2e/status_command_test.go,
  internal/e2e/slice_list_test.go, internal/state/state.go,
  internal/state/moves.go, internal/state/replay.go,
  internal/state/state_test.go, internal/git/git.go, the brief
- Ran:
  go vet ./... (pass)
  go test -count=1 ./... (pass)
  git worktree list --porcelain -z in a temporary repo after rm -rf of a worktree folder (the worktree is still listed, marked prunable)
  go build ./cmd/sliceflow, then in a temporary repo: slice create demo, a committed log that ends in review 1 passed, rm -rf of the worktree folder, slice status demo and slice list (both: Error: chdir ...: no such file or directory, exit 1)
- Requirements: R1, R3, R29, R30, R34
- Interfaces: SliceStatus fields State, SpecText, Next,
  ChangedAfterReview, Merged; findWorktree, specChanged, specNext, and
  currentSliceName match the task's signatures

## Tests

- `TestSpecNext`: ok
- `TestStatusState`: ok
- `TestSpecChanged`: ok
- `TestFindWorktree`: ok

## Previous findings

None.

## Findings

### F1. A deleted worktree folder makes slice status and slice list fail

Level: high
Where: internal/slice/worktree.go:13-32, internal/slice/next.go:22-30,
  internal/slice/status.go:171-181
Breaks: R30, plan Flow 1 step 3
Why: `git worktree list --porcelain -z` still lists a worktree whose
  folder was removed by hand, with a `prunable` field, until
  `git worktree prune` runs. I checked this in a temporary repo.
  findWorktree returns that path with true. specChanged then runs git
  with that folder as its directory. The process cannot start, so the
  error is a `chdir` error, not a `*git.Error`, and specChanged returns
  it. readSlice returns it too. With the built binary, a slice whose log
  ends in review 1 passed, after `rm -rf` of its worktree folder, gives
  `Error: chdir <path>: no such file or directory` and exit code 1 for
  `slice status demo`. `slice list` fails the same way, for every slice.
  R30 says that with no worktree, sliceflow does not compare. Flow 1
  step 3 says any other failure means not compared, and the command goes
  on.
Fix: in findWorktree, skip a record that has a `prunable` field, or whose
  folder does not exist. Add that case to `TestFindWorktree`.

### F2. findWorktree returns git's form of the path

Level: low
Where: internal/slice/worktree.go:24-28
Breaks: nothing
Why: on Windows, git prints worktree paths with forward slashes, e.g.
  `C:/repo/.sliceflow/worktrees/x`. Such a path works as a process
  folder and in `filepath.Join`, which cleans it, so nothing breaks
  today, on any of the three systems. But tasks 09, 11, 12, and 14 use
  findWorktree. A later task that compares this path with one from
  `filepath`, e.g. `Created.Worktree`, would see two forms of one folder.
Fix: return `filepath.Clean(dir)`, so callers always get the system's
  form.

### F3. No test for a refused submit after a pass, with no change

Level: low
Where: internal/slice/next_test.go:24-41
Breaks: nothing
Why: R29 shows the refused mark only when `spec.md` is `draft` or changed
  after its review. The code does this right: `review_passed` with
  Refused set and changed false gives `review_passed (review 1, placeholder)`
  and `write the plan`. But no row checks it. If the check `|| changed`
  at next.go:80 were dropped for `review_passed`, every row would still
  pass.
Fix: add a row: `review_passed`, refused, changed false.

### F4. A refused submit with 0 errors and 0 warnings gives an empty count

Level: low
Where: internal/slice/next.go:83-91
Breaks: nothing
Why: eventlog.Parse accepts a refused `spec.submitted` with errors 0 and
  warnings 0. specNext then gives `draft (submit refused: )`. sliceflow
  never writes such a line itself, so only a hand-made log shows it.
  The spec does not say what to show here.

## Requirements check

None.

## Needs human decision

None.
