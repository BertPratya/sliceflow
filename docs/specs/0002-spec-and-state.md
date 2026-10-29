---
slice: 0002-spec-and-state
title: Spec and state
created: 2026-10-29
---

# Spec

## Summary

A user checks a slice's spec with one command, the way a compiler checks
code: each problem with its line and a hint on how to fix it. `slice
status` also shows the slice's state and what to do next, worked out from
the event log by a state machine that later slices grow move by move.
A user can set or change a slice's title, which is the first command
that adds an event to an existing log.
Later steps (the plan, the tasks, the run) read the spec as data, so the
facts they rely on must be checked by code, not by eye.

## User stories

### Story 1: Check a spec

A user, or an AI agent, checks that a slice's spec is complete before
anyone builds from it.

- **S1.** **Given** the slice `spec-submit`, whose spec has no problems,
  and I am at the root of the main checkout,
  **When** I run `sliceflow spec validate spec-submit`,
  **Then** sliceflow prints:
  ```
  $ sliceflow spec validate spec-submit
  No problems in .sliceflow/worktrees/spec-submit/docs/slices/spec-submit/spec.md.
  ```
  and the exit code is 0.
  Shows: R1, R3, R36
- **S2.** **Given** the spec of `spec-submit` has no `Check:` line under
  R3 on line 48, and S4 on line 61 has `Shows: R9`, and there is no R9,
  **When** I run `sliceflow spec validate spec-submit`,
  **Then** sliceflow prints:
  ```
  $ sliceflow spec validate spec-submit
  .sliceflow/worktrees/spec-submit/docs/slices/spec-submit/spec.md:48: error: R3 has no Check: line
    hint: add a line "Check: e2e" under R3. The kinds are unit, api, e2e, and manual.
  .sliceflow/worktrees/spec-submit/docs/slices/spec-submit/spec.md:61: error: S4 shows R9, but there is no R9
    hint: name only requirements that exist, e.g. "Shows: R1, R2".
  2 errors.
  ```
  and the exit code is 1.
  Shows: R2, R3, R18, R19
- **S3.** **Given** the spec of `spec-submit` has no problems except
  the open question Q1 on line 210,
  **When** I run `sliceflow spec validate spec-submit`,
  **Then** sliceflow prints:
  ```
  $ sliceflow spec validate spec-submit
  .sliceflow/worktrees/spec-submit/docs/slices/spec-submit/spec.md:210: error: Q1 is not answered
    hint: write the answer into the spec, and remove Q1. When no question is left, write None.
  1 error.
  ```
  and the exit code is 1.
  Shows: R3, R22
- **S4.** **Given** I am in the worktree of `spec-submit`, on branch
  `slice/spec-submit`, and its spec has the problems of S2,
  **When** I run `sliceflow spec validate`,
  **Then** sliceflow prints the lines of S2, with the path
  `docs/slices/spec-submit/spec.md`.
  Shows: R1, R2
- **S5.** **Given** the spec of `spec-submit` still has the placeholder
  `{{Two or three sentences: what this feature lets someone do, and why.}}`
  on line 12,
  **When** I run `sliceflow spec validate spec-submit`,
  **Then** the output starts with:
  ```
  .sliceflow/worktrees/spec-submit/docs/slices/spec-submit/spec.md:12: error: placeholder left from the template: {{Two or three sentences: what this feat…
    hint: replace it with real text.
  ```
  and the exit code is 1.
  Shows: R14
- **S6.** **When** I run `sliceflow spec validate spec-submit --json`
  for the spec of S2,
  **Then** sliceflow prints one JSON object:
  ```
  {"slice":"spec-submit","file":"<full path>/.sliceflow/worktrees/spec-submit/docs/slices/spec-submit/spec.md","problems":[{"line":48,"level":"error","message":"R3 has no Check: line","hint":"add a line \"Check: e2e\" under R3. The kinds are unit, api, e2e, and manual."},{"line":61,"level":"error","message":"S4 shows R9, but there is no R9","hint":"name only requirements that exist, e.g. \"Shows: R1, R2\"."}],"errors":2}
  ```
  and the exit code is 1.
  Shows: R4
