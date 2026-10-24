---
review: task-08-1
slice: 0001-slice-and-events
kind: task
target: task 08
round: 1
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

The checks run in R11's order, print the spec's exact words, and make
nothing when one fails. The tests are real. One high problem: the R16
check accepts revision syntax, so `--from main~1` or `--from main@{0}`
creates a slice from a base that is not a local branch (F1). The
`ASSUMED` folder lookup on `refs/heads/<base>` is a small departure from
the plan's "commit id once" rule; it only matters if the base moves
during `create` (F2).

## Checked

- Commits: 9cdc0eb
- Files read: internal/slice/check.go, internal/slice/create.go,
  internal/slice/base.go, internal/slice/check_test.go,
  internal/e2e/create_checks_test.go, internal/git/git.go (Run,
  RefExists, Show), docs/slices/0001-slice-and-events/plan.md (Flow 1,
  Flow 2, Git arguments)
- Ran:
  go test ./... (pass)
  built the binary; in a temporary repo with two commits,
    `slice create aa --from main~1` (exit 0, slice made, event log has
    `"branch":"main~1"`)
  `slice create bb --from main@{0}` (exit 0, slice made)
  `slice create cc --from HEAD` (exit 1, R16 error, as expected)
  `slice create "My Slice"` outside a repo (R12 error) and in a repo
    with no commits (R13 error): R12 and R13 come before R14
  `slice create zz --from main~1` with a branch `slice/zz` (R15 error)
- Requirements: R2, R11, R12, R13, R14, R15, R16, R17, R42
- Interfaces: `checkCreate(ctx, r, opts, base, source) error` matches the
  task; uses `CheckName`, `git.RefExists`, `git.Run`

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

None.

## Findings

### F1. The base check accepts revision syntax, not only a local branch

Level: high
Where: internal/slice/check.go:20-26
Breaks: R16
Why: The check runs `git rev-parse --verify --quiet
  refs/heads/<base>^{commit}`. Git reads `refs/heads/main~1` and
  `refs/heads/main@{0}` as revisions, and they resolve to commits. I
  ran `slice create aa --from main~1` in a temporary repo: exit 0, the
  slice was made from the parent commit, and the event log says
  `"from":{"branch":"main~1",...}`. `main~1` is not a local branch, so R16
  must print `Error: base branch main~1 does not exist (from --from).
  Nothing was changed.` The same goes for `git config sliceflow.base`.
  The plan's Flow 2 step 4 ("does not resolve to a commit") reads the
  same way, so the code follows the plan's words, but the spec decides.
  Later commands would also store and read a base "branch" that is not
  a branch.
Fix: check the exact ref, e.g. `git show-ref --verify --quiet
  refs/heads/<base>`, or `git check-ref-format refs/heads/<base>` first,
  then resolve the commit. Add a unit case for `main~1`.

### F2. The folder lookup reads the base ref, not the commit id taken once

Level: low
Where: internal/slice/check.go:67, internal/slice/create.go:36-40
Breaks: nothing
Why: The plan's Git arguments say the base is turned into a commit id
  once, at the start, and every later step uses it; Flow 2 step 5 says
  `git cat-file -e <commit>:docs/slices/<name>`. The worker's `ASSUMED`
  uses `refs/heads/<base>:docs/slices/<name>`, and `Create` resolves the
  commit only after the checks. Flow 1 also puts the checks (step 3)
  before the commit id (step 4), so the plan does not fully agree with
  itself, and the worker's choice is a fair reading. The effect is
  small: the base is read twice, a few milliseconds apart. If another
  process moves the base in that window (for example, merges slice
  `<name>` into it), the folder check sees the old commit while the new
  slice starts from the new one. No visible behavior changes in normal
  use, and no document needs to change.
Fix: resolve the commit once in the R16 check (it fits with F1's fix),
  and pass it to the folder lookup and to the rest of `Create`.

## Requirements check

None.

## Needs human decision

None.
