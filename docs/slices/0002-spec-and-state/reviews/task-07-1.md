---
review: task-07-1
slice: 0002-spec-and-state
kind: task
target: task 07
round: 1
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

The task covers R4, R29, and R31. `slice status` and `--json`, in both
`slice status` and `slice list`, match S1 to S5 and R31 character for
character, and the tests are real. One low finding: the CLI keeps a
second copy of R29's refused-mark rule, and no test covers the parts of
that copy that differ from a plain "refused" check (F1).

## Checked

- Commits: be44aaf
- Files read: internal/cli/status.go, internal/cli/list.go,
  internal/slice/status.go, internal/slice/next.go,
  internal/state/state.go, internal/state/moves.go (Refused),
  internal/e2e/status_spec_test.go, internal/e2e/status_command_test.go,
  internal/e2e/slice_list_test.go, docs/slices/0002-spec-and-state/plan.md
  (Core API), the brief and its Lead's note
- Ran:
  go test ./internal/e2e/ -run 'TestStatus|TestList' -count=1 (pass)
  go test ./... (pass)
  go build ./cmd/sliceflow, then in a temporary repo: slice create
    login-google --title "Login with Google", slice status and
    slice status --json (S1 and S2 output, exact), log lines for a passed
    review 1 (S20's two lines), spec.md edited on disk (S28's two lines),
    a refused submit 2 errors 1 warning (`passed review 1, but spec.md was
    changed after it (submit refused: 2 errors, 1 warning)`, and JSON with
    `submit_refused` and `changed_after_review`), spec.md restored (no
    mark, `write the plan`), slice list --json (same docs and next)
- Requirements: R4, R29, R31
- Interfaces: `SliceStatus.State`, `SpecText`, `Next`,
  `ChangedAfterReview` are used as the task says; `Merged` is used too,
  as the Lead's note allows

## Tests

- `TestStatusSpecDraft`: ok
- `TestStatusSpecJSON`: ok
- `TestStatusNotAllowedLine`: ok
- `TestStatusReviewChanges`: ok
- `TestStatusMerged`: ok
- `TestStatusJSONReview`: ok

## Previous findings

None.

## Findings

### F1. The CLI repeats R29's refused-mark rule, and its extra conditions are not tested

Level: low
Where: internal/cli/status.go:160-162, internal/slice/next.go:86
Breaks: nothing
Why: `docsJSON` checks `d.Refused != nil && !st.Merged && (d.Status ==
  "draft" || st.ChangedAfterReview)`, a copy of the check in `specNext`.
  Today the two agree: `ChangedAfterReview` is only set for a
  `review_passed` spec that is not merged, so it equals `specNext`'s
  masked `changed`. I checked this by hand in a temporary repo. But the
  rule now lives in two places, and the copy is outside the core that the
  plan's Core API says a later server calls (SD13 of 0001). And no test
  covers the cases where the copy differs from a plain `d.Refused != nil`
  check: a refused submit on a `review_passed` spec that is not changed,
  or on a merged slice. `TestStatusJSONReview` only has a refused draft.
  So if one copy changes, nothing fails. This is not a break of the
  AGENTS.md rule "Shared Go logic goes in internal/": `internal/cli` is in
  `internal/`.
Fix: let the core set the mark, e.g. a `SubmitRefused *state.Counts`
  field in `SliceStatus` that `specNext` fills only when the mark shows,
  and have `docsJSON` copy it. Or add a JSON case for a refused submit on
  an unchanged `review_passed` spec, with no `submit_refused`.

## Requirements check

None.

## Needs human decision

None.
