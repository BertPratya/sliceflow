---
slice: spec
---

# Tasks

## Task 01: split-cli

```yaml
covers: []
depends_on: []
touches: [internal/cli/]
follows: [TD1]
risk: normal
```

### Goal

`internal/cli` keeps only Run, parsing, routing, and help. The slice
commands move to `internal/cli/slicecmd`, and what every group shares
(the command types, `Input`, the printing helpers, the shared message
text) moves to `internal/cli/command`. No output changes.

### Steps

1. Make `internal/cli/command` with the exported command types
   (`Registry`, `Group`, `Action`, `Arg`, `Option`), `Input` with
   exported `Args` and `Options`, and the printing helpers that are not
   about one group: `PrintFailure`, `PrintGitText`, `IsStop`,
   `PrintBadLines`.
2. Move the slice commands and their tests to `internal/cli/slicecmd`.
   It exports `Group`, the slice group of the command table.
3. `internal/cli` builds its table from `slicecmd.Group`, and keeps
   parsing, routing, help, and their tests. Its tests that read the
   command table use the package it is now in.

### Interfaces

- `command.Input`: creates. Every group's Run gets it.
- `command.PrintFailure`, `command.PrintGitText`, `command.IsStop`,
  `command.PrintBadLines`: creates. Tasks 03, 04, 06, 07 use them.
- `slicecmd.Group`: creates. Task 07 and task 08 change it.

### Done when

- Every existing test passes, unchanged except for the package it is in
  and the names that became exported.
  Proves: none

### Out of scope

- Any change in what a command prints. Done in: task 03, task 07,
  task 08

## Task 02: state-machine

```yaml
covers: [created-starts-in-draft, edited-keeps-the-state, unknown-event, event-not-allowed]
depends_on: []
touches: [internal/state/, internal/eventlog/]
follows: [TD2, TD3]
risk: core
```

### Goal

Done on `main` in 5058180 and 3d224e7, written with the human:
`state.Replay`, `Event.Line`, and `Event.Payload`.

### Steps

None

### Done when

- `internal/state` tests 1 to 13, and `internal/eventlog` tests 14 to
  18.
  Proves: created-starts-in-draft, edited-keeps-the-state,
  unknown-event, event-not-allowed

### Out of scope

- Printing the warnings. Done in: task 03

## Task 03: status-with-state

```yaml
covers: [state-and-next, status-of-unknown-state, damaged-line-skipped, unknown-event, event-not-allowed, status-shows-edits]
depends_on: ["01", "02"]
touches: [internal/slice/status.go, internal/slice/find.go, internal/cli/slicecmd/, e2e/status_test.go, e2e/status_state_test.go]
follows: [Flow 1, TD3, Testing]
risk: normal
```

### Goal

`slice status` prints the State: line, the next step for a known state,
the current title and description, and the warning about the state
after the warning about damaged lines.

### Steps

1. `slice.StatusResult` gets `State state.State` and `Spec string`, the
   full path of the spec in the user's checkout. `Status` replays the
   events with `state.Replay`.
2. `slicecmd` prints the State: line after the first line, the title and
   description from `State.Detail`, then the empty line, Next: and Then
   run: when the state is known.
3. `command.PrintStateWarning` prints the warning of `unknown-event` or
   `event-not-allowed`, after the damaged-lines warning.
4. The existing status e2e tests get the State: line and the next step,
   since `state-and-next` changes the output. Name: `e2e/status_test.go`.

### Interfaces

- `slice.StatusResult.State`, `slice.StatusResult.Spec`: creates.
- `command.PrintStateWarning(w io.Writer, logPath string, u state.Unknown)`:
  creates. Task 07 uses it.

### Done when

- `TestStatusShowsStateAndNext`: the example of `state-and-next`.
  Proves: state-and-next, created-starts-in-draft
- `TestStatusUnknownEvent`, `TestStatusEventNotAllowed`: the examples of
  `unknown-event` and `event-not-allowed`.
  Proves: unknown-event, event-not-allowed
- `TestStatusOfUnknownState`: the example of `status-of-unknown-state`.
  Proves: status-of-unknown-state
- `TestStatusDamagedLineFirst`: both examples of `damaged-line-skipped`.
  Proves: damaged-line-skipped
- `TestStatusShowsEdits`: both examples of `status-shows-edits`, with
  the log written by the test.
  Proves: status-shows-edits

### Out of scope

- `slice list`. Done in: task 08

## Task 04: spec-template

```yaml
covers: [template-printed, template-help, template-in-any-folder, spec-help]
depends_on: ["01"]
touches: [internal/cli/, e2e/spec_template_test.go]
follows: [TD1, Testing]
risk: low
```

### Goal

The `spec` group, with `spec template`, which prints `templates.Spec`.

### Steps

1. `internal/cli/speccmd` with `Group`: description "Print the spec
   template, and find problems in specs", and the command `template`.
2. `internal/cli` adds `spec` to its table.

### Interfaces

- `speccmd.Group`: creates. Task 06 changes it.

### Done when

- `TestSpecTemplate`: prints `templates.Spec` exactly, exit 0, in a
  folder that is not a git repo.
  Proves: template-printed, template-in-any-folder
