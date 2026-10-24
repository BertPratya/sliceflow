---
slice: 0001-slice-and-events
created: 2026-10-24
touches:
  - go.mod
---

# Plan

## Approach

sliceflow is one Go program. A thin command layer reads the arguments
and prints the result, as text or as JSON. All the real work is in a
core package, which a later server can call too. The core does every
git step by running `git` as its own process, one command at a time,
with its own arguments. The event log format is its own small package,
with no file or git access, so it is easy to test.

`slice create` never touches the main checkout. `git worktree add`
makes the branch and the worktree in one step. sliceflow writes the two
files inside the worktree, and makes a normal commit there.

| Part | Folder | Change |
|---|---|---|
| Program entry: only `main`, which calls the CLI and exits with its code | cmd/sliceflow/ | new |
| CLI: commands, flags, help, text and `--json` output, exit codes | internal/cli/ | new |
| Slice core: create (checks, steps, undo), status, list | internal/slice/ | new |
| Event log: write an event line; read a log into events and warnings | internal/eventlog/ | new |
| Git: small functions that each run one git command | internal/git/ | new |
| Spec template, built into the binary | internal/template/ | new |
| Test helpers: temporary repos with a test git config; only tests import it | internal/testrepo/ | new |
| End-to-end tests: build the binary, run it in temporary repos | internal/e2e/ | new |

New dependencies: None.

## Flows

### Flow 1: Create a slice

Covers: R1, R2, R3, R4, R5, R6, R7, R8, R9, R10, R19, R20, R41

1. The CLI reads the name and the flags (TD2).
2. The CLI calls `FindRepo`, which finds the folder with
   `git rev-parse --show-toplevel`, and the main checkout's root as the
   first `worktree` line of `git worktree list --porcelain`. That works
   from any worktree, and also when the `.git` folder is somewhere
   else. (lead's call) The CLI passes the `Repo` to `Create`.
3. It runs the checks of Flow 2, in R11's order. While they run,
   nothing is made.
4. It picks the base branch (R3), and turns it into a commit id once
   (Git arguments).
5. It makes `.sliceflow/` and `.sliceflow/.gitignore` when missing
   (R6), and adds each to the undo list.
6. It runs `git worktree add -b slice/<name> <worktree> <commit>` from
   the main checkout's root. This makes the branch and the worktree (R4,
   R5), and both go on the undo list. The main checkout's own files do
   not change.
7. It writes `docs/slices/<name>/spec.md` from the built-in template,
   with the front matter filled (R7), and `.events.jsonl` with one
   `slice.created` line (R19, R20), with `at` from `Repo.Now`.
8. In the worktree, it runs `git add -f -- <the two files>` and
   `git commit -m "<name>: slice created"`, so the user's hooks and
   signing run (R8). `-f` adds the two files also when the project's or
   the user's ignore rules match them, e.g. `*.jsonl`.
9. The CLI prints R9's lines, or R10's JSON, with `/` in every path
   (R41).

### Flow 2: A check fails before creating

Covers: R11, R12, R13, R14, R15, R16, R17, R42

