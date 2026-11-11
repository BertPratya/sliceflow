---
slice: slice-and-events
---

# Tasks

This file holds plan step 7 only: the create flow, with its undo list
(decided 2026-11-11, the human). Steps 0 to 6 were built before this
file existed, and steps 8 and 9 are added here when they are designed.
So the rule "every requirement is in a task" holds for step 7's part of
the spec only.

The human and the lead designed step 7 together on 2026-11-11. The
decisions of that design:

1. `slice.Create` does the work and prints nothing. `internal/cli`
   prints the lines of R8.
2. `Create` checks in the order of R40, then makes the branch, the
   worktree, the two files, and the commit (R37).
3. Each thing made goes on an undo stack. The undo runs last made first,
   and tries every entry even when one fails; it returns what is left,
   for R41.
4. `git check-ignore` is asked before the add, then both files are
   added by name with `git add -f` (R5).
5. The undo's git commands, and the questions asked after a failure,
   run in a process group of their own, so a Ctrl+C, even pressed many
   times, cannot stop them.
6. When the context ends, git gets an interrupt, not a kill, so git can
   clean up after itself. After a git step that failed or was stopped,
   `Create` asks git, with a new context, whether the branch or the
   worktree is there anyway, and if so it goes on the stack. The undo
   removes a worktree with `git worktree remove -f -f`, because a half
   made worktree is locked.
7. An empty `--from=` counts as no value (R74, SD29).
8. `spec.md` has no header (R3, SD30).
9. The lock file of the log is ignored through `info/exclude` (R2,
   SD31), and R41 prints `--force --force` (both decided in the review
   of this file).

## Task 01: git-steps

```yaml
covers: [R2, R5, R6, R25, R36, R37, R39]
depends_on: []
touches: [internal/git/run.go, internal/git/run_test.go,
  internal/git/run_unix_test.go, internal/git/proc_unix.go,
  internal/git/proc_windows.go, internal/git/refs.go,
  internal/git/refs_test.go, internal/git/worktree.go,
  internal/git/worktree_test.go, internal/git/commit.go,
  internal/git/commit_test.go]
follows: [TD5, One process per git command, Tool errors word for word,
  Tests make their own repos, Testing]
risk: core
```

### Goal

`internal/git` gets the git steps that make a slice, and the ones that
remove it again. The runner stops git with an interrupt instead of a
kill, and gets a second way to run git, shielded from Ctrl+C. The
hardest part is the two process settings: they differ on Windows and on
Linux and macOS, and only CI runs the Windows one.

### Steps

1. Write the tests listed under Done when. Run them, and see them fail.
   Tests that use `syscall.Kill`, `Setpgid`, or `syscall.Getpgrp` go in
   `internal/git/run_unix_test.go`, with `//go:build !windows`, so the
   package still builds on Windows.
2. In `internal/git/run.go`, change `run`: set `cmd.Cancel` to
   `interrupt(cmd)` from step 3, so a cancelled context sends git an
   interrupt. Keep `WaitDelay`: when git has not exited 5 seconds after
   the interrupt, Go kills it. Add this to `run`'s doc comment. A git
   that exits with 0 after the interrupt still returns the context's
   error (step 3 of the doc comment); the caller then asks git what is
   there (task 03).
3. Add `internal/git/proc_unix.go` (`//go:build !windows`) and
   `internal/git/proc_windows.go`, in the style of
   `internal/filelock/lock_unix.go` and `lock_windows.go`. Each has:
   - `interrupt(cmd *exec.Cmd) func() error`. Linux and macOS: send
     `os.Interrupt` to the process. Windows: a single process cannot get
     Ctrl+C, so it returns nil and does nothing; git in a console got
     Ctrl+C from the console already, and `WaitDelay` kills it after 5
     seconds if not.
   - `shield(cmd *exec.Cmd)`. Linux and macOS: `SysProcAttr` with
     `Setpgid: true`. Windows: `SysProcAttr` with `CreationFlags:
     syscall.CREATE_NEW_PROCESS_GROUP`.
4. In `run.go`, add `runShielded(ctx, dir, args...)`: `run` with
   `shield(cmd)`. Share the code with `run`; do not copy it.
