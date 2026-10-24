---
slice: 0002-spec-and-state
---

# Tasks

## Task 01: eventlog-events

```yaml
covers: [R5]
depends_on: []
touches: [internal/eventlog/]
follows: [Event log, Data, Tests, Testing]
risk: core
```

### Goal

`eventlog` reads and writes the four new events, and keeps each event's
line number. The hardest part is `parseLine`: today it knows only
`slice.created`, and it must check each event's own `data` fields with
R26's reasons of 0001-slice-and-events, in the same order of checks.

### Steps

1. Write the tests listed under Done when. Run them, and see them fail.
2. `internal/eventlog/event.go`: the data types below, `New`, and the
   `Line` field on `Event`.
3. `internal/eventlog/parse.go`: one field list per event; `parseLine`
   checks unknown, missing, wrong-type, and wrong-value fields per event
   (R5), as it does for `slice.created` now. `Parse` sets `Line`.
4. Run the tests until they pass.

### Files

- `internal/eventlog/event.go`: changed. The new data types, `New`, and
  `Event.Line`.
- `internal/eventlog/parse.go`: changed. The new events in `parseLine`;
  `Parse` sets `Line`.
- `internal/eventlog/event_test.go`, `internal/eventlog/parse_test.go`:
  changed. The new tests.

### Interfaces

- `Event.Line int` with tag `json:"-"`: creates. The line number in the
  log, set by `Parse`; never written. Task 02 uses it for R4's warnings.
- `type Titled struct{ Title string }`: creates. JSON `title`.
- `type Submitted struct{ Result string; Errors, Warnings int }`:
  creates. JSON `result`, `errors`, `warnings`.
- `type ReviewStarted struct{ Doc string; Round int; Reviewer, File, Commit string }`:
  creates. JSON `doc`, `round`, `reviewer`, `file`, `commit`, in that
  order.
- `type ReviewFinished struct{ Doc string; Round int; Result string }`:
  creates. JSON `doc`, `round`, `result`.
- `func New(seq int, at time.Time, event string, data any) Event`:
  creates. `version` 1, `at` in `TimeLayout`. Tasks 11, 12, and 14 use
  it.
- `Parse`, `Encode`, `Phase`: unchanged here. Task 06 removes `Phase`.

### Done when

- `TestEncodeNewEvents`: `Encode` of each new event gives exactly the
  JSON lines of S13, S14, and S25, e.g.
  `{"version":1,"seq":2,"event":"slice.titled","at":"2026-10-24T07:50:00Z","data":{"title":"Sign in with Google"}}`.
  Proves: R5
- `TestParseNewEvents`: those lines parse back to the same values, with
  `Line` 1, 2, ... in log order.
  Proves: R5
- `TestParseNewEventsBad`: one bad line per new rule gives the reason of
  R26 of 0001-slice-and-events: `slice.titled` with `"title":""` gives
  `field "data.title" has a wrong value`; `spec.submitted` with
  `"result":"maybe"`, or `"errors":-1`, a wrong value; `review.started`
  with `"round":0`, `"commit":"abc"`, or `"doc":"plan.md"`, a wrong value;
  `review.finished` with `"result":"ok"`, a wrong value; a missing
  `data.round` gives `missing field "data.round"`; an extra `data.x`
  gives `unknown field "data.x"`.
  Proves: R5

### Out of scope

- What the events mean for the state. Done in: task 02
- Writing the log to disk. Done in: task 11

## Task 02: state-machine

```yaml
covers: [R1, R2, R3, R4]
depends_on: ["01"]
touches: [internal/state/]
follows: ["Flow 1: Replay a slice, and show its status", State machine, Tests, Testing]
risk: core
```

### Goal

A new package `internal/state` replays events into a `State` through
the table of moves (R2), and says whether one move is allowed now. It
touches no files and no git. The table is one list in this package, so
later slices add rows to it.

### Steps

1. Write the tests listed under Done when. Run them, and see them fail.
2. `internal/state/state.go`: the types below.
3. `internal/state/moves.go`: the table of R2, `Allowed`, and the
   change each move makes.
4. `internal/state/replay.go`: `Replay`, which applies each allowed
   event in order, and adds R4's warning for each other one.
5. Run the tests until they pass.

### Files

- `internal/state/state.go`: new. The types.
- `internal/state/moves.go`: new. The table, `Allowed`.
- `internal/state/replay.go`: new. `Replay`.
- `internal/state/state_test.go`: new.

### Interfaces

- `type State struct{ Phase, Title string; Spec Doc }`: creates.
  `Phase` is `planning` or `unknown` (R3); `Title` is the newest title,
  `""` when none was saved.
- `type Doc struct{ Status string; Review *Review; Refused *Counts; MaxRound, SubmitSeq int }`:
  creates. `Status` is `""` before `slice.created`, then `draft`,
  `submitted`, `in_review`, or `review_passed`. `Review` is the last
  started review. `Refused` is set while the last spec event is a
  refused `spec.submitted`. `MaxRound` is the highest round started.
  `SubmitSeq` is the `seq` of the last accepted `spec.submitted`.
- `type Review struct{ Round int; Reviewer, File, Commit, Result string }`:
  creates. `Result` is `""` while the review runs.
- `type Counts struct{ Errors, Warnings int }`: creates.
- `func Replay(events []eventlog.Event) (State, []string)`: creates.
  The strings are R4's warnings, e.g.
  `.events.jsonl line 4: review.finished is not allowed when spec.md is draft; skipped`.
  Task 06 uses it; task 11 replays with it before each write.
