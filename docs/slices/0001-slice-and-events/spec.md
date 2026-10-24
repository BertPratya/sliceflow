---
slice: 0001-slice-and-events
title: Slice and events
created: 2026-10-23
ideas: []
---

# Spec

## Summary

A user creates a slice with one command, and sees one slice or all slices
with two more. Every step that matters is saved as an event in the slice's
event log, and the status is read from that log.

## Why

Today the lead creates a slice by hand, step by step, and keeps its status
by hand. A step can be skipped, or done a little differently each time,
and the status can drift from what really happened. Steps with one right
answer belong in a program. This slice builds the base every later slice
uses: a slice, its event log, and a status worked out from the log.

## User stories

### Story 1: Create a slice

A user starts a new slice with one command, and gets a branch, a worktree,
a spec to fill in, and an event log.

- **S1.** **Given** a repo on `main` at commit `9dd7ebf`, with no
  `git config sliceflow.base`,
  **When** I run `sliceflow slice create spec-submit`,
  **Then** sliceflow creates the slice, and prints:
  ```
  $ sliceflow slice create spec-submit
  Created slice spec-submit.
    Branch:   slice/spec-submit (from main @ 9dd7ebf)
    Worktree: .sliceflow/worktrees/spec-submit
    Spec:     docs/slices/spec-submit/spec.md
  Next: write the spec in .sliceflow/worktrees/spec-submit/docs/slices/spec-submit/spec.md
  ```
  and the exit code is 0,
  and `.sliceflow/.gitignore` holds the one line `*`,
  and the branch `slice/spec-submit` has one new commit,
  `spec-submit: slice created`, with two files:
  `docs/slices/spec-submit/spec.md` and
  `docs/slices/spec-submit/.events.jsonl`,
  and `.events.jsonl` holds one line, like:
  ```
  {"version":1,"seq":1,"event":"slice.created","at":"2026-10-24T07:42:00Z","data":{"from":{"branch":"main","commit":"9dd7ebf4c2a1e8d3b5f6a7c8d9e0f1a2b3c4d5e6"}}}
  ```
  Shows: R1, R3, R4, R5, R6, R7, R8, R9, R19, R20, R41
- **S2.** **Given** `git config sliceflow.base develop`,
  **When** I run
  `sliceflow slice create login-google --title "Login with Google" --json`,
  **Then** the slice starts from `develop`, and sliceflow prints one JSON
  object:
  ```
  {"slice":"login-google","branch":"slice/login-google","from":{"branch":"develop","commit":"<full commit id>"},"worktree":"<full path>/.sliceflow/worktrees/login-google","spec":"docs/slices/login-google/spec.md"}
  ```
  and the event's `data` holds `"title":"Login with Google"`.
  Shows: R3, R10, R20, R41
- **S3.** **Given** `git config sliceflow.base develop`,
  **When** I run `sliceflow slice create hotfix-ui --from main`,
  **Then** the slice starts from `main`, because `--from` wins.
  Shows: R3
- **S4.** **Given** I am inside the worktree of another slice,
  **When** I run `sliceflow slice create spec-submit`,
  **Then** the new worktree is in `.sliceflow/worktrees/spec-submit` at the
  root of the main checkout, not inside the worktree I am in.
  Shows: R5
- **S5.** **When** I run `sliceflow slice create "My Slice"`,
  **Then** sliceflow changes nothing, prints on stderr:
  ```
  Error: invalid slice name "My Slice": use 1-64 lowercase letters, digits, and dashes, starting with a letter or digit. Nothing was changed.
  ```
  and the exit code is 1.
  Shows: R2, R14, R38
- **S6.** **When** I run `sliceflow slice create con`,
  **Then** sliceflow changes nothing and prints
  `Error: invalid slice name "con": it is a reserved name on Windows. Nothing was changed.`
  Shows: R2, R14
- **S7.** **Given** a teammate pushed `slice/spec-submit`, and my last
  fetch brought `origin/slice/spec-submit`,
  **When** I run `sliceflow slice create spec-submit`,
  **Then** sliceflow changes nothing and prints
  `Error: slice spec-submit already exists: branch origin/slice/spec-submit. Nothing was changed.`
  Shows: R15
