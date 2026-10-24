---
review: spec-2
slice: 0001-slice-and-events
kind: spec
target: docs/slices/0001-slice-and-events/spec.md
round: 2
reviewer: doc-reviewer (claude opus)
---

# Review

## Summary

The fixes from spec-1 hold, and the spec has no gap that needs the
human. Two fixes are not complete (F6: the form of `at`; F10: two slices
with the same time). One fix made a new conflict: R9 says every relative
path is relative to the main checkout, but S1's `Spec:` path is relative
to the worktree (F1).

## Checked

- Files read: docs/slices/0001-slice-and-events/spec.md,
  docs/slices/0001-slice-and-events/status.md,
  docs/slices/0001-slice-and-events/reviews/spec-1.md,
  docs/templates/spec.md, docs/templates/review.md,
  .agents/prompts/doc-reviewer.md, AGENTS.md
- Ran:
  grep for leftover `{{` and `<!--` in spec.md (none found)
  git show 622ae40 --stat, to see the fix commit
  counted the column widths of the S20 table against R34 (all match)
- Requirements: R1-R41, S1-S28, SD1-SD13, every Shows: and Rule: line

## Tests

None.

## Previous findings

- spec-1/F1: verified (fixed)
- spec-1/F2: verified (fixed)
- spec-1/F3: verified (fixed)
- spec-1/F4: verified (fixed)
- spec-1/F5: verified (fixed)
- spec-1/F6: reopened. R26's reasons name only the type of a field, not
  its form. Input: `"at":"yesterday"`, or `"at":"2026-10-24 07:42"`
  (a string, so the type is right), or a `data.from.commit` that is not
  a full commit id. One worker skips the line with `field "at" has the
  wrong type`; another keeps it, and S13 prints a bad `Created:`. Add a
  reason for a field with the wrong form, e.g. `field "<f>" has the wrong
  form`, and say which fields have a form (`at`, `data.from.commit`).
- spec-1/F7: verified (fixed)
- spec-1/F8: verified (fixed)
- spec-1/F9: verified (fixed)
- spec-1/F10: reopened. The widths are fixed, and S20 now matches R34.
  Two slices with the same created time still have no order. `at` is to
  the second, so an e2e test that creates two slices in a row often gets
  the same time, and the row order changes from run to run. Say they
  sort by name.
- spec-1/F11: verified (fixed)
- spec-1/F12: verified (wontfix)
- spec-1/F13: verified (fixed)
- spec-1/F14: verified (fixed)
- spec-1/F15: verified (fixed)

## Findings

### F1. R9 says the Spec: path is relative to the main checkout, but S1 shows it relative to the worktree

Level: medium
Where: spec.md → Requirements → R9, Story 1 → S1
Breaks: R9
Why: The F3 fix added to R9: "Each relative path is relative to the root
  of the main checkout". S1 prints `Spec:     docs/slices/spec-submit/spec.md`.
  From the main checkout, that file does not exist; it is in the
  worktree. The `Next:` line follows R9, the `Spec:` line does not.
  Input: `sliceflow slice create spec-submit`. A worker who follows R9
  prints `Spec:     .sliceflow/worktrees/spec-submit/docs/slices/spec-submit/spec.md`;
  a worker who follows S1 prints the short path. R10 says `spec` in
  `--json` is "the path inside the worktree", which matches S1.
Fix: in R9, say the `Spec:` line is the path inside the worktree, as in
  R10, and every other relative path is relative to the root of the main
  checkout.

### F2. The base-branch warning does not fit R28, and its place in --json is not given

Level: medium
Where: spec.md → Requirements → R23, R28, R30, R36; Story 3 → S25
Breaks: nothing
Why: R28 says every warning prints as `Warning: <name>: <warning>`, but
  R23's warning (added for spec-1/F5) has no slice name, as S25 shows. A
  worker must guess whether R23 is an exception to R28. It is also not
  said whether that warning goes in each slice's `warnings` list in
  `--json`, and whether it makes R36 show ` !` on every row. Input: S25's
  repo, then `sliceflow slice list --json`. One worker prints the warning
  on stderr only; another adds it to the `warnings` of every object.
  The form of an item in `warnings` is not given either: with or without
  `Warning: <name>: ` before it.
Fix: say R28 covers the warnings of one slice (R26, R27); R23's warning
  goes to stderr only, and not in `warnings` or ` !`. Say each item in
  `warnings` is the `<warning>` text alone, e.g.
  `.events.jsonl line 2: not valid JSON; skipped`.

### F3. slice list --json with no slices has no output given

Level: medium
Where: spec.md → Requirements → R34, R35, R37
Breaks: nothing
Why: R34 and R35 print `No slices in progress.` or `No slices.` when
  there are no rows. R37 says `--json` prints one JSON list. Input:
  `sliceflow slice list --json` in a repo with no slices. One worker
  prints `[]`; another prints `No slices in progress.`, which an agent
  cannot parse.
Fix: say R37 prints `[]` when there are no rows.

## Requirements check

None.

## Needs human decision

None.
