---
slice: spec
created: 2026-11-23
touches:
  - templates/default/spec.md
---

# Plan

## Approach

0002 adds three things on top of plan-on-base: a state worked out by
replaying a slice's log, `slice edit`, and the `spec` commands
(`template` and `validate`). The state machine and the spec reader are
new packages, one job each. `slice edit` reuses the pieces `slice
create` already has: the branch check, the lock, the commit of only
some files, and the undo of steps.

The build starts by splitting `internal/cli` into one package per
command group (backlog, 2026-11-23), before the first `spec` command
is written. Then it goes from the core up: the state machine, then
`status` with the state, then the commands that use it.

| Part | Folder | Change |
|---|---|---|
| Parsing, routing, help, shared printing | internal/cli/ | changed |
| Shared command types and printing helpers | internal/cli/command/ | new |
| The slice commands | internal/cli/slicecmd/ | new (moved from internal/cli/) |
| The spec commands | internal/cli/speccmd/ | new |
| The state machine, and the current title and description | internal/state/ | new |
| Edit, and status and list with the state | internal/slice/ | changed |
| Reading a spec and finding its problems | internal/spec/ | new |
| The template, without a placeholder in code | templates/default/spec.md | changed |
| End-to-end tests | e2e/ | changed |

New dependencies: None.

### How the pieces connect

Calls only go down.

```
internal/cli            Run, parsing, routing, help
   |-- internal/cli/slicecmd   create, edit, list, status
   |-- internal/cli/speccmd    template, validate
   '-- both use internal/cli/command  (Input, the command types, printing helpers)

internal/cli/slicecmd --> internal/slice --> internal/state, internal/eventlog, internal/git
internal/cli/speccmd  --> internal/spec  (reads a file, knows nothing of git)
                      --> internal/slice (only to find the slice's log and spec)
```

### Build order

1. Split `internal/cli` (no change in behavior; all tests pass before
   and after).
2. The state machine (`internal/state`).
3. `slice status` with the state, the next step, and the current title
   and description.
4. `spec template`.
5. The spec reader: lines, sections, comments, code, backticks.
6. The checks of `spec validate`, and the command.
7. `slice edit`.
8. `slice list` with the current title, and the help pages.

## Flows

### Flow 1: slice status

1. Check the name and find the user's checkout, as in plan-on-base.
2. Read the log. No log gives the no-slice error.
3. Read the events. A damaged first line is an error
   (slice-and-events/R55).
4. Replay the events into a state (`internal/state`): the position,
   and the current title and description, in one pass.
5. Print the warnings in their order (SD37): damaged lines, then the
   state warning.
6. Print the lines; for a known state, the next step.

### Flow 2: spec validate

1. Check the name, find the checkout, and check that the log exists.
   The log is not read (`validate-errors-as-before`).
2. Read `docs/slices/<name>/spec.md` as it is on the disk. No file
   gives the no-spec error.
3. The reader turns the text into lines, each marked as plain text,
   inside a code block, or inside a comment, and finds the code spans
   of each plain line (`spec-lines-read`, SD35, SD25).
4. The reader builds the sections, stories, rules, examples,
   decisions, and questions from the plain lines.
5. Each check returns its problems. The problems are sorted by line,
   then by the order of their checks in Story 3 (SD39), and a line with
   a placeholder keeps only that problem (SD41).
6. Print the problems and the count, or "no problems found".

### Flow 3: slice edit

1. Check, in the order of `edit-errors-as-before`: control
   characters, an option, the name, the project, the log, its first
   line, the base branch, the state.
2. Work out the current values, and the values asked for (SD3). The
   same values print "Nothing to change" and stop.
3. Steps, with undo:
   - append `slice.edited` under the lock (undo: cut the log back to
     its size before);
   - commit only the log (undo: unstage the log).
4. When the commit fails, check whether it was made anyway, as
   `create` does. If it was, nothing is removed.
5. Print the new title and description.

### Shared rules

#### Paths and line ends on every system

Project rule: AGENTS.md, "Paths, line endings, and processes work on
  Windows, macOS, and Linux."
Paths are built with `filepath.Join`. The spec reader accepts `\n` and
`\r\n`, and skips a byte order mark (`spec-lines-read`). Line numbers
are counted the same way for both line ends.

#### Each git command is its own process

Project rule: AGENTS.md, "Run each git command or check as its own
  process with its own arguments, never as one shell string."
`slice edit` uses the existing functions of `internal/git`: `Commit`,
`Unstage`, `PathExists`, `CheckedOutBranch`, `GitPath`.

#### The tool's own message

Project rule: AGENTS.md, "show the tool's own message word for word,
  after one line of sliceflow's own."
A failed commit in `slice edit` is printed with git's text, as in
`create`.

#### No data out

Project rule: AGENTS.md, "sliceflow sends no data out."
Nothing in this slice uses the network. `spec validate` reads one file.

#### Tests make their own repos

Project rule: AGENTS.md, "Tests create their own temporary git repos."
The e2e tests use the existing helpers in `e2e/`.

## Interfaces

### CLI

As the spec says, and nothing more:
- `sliceflow slice status <name>`: gains the State: line and the next
  step (Story 1).
- `sliceflow slice edit <name> [--title <text>] [--description <text>]`
  (Story 4).
- `sliceflow slice list`: shows the current title and description.
- `sliceflow spec template` (Story 2).
- `sliceflow spec validate <name>` (Story 3).
- The help of `sliceflow`, `slice`, `slice edit`, `slice status`,
  `spec`, `spec template`, and `spec validate`.

