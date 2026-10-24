---
review: task-06-1
slice: 0001-slice-and-events
kind: task
target: task 06
round: 1
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

`Create` follows Flow 1 steps 4 to 8. The five tests are real, and the
shape leaves room for tasks 07, 08, and 09. One medium problem: the
worktree path is built with `path.Join`. That drops the leading `//` of a
Windows network-share (UNC) root, so the worktree lands in the wrong
place (F1).

## Checked

- Commits: 94e0cd6
- Files read: internal/slice/create.go, internal/slice/create_test.go,
  internal/slice/repo.go, internal/slice/base.go, internal/git/git.go,
  internal/testrepo/testrepo.go, internal/eventlog/event.go,
  internal/template/template.go, the brief, tasks.md (tasks 07 to 09),
  plan.md (Flow 3, Undo on failure, Interrupts, Core API)
- Ran:
  go test -count=1 ./internal/slice/ (pass)
  go vet ./... (pass)
  go build ./... (pass)
  gofmt -l cmd internal (prints nothing)
  go test -count=1 ./... (pass)
  go run of a small program that prints path.Join("//server/share/repo", ".sliceflow", "worktrees", "x") (prints /server/share/repo/.sliceflow/worktrees/x)
- Requirements: R1, R4, R5, R6, R7, R8, R19, R20
- Interfaces: CreateOptions, Created, Create, checkCreate (with base and
  source) match the task. Uses eventlog.NewCreated, eventlog.Encode,
  template.Spec, BaseBranch, git.Run.

## Tests

- `TestCreateMakesSlice`: ok
- `TestCreateWritesGitignore`: ok
- `TestCreateTitle`: ok
- `TestCreateHookRuns`: ok
- `TestCreateIgnoredFiles`: ok

## Previous findings

None.

## Findings

### F1. The worktree path loses a UNC root on Windows

Level: medium
Where: internal/slice/create.go:57-58
Breaks: R5
Why: `worktree := path.Join(r.Root, ".sliceflow", "worktrees", opts.Name)`.
  `path.Join` cleans the path, and that turns a leading `//` into `/`.
  For a repo on a Windows network share, git gives the root as
  `//server/share/repo`. Then `worktree` is
  `/server/share/repo/.sliceflow/worktrees/x`, and `filepath.FromSlash`
  makes `\server\share\repo\...`. That path is on the current drive,
  not the share, so `git worktree add` puts the worktree in the wrong
  place, or fails. I checked `path.Join` with a small program: it prints
  `/server/share/repo/.sliceflow/worktrees/x`. On Windows,
  `filepath.Join(filepath.FromSlash(r.Root), ...)` keeps the UNC prefix.
  The tests use `t.TempDir()`, which is never UNC, so they miss this.
Fix: build `wtDir` with `filepath.Join(root, ".sliceflow", "worktrees", opts.Name)`,
  and set `Created.Worktree` to `filepath.ToSlash(wtDir)`.

### F2. .sliceflow/.gitignore is checked, then written in a second step

Level: low
Where: internal/slice/create.go:49-56
Breaks: nothing
Why: `os.Stat`, then `os.WriteFile`. If another process writes the file
  between those two calls, `WriteFile` replaces it. Task 09 must add the
  file to the undo list only when this run made it.
  `os.OpenFile` with `O_CREATE|O_EXCL` answers "made by us" in one call.
  The same goes for `MkdirAll` on `.sliceflow/`, which does not say
  whether it made the folder. Task 09 can change both.
Fix: in task 09, use `O_EXCL` for the file and `os.Mkdir` (treating
  `fs.ErrExist` as "already there") for the folder.

## Requirements check

None.

## Needs human decision

None.
