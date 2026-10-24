---
review: task-04-1
slice: 0001-slice-and-events
kind: task
target: task 04
round: 1
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

The task covers R2, R3, R12, R13, R14, R21, and R22's core part, with the
exact messages, and the names match Interfaces. The tests are real: each
one failed when I broke the code it checks, in a copy of the package. One
low finding: a repo git refuses to open is reported as "not inside a git
repository" (F1).

## Checked

- Commits: 5550758
- Files read: internal/slice/errors.go, internal/slice/repo.go,
  internal/slice/name.go, internal/slice/base.go,
  internal/slice/repo_test.go, internal/slice/name_test.go,
  internal/slice/base_test.go, internal/git/git.go,
  internal/testrepo/testrepo.go, the brief, the commit's ASSUMED notes,
  status.md (no finding deferred to task 04)
- Ran:
  gofmt -l internal (prints nothing)
  go vet ./... (pass)
  go test ./internal/... (pass)
  mutation runs of go test ./internal/slice/ in a scratch copy, not in the worktree: Root from rev-parse --show-toplevel (TestFindRepoFromWorktree fails); no-commits check removed (TestFindRepoNoCommits fails); reserved check removed (TestCheckName fails); sliceflow.base ignored (TestBaseBranch fails)
- Requirements: R2, R3, R12, R13, R14, R21, R22
- Interfaces: Error, Error.Error, Warning, Repo, FindRepo, CheckName,
  BaseBranch match the task; FindRepo sets Now to time.Now; BaseBranch's
  three source texts match exactly

## Tests

- `TestCheckName`: ok
- `TestFindRepoNotARepo`: ok
- `TestFindRepoNoCommits`: ok
- `TestFindRepoFromWorktree`: ok
- `TestBaseBranch`: ok

## Previous findings

None.

## Findings

### F1. A repo git refuses to open is reported as not a repo

Level: low
Where: internal/slice/repo.go:23-28
Breaks: nothing
Why: FindRepo turns every failed `rev-parse --show-toplevel` into R12's
  message. In a repo owned by another user, git exits 128 with "detected
  dubious ownership", and sliceflow then prints "not inside a git
  repository", which hides the real cause. The ASSUMED note covers this
  choice, and the plan says sliceflow does not read git's error text, so
  the two cases cannot be told apart by exit code. A rare case; it does not
  break R12's stated scenario.

## Requirements check

None.

## Needs human decision

None.
