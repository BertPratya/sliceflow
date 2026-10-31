# Plan: 0002 spec and state

How the rest of `docs/specs/0002-spec-and-state.md` is built, in fast
mode (the human, 2026-10-31). Already built: R23, R27 (`internal/state`),
and R29-R35 (`slice title`). This plan covers R1-R22, R24-R26, R28, R36.

## Packages

0001's table still holds. One new package:

| Package | Job | Never |
|---|---|---|
| `internal/spec` | reads a spec's text and checks it (R9-R22); returns a list of problems | runs git, reads files, prints |

`internal/slice` finds the slice, its worktree, and the file (R1, R5-R8),
reads the file, and calls `spec.Check`. `internal/cli` prints (R2-R4).
`internal/slice/spec.go` stays: it makes the spec of a new slice from the
template. `internal/spec` only reads specs.

## Plan decisions

0001's PD1-PD8 still hold. New ones:

- **PD9. A problem is data.** `spec.Problem{Line int; Message, Hint
  string}`. `Line` is 0 for a problem with no line. `Hint` has no period
  at the end; `cli` adds it. The level is always `error` in 0002, so it is
  not a field; `cli` prints it.
- **PD10. One entry point.** `spec.Check(content []byte, name string,
  today time.Time) []Problem`. `name` is the slice name (R12's `slice`
  check). `today` gives the date in R12's hint for a missing `created`
  (PD8: time comes in as a value). The checks run in the order R12-R22,
  each adding its problems. Then a stable sort puts line 0 first, then by
  line. So on one line the R12-R22 order stays (R2).
- **PD11. The reader marks what is content.** The reader (R9-R11) turns
  the text into lines. Each line keeps its number, its text, and a
  `masked` copy where inline code is replaced by spaces. Lines in a code
  block, and lines inside an HTML comment, are flagged. The checks read
  only what the flags allow, so no check has to know about code or
  comments.
- **PD12. The next step comes from the state.** `state.Next(p Place,
  name string) string` returns the `Next:` text for a place, or `""` when
  there is none (R25). `slice.Info` gets `Next string`, filled in
  `ReadLog`. It never reads `spec.md` (SD5).
- **PD13. The spec template has one source (PD5 again).**
  `slice.SpecTemplate() (string, error)` runs only `specEdits` on
  `templates.Spec`, with `\r\n` turned into `\n`. `specFile` calls it,
  then fills `{{slice}}` and `{{date}}`. `spec template` prints
  `SpecTemplate()` (R28).

## Assumed (the spec does not settle these)

