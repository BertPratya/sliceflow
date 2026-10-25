---
review: task-07-2
slice: 0002-spec-and-state
kind: task
target: task 07
round: 2
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

task-07-1/F1 is fixed with tests, as the lead decided. `TestStatusJSONReview`
now fails if either copy of R29's refused-mark rule loses its `merged` or
its draft-or-changed condition; I checked all four cases by changing the
code. No new findings.

## Checked

- Commits: be44aaf, 290c964 (`git diff be44aaf 290c964`)
- Files read: internal/e2e/status_spec_test.go, internal/e2e/slice_list_test.go
  (mergeAndRemove), internal/cli/status.go (docsJSON),
  internal/slice/next.go (specNext), reviews/task-07-1.md, status.md (Lead's
  decisions), the brief, and the commit message of 290c964 with its
  `ASSUMED` note
- Ran:
  go test ./internal/e2e/ -run 'TestStatus|TestList' -count=1 (pass)
  in a temporary worktree, four one-line changes, each with
    go test ./internal/e2e/ -run TestStatusJSONReview -count=1: drop the
    draft-or-changed condition in docsJSON (fails at line 199), drop
    `!st.Merged` in docsJSON (fails at line 234), drop the draft-or-changed
    condition in specNext (fails at line 199), drop `!merged` in specNext
    (fails at line 238); the temporary worktree removed afterwards
- Requirements: R4, R29, R31
- Interfaces: `SliceStatus.State`, `SpecText`, `Next`, `ChangedAfterReview`
  unchanged since round 1

## Tests

- `TestStatusSpecDraft`: ok
- `TestStatusSpecJSON`: ok
- `TestStatusNotAllowedLine`: ok
- `TestStatusReviewChanges`: ok
- `TestStatusMerged`: ok
- `TestStatusJSONReview`: ok

## Previous findings

- task-07-1/F1: verified (fixed). The fix is the second option the finding named (tests, no new `SliceStatus` field), as the lead decided in status.md. A refused submit on an unchanged `review_passed` spec and on a merged draft slice now have JSON and plain checks, and each of the four ways the two copies can drift makes the test fail.

## Findings

None.

## Requirements check

None.

## Needs human decision

None.