- **S7.** **Given** I am on `main`,
  **When** I run `sliceflow spec validate`,
  **Then** sliceflow prints on stderr:
  ```
  Error: not on a slice branch (you are on main).
  Give a name: sliceflow spec validate <name>
  See all slices: sliceflow slice list
  ```
  and the exit code is 1.
  Shows: R5
- **S8.** **Given** the branch `slice/old-idea` exists, but no worktree
  has it checked out,
  **When** I run `sliceflow spec validate old-idea`,
  **Then** sliceflow prints on stderr
  `Error: slice old-idea has no worktree: branch slice/old-idea is not checked out anywhere.`
  and the exit code is 1.
  Shows: R7
- **S9.** **When** I run `sliceflow spec validate no-such-slice`,
  **Then** sliceflow prints on stderr
  `Error: no slice named no-such-slice. See all slices: sliceflow slice list`
  and the exit code is 1.
  Shows: R6
- **S10.** **Given** the spec file of `spec-submit` was deleted in its
  worktree,
  **When** I run `sliceflow spec validate spec-submit`,
  **Then** sliceflow prints on stderr
  `Error: spec not found: .sliceflow/worktrees/spec-submit/docs/slices/spec-submit/spec.md.`
  and the exit code is 1.
  Shows: R8

### Story 2: See where a slice is, and what comes next

A user, or an AI agent, sees the state of a slice and the next thing to
do.

- **S11.** **Given** the slice `spec-submit`, just created from `main` at
  commit `9dd7ebf`,
  **When** I run `sliceflow slice status spec-submit`,
  **Then** sliceflow prints:
  ```
  $ sliceflow slice status spec-submit
  Slice:    spec-submit
  Title:    spec-submit
  Phase:    planning
  State:    spec_draft
  From:     main @ 9dd7ebf
  Created:  2026-10-24 07:42 UTC
  Next:     write the spec, then check it: sliceflow spec validate spec-submit
  ```
  Shows: R24, R25
- **S12.** **When** I run `sliceflow slice status login-google --json`,
  **Then** sliceflow prints one JSON object:
  ```
  {"slice":"login-google","title":"Login with Google","phase":"planning","state":"spec_draft","next":"write the spec, then check it: sliceflow spec validate login-google","from":{"branch":"develop","commit":"<full commit id>"},"created":"2026-10-24T08:10:00Z","warnings":[]}
  ```
  Shows: R26
- **S13.** **Given** the log of `spec-submit` has a second
  `slice.created` event on line 2, with `seq` 2,
  **When** I run `sliceflow slice status spec-submit`,
  **Then** sliceflow prints the status of S11, and prints on stderr:
  ```
  Warning: spec-submit: .events.jsonl line 2: slice.created is not allowed in planning (spec_draft); skipped
  ```
  and the exit code is 0.
  Shows: R27

### Story 3: Read the spec rules again

A user, or an AI agent, reads the template's rules again before changing
a finished spec, which no longer has its rule comments.

- **S14.** **When** I run `sliceflow spec template`,
  **Then** sliceflow prints the built-in spec template, the same text
  `slice create` writes, with `{{slice}}` and `{{date}}` left as they
  are:
  ```
  $ sliceflow spec template
  ---
  slice: {{slice}}
  created: {{date}}
  ---

  <!-- HOW TO FILL THIS TEMPLATE
  ...
  ```
  and the exit code is 0.
  Shows: R28

### Story 4: Name a slice

A user, or an AI agent, sets or changes the title of a slice.

- **S15.** **Given** the slice `spec-submit` from S11, and an
  uncommitted change to its `spec.md`,
  **When** I run
  `sliceflow slice title spec-submit "Submit a spec for review"`,
  **Then** sliceflow prints:
  ```
  $ sliceflow slice title spec-submit "Submit a spec for review"
  Title of spec-submit set: Submit a spec for review
  ```
  and the exit code is 0,
  and the branch `slice/spec-submit` has one new commit,
  `spec-submit: title set`, that changes only
  `docs/slices/spec-submit/.events.jsonl`, with one new line, like:
  ```
  {"version":1,"seq":2,"event":"slice.titled","at":"2026-10-29T13:05:00Z","data":{"title":"Submit a spec for review"}}
  ```
  and the change to `spec.md` is still not committed,
  and `sliceflow slice status spec-submit` shows
  `Title:    Submit a spec for review`.
  Shows: R29, R32
