---
review: task-09-1
slice: 0001-slice-and-events
kind: task
target: task 09
round: 1
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

The undo list, the reasons, the signal `ctx`, and all nine Done when tests
are real, and task-06-1/F2 is fixed. One high problem: an interrupt while
`git worktree add` runs leaves the branch and the worktree behind, but
the error says everything was removed (F1). How to fix it is not decided
by the plan (Q1). Also, a kept branch is not named in `Could not remove:`
(F2).

## Checked

- Commits: b99ebcb
- Files read: internal/slice/create.go, internal/slice/undo.go,
  internal/slice/undo_test.go, internal/slice/check.go,
  internal/slice/errors.go, internal/cli/create.go, internal/cli/output.go,
  internal/git/git.go, internal/e2e/create_undo_test.go, the brief
- Ran:
  gofmt -l cmd internal (prints nothing)
  go vet ./... (pass)
  go build ./... (pass)
  go test -count=1 ./... (pass)
  go test -count=3 for the task's unit and e2e tests (pass)
  git worktree add -b in a temporary repo with 20,000 files, killed with SIGKILL after 0.05 to 0.4 s (each left the branch, the folder, and .git/worktrees/<name>)
  the built binary, slice create in that repo, SIGINT after 0.15, 0.25, 0.35 s (each printed "interrupted. Everything sliceflow made was removed." and left slice/<name> and .sliceflow/worktrees/<name>)
- Requirements: R8, R18
- Interfaces: Create returns R18's *Error after the checks; git.Error used as in task 01. Extra export slice.Interrupted (F3).

## Tests

- `TestCreateHookRefuses`: ok
- `TestCreateHookOutput`: ok
- `TestCreateInterrupt`: ok
- `TestUndoCancelledContext`: ok
- `TestUndoCancelledDuringChecks`: ok
- `TestUndoKeepsOthers`: ok
- `TestUndoCouldNotRemove`: ok
- `TestReasonFileError`: ok
- `TestReasonGitCommand`: ok

## Previous findings

- task-06-1/F2: verified (fixed)

## Findings

### F1. An interrupt during git worktree add leaves the branch and the worktree, and says all was removed

Level: high
Where: internal/slice/create.go:133-144
Breaks: R18
Why: `git.Run` uses `exec.CommandContext`, so a signal kills `git worktree
  add` with SIGKILL. Git cannot clean up then: it has already made the
  branch, the folder, and `.git/worktrees/<name>`. `create` adds the slice
  to the undo list only when `worktree add` succeeds, so the undo skips
  them, and `.sliceflow/worktrees/` is not empty, so it stays too. I ran
  the built binary in a repo with 20,000 files and sent SIGINT after 0.15,
  0.25, and 0.35 seconds. Each run printed
  `Error: could not create slice s015: interrupted. Everything sliceflow made was removed.`
  and exit code 1, but `slice/s015` and `.sliceflow/worktrees/s015` were
  left. A later `create` with that name then fails with R15. The checkout
  of a real repo can take seconds, so this window is easy to hit.
Fix: see Q1. Either let `worktree add` finish (run it with
  `context.WithoutCancel(ctx)`, then check `ctx`), or after an
  interrupted `worktree add`, add the branch and worktree to the list when
  git registered a worktree at this path on `slice/<name>` at `commit`.
  Add a unit test that cancels during `worktree add`.

### F2. A branch kept because its worktree could not be removed is not named in Could not remove

Level: medium
Where: internal/slice/undo.go:81-85
Breaks: R18
Why: When `git worktree remove --force` fails and the folder stays,
  `removeSlice` returns only the folder. The branch `slice/<name>` was
  made by this run and also stays, but the message does not name it. R18:
  "If it cannot remove something, the message ends with
  `Could not remove: <list>.`" The plan's Undo on failure: "Anything it
  could not remove goes into `Could not remove`." In
  `TestUndoCouldNotRemove` (a locked worktree), the error ends with
  `Could not remove: <folder>.` and `slice/x` is still there. Keeping the
  branch is right (git refuses to delete a branch a worktree holds); only
  the list is short.
Fix: return both the folder and the branch, e.g.
  `Could not remove: <folder>, slice/x.`, and check it in the test.

### F3. New export slice.Interrupted is not in the task's Interfaces

Level: low
Where: internal/slice/create.go:41
Breaks: task 09
Why: tasks.md Interfaces for task 09 lists only `Create` and `git.Error`.
  The CLI needs R18's `interrupted` text when `FindRepo` is stopped, and
  keeping that text in one place is the right design. It changes no
  behavior. The note is only so the integration review knows the name.
Fix: none needed in code; the lead may note the name in tasks.md
  Interfaces if the human agrees.

### F4. Could not remove lists absolute paths

Level: low
Where: internal/slice/undo.go:53, internal/slice/undo.go:57, internal/slice/undo.go:84
Breaks: nothing
Why: The list holds the full path, e.g.
  `/home/me/project/.sliceflow/worktrees/x`. The other messages of
  `create` (R16, S1) print paths from the repo root, e.g.
  `.sliceflow/worktrees/x`. Neither the spec nor the plan says which form
  this list uses, so this is only for the same look.

## Requirements check

None.

## Needs human decision

- **Q1.** The plan's Undo on failure adds the branch and the worktree
  "after `git worktree add` succeeds", and Interrupts says "A running git
  command is stopped". Together they leave the branch and worktree of an
  interrupted `worktree add` behind (F1). Which rule changes: let
  `worktree add` finish before the signal is honored, or undo what an
  interrupted `worktree add` left when git registered it at this path?
  Guess: run `worktree add` with `context.WithoutCancel(ctx)` and check
  `ctx` right after it, so the normal undo removes it. It is the simplest
  and never touches another run's branch, but a Ctrl-C waits for the
  checkout to end.