- **S8.** **Given** `git config sliceflow.base develop`, and no branch
  `develop`,
  **When** I run `sliceflow slice create spec-submit`,
  **Then** sliceflow changes nothing and prints
  `Error: base branch develop does not exist (from git config sliceflow.base). Nothing was changed.`
  Shows: R16
- **S9.** **Given** an old folder `.sliceflow/worktrees/spec-submit` was
  left behind, and no slice `spec-submit` exists,
  **When** I run `sliceflow slice create spec-submit`,
  **Then** sliceflow changes nothing and prints
  `Error: folder .sliceflow/worktrees/spec-submit already exists. Remove it, or pick another name. Nothing was changed.`
  Shows: R17
- **S10.** **Given** a folder that is not a git repo,
  **When** I run `sliceflow slice create spec-submit` in it,
  **Then** sliceflow prints
  `Error: not inside a git repository. Nothing was changed.`
  Shows: R12
- **S11.** **Given** a git repo with no commits yet,
  **When** I run `sliceflow slice create spec-submit`,
  **Then** sliceflow prints
  `Error: the repository has no commits yet. Make a first commit, then try again. Nothing was changed.`
  Shows: R13
- **S12.** **Given** a git pre-commit hook that refuses every commit,
  **When** I run `sliceflow slice create spec-submit`,
  **Then** sliceflow removes the worktree and the branch it made, and
  prints
  `Error: could not create slice spec-submit: <git's message>. Everything sliceflow made was removed.`
  and the exit code is 1.
  Shows: R8, R18, R38

### Story 2: See one slice

A user, or an AI agent, sees where one slice stands.

- **S13.** **Given** the slice `spec-submit` from S1, and I am on `main`,
  **When** I run `sliceflow slice status spec-submit`,
  **Then** sliceflow prints:
  ```
  $ sliceflow slice status spec-submit
  Slice:    spec-submit
  Title:    spec-submit
  Phase:    planning
  From:     main @ 9dd7ebf
  Created:  2026-10-24 07:42 UTC
  ```
  Shows: R24, R25, R29
- **S14.** **Given** I am in the worktree of `spec-submit`, on branch
  `slice/spec-submit`,
  **When** I run `sliceflow slice status`,
  **Then** sliceflow prints the status of `spec-submit`, as in S13.
  Shows: R31
- **S15.** **Given** I am on `main`,
  **When** I run `sliceflow slice status`,
  **Then** sliceflow prints on stderr:
  ```
  Error: not on a slice branch (you are on main).
  Give a name: sliceflow slice status <name>
  See all slices: sliceflow slice list
  ```
  and the exit code is 1.
  Shows: R32, R38
- **S16.** **When** I run `sliceflow slice status login-google --json`,
  **Then** sliceflow prints one JSON object:
  ```
  {"slice":"login-google","title":"Login with Google","phase":"planning","from":{"branch":"develop","commit":"<full commit id>"},"created":"2026-10-24T08:10:00Z","warnings":[]}
  ```
  Shows: R30
- **S17.** **Given** the log of `spec-submit` has a line 2 that is not
  valid JSON, and a line 3 with `seq` 4,
  **When** I run `sliceflow slice status spec-submit`,
  **Then** sliceflow prints the status from the good lines, and prints on
  stderr:
  ```
  Warning: spec-submit: .events.jsonl line 2: not valid JSON; skipped
  Warning: spec-submit: .events.jsonl line 3: seq 4, expected 2
  ```
  and the exit code is 0.
  Shows: R26, R27, R28
- **S18.** **Given** the branch `slice/old-idea` has no `.events.jsonl`,
  **When** I run `sliceflow slice status old-idea`,
  **Then** sliceflow prints the status with `Phase:    unknown`, and `-`
  for `From:` and `Created:`, and prints on stderr
  `Warning: old-idea: .events.jsonl not found`.
  Shows: R25, R26, R29
- **S19.** **When** I run `sliceflow slice status no-such-slice`,
  **Then** sliceflow prints on stderr
  `Error: no slice named no-such-slice. See all slices: sliceflow slice list`
  and the exit code is 1.
  Shows: R33