- `func Allowed(s State, e eventlog.Event) string`: creates. `""` when
  allowed, else R4's reason without `.events.jsonl line <n>: ` and
  `; skipped`, e.g. `review.finished is not allowed when spec.md is draft`.
  Task 11 uses it.

### Done when

- `TestMoves`: one case per row of R2's table, from each state: the
  allowed ones change `Spec.Status` as the table says; `slice.titled`
  sets `Title` in every state; `spec.submitted` `refused` keeps the
  state and sets `Refused` to its counts; an accepted submit or a
  `review.started` clears `Refused`.
  Proves: R2
- `TestReplayCreated`: a log of only `slice.created` gives phase
  `planning` and `draft`; an empty list gives phase `unknown` and
  `Status` `""`.
  Proves: R1, R3
- `TestReplayNotAllowed`: `review.finished` at line 4 with no review
  started is skipped, with the warning of S3, and the state stays
  `draft`; a `slice.titled` before `slice.created` gives
  `.events.jsonl line 1: slice.titled comes before slice.created; skipped`;
  a second `slice.created` at line 2 gives
  `.events.jsonl line 2: slice.created is not allowed when spec.md is draft; skipped`;
  a `review.finished` with
  round 2 while review 1 runs gives
  `.events.jsonl line 5: review.finished is for round 2, but review 1 is running; skipped`.
  Proves: R4
- `TestReplayRounds`: two passed reviews give `MaxRound` 2 and `Review`
  round 2; a skipped `review.started` does not count.
  Proves: R2

### Out of scope

- Reading the log from git, and the status texts. Done in: task 06
- Saving events. Done in: task 11

## Task 03: spec-reader

```yaml
covers: [R18]
depends_on: []
touches: [internal/spec/]
follows: [Spec reader, Cross-platform paths, Tests, Testing]
risk: core
```

### Goal

A new package `internal/spec` reads the text of a spec into a `Doc`:
the front matter, the `##` sections, the marked items with their
labeled lines, the `{{` placeholders, and the template comments, each
with its line number. The hardest part is what is skipped: fenced code
blocks, `inline code`, and comments (R18).

### Steps

1. Write the tests listed under Done when. Run them, and see them fail.
2. `internal/spec/doc.go`: the types below.
3. `internal/spec/parse.go`: `Parse`.
4. Run the tests until they pass.

### Files

- `internal/spec/doc.go`: new. The types.
- `internal/spec/parse.go`: new. `Parse`.
- `internal/spec/parse_test.go`: new.

### Interfaces

- `func Parse(text []byte) Doc`: creates. Never fails: what it cannot
  read, it records for the checks. Task 04 uses it.
- `type Doc struct{ HasFront bool; Front []Field; Sections []Section; Placeholders []Mark; Comments []Mark; LastLine int }`:
  creates.
- `type Field struct{ Key, Value string; Line int; ListItems []string }`:
  creates. A front matter key; `ListItems` holds the `- ` lines under
  it.
- `type Section struct{ Name string; Line int; Items []Item; BadMarkers []Mark; Empty bool; Text string }`:
  creates. `Name` without `## `. `Text` is the section's text outside
  code and comments, for checks like `None.`.
- `type Item struct{ ID string; Line int; Removed bool; Labels []Label; Tail string }`:
  creates. `ID` like `R7`; `Tail` is the item's first line after the
  marker, e.g. for `Goes to:` on an Out of scope item.
- `type Label struct{ Name, Value string; Line int }`: creates. `Name`
  like `Check`; `Value` with its continued lines joined by spaces.
- `type Mark struct{ Line int; Text string }`: creates.

### Done when

- `TestParseSections`: a small spec gives its sections in order, with
  their lines, and its items with IDs, lines, and labels (a `Check:`
  line, and a `Why:` that goes on over 2 lines).
  Proves: R18
- `TestParseSkips`: in a fenced code block, `**R9.**`, `{{slice}}`, and
  `<!--` are not read; in `inline code`, `{{x}}` and `<!--` are not; in
  a comment, `**S1.**` and `{{y}}` are not, and the comment is one
  `Comments` mark at its first line.
  Proves: R18
- `TestParseCRLF`: the same spec with `\r\n` line ends gives the same
  `Doc`.
  Proves: R18
- `TestParseRemoved`: `- **R4.** (removed in R002: why)` gives an item
  with `Removed` true.
  Proves: R18
- `TestParseFrontMatter`: `ideas: []`, `ideas: [docs/ideas/a.md]`, and
  `ideas:` followed by `  - docs/ideas/a.md` are read; a file with no
  first `---` line gives `HasFront` false.
  Proves: R18

### Out of scope

- The checks and their messages. Done in: task 04, task 05

## Task 04: spec-errors

```yaml
covers: [R15, R18]
depends_on: ["03"]
touches: [internal/spec/check.go, internal/spec/errors.go, internal/spec/errors_test.go, internal/spec/testdata/]
follows: [Spec reader, Tests, Testing]
risk: core
```

### Goal

`spec.Check` returns every error of R15's table, with R15's messages and
lines. A value that holds a `{{` is not checked (R18). Two real files
are the test data: a fresh spec from the template, and 0001's approved
spec; both give 0 errors.

### Steps

1. Write the tests listed under Done when. Run them, and see them fail.
2. `internal/spec/check.go`: `Problem` and `Check`, which sorts the
   problems by line, then errors before warnings before notes.
