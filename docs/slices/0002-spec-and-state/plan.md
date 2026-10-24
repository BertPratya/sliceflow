---
slice: 0002-spec-and-state
created: 2026-10-24
touches: []
---

# Plan

## Approach

A new state machine replays a slice's events through one table of moves,
and gives the state: the spec's state, the title, the last review, and
the last refused submit. The same table answers whether a move is allowed
now. It touches no files and no git. Every write goes one way: take the
lock, replay, check the move, write the log, commit, and undo when the
commit fails. A new spec reader turns `spec.md` into a structure with line
numbers, and the checks run on that structure. The reviewer sits behind an
interface; the placeholder is the only one in this slice.

| Part | Folder | Change |
|---|---|---|
| State machine: replay, and the table of moves | internal/state/ | new |
| Spec reader, and the checks | internal/spec/ | new |
| Reviewer interface, the placeholder, and the review file | internal/review/ | new |
| Event lines: the new events' data, and each event's line number | internal/eventlog/ | changed |
| The write path (lock, save, undo), submit, title, validate, and status | internal/slice/ | changed |
| The unfilled spec template, for `template show` | internal/template/ | changed |
| The new commands, and the new status output | internal/cli/ | changed |

The packages depend on each other in one direction only: `cli` uses
`slice`, `template`, and the types of `spec` and `state`; `slice` uses `state`,
`spec`, `review`, `eventlog`, `git`, and `template`; `state` uses
`eventlog`. Only `slice`, `git`, and a reviewer (which writes its review
file) touch files or run git.

New dependencies: None.

## Flows

### Flow 1: Replay a slice, and show its status

Covers: R1, R2, R3, R4, R5, R29, R30, R31, R34

1. `slice status` and `slice list` read the log from the last commit, as
   0001 does. `eventlog.Parse` now also checks the new events' `data`
   (R5), and keeps each event's line number, for R4's warnings. The line
   number is never written to the log.
2. `state.Replay` goes through the events in order. For each one,
   `state.Allowed` says whether the move is allowed now: allowed, it is
   applied; not allowed, it is skipped with R4's warning. The result is
   the `State`: the phase (R3, which replaces `eventlog.Phase`), the
   title (R34), and `spec.md`'s state, last review, and last refused
   submit.
