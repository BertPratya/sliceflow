---
review: task-11-1
slice: 0002-spec-and-state
kind: task
target: task 11
round: 1
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

The write path follows the plan, R10's text matches word for word, `seq`
is right after skipped lines and seq warnings, and all eleven tests are
real. One high problem: Ctrl-C after git has made the commit (e.g. during
a `post-commit` hook, or during the `rev-parse HEAD` after it) undoes the
files on disk and reports "Nothing was changed.", while the commit is on
the branch (F1). The `index.lock` removal after Ctrl-C can delete a lock
that another git command holds (F2).

## Checked

- Commits: d994f92
- Files read: internal/slice/save.go, internal/slice/save_test.go,
  internal/slice/undo.go, internal/slice/create.go (fail),
  internal/slice/errors.go, internal/slice/status_test.go (appendLog),
  internal/git/git.go, internal/eventlog/parse.go, internal/cli/output.go,
  the brief, docs/slices/0002-spec-and-state/plan.md (Saving an event,
  Undo after a failed commit, Interrupts, TD1),
  docs/slices/0001-slice-and-events/plan.md (Undo on failure),
  docs/slices/0002-spec-and-state/status.md (Findings, spec-1/F9)
- Ran:
  go test ./internal/slice/ -run TestSave -count=1 (pass)
  a scratch test in a copy of the repo, outside the worktree: a
    `post-commit` hook `sleep 3`, ctx cancelled after 500 ms (the commit
    was made: 2 commits before, 3 after; error `could not save the event:
    interrupted. Nothing was changed.`; `git status` shows
    `M docs/slices/login-google/.events.jsonl`)
  the same with a `pre-commit` hook `sleep 3` (undone; the worktree's git
    folder keeps `next-index-<pid>.lock`)
- Requirements: R2, R6, R8, R9, R10
- Interfaces: prepared, saveOp, and saveEvent(ctx, r, name, worktree, op)
  match the task; takeLock, state.Replay, state.Allowed, eventlog.New,
  eventlog.Encode, git.Run, git.Show are used

## Tests

- `TestSaveCommit`: ok
- `TestSaveOnlyOwnFiles`: ok
- `TestSaveNoNewline`: ok
- `TestSavePrepareFails`: ok
- `TestSaveNotAllowed`: ok
- `TestSaveHookFails`: ok
- `TestSaveTail`: ok
- `TestSaveCancelled`: ok
- `TestSaveTwoAtOnce`: ok
- `TestSaveSeqAfterSkipped`: ok
- `TestSaveLeftoverTemp`: ok

## Previous findings

None.

## Findings

### F1. Ctrl-C after the commit is made undoes the disk and says nothing changed

Level: high
Where: internal/slice/save.go:220-226
Breaks: R10
Why: R10 counts only Ctrl-C "before the commit is made" as a failed
  commit. `git commit` makes the commit and moves the branch before it
  runs the `post-commit` hook. A Ctrl-C during that hook kills git, Run
  returns `ctx.Err()`, and `fail` runs the undo: it writes the old log
  back to disk, deletes the event's new files, and returns `could not
  save the event: interrupted. Nothing was changed.` But the commit is on
  `slice/<name>`. My scratch test (post-commit `sleep 3`, cancel after
  500 ms) gave 3 commits instead of 2, and `git status` then showed the
  log as modified, older than the commit. A caller (task 14's two events)
  would also stop with the wrong tail. The same happens when Ctrl-C comes
  just after a successful commit: `rev-parse HEAD` at line 223 runs with
  the cancelled ctx and returns the plain error `context canceled`,
  though the event was saved.
Fix: note the branch's commit before `git commit`; when the commit
  fails or is stopped, check it again with `context.WithoutCancel(ctx)`.
  If it moved, the event is saved: no undo, return success. Run the
  `rev-parse` with `context.WithoutCancel(ctx)` too.

### F2. The undo may delete an index.lock that another git command holds

Level: medium
Where: internal/slice/save.go:151-165
Breaks: R8
Why: After Ctrl-C, any failure of `git reset` makes the undo delete the
  worktree's `index.lock`, without checking why reset failed or whose lock
  it is. The lock is stale only when Go's kill hit git before git's own
  signal handler cleaned up. With a normal Ctrl-C in a terminal, git gets
  the SIGINT too and usually removes its own lock first. Then reset fails
  only if another git command in that worktree holds the lock right then:
  the user's `git add`, or an editor's git. The undo then deletes that
  live lock, and both processes write the index at once. The other
  command's change to the index is lost, or it fails with "unable to
  write new index file". R8 says other staged changes stay as they were.
  0001 never needed this: its undo removes the whole worktree, git folder
  and all, and "never touches a thing that is not on the list".
  Neither the plan's "Interrupts" nor "Undo after a failed commit" says to
  remove git's lock.
Fix: safest: stop git with an interrupt instead of a kill
  (`cmd.Cancel` sends `os.Interrupt` on Unix, kill after `WaitDelay`), so
  git removes its own locks. That is in internal/git, outside this task's
  touches: the lead's call. Where a stale lock still comes up (Windows, or
  a git that does not exit), remove it only when reset's error is git's
  `index.lock': File exists` and the lock file is not newer than the
  moment the stopped git command ended. While a stale lock exists, no
  other git can make a new one, so an older lock is surely the dead git's.

### F3. A stopped commit leaves next-index-<pid>.lock in the git folder

Level: low
Where: internal/slice/save.go:220
Breaks: nothing
Why: `git commit --only` builds a temporary index named
  `next-index-<pid>.lock` in the worktree's git folder. When the kill
  stops git, that file stays; my scratch test found it after the undo. It
  does not block later commits (the pid differs), but it stays there for
  good. The hook git was running is not stopped either: after a kill, the
  `pre-commit` hook's process keeps running on its own.
Fix: the interrupt in F2's fix lets git remove the file itself.

### F4. The undo deletes a file that was on disk before the run but not in the commit

Level: low
Where: internal/slice/save.go:193-205
Breaks: nothing
Why: A file `Prepare` returns that is not in the commit goes on the
  `created` list, even when it was already on disk. If the commit then
  fails, the undo deletes it. The plan says to delete only "the files
  step 5 created". For the review file, R24's check in `Prepare` should
  stop this earlier, so no current caller reaches it.
Fix: before writing, check `exists(target)`; put only files that were
  not there on the `created` list, and keep the old bytes of the others
  to write back.

## Requirements check

None.

## Needs human decision

- **Q1.** What does R10's message say when the undo cannot put something
  back? R10 always ends with `Nothing was changed.` (or the Tail). 0001's
  R18 replaces its last sentence with `Could not remove: <list>.` in that
  case. The worker keeps `Nothing was changed.` and adds a second line
  `Could not put back: <list>.`, so the message says nothing was changed
  and then lists what was changed. This choice is ASSUMED in the commit
  and changes what users see.
  Guess: as R18: `Could not put back: <list>.` replaces the last
  sentence; or keep the worker's second line.