3. `internal/spec/errors.go`: one function per group of R15's rows.
4. `internal/spec/testdata/`: `fresh.md`, written by
   `template.Spec("login-google", <2026-10-24>)`, and `0001-spec.md`, a
   copy of `docs/slices/0001-slice-and-events/spec.md`.
5. Run the tests until they pass.

### Files

- `internal/spec/check.go`: new. `Problem`, `Check`.
- `internal/spec/errors.go`: new. R15's checks.
- `internal/spec/errors_test.go`: new.
- `internal/spec/testdata/fresh.md`, `internal/spec/testdata/0001-spec.md`:
  new.

### Interfaces

- `type Problem struct{ Line int; Level, Message string }`: creates.
  `Level` is `error`, `warning`, or `note`. Tasks 05, 09, and 14 use it.
- `func Check(d Doc, slice string) []Problem`: creates. Task 05 adds the
  warnings and notes; task 09 calls it.
- `Parse`, `Doc`: uses.

### Done when

- `TestErrorRows`: one small spec per row of R15's table gives exactly
  that row's message at that row's line, e.g. a requirement R7 at line
  42 with no `Check:` line gives `{42, "error", "R7 has no Check: line"}`;
  `Shows: R99` gives `S3 shows R99, which does not exist`; a `Rule:`
  naming a removed R4 gives `SD2: Rule: names R4, which was removed`.
  Proves: R15
- `TestErrorPlaceholderValue`: `Check: {{unit, api, e2e, or manual}}`,
  `Shows: {{R3}}`, `Rule: {{R11}}`, `Goes to: {{a later slice}}`, and
  the front matter `created: {{date}}` give no error.
  Proves: R18
- `TestErrorsFresh`: `testdata/fresh.md` gives 0 errors.
  Proves: R15, R18
- `TestErrors0001`: `testdata/0001-spec.md`, checked as slice
  `0001-slice-and-events`, gives 0 errors.
  Proves: R15

### Out of scope

- Warnings and notes. Done in: task 05
- Reading the file from disk, and printing. Done in: task 09

## Task 05: spec-warnings

```yaml
covers: [R16, R17, R18]
depends_on: ["04"]
touches: [internal/spec/check.go, internal/spec/warnings.go, internal/spec/warnings_test.go]
follows: [Spec reader, Tests, Testing]
risk: core
```

### Goal

`spec.Check` also returns R16's warnings and R17's notes. A fresh spec
then gives warnings for its placeholders and comments; 0001's spec gives
the notes for R11, R40, and R43.

### Steps

1. Write the tests listed under Done when. Run them, and see them fail.
2. `internal/spec/warnings.go`: R16's checks, and R17's note.
3. `internal/spec/check.go`: `Check` adds them.
4. Run the tests until they pass.

### Files

- `internal/spec/warnings.go`: new.
- `internal/spec/check.go`: changed. `Check` adds the warnings and notes.
- `internal/spec/warnings_test.go`: new.

### Interfaces

- `Check`, `Problem`, `Doc`: uses; `Check` changes to add them.

### Done when

- `TestWarningRows`: one small spec per row of R16's table gives that
  row's message, e.g. Open questions with Q1 and Q2 at line 140 gives
  `{140, "warning", "2 open questions left (Q1, Q2)"}`; one gives
  `1 open question left (Q1)`; `{{short title}}` gives
  `placeholder left: {{short title}}`; a `{{` with no `}}` gives the text
  to the end of the line; an empty `## Why` gives
  `"## Why" is empty; write None. if there is nothing to say`.
  Proves: R16
- `TestNoteNoScenario`: R8 that no `Shows:` names gives
  `{44, "note", "no scenario shows R8"}`; a removed R4 gives no note.
  Proves: R17, R18
- `TestWarnings0001`: `testdata/0001-spec.md` gives 0 warnings and
  exactly the notes for R11, R40, and R43.
  Proves: R16, R17
- `TestCheckOrder`: problems at lines 42, 44, and 140 of S5 come in line
  order; at one line, the error comes before the warning.
  Proves: R16

### Out of scope

- Printing, the summary line, and the exit code. Done in: task 09

## Task 06: status-core

```yaml
covers: [R1, R3, R29, R30, R34]
depends_on: ["02"]
touches: [internal/slice/status.go, internal/slice/list.go, internal/slice/worktree.go, internal/slice/next.go, internal/slice/status_test.go, internal/slice/list_test.go, internal/slice/worktree_test.go, internal/slice/next_test.go, internal/eventlog/parse.go, internal/eventlog/parse_test.go]
follows: ["Flow 1: Replay a slice, and show its status", Core API, Git arguments, Cross-platform paths, Tests, Testing]
risk: core
```

### Goal

`slice.Status` and `slice.List` work out the state with `state.Replay`,
and give the `Spec:` and `Next:` texts of R29. For a passed spec, they
check whether `spec.md` changed after its review (R30). The hardest
part is R29's refused mark, which depends on both the state and R30.

### Steps

1. Write the tests listed under Done when. Run them, and see them fail.
2. `internal/slice/worktree.go`: `findWorktree`.
3. `internal/slice/next.go`: `specChanged` and `specNext`.
4. `internal/slice/status.go`: `readSlice` uses `state.Replay`, adds its
   warnings, and fills the new fields; `currentSliceName` takes the
   command for R32's second line.
5. `internal/eventlog/parse.go`: remove `Phase`, now unused.
6. Run the tests until they pass.

