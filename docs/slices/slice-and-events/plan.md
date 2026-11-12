---
slice: slice-and-events
created: 2026-11-03
touches:
  - go.mod
  - templates/default/spec.md
  - .github/workflows/test.yml
---

# Plan

## Approach

sliceflow is built as six packages, one job each. The lock and the event
log come first, because everything else writes events. The git runner
and the name check come next. The create flow comes last of the core
pieces, because it uses all of them. Status and list only read, so they
come at the end.

This plan sets the build order, the packages, and the choices that cover
the whole slice. It does not set the inside of a piece: its functions,
inputs, outputs, and steps. The human designs those one piece at a time,
in the loop (AGENTS.md), and each design is written into tasks.md when
it is done.

| Part | Folder | Change |
|---|---|---|
| File lock | internal/filelock/ | new |
| Event log | internal/eventlog/ | new |
| Git runner and git operations | internal/git/ | new |
| Name check, create, find a log, status, list | internal/slice/ | new |
| Commands, help, wrong use, exit codes | internal/cli/ | new |
| Start, crash, Ctrl+C | cmd/sliceflow/ | changed |
| Temporary git repos for tests | internal/gittest/ | new |
| End-to-end tests | e2e/ | new |
| Spec template | templates/default/spec.md | changed |
| Tests on three systems | .github/workflows/test.yml | new |

New dependencies:
- `golang.org/x/sys`, for the lock calls of each system (TD2).

### How the pieces connect

Each arrow means "calls". Calls only go down, never up.

```
cmd/sliceflow       starts the program, catches a crash, handles Ctrl+C
      |
internal/cli        the routing: reads the command line with a table,
      |             picks the command, turns the result into an exit code
      |
      |-- slice create  --> internal/slice  (create flow)
      |-- slice status  --> internal/slice  (find a log, status)
      '-- slice list    --> internal/slice  (list)

internal/slice  ------> internal/git        (runner, then git operations)
      '-------------->  internal/eventlog  ------> internal/filelock
```

- `internal/cli` holds only the routing. It knows the commands and the
  options, but not how a slice is made.
- `internal/slice` holds the work. It knows nothing about the command
  line, so its tests call it directly.
- `internal/git`, `internal/eventlog`, and `internal/filelock` know
  nothing about slices.

### Build order

"Human" means the human writes the code, and the agent writes the tests
first (AGENTS.md, "The human's piece"). "Agent" means the agent writes
the code from the human's design. The human designs every piece. The
hours are for the human's pieces, for the stop-loss rule: a piece that
takes more than twice its estimate goes to the agent.

| Step | Piece | Who | Requirements | Estimate |
|---|---|---|---|---|
| 0 | Tests on Windows, macOS, and Linux in CI | Agent | none | |
| 1 | File lock | Human | R30 | 4–6 h |
| 2 | Event log, with the lock | Human | R27, R30–R35 | 6–10 h |
| 3 | Name check | Human | R19–R21 | 2–3 h |
| 3b | Control-character check | Agent | R29 | |
| 4 | Git runner | Human | R14 | 3–5 h |
| 5 | Git operations, and finding the project | Agent | R10–R13, R15–R18, R23–R26 | |
| 6 | Commands, help, wrong use, exit codes, crash | Agent | R67–R76 | |
| 7 | Create flow with its undo list | Human | R1–R9, R22, R28, R36–R41 | 10–15 h |
| 8 | Finding a log, and status | Agent | R42–R55 | |
| 9 | List | Agent | R56–R66 | |

The human's pieces add up to about 25–40 hours.

Step 0 comes first because the human works on Linux only, and the lock
must be tested on Windows and macOS from its first commit. Step 6 comes
before step 7 because most tests of the create flow are e2e tests: they
run the real program, so the `slice create` command must exist. Until
step 7 is done, the command calls a create function that does not work
yet, and its tests fail.

## Flows

### Flow 1: Create a slice

1. `internal/cli` reads the command line. A wrong command line stops
   here (R70 to R75), with `--help` first.
2. The create flow checks, in the order of R40: control characters,
   the name, git found, the repository, the base, then a name that is
   not free.
3. It takes the git steps in the order of R37: create the branch, create
   the worktree, then write the two files and make the commit. The first
   event is written under the lock (SD12).