5. In `internal/git/refs.go`, add `CreateBranch`, `DeleteBranch`, and
   `BranchExistsShielded` (see Interfaces). `CreateBranch` runs `git
   branch --no-track --end-of-options <branch> <commit>` with `run`.
   `DeleteBranch` runs `git branch -D --end-of-options <branch>` with
   `runShielded`. `BranchExistsShielded` is `BranchExists` with
   `runShielded`; share the code.
6. Add `internal/git/worktree.go` with `AddWorktree`,
   `WorktreeRegistered`, and `RemoveWorktree`. `AddWorktree` runs `git
   worktree add --end-of-options <path> <branch>` with `run`.
   `WorktreeRegistered` reads `git worktree list --porcelain` (no `-z`:
   it needs git 2.36) with `runShielded`, and reports whether one
   `worktree <path>` line names the same folder as path. To compare,
   resolve each path with `filepath.EvalSymlinks` on its deepest folder
   that exists, then add the rest, then `filepath.Clean`; ignore case on
   Windows. This handles macOS's `/var` link, Windows short names such
   as `RUNNER~1`, and a half made worktree. `RemoveWorktree` runs `git
   worktree remove -f -f --end-of-options <path>` with `runShielded`.
7. Add `internal/git/commit.go` with `IsIgnored`, `AddForce`, `Commit`,
   and `ExcludeFile`. `IsIgnored` runs `git check-ignore -q
   --end-of-options <path>` with `run`: exit 0 is yes, exit 1 is no, as
   in `BranchExists`. `AddForce` runs `git add -f --end-of-options
   <paths>`, with no `--` (after `--end-of-options`, git reads `--` as a
   path). `Commit` runs `git commit -q -m <message>`: never
   `--no-verify`, never `-c`, so the user's hooks and signing run (R6).
   `ExcludeFile` returns the trimmed answer of `git rev-parse
   --path-format=absolute --git-path info/exclude`, which is the shared
   file of the main repository, also from a worktree.
8. Each new function returns git's errors as they are (`*Error`,
   `ErrNotFound`, or the context's error), with a doc comment of
   numbered steps, in the style of `refs.go`.
9. Run the tests until they pass, on Linux. Windows and macOS run in
   CI.

### Interfaces

- `func CreateBranch(ctx context.Context, dir, branch, commit string) error`:
  creates. Task 03 uses it.
- `func DeleteBranch(ctx context.Context, dir, branch string) error`:
  creates. Shielded. Task 03 uses it.
- `func BranchExistsShielded(ctx context.Context, dir, branch string) (bool, error)`:
  creates. Task 03 uses it.
- `func AddWorktree(ctx context.Context, dir, path, branch string) error`:
  creates. Task 03 uses it.
- `func WorktreeRegistered(ctx context.Context, dir, path string) (bool, error)`:
  creates. Shielded. Task 03 uses it.
- `func RemoveWorktree(ctx context.Context, dir, path string) error`:
  creates. Shielded. Task 03 uses it.
- `func IsIgnored(ctx context.Context, dir, path string) (bool, error)`:
  creates. Task 03 uses it.
- `func AddForce(ctx context.Context, dir string, paths ...string) error`:
  creates. Task 03 uses it.
- `func Commit(ctx context.Context, dir, message string) error`:
  creates. Task 03 uses it.
- `func ExcludeFile(ctx context.Context, dir string) (string, error)`:
  creates. Task 03 uses it.

### Done when

- `TestRunInterruptsOnCancel` (`run_unix_test.go`): a fake git sets a
  trap for the interrupt, then writes a file `ready`, and waits; on the
  interrupt it writes `got-interrupt` and exits 0. The test waits for
  `ready` (with a time limit), then cancels: `got-interrupt` exists, and
  `run` returns `context.Canceled`.
  Proves: R39
- `TestRunShieldedOwnGroup` (`run_unix_test.go`): a fake git prints its
  process group. With `runShielded` it differs from the test's own
  group; with `run` it is the same.
  Proves: R39
- `TestCreateBranch`: in a repo from `gittest.New`, `CreateBranch(ctx,
  dir, "slice/login", <hash of main>)` makes `slice/login` at that hash,
  and the checked-out branch is still `main`. On an existing branch it
  returns an `*Error`.
  Proves: R37