### Files

- `internal/slice/worktree.go`: new. `findWorktree`.
- `internal/slice/next.go`: new. `specChanged`, `specNext`.
- `internal/slice/status.go`: changed.
- `internal/slice/list.go`: changed, if needed for the new fields.
- `internal/eventlog/parse.go`, `internal/eventlog/parse_test.go`:
  changed. `Phase` and its test removed.
- `internal/slice/status_test.go`, `internal/slice/list_test.go`:
  changed. `internal/slice/worktree_test.go`, `internal/slice/next_test.go`:
  new.

### Interfaces

- `SliceStatus` gains `State state.State`, `SpecText, Next string`, and
  `ChangedAfterReview bool`: changes. The fields `Phase` and `Title`
  stay, filled from `State`, so `internal/cli` keeps working. Task 07
  prints the new ones.
- `func findWorktree(ctx context.Context, r Repo, name string) (string, bool, error)`:
  creates. The full path of the worktree whose branch is
  `refs/heads/slice/<name>`, from `git worktree list --porcelain -z`.
  Tasks 09, 11, 12, and 14 use it.
- `func specChanged(ctx context.Context, worktree, name, commit string) (bool, error)`:
  creates. Flow 1, step 3. Task 14 uses it for R20.
- `func specNext(st state.State, changed bool) (spec, next string)`:
  creates. R29's table and refused mark.
- `func currentSliceName(ctx context.Context, r Repo, command string) (string, error)`:
  changes. `command` like `slice status` or `spec validate` goes into
  R32's second line. Task 09 uses it.
- `state.Replay`, `git.Run`, `git.Show`: uses.

### Done when

- `TestSpecNext`: one case per row of R29's table, e.g. `draft` after a
  refused submit with 0 errors and 1 warning gives
  `draft (submit refused: 1 warning)` and
  `fix spec.md, then: sliceflow spec submit`; with 2 errors and 1
  warning, `(submit refused: 2 errors, 1 warning)`; changed after review
  1 gives `passed review 1, but spec.md was changed after it` and
  `sliceflow spec submit`.
  Proves: R29
- `TestStatusState`: a slice whose committed log has `slice.created` and
  `slice.titled` "Sign in with Google" gives phase `planning`, that
  title, and `draft`; the log's lines are read from the commit, not from
  the file on disk.
  Proves: R1, R3, R34
- `TestSpecChanged`: in a temporary repo, a commit with `spec.md`, then
  an edit on disk, gives true; no edit gives false; with
  `core.autocrlf true`, a file that differs only by `\r\n` gives false.
  Proves: R30
- `TestFindWorktree`: finds the worktree of `slice/login-google`; a slice
  whose worktree was removed gives false.
  Proves: R30

### Out of scope

- Printing the lines, and the JSON. Done in: task 07

## Task 07: status-spec

```yaml
covers: [R4, R29, R31]
depends_on: ["06"]
touches: [internal/cli/status.go, internal/cli/list.go, internal/e2e/status_command_test.go, internal/e2e/slice_list_test.go, internal/e2e/status_spec_test.go]
follows: [CLI, Core API, Cross-platform paths, Tests, Testing]
risk: normal
```

### Goal

`slice status` prints the `Spec:` and `Next:` lines, and `--json` adds
`docs` and `next`, also in `slice list --json`. 0001's e2e tests are
updated for the two new lines.

### Steps

1. Write the tests listed under Done when. Run them, and see them fail.
2. `internal/cli/status.go`: the two lines after `Created:`, and `docs`
   and `next` in the JSON (R31).
3. `internal/cli/list.go`: the same two JSON fields per slice.
4. Update 0001's status and list e2e tests for the new lines and fields.
5. Run the tests until they pass.

### Files

- `internal/cli/status.go`, `internal/cli/list.go`: changed.
- `internal/e2e/status_command_test.go`, `internal/e2e/slice_list_test.go`:
  changed. 0001's tests, with the new lines.
- `internal/e2e/status_spec_test.go`: new.

### Interfaces

- `SliceStatus.State`, `SpecText`, `Next`, `ChangedAfterReview`: uses.

### Done when

- `TestStatusSpecDraft`: S1's slice prints S1's 7 lines exactly.
  Proves: R29
  Scenario: S1
- `TestStatusSpecJSON`: S2's command prints S2's JSON, with
  `"docs":[{"doc":"spec.md","status":"draft"}]` and `next`.
  Proves: R31
  Scenario: S2
- `TestStatusNotAllowedLine`: S3's log prints `Spec:     draft` and S3's
  warning on stderr, exit code 0.
  Proves: R4
  Scenario: S3
- `TestStatusReviewChanges`: S4's log prints S4's two lines.
  Proves: R29
  Scenario: S4
- `TestStatusJSONReview`: a log with review 1 `passed` gives
  `"review":{"round":1,"reviewer":"placeholder","file":"reviews/spec-1.md","result":"passed"}`;
  a refused submit gives `"submit_refused":{"errors":0,"warnings":1}`;
  a running review has no `result`; a passed review with `spec.md`
  edited on disk gives `"changed_after_review":true`; no `slice.created`
  gives `"docs":[]` and no `next`; `slice list --json` has the same
  `docs` and `next` for each slice.
  Proves: R31

### Out of scope

- The status after a real submit (S13 to S17, S23). Done in: task 15

## Task 08: cli-groups-and-template