4. Each thing it makes goes on an undo list. When a later step fails, or
   the user presses Ctrl+C, it removes the things on the list, last made
   first (R36, R39). What it could not remove is reported (R41).
5. On success, it prints the lines of R8.

### Flow 2: See one slice

1. `internal/cli` reads the command line, and the name is checked as
   in R40.
2. The log is found in the order of R51: the worktree file, the branch
   tip, then the `main` tip.
3. The first line must be a readable `slice.created` (R55). The other
   lines are read, and bad ones are named in a warning (R53, R54).
4. The status lines are printed (R43 to R50).

### Flow 3: See all slices

1. The candidates are the branches `slice/*` and the folders in `main`
   (R57).
2. For each candidate, the log is found as in Flow 2, step 2.
3. Only the first line of each log is read (R61). Warnings come first,
   then the table (R59, R63 to R65).

### Shared rules

#### No data out

Project rule: AGENTS.md, "sliceflow sends no data out".
sliceflow makes no network call. It never runs a git command that talks
to a remote, e.g. fetch (SD2).

#### Paths on every system

Project rule: AGENTS.md, "Paths, line endings, and processes work on
  Windows, macOS, and Linux."
Paths are built with `filepath.Join`, never with "/". Paths that git
prints are turned into paths of the system before they are printed (R9).
The tests run on all three systems in CI (step 0).

#### One process per git command

Project rule: AGENTS.md, "Run each git command or check as its own
  process with its own arguments, never as one shell string."
Every git call goes through the git runner of step 4. The runner takes
the arguments as a list, and never starts a shell.

#### Tool errors word for word

Project rule: AGENTS.md and design 1.17, "show the tool's own message
  word for word, after one line of sliceflow's own".
The git runner keeps everything git printed to its error output. The
create flow prints it as in R37. sliceflow has its own message only
where the spec gives one, e.g. a branch that exists (R24).

#### Code in internal

Project rule: AGENTS.md, "Shared logic goes in internal/; cmd/ only
  starts the program."
`cmd/sliceflow/main.go` only starts the commands, catches a crash, and
turns Ctrl+C into a cancelled context (TD4, TD5).

#### Tests make their own repos

Project rule: AGENTS.md, "Tests create their own temporary git repos."
Every test that needs git makes a new repo in a temporary folder, with
the helpers in `internal/gittest` (TD6).

## Interfaces

### CLI

The commands `slice create`, `slice status`, and `slice list`, their
options, their output, their messages, and their exit codes are given in
full by the spec (Stories 1 to 4, R1 to R76). This plan adds nothing to
them.

## Data

A new file per slice: `docs/slices/<name>/.events.jsonl`, in the format
of R27 and R32. There is no old data to move.

The lock may need a file of its own. Where it lives is part of the lock
piece's design (step 1).

## Testing

| Check | How |
|---|---|
| unit | Go tests next to the code, in each package's `*_test.go`. Folders from `t.TempDir()`. Git repos from `internal/gittest`. The lock test (R30) uses two real processes (TD7). |
| e2e | `e2e/main_test.go` builds the program once, in `TestMain`. Each test makes a repo, runs the program with `os/exec`, and checks the output, the error output, and the exit code. One file per piece, e.g. `e2e/create_test.go`. Each test runs with its own `HOME`, `GIT_CONFIG_NOSYSTEM=1`, and a global git config that sets a name and an email. |

Tests compare sliceflow's own lines word for word. They never compare
git's own text, because it changes between git versions and languages.
They only check that it is there.

The test for "git cannot work out a name and an email" (S11, S12, S22)
sets `user.useConfigOnly=true` and gives git no name, no email, and no
global config. git then refuses to guess an email from the host name,
so the test fails the same way on every machine (TD8).

### Trying it by hand

#### Setup

1. Run `go build -o sliceflow ./cmd/sliceflow`.
2. Make a test repo in a temporary folder: `git init demo`, then add one
   commit in it.
3. In `demo`, run `../sliceflow slice create login`.
   Expect: the lines of R8, and a new branch `slice/login`.

#### Manual checks

None.

## Decisions

