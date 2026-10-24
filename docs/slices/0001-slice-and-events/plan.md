---
slice: 0001-slice-and-events
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
| End-to-end tests: build the binary, run it in temporary repos | e2e/ | new |

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

#### {{rule name}}

Covers: {{R5, or none}}
Project rule: {{where it comes from, for a project rule; otherwise delete this line}}

{{the rule}}

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

New. The spec describes it fully; the code follows the spec's words.

- `sliceflow slice create <name> [--title <title>] [--from <branch>] [--json]`:
  R1 to R18, R42; output R9 and R10.
- `sliceflow slice status [<name>] [--json]`: R21 to R33.
- `sliceflow slice list [--all] [--json]`: R21 to R28, R34 to R37.
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
func Status(ctx context.Context, r Repo, name string) (Status, error) // name "" = the current branch
func List(ctx context.Context, r Repo, all bool) (List, error)

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

<!-- What stored data changes: new or changed tables, fields, or local
     files. How existing data is moved over, if at all. Link the project
     data model doc for full details.
     None? Write: None. -->

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

- A line of only spaces, tabs, or `\r` is a `blank line`. A good line
  that ends in `\r\n` reads normally, because JSON counts `\r` as a
  space. (lead's call)
- A last line with no `\n` at the end is read normally, with no
  warning. (lead's call)
- `version` and `seq` must be JSON whole numbers: `1.0` or `"1"` is
  `has the wrong type`. (lead's call)
- The whole log is read from git in one piece, so a line has no length
  limit. (lead's call)

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
  Plan only: it picks the Go version; no task's code depends on it.
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

<!-- Optional. Delete this section if not needed.
     What could go wrong, and how it is reduced. Migration order, feature
     flags, compatibility with older versions. -->

{{A risk, and how it is reduced.}}

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