```yaml
covers: [R35, R36]
depends_on: []
touches: [internal/cli/cli.go, internal/cli/cli_test.go, internal/cli/template.go, internal/template/template.go, internal/template/template_test.go, internal/e2e/cli_skeleton_test.go, internal/e2e/cli_groups_and_template_test.go]
follows: ["Flow 7: Show the template", "Flow 8: Wrong use, and the order of the first checks", CLI, Tests, Testing]
risk: core
```

### Goal

The CLI knows the command groups `slice`, `spec`, and `template`, and
its usage lists every new command. `sliceflow template show spec` prints
the built-in template as it is embedded.

### Steps

1. Write the tests listed under Done when. Run them, and see them fail.
2. `internal/cli/cli.go`: groups in `Run`; `topUsage` lists all seven
   commands; `sliceflow spec` and `sliceflow template` alone print it.
3. `internal/template/template.go`: `Raw`.
4. `internal/cli/template.go`: `template show`, with
   `missing template name`, `too many arguments`, and R35's error.
5. Run the tests until they pass.

### Files

- `internal/cli/cli.go`, `internal/cli/cli_test.go`: changed.
- `internal/cli/template.go`: new.
- `internal/template/template.go`, `internal/template/template_test.go`:
  changed.
- `internal/e2e/cli_skeleton_test.go`: changed, for the new usage text.
- `internal/e2e/cli_groups_and_template_test.go`: new.

### Interfaces

- `func Raw() []byte`: creates, in `internal/template`. The embedded
  spec template, unchanged.
- `topUsage`: changes. Lines for `spec validate [<name>] [--strict]`,
  `spec submit [<name>]`, `slice title <name> <title>`, and
  `template show <template>`. Tasks 09, 12, and 15 register those
  commands with `register`.

### Done when

- `TestTemplateShow`: S29's command prints exactly `template.Raw()` on
  stdout, starting with S29's lines, exit code 0, in a folder that is
  not a git repo.
  Proves: R35
  Scenario: S29
- `TestTemplateShowUnknown`: S30's command prints S30's error, exit 1.
  Proves: R35
  Scenario: S30
- `TestGroupsAlone`: `sliceflow spec` and `sliceflow template` print the
  usage on stdout, exit 0; `sliceflow template show` prints
  `Error: missing template name. Run sliceflow --help.`, exit 2;
  `sliceflow template show spec --json` gives `unknown flag --json`,
  exit 2.
  Proves: R36

### Out of scope

- The other new commands. Done in: task 09, task 12, task 15

## Task 09: validate-command

```yaml
covers: [R11, R12, R13, R14, R18, R19, R36, R37]
depends_on: ["05", "06", "08"]
touches: [internal/slice/validate.go, internal/slice/open.go, internal/slice/validate_test.go, internal/cli/validate.go, internal/cli/problems.go, internal/e2e/validate_command_test.go]
follows: ["Flow 2: Check a spec", "Flow 8: Wrong use, and the order of the first checks", Core API, CLI, Cross-platform paths, Tests, Testing]
risk: core
```

### Goal

`sliceflow spec validate [<name>] [--strict]` works as Flow 2 says.
`openSlice` runs R11's first checks in order, for this command and the
two that write.

### Steps

1. Write the tests listed under Done when. Run them, and see them fail.
2. `internal/slice/open.go`: `openSlice`.
3. `internal/slice/validate.go`: `Validation`, `Validate`.
4. `internal/cli/problems.go`: `printProblems`.
5. `internal/cli/validate.go`: the command, with the summary line and
   R14's exit code.
6. Run the tests until they pass.

### Files

- `internal/slice/open.go`: new. `openSlice`.
- `internal/slice/validate.go`, `internal/slice/validate_test.go`: new.
- `internal/cli/problems.go`: new. `printProblems`.
- `internal/cli/validate.go`: new.
- `internal/e2e/validate_command_test.go`: new.

### Interfaces

- `func openSlice(ctx context.Context, r Repo, name string) (string, error)`:
  creates. R14 of 0001-slice-and-events for a bad name, R33 for no
  slice, R11 for no worktree; returns the worktree. Tasks 12 and 14 use
  it.
- `type Validation struct{ Problems []spec.Problem; Errors, Warnings, Notes int }`:
  creates.
- `func Validate(ctx context.Context, r Repo, name string, strict bool) (Validation, error)`:
  creates. `name` `""` means the current branch's slice.
- `func printProblems(w io.Writer, ps []spec.Problem)`: creates. One
  `spec.md:<line>: <level>: <message>` line each. Task 15 uses it.
- `currentSliceName`, `findWorktree`, `spec.Parse`, `spec.Check`: uses.

### Done when

- `TestValidateProblems`: S5's spec prints S5's 4 lines on stdout, exit 1.
  Proves: R13, R14, R37
  Scenario: S5
- `TestValidateWarningOnly`: S6 prints `0 errors, 1 warning, 1 note`,
  exit 0.
  Proves: R14
  Scenario: S6
- `TestValidateStrict`: S7 exits with 1.
  Proves: R14
  Scenario: S7
- `TestValidateByName`: S8 on `main` prints `0 errors, 0 warnings, 0 notes`.
  Proves: R13
  Scenario: S8
- `TestValidateNoBranch`: S9 prints S9's 3 lines on stderr, exit 1.
  Proves: R12
  Scenario: S9
- `TestValidateSkipsCode`: S10's spec reports neither line.
  Proves: R18
  Scenario: S10
- `TestValidateNoWorktree`: S11 prints S11's error, exit 1.
  Proves: R11
  Scenario: S11