- **S16.** **When** I run `sliceflow slice title spec-submit "  "`,
  **Then** sliceflow prints on stderr
  `Error: the title is empty. Nothing was changed.`
  and the exit code is 1.
  Shows: R31
- **S17.** **Given** the log of `spec-submit` in its worktree ends with
  a half-written line 2, with no `\n` at its end,
  **When** I run `sliceflow slice title spec-submit "Submit a spec"`,
  **Then** sliceflow prints on stderr
  `Warning: spec-submit: .events.jsonl line 2: half-written last line; removed`,
  and sets the title as in S15, with the new event on line 2,
  and the exit code is 0.
  Shows: R33
- **S18.** **Given** another sliceflow is adding an event to the log of
  `spec-submit`, and does not finish within 10 seconds,
  **When** I run `sliceflow slice title spec-submit "Submit a spec"`,
  **Then** after 10 seconds sliceflow prints on stderr
  `Error: the event log of spec-submit is busy: another sliceflow is writing to it. Try again. Nothing was changed.`
  and the exit code is 1.
  Shows: R34
- **S19.** **Given** the repo has a `pre-commit` hook that prints
  `error: commits are blocked` and fails,
  **When** I run `sliceflow slice title spec-submit "Submit a spec"`,
  **Then** sliceflow prints on stderr
  `Error: could not set the title of spec-submit: error: commits are blocked. The event log was put back as it was.`
  and the log in the worktree is as it was before,
  and the exit code is 1.
  Shows: R35

## Requirements

### The command

- **R1.** When the user runs `sliceflow spec validate <name>`, sliceflow
  checks the file `docs/slices/<name>/spec.md` in the worktree where the
  branch `slice/<name>` is checked out, as it is on disk, also when it is
  not committed. When no name is given and the current branch is
  `slice/<name>`, sliceflow checks the spec of `<name>`.
  Check: e2e
- **R2.** sliceflow prints each problem as two lines on stdout:
  `<path>:<line>: <level>: <message>`, then `  hint: <hint>.` `<level>`
  is `error`: in this slice every problem is an error, also an open
  question, because nothing may be built from a spec with a problem.
  Every hint ends with a period; R12-R22 write
  the hints without it. `<path>` is relative to the current folder.
  A problem that belongs to no line, e.g. a missing section, prints as
  `<path>: <level>: <message>`. Problems with no line come first, then the
  others by line, and on one line in the order of R12-R22.
  Check: e2e
- **R3.** After the problems, sliceflow prints `<n> errors.`
  (`1 error.` for one). With no problems at all, it prints
  only `No problems in <path>.` The exit code is 1 when there is at least
  one error, else 0. A spec with errors is a result, not a failed command:
  this is the one exception to 0001's R38 ("nothing on stdout when a
  command fails").
  Check: e2e
- **R4.** Where `--json` is given, sliceflow prints, instead of R2's and
  R3's lines, one JSON object with `slice`, `file` (the full path),
  `problems` (a list, empty when there are none), and `errors` (the
  count). Each problem has `line` (left out when the
  problem has no line), `level`, `message`, and `hint` (with its
  period), in the order of R2. The exit code is as in R3.
  Check: e2e
- **R5.** If no name is given and the current branch is not a `slice/`
  branch, sliceflow prints the three lines of S7, with `on <branch>` or
  `on no branch`, and exits with code 1. Outside a git repo, or in a repo
  with no commits, it prints the error of 0001's R12 or R13 and exits
  with code 1.
  Check: e2e
- **R6.** If no local branch `slice/<name>` exists, sliceflow prints
  `Error: no slice named <name>. See all slices: sliceflow slice list`
  and exits with code 1.
  Check: e2e
- **R7.** If no worktree has the branch `slice/<name>` checked out,
  sliceflow prints
  `Error: slice <name> has no worktree: branch slice/<name> is not checked out anywhere.`
  and exits with code 1.
  Check: e2e