### Story 3: See all slices

A user, or an AI agent, sees every slice in progress.

- **S20.** **Given** the slices `spec-submit` and `login-google` are in
  progress, and `first-slice` is merged into the base branch,
  **When** I run `sliceflow slice list`,
  **Then** sliceflow prints the slices in progress, oldest first:
  ```
  $ sliceflow slice list
  NAME          TITLE              PHASE     CREATED (UTC)
  spec-submit   spec-submit        planning  2026-10-24 07:42
  login-google  Login with Google  planning  2026-10-24 08:10
  ```
  Shows: R34, R35
- **S21.** **Given** the slices of S20,
  **When** I run `sliceflow slice list --all`,
  **Then** `first-slice` is in the list too, at its place by created time.
  Shows: R35
- **S22.** **Given** the log of `spec-submit` has a bad line, as in S17,
  **When** I run `sliceflow slice list`,
  **Then** its row shows `planning !` as the phase, and the other rows
  print normally.
  Shows: R36
- **S23.** **When** I run `sliceflow slice list --json`,
  **Then** sliceflow prints one JSON object, with each slice in the form
  of S16:
  ```
  {"slices":[{"slice":"spec-submit","phase":"planning",...,"warnings":[]},{"slice":"login-google",...}],"warnings":[]}
  ```
  Shows: R37
- **S24.** **Given** no slice is in progress,
  **When** I run `sliceflow slice list`,
  **Then** sliceflow prints `No slices in progress.`
  Shows: R34
- **S25.** **Given** a repo whose main branch is `master`, with no
  `git config sliceflow.base`, and the slice `spec-submit` in progress,
  **When** I run `sliceflow slice list`,
  **Then** sliceflow prints the row of `spec-submit`, and prints on stderr
  `Warning: base branch main does not exist (the default); merged slices are not shown`
  and the exit code is 0.
  Shows: R21, R23

### Story 4: Get help, and wrong use

A user learns how to use a command, and sees what was wrong.

- **S26.** **When** I run `sliceflow slice create --help`,
  **Then** sliceflow prints the usage of `slice create` on stdout, and
  the exit code is 0.
  Shows: R39
- **S27.** **When** I run `sliceflow slice list --colour`,
  **Then** sliceflow prints on stderr
  `Error: unknown flag --colour. Run sliceflow --help.`
  and the exit code is 2.
  Shows: R38
- **S28.** **Given** a folder that is not a git repo,
  **When** I run `sliceflow slice list` in it,
  **Then** sliceflow prints
  `Error: not inside a git repository. Nothing was changed.`
  and the exit code is 1.
  Shows: R22

## Key concepts

- **Slice name:** the slice's ID, exactly as the user typed it, e.g.
  `spec-submit`. It has no number added by sliceflow.
- **Base branch:** the branch a slice starts from and later merges into.
  It is `--from`, else `git config sliceflow.base`, else `main`.
- **Event log:** the file `docs/slices/<name>/.events.jsonl`, one event
  per line. It is the only source of truth for a slice. Only sliceflow
  writes it.
- **A slice exists** when a local branch `slice/<name>` exists, or the
  folder `docs/slices/<name>/` is on the base branch.
- **In progress:** a slice whose local branch `slice/<name>` exists, and
  whose folder is not on the base branch yet. A branch with no log is in
  progress too.
- **Merged:** a slice whose folder `docs/slices/<name>/` is on the base
  branch.
- **Short commit:** the first 7 characters of a commit id.
- **Phase:** where a slice stands. In this slice: `planning`, or
  `unknown` when the log does not say.

## Requirements

### Creating a slice

- **R1.** When the user runs `sliceflow slice create <name>`, sliceflow
  shall use `<name>` as the slice name exactly as typed, with no number
  added.
  Check: e2e
- **R2.** sliceflow shall accept a slice name only when it has 1 to 64
  characters, uses only lowercase letters a-z, digits, and dashes, starts
  with a letter or a digit, and is not a name Windows reserves (`con`,
  `prn`, `aux`, `nul`, `com1` to `com9`, `lpt1` to `lpt9`).
  Check: unit