- `TestValidateNoSpec`: S12 prints S12's error, exit 1.
  Proves: R19
  Scenario: S12
- `TestValidateJSON`: S31 prints S31's error, exit 2.
  Proves: R36
  Scenario: S31
- `TestOpenSliceOrder`: `spec validate Bad_Name` gives R14's error of
  0001-slice-and-events; `no-such-slice` gives R33's; a slice with no
  worktree gives R11's.
  Proves: R11

### Out of scope

- Submitting. Done in: task 14, task 15

## Task 10: lock

```yaml
covers: [R6, R7]
depends_on: []
touches: [internal/slice/lock.go, internal/slice/lock_test.go, internal/slice/repo.go]
follows: [The lock, Data, TD2, Interrupts, Cross-platform paths, Tests, Testing]
risk: core
```

### Goal

`takeLock` takes and frees a slice's lock as The lock says, with the
lock file of Data, and gives R7's error after the wait.

### Steps

1. Write the tests listed under Done when. Run them, and see them fail.
2. `internal/slice/repo.go`: `Repo.LockWait`.
3. `internal/slice/lock.go`: `takeLock`.
4. Run the tests until they pass.

### Files

- `internal/slice/lock.go`, `internal/slice/lock_test.go`: new.
- `internal/slice/repo.go`: changed. `LockWait`.

### Interfaces

- `Repo.LockWait time.Duration`: creates. `0` means 10 s.
- `func takeLock(ctx context.Context, r Repo, name string) (func(), error)`:
  creates. The returned function frees the lock. Task 11 uses it.

### Done when

- `TestLockFile`: while held, `.sliceflow/locks/login-google.lock` holds
  one JSON line with this process's `pid` and `since` in `TimeLayout`;
  after the free, it is gone; `.sliceflow/.gitignore` is `*\n` when it
  was missing.
  Proves: R6
- `TestLockBusy`: with a lock file holding pid 4312 and
  `2026-10-24T10:02:00Z`, and `LockWait` 200 ms, `takeLock` returns
  `slice login-google is busy: another sliceflow command holds .sliceflow/locks/login-google.lock (pid 4312, since 10:02 UTC). If no sliceflow command is running, delete that file.`
  Proves: R7
- `TestLockUnreadable`: a lock file holding `x` gives
  `(pid unknown, since unknown UTC)`.
  Proves: R7
- `TestLockWaits`: a lock freed after 100 ms is taken, with `LockWait`
  1 s.
  Proves: R6

### Out of scope

- What happens while the lock is held. Done in: task 11

## Task 11: save-event

```yaml
covers: [R2, R6, R8, R9, R10]
depends_on: ["01", "02", "06", "10"]
touches: [internal/slice/save.go, internal/slice/save_test.go]
follows: [Saving an event, Undo after a failed commit, Leftovers after a crash, Interrupts, TD1, Git arguments, Cross-platform paths, Tests, Testing]
risk: core
```

### Goal

`saveEvent` is the one write path: with the lock held, it replays the
committed log, checks the move, writes the log and the event's files,
commits only those, and undoes all of it when the commit fails. The
hardest part is the undo, and Ctrl-C during the commit.

### Steps

1. Write the tests listed under Done when. Run them, and see them fail.
2. `internal/slice/save.go`: `saveOp` and `saveEvent`, as Saving an
   event and Undo after a failed commit say.
3. Run the tests until they pass.

### Files

- `internal/slice/save.go`, `internal/slice/save_test.go`: new.

### Interfaces

- `type saveOp struct{ Event string; Data any; Message string; Write map[string][]byte; Keep []string; Check func(state.State) error; Tail string }`:
  creates. `Write`: files to write, by path in the slice folder (e.g.
  `reviews/spec-1.md`). `Keep`: files committed as they are on disk
  (e.g. `spec.md`). `Check`: the command's own check of the state, run
  after `state.Allowed`. `Tail`: R10's last sentence, `""` for
  `Nothing was changed.`.
- `func saveEvent(ctx context.Context, r Repo, name, worktree string, op saveOp) (string, state.State, error)`:
  creates. Returns the new commit's full id and the state after the
  event. Tasks 12 and 14 use it.
- `takeLock`, `state.Replay`, `state.Allowed`, `eventlog.New`,
  `eventlog.Encode`, `git.Run`, `git.Show`: uses.

### Done when

- `TestSaveCommit`: saving `slice.titled` makes one commit
  `login-google: title set` holding only the log, whose last line has
  `seq` 2.
  Proves: R8, R9
- `TestSaveOnlyOwnFiles`: with `src/app.go` staged, the commit holds
  only the log and `spec.md`, and `src/app.go` is still staged after.
  Proves: R8
- `TestSaveNoNewline`: a committed log whose last line is half written
  with no `\n` keeps that line, and the new event starts on its own
  line with `seq` 1 more than the last good line.
  Proves: R9
- `TestSaveNotAllowed`: a `review.finished` while `draft` saves nothing,
  and returns R4's reason.
  Proves: R2
- `TestSaveHookFails`: with a `pre-commit` hook that exits with 1, no
  commit is made, the log on disk equals the committed log, a file in
  `Write` is gone, nothing is staged, and the error is
  `could not save the event: git commit failed (exit code 1). Nothing was changed.`
  Proves: R10
- `TestSaveTail`: with `Tail` `The steps before it were saved; spec.md is submitted.`,
  the error ends with it.
  Proves: R10