- `TestDeleteBranch`: after `CreateBranch`, `DeleteBranch` removes
  `slice/login`; `BranchExistsShielded` says no.
  Proves: R36
- `TestAddWorktree`: `AddWorktree` makes the worktree at a path in
  `t.TempDir()`, on `slice/login`. A folder that exists and is not empty
  gives an `*Error`; an empty one is used.
  Proves: R25, R37
- `TestWorktreeRegistered`: true for the worktree just added, also when
  the path is given through a symlink (skipped where links cannot be
  made); false for another folder.
  Proves: R36
- `TestRemoveWorktree`: removes a worktree that has an untracked file;
  removes a locked worktree (`git worktree lock`); afterwards
  `WorktreeRegistered` is false and the folder is gone.
  Proves: R36
- `TestIsIgnored`: with `.gitignore` holding `*.jsonl`, `a.jsonl` is
  ignored and `a.md` is not.
  Proves: R5
- `TestAddForceAndCommit`: with `.gitignore` holding `spec.md`,
  `AddForce(ctx, dir, "spec.md", "b.md")` then `Commit(ctx, dir, "login:
  add the spec and the event log")` makes a commit with that message
  that holds both files.
  Proves: R5
- `TestCommitRunsHooks`: a `pre-commit` hook that exits 1 makes `Commit`
  return an `*Error` with the hook's text in `Stderr`. Skipped on
  Windows, where a shell hook may not run.
  Proves: R6, R37
- `TestExcludeFile`: from the main folder and from a worktree, the
  answer is the same full path, ending in `info/exclude`.
  Proves: R2

### Out of scope

- The order of the steps, the undo stack, and the messages. Done in:
  task 02, task 03
- Printing anything. Done in: task 04

## Task 02: errors-and-undo

```yaml
covers: [R12, R13, R14, R15, R20, R23, R24, R29, R36, R37, R38, R39, R41]
depends_on: []
touches: [internal/slice/errors.go, internal/slice/errors_test.go,
  internal/slice/undo.go, internal/slice/undo_test.go]
follows: [Tool errors word for word, Paths on every system, Testing]
risk: core
```

### Goal

`internal/slice` gets every error text of `slice create`, and the undo
stack. Both are small and have no git in them, so they are tested alone,
including R41's lines, which an e2e test can hardly make.

### Steps

