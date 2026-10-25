---
review: task-18-1
slice: 0002-spec-and-state
kind: task
target: task 18
round: 1
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

The fix works for the case integration-1/F1 describes: a failed or
stopped `git commit` now puts back the staged `spec.md`, the stale
`index.lock` rule still holds for both index commands, and the success
path is unchanged. One medium problem: a Ctrl-C that comes before
`git add` runs, e.g. while `spec submit` checks the spec, still resets
the staged `spec.md` to the last commit, because the undo then has no
saved entries and resets every file (F1).

## Checked

- Commits: 7aea506
- Files read: internal/slice/save.go, internal/slice/save_test.go
  (the new test, and the index.lock tests and helpers),
  internal/e2e/submit_review_failures_test.go (the new test),
  internal/slice/submit.go (the op of `spec submit`),
  internal/git/git.go (`Run`),
  docs/slices/0002-spec-and-state/reviews/integration-1.md (F1),
  the brief of task 18
- Ran:
  go test ./internal/slice/ -run 'TestSaveUndoKeepsStaged|IndexLock|TestSaveCancelled|TestSaveHookFails' -count=1 -v (pass)
  go test ./internal/e2e/ -run 'TestSubmitStagedSpecKept|TestSubmit' -count=1 (pass)
  go test ./... -count=1 (pass)
  a probe test, in a copy of the commit outside the worktree: `spec.md`
    staged as `staged text` and changed on disk, an op with
    `Keep: spec.md` and no files, ctx cancelled inside `Prepare`
    (fails: error `interrupted. Nothing was changed.`, but `:spec.md`
    is the committed template again)
- Requirements: R8, R10
- Interfaces: `saveEvent` and `saveOp` unchanged

## Tests

- `TestSaveUndoKeepsStaged`: ok
- `TestSubmitStagedSpecKept`: ok

## Previous findings

None.

## Findings

### F1. A Ctrl-C before git add still resets a staged spec.md

Level: medium
Where: internal/slice/save.go:160-190, internal/slice/save.go:246-254
Breaks: R10
Why: The saved index entries are filled only after `git ls-files`
  runs (line 251). Every step before it that fails calls the same
  undo, which then finds no entry for any file and runs
  `git reset -q --` on all of `gitFiles`. For `spec submit`,
  `spec.md` is in `gitFiles` from line 247, and nothing between
  `op.Prepare` and `ls-files` checks the ctx. So a Ctrl-C while the
  spec checks run in `Prepare`, while the log is written, or while
  `ls-files` runs, makes `ls-files` fail with the cancelled ctx, and
  the undo resets `spec.md` in the index to the last commit, while the
  message says `interrupted. Nothing was changed.` I showed it with a
  probe test: `spec.md` staged as `staged text` and changed on disk,
  `Keep: spec.md`, ctx cancelled inside `Prepare`. The error is
  `could not save the event: interrupted. Nothing was changed.`, and
  `git show :spec.md` then gives the committed template, not
  `staged text`. This is the same loss as integration-1/F1, in a
  smaller window. The R10 rule `unstage what it staged` means the
  index should not be touched at all when `git add` never ran.
Fix: set a flag just before `git add` runs, and in the undo touch the
  index (reset and update-index) only when it is set. Add a case to
  `TestSaveUndoKeepsStaged` with the ctx cancelled in `Prepare`.

### F2. A conflicted spec.md loses its conflict state in the undo

Level: low
Where: internal/slice/save.go:255-262
Breaks: nothing
Why: Entries with a stage other than 0 are skipped, so a file with a
  merge conflict goes to `noEntry` and is reset to the last commit at
  stage 0. While a merge is in progress, `git commit --only` always
  fails, so every save takes this path. Then `git status` no longer
  shows `spec.md` as unmerged, while the file on disk still has the
  conflict markers. `update-index --cacheinfo` cannot put back stages
  1 to 3, and `--index-info` needs stdin, which `git.Run` does not
  take. The old code did the same, so this fix makes nothing worse.
  It is rare: the lead merges `main` into the slice branch, gets a
  conflict in `spec.md`, and runs `spec submit` before resolving it.
Fix: optional: refuse the save before `git add` when `ls-files -s`
  shows a stage other than 0 for one of the files, or note it in the
  backlog.

## Requirements check

None.

## Needs human decision

None.