1. R12: `git rev-parse --show-toplevel` fails.
2. R13: no local branch has a commit:
   `git for-each-ref --count=1 refs/heads/` prints nothing.
   (lead's call)
3. R14: the name breaks R2.
4. R16: `refs/heads/<base>` does not resolve to a commit.
5. R15, in R15's order: `refs/heads/slice/<name>` exists; for a remote
   named by `git remote` (it reads only the local config),
   `refs/remotes/<remote>/slice/<name>` exists, so a remote name with
   `/` works and a branch like `user/slice/<name>` does not match
   (lead's call); `git cat-file -e <commit>:docs/slices/<name>`
   succeeds.
6. R17: `.sliceflow/worktrees/<name>` exists (`os.Lstat`).

The first check that fails returns an `*Error` with the spec's words.
The CLI prints it on stderr, prints nothing on stdout, also with
`--json` (R42), and exits with 1.

### Flow 3: A step fails, or the user stops it

Covers: R18

1. A step of Flow 1, 5 to 8, fails, or a signal arrives (Interrupts).
2. `create` runs Undo on failure.
3. The CLI prints R18's error and exits with 1. The reason is
   `interrupted` when a signal arrived, whatever git printed. Otherwise
   it is the first line of git's stderr that is not empty, or
   `git <command> failed (exit code <n>)`; the other lines follow,
   each indented by 2 spaces (R18).

### Flow 4: Status of one slice

Covers: R21, R22, R23, R24, R25, R26, R27, R28, R29, R30, R31, R32, R33

1. The CLI calls `Status`. It finds the repo, and runs the checks of
   R22.
2. It picks the base branch (R21). When that branch does not exist, it
   adds R23's warning and goes on.
3. It finds the name: the one given, or from the current branch with
   `git symbolic-ref --quiet --short HEAD` (R31). Not a `slice/` branch:
   R32's error.
4. A name that breaks R2 gets R33's error, before any git command uses
   it. (lead's call)
5. It picks where to read (R24): `refs/heads/slice/<name>` when it
   exists, else the base branch when `docs/slices/<name>` is on it,
   else R33's error.
6. It reads the log with `git show <ref>:docs/slices/<name>/.events.jsonl`
   (TD6). A log that is not there gives `.events.jsonl not found`
   (R26).
7. `eventlog` reads the lines into events and warnings (R26, R27,
   TD5). The phase is `planning` with a good `slice.created`, else
   `unknown` (R25).
8. The CLI prints R29's lines or R30's JSON, and each warning on
   stderr as R28 says.

### Flow 5: List of slices

Covers: R21, R22, R23, R24, R28, R34, R35, R36, R37, R43

1. The CLI calls `List`. Steps 1 and 2 of Flow 4 run first.
2. It lists the local branches with
   `git for-each-ref --format=%(refname) refs/heads/slice/`. A branch
   whose name after `slice/` breaks R2 is left out, with R43's warning
   about the whole repo.
3. It lists the merged slices: the folders under `docs/slices/` on the
   base branch, with `git ls-tree -d --name-only <base> docs/slices/`.
4. In progress: a local branch whose folder is not on the base branch.
   With `--all`, the merged slices are added (R35).
5. It reads each slice's log as Flow 4, steps 5 to 7.
6. It sorts as R34 says. The CLI prints the table, with ` !` for a
   slice with warnings (R36), or R37's JSON, and the warnings on
   stderr (R28).

### Flow 6: Help and wrong use

Covers: R38, R39

1. `--help` or `-h` on a command prints its usage on stdout, and exits
   with 0 (R39).
2. `sliceflow` alone and `sliceflow slice` alone print the usage on
   stdout, and exit with 0 (R38).
3. An unknown command or flag, a flag with no value, no slice name for
   `slice create`, or more positional arguments than the command takes,
   prints `Error: <what is wrong>. Run sliceflow --help.` on stderr with
   R38's words, and exits with 2.
4. All of these happen in the CLI, before any git command.

### Shared rules

#### Cross-platform paths

Covers: R41
Project rule: AGENTS.md, paths, line endings, and processes must work
  on Windows, macOS, and Linux.

Files on disk are opened with `filepath`. Paths inside git (`<ref>:<path>`)
are built with `path`, which always uses `/`. Every path the CLI prints
goes through `filepath.ToSlash`. Files are written as bytes, so `\n`
stays `\n` on Windows. The built-in template is a `.md` file, which
`.gitattributes` checks out with `\n` on every system, so the binary
holds the same bytes everywhere.

#### Tests

Covers: none
Project rule: AGENTS.md, every change that adds behavior adds a test.
  Tests create their own temporary git repos; never use shared services.

Every task adds the tests its behavior needs, as Testing says, and
names this rule in its follows. Each
test makes its own repo in `t.TempDir()`, with the test git config.
No test uses the network.

#### Shared logic in internal

Covers: none
Project rule: AGENTS.md, shared Go logic goes in `internal/`, not
  `cmd/`.

`cmd/sliceflow/main.go` only calls `cli.Run(os.Args[1:], os.Stdout,
os.Stderr)` and exits with its code. Everything else is in
`internal/`, as in Approach.

#### Undo on failure

Covers: R18

`create` keeps a list of what it made in this run. It adds each thing
right after making it:

1. `.sliceflow/` and `.sliceflow/.gitignore`, only those that were
   missing.
2. The branch and the worktree, after `git worktree add` succeeds.
   `git worktree add` also makes `.sliceflow/worktrees/` when it is
   missing, so `create` notes before step 6 whether that folder exists,
   and adds it to the list after step 6 when it did not.

The files in the worktree and the commit need no entry: removing the
worktree and the branch removes them. If a step fails, `create`
removes the list's items, newest first:

- the worktree with `git worktree remove --force`;
- the branch with `git branch -D`, only if it still points at the
  commit `create` made or started from;
- the files and folders with an OS call, a folder only if it is empty.

It never removes a thing that is not on the list. So when two `create`
commands with the same name run at once, the one that loses never
removes the other's branch. Anything it could not remove goes into
`Could not remove: <list>.` (R18).

#### Interrupts

Covers: R18

While `create` runs, sliceflow catches `os.Interrupt` (Ctrl-C) and
`syscall.SIGTERM`, on every system. On Linux and macOS, a system
shutdown sends `SIGTERM` first. On Windows, Go turns closing the
console window, logging off, and shutting down into `SIGTERM`, with a
few seconds to finish. The signal cancels the `ctx`. A running git
command is stopped, and `create` waits for it to end; between steps,
`create` checks the `ctx` before the next one. Then it runs Undo on
failure, and prints R18's error with the reason `interrupted`.

- **The undo has its own context.** It runs with
  `context.WithoutCancel(ctx)`, so the signal does not also stop the
  undo's git commands.
- **A hook can outlive git.** When git is stopped, a hook it started
  (e.g. `sh` running `sleep 30`) can keep the output pipe open. So
  `git.Run` sets `cmd.WaitDelay` to 2 seconds: after that, Go closes
  the pipes and `Run` returns. (lead's call)
- **The reason is `interrupted` when a signal arrived.** A real Ctrl-C
  reaches git and the hook too, so `git commit` may fail first, often
  with an empty message. `create` checks `ctx.Err()` before it picks
  git's message.
- **Ctrl-C during the checks.** The checks also use the `ctx`. When it
  is cancelled, `git.Run` returns `ctx.Err()`, never a "not found"
  answer, so `RefExists` cannot report a missing branch by mistake.
  `create` then prints R18's error with `interrupted`; nothing was made
  yet, so there is nothing to remove.
`status` and `list` change nothing, so they keep Go's default: they
stop at once. (lead's call)

#### Git arguments

Covers: R3, R4, R40
Project rule: AGENTS.md, run each check or git command as its own
  process with its own arguments, never as one shell string.

Each git command runs with `os/exec`, with its arguments as a list,
never through a shell (TD8).

- A branch from `--from` or `sliceflow.base` is passed as
  `refs/heads/<branch>`, so it never starts with `-` and is never read
  as an option.
- The base branch is turned into a commit id once, at the start.
  Every later step uses that id, so the base cannot move during
  `create`.
- The slice name passes R2 before any git command uses it, so it has
  no `/`, `..`, or leading `-`.
- The title only goes into the JSON line, never into a git argument.
- sliceflow reads only output meant for programs: `rev-parse`,
  `for-each-ref`, `show`, and `worktree list --porcelain`. Git's own
  error text is only passed on, as R18's `<reason>`. (lead's call)

#### No network

Covers: R40
Project rule: AGENTS.md, sliceflow sends no data out: no telemetry, no
  uploads, no calls home.

sliceflow never runs `git fetch`, `git pull`, or `git push`, and reads
remote branches only from what git already has (`refs/remotes/`).
Every git command runs with `GIT_TERMINAL_PROMPT=0`, so git never asks
for a password, and `GIT_NO_LAZY_FETCH=1`, so a partial clone does not
download a missing file. Git 2.44 and newer honor it; older git ignores
it. (lead's call) The user's own git hooks run as the user set them up
(R8); they are the user's code, not sliceflow's.

## Interfaces

### CLI

New. The spec describes it fully; the code follows the spec's words.

- `sliceflow slice create <name> [--title <title>] [--from <branch>] [--json]`:
  R1 to R18, R42; output R9 and R10.
- `sliceflow slice status [<name>] [--json]`: R21 to R33.
- `sliceflow slice list [--all] [--json]`: R21 to R28, R34 to R37, R43.
- `--help` and `-h` on every command: R39.
- Flags may come before or after the name (TD2).
- Exit codes 0, 1, and 2, errors on stderr, nothing on stdout when a
  command fails: R38, R42.

### Event log

New. The file `docs/slices/<name>/.events.jsonl`, as R19 and R20 say,
read as R26 and R27 say. See Data.

### Settings

New. `git config sliceflow.base`: the base branch, as R3 and R21 say.
sliceflow only reads it.

### Core API

New, in `internal/slice`. The CLI calls it; a later server calls the
same functions (TD7).

```go
func Create(ctx context.Context, r Repo, opts CreateOptions) (Created, error)
func Status(ctx context.Context, r Repo, name string) (SliceStatus, error) // name "" = the current branch
func List(ctx context.Context, r Repo, all bool) (SliceList, error)

type CreateOptions struct{ Name, Title, From string }
type Warning struct{ Slice, Text string } // Slice "" = about the whole repo
type Error struct{ Message string }
```

- `Repo` holds the folder the command runs in, and the root of the main
  checkout (for R5 and R9).
- A result holds its own warnings. The core never prints. The CLI
  prints each warning as R28 says, and `--json` adds its text to
  `warnings` (R30, R37).
- Every failure the spec names is an `*Error` with the spec's words,
  without `Error: `. The CLI prints `Error: ` and the message, and exits
  with 1. Any other error is a bug, also printed with exit 1.
- Wrong use (an unknown flag, a missing value) is found by the CLI, and
  exits with 2. It never reaches the core.
- `Repo` has a `Now func() time.Time` for `at` and `created`; tests set
  a fixed time. (lead's call)

## Data

`slice create` writes three files. This is the first version of each,
so no old data is moved over.

| File | Where | What it holds |
|---|---|---|
| `docs/slices/<name>/.events.jsonl` | committed on the slice branch | one `slice.created` line, as the spec's R19 and R20 say |
| `docs/slices/<name>/spec.md` | committed on the slice branch | the built-in template |
| `.sliceflow/.gitignore` | in the main checkout, never committed | the one line `*` |

### Events in the code

One Go type holds any event: `version`, `seq`, `event`, `at`, and
`data`, declared in that order. `encoding/json` writes struct fields in
the order they are declared, so every line has R19's order. `data` of
`slice.created` has its own type: `from` (`branch`, `commit`) and an
optional `title`, left out when empty. See TD4.

### Reading a line

The reader opens each line as a map of field names first, then checks
each field itself, so each warning has R26's exact words. See TD5.
A map keeps only one value per name, so before that, the reader walks
the line's tokens with `json.Decoder.Token` and finds a name that
appears twice in the same object: `field "<f>" appears twice` (R26),
e.g. `data.from.branch`.

- A line of only spaces, tabs, or `\r` is a `blank line`. A good line
  that ends in `\r\n` reads normally, because JSON counts `\r` as a
  space. (lead's call)
- A last line with no `\n` at the end is read normally, with no
  warning. (lead's call)
- `version` and `seq` must be JSON whole numbers: `1.0` or `"1"` is
  `has the wrong type`. (lead's call)
- The whole log is read from git in one piece, so a line has no length
  limit. (lead's call)
- A line with more than one fault gets one warning: the checks run in
  the order of R26's list of reasons, and the first fault found is the
  warning, e.g. a line with an unknown field and no `at` gives
  `unknown field "<f>"`. (lead's call)

## Testing

| Check | How |
|---|---|
| unit | Go tests next to the code. Table tests for the name rules (R2), the flag helper (TD2), the event log writer (R19) and reader (R25 to R27: one bad line for each reason), and paths (R41). The core's tests run real `git` in a temporary repo, with a fixed `Repo.Now`. |
| e2e | `TestMain` builds the binary once into a temporary folder. Each test runs it in its own temporary repo, and checks stdout, stderr, and the exit code. Shared helpers in `internal/e2e/helpers_test.go`; one file per task: `internal/e2e/<task_name>_test.go`. They live under `internal/` so `gofmt -l cmd internal` checks them. (lead's call) |

`internal/testrepo` makes the temporary repos for both kinds: `New`
makes a repo on `main` with one commit, with the test git config below.
(lead's call)

How the tests stay apart from the developer's own setup:

- Every git command in a test runs with `GIT_CONFIG_GLOBAL` set to a
  temporary file with a test name and email, and
  `GIT_CONFIG_NOSYSTEM=1`.
- Remote branches are made with `git update-ref refs/remotes/origin/slice/<name>`;
  no test has a real remote.
- A failed step (R18) comes from a test `pre-commit` hook that exits
  with 1. Git for Windows runs `#!/bin/sh` hooks too.
- The interrupt test (R18) starts `create` with a slow `pre-commit`
  hook and sends `os.Interrupt`. Go cannot send it to another program
  on Windows, so this test skips there; a unit test of the undo after a
  cancelled `ctx` runs on every system, and preview check P5 covers
  Windows by hand. (lead's call)
- R40: a test sets `GIT_TRACE` to a file, runs each command, and checks
  that no `fetch`, `pull`, or `push` ran.

## Decisions

- **TD1.** The Go module path is `github.com/BertPratya/sliceflow`.
  Why: it is where the repo is going after its rename, so the path is
  set once, and `go install` works after the rename.
  Not chosen: `github.com/BertPratya/flow-cli`, the repo's name today,
  because it must change after the rename.
  Cost: until the repo is renamed, `go install` with this path does not
  work. Builds and tests in the repo work.
- **TD2.** Flags are parsed with Go's standard `flag` package, plus a
  small helper that moves the flags in front of the positional
  arguments before parsing.
  Why: `flag` stops at the first argument that is not a flag, but the
  spec puts flags after the name, e.g.
  `sliceflow slice create login-google --title "Login with Google" --json`.
  The helper keeps sliceflow free of dependencies.
  Not chosen: `spf13/pflag` or `cobra`, because each is a new
  dependency, and cobra is large.
- **TD3.** `go.mod` says `go 1.25`. (lead's call)
  Why: it is the Go version on the development machine, and CI uses the
  version in `go.mod`.
  Not chosen: an older version, because no user needs one yet.
- **TD4.** An event is written from Go structs whose fields are
  declared in R19's order, with `encoding/json`.
  Why: the standard library keeps the declared order, so the order
  needs no extra code and no extra test helper.
  Not chosen: writing the JSON by hand, because escaping strings by
  hand is easy to get wrong.
- **TD5.** An event line is read in two steps: first into a map of
  field names to raw JSON values, then each field is checked by our own
  code.
  Why: R26 needs exact warnings, like `unknown field "f"` and
  `field "data.from.branch" has the wrong type`. Our own checks give
  them directly.
  Not chosen: `json.Decoder` with `DisallowUnknownFields`, because its
  error messages are not in R26's words, and parsing them would break
  with a new Go version.
- **TD6.** sliceflow reads a log with one `git show <ref>:<path>`, as
  one piece in memory.
  Why: a log is small, and one piece has no line length limit.
  Not chosen: `bufio.Scanner` line by line, because it fails on a line
  over 64 KB.
- **TD7.** The core is three plain functions, `Create`, `Status`, and
  `List`, as in Interfaces, Core API. Results carry their warnings, and
  errors carry the spec's words.
  Why: the CLI and a later server share one core, and neither copies
  the other's logic. With no shared state in this slice, plain
  functions are enough.
  Not chosen: an object with methods, because there is no state to
  keep yet; and calling git and the log from the CLI directly, because
  a server would have to copy that logic.
- **TD8.** sliceflow finds `git` once on `PATH`, and runs each git
  command with the repo folder as its working directory. (lead's call)
  Why: it uses the git the user already runs, with their settings.
  Not chosen: `git -C <dir>` on every call, because a working
  directory is the same thing without an extra argument.

## Risks and rollout

- **A hard kill leaves a half-made slice.** A crash, `kill -9`, or a
  power cut stops `create` before Undo on failure can run. What is left
  depends on the step:
  - after Flow 1, step 5: `.sliceflow/.gitignore`, which is ignored and
    harmless;
  - after Flow 1, step 6 or 7: a branch `slice/<name>`
    with no committed log. `slice list` shows it with `!`, `slice
    status` shows `Phase: unknown` and `.events.jsonl not found` (S18),
    and a new `create` with that name says it already exists.

  The commit itself is all or nothing in git. The user removes the
  leftover by hand:
  `git worktree remove --force .sliceflow/worktrees/<name>` and
  `git branch -D slice/<name>`. A command for this is in the backlog,
  with cancelling a slice.

## Open questions

None.
