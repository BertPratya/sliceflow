---
review: integration-1
slice: 0001-slice-and-events
kind: integration
target: slice/0001-slice-and-events
round: 1
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

The slice works as the spec says: every requirement from R1 to R43 holds
in the code, the tests, and my own runs of the binary. One high problem,
the defect the lead handed over: `removeWorktree` runs `git worktree prune`,
which removes other worktrees' stale entries and still leaves this run's
locked entry (F1). One low: small pieces of logic are copied between
files (F2). task-09-1/F3 (`slice.Interrupted`) is fine as wontfix: the CLI
needs R18's exact text when `FindRepo` is interrupted (task 09, step 4),
and the export adds no behavior.

## Checked

- Commits: main...HEAD (b7b34d1..0408914), all 13 tasks, and the merges
  b7b34d1 and 28e9abb
- Files read: docs/slices/0001-slice-and-events/spec.md, plan.md, tasks.md
  (task 09 in full, and the Interfaces of the others), status.md (Findings,
  Lead's decisions); cmd/sliceflow/main.go; internal/git/git.go;
  internal/slice/repo.go, errors.go, base.go, name.go, check.go, create.go,
  undo.go, status.go, list.go; internal/eventlog/event.go, parse.go;
  internal/cli/cli.go, flags.go, output.go, create.go, status.go, list.go;
  internal/template/template.go; internal/testrepo/testrepo.go;
  internal/e2e/helpers_test.go; the list of every test name
- Ran:
  gofmt -l cmd internal (prints nothing)
  go vet ./... (pass)
  go build ./... (pass)
  go test -count=1 ./... (pass)
  GOOS=windows go vet ./... (pass)
  GOOS=darwin go vet ./... (pass)
  git diff sliceflow/0001-slice-and-events/planning/R004 HEAD on spec.md, plan.md, tasks.md, preview.md (only task 12's touches gained internal/cli/status.go: allowed)
  git log --merges main..HEAD (b7b34d1, 28e9abb); git show --remerge-diff on both (empty: no conflicts; they bring only docs, agent files, and templates)
  diff docs/templates/spec.md internal/template/spec.md (only the title: line and its rules differ, as R7 says)
  the built binary in temporary repos (scratchpad, test git config): S1, S2 (title, --json), S4 (create inside a worktree), S5, S6, S7 (with a configured remote, also a remote named team/a), S8, S9, S10, S11, S12, a pre-commit hook with output (R18's indented lines), a failing post-checkout hook, SIGINT and SIGTERM during a slow pre-commit hook (nothing left, reason interrupted), S13 to S19, S20 to S25, R43 (slice/My_Thing, slice/team/login), --all with a merged slice, from a subfolder, --from master~1, every wrong use of R38, --help (all as the spec says)
  git 2.43: a locked worktree whose folder is gone, then git worktree prune (removes another stale entry, keeps the locked one; see F1); then git worktree remove --force --force <path> (removes only that entry)
- Requirements: R1 to R43
- Interfaces: Create, Status, List, CreateOptions, Warning, Error, Repo
  (with Now) match plan.md, Core API; one export not in tasks.md,
  slice.Interrupted (task-09-1/F3, accepted above)

## Tests

None.

## Previous findings

None.

## Findings

### F1. The undo's git worktree prune removes other worktrees' entries, and keeps this run's

Level: high
Where: internal/slice/undo.go:164-172
Breaks: plan.md, Shared rules, Undo on failure ("It never removes a thing
  that is not on the list"; an old entry whose folder the user deleted
  "is not this run's, and it stays")
Why: when `git worktree remove --force <wt>` fails and the folder is gone,
  `removeWorktree` runs `git worktree prune`. With git 2.43, `remove --force`
  of a missing folder succeeds, so this branch runs when this run's
  worktree is locked (or `remove` failed after it deleted the folder). I
  tried it: worktree `old` with its folder deleted by hand, worktree `mine2`
  locked with its folder deleted. `git worktree remove --force mine2` fails
  ("cannot remove a locked working tree"); then `git worktree prune`
  removes `old`'s entry, which is not on the list, and keeps `mine2`'s,
  because prune skips locked entries. Then `git branch -D` fails ("used by
  worktree"), so the user gets `Could not remove: branch slice/<name>.`,
  and a stale entry of someone else is gone. An unlocked entry for a
  worktree on a drive that is not mounted loses its admin files the same
  way. The trigger is narrow, but the rule it breaks is the main promise
  of the undo, and no test covers this branch.
Fix: replace the prune with `git worktree remove --force --force <wt>`,
  which removes only this run's entry (I checked: it works on a locked
  entry with a missing folder, and leaves the other stale entry).

### F2. Small pieces of logic are copied between files

Level: low
Where: internal/cli/create.go:70-73, internal/cli/status.go:103-106,
  internal/cli/create.go:87-90, internal/cli/status.go:74-77,
  internal/slice/check.go:23-30, internal/slice/status.go:90-97,
  internal/slice/status.go:154
Breaks: nothing
Why: `runCreate` declares its own `fromJSON` type, which shadows the
  package's `fromJSON` in status.go. The 7-character short commit is cut in
  both create.go and status.go. The "only an existing local branch is a
  base" check (`show-ref --verify --quiet refs/heads/<base>`) is written
  twice, in `checkCreate` and `baseWarning`. `readSlice` repeats the time
  layout that eventlog keeps in `timeLayout`. Each copy is right today;
  a later change to one can miss the other.
Fix: one helper each, e.g. `shortCommit` in internal/cli, and a
  `baseExists` in internal/slice that both callers use.

## Requirements check

- R1: ok
- R2: ok
- R3: ok
- R4: ok
- R5: ok
- R6: ok
- R7: ok
- R8: ok
- R9: ok
- R10: ok
- R11: ok
- R12: ok
- R13: ok
- R14: ok
- R15: ok
- R16: ok
- R17: ok
- R18: ok
- R19: ok
- R20: ok
- R21: ok
- R22: ok
- R23: ok
- R24: ok
- R25: ok
- R26: ok
- R27: ok
- R28: ok
- R29: ok
- R30: ok
- R31: ok
- R32: ok
- R33: ok
- R34: ok
- R35: ok
- R36: ok
- R37: ok
- R38: ok
- R39: ok
- R40: ok
- R41: ok
- R42: ok
- R43: ok

## Needs human decision

None.