- `TestSaveCancelled`: a cancelled `ctx` during the commit undoes as
  above, with the reason `interrupted`, and frees the lock.
  Proves: R10, R6
- `TestSaveTwoAtOnce`: two `saveEvent` calls of `slice.titled` started at
  once both succeed, with `seq` 2 and 3.
  Proves: R6
- `TestSaveSeqAfterSkipped`: a committed log whose last line is a
  well-formed `review.finished` with `seq` 2 that R4 skips gives the new
  event `seq` 3.
  Proves: R9
- `TestSaveLeftoverTemp`: a `.events.jsonl.tmp` left on disk is gone
  after the next save.
  Proves: R9

### Out of scope

- The commands that call it. Done in: task 12, task 14

## Task 12: title-command

```yaml
covers: [R32, R33, R34, R36]
depends_on: ["08", "09", "11"]
touches: [internal/slice/title.go, internal/slice/title_test.go, internal/cli/title.go, internal/e2e/title_command_test.go]
follows: ["Flow 6: Set a title", Core API, CLI, Interrupts, Git arguments, Tests, Testing]
risk: normal
```

### Goal

`sliceflow slice title <name> <title>` saves `slice.titled` as Flow 6
says.

### Steps

1. Write the tests listed under Done when. Run them, and see them fail.
2. `internal/slice/title.go`: `SetTitle`.
3. `internal/cli/title.go`: the command, with `missing title`, and the
   signal `ctx` as `slice create` has.
4. Run the tests until they pass.

### Files

- `internal/slice/title.go`, `internal/slice/title_test.go`: new.
- `internal/cli/title.go`: new.
- `internal/e2e/title_command_test.go`: new.

### Interfaces

- `func SetTitle(ctx context.Context, r Repo, name, title string) error`:
  creates.
- `openSlice`, `saveEvent`: uses.

### Done when

- `TestTitleSet`: S25 prints S25's line, makes S25's commit and log
  line, and `slice list` then shows S25's row.
  Proves: R32, R34
  Scenario: S25
- `TestTitleSame`: S26 prints S26's error, exit 1, no commit.
  Proves: R33
  Scenario: S26
- `TestTitleEmpty`: S27 prints S27's error, exit 1.
  Proves: R33
  Scenario: S27
- `TestTitleMissing`: S28 prints S28's error, exit 2.
  Proves: R36
  Scenario: S28
- `TestTitleUnknownState`: a slice whose log has no good `slice.created`
  gives R32's error, exit 1.
  Proves: R32

### Out of scope

- Showing the title. Done in: task 06, task 07

## Task 13: reviewer

```yaml
covers: [R25, R26, R27]
depends_on: []
touches: [internal/review/]
follows: [Reviewer, Review files, Tests, Testing]
risk: core
```

### Goal

A new package `internal/review` holds the `Reviewer` contract, the
placeholder, the review file's front matter, and the reader that gets
the result from a review file.

### Steps

1. Write the tests listed under Done when. Run them, and see them fail.
2. `internal/review/review.go`: `Reviewer`, `Request`, `FrontMatter`,
   `ReadResult`.
3. `internal/review/placeholder.go`: `Placeholder`.
4. Run the tests until they pass.

### Files

- `internal/review/review.go`, `internal/review/placeholder.go`: new.
- `internal/review/review_test.go`: new.

### Interfaces

- `type Reviewer interface{ Name() string; Review(ctx context.Context, req Request) error }`:
  creates. Task 14 uses it.
- `type Request struct{ Slice, Doc, File, Commit, Worktree string; Round int }`:
  creates. `File` is the review file's full path, as in the plan.
- `func Placeholder() Reviewer`: creates. `Name` is `placeholder`;
  `Review` adds R26's text after the front matter.
- `func FrontMatter(slice string, round int, reviewer string) []byte`:
  creates. R25's front matter.
- `func ReadResult(file []byte) string`: creates. `needs_human` when
  `## Needs human decision` is not `None.`, or is missing; else `passed`.

### Done when

- `TestFrontMatter`: `FrontMatter("login-google", 1, "placeholder")` is
  exactly R25's block.
  Proves: R25
- `TestPlaceholder`: after `Review`, the file is R25's block and then
  exactly R26's text.
  Proves: R26
- `TestReadResult`: R26's file gives `passed`; a file with `- **Q1.**`
  under `## Needs human decision` gives `needs_human`; a file with no
  such section gives `needs_human`.
  Proves: R27

### Out of scope

- Saving the review events. Done in: task 14

## Task 14: submit-core

```yaml
covers: [R20, R21, R22, R23, R24, R25, R27, R28]
depends_on: ["05", "09", "11", "13"]
touches: [internal/slice/submit.go, internal/slice/submit_test.go]
follows: ["Flow 3: Submit, refused", "Flow 4: Submit, accepted, and the review", "Flow 5: Submit again from submitted", Core API, Saving an event, Interrupts, Tests, Testing]
risk: core
```

### Goal

`slice.Submit` runs Flows 3, 4, and 5: the state check, the strict
check, `spec.submitted`, and the review with its two events, reporting
each step as it happens. The hardest parts are the three separate
saves, each under its own lock, and finding the submit's commit when it
goes on from `submitted`.

### Steps

1. Write the tests listed under Done when. Run them, and see them fail.
2. `internal/slice/submit.go`: `Step`, `Submission`, `Submit`.
3. Run the tests until they pass.

### Files

- `internal/slice/submit.go`, `internal/slice/submit_test.go`: new.

