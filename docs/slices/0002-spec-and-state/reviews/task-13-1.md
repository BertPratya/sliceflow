---
review: task-13-1
slice: 0002-spec-and-state
kind: task
target: task 13
round: 1
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

The task covers R25 to R27 as the brief asks, and the three tests are
real. The front matter and the placeholder's text match R25 and R26 byte
for byte. One low problem: a `## ` line inside a code block or a comment,
after `None.` in `## Needs human decision`, ends the section early, so
that file gives `passed` (F1).

## Checked

- Commits: d10236e (d10236e~1..d10236e), and its `ASSUMED` lines
- Files read: internal/review/review.go, internal/review/placeholder.go,
  internal/review/review_test.go, the brief, spec.md (R25, R26, R27,
  S20), plan.md (Reviewer, Review files), tasks.md (tasks 13 and 15),
  docs/templates/review.md, AGENTS.md
- Ran:
  go test ./internal/review/ (pass)
  grep for os/exec and net/ in internal/review/ (no match)
  cat -A on spec.md R25 and R26, compared with the Go strings by hand
  (same bytes, "\n" line ends, the file ends with "None.\n")
- Requirements: R25 (front matter), R26, R27 (the result rule of the
  plan's Review files)
- Interfaces: `Reviewer`, `Request`, `Placeholder`, `FrontMatter`, and
  `ReadResult` match task 13's Interfaces and plan "Reviewer". `Request`
  has the same fields as the plan, in one line.

## Tests

- `TestFrontMatter`: ok
- `TestPlaceholder`: ok
- `TestReadResult`: ok

## Previous findings

None.

## Findings

### F1. A `## ` line in a code block or comment after `None.` gives `passed`

Level: low
Where: internal/review/review.go:43-45
Breaks: nothing
Why: `ReadResult` ends the section at the first line that starts with
  `## `, and does not know about code blocks or comments. So a section
  of `None.`, then a fenced block whose first line is `## x`, gives
  `passed`, though the section is not only `None.` (plan "Review
  files"). I read the loop; every other case I traced fails safe: a
  question, a comment, or a quoted `## Needs human decision` earlier in
  the file all leave more than `None.` in the body, so they give
  `needs_human`. The template says a section that does not apply gets
  one line, `None.`, and the reviewer deletes every comment, so this
  file already breaks the template. The placeholder never writes it. It
  matters only when real reviewers come (0004).
Fix: skip `## ` lines inside a ``` fence or a comment, or end the
  section only at a `## ` heading of the template. Can wait for 0004.

## Requirements check

None.

## Needs human decision

None.
