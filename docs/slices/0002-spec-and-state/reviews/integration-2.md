---
review: integration-2
slice: 0002-spec-and-state
kind: integration
target: slice 0002-spec-and-state
round: 2
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

integration-1/F1 is fixed. When a save fails, the undo now puts back
what the index held for the event's files. This works for every command
that saves an event: `spec submit`, both saves of `spec review`, and
`slice title`. Since integration-1, only task 18's three files and the
backlog changed. All 37 requirements hold. No new findings.

## Checked

- Commits: d3865fa..ac43737 (since integration-1: 7aea506 and 715ebe3
  of task 18, the lead's status and review commits, and the merge
  225d006 from `main`, backlog only)
- Files read: docs/slices/0002-spec-and-state/spec.md (R8, R10),
  tasks.md (task 18), status.md (Findings, Lead's decisions),
  reviews/integration-1.md, reviews/task-18-2.md, docs/backlog.md (the
  rows for integration-1/F2 and task-18-1/F2), docs/templates/review.md,
  internal/slice/save.go in full, the calls of `saveEvent` in
  internal/slice/submit.go, review.go (both saves), and title.go, the
  diff of internal/slice/save_test.go and
  internal/e2e/submit_review_failures_test.go
- Ran:
  `git diff sliceflow/0002-spec-and-state/planning/R003 HEAD -- docs/slices/0002-spec-and-state/` (spec, plan, preview unchanged; tasks.md only gained task 18 at the end)
  `git diff d3865fa HEAD --stat` (code: only save.go, save_test.go, submit_review_failures_test.go, all in task 18's touches)
  `git log --merges main..HEAD`, and `git show --remerge-diff 225d006` (no conflict; it adds two backlog rows)
  `gofmt -l cmd internal` (nothing), `go vet ./...` (pass), `go build ./...` (pass), `go test ./...` (pass)
  by hand, a built binary in a temporary repo with a `pre-commit` hook that exits 1: `spec submit` with `spec.md` staged and then changed on disk (exit 1, `Nothing was changed.`, the staged text is the same, the disk text is kept); `slice title` with the log staged (exit 1, index as before); `spec review` failing on its first save (exit 1, `reviews/` gone, a staged `other.txt` kept) and on its second save (exit 1, `review 1 started` committed, `reviews/spec-1.md` and the index as in that commit, `other.txt` still staged)
- Requirements: R1 to R37

## Tests

None.

## Previous findings

- integration-1/F1: verified (fixed)
- integration-1/F2: verified (deferred)

## Findings

None.

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

## Needs human decision

None.