- **A code fence may be indented.** A line whose text, after leading
  spaces, starts with a run of 3 or more backticks, with no backtick in
  the rest of the line, opens a code block. It closes at a fence line
  whose run is at least as long, with only spaces after it. The spec's
  own code blocks are indented 2 spaces under list items, and S14's
  output holds `---`, `<!--`, and `{{slice}}`, which must stay content.
  A line like `` ```` ``` ```` `` (R10's own text) is inline code, not a
  fence.
- **Inline code** is a run of N backticks up to the next run of exactly N
  backticks on the same line. A run with no partner is plain text.
- **HTML comments.** From a `<!--` to the next `-->` (also across lines),
  sliceflow reads no heading, item, or labeled line. R14 still reports
  each `<!--` outside code, and each `{{` outside code, also inside a
  comment.
- **The front matter is not closed:** only that problem from R12 (line 1),
  and the body starts at line 2.
- **No front matter:** that problem has no line, and the body starts at
  line 1.
- **Line of a problem.** A problem about a line gets that line: a labeled
  line's problem gets the labeled line; a missing label gets the item's
  line; a story problem gets its `###` heading. A problem about something
  missing from the whole file (a missing section, a missing front matter
  key, no front matter) has no line.
- **R15, the first item is not 1:** the message is
  `<id> comes first: expected <kind>1`, with R15's hint. After an item
  out of order, the next expected number is its number + 1, so one gap
  gives one problem, not a chain.
- **R13, empty section:** a section is empty when it has no line other
  than blank lines and comment lines.
- **R13, out of order:** a section is out of order when it is not in the
  longest run of sections that is already in the right order. So one
  moved section gives one problem, not one for each section it passed.
- **R16 and R17, no cascades.** An item whose home section is missing
  gets no R16 problem: the missing section covers it. A scenario under
  any `###` heading in `User stories` is under a story; a bad heading
  gets only R17's heading problem. Story numbers are read loosely
  (`story 2`, no colon) for R15.

## Chunks

Each chunk is one commit. All chunks keep `gofmt`, `go vet ./...` and
`go test ./...` green.

### Chunk 1: status shows State and Next; spec template

- **Covers:** R24, R25, R26, R28, PD12, PD13.
- **Files:** `internal/state/state.go` (+test), `internal/slice/status.go`,
  `internal/slice/spec.go` (+test), `internal/cli/slice.go`,
  `internal/cli/cli.go`, new `internal/cli/spec.go`,
  `internal/e2e/status_test.go`, `internal/e2e/list_test.go`, new
  `internal/e2e/spec_template_test.go`.
- **Behavior:** `slice status` prints `State:` after `Phase:` and `Next:`
  last, as in S11; both `-` when the phase is `unknown`. The JSON of
  `status` and of each slice in `list --json` gets `state` and `next`
  after `phase`, both left out when unknown (S12). `slice list` text does
  not change. A new top-level group `spec` with the command `template`;
  it works outside a repo, exit 0. The `spec` group's `about`:
  `Check and read slice specs`.
- **Tests:** unit for `state.Next`; unit that `SpecTemplate()` keeps
  `{{slice}}` and `{{date}}` and has no `title:` line; e2e S11, S12 (the
  existing 0001 status and list tests are updated), S14.

### Chunk 2: reading a spec

- **Covers:** R9, R10, R11, PD11.
- **Files:** new `internal/spec/read.go` (+test).
- **Names (inside the package):** `doc` holds `lines`, the front matter
  range, `headings` (`level`, `text`, `line`), and `items`. An item has
  `id` (e.g. `SD3`), `kind` (`R`, `S`, `SD`, `Q`), `num`, `line`,
  `removed`, `section` (the `##` heading it is under, or `""`),
  `storyLine` (the line of the `###` heading it is under, whatever its
  text, or 0), and `labels`.
  A label has `name`, `value` (continuation lines joined with one space,
  trimmed), and `line`. Known labels: `Check`, `Shows`, `Why`,
  `Not chosen`, `Cost`, `Rule`, `Replaces`.
- **Tests:** table tests for each rule of R9-R11 and each Assumed line
  above: CRLF, line numbers, an indented fence, a heading and an item
  inside a fence, inline code with one and two backticks, an unmatched
  backtick, a comment over several lines with an item inside, a labeled
  line with a 4-space continuation, a removed item.

### Chunk 3: checks R12-R17

- **Covers:** R12, R13, R14, R15, R16, R17, PD9, PD10.
- **Files:** new `internal/spec/check.go`, `internal/spec/check_test.go`.
- **Names:** `spec.Problem`, `spec.Check`.
- **Tests:** for each message of R12-R17, a small spec with that one
  problem gives exactly that problem (message, hint, line). The built-in
  template filled by hand gives no problems. The raw template (PD13)
  gives placeholder and comment problems.

### Chunk 4: checks R18-R22

- **Covers:** R18, R19, R20, R21, R22.
- **Files:** `internal/spec/check.go`, `internal/spec/check_test.go`.
- **Tests:** each message of R18-R22, as in chunk 3. A removed item is
  checked only for its number. The order of problems on one line.

### Chunk 5: the command

- **Covers:** R1-R8, R36, S1-S10.
- **Files:** new `internal/slice/validate.go`, `internal/cli/spec.go`,
  `internal/cli/cli.go`, new `internal/e2e/spec_validate_test.go`.
- **Names:** `slice.ValidateSpec(dir, name string, today time.Time)
  (SpecResult, error)`; `SpecResult{Name, File, Path string; Problems
  []spec.Problem}`. `File` is the full path; `Path` is relative to `dir`
  with `/` (R2), after both sides are resolved with
  `filepath.EvalSymlinks` (macOS `/tmp` is a link).
- **Behavior:** R5-R8 in that order, using `findWorktree` (R6, R7) and
  R5's three lines with `sliceflow spec validate <name>`. The file is
  read from disk, also when not committed. `cli` prints R2-R4, exit 1
  when there is a problem. Nothing is written (R36).
- **Tests:** e2e S1-S10, each named after its scenario, e.g.
  `TestValidateS1`; R36: `git status --porcelain` and every branch are
  the same before and after.
