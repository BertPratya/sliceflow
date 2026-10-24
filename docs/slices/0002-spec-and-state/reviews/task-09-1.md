---
review: task-09-1
slice: 0002-spec-and-state
kind: task
target: task 09
round: 1
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

`spec validate` works as Flow 2 says. The output matches S6 to S13 and
S36 character for character, and the tests are real. One high problem:
R19's error always names `.sliceflow/worktrees/<name>/...`, even when the
slice's worktree is somewhere else, so it can name a path that does not
exist (F1). One low: S6's test does not use S6's line numbers (F2).

## Checked

- Commits: 26602dc
- Files read: internal/slice/open.go, internal/slice/validate.go,
  internal/slice/validate_test.go, internal/cli/validate.go,
  internal/cli/problems.go, internal/e2e/validate_command_test.go,
  internal/slice/check.go (existsWhere, baseExists), internal/slice/status.go
  (noSliceError, currentSliceName, readSlice), internal/slice/worktree.go,
  internal/slice/base.go, internal/slice/repo.go, internal/e2e/slice_list_test.go
  (mergeAndRemove), docs/slices/0001-slice-and-events/spec.md (R15, R24,
  R41, SD10), docs/templates/tasks.md, the brief
- Ran:
  go test ./internal/slice/ ./internal/e2e/ ./internal/cli/ (pass)
  built the binary; in a temporary repo, moved the slice's worktree with `git worktree move` and deleted `spec.md`, then ran `spec validate demo` and `spec validate` (both print `Error: .sliceflow/worktrees/demo/docs/slices/demo/spec.md not found. Nothing was changed.`, a path that does not exist)
  in the same repo, a slice whose branch is only `refs/remotes/origin/slice/demo`: `spec validate demo` and `slice status demo` (both print R33's `no slice named demo`, exit 1)
  `spec validate --json --strict x y` (unknown flag --json, exit 2), `spec validate a b` (too many arguments, exit 2), `spec validate main~1` (R14 of 0001, exit 1)
- Requirements: R11, R12, R13, R14, R18, R19, R36, R37; R13, R14, R21,
  R32, R33, R38 of 0001-slice-and-events
- Interfaces: openSlice, Validation, Validate, printProblems match the
  task; currentSliceName, findWorktree, BaseBranch, git.RefExists,
  spec.Parse, spec.Check used as named

On the lead's points:
- A slice that exists only on a remote: `openSlice` decides "exists" the
  same way 0001's `readSlice` does (local branch, else the folder on the
  base branch, as R24 of 0001 says), so `spec validate` and
  `slice status` agree: both give R33's error. `existsWhere` is for R15
  of 0001 (a name already taken, for `slice create`), which names remote
  branches on purpose; SD10 of 0001 keeps remote branches out of what
  sliceflow shows. R11's order (R33, then R11) is followed. So the two
  ways should not agree, and this is not a finding.
- S6's line numbers: see F2.
- `printProblems` and `openSlice` are fit for tasks 12, 14, and 16, but
  F1's path code sits in `Validate`; task 14 will need the same path for
  R19, so the fix should be shared.

## Tests

- `TestValidateProblems`: ok, see F2
- `TestValidateWarningOnly`: ok
- `TestValidateStrict`: ok
- `TestValidateByName`: ok
- `TestValidateNoBranch`: ok
- `TestValidateSkipsCode`: ok
- `TestValidateNoWorktree`: ok
- `TestValidateNoSpec`: ok
- `TestValidateJSON`: ok
- `TestValidateMerged`: ok
- `TestOpenSliceOrder`: ok

## Previous findings

None.

## Findings

### F1. R19's path is fixed to .sliceflow/worktrees, not the real worktree

Level: high
Where: internal/slice/validate.go:35-37
Breaks: R19
Why: R19 says the path is "the path from the root of the main checkout".
  `Validate` builds it from fixed parts,
  `.sliceflow/worktrees/<name>/docs/slices/<name>/spec.md`, while it reads
  the file from `wt`, which `findWorktree` finds wherever the worktree is.
  I moved a slice's worktree with `git worktree move` and deleted
  `spec.md`: both `spec validate demo` on `main` and `spec validate` in
  the moved worktree printed
  `Error: .sliceflow/worktrees/demo/docs/slices/demo/spec.md not found. Nothing was changed.`,
  a path that does not exist. S13's value is right only because the
  worktree is in the default place. Task 14 (`spec submit`) needs the
  same R19 path.
Fix: build the path from `wt`: `filepath.Rel(root, specPath)`, then
  `filepath.ToSlash` (R41 of 0001), in a helper task 14 can reuse; add a
  unit test with a moved worktree.

### F2. S6's test finds its line numbers instead of using 42, 44, and 140

Level: low
Where: internal/e2e/validate_command_test.go:88-93, 95-110
Breaks: docs/templates/tasks.md, "A scenario from the spec becomes an
  e2e test with the scenario's exact values"
Why: `s6Lines` searches the test's own spec text for the line numbers,
  so the test prints e.g. `spec.md:39:` where S6 shows `spec.md:42:`. The
  behavior is still proven (format, order, counts, exit code, no commit),
  and the spec does not give the file, so the numbers cannot be wrong in
  the code. But the test could pad its spec text so R7 is on line 42, R8
  on line 44, and `## Open questions` on line 140, and then check S6's
  four lines as written.
Fix: pad the test spec (e.g. more text lines under Summary and Why) to
  put the three items on lines 42, 44, and 140, and compare with S6's
  literal lines.

## Requirements check

None.

## Needs human decision

None.