### Interfaces

- `type Step struct{ Kind string; Problems []spec.Problem; Refused *state.Counts; Before string; Round int; Reviewer, File, Result string }`:
  creates. `Kind` is `problems`, `refused`, `submitted`, `started`, or
  `finished`. `Before` is the spec's state before a refused submit, for
  R22's last sentence. Task 15 prints each kind.
- `type Submission struct{ Refused bool }`: creates.
- `func Submit(ctx context.Context, r Repo, name string, step func(Step)) (Submission, error)`:
  creates. Task 15 calls it.
- `var reviewer = review.Placeholder()`: creates. The one reviewer of
  this slice. (lead's call)
- `openSlice`, `saveEvent`, `specChanged`, `spec.Parse`, `spec.Check`,
  `review.FrontMatter`, `review.ReadResult`: uses.

### Done when

- `TestSubmitRefused`: S13's spec gives the steps `problems` and
  `refused` (0 errors, 1 warning, before `draft`), one commit
  `login-google: spec submit refused` holding `spec.md` and the log.
  Proves: R21, R22
- `TestSubmitAccepted`: a finished spec gives `submitted`, `started`
  (round 1, `placeholder`, `reviews/spec-1.md`), and `finished`
  (`passed`), three commits as S14, and `review.started`'s `commit` is
  the first of them.
  Proves: R21, R23, R25, R27, R28
- `TestSubmitStates`: `review_passed` unchanged gives S15's error;
  `in_review` gives S18's; no `slice.created` gives R20's third error;
  `review_passed` with an edited `spec.md` is accepted with round 2.
  Proves: R20
- `TestSubmitReviewFileExists`: S21's file gives R24's error, no commit.
  Proves: R24
- `TestSubmitGoOn`: a log ending with an accepted `spec.submitted`,
  then a `slice.titled`, goes on with `started` and `finished` only;
  `commit` is the commit that added that `spec.submitted`.
  Proves: R20, R25
- `TestSubmitMovedFirst`: when another save moves the state between the
  steps, the next step returns R6's error, with
  `The steps before it were saved.`
  Proves: R20

### Out of scope

- Printing, and the exit codes. Done in: task 15

## Task 15: submit-command

```yaml
covers: [R6, R7, R8, R9, R10, R20, R21, R22, R23, R24, R25, R28, R29, R30, R37]
depends_on: ["07", "12", "14"]
touches: [internal/cli/submit.go, internal/e2e/submit_command_test.go, internal/e2e/no_network_check_test.go]
follows: ["Flow 3: Submit, refused", "Flow 4: Submit, accepted, and the review", "Flow 5: Submit again from submitted", CLI, Interrupts, No network, Tests, Testing]
risk: normal
```

### Goal

`sliceflow spec submit [<name>]` prints each step as it happens, with
R22's and R28's lines, on stdout, and exits as the spec says. Its e2e
tests cover the submit scenarios end to end, and 0001's no-network test
runs the new commands too.

### Steps

1. Write the tests listed under Done when. Run them, and see them fail.
2. `internal/cli/submit.go`: the command, with the signal `ctx` as
   `slice create` has; `printProblems` for the problems step.
3. `internal/e2e/no_network_check_test.go`: add `spec validate`,
   `spec submit`, `slice title`, and `template show`.
4. Run the tests until they pass.

### Files

- `internal/cli/submit.go`: new.
- `internal/e2e/submit_command_test.go`: new.
- `internal/e2e/no_network_check_test.go`: changed.

### Interfaces

- `Submit`, `Step`, `Submission`, `printProblems`: uses.

### Done when

- `TestSubmitRefusedCmd`: S13's output, exit 1, commit, log line, and
  status lines.
  Proves: R20, R21, R22, R29, R37
  Scenario: S13
- `TestSubmitAcceptedCmd`: S14's output, exit 0, three commits, three
  log lines, and status lines.
  Proves: R8, R21, R23, R25, R28, R29
  Scenario: S14
- `TestSubmitUnchanged`: S15's error, exit 1, no commit.
  Proves: R20
  Scenario: S15
- `TestSubmitChangedStatus`: S16's status lines.
  Proves: R29, R30
  Scenario: S16
- `TestSubmitRound2`: S17's lines.
  Proves: R20, R25, R30
  Scenario: S17
- `TestSubmitInReview`: S18's error, exit 1.
  Proves: R20
  Scenario: S18
- `TestSubmitBusy`: S19's error after 10 s, exit 1, no commit.
  Proves: R6, R7
  Scenario: S19
- `TestSubmitHookFails`: S20's error, exit 1, log as committed, no
  commit.
  Proves: R9, R10
  Scenario: S20
- `TestSubmitFileExists`: S21's error, exit 1.
  Proves: R24
  Scenario: S21
- `TestSubmitOwnFiles`: S22: `src/app.go` is in no new commit, and still
  staged.
  Proves: R8
  Scenario: S22
- `TestSubmitStopsHalfway`: S23's output with a `commit-msg` hook, exit
  1, and its status lines.
  Proves: R10, R29
  Scenario: S23
- `TestSubmitGoesOn`: S24's output, exit 0, two new commits.
  Proves: R20
  Scenario: S24
- `TestNoNetworkNewCommands`: with `GIT_TRACE` set, the four new commands
  run no `fetch`, `pull`, or `push` (the project rule No network).
  Proves: R40 of 0001-slice-and-events

### Out of scope

- A real reviewer. Done in: not in this slice
