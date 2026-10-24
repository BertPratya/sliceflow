---
review: task-13-1
slice: 0001-slice-and-events
kind: task
target: task 13
round: 1
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

TestNoNetwork does what Done when says, and it fails when the binary runs
`git fetch origin`. One medium problem: the guard against a broken
`GIT_TRACE` never fires. The test's own `run` helper writes git lines into
the same trace file, so the test also passes when the binary's git writes
no trace at all (F1). One low gap: commands that reach a local or ssh
remote without fetch, such as `git ls-remote origin`, are not caught (F2).

## Checked

- Commits: 7d41700
- Files read: internal/e2e/no_network_check_test.go,
  internal/e2e/helpers_test.go, internal/testrepo/testrepo.go,
  internal/git/git.go, internal/slice/check.go,
  docs/slices/0001-slice-and-events/status.md (Findings: none deferred
  to task 13)
- Ran:
  gofmt -l cmd internal (prints nothing)
  go vet ./... (pass); GOOS=windows and GOOS=darwin go vet ./internal/e2e (pass)
  go test ./... (pass)
  go build ./... (pass)
  git 2.43.0 with GIT_TRACE, by hand in a temporary repo: fetch, an alias for fetch, `-c ... fetch`, ls-remote, and remote show, against a missing-folder remote and an https remote (trace line forms below)
  in a copy of the tree from `git archive HEAD`, outside the worktree: added `git fetch origin` to the remote check in check.go (TestNoNetwork fails on `trace: built-in: git fetch origin`)
  same copy: added `GIT_TRACE=0` to git.Run's env, with no other change (TestNoNetwork passes)
  same copy: `GIT_TRACE=0` in git.Run plus `git fetch origin` in check.go (TestNoNetwork passes)
  same copy: `git ls-remote origin` in check.go (TestNoNetwork passes)
- Requirements: R40
- Interfaces: `run`, `testrepo.New`, `testrepo.Git` used as the task says

Trace line forms, as asked by the lead. Every git process prints its own
`trace: built-in: git <sub>` line after git's own option handling, so
`-c a=b` and `-C dir` never show there. An alias shows as
`trace: alias expansion: f => fetch`, then `trace: built-in: git fetch`,
so it is caught. `run_command:` lines can start with
`unset GIT_PREFIX;` or `GIT_DIR=.git`; the parser then returns the wrong
word. That misses nothing today: each such child git also prints its own
`built-in:` or `exec:` line, and a remote helper shows as
`trace: exec: git-remote-https`, which the parser catches. `GIT_TRACE`
reaches every git the binary starts: git.Run uses `os.Environ()` plus two
variables, the test sets it with `t.Setenv` before `run`, and the binary
inherits the test's env. Git for Windows and Apple git use the same
trace.c, the same line text, and accept an absolute path with a drive
letter; Windows `\r` would be removed by `strings.Fields`.

## Tests

- `TestNoNetwork`: ok

## Previous findings

None.

## Findings

### F1. The empty-trace guard counts the test's own git lines

Level: medium
Where: internal/e2e/no_network_check_test.go:93-127, internal/e2e/helpers_test.go:48
Breaks: nothing
Why: The commit says the test "also fails when the trace holds no git
  command, so a broken GIT_TRACE cannot make it pass". But `run` calls
  `testrepo.Setup`, which runs `git rev-parse --local-env-vars` after
  `t.Setenv("GIT_TRACE", ...)`. So the trace always holds four lines from
  the test itself, and `seen` is never 0. I added `"GIT_TRACE=0"` to the
  env in git.Run and a `git fetch origin` in the remote check, in a copy of
  the tree: TestNoNetwork passed. So a later change that builds git's env
  without inheriting it would hide any fetch from this test.
Fix: require a line only the binary makes, e.g. a `worktree` subcommand
  from `slice create`, or skip `rev-parse` lines when counting `seen`.

### F2. A connection without fetch, e.g. git ls-remote origin, passes

Level: low
Where: internal/e2e/no_network_check_test.go:55-61
Breaks: nothing
Why: The remote is a local folder, so git never starts a `remote-`
  helper for it; it runs `git-upload-pack`. Only the fetch, pull, and push
  names can fire. I added `git ls-remote origin` to the remote check in a
  copy of the tree: TestNoNetwork passed. `git remote show origin` traces
  the same way. Done when lists only fetch, pull, push, and `remote-`,
  so the test meets it. R40's first sentence, no network connection,
  is wider than that.
Fix: also forbid `ls-remote`, `upload-pack`, and `receive-pack`. Or
  decide that the Done when list is enough.

## Requirements check

None.

## Needs human decision

None.
