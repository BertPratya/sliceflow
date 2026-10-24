---
review: task-03-2
slice: 0002-spec-and-state
kind: task
target: task 03
round: 2
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

All five findings of round 1 are fixed, and the new tests prove the
fixes. `Doc` now gives tasks 04 and 05 what R15 to R18 need. One new
problem comes from the F5 fix: a stray backtick in one list item now
hides the next item, so a valid requirement disappears from `Doc` (F1).

## Checked

- Commits: 5b38b6f, 31a6ab8 (the diff 5b38b6f..31a6ab8 in full)
- Files read: internal/spec/doc.go, internal/spec/parse.go,
  internal/spec/parse_test.go, the brief,
  docs/slices/0002-spec-and-state/spec.md (R15 to R18),
  docs/slices/0002-spec-and-state/reviews/task-03-1.md, status.md
  (Lead's decisions), both commit messages and their ASSUMED lines
- Ran:
  go test ./internal/spec/ (pass)
  a copy of the package in a scratch module, run on the spec template,
  0001's spec, and 0002's spec (10, 9, and 9 sections; every live R has
  a Check: line; no bad markers; the template gives 32 placeholders and
  11 comments, the two specs 0 and 0)
  the same copy on small inputs: an unindented and a mixed ideas list,
  an unclosed front matter, a file of only `---`, bad marker forms
  (`**R7**`, `**R7.**x`, two spaces after `-`, `**SC 1.**`,
  `**Read**`), a fence inside a comment, inline code over a line break,
  and a stray backtick in one list item followed by another item
  the stray-backtick input on the round 1 code from 5b38b6f (both items
  read)
- Requirements: R18; R15 and R16 as far as `Doc` must serve them
- Interfaces: Parse, Doc, Field, Section, Item, Label, and Mark still
  match the brief exactly

## Tests

- `TestParseSections`: ok
- `TestParseSkips`: ok
- `TestParseCRLF`: ok
- `TestParseRemoved`: ok
- `TestParseFrontMatter`: ok

## Previous findings

- task-03-1/F1: verified (fixed)
- task-03-1/F2: verified (fixed)
- task-03-1/F3: verified (fixed)
- task-03-1/F4: verified (fixed)
- task-03-1/F5: verified (fixed)

## Findings

### F1. A stray backtick in one list item hides the next item

Level: medium
Where: internal/spec/parse.go:224-227, internal/spec/parse.go:284-291
Breaks: nothing
Why: `spansLines` looks for the closing backticks on every later line
  up to a blank line, so it runs over the start of a new list item. I
  parsed `## Requirements`, then an item R1 whose text holds one lone
  backtick, its `Check: unit` line, then an item R2 whose text holds a
  pair of backticks (real inline code), and its `Check: unit` line.
  Parse gives one item, R1, with two `Check:` labels (lines 4 and 6);
  R2 is neither an item nor a bad marker. The round 1 code read both
  items. Task 04 would then report `<id> shows R2, which does not
  exist` for every scenario that shows R2, and nothing points at the
  stray backtick. Markdown ends inline code at the end of its block,
  and a new list item is a new block. The fix of task-03-1/F5 caused
  this.
Fix: in `spansLines`, also stop at a line that starts a new block: a
  list item (`- ` after any spaces), a `#` heading, or a fence. Add the
  input above to a test.

## Requirements check

None.

## Needs human decision

None.
