---
review: integration-1
slice: 0002-spec-and-state
kind: integration
target: slice 0002-spec-and-state
round: 1
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

The slice works as the spec says. I ran every story by hand in a
temporary repo, and all 37 requirements hold, except one edge of R10.
One medium problem: when the commit of `spec submit` fails, the undo
also unstages a `spec.md` the user had staged before, and still prints
`Nothing was changed.` (F1). Of the lead's document fixes, all are
verified except one leftover in plan.md Flow 8 (plan-5/F3, low).

## Checked

- Commits: cf38fa8..d3865fa (the whole slice against `main`; the two
  merges from `main`, 9eb0c84 and 0f0c1ed, have no conflicts and an
  empty `--remerge-diff`)
- Files read: docs/slices/0002-spec-and-state/spec.md, plan.md,
  tasks.md (tasks 04 and 16 in full, and the task-number lines),
  preview.md (P1 to P5), status.md, reviews/change-1.md,
  reviews/change-2.md, reviews/plan-5.md, docs/backlog.md (the three
  deferred rows), docs/templates/review.md,
  docs/slices/0001-slice-and-events/spec.md (R18, R24 to R33),
  internal/state/state.go, replay.go, moves.go, state_test.go,
  internal/eventlog/event.go, the diff of parse.go,
  internal/slice/save.go, lock.go, open.go, worktree.go, validate.go,
  submit.go, review.go, title.go, next.go, the diffs of status.go and
  repo.go, internal/review/review.go, placeholder.go,
  internal/cli/problems.go, review.go, submit.go, title.go,
  validate.go, template.go, the diffs of cli.go, status.go,
  cli_test.go, internal/template/template.go, and the list of test
  names in internal/slice and internal/e2e
- Ran:
  `git diff sliceflow/0002-spec-and-state/planning/R003 HEAD -- docs/slices/0002-spec-and-state/` (only status.md and reviews changed; spec, plan, tasks, preview are as approved)
  `git log --merges main..HEAD`, and `git show --remerge-diff` of both merges (no conflict)
  `gofmt -l cmd internal` (nothing), `go vet ./...` (pass), `go build ./...` (pass), `go test ./...` (pass)
  `go list -f` with the imports template, for state, spec, review, eventlog, cli, template (the dependency direction of Approach holds)
  the built binary in a temporary repo: S1, S2, S6 to S10, S12 to S17, S18 (pre-commit hook), S19, S20, S26 to S35, a refused submit from `review_passed`, round 2 and 3, a commit-msg hook that refuses `passed` (the error ends with `The steps before it were saved; spec.md is in_review.`), a busy lock with good and with unreadable content (10 s, then R7's text), no worktree, a merged slice (`Next:     -`, no `next` in JSON), 0001's spec (3 notes, as P2 says), this slice's own spec (0 errors), and a spec with tilde fences, a multi-line comment, a removed item, and CRLF line ends (all as R15, R16, R18 say)
  the same repo: a staged `spec.md`, a different `spec.md` on disk, a failing pre-commit hook, then `spec submit` (finds F1)
- Requirements: R1 to R37

## Tests

None.

## Previous findings

- change-1/F1: verified (fixed)
- change-1/F2: verified (fixed)
- change-1/F3: verified (fixed)
- change-1/F4: verified (fixed)
- change-2/F1: verified (fixed)
- change-2/F2: verified (fixed)
- change-2/F3: verified (fixed)
- plan-5/F2: verified (fixed)
- plan-5/F3: reopened. The lock (step 4), the Approach row, and Testing
  (task 17) are fixed, but plan.md Flow 8 step 3 (line 176) still
  starts "For `spec validate`, `spec submit`, and `slice title`", and
  leaves out `spec review`, although its own bracket now says "(submit,
  review, and title only)". R11 and Flow 5 are right, and the code
  follows them, so nothing breaks.
- plan-5/F4: verified (fixed)

## Findings

### F1. A failed submit unstages a spec.md the user had staged, and says nothing was changed

Level: medium
Where: internal/slice/save.go:159-160, internal/slice/save.go:229-232
Breaks: R10
Why: For `spec submit`, `spec.md` is in `op.Keep`, so it is in
  `gitFiles`. `git add -f` stages it, and when the commit fails the undo
  runs `git reset -q -- <gitFiles>`, which resets the index entry of
  `spec.md` to the last commit. That removes more than sliceflow
  staged: the user's own staged version is gone. I tried it: in the
  worktree, append a line to `spec.md` and `git add` it, change that
  line on disk, add a pre-commit hook with `exit 1`, run
  `sliceflow spec submit`. It prints
  `Error: could not save the event: git commit failed (exit code 1). Nothing was changed.`,
  but `git diff --cached` is now empty: the staged version of
  `spec.md` was dropped. The file on disk is kept, so the lead's text is
  not lost, but the staged copy is, and R10 says the undo unstages only
  "what it staged for it". A Ctrl-C before the commit takes the same
  path.
Fix: before `git add`, save the index entry of each Keep file
  (`git ls-files -s -- <file>`), and in the undo put it back with
  `git update-index --cacheinfo`, or reset a Keep file only when it was
  not staged before. Add a case to `TestSaveHookFails` or
  `TestSubmitOwnFiles` with a staged `spec.md`.

### F2. Small pieces of logic are written two or three times

Level: low
Where: internal/cli/problems.go:18, internal/slice/next.go:39,
  internal/slice/submit.go:54, internal/slice/review.go:60,
  internal/slice/title.go:31, internal/slice/validate.go:42-51,
  internal/slice/submit.go:84-91
Breaks: nothing
Why: `plural` is in both `cli` and `slice`. R20's error for no good
  `slice.created` is written out in full in three places. Counting the
  problems by level is done in both `Validate` and `Submit`. They agree
  today, but a wording change has to be made in each copy.
Fix: one `noCreatedError(name)` in `internal/slice`, and one helper
  that counts problems by level; `cli` can keep its own `plural`, or
  `slice` can export it.

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
- R10: fails, see F1
- R11: ok
- R12: ok
- R13: ok
- R14: ok
- R15: ok
- R16: ok
- R17: ok
- R18: ok
- R19: ok
- R20: ok, see F2
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
- R32: ok, see F2
- R33: ok
- R34: ok
- R35: ok
- R36: ok
- R37: ok

## Needs human decision

None.