### Events

A new event, `slice.edited`. Its data holds `title`, `description`, or
both, as strings, and no other key (`edited-event`).

## Data

The event log gets a new kind of line, `slice.edited`. Logs written by
0001 and plan-on-base hold only `slice.created`, so they need no
change. An older sliceflow that reads a `slice.edited` line does not
know it; that is the case of `unknown-event`, for a newer log read by
an older version.

## Testing

| Check | How |
|---|---|
| unit | Go tests in each package. The state machine is tested with lists of events, with no files. The spec reader and the checks are tested with spec text in the test, with no files and no git. |
| e2e | The built binary, run in a temporary git repo with the existing helpers. One file per task: `e2e/<task_name>_test.go`. Output is compared word for word with the spec's examples. |

### Trying it by hand

#### Setup

1. Run `go build -o /tmp/sliceflow ./cmd/sliceflow`.
2. In a new folder, run `git init`, make one commit, then
   `/tmp/sliceflow slice create login`.
   Expect: "Created slice login on branch main."

#### Manual checks

None

## Decisions

- **TD1.** `internal/cli` is split into `internal/cli` (Run, parsing,
  routing, help), `internal/cli/command` (the command types, `Input`,
  and the printing helpers every group uses), and one package per
  group: `slicecmd`, `speccmd`.
  Why: the backlog plans the split for the start of 0002, while only
  three slice files have to move. A group package cannot import
  `internal/cli`, because `internal/cli` imports every group to route
  to it; the shared part goes in `command`, which both import.
  Not chosen: one `cli` package for every group, because it grows to
  some 60 files by 0006 (backlog row, 2026-11-23); putting the shared
  part in `internal/cli` itself, because Go refuses the import cycle.
  Cost: the types and helpers that groups share become exported, e.g.
  `command.PrintFailure`.

- **TD2.** The state machine is its own package, `internal/state`, that
  knows only events and states. It reads no file and runs no git. The
  current title and description are worked out in the same package, in
  the same pass over the events, as `State.Detail`.
  Why: the state machine is the core that 0003 to 0005 grow, and a
  package with no files or git is tested with plain lists of events.
  The title is not a state (SD36), but it is replayed from the same
  events in the same loop, so one `Replay` gives both, and `status`,
  `list`, and `edit` all get them from one call.
  Not chosen: the state inside `internal/slice`, because that package
  already does files and git, so its tests need repos; the title in
  `internal/slice`, because it would need a second loop over the same
  events.
  Cost: one more package.

- **TD3.** The state machine is a table of moves. `Replay(events)`
  returns the whole `State`: the `Position` (phase and step), the
  `Detail` (title and description), and `Unknown`. Each row of
  `transitions` has `From`, `Event`, `To`, and an optional `Guard`;
  `allowedEverywhere` lists the events allowed in every position that do
  not move the slice (`slice.edited`). An event is known when a row or
  `allowedEverywhere` names it. Replay reads every event: the detail
  follows each one, and the position freezes at the first event that is
  unknown or not allowed, where `Unknown` records its line, its name,
  and the reason. `eventlog` decodes each event's data once, into
  `Event.Payload`, and records `Event.Line`, so `state` never reads
  JSON.
  Why: a new event or step is one new row (roadmap rule 4). Returning
  the whole state is simple, and the state is small. Freezing only the
  position keeps SD9 and SD36 apart: the title still follows edits after
  the state became unknown.
  Not chosen: a switch per event, because each new event would touch
  the code, not the data; choosing which fields to replay, because the
  state is small and the saving cannot be measured (0.84 ms for 100
  events); stopping the loop at the first problem, because the title
  would stop too.
  Cost: a guard is a function, so the table cannot be printed or saved
  as data.

- **TD4.** The spec reader works in two passes: first it marks each
  line as plain, in a code block, or in a comment, and finds its code
  spans; then it builds the sections and items from the plain lines.
  Each check is a function that gets the built spec and returns its
  problems; one list holds the checks in the order of Story 3.
  Why: the hard part, what is code and what is a comment (SD25, SD35),
  is done once, so no check can get it wrong. The order of the list is
  the order of problems on one line (SD39).
  Not chosen: each check reading the raw text, because every check
  would then need its own rule for code and comments.
  Cost: the first pass must be right for every check to be right; it
  gets most of the unit tests.

- **TD5.** `slice edit`'s undo cuts the log back to the size it had
  before the event was added, then unstages the log. It reuses
  `runSteps` and the "was the commit made" check of `create`.
  `CommittedError` gets the word for what was done ("created" or
  "edited"), so one message form serves both commands.
  Why: the log is only ever added to, so its old size is exactly where
  the new line starts. Reusing the steps keeps one undo pattern.
  Not chosen: restoring the log from git (`git restore`), because it
  would also throw away a line another process added under the lock in
  the meantime.
  Cost: when another process appends after `edit` and before its
  commit fails, cutting back removes that line too. The lock makes this
  rare: `edit` holds it only while it appends.

## Risks and rollout

- The spec reader is the largest new piece, and a wrong first pass
  breaks every check. It gets the most unit tests, including this
  spec's own text, which must pass with no problems.
- Splitting `internal/cli` moves code under every command. It is done
  first and alone, with no change in behavior, so the existing tests
  catch any mistake.
- Logs written by this version can hold `slice.edited`, which an older
  sliceflow does not know. It shows the state as unknown with a
  warning (`unknown-event`), and changes nothing.

## Open questions

None
