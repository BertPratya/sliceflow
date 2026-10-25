---
review: change-1
slice: 0002-spec-and-state
kind: change
target: docs/slices/0002-spec-and-state/tasks.md
round: 1
reviewer: doc-reviewer (claude opus)
---

# Review

## Summary

The change to task 04 is sound: `Doc.Preamble` lets `Check` see markers
above the first `##` heading, as R15 asks, and no task that can run at
the same time touches `internal/spec/`. One medium problem: Files says
`parse.go` changes for task-04-1/F1 only, but the F3 fix changes it too
(F1). The rest is low.

## Checked

- Files read: docs/slices/0002-spec-and-state/tasks.md (the diff from
  sliceflow/0002-spec-and-state/planning/R001; tasks 03, 04, 05, and the
  touches and depends_on of every task), spec.md (R15, R16, R17, R18),
  plan.md (Components, Spec reader), status.md, reviews/task-04-1.md,
  docs/templates/tasks.md, docs/templates/review.md, AGENTS.md,
  internal/spec/doc.go, the working-tree diff of internal/spec/parse.go,
  internal/spec/errors.go:194, internal/spec/parse_test.go
  (TestParsePreamble), the test names added in internal/spec/errors_test.go
- Ran:
  git diff sliceflow/0002-spec-and-state/planning/R001 -- docs/slices/0002-spec-and-state/tasks.md (task 04 only)
  git diff --stat (5 files in internal/spec, all in task 04's touches)
  go test ./internal/spec/ (pass)
- Requirements: R15, R18
- Interfaces: `Doc.Preamble Section` against task 03's `Doc`; `Check`,
  `Problem`, `Doc` as used by tasks 05, 09, and 14; `spec.Parse` and
  `spec.Check` in tasks 09 and 14

## Tests

None.

## Previous findings

None.

## Findings

### F1. Files says parse.go changes for F1 only, but the F3 fix changes it too

Level: medium
Where: tasks.md → Task 04 → Files
Breaks: task 04
Why: Files says `parse.go`, `parse_test.go`, and `doc.go` change "for
  review task-04-1/F1 only". The working-tree diff of parse.go also
  changes `BadMarkers` text from `clean` to `sc[i].vis`, which is the
  fix for task-04-1/F3 (inline code kept in the "is not a marker" text).
  That changes what task 03's `Mark.Text` holds for a bad marker. A code
  reviewer who reads the brief would flag it as outside the task.
Fix: write "for review task-04-1/F1 and F3" and say what F3 changes:
  `BadMarkers` text keeps inline code. Or move the F3 fix out of
  parse.go.

### F2. Done when names no test for a marker above the first heading

Level: low
Where: tasks.md → Task 04 → Done when
Breaks: nothing
Why: The change adds `Doc.Preamble` and its checks, but Done when lists
  only the four first tests. The code has `TestParsePreamble` and
  `TestErrorMarkerBeforeFirstSection`, but the code reviewer's Tests
  section checks only the tests that Done when names, so these may not
  be read.
Fix: add `TestErrorMarkerBeforeFirstSection`: `- **R9.** x` between
  `# Spec` and `## Summary` gives `R9 is outside "## Requirements"`, and
  `- **R2 x` there gives the "is not a marker" error. Proves: R15

### F3. The changed name is not written as task 03 writes it

Level: low
Where: tasks.md → Task 04 → Interfaces
Breaks: nothing
Why: The tasks template says to write a name the same way in every
  task, so sliceflow can match them. Task 03 writes
  `type Doc struct{ ... }`; task 04 writes `Doc.Preamble Section`. The
  depends_on chain is right either way (05 needs 04; 09 and 14 come
  after 05), so nothing breaks now.
Fix: write it as `type Doc struct{ ...; Preamble Section; ... }`:
  changes, with the full field list, or name `Doc` first.

### F4. The lead's decision says no interface changes

Level: low
Where: status.md → Lead's decisions (task 04, task-04-1/F1)
Breaks: nothing
Why: The decision says the fix is "the same package, no interface
  changes", and names only parse.go and parse_test.go. After the
  human's answer, task 04 changes `Doc` and touches doc.go. The line is
  now out of date.
Fix: add that `Doc` gets `Preamble`, by the human's answer (option 1),
  and that doc.go was added to touches.

## Requirements check

None.

## Needs human decision

None.
