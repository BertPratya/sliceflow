---
review: task-04-1
slice: 0002-spec-and-state
kind: task
target: task 04
round: 1
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

The task covers R15 and R18's skip of placeholder values. Every message
matches R15's table word for word, and the tests are real. One medium
problem: a marker above the first `##` heading is never reported (F1).
The other findings are low.

## Checked

- Commits: c25ef94
- Files read: internal/spec/check.go, internal/spec/errors.go,
  internal/spec/errors_test.go, internal/spec/doc.go,
  internal/spec/parse.go, internal/template/template.go,
  internal/template/spec.md, internal/spec/testdata/fresh.md,
  internal/spec/testdata/0001-spec.md,
  docs/slices/0002-spec-and-state/spec.md (R15, R16, R18, Out of scope),
  the brief, the commit message with its ASSUMED lines
- Ran:
  go test ./internal/spec/ (pass)
  diff internal/spec/testdata/0001-spec.md docs/slices/0001-slice-and-events/spec.md (same)
  diff of testdata/fresh.md after its front matter against the body of internal/template/spec.md (same)
  probe tests on a copy of the package in my scratchpad: empty labels, quotes, markers before the first section, inline code in a bad marker, sections out of order, a removed duplicate, and Check run on 0002's own spec.md (0 errors)
- Requirements: R15, R18 (placeholder values only)
- Interfaces: `Problem{Line int; Level, Message string}` and
  `Check(d Doc, slice string) []Problem` match the task; `Parse` and `Doc`
  used as given

## Tests

- `TestErrorRows`: ok
- `TestErrorPlaceholderValue`: ok
- `TestErrorsFresh`: ok
- `TestErrors0001`: ok

## Previous findings

None.

## Findings

### F1. A marker above the first section is never reported

Level: medium
Where: internal/spec/parse.go:89, internal/spec/errors.go:180-230
Breaks: R15
Why: R15 says each marker is in its section, e.g. R in Requirements. I
  put `- **R9.** x` between `# Spec` and `## Summary` of the test spec:
  Check gives no error. The same holds for `- **R2 x` there: no "is not
  a marker" error. Parse drops every line before the first `##` heading
  (`if cur == nil { continue }`), so Check never sees them. R15 expects
  `R9 is outside "## Requirements"`. The input is rare, but the error is
  silent.
Fix: Doc needs the items and bad markers found before the first section,
  e.g. in a section with no name. parse.go is task 03's file and not in
  this task's touches, so the lead decides: widen touches, or defer to a
  fix task.

### F2. Quotes and backslashes in a value are shown with Go escapes

Level: low
Where: internal/spec/errors.go:86, internal/spec/errors.go:96, internal/spec/errors.go:197, internal/spec/errors.go:321
Breaks: nothing
Why: The messages use `%q`. `Goes to: "x"` gives `Goes to: "\"x\"" is
  not a place; ...`. The user sees backslashes that are not in the file.
  R15's form is `"<x>"`.
Fix: write `"` + value + `"` with `%s` instead of `%q`.

### F3. The "is not a marker" text loses its inline code

Level: low
Where: internal/spec/errors.go:197
Breaks: nothing
Why: The text comes from `Mark.Text`, which parse.go takes from the
  line with inline code blanked. ``- **R2 uses `x` here`` gives
  `"**R2 uses     here" is not a marker; ...`, with spaces where the
  code was.
Fix: keep the visible text (inline code kept) for the message, or ask
  task 03's owner to store it in `BadMarkers`.

### F4. An empty Check:, Shows:, or Rule: gives no error

Level: low
Where: internal/spec/errors.go:242, internal/spec/errors.go:261, internal/spec/errors.go:285
Breaks: nothing
Why: `Check:` with nothing after it has no kinds, so no "unknown check"
  error; the requirement has no check kind at all. The same for an empty
  `Shows:` and an empty `Rule:`, which names no requirement and does not
  say none. R15 does not say what an empty value gets, so this is not a
  break.
Fix: if wanted, report an empty value with the row's "has no ... line"
  message, after the human agrees.

## Requirements check

None.

## Needs human decision

None.