- **R8.** If the worktree has no `docs/slices/<name>/spec.md`, sliceflow
  prints `Error: spec not found: <path>.` with `<path>` as in R2, and
  exits with code 1. The checks of R5-R8 run in that order, and stop at
  the first that fails.
  Check: e2e

### Reading the spec

- **R9.** sliceflow reads the spec as UTF-8, with `\n` or `\r\n` line
  endings. Line numbers count from 1, at the first line of the file.
  Check: unit
- **R10.** A code block is the lines from a line that starts with
  ```` ``` ```` up to the next such line, both included. Text inside a
  code block, and inside `inline code`, is content: sliceflow never reads
  a heading, an item, a labeled line, a placeholder, or a comment there.
  Check: unit
- **R11.** An item is a line that starts with `- **<id>.**`, e.g.
  `- **R3.**`, and the lines under it, up to the next item or heading.
  A labeled line is a line of an item that starts with 2 spaces and then
  `<label>:`, e.g. `  Check: e2e`. Its value goes on over the next lines
  that are indented by 4 spaces or more. A removed item is an item whose
  text is `(removed in R<3 digits>: <why>)`, e.g.
  `- **R4.** (removed in R002: merged into R3)`; sliceflow checks only
  its number.
  Check: unit

### The checks

Each check gives an error. sliceflow
runs every check and reports every problem it finds, not only the first.
In the messages, `<id>` is an item's id, e.g. `R3`.

- **R12.** The front matter. The file starts with a line `---`, and the
  front matter ends at the next line `---`. Each line between them is
  `<key>: <value>`. The keys are exactly `slice` and `created`, each once.
  `slice` is the slice's name; `created` is a date as `YYYY-MM-DD`. The
  messages and hints:
  - `no front matter` / `start the file with the front matter of: sliceflow spec template`
  - `front matter is not closed` / `add a line "---" after the last key`
  - `front matter line is not "<key>: <value>"` / `write it as e.g. "created: 2026-10-29"`
  - `front matter has no <key>` / `add the line "<key>: <example>"`,
    where the example is the slice's name for `slice`, and today's date
    for `created`
  - `front matter has <key> twice` / `keep one`
  - `unknown key <key> in the front matter` / `remove it. The keys are slice and created; the title lives in the event log`
  - `front matter says slice <value>, but this is the spec of <name>` / `change it to "slice: <name>"`
  - `created <value> is not a date as YYYY-MM-DD` / `write it as e.g. "created: 2026-10-29"`
  Check: unit
- **R13.** The headings. After the front matter, the first heading is
  `# Spec`. The sections are these `##` headings, each once, in this
  order: `Summary`, `User stories`, `Requirements`, `Out of scope`,
  `Decisions`, `Open questions`. A section's text runs to the next `##`
  heading. A section may not be empty. Only `Out of scope`, `Decisions`,
  and `Open questions` may be the one line `None.` The messages and
  hints:
  - `the first heading is not "# Spec"` / `add the line "# Spec" after the front matter`
  - `missing section "## <section>"` / `add it after "## <section before it>". With nothing to say, write None`,
    or for `Summary`: `add it after "# Spec"`
  - `unknown section "## <heading>"` / `the sections are: Summary, User stories, Requirements, Out of scope, Decisions, Open questions`
  - `section "## <section>" appears twice` / `merge them into one`
  - `section "## <section>" is out of order` / `put the sections in this order: Summary, User stories, Requirements, Out of scope, Decisions, Open questions`
  - `section "## <section>" is empty` / `write its text, or None`,
    or, for a section that may not be `None.`: `write its text`
  - `section "## <section>" cannot be None` / `a spec needs a summary, at least one story, and at least one requirement`
  Check: unit
- **R14.** Nothing is left from the template: no `{{` and no `<!--`
  outside code (R10). Each one is a problem on its own line. A
  placeholder in the message is shown up to its `}}`, and cut after 40
  characters: the first 40 characters, then `…`. The messages and hints:
  - `placeholder left from the template: <placeholder>` / `replace it with real text`
  - `comment left from the template` / `delete it. A finished spec has no comments`
  Check: unit