- `TestSpecTemplateHelp`, `TestSpecHelp`: the help examples, and
  `spec template extra` exits 2.
  Proves: template-help, spec-help

### Out of scope

- `spec validate`. Done in: task 06

## Task 05: spec-reader

```yaml
covers: [spec-lines-read, code-blocks-skipped]
depends_on: []
touches: [internal/spec/]
follows: [TD4, Paths and line ends on every system]
risk: core
```

### Goal

`internal/spec` turns spec text into marked lines, then into a built
spec: sections, stories, rules, examples, decisions, and questions,
each with its line number.

### Steps

1. `lines.go`: split the text into lines (`\n` or `\r\n`, a byte order
   mark skipped), mark each line plain, code, or comment, and find the
   comments and the code spans of each plain line.
2. `build.go`: the sections, stories, rules, Check:, Uses:, examples,
   decisions with their Rule: lines, and questions, from the plain
   lines.

### Interfaces

- `spec.Read(text []byte) *Spec`: creates. Task 06 uses it.

### Done when

- Unit tests of the line marking: code blocks, comments, both orders,
  backtick runs, line ends, and the byte order mark.
  Proves: spec-lines-read, code-blocks-skipped
- Unit tests of the build: headings, Check:, Uses:, examples, items.
  Proves: spec-lines-read

### Out of scope

- The problems. Done in: task 06

## Task 06: spec-validate

```yaml
covers: [spec-has-no-problems, problems-listed, spec-needs-a-story, story-heading-shape, stories-numbered-in-order, story-needs-a-rule, story-needs-uses, uses-names-shared-rules, shared-rule-used, rule-in-unknown-section, section-heading-exact, rule-heading-shape, rule-name-shape, rule-name-unique, rule-has-one-check, check-values, rule-needs-an-example, example-has-when-then, placeholder-left, comment-left, open-questions-left, decisions-numbered-in-order, item-shape, decision-needs-rule-line, decision-rule-names-exist, code-blocks-skipped, validate-errors-as-before, validate-needs-the-spec, validate-in-any-state, validate-help]
depends_on: ["04", "05"]
touches: [internal/spec/, internal/slice/, internal/cli/speccmd/, e2e/spec_validate_test.go]
follows: [Flow 2, TD4, Testing]
risk: normal
```

### Goal

`spec validate <name>` prints every problem of the spec, sorted, with a
hint, and a count; or "no problems found".

### Steps

1. `spec.Check(s *Spec) []Problem`: the checks, in the order of Story 3,
   then sorted by line, with only the placeholder problem kept on a line
   that has one.
2. `slice.SpecPath(ctx, dir, name)`: the checks of
   `validate-errors-as-before`, and the full path of the spec.
3. `speccmd`: the command `validate`.

### Interfaces

- `spec.Problem{Line int; Message, Hint string}`, `spec.Check`: creates.
- `slice.SpecPath`: creates.

### Done when

- One unit test per rule of Story 3 from `spec-needs-a-story` to
  `code-blocks-skipped`, with the examples' values.
- The 0002 spec itself has no problems.
- e2e: `spec-has-no-problems`, `problems-listed`,
  `validate-errors-as-before`, `validate-needs-the-spec`,
  `validate-in-any-state`, `validate-help`.

### Out of scope

- Warnings. Done in: not in this slice

## Task 07: slice-edit

```yaml
covers: [edit-changes-the-slice, edited-event, edit-empty-values, nothing-to-change, edit-needs-an-option, edit-commits-the-log, edit-on-the-base, edit-undo, no-edit-without-state, edit-errors-as-before, edit-help]
depends_on: ["03"]
touches: [internal/slice/, internal/git/, internal/cli/slicecmd/, e2e/edit_test.go]
follows: [Flow 3, TD5, Testing]
risk: core
```

### Goal

`slice edit <name> [--title] [--description]` appends `slice.edited`
under the lock and commits only the log, with undo.

### Steps

1. `slice.Edit(ctx, dir, EditOptions) (EditResult, error)`: the checks
   in the order of `edit-errors-as-before`, then the values, then the
   steps: append (undo: cut the log back), commit only the log (undo:
   unstage it).
2. `slicecmd`: the command `edit`, with its messages.

### Interfaces

- `slice.Edit`, `slice.EditOptions`, `slice.EditResult`: creates.
- `slice.NoStateError`: creates.

### Done when

- e2e tests of every example of Story 4 from `edit-changes-the-slice` to
  `edit-help`.

### Out of scope

- `slice list`. Done in: task 08

## Task 08: list-and-help

```yaml
covers: [list-shows-edits, slice-help-lists-edit]
depends_on: ["07"]
touches: [internal/slice/list.go, internal/cli/, e2e/list_test.go, e2e/help_test.go]
follows: [Testing]
risk: low
```

### Goal

`slice list` shows the current title and description; the help lists
`edit` and the new group descriptions.

### Steps

1. `List` replays each whole log with `state.Replay`.
2. The slice group's description becomes "Make, see, and change
   slices".

### Done when

- `TestListShowsEdits`: the example of `list-shows-edits`.
  Proves: list-shows-edits
- `TestSliceHelpListsEdit`: the example of `slice-help-lists-edit`.
  Proves: slice-help-lists-edit

### Out of scope

- None
