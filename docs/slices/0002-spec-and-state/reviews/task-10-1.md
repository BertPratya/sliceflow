---
review: task-10-1
slice: 0002-spec-and-state
kind: task
target: task 10
round: 1
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

`takeLock` and `Repo.LockWait` follow The lock, Data, and TD2. R7's message
matches R7 and S17 word for word, and the four Done when tests are real.
One medium problem: on Windows, the free can fail and its error is
dropped, which leaves a lock that every later command reports as busy
(F1).

## Checked

- Commits: 76aa1c8
- Files read: internal/slice/lock.go, internal/slice/lock_test.go,
  internal/slice/repo.go, internal/slice/create.go (the `.gitignore`
  part), internal/slice/create_test.go (`newRepo`, `fixedNow`),
  internal/slice/errors.go, internal/cli/output.go,
  internal/eventlog/event.go (`TimeLayout`), the brief, plan.md (The
  lock, Data, TD2, Interrupts, Risks), spec.md (R6, R7, S17), status.md
  (Findings, Decisions), Go 1.25's src/syscall/syscall_windows.go
  (`Open`) and src/os/stat_windows.go (`stat`)
- Ran:
  go test -race -count=3 -run TestLock ./internal/slice/ (pass)
  GOOS=windows go vet ./internal/slice/ (pass)
- Requirements: R6 (the lock file, its content, the free), R7
- Interfaces: `Repo.LockWait time.Duration` (0 means 10 s) and
  `takeLock(ctx context.Context, r Repo, name string) (func(), error)`
  match the task

## Tests

- `TestLockFile`: ok
- `TestLockBusy`: ok
- `TestLockUnreadable`: ok
- `TestLockWaits`: ok

## Previous findings

None.

## Findings

### F1. On Windows, a failed free is ignored and leaves the lock

Level: medium
Where: internal/slice/lock.go:62
Breaks: nothing
Why: the returned function is `func() { os.Remove(path) }`, and its error
  is dropped. On Windows, Go opens files without `FILE_SHARE_DELETE`
  (src/syscall/syscall_windows.go:396). So while any other process has
  the lock file open, `os.Remove` fails with a sharing violation. Input:
  command A frees its lock at the moment a waiting command B reads that
  file for R7's message (`busyError`), or a virus scanner has it open.
  Wrong result: the file stays, and every later command waits 10 s and
  prints R7's busy error, until the user deletes the file by hand. R6
  says the lock is removed when the command ends. Task 11 sets this
  function as its `defer`, so the gap goes into every write. Found by
  reading the code and Go's Windows `Open`; not run on Windows.
Fix: in the free, try `os.Remove` again a few times (e.g. every 10 ms
  for up to 1 s) while it fails and the file still exists.

### F2. A lock being deleted on Windows can give "access denied" at once

Level: low
Where: internal/slice/lock.go:64-69
Breaks: nothing
Why: when `os.OpenFile` fails with an error other than "exists", the
  code retries only if `os.Stat` succeeds. A file in Windows' "delete
  pending" state (another handle opened with share-delete, e.g. a
  scanner, is still open on it) gives "access denied" to `CreateFile`.
  In Go's `stat` (src/os/stat_windows.go) `GetFileAttributesEx` and then
  `CreateFile` most likely fail the same way, so `os.Stat` fails too and
  `takeLock` returns the raw "access denied" error instead of trying
  again. Rare, and the user can run the command again. Found by reading
  the code; not run on Windows.
Fix: retry also when `os.Stat` fails with anything but "not exist", or
  when the open error is `fs.ErrPermission`, until the deadline.

## Requirements check

None.

## Needs human decision

None.