- **R15.** The numbers. Requirements (R), scenarios (S), decisions (SD),
  questions (Q), and stories (`### Story <n>:`) are each numbered 1, 2,
  3, ... from top to bottom, with no number missing and none twice. A
  removed item keeps its number. The messages and hints:
  - `<id> appears twice (first on line <n>)` / `number them in order from the top, and update the Shows: and Rule: lines`
  - `<id> comes after <id before it>: expected <expected id>` / the same hint
  Check: unit
- **R16.** Each item is in its section: R in `Requirements`, S in `User
  stories`, SD in `Decisions`, Q in `Open questions`. Message and hint:
  `<id> is in the wrong section` / `move it to "## <its section>"`
  Check: unit
- **R17.** The stories. Under `User stories`, each `###` heading is
  `### Story <n>: <title>`, with a title. Under it comes at least one
  line of text before its first scenario, then at least one scenario.
  Every scenario is under a story. The messages and hints:
  - `heading under "## User stories" is not "### Story <n>: <title>"` / `write it as e.g. "### Story 1: Start a new slice"`
  - `Story <n> has no sentence` / `add one sentence under the heading: who wants to do what`
  - `Story <n> has no scenarios` / `add a scenario under it, e.g. "- **S1.** **When** ..."`
  - `<id> is not under a story` / `add a heading "### Story <n>: <title>" above it`
  Check: unit
- **R18.** Each scenario that is not removed has one `Shows:` line, which
  lists one or more requirement ids, split by commas, e.g.
  `Shows: R1, R2`. Each one names a requirement that exists and is not
  removed. The messages and hints:
  - `<id> has no Shows: line` / `add a line "Shows: R1" under <id>, naming the requirements it shows`
  - `<id> has two Shows: lines` / `merge them into one`
  - `<id>'s Shows: line does not list requirements` / `write it as e.g. "Shows: R1, R2"`
  - `<id> shows <R id>, but there is no <R id>` / `name only requirements that exist, e.g. "Shows: R1, R2"`
  - `<id> shows <R id>, but <R id> is removed` / `remove <R id> from the Shows: line`
  Check: unit
- **R19.** Each requirement that is not removed has one `Check:` line,
  which lists one or more of `unit`, `api`, `e2e`, and `manual`, split by
  commas. The messages and hints:
  - `<id> has no Check: line` / `add a line "Check: e2e" under <id>. The kinds are unit, api, e2e, and manual`
  - `<id> has two Check: lines` / `merge them into one`
  - `<id> has an unknown check kind "<kind>"` / `use unit, api, e2e, or manual`
  Check: unit
- **R20.** Each decision that is not removed has the labeled lines
  `Why:`, `Not chosen:`, `Cost:`, and `Rule:`, each once, each with a
  value, in this order, and may have `Replaces:` as the last one. The
  messages and hints:
  - `<id> has no <label>: line` / `add a line "<label>: ..." under <id>`,
    or for `Cost:`: `add a line "Cost: ..." under <id>: the downside we accept, or None`
  - `<id> has two <label>: lines` / `keep one`
  - `<id>'s <label>: line is empty` / `write its value`
  - `<id>'s lines are out of order` / `put them in this order: Why:, Not chosen:, Cost:, Rule:, Replaces:`
  Check: unit
- **R21.** A decision's `Rule:` is `none`, or one or more requirement ids
  split by commas, each one a requirement that exists and is not
  removed. A `Replaces:` is `<slice>/SD<n>`, e.g. `device-pairing/SD3`.
  The messages and hints:
  - `<id>'s Rule: names <R id>, but there is no <R id>` / `name a requirement that exists, or write none`
  - `<id>'s Rule: names <R id>, but <R id> is removed` / the same hint
  - `<id>'s Rule: is not "none" or a list of requirements` / `write it as e.g. "Rule: R11", or "Rule: none"`
  - `<id>'s Replaces: is not "<slice>/SD<n>"` / `write it as e.g. "Replaces: device-pairing/SD3"`
  Check: unit