1. Write the tests listed under Done when. Run them, and see them fail.
2. In `internal/slice/errors.go`, add one function per error (see
   Interfaces), with the texts of R12, R13, R14, R15, R20, R23, R24,
   R29, R37, R38, and R39, word for word, and R41's lines. An error's
   text is the whole message: full lines, no line break at the end (as
   `runCreate`'s comment in `internal/cli/commands.go` says). The errors
   of R20 and R29 have the method `WrongUse() bool`, which returns true,
   so `internal/cli` exits with 2 (see `wrongUseMarker` in
   `internal/cli/errors.go`); the others give 1.
3. Git's text, for R37: each line indented by two spaces, an empty line
   stays empty, and a last line break is not a line; with no text at
   all, "git printed no message." takes the place of "git said:".
4. A check that fails in a way with no message of its own (Assumed: not
   in the spec):
   - A git error: "Error: git could not <question>. Nothing was
     changed.", then git's text as in step 3. `<question>` is "find the
     repository", "check for commits", "find the base branch", "check
     the name", or "find the worktree folder".
   - Any other error: "Error: could not <question>. Nothing was
     changed.", then "The system said:" and the error's text, indented
     by two spaces, as in R38.
5. In `internal/slice/undo.go`, add the undo stack: `push` adds a step;
   `unwind` runs every step's `undo`, last pushed first, keeps going
   when one fails, and returns the steps whose undo failed, in the order
   it ran them. A step with an empty `left` gets no R41 line: when only
   such steps are left, the message is as if nothing were left.
6. Run the tests until they pass.

### Interfaces

- `type undoStep struct { left, cleanup string; undo func(ctx context.Context) error }`:
  creates. `left` is the R41 line without its indent, e.g. `worktree
  <full path>`; `cleanup` is the command, e.g. `git branch -D
  "slice/<name>"`. Task 03 uses it.
- `func worktreeStep(path string, undo func(context.Context) error) undoStep`
  and `func branchStep(name string, undo func(context.Context) error) undoStep`:
  creates (added in the review of task 02). They build `left` and
  `cleanup` exactly as R41 says. Task 03 uses them.
- `type undoStack []undoStep`, `func (s *undoStack) push(step undoStep)`,
  `func (s *undoStack) unwind(ctx context.Context) []undoStep`: creates.
  Task 03 uses them.
- `func failedStep(step string, err error, left []undoStep) error`:
  creates. R37 for a git step ("create the branch", "create the
  worktree", "make the commit"), with R41 when left is not empty.
  Task 03 uses it.
- `func failedWrite(path string, err error, left []undoStep) error`:
  creates. R38, with R41 when left is not empty. Task 03 uses it.
- `func stopped(left []undoStep) error`: creates. R39, with R41 when
  left is not empty. Task 03 uses it.
- `func failedCheck(question string, err error) error`: creates. Step 4.
  Task 03 uses it.
- One function per check error, e.g. `errNoBase(name, base string,
  fromGiven bool) error`, `errNotARepo() error`, `errGitNotFound()
  error`, `errNoCommits() error`, `errBadName(given string, err error)
  error`, `errControl(option, text string) error`, `errFolderTaken(name,
  base string) error`, `errBranchTaken(name string) error`: creates; the
  worker picks the exact names and lists them in the commit. Task 03
  uses them.

### Done when

- `TestUndoStackOrder`: push a, b, c; `unwind` runs c, b, a.
  Proves: R36
- `TestUndoStackKeepsGoing`: b's undo fails; `unwind` still runs a, and
  returns only b.
  Proves: R36, R41
- `TestFailedStepText`: "make the commit" with git's text "line 1\n\nline
  2\n" gives R37's three parts, the empty line with no spaces; with no
  text, "git printed no message.".
  Proves: R37
- `TestLeftoverLines`: `failedStep` with a worktree and a branch left
  gives S12's text: no "Nothing was changed." on the first line, then
  the R41 lines, the worktree first, with `--force --force` and the
  double quotes.
  Proves: R41
- `TestStoppedText`: `stopped(nil)` is R39's line; with a branch left,
  the first line has no "Nothing was changed." and R41's lines follow;
  with only a step whose `left` is empty, it is R39's line alone.
  Proves: R39, R41
- `TestFailedWriteText`: R38's lines with the path and the system's
  text.
  Proves: R38
- `TestCheckErrorTexts`: each check error has its spec text word for
  word, with S2, S3, S4, S6, S7, S8, S17, S20, and S21's values; R20's
  and R29's have `WrongUse()` true, the others have no `WrongUse`.
  Proves: R12, R13, R14, R15, R20, R23, R24, R29

### Out of scope

- Calling git, and the order of the steps. Done in: task 03

## Task 03: create-flow

```yaml
covers: [R1, R2, R3, R4, R5, R7, R27, R28, R36, R37, R38, R39, R40]
depends_on: ["01", "02"]
touches: [internal/slice/create.go, internal/slice/create_test.go,
  internal/slice/create_unix_test.go]
follows: [TD5, Paths on every system, Tests make their own repos,
  Testing]
risk: core
```

### Goal

`slice.Create` makes a slice: the checks in the order of R40, then the
steps of R37, and the undo when a step fails or the user presses Ctrl+C.
It returns what it made, or an error from task 02. The hardest part is
R36: remove only what this run made, never what was there before, even
when git was stopped halfway.

### Steps

1. Write the tests listed under Done when. Run them, and see them fail.
2. In `internal/slice/create.go`, add `CreateInput`, `Created`, and
   `Create`. The checks; each returns its error at once, with nothing to
   undo. When the context has ended at any check, return `stopped(nil)`.
   1. A control character in `Title`, then in `Description`, with
      `HasControlCharacter` (R29).
   2. The name, with `ParseName` (R19 to R21). `Lowered` is true when
      the name differs from `GivenName`.
   3. The project, with `git.FindProject`: `git.ErrNotFound` gives R14,
      `git.ErrNotARepo` R13.
   4. Commits, with `git.HasCommits` (R15).
   5. The base: `From`, or `main` when `From` is "". Its commit, with
      `git.BranchCommit`; `git.ErrNoBranch` gives R12, whose second line
      depends on whether `From` was "".
   6. The folder `docs/slices/<name>` in the base commit, with
      `git.FolderInCommit` (R23), then the branch `slice/<name>`, with
      `git.BranchExists` (R24).
   7. The worktree path, with `WorktreePath`. Then record, before
      anything is made: the folders of the path that do not exist yet;
      whether the worktree folder exists (it is empty then, or git
      refuses it, R25); and whether git has a worktree registered at the
      path already, with `git.WorktreeRegistered`.
3. The steps. A context that ended between two steps counts as Ctrl+C.
   0. Before anything else (moved here on 2026-11-11, in the review of
      task 03, so a failure cannot undo a finished slice): add the line
      `*.events.jsonl.lock` to the file of `git.ExcludeFile`, which check
      7 asks for (a failure there is `failedCheck("find the info/exclude
      file", err)`). Add it unless a line of the file is exactly that,
      with a `\r` at its end cut off. When the file is not empty and does
      not end with a line break, write `\n` first. Make the file and its
      folder when they are missing (R2, SD31). A failure is R38, with
      the file's full path, and "Nothing was changed.". The line is never
      undone.
   1. `git.CreateBranch` at the base commit, in the project's folder.
      Push the branch with `branchStep`: its undo is `git.DeleteBranch`.
   2. Push the folders step: its undo makes the worktree folder again,
      empty, when it was there before and is gone; then removes the
      folders recorded as missing, deepest first, with `os.Remove`. It
      is best effort: it ignores every error of `os.Remove` and
      `os.Mkdir`, and always returns nil. Its `left` is empty, so it has
      no R41 line.
      This step comes before the worktree, so it runs after the
      worktree's undo; git deletes an empty folder that was there
      before when it is stopped during `worktree add`.
   3. `git.AddWorktree` at the path, in the project's folder. Push the
      worktree with `worktreeStep`: its undo is `git.RemoveWorktree`.
   4. In the worktree, make `docs/slices/<name>`, write `spec.md` from
      `templates.Spec` as it is (R3, SD30), then the log with
      `eventlog.Create`: name, title (`Title`, or the name when empty,
      R28), description, base, and the full base commit (R27).
   5. `git.IsIgnored` for each file, before the add (R5); then
      `git.AddForce` on both, by their paths in the worktree; then
      `git.Commit` with the message of R4. All three run in the worktree,
      and each failure is the step "make the commit".
   6. Return `Created`, with the full paths of the worktree and of
      `spec.md`, and `IgnoredFiles` in the order `spec.md`,
      `.events.jsonl`.
4. When a step fails, or the context ends:
   1. Make a new context with `context.WithoutCancel` (TD5).
   2. When step 3.1 or 3.3 failed before its push, ask git with the new
      context whether the thing is there anyway, and push it if so: the
      branch with `git.BranchExistsShielded`; the worktree with
      `git.WorktreeRegistered`, but only when it was not registered
      before (step 2.7).
   3. `unwind` with the new context.
   4. Return `stopped(left)` when the original context has ended, even
      when the step's own error is a git error (Ctrl+C reaches git and
      sliceflow at once, so git may fail first). Otherwise
      `failedStep(...)` after a git step, `failedWrite(...)` after a file
      write or the lock.
5. Run the tests until they pass.

### Interfaces

- `type CreateInput struct { GivenName, Title, Description, From string }`:
  creates. `From` "" means not given. Task 04 uses it.
- `type Created struct { Name string; Lowered bool; Base, BaseCommit, Worktree, SpecPath string; IgnoredFiles []string }`:
  creates. Task 04 prints it.
- `func Create(ctx context.Context, dir string, in CreateInput) (Created, error)`:
  creates. Task 04 calls it.
- The functions of task 01, and the stack and errors of task 02: uses.

### Done when

- `TestCreateMakesSlice`: in a repo from `gittest.New`, `Create` with
  `GivenName` "Login" returns `Name` "login", `Lowered` true, `Base`
  "main"; `slice/login` has one commit on top of the base with exactly
  `spec.md` and `.events.jsonl` in `docs/slices/login/`; `spec.md`
  equals `templates.Spec`; the user's checkout has the same branch, the
  same HEAD, and no changed file.
  Proves: R1, R2, R3, R4, R7
- `TestCreateExcludesLock`: after `Create`, `git status --porcelain` in
  the worktree is empty, and `info/exclude` has the line once; a second
  slice does not add it again. An `info/exclude` whose last line is
  `*.tmp` with no line break keeps that line, and gets ours on a line of
  its own; one with `*.events.jsonl.lock\r\n` is not changed.
  Proves: R2
- `TestCreateFirstEventData`: the log's first event has the name, the
  title, the description, the base, and the full base commit; no
  `Title` gives the name as the title; `Title` "Card payment" and
  `Description` "Pay by card." are stored as they are.
  Proves: R27, R28
- `TestCreateIgnoredFiles`: with `.gitignore` holding `*.jsonl`,
  `IgnoredFiles` is `[".events.jsonl"]`, and both files are in the
  commit.
  Proves: R5
- `TestCreateCheckOrder`: a table of cases with two problems each
  returns the first by R40: a control character and a bad name gives
  R29; a missing base and a taken branch gives R12; a folder in the base
  and a taken branch gives R23.
  Proves: R40
- `TestCreateUndoOnCommitFail`: with no name or email (TD8), `Create`
  returns R37's "make the commit" error; the branch, the worktree, and
  the folders made for it are gone, and folders that were there before
  are not.
  Proves: R36, R37
- `TestCreateKeepsEmptyFolder`: the worktree folder exists and is
  empty; the commit fails; afterwards the folder exists and is empty.
  Proves: R36
- `TestCreateKeepsRegisteredWorktree`: git has a worktree registered at
  the path whose folder is gone (prunable); `Create` fails at the
  worktree step, and that registration is still there afterwards.
  Proves: R36
- `TestCreateWriteFails`: the base commit has a file `docs`, so
  `docs/slices/<name>` cannot be made in the worktree; the error is
  R38's, with the path; the branch and the worktree are gone.
  Proves: R38
- `TestCreateStopped` (`create_unix_test.go`, `//go:build !windows`): a
  `pre-commit` hook writes a file `waiting`, then waits for a file
  `release` (at most 10 seconds). The test waits for `waiting`, cancels
  the context, then writes `release`. `Create` returns R39's line, and
  nothing it made is left.
  Proves: R39

### Out of scope

- Printing, and the lines of R8, R9, and R22. Done in: task 04
- The e2e tests of the scenarios. Done in: task 04, task 05, task 06

## Task 04: create-command

```yaml
covers: [R1, R2, R3, R5, R8, R9, R10, R11, R16, R17, R18, R21, R22, R27,
  R28, R74]
depends_on: ["03"]
touches: [internal/cli/commands.go, internal/cli/cli_test.go,
  internal/cli/line.go, internal/cli/line_test.go, e2e/main_test.go,
  e2e/create_test.go]
follows: [CLI, Paths on every system, Tests make their own repos,
  Testing]
risk: normal
```

### Goal

`sliceflow slice create` works: `runCreate` calls `slice.Create` and
prints the lines of R8, with the notes of R22 and R5. An empty `--from`
is "needs a value". The e2e harness can set the home folder, the global
git config, and the path, and the e2e tests run the scenarios of a
slice that is made.

### Steps

1. Write the tests listed under Done when. Run them, and see them fail.
2. In `internal/cli/line.go`, an empty value of `--from`, as in
   `--from=` or `--from ""`, gives the error of R74, naming `--from`
   (SD29). `--title` and `--description` keep an empty value (R28).
3. In `internal/cli/commands.go`, add `printCreated` (see Interfaces):
   to stdout, in this order, `Created slice <name>.`; the R22 note when
   `Lowered`; one R5 note per file of `IgnoredFiles`; the `Branch:` line
   with the first 7 characters of `BaseCommit`; the `Worktree:` line;
   the `Next:` line.
4. Replace the placeholder `runCreate`: get the folder with `os.Getwd`,
   build `slice.CreateInput` from the call, call `slice.Create`, and
   print with `printCreated`. On an error, return it as it is. Remove
   `runCreate`'s part of the TEMPORARY comment, and its row in
   `TestPlaceholders` in `cli_test.go`.
5. In `e2e/main_test.go`, add `createCmd` (see Interfaces), next to
   `runIn`, without changing what `runIn` does for the tests that use it
   now. It prepares, and does not start, the program with: its own
   `HOME` (a new temp folder, returned); a global git config with a name
   and an email, or with no identity and `user.useConfigOnly=true` when
   asked (TD8), or a given setting such as `sliceflow.worktree.dir`;
   and the normal `PATH`, or a given one (S21, and task 06's fake git).
   Add `runCmd`, which starts a prepared command, waits, and returns
   the `result` that `check` compares.
6. Add `e2e/create_test.go`. Each test makes its own repo in a temp
   folder, and uses `createCmd`, so the worktree goes under the test's
   `HOME` (R17). Paths in the scenarios, such as `/home/you`, become that
   `HOME`. Build each expected path from `filepath.EvalSymlinks` of the
   folder, because git gives real paths (`/private/var` on macOS, long
   names on Windows).
7. Run the tests until they pass.

### Interfaces

- `slice.CreateInput`, `slice.Created`, `slice.Create`: uses. Task 03
  creates them.
- `func printCreated(w io.Writer, givenName string, c slice.Created) error`:
  creates. Used only in this task. `givenName` was added when it was
  built, because R22's note needs the name as the user gave it.
- `type createOpts struct { NoIdentity bool; Config map[string]string; Path string }`:
  creates. `Path` "" means the normal `PATH`. Tasks 05 and 06 use it.
- `func createCmd(t *testing.T, dir string, opts createOpts, args ...string) (cmd *exec.Cmd, home string)`:
  creates. The command is not started, so a test may set `SysProcAttr`
  and call `Start` itself. Tasks 05 and 06 use it.
- `func runCmd(t *testing.T, cmd *exec.Cmd) result`: creates. Tasks 05
  and 06 use it.

### Done when

- `TestCreateEmptyFrom` (`line_test.go`): `slice create x --from=` and
  `slice create x --from ""` give R74's error for `--from`; `--title=`
  gives no error.
  Proves: R74
- `TestPrintCreated` (`cli_test.go`): a `Created` with `Lowered` and
  both files ignored prints the six lines in the order of R8.
  Proves: R5, R8, R22
- e2e tests in `e2e/create_test.go`, one per scenario, with the
  scenario's values:
  - `TestCreateSlice`: Scenario: S1. Proves: R1, R2, R8, R10, R11, R16, R17
  - `TestCreateLowercase`: Scenario: S5. Proves: R8, R21, R22
  - `TestCreateFrom`: Scenario: S9. Proves: R10, R11
  - `TestCreateFirstEvent`: Scenario: S13. Proves: R27
  - `TestCreateCommit`: Scenario: S14. Proves: R2, R4, R7
  - `TestCreateTitle`: Scenario: S15. Proves: R27, R28
  - `TestCreateFromSubfolder`: Scenario: S16. Proves: R1, R16
  - `TestCreateSpecFile`: Scenario: S18. Proves: R2, R3
  - `TestCreateWorktreeDir`: Scenario: S19. Proves: R18
  - `TestCreateIgnored`: Scenario: S23. Proves: R5, R8
- `TestCreateIgnoredFilesStayTracked`: `.gitignore` holds `spec.md` and
  `*.jsonl`. `slice create login` exits 0 with both R5 notes, and its
  commit holds both files. Then, in the worktree, add a line to each
  file, run a plain `git add` on both (no `-f`) and `git commit -m
  edit`: both work, the new commit changes both files, and `git status
  --porcelain` is empty (the lock file is ignored, SD31). Then a new
  file `docs/slices/login/notes.jsonl` is refused by a plain `git add`,
  which proves the rules are active.
  Proves: R2, R5
- `TestCreatePaths`: on Windows the printed paths use `\`; the paths in
  the commit use `/`.
  Proves: R9

### Out of scope

- The scenarios of a check that fails. Done in: task 05
- The scenarios where a step fails, and Ctrl+C. Done in: task 06

## Task 05: create-checks

```yaml
covers: [R12, R13, R14, R15, R19, R20, R23, R24, R29, R40]
depends_on: ["04"]
touches: [e2e/create_check_test.go]
follows: [Tests make their own repos, Testing]
risk: normal
```

### Goal

e2e tests for each check that stops `slice create` before anything is
made: the message, the exit code, and nothing changed.

### Steps

1. Add `e2e/create_check_test.go`, with `createCmd` and `runCmd` from
   task 04. Each test also checks that nothing changed: no branch
   `slice/<name>`, and no worktree folder.
2. Run the tests until they pass. If a test shows a bug in task 01 to
   04, stop and report it; do not change their files.

### Done when

- `TestCreateBadName`: Scenario: S2. Proves: R19, R20
- `TestCreateBranchTaken`: Scenario: S3. Proves: R24
- `TestCreateFolderTaken`: Scenario: S4. Proves: R23
- `TestCreateNotARepo`: Scenario: S6. Proves: R13
- `TestCreateNoCommits`: Scenario: S7. Proves: R15
- `TestCreateNoMain`: Scenario: S8. Proves: R12
- `TestCreateFromMissing`: Scenario: S17. Proves: R12
- `TestCreateControlCharacter`: Scenario: S20. Proves: R29
- `TestCreateNoGit`: Scenario: S21, with a `PATH` that has no git.
  Proves: R14
- `TestCreateTwoProblems`: `slice create "bad name" --from nope` in a
  folder that is not a repo gives R20, not R13 or R12.
  Proves: R40

### Out of scope

- Any change to the code of tasks 01 to 04. Done in: task 01, task 02,
  task 03, task 04

## Task 06: create-failures

```yaml
covers: [R6, R25, R36, R37, R39, R41]
depends_on: ["04"]
touches: [e2e/create_fail_test.go, e2e/create_fail_unix_test.go]
follows: [Tests make their own repos, Testing]
risk: normal
```

### Goal

e2e tests for a slice that cannot be made: a step fails, the undo runs,
and what is left is reported. The hardest part is Ctrl+C: the test must
send it the way a terminal does, to the whole process group, and many
times.

### Steps

1. Add `e2e/create_fail_test.go`, with `createCmd` and `runCmd` from
   task 04. The tests of S11 and S22 use `NoIdentity` (TD8). git's text
   is not compared, only checked to be there (Testing).
2. Add `e2e/create_fail_unix_test.go`, with `//go:build !windows`, for
   the tests that need a process group or a fake git:
   - Ctrl+C: prepare the command with `createCmd`, set
     `SysProcAttr.Setpgid`, and `Start` it. A `pre-commit` hook writes a
     file and then waits (at most 10 seconds); the test waits for that
     file and sends `SIGINT` to the whole group, like a terminal does.
   - S12: a fake `git` first on `PATH`, through `createOpts.Path`: a
     script that fails `git worktree remove` and `git branch -D` with
     exit 1, and runs the real git for every other command. Use
     `NoIdentity`, so the commit fails.
3. Run the tests until they pass. If a test shows a bug in task 01 to
   04, stop and report it; do not change their files.

### Done when

- `TestCreateWorktreeFolderNotEmpty`: Scenario: S10. The branch is gone
  afterwards; the folder and its file are still there.
  Proves: R25, R36, R37
- `TestCreateCommitFails`: Scenario: S11. The files, the worktree, the
  branch, and the empty folders made for the worktree are gone.
  Proves: R36, R37
- `TestCreateHookRejects`: a `pre-commit` hook that prints "no" and
  exits 1 gives R37's "make the commit" error with "no" in git's text,
  and nothing is left. Skipped on Windows, where a shell hook may not
  run.
  Proves: R6, R37
- `TestCreateKeepsEmptyWorktreeFolder`: Scenario: S22.
  Proves: R36
- `TestCreateCannotRemove` (`create_fail_unix_test.go`): Scenario: S12,
  with the fake git; the output has S12's lines, `--force --force`
  included.
  Proves: R41
- `TestCreateCtrlC` (`create_fail_unix_test.go`): Scenario: S24. One
  `SIGINT` to the group: exit 1, R39's line, and nothing left.
  Proves: R39
- `TestCreateCtrlCMany` (`create_fail_unix_test.go`): as S24, but
  `SIGINT` goes to the group five times, 20 ms apart: still exit 1,
  R39's line, and nothing left.
  Proves: R36, R39

### Out of scope

- Any change to the code of tasks 01 to 04. Done in: task 01, task 02,
  task 03, task 04