- **TD1.** Read the command line with a table of our own: a map from
  the command words ("slice create") to each command's help, whether it
  takes a name, its options, and the function that does its work. One
  function, `checkLine`, reads the words from the left against the
  table. Changed on 2026-11-10, in the review of PR #1: the first build
  used Cobra.
  Why: sliceflow has three commands. The table is short, has no outside
    library, and prints only the spec's own text. Go's own `go` command
    and Terraform read their command lines the same way.
  Not chosen: Cobra. It needed each of its own texts turned off: usage
    on an error, the `help` and `completion` commands, its help layout.
    Not chosen either: the standard `flag` package. It stops reading
    options at the first word that is not an option, and has no
    commands.
  Cost: we write and test the reading ourselves, e.g. options before or
    after the name, `--title=<text>`, and the last value counting (R71).
- **TD2.** Write the lock with `golang.org/x/sys`: `unix.Flock` on Linux
  and macOS, `windows.LockFileEx` on Windows, one file for each, picked
  by Go when it builds.
  Why: the lock is the human's first-time piece. Writing it teaches how
    locking works and how it differs between systems.
  Not chosen: the `gofrs/flock` library. It is faster to use, but it
    skips what the piece is meant to teach.
  Cost: more code, and code for two systems. The human works on Linux
    only, so Windows and macOS bugs show up only in CI.
- **TD3.** The packages of the Approach table, one job each.
  Why: each human piece is a package of its own, so it can be written
  and tested alone. The names match the commit areas of AGENTS.md.
  Not chosen: one package for everything. It would mix the human's
  pieces with the agent's.
  Cost: a few more imports between packages.
- **TD4.** `main` catches a crash (a panic), prints what Go would print,
  and exits with 1 (SD17). (lead's call)
  Why: Go exits with 2 after a crash, and 2 means a wrong command line
    in this spec.
  Not chosen: letting Go crash. The exit code would be wrong.
  Cost: None.
- **TD5.** `main` turns Ctrl+C into a cancelled `context.Context`, and
  passes it down to the create flow and the git runner. (lead's call)
  Why: it is how Go stops work in progress, and `os/exec` stops a running
    git when its context is cancelled.
  Not chosen: catching the signal inside the create flow. Then the
    create flow would depend on how the program was started.
  Cost: the clean-up after Ctrl+C must run its git commands with a new
    context, not the cancelled one, or it stops at once.
- **TD6.** One helper package, `internal/gittest`, makes temporary git
  repos for tests. (lead's call)
  Why: every package that runs git needs the same setup.
  Not chosen: setup copied into each test file.
  Cost: None.
- **TD7.** The lock test starts a second copy of the test program, as a
  second process. (lead's call)
  Why: on Linux and macOS, some locks do not block a second lock inside
    the same process. A test with two goroutines could pass while the
    lock does not work.
  Not chosen: two goroutines in one test.
  Cost: the test is harder to read.
- **TD8.** The identity test uses `user.useConfigOnly=true`. (lead's
  call)
  Why: without it, git guesses an email on hosts like `mac.local`, and
    the commit works, so the test passes on some machines and not on
    others.
  Not chosen: an empty `HOME` alone. It does not stop the guess.
  Cost: None.

## Risks and rollout

- **Windows locks block reading.** A lock taken with `LockFileEx` stops
  other programs from reading the locked part of the file. If the lock is
  held on the log itself, `status` could fail on Windows while a write is
  in progress. The lock design (step 1) must handle this, and CI runs the
  tests on Windows.
- **A cleaned-up folder that was there before.** `git worktree remove`
  deletes the worktree folder. When the folder was there and empty before
  `create`, the clean-up must make it again (R36).
- **Ctrl+C reaches git too.** In a terminal, Ctrl+C stops git as well as
  sliceflow. The clean-up must not count on the stopped git, and must
  use a new context (TD5). Sending Ctrl+C from a test is easy on Linux
  and macOS but hard on Windows. If it does not work on Windows, that
  test is skipped there, and the skip is written in the test.
- **The template's `title:` line.** `templates/default/spec.md` still has
  a `title:` line in its front matter. The title now lives only in the
  event log (SD9), so step 7 removes the line.
- **Size.** The slice has 76 requirements. The build order splits it into
  ten steps, and the stop-loss rule limits the human's pieces.

## Open questions

None.