- **R22.** Each question that is not removed gives the error
  `<id> is not answered`, with the hint
  `write the answer into the spec, and remove <id>. When no question is left, write None`
  Check: unit

### State and what comes next

- **R23.** sliceflow works out a slice's phase and state by replaying its
  log, event by event. A move says which event moves a slice from which
  phase and state to which. In this slice there is one move:
  `slice.created` moves a new slice to the phase `planning` and the state
  `spec_draft`. A slice with no good `slice.created` event has the phase
  `unknown` and no state. This replaces 0001's R25.
  Check: unit
- **R24.** `sliceflow slice status` prints a `State:` line after
  `Phase:`, and a `Next:` line at the end, as in S11. Both show `-` when
  the phase is `unknown`. This changes 0001's R29 and S13.
  Check: e2e
- **R25.** While a slice is in `planning` (`spec_draft`), `Next:` is
  `write the spec, then check it: sliceflow spec validate <name>`.
  `Next:` comes from the log only: sliceflow never reads `spec.md` for
  it.
  Check: unit
- **R26.** Where `--json` is given, the object of `slice status` holds
  `state` and `next` after `phase`, as in S12, both left out when the
  phase is `unknown`. The slice objects of `slice list --json` hold them
  too. `slice list` without `--json` does not change.
  Check: e2e
- **R27.** If an event is good, but no move allows it in the slice's
  phase and state at that point, sliceflow skips it and adds the warning
  `.events.jsonl line <n>: <event> is not allowed in <phase> (<state>); skipped`,
  or `in <phase>` alone when the slice has no state, e.g. `in unknown`.
  The skipped line still counts as the last good line for 0001's R27, so
  its `seq` is checked and the next line's `seq` follows it. This adds a
  third warning to 0001's S17, whose line 3 is a second `slice.created`.
  Check: unit

### The template

