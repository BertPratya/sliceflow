---
slice: 0002-spec-and-state
created: 2026-10-24
touches:
  - {{a project file this slice changes, e.g. api/openapi.yaml}}
---

<!-- HOW TO FILL THIS TEMPLATE

     A plan says HOW the spec will be built: the parts that
     change, the flows, the interfaces, the data, the tests, and the
     decisions. The tasks go in tasks.md, written after this plan is
     designed and its Open questions say None. The human approves the
     spec, plan, tasks, and preview together, in one approval.

     The plan never adds scope. Anything not in the spec needs a spec
     change first. If the spec leaves out something the user sees (an
     output layout, an exit code, a message), don't decide it here: add
     an Open question, so the spec is changed.

     Write plain English, short sentences, one idea per sentence. A
     point the lead decided alone ends with (lead's call); the human sees
     every one at approval.

     sliceflow reads this file. Keep every marker exactly as shown: the
     headings, the numbers (**TD1.**, **Q1.**), the labeled lines
     (Success criteria:, Covers:, Project rule:, Why:, Not chosen:, Cost:,
     Replaces:, Plan only:), and the front matter. Don't rename or reorder
     them. A labeled line may go on over more lines:
     indent the extra lines 2 spaces more than the label.

     - Replace every {{...}} with real text. The text inside says what
       goes there. The lead fills slice and created when it copies
       this template.
     - Each section has a comment like this one with its rules. Follow it.
     - The ## sections are fixed. Add ### subsections freely, except in
       Flows and Interfaces, which have their own ### rules.
     - Repeat a block as often as needed: flows, rules, decisions.
     - A required section with nothing to say gets one line: None.
     - Two sections are optional: Risks and rollout, and Parked. Delete
       them when they don't apply.
     - When you are done, delete every comment and check that no {{ is
       left. A plan with a comment or a {{ left in it is not finished.
       Every placeholder of this template counts, even inside code. A {{
       or a comment marker you write yourself inside a code block or
       `inline code` is real content.
     - To change a plan later, read this template again first. The
       finished plan no longer has these rules in it.

     The front matter:
     - A stacked slice (one that waits for another slice) adds a line:
       after: { slice: <that slice's name>, planning: <its approved revision> }
     - touches: the project-level files this slice changes, e.g. a
       contract or the data model doc. None? Write: touches: [] -->

# Plan

## Approach

<!-- The whole idea in a few sentences. A reader should understand the
     plan from this section alone.

     Then a table of the parts that change. A part is a module or
     package, e.g. internal/store. Change is one of: new, changed,
     removed.

     Then one line for new dependencies (libraries, tools, services).
     Each one needs the human's approval, so name it and say why it is
     needed. None? Write: New dependencies: None.

     Then, only if the spec has success criteria: a Success criteria:
     line, and under it one bullet per criterion. Each bullet starts with
     the criterion's number and a colon, then the flows, decisions, or
     shared rules that reach it, then a period and one or two sentences
     on how. sliceflow checks that every criterion has a bullet and that the
     names exist; the reasoning is checked in the plan review.
     No success criteria in the spec? Leave the line out.

     Example:
     The due date is a new optional field on a task. Sorting and the
     overdue mark happen when the list is printed.

     | Part | Folder | Change |
     |---|---|---|
     | Task store | internal/store/ | changed |
     | List command | cmd/todo/ | changed |

     New dependencies: None.

     Success criteria:
     - SC1: Flow 2, TD3. The run loop never calls the lead when a task
       passes, and sliceflow writes the merge messages itself. -->

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
`slice`; `slice` uses `state`, `spec`, `review`, `git`, and `template`;
`state` uses `eventlog`. Only `slice` and `git` touch files or run git.

New dependencies: None.

## Flows

<!-- One ### Flow subsection per main path, including the important
     failure paths. Each one starts with a Covers: line, then lists the
     steps. Add a Mermaid diagram for complex flows.

     Rules that apply across several flows (expiry, retries, locking,
     security) go under ### Shared rules, one #### subsection each, also
     with a Covers: line. Give each rule a short heading that no other
     shared rule or interface uses: tasks name it in follows.

     Shared rules also list each project rule this slice must handle
     (from AGENTS.md or the project's rules), and how the slice handles
     it. Such a rule has a Project rule: line quoting where it comes
     from, and Covers: names the requirements it touches, or none. List
     every project rule that applies: a missing one is a plan review
     finding.
     Tasks name these rules in follows like any other, so they reach the
     worker.

     - Number the flows Flow 1, Flow 2, ...
     - Covers: names spec requirements, e.g. Covers: R1, R5
     - Every requirement in the spec is covered by at least one flow or
       shared rule.

     Example:
     ### Flow 1: Add a task with a due date

     Covers: R1, R4

     1. The add command reads --due and checks the date.
     2. A bad date stops the command before anything is saved.
     3. The store saves the task with its due date.

     ### Shared rules

     #### Cross-platform paths

     Covers: none
     Project rule: AGENTS.md, paths must work on Windows, macOS, and
       Linux.
     Paths are built with filepath.Join, never with "/". The tests run
     on all three systems in CI. -->

### Flow 1: {{name}}

Covers: {{R1, R2}}

1. {{step}}

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
2. Check the move (`state.Allowed`). Not allowed: save nothing, and
   return the spec's error, e.g. R20's or R6's.
3. Delete a `.events.jsonl.tmp` left by a crash. Write the committed log,
   a `\n` when it does not end with one (SD15), and the new line, to
   `.events.jsonl.tmp`; then rename it over `.events.jsonl` (R9).
4. Write the event's other files, e.g. the review file.
5. `git add -f -- <the files>`, then
   `git commit --only -m "<message>" -- <the files>` in the slice's
   worktree (TD1). `-f` adds them also when an ignore rule matches, as in
   0001.

#### Undo after a failed commit

Covers: R10

When step 5 of Saving an event fails, or Ctrl-C comes before it ends:

1. Write the log back as it was committed (by the same temporary file
   and rename).
2. Delete the files step 4 created; put back the files it changed, as
   they were committed.
3. Unstage what step 5 staged: `git reset -q -- <the files>`.
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

## Interfaces

<!-- Every way the outside world talks to the system that this slice
     touches. One ### subsection per kind: CLI, HTTP API, Events,
     Settings, UI routes. Tasks name a subsection in follows by its
     heading, so each heading is used once.

     - If a contract file exists (e.g. OpenAPI), list the operations and
       link the file. Never copy request or response details.
     - If there is no contract file (e.g. a CLI or settings), describe it
       fully here: commands, flags, output, exit codes. What the user
       sees comes from the spec. If the spec doesn't give it, don't
       decide it here: add an Open question.
     - None? Write: None. -->

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
func Allowed(s State, e eventlog.Event) error          // R2: is this move allowed now?
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
	Slice, Doc, File, Commit string // File: the review file's path in the worktree
	Round                    int
	Worktree                 string
}
```

The reviewer writes the review file; sliceflow reads the result from
it (Review files). The placeholder writes R26's text.

## Data

<!-- What stored data changes: new or changed tables, fields, or local
     files. How existing data is moved over, if at all. Link the project
     data model doc for full details.
     None? Write: None. -->

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

<!-- How the checks named in the spec will run. The spec says which kind
     of check each requirement gets. This section says how those checks
     are done, so every task uses the same helpers.

     One row per check kind the spec uses (unit, api, e2e). Say which
     test tools, fakes, and test data are used, and where the test files
     live. Each task puts its e2e tests in its own file, so say the
     folder and the file name form, e.g. e2e/<task_name>_test.go (the
     task's name with - written as _, as Go file names usually are). manual
     checks are not listed here: they go in preview.md.

     Example:
     | Check | How |
     |---|---|
     | unit | Go tests; a fake clock for "today" |
     | e2e | build the binary, run it in a temporary folder; one file per task: e2e/<task_name>_test.go | -->

| Check | How |
|---|---|
| {{unit, api, or e2e}} | {{test tools, fakes, test data}} |

## Decisions

<!-- Technical decisions, numbered TD1, TD2, ...

     - Each one has the choice, Why:, and Not chosen: (the main other
       option, and why not).
     - Optional: a Cost: line for the downside we accept with this
       choice. Leave it out when there is no real downside.
     - To replace a decision from an earlier slice, add Replaces: with
       that slice's name and the decision's number, e.g. checkout/TD2.
       Otherwise remove the whole Replaces: line.
     - A decision that goes against a project rule says so in Why:.
     - Only technical decisions. Don't repeat the spec's decisions (SD1,
       ...); they are already approved.
     - A worker sees a decision only when its task names it in follows
       (in tasks.md). So every decision is named by at least one task.
       A decision that no task's code depends on says why, in a
       Plan only: line, e.g. Plan only: it picks a dependency; no task's
       code depends on the choice. sliceflow warns about a decision that is
       neither followed nor plan only, and about one that is both.

     Example:
     - **TD1.** Store the due date as text in the form YYYY-MM-DD.
       Why: it sorts correctly as text and needs no time zone.
       Not chosen: a timestamp, because a due date has no time of day. -->

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

<!-- Optional. Delete this section if not needed.
     What could go wrong, and how it is reduced. Migration order, feature
     flags, compatibility with older versions. -->

- **A slow git hook gives a false "busy".** The lock is held while
  `git commit` runs the project's hooks. A hook that takes 40 s makes a
  second command give up after 10 s, although nothing crashed. This is
  rare: one lead works on one slice. The message is clear, and running
  the command again works. The human chose to keep 10 s; if it happens
  in practice, a later slice can wait longer or show a waiting line.

## Parked

<!-- Optional. Delete this section if not needed.
     Ideas found while planning that belong to a later slice. Only new
     ideas: don't repeat an item the spec's Out of scope already sends
     to the backlog.

     Each idea is also added to docs/backlog.md on main at once,
     as plan-slice says. So:
     - One - bullet per idea. Start each idea with what it is.
     - Lines indented under a bullet belong to that idea.
     - No sub-bullets, headings, or code blocks.

     Example:
     - Show overdue tasks in red.
     - Let `todo due` take a number of days, e.g. `todo due 2 +3`.
       Saves typing the full date. -->

- {{An idea for later.}}

## Open questions

<!-- Questions for the human, numbered Q1, Q2, ... An answered question
     is written into the plan and removed from here. A deferred question
     that does not block the build moves to Parked and the backlog.
     The tasks are written only when this section says None. -->

- **Q1.** {{A question for the human.}}
