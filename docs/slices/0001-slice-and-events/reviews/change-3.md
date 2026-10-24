---
review: change-3
slice: 0001-slice-and-events
kind: change
target: docs/slices/0001-slice-and-events/plan.md
round: 1
reviewer: doc-reviewer (claude opus)
---

# Review

## Summary

All three fixes resolve their findings and agree with R7, R18, and R41.
One new problem comes from the F1 fix: the new unit test
`TestReasonFileError` uses the path `C:\a\b\spec.md`, and
`filepath.ToSlash` changes `\` only on Windows, so the test fails on
Linux and macOS in CI (F1, medium). Nothing needs the human.

## Checked

- Files read: docs/slices/0001-slice-and-events/reviews/plan-3.md,
  spec.md (R7, R18, R41), plan.md (Flow 3, Cross-platform paths),
  tasks.md (task 01 Interfaces, task 05, task 09), status.md,
  docs/templates/spec.md (front matter), docs/templates/review.md,
  .agents/prompts/doc-reviewer.md
- Ran:
  git show cdbf68f -- plan.md tasks.md (the three fixes)
  go run of a small program in the scratchpad: `filepath.ToSlash` of `C:\a\b\spec.md` on Linux prints `C:\a\b\spec.md`; of `filepath.FromSlash("C:/a/b/spec.md")` prints `C:/a/b/spec.md`
- Requirements: R7, R18, R41

## Tests

None.

## Previous findings

- plan-3/F1: verified (fixed)
- plan-3/F2: verified (fixed)
- plan-2/F2: verified (fixed)

## Findings

### F1. `TestReasonFileError` fails on Linux and macOS

Level: medium
Where: tasks.md → Task 09 → Done when, `TestReasonFileError`
Breaks: AGENTS.md, CI runs the checks on Linux, Windows, and macOS
Why: The test gives the `*fs.PathError` the `Path` `C:\a\b\spec.md`
  and expects `open C:/a/b/spec.md: Access is denied.`. Flow 3, step 4
  says to use `filepath.ToSlash(Path)`, which changes only the system's
  own separator. I ran it on Linux: `filepath.ToSlash` of
  `C:\a\b\spec.md` gives `C:\a\b\spec.md`, so the test fails there,
  while the code is right.
Fix: give the test the `Path` `filepath.FromSlash("C:/a/b/spec.md")`;
  the expected text stays the same on every system.

### F2. The example reason ends with `.`, so the error line has `..`

Level: low
Where: plan.md → Flow 3, step 4, tasks.md → Task 09 → Done when,
  `TestReasonFileError`
Breaks: nothing
Why: R18 prints `<reason>. Everything sliceflow made was removed.`
  With the new example reason `... Access is denied.`, the line reads
  `...: Access is denied.. Everything sliceflow made was removed.`
  Neither the plan nor task 09 says whether a final `.` of the reason
  is dropped.
Fix: say in Flow 3 that a final `.` of the reason is dropped, or leave
  it as Windows prints it.

## Requirements check

None.

## Needs human decision

None.