- **R28.** When the user runs `sliceflow spec template`, sliceflow prints
  on stdout the spec template that `slice create` writes (0001's R7),
  with `{{slice}}` and `{{date}}` left as they are, and exits with code 0.
  It works outside a git repo too.
  Check: e2e

### The title

- **R29.** When the user runs `sliceflow slice title <name> <title>`,
  sliceflow adds a `slice.titled` event, with `data` holding `title`, to
  the log in the slice's worktree (the worktree of R1), and commits the
  log on the slice branch with the message `<name>: title set`. The
  commit holds only the log file. Other changes in the worktree stay as
  they are, not committed. It is a normal git commit, as in 0001's R8.
  Then sliceflow prints `Title of <name> set: <title>`.
  Check: e2e
- **R30.** Where `--json` is given, sliceflow prints, instead of R29's
  line, one JSON object with `slice` and `title`.
  Check: e2e
- **R31.** If the title is empty, or only spaces, sliceflow prints
  `Error: the title is empty. Nothing was changed.` and exits with code
  1. Without a slice of that name, or without its worktree, sliceflow
  prints the error of R6 or R7. Both the name and the title must be
  given; a missing one is wrong use (0001's R38), with
  `missing slice name` or `missing title`.
  Check: e2e
- **R32.** `slice.titled` is a move from `planning` (`spec_draft`) to
  `planning` (`spec_draft`). `slice status` and `slice list` show the title of
  the newest `slice.titled` event, else the title of `slice.created`.
  Check: unit
- **R33.** Before it adds an event, if the log in the worktree does not
  end with `\n`, sliceflow removes the text after the last `\n` and adds
  the warning `.events.jsonl line <n>: half-written last line; removed`.
  The new event's `seq` is 1 more than the `seq` of the last good line,
  or 1 when there is none.
  Check: unit
- **R34.** Only one sliceflow at a time adds an event to a slice's log.
  If another sliceflow is adding one, sliceflow waits up to 10 seconds,
  then prints
  `Error: the event log of <name> is busy: another sliceflow is writing to it. Try again. Nothing was changed.`
  and exits with code 1. A lock left by a sliceflow that stopped, e.g.
  after a crash, never blocks.
  Check: unit, e2e
- **R35.** If a step fails after sliceflow started to write, sliceflow
  puts the log file back as it was, and prints
  `Error: could not set the title of <name>: <reason>. The event log was put back as it was.`
  `<reason>` is as in 0001's R18.
  Check: e2e

### Every command

- **R36.** `sliceflow spec validate` never changes a file, a branch, or
  the event log.
  Check: e2e

## Out of scope

- Approving the spec. The spec, plan, and tasks are approved once, together, in 0003.
- `spec submit` and the spec review. They need real agents, in 0004.
- Showing the title in `spec.md`, written from the log. Stays in the backlog.
- Renaming a slice. The name is the slice's ID, fixed for good (0001's SD1); only the title changes.
- Checks of writing quality, e.g. the EARS form, "shall", or a user-visible requirement with no scenario. The spec reviewer judges these, from 0004.
- A requirement or scenario number reused after approval. Comes with the approval, in 0003.
- Checking that a `Replaces:` names a real decision of another slice.
- Errors as JSON with a stable code. Stays in the backlog.

## Decisions

- **SD1.** The spec, plan, and tasks are approved once, together, at the
  end of planning in 0003. There is no spec approval of its own (the
  human, 2026-10-29).
  Why: it matches the 2026-10-28 templates, and the approval code (the
  terminal prompt, `via`, the hash, the tag) is written once.
  Not chosen: a spec approval in 0002 and a plan approval in 0003,
  because the human would approve twice per slice, and 0002 would be
  bigger.
  Cost: a spec is only checked, never approved, until the plan is done.
  Rule: none
- **SD2.** `spec validate` checks only the facts that later code relies
  on. Writing quality is the spec reviewer's job (the human, 2026-10-29).
  Why: rules go in code, judgment goes to agents (design 1.3). Text
  matching for good writing is fragile.
  Not chosen: also checking the EARS form and that every user-visible
  requirement has a scenario, because it costs about 8-15 more hours.
  Cost: until 0004, only the human and the lead catch a badly written
  requirement.
  Rule: none
- **SD3.** Problems go to stdout, and the exit code is 1 when there is an
  error.
  Why: the text and the `--json` output go to the same place, and both
  can be piped. A spec with errors is the command's result.
  Not chosen: problems on stderr, as compilers print them, because then
  `--json` and the text would go to different places.
  Cost: an exception to 0001's R38, which says nothing goes to stdout
  when a command fails.
  Rule: R2, R3
- **SD4.** Paths in the problems are relative to the current folder.
  Why: an editor's terminal opens `<path>:<line>` with one click only
  when the path is relative to where the command ran.
  Not chosen: relative to the root of the main checkout, as in 0001's
  R9, because those paths cannot be clicked from inside a worktree.
  Cost: the same spec shows a different path from different folders.
  Rule: R2
- **SD5.** `Next:` comes from the event log only.
  Why: `slice status` stays a quick view of the log. The same log always
  gives the same answer.
  Not chosen: `slice status` runs the spec check and says how many
  problems are left, because then a status depends on a file that is
  not committed.
  Cost: `Next:` says "check it" even when the spec is already clean.
  Rule: R25
- **SD6.** An event that no move allows is skipped with a warning, like
  a bad line.
  Why: a warning never stops a command (0001's SD5), so one wrong line
  never hides the rest of the slice.
  Not chosen: stopping with an error, because one bad line would make
  the slice unreadable.
  Cost: a slice may show a state that ignores a line someone wrote by
  hand.
  Rule: R27
- **SD7.** `slice title` is in this slice (the human, 2026-10-29).
  Why: it is a small command that users want early.
  Not chosen: moving it to 0003, with the approval, the next command
  that adds an event.
  Cost: this slice also builds the first write to a log: the lock, the
  fix for a half-written last line, and the commit of the log alone.
  Rule: R29
- **SD8.** A half-written last line is removed, with a warning.
  Why: such a line is never a whole event, and it can only come from a
  write that stopped halfway. Kept, it would give a warning on every
  later read.
  Not chosen: ending it with `\n` and keeping it as a bad line, because
  it holds nothing that can be read.
  Cost: the text of the half-written line is lost.
  Rule: R33

## Open questions

None.