- **R3.** When creating a slice, sliceflow shall use as the base branch
  the value of `--from`, else the value of `git config sliceflow.base`,
  else `main`.
  Check: e2e
- **R4.** When every check passes, sliceflow shall create the branch
  `slice/<name>` at the newest commit of the base branch.
  Check: e2e
- **R5.** sliceflow shall create the slice's worktree in
  `.sliceflow/worktrees/<name>/` at the root of the main checkout, also
  when the command runs inside another worktree.
  Check: e2e
- **R6.** When the `.sliceflow/` folder has no file `.gitignore`,
  sliceflow shall write one in it that holds the one line `*`.
  Check: e2e
- **R7.** sliceflow shall add exactly two files to the slice branch:
  `docs/slices/<name>/spec.md`, from the spec template built into
  sliceflow, and
  `docs/slices/<name>/.events.jsonl`, with one `slice.created` event.
  The built-in template is this repo's spec template without the
  `title:` line and its rules about the title. Its front matter is
  filled with `slice`, `created` (today's date in UTC, as `YYYY-MM-DD`),
  and `ideas: []`.
  Check: e2e
- **R8.** sliceflow shall commit those two files on the slice branch as
  one commit, with the message `<name>: slice created`. It is a normal
  git commit: it uses the user's git name and email, and the user's git
  hooks and commit signing run as they are set up.
  Check: e2e
- **R9.** When the slice is created, sliceflow shall print the lines of
  S1: `Created slice <name>.`, the branch with its base branch and short
  commit, the worktree path, the spec path, and the `Next:` line. The
  `Worktree:` and `Next:` paths are relative to the root of the main
  checkout, also when the command runs in a subfolder or in another
  worktree. The `Spec:` path is the path inside the worktree.
  Check: e2e
- **R10.** Where `--json` is given, sliceflow shall print, instead of R9's
  lines, one JSON object with `slice`, `branch`, `from` (`branch` and the
  full `commit`), `worktree` (the full path), and `spec` (the path inside
  the worktree).
  Check: e2e

### Checks before creating

Each check runs before sliceflow creates anything. Each error goes to
stderr, changes nothing, and exits with code 1.

- **R11.** sliceflow shall run the checks in this order, and stop at the
  first that fails: R12, R13, R14, R16, R15, R17.
  Check: e2e

- **R12.** If the command does not run inside a git repo, then sliceflow
  shall print `Error: not inside a git repository. Nothing was changed.`
  Check: e2e
- **R13.** If the repo has no commits, then sliceflow shall print
  `Error: the repository has no commits yet. Make a first commit, then try again. Nothing was changed.`
  Check: e2e
- **R14.** If the name breaks R2, then sliceflow shall print
  `Error: invalid slice name "<name>": use 1-64 lowercase letters, digits, and dashes, starting with a letter or digit. Nothing was changed.`
  or, for a name Windows reserves,
  `Error: invalid slice name "<name>": it is a reserved name on Windows. Nothing was changed.`
  Check: e2e
- **R15.** If a local branch `slice/<name>` exists, or a branch
  `slice/<name>` of any remote as git last saw it, or a folder
  `docs/slices/<name>/` on the base branch, then sliceflow shall print
  `Error: slice <name> already exists: <where>. Nothing was changed.`
  where `<where>` is `branch slice/<name>`, `branch <remote>/slice/<name>`,
  or `folder docs/slices/<name> on <base branch>`. When the name is in
  more than one place, `<where>` names the first found, in that order.
  Check: e2e
- **R16.** If the base branch is not a local branch, then sliceflow shall
  print
  `Error: base branch <branch> does not exist (<source>). Nothing was changed.`
  where `<source>` is `from --from`, `from git config sliceflow.base`, or
  `the default`.
  Check: e2e
- **R17.** If the folder `.sliceflow/worktrees/<name>/` exists, then
  sliceflow shall print
  `Error: folder .sliceflow/worktrees/<name> already exists. Remove it, or pick another name. Nothing was changed.`
  Check: e2e
- **R18.** If a step fails after sliceflow started creating, then
  sliceflow shall remove what it made, newest first, and print
  `Error: could not create slice <name>: <reason>. Everything sliceflow made was removed.`
  If it cannot remove something, the message ends with
  `Could not remove: <list>.` instead of the last sentence.
  Check: e2e

### The event log

- **R19.** sliceflow shall write the event log as UTF-8 with no byte order
  mark, one JSON object per line, each line ending with `\n` on every
  system. Each event has these fields, in this order, and no others:
  `version` (1), `seq` (1 for the first event, then 1 more each time),
  `event` (the event's name), `at` (the time in UTC, as RFC 3339 to the
  second, ending in `Z`), and `data` (an object).
  Check: unit
- **R20.** The `data` of `slice.created` shall hold `from`, with the base
  branch's name (`branch`) and the full commit id (`commit`), and `title`
  only when the user gave `--title`.
  Check: e2e

### Reading a slice

- **R21.** For `slice status` and `slice list`, sliceflow shall use as
  the base branch the value of `git config sliceflow.base`, else `main`.
  Check: e2e
- **R22.** If `slice status` or `slice list` runs outside a git repo, or
  in a repo with no commits, then sliceflow shall print the error of R12
  or R13, and exit with code 1.
  Check: e2e
- **R23.** If the base branch does not exist, then `slice status` and
  `slice list` shall print on stderr
  `Warning: base branch <branch> does not exist (<source>); merged slices are not shown`
  with `<source>` as in R16, and go on with the local `slice/` branches.
  Check: e2e
- **R24.** sliceflow shall read a slice's log from the newest commit of
  the local branch `slice/<name>` when that branch exists, else from the
  newest commit of the base branch. It shall never read the log from files
  that are not committed.
  Check: e2e
- **R25.** sliceflow shall give a slice the phase `planning` when its log
  has a good `slice.created` event, and `unknown` otherwise.
  Check: unit
- **R26.** If a line of the log is bad, then sliceflow shall skip that
  line and add the warning `.events.jsonl line <n>: <reason>; skipped`.
  The reasons are: `blank line`, `not valid JSON`,
  `unknown version <v>`, `unknown event "<e>"`, `unknown field "<f>"`,
  `missing field "<f>"`, `field "<f>" has the wrong type`, and
  `field "<f>" has a wrong value`. A wrong value is: an `at` that is not
  in the form of R19, e.g. `2026-10-24T07:42:00Z`, a `commit` that is not 40 or
  64 lowercase hex characters, or an empty `branch` or `title`. A field
  inside `data` is named like `data.from.branch`. The order of the fields
  is not checked when reading. If the log does not exist, sliceflow
  shall add the warning `.events.jsonl not found`.
  Check: unit
- **R27.** If a line's `seq` is not 1 more than the last good line's
  `seq`, then sliceflow shall keep the line and add the warning
  `.events.jsonl line <n>: seq <found>, expected <expected>`. The kept
  line counts as the last good line: `seq` 1, 2, 4, 5 gives one warning.
  Check: unit
- **R28.** sliceflow shall print each warning about a slice on stderr as
  `Warning: <name>: <warning>`, and the warning of R23 once, as R23
  writes it. A warning shall never stop a command or change its exit
  code.
  Check: e2e

### Status of one slice

- **R29.** When the user runs `sliceflow slice status <name>`, sliceflow
  shall print the lines of S13: `Slice:`, `Title:`, `Phase:`, `From:`
  (the slice's base branch and short commit), and `Created:` (UTC, to the minute).
  `Title:` shows the slice name when no title was saved. A value the log
  does not have prints as `-`.
  Check: e2e
- **R30.** Where `--json` is given, sliceflow shall print, instead of
  R29's lines, one JSON object with `slice`, `title` (only when one was
  saved), `phase`, `from`, `created` (RFC 3339 in UTC), and `warnings`
  (a list, empty when there are none). Each item of `warnings` is the
  warning's text only: without `Warning: `, and without `<name>: ` when
  the warning has a slice name. The list also holds the warning of R23,
  when there is one, e.g.
  `base branch main does not exist (the default); merged slices are not shown`. A value the log does not have
  is left out.
  Check: e2e
- **R31.** When no name is given and the current branch is
  `slice/<name>`, sliceflow shall show the status of `<name>`.
  Check: e2e
- **R32.** If no name is given and the current branch is not a
  `slice/` branch, then sliceflow shall print the three lines of S15,
  with `on <branch>` or `on no branch`, and exit with code 1.
  Check: e2e
- **R33.** If no slice with the given name exists, then sliceflow shall
  print `Error: no slice named <name>. See all slices: sliceflow slice list`
  and exit with code 1.
  Check: e2e

### List of slices

- **R34.** When the user runs `sliceflow slice list`, sliceflow shall
  print a header and one row per slice in progress, with the columns
  `NAME`, `TITLE`, `PHASE`, and `CREATED (UTC)`, oldest first. Columns
  are separated by 2 spaces, and each is as wide as its widest value.
  Slices with the same created time are in name order, a to z. Slices
  with no created time come last, by name. With no rows, it shall print
  `No slices in progress.` instead.
  Check: e2e
- **R35.** sliceflow shall leave merged slices out of the list, except
  where `--all` is given. With `--all` and no rows, it shall print
  `No slices.`
  Check: e2e
- **R36.** If a slice has warnings of its own, then sliceflow shall show
  ` !` after its phase in the list, and still print every other row. The
  warning of R23 adds no ` !`.
  Check: e2e
- **R37.** Where `--json` is given, sliceflow shall print, instead of the
  rows or the `No slices` lines, one JSON object with `slices`, a list of
  objects in the form of R30 in the order of R34, and `warnings`, a list
  of the warnings about the whole repo (the warning of R23), each item
  written as in R30. Each slice
  object holds only its own warnings. With no slices, `slices` is `[]`,
  e.g. `{"slices":[],"warnings":[]}`.
  Check: e2e

### Every command

- **R38.** sliceflow shall exit with code 0 on success, 1 on an error,
  and 2 on wrong use (an unknown command or flag, or a missing value).
  It shall print errors on stderr, and nothing on stdout when a command
  fails. For wrong use, the error is
  `Error: <what is wrong>. Run sliceflow --help.`
  Check: e2e
- **R39.** Where `--help` or `-h` is given, sliceflow shall print the
  usage of that command on stdout, and exit with code 0.
  Check: e2e
- **R40.** sliceflow shall make no network connection. It shall never run
  `git fetch`, `git pull`, or `git push`, and reads only what git already
  has.
  Check: e2e
- **R41.** sliceflow shall print every path with `/` between its parts,
  on every system, also in `--json`. On Windows a full path looks like
  `C:/Users/me/project/.sliceflow/worktrees/spec-submit`.
  Check: unit, e2e

## Out of scope

- User hooks before and after events. Goes to: the backlog
- Cancelling a slice, and the `slice.cancelled` event. Goes to: the backlog
- Who caused an event (a `by` field, detected from the agent tool). Goes to: the backlog
- A `status.md` written from the log. Goes to: the backlog
- A local server and browser app (dashboard, starting agents). Goes to: the backlog
- Teammates' slices from remote branches in `slice list`. Goes to: the backlog
- Every event other than `slice.created`. Goes to: a later slice (each one with the step it records)
- Adding an event to an existing log: the lock, and cutting off a half-written last line. Goes to: a later slice (the first one that adds a second event)
- A command to set or change a slice's title, and showing the title in `spec.md`, written from the log. Goes to: the backlog
- A settings file (`sliceflow.yaml`). Goes to: a later slice (when there are settings a whole team shares)
- Templates a project can change. Goes to: a later slice (design 1.10)
- Fetching from a remote. Goes to: not planned

## Decisions

- **SD1.** The user types the full slice name. sliceflow adds no number.
  Why: a number picked by counting slices clashes when several people
    work on the same repo.
  Not chosen: sliceflow picks the next number, because two clones can
    pick the same one.
  Rule: R1
- **SD2.** The event log is the only source of truth. This slice writes
  no `status.md`. `slice status` is the view.
  Why: one truth, and no copy that can go stale. A dashboard will read
    the same log later.
  Not chosen: a `status.md` written from the log (the design, 2026-10-18),
    because it is a copy that can drift, and the dashboard replaces it.
  Cost: the status is not readable on GitHub without running sliceflow.
  Rule: R7
- **SD3.** Each event has the fields `version`, `seq`, `event`, `at`,
  and `data`. The slice name is not in the line, and `seq` replaces a
  random id.
  Why: the folder already names the slice, so a name in each line would
    be a second copy. A number shows the order and a lost line.
  Not chosen: CloudEvents' `id` and `source`, because `seq` and the
    folder already make each event unique.
  Rule: R19
- **SD4.** The log is read from commits on the slice branch, else the
  base branch.
  Why: it works from any branch, also from `main`, and when the worktree
    is gone. An event counts once it is committed.
  Not chosen: reading the worktree's file, because it needs the
    worktree, and teammates' slices have none.
  Rule: R24
- **SD5.** A bad line is skipped with a warning, and never stops a
  command.
  Why: one bad line must not block the whole slice.
  Not chosen: refusing a damaged log, because a crash would then need a
    fix by hand, and nobody edits the log by hand.
  Cost: a lost line can make the phase shown wrong. The warnings, also
    in `--json`, make that visible.
  Rule: R26, R27, R28
- **SD6.** The base branch comes from `--from`, else `git config
  sliceflow.base`, else `main`.
  Why: a personal base branch is set once per clone, and needs no new
    file and no new dependency.
  Not chosen: a `sliceflow.yaml` setting, because it needs a YAML
    library, and a personal branch does not belong in a shared file.
  Rule: R3
- **SD7.** Worktrees live in `.sliceflow/worktrees/<name>/`, and
  `.sliceflow/` ignores itself with its own `.gitignore`.
  Why: one folder for all of sliceflow's local files, and the project's
    own `.gitignore` is not changed.
  Not chosen: `.worktrees/` with a line added to the project's
    `.gitignore`, because sliceflow would edit a project file.
  Rule: R5, R6
- **SD8.** The spec template is built into sliceflow.
  Why: it works in any repo with no setup.
  Not chosen: the project's own template first, because templates a
    project can change come later.
  Cost: a new template needs a new sliceflow release.
  Rule: R7
- **SD9.** A title is saved only when the user gives `--title`. Without
  it, the name is shown.
  Why: a title made from the name adds nothing, and would be a second
    copy of the name.
  Not chosen: a title made from the name, because it saves the same fact
    twice and can read badly, e.g. "Oauth2 api".
  Rule: R20, R29
- **SD10.** `slice list` shows only local `slice/` branches and the base
  branch, hides merged slices unless `--all`, and has no column for
  where a slice is.
  Why: everything in the list is the user's own or merged, so nothing
    is unclear. The list answers "what is going on now".
  Not chosen: listing remote branches too, because it is hard to tell
    whose slice it is.
  Rule: R34, R35
- **SD11.** `slice status` shows the overview only, not the events.
  Why: it is the one place a user or an AI agent reads where a slice
    stands.
  Not chosen: the overview plus the events, because the events are not
    needed to know the status.
  Rule: R29
- **SD12.** The title lives only in the event log. The spec's front
  matter has no `title`.
  Why: one source of truth. A title in the spec could be edited and
    then disagree with the log.
  Not chosen: the title only in the spec, because `status` would then
    depend on reading the spec.
  Cost: the spec file does not show its title, until it is written there
    from the log later.
  Rule: R7
- **SD13.** The commands only format what a shared core returns. A later
  server for the browser app calls the same core, not the commands.
  Why: the CLI and the server are two ways into one program, as in Theia.
  Not chosen: the server running `sliceflow slice status --json`, because
    it would run a new program for every request.
  Rule: none

## Assumptions

- The user has git installed, a version with `git worktree`.
  Noticed: `slice create` fails at the worktree step.
- The spec template text of this repo also fits other projects.
  Noticed: a user in another project finds parts that only make sense
    here.

## Open questions

None.
