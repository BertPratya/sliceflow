---
review: task-09-2
slice: 0002-spec-and-state
kind: task
target: task 09
round: 2
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

Both findings of round 1 are fixed. R19's path now comes from the real
worktree, relative to the main checkout's root, from any folder. S6's test
uses S6's exact line numbers. No new findings.

## Checked

- Commits: 26602dc, 07bb7d5 (`git diff 26602dc 07bb7d5`)
- Files read: internal/slice/validate.go, internal/slice/validate_test.go,
  internal/e2e/validate_command_test.go, internal/slice/repo.go
  (FindRepo, how `Root` is set), internal/slice/worktree.go
  (findWorktree), docs/slices/0002-spec-and-state/reviews/task-09-1.md,
  status.md (Lead's decisions, task 09), the commit messages and their
  ASSUMED lines, the brief
- Ran:
  `TestValidateMovedWorktree` against the round 1 code (26602dc, in a temporary detached worktree, with the new test file): fails, got `.sliceflow/worktrees/demo/...`, want `.worktrees/elsewhere/demo/...`
  go test ./internal/slice/ ./internal/e2e/ ./internal/cli/ on 07bb7d5 (pass)
  built the binary; in a temporary repo, moved the worktree of `demo` to `.worktrees/else/demo` and deleted its `spec.md`: `spec validate demo` on `main`, in a subfolder of `main`, `spec validate` in the moved worktree, and `spec validate demo` in its `docs/` subfolder all print `Error: .worktrees/else/demo/docs/slices/demo/spec.md not found. Nothing was changed.`, exit 1
  moved the worktree outside the repo: on `main`, in the worktree, and through a symlink to the repo, all print `Error: ../ext/docs/slices/demo/spec.md not found. Nothing was changed.`, exit 1
  moved it back to `.sliceflow/worktrees/demo`: `spec validate` in the worktree prints S13's line, exit 1
- Requirements: R19, R13, R14, R37 (S6 to S8 tests)
- Interfaces: `specNotFoundError(r Repo, wt, name string) *Error` is
  private in `internal/slice`, as the lead decided, and takes only what
  task 14 has after `openSlice` (the repo, the worktree, the name)

On the lead's points:
- `TestValidateMovedWorktree` fails on the round 1 code, so it proves F1's
  fix.
- `r.Root` is the first entry of `git worktree list --porcelain`, which is
  always the main checkout, also when the command runs in a linked
  worktree. The binary runs above show the path starts from the main
  checkout's root both in the moved worktree and on `main` by name.
  `findWorktree` also gets its path from git, so both sides of
  `filepath.Rel` have the same form, also through a symlink.
- Task 14 can call `specNotFoundError(r, wt, name)` with the `wt` that
  `openSlice` returns. The ASSUMED fallback to the full path, when
  `filepath.Rel` fails (another drive on Windows), changes no document:
  R19 does not say what to print then, and no relative path exists.

## Tests

- `TestValidateProblems`: ok
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

- task-09-1/F1: verified (fixed)
- task-09-1/F2: verified (fixed)

## Findings

None.

## Requirements check

None.

## Needs human decision

None.
