---
review: task-03-1
slice: 0002-spec-and-state
kind: task
target: task 03
round: 1
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

`Parse` covers R18 well: code, inline code, comments, and `\r\n` are
handled, the five tests are real, and the template and 0001's and
0002's specs read cleanly. But two gaps leave tasks 04 and 05 without
what they need, and they cannot fix them, since `parse.go` is not in
their touches: an `ideas` list with no spaces before the `-` is lost
(F1), and a marker like `**R 7.**` is neither an item nor a bad marker
(F2).

## Checked

- Commits: 5b38b6f
- Files read: internal/spec/doc.go, internal/spec/parse.go,
  internal/spec/parse_test.go, internal/template/template.go,
  internal/template/spec.md, docs/slices/0002-spec-and-state/spec.md
  (R15 to R18), docs/slices/0002-spec-and-state/tasks.md (tasks 03 to
  05), docs/slices/0001-slice-and-events/spec.md, the brief, status.md
  Findings
- Ran:
  go test ./internal/spec/ -v (pass, 5 tests)
  a copy of the package in a scratch module, run on the spec template,
  0001's spec, and 0002's spec (every R has Check:, every S has Shows:,
  every SD has Why:, Not chosen:, and Rule:; 0001 and 0002 give 0
  placeholders and 0 comments; the template gives 32 and 11)
  the same copy on small inputs: an unindented ideas list, an unclosed
  front matter, bad marker forms, a section of only code, a label with a
  placeholder over 2 lines, inline code over a line break
- Requirements: R18; R15 to R17 only as far as `Doc` must serve them
- Interfaces: Parse, Doc, Field, Section, Item, Label, and Mark match
  the brief exactly

## Tests

- `TestParseSections`: ok
- `TestParseSkips`: ok
- `TestParseCRLF`: ok
- `TestParseRemoved`: ok
- `TestParseFrontMatter`: ok

## Previous findings

None.

## Findings

### F1. An ideas list with no spaces before the - is lost

Level: high
Where: internal/spec/parse.go:15, internal/spec/parse.go:146-150
Breaks: R15 (`ideas` is a list when followed by lines that start with
  `- `, with or without spaces before the `-`), task 03 Interfaces
  (`ListItems` holds the `- ` lines under the key)
Why: `listItemRe` is `^\s+-\s+`, so it needs at least one space. With
  `---`, `ideas:`, `- docs/ideas/a.md`, `---`, Parse gives the field
  `ideas` with Value "" and no ListItems; the `- ` line matches neither
  regexp and is dropped. Task 04 would then report
  `front matter ideas is not a list` for a valid spec. Task 04 cannot
  fix it: `parse.go` is not in its touches. The test only tries the
  indented form.
Fix: use `^\s*-\s+` for list lines, and add the unindented case to
  `TestParseFrontMatter`.

### F2. A marker with a space, like **R 7.**, is not recorded

Level: high
Where: internal/spec/parse.go:11
Breaks: R15 (the marker row: checked on a list item that starts with
  `**` and then R, S, SD, SC, or Q, and a digit or a space), task 03
  Interfaces (Parse records what it cannot read, for the checks)
Why: `markerLike` needs a digit right after the letters. I parsed
  `- **R 7.** x` and `- **SD 1.** z`: neither is an item, and neither is
  in `BadMarkers`, so task 04 cannot report
  `"<text>" is not a marker; write it like **R7.**`. The regexp also
  catches lines R15 does not check: `**R9.** prose` (not a list item)
  and `- **AB12** odd` (not R, S, SD, SC, or Q) land in `BadMarkers`,
  and a nested `  - **R8.** x`, a good marker, does too. Task 04 can
  filter those out by their text, but it cannot recover the missing ones.
Fix: match R15's row exactly, e.g. `^- \*\*(R|S|SD|SC|Q)[0-9 ]` on a
  line that is not a good item, and add these cases to a test.

### F3. A front matter that never closes hides the whole file

Level: medium
Where: internal/spec/parse.go:43-53
Breaks: task 03 Interfaces (Parse records what it cannot read, for the
  checks)
Why: with `---`, `slice: x`, then `## Summary` and text, and no second
  `---`, Parse gives HasFront true, the field `slice`, and no sections.
  `Doc` has no sign that the front matter never closed, so task 04
  cannot give R15's `no front matter: the file must start with a line
  ---` (the front matter must be between two `---` lines). It would
  report every section as missing instead. The ASSUMED line chose this,
  and it changes what the checks can report.
Fix: set HasFront true only when the closing `---` is found, and then
  read the file from line 1 for sections; or ask the lead.

### F4. A section with only code in it is Empty

Level: medium
Where: internal/spec/parse.go:60-63, internal/spec/parse.go:88
Breaks: R16 (a section is empty when it holds nothing but blank lines
  and comments)
Why: `Text` is built from the line with code blanked, and `Empty` is
  `Text == ""`. `## Why`, then a fenced block, gives Empty true; so does
  a section whose only text is `` `only` ``. Task 05 would then warn
  `"## Why" is empty; write None. if there is nothing to say` for a
  section that has content.
Fix: set Empty from the lines with only comments blanked (the `vis`
  form, with fence lines counted as content), and keep Text as it is.

### F5. Inline code over a line break is read wrong

Level: low
Where: internal/spec/parse.go:197-210
Breaks: nothing
Why: Markdown lets inline code go over a line break in a paragraph, but
  the scan pairs backticks within one line only. I parsed a paragraph
  whose first line ends inside inline code, and whose second line holds
  the end of that code, then a real placeholder, then more inline code.
  The second line's backticks pair the wrong way, so the placeholder is
  hidden, and no Placeholders mark is made. The ASSUMED line chose this.
  Specs here wrap at 72 columns, so this can happen, though 0001's and
  0002's specs read cleanly.

## Requirements check

None.

## Needs human decision

None.