3. When `spec.md` is `review_passed`, and a worktree has the branch
   `slice/<name>`, `slice` runs
   `git diff --quiet <reviewed commit> -- docs/slices/<name>/spec.md` in
   that worktree. Exit code 1 means changed after its review (R30). git
   compares through the same line-end rules as `git status`, so line
   ends that git converts do not count. Any other failure, e.g. exit
   code 128 when the reviewed commit is missing, means not compared: no
   change is shown, and the command goes on. (lead's call)
4. `slice` works out the `Spec:` and `Next:` texts from the `State` and
   step 3, by R29's table.
5. The CLI prints them after `Created:`, or adds `docs` and `next` to
   the JSON (R31), also in `slice list --json`.

### Flow 2: Check a spec

Covers: R12, R13, R14, R15, R16, R17, R18, R19, R37

1. The name is the argument, else the slice of the current branch
   (R12); the second line of R32 of 0001-slice-and-events names
   `spec validate`.
2. The first checks run in R11's order. The slice's worktree is the
   entry of `git worktree list --porcelain` whose branch is
   `refs/heads/slice/<name>`. No lock is taken.
3. `slice` reads `docs/slices/<name>/spec.md` from that worktree, from
   disk. A missing file gives R19's error.
4. `spec.Parse` reads the text into a `Doc`; `spec.Check` returns the
   problems (R15 to R18), sorted by line, then errors before warnings
   before notes.
5. The CLI prints the problem lines and the summary on stdout (R13,
   R37), and exits with R14's code.

### Flow 3: Submit, refused

Covers: R19, R20, R21, R22

1. Steps 1 and 2 of Flow 2, with `spec submit` in R32's line. Then
   `slice` takes the lock, and replays.
2. The state decides (R20): `draft`, or `review_passed` and changed after
   its review (Flow 1, step 3), go on; `submitted` goes to Flow 5; any
   other state gives R20's error, with no event.
3. `slice` reads `spec.md` from disk (R19), and checks it as
   `spec validate --strict` would.
4. With an error or a warning, it saves `spec.submitted` with `refused`
   (Saving an event: the files are `spec.md` and the log; the message is
   `<name>: spec submit refused`), and frees the lock.
5. `step` gets the problems first, then the result. The CLI prints the
   problem lines, then R22's line, whose last sentence depends on the
   state before (R22), on stdout, and exits with 1.

### Flow 4: Submit, accepted, and the review

Covers: R23, R24, R25, R26, R27, R28

1. Flow 3, steps 1 to 3, find no error and no warning.
2. If `reviews/spec-<next round>.md` exists in the slice folder on disk,
   R24's error; no event.
3. `slice` saves `spec.submitted` with `accepted` (files: `spec.md` and
   the log; message `<name>: spec submitted`), frees the lock, and calls
   `step`: the CLI prints `Submitted spec.md.` (R23).
4. It takes the lock again, replays, and checks that `review.started` is
   allowed; when another command moved the state first, R6's error. It
   writes the review file with only R25's front matter, saves
   `review.started` (files: the review file and the log; message
   `<name>: review <n> started`), frees the lock, and calls `step` (R25).
   `commit` is the commit step 3 made.
5. With no lock held, it calls the reviewer. The placeholder writes R26's
   text after the front matter.
6. It reads the result from the file (Interfaces → Review files).
7. It takes the lock again, replays, checks that `review.finished` is
   allowed, saves it (files: the review file and the log; R27's
   message), frees the lock, and calls `step`. The CLI prints
   `Review <n> passed.` and `Next: write the plan.` (R28), and exits
   with 0.

### Flow 5: Submit again from submitted

Covers: R20, R24, R25

1. Flow 3, step 2, finds `spec.md` `submitted`: `slice` skips the check
   and `spec.submitted`.
2. Flow 4 goes on from step 2 (R24), then step 4. `Submitted spec.md.` is
   not printed.
3. `commit` in `review.started` is the commit that added the last
   accepted `spec.submitted`. `slice` finds it by reading the log of each
   commit that changed it, newest first (`git log --format=%H -- <log>`),
   and takes the oldest one whose log still holds that event's `seq`.
   (lead's call)

### Flow 6: Set a title

Covers: R11, R32, R33, R34

1. The CLI needs exactly two arguments (R36).
2. The first checks run in R11's order; then `slice` takes the lock, and
   replays.
3. No good `slice.created`: R32's error. An empty title, or the title
   the slice has now: R33's error. No event.
4. `slice` saves `slice.titled` (files: only the log; message
   `<name>: title set`), frees the lock, and the CLI prints R32's line.

### Flow 7: Show the template

Covers: R35

1. `template show spec` prints the built-in template file as it is
   embedded, with its own front matter and every comment, on stdout.
2. It needs no git repo, and runs no git command (R35).
3. Any other name gives R35's error.

### Flow 8: Wrong use, and the order of the first checks

Covers: R11, R36

1. The CLI knows the command groups `slice`, `spec`, and `template`; a
   command is a group and a second word. `sliceflow spec` or
   `sliceflow template` alone prints the usage, and exits with 0.
2. Wrong use exits with 2, as R38 of 0001-slice-and-events says, with
   the new texts `missing title` and `missing template name`. The new
   commands have no `--json`, so it is `unknown flag --json`.
3. For `spec validate`, `spec submit`, and `slice title`, the first
   checks run in R11's order, then the lock (submit and title only),
   then the command's own checks.

### Shared rules

#### The lock

Covers: R6, R7

1. To take the lock, sliceflow creates `.sliceflow/locks/<name>.lock`
   in the root of the main checkout with "create only if it does not
   exist" (`os.O_CREATE|os.O_EXCL`), and writes the JSON line of Data
   into it. It first creates `.sliceflow/locks/`, and `.sliceflow/.gitignore`
   when it is missing.
2. When the file exists, it tries again every 100 ms, for up to 10 s
   (TD2). Then it reads the file for R7's message; a file it cannot read
   or parse gives `unknown` (R7).
3. It deletes the lock as soon as the event is saved or has failed, also
   on Ctrl-C (see Interrupts), with a `defer` set right after the lock is
   taken.
4. Only `spec submit` and `slice title` take the lock. Reads never do:
   they read the last commit of the slice branch, which never changes
   (R24 of 0001-slice-and-events).

#### Saving an event

Covers: R1, R2, R8, R9

With the lock held:

1. Read the log from the last commit of `slice/<name>` (as 0001 reads
   it), and replay it (`state.Replay`).
2. Run the command's prepare step on the state. It does the command's
   own checks, in the spec's order, and gives the spec's error when one
   fails, e.g. R20's, R19's, R24's, R32's, R33's, or R6's: then nothing
   is saved. Otherwise it gives the event's `data` and the files to
   write. (lead's call)
3. Check the move with `state.Allowed`, as a last guard. A prepare step
   that passed but a move that is not allowed is a bug: save nothing,
   and return a plain error.
4. Delete a `.events.jsonl.tmp` left by a crash. Write the committed log,
   a `\n` when it does not end with one (SD15), and the new line, to
   `.events.jsonl.tmp`; then rename it over `.events.jsonl` (R9).
5. Write the event's other files, e.g. the review file.
6. `git add -f -- <the files>`, then
   `git commit --only -m "<message>" -- <the files>` in the slice's
   worktree (TD1). `-f` adds them also when an ignore rule matches, as in
   0001.

#### Undo after a failed commit

Covers: R10

When step 6 of Saving an event fails, or Ctrl-C comes before it ends:

1. Write the log back as it was committed (by the same temporary file
   and rename).
2. Delete the files step 5 created; put back the files it changed, as
   they were committed.
3. Unstage what step 6 staged: `git reset -q -- <the files>`.
4. Never touch `spec.md`: it holds the lead's work.

R10's message then names git's reason, as R18 of 0001-slice-and-events
says.

#### Leftovers after a crash

Covers: R9, R24

A crash (power loss, `kill -9`) can leave these; none of them breaks the
committed log:

- `.events.jsonl.tmp`: removed at the next write.
- A log on disk newer than the commit: the next write starts again from
  the committed log, and overwrites it.
- A lock: R7's message; a person deletes it (0007 takes it over later).
- A review file with no commit: R24's message; a person moves it away.

#### Interrupts

Covers: R6, R10

`spec submit` and `slice title` catch Ctrl-C and `SIGTERM` as
`slice create` does in 0001 (its plan, Interrupts): the signal cancels
the `ctx`, a running git command is stopped, and the undo runs with
`context.WithoutCancel(ctx)`. R10's reason is then `interrupted`. The
lock is freed last. A signal while the reviewer runs stops it; no event
is saved for that step, so `spec.md` stays `submitted` or `in_review`.
Read commands keep Go's default: they stop at once.

#### Cross-platform paths

Covers: R7, R13, R19
Project rule: AGENTS.md, paths, line endings, and processes must work
  on Windows, macOS, and Linux.

As in 0001: files on disk use `filepath`, paths inside git use `path`,
and every printed path goes through `filepath.ToSlash`. The spec reader
reads `\r\n` as `\n` (R18). Renaming over an existing file works on
all three systems with `os.Rename`.

#### Tests

Covers: none
Project rule: AGENTS.md, every change that adds behavior adds a test.
  Tests create their own temporary git repos; never use shared services.

Every task adds the tests its behavior needs, as Testing says, and
names this rule in its follows. Each test makes its own repo with
`internal/testrepo`. No test uses the network.

#### Shared logic in internal

Covers: none
Project rule: AGENTS.md, shared Go logic goes in `internal/`, not
  `cmd/`.

`cmd/sliceflow/main.go` does not change. Everything new is in
`internal/`, as in Approach.

#### Git arguments

Covers: R8
Project rule: AGENTS.md, run each check or git command as its own
  process with its own arguments, never as one shell string.

As in 0001: every git command runs through `git.Run`, with its
arguments as a list. The file paths come after `--`. The slice name
passes R2 of 0001-slice-and-events before any git command uses it. The
title only goes into the JSON line and the output, never into a git
argument or a commit message.

#### No network

Covers: none
Project rule: AGENTS.md, sliceflow sends no data out: no telemetry, no
  uploads, no calls home.

0002 adds no network use. Every git command still goes through
`git.Run`, which keeps 0001's settings, and R40 of 0001-slice-and-events
keeps holding: 0001's no-network test runs the new commands too.

## Interfaces

### CLI

Changed. The spec describes it fully; the code follows the spec's words.

- `sliceflow spec validate [<name>] [--strict]`: R11 to R19, R37.
- `sliceflow spec submit [<name>]`: R6 to R12, R19 to R28, R37.
- `sliceflow slice title <name> <title>`: R11, R32 to R34.
- `sliceflow template show <template>`: R35.
- `sliceflow slice status [<name>] [--json]`: gains `Spec:` and `Next:`,
  and `docs` and `next` in JSON (R29 to R31). `slice list --json` gains
  the same two fields (R31). Both show the newest title (R34).
- Wrong use, and `sliceflow spec` or `sliceflow template` alone: R36. A
  command is now one word, or `slice`, `spec`, or `template` and a second
  word.

### Event log

Changed. Four new events, read and checked as R2, R4, and R5 say, and
written as R9 says. See Data.

### Review files

New. `docs/slices/<name>/reviews/spec-<round>.md`, as R25 to R27 say.
sliceflow reads the result from the file, as the review template says
for a doc review: when `## Needs human decision` is not `None.`, the
result is `needs_human`; else it is `passed`. A file with no such
section gives `needs_human`. (The section rule is the human's choice;
the missing section is the lead's call.)

### Lock file

New. See Data, and R6 and R7.

### Core API

Changed, in `internal/slice`. The CLI calls it; a later server calls
the same functions (SD13 of 0001-slice-and-events).

```go
func Validate(ctx context.Context, r Repo, name string, strict bool) (Validation, error)
func Submit(ctx context.Context, r Repo, name string, step func(Step)) (Submission, error)
func SetTitle(ctx context.Context, r Repo, name, title string) error
// Status and List: SliceStatus gains the spec's state and the next step.
```

- `Submit` calls `step` once for each step done, as it happens (e.g.
  "Submitted spec.md."), so a long review shows its start at once. The
  core never prints, as in 0001. (lead's call)
- Every failure the spec names is an `*Error` with the spec's words, as
  in 0001.

### State machine

New, in `internal/state`. No files, no git.

```go
func Replay(events []eventlog.Event) (State, []string) // the state, and R4's warnings
func Allowed(s State, e eventlog.Event) string         // R2: "" when allowed, else R4's reason
```

`State` holds the phase, the title, and for `spec.md`: its state, the
last review (round, reviewer, file, reviewed commit, result), and the
last refused submit. The table of moves is one list in this package;
later slices add rows to it. (lead's call)

### Spec reader

New, in `internal/spec`. No files, no git.

```go
func Parse(text []byte) Doc               // sections, items, labeled lines, with line numbers
func Check(d Doc, slice string) []Problem // R15 to R18
```

`Doc` is meant to be reused, e.g. by a later `spec show --json` (0015).
(lead's call)

### Reviewer

New, in `internal/review`. A contract for later reviewers (0004, 0011).

```go
type Reviewer interface {
	Name() string                                      // "placeholder"; the event's reviewer
	Review(ctx context.Context, req Request) error     // writes the review file
}
type Request struct {
	Slice, Doc, File, Commit string // File: the review file's full path
	Round                    int
	Worktree                 string
}
```

The reviewer writes the review file; sliceflow reads the result from
it (Review files). The placeholder writes R26's text.

## Data

Committed, on the slice branch:

- **The event log** `docs/slices/<name>/.events.jsonl` gets four new
  events: `slice.titled`, `spec.submitted`, `review.started`, and
  `review.finished`, with the `data` fields of R5, in the order of the
  spec's examples. `version` stays 1: the line form did not change. A
  0001 log holds only `slice.created`, which keeps its meaning, so no old
  log is changed. An older sliceflow skips the new events with the
  warning `unknown event` (R26 of 0001-slice-and-events).
- **The review files** `docs/slices/<name>/reviews/spec-<round>.md`, in
  the form of R25 and R26.

Local, in the main checkout, never committed:

- **The lock file** `.sliceflow/locks/<name>.lock`: one JSON line with
  `pid` (a whole number) and `since` (the time in UTC, as the events'
  `at`), e.g. `{"pid":4312,"since":"2026-10-24T10:02:00Z"}`. It exists
  only while one event is saved. `.sliceflow/` ignores itself through its
  `.gitignore` (R6 of 0001-slice-and-events); when that file is missing,
  taking the lock writes it first. (lead's call)

Local, in the slice's worktree, for a moment only:

- **The temporary log** `docs/slices/<name>/.events.jsonl.tmp`: the whole
  new log, renamed over `.events.jsonl` once written (R9). A file left by
  a crash is removed at the next write. (lead's call)

## Testing

| Check | How |
|---|---|
| unit | Go tests next to the code. `state`: one table test per row of R2's table, and each warning of R4. `spec`: one small spec per check of R15 to R18 that breaks just that check, plus two real files as test data: a fresh spec from `template.Spec`, which gives 0 errors, and 0001's approved spec, which gives 0 errors and the notes for R11, R40, and R43. `eventlog`: each new event's fields (R5). `slice`: real git in a temporary repo, a fixed `Repo.Now`, and a short lock wait (`Repo.LockWait`, 10 s by default). Two writes started at once both save, with `seq` 2 and 3. (lead's call) |
| e2e | As in 0001: `TestMain` builds the binary once; each test runs it in its own temporary repo; helpers in `internal/e2e/helpers_test.go`; one file per task: `internal/e2e/<task_name>_test.go`. One test per scenario. |

How the tests make things fail:

- A failed commit comes from a test git hook that exits with 1: a
  `pre-commit` hook, or a `commit-msg` hook that refuses only messages
  with `review` (S23). Git for Windows runs `#!/bin/sh` hooks too.
- The busy test (S19) makes the lock file itself, and waits the full
  10 s: one slow test, which checks exactly what the user sees. (lead's
  call)
- Ctrl-C: as in 0001, the e2e test (task 16) skips on Windows, and a
  unit test of the undo after a cancelled `ctx` (task 11) runs on every
  system.
- A log with a bad line, or a move that is not allowed (S3, S4), is
  written by the test and committed with plain git.

## Decisions

- **TD1.** sliceflow commits with `git commit --only -- <its files>`.
  Why: git then commits those files as they are on disk, and leaves
    what is staged for other files staged and out of the commit (R8,
    S22).
  Not chosen: a separate index file (`GIT_INDEX_FILE`), because it is
    more code for the same result.
- **TD2.** The lock is tried again every 100 ms, for up to 10 s.
  Why: a write usually takes well under a second, so a second command
    rarely waits long; 10 s is the spec's limit (R7).
  Not chosen: a longer wait for slow git hooks, because a real leftover
    lock would then keep the user waiting (the human's choice; see Risks).

## Risks and rollout

- **A slow git hook gives a false "busy".** The lock is held while
  `git commit` runs the project's hooks. A hook that takes 40 s makes a
  second command give up after 10 s, although nothing crashed. This is
  rare: one lead works on one slice. The message is clear, and running
  the command again works. The human chose to keep 10 s; if it happens
  in practice, a later slice can wait longer or show a waiting line.

## Open questions

None.
