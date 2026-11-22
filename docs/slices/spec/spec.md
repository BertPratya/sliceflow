---
slice: spec
created: 2026-11-18
---

# Spec

## Summary

A developer can see where a slice is and what to do next, print the
spec template, check that a spec is in the right shape, and change a
slice's title and description. A spec must be in the right shape
before it is submitted and moves on to the next round.

## Shared rules

### Rule `damaged-line-skipped`: a line that cannot be read does not hide the state

If a line of a slice's log cannot be read, sliceflow skips it with the
warning from slice-and-events ("Warning: line <n> of <log> could not be
read, so it was skipped."), and works out the state from the other
events.
Check: e2e

- **Example:** **Given** a slice `login`, made with `slice create`,
  whose log has a second line `cut off`,
  **When** I run `sliceflow slice status login`,
  **Then** sliceflow prints to stderr:
  ```
  Warning: line 2 of docs/slices/login/.events.jsonl could not be read, so it was skipped.
  ```
  and its State: line is `  State:        spec (draft)`.

### Rule `unknown-event`: an event this version of sliceflow does not know

If a line of a slice's log can be read, but its event is not one this
version of sliceflow knows, sliceflow prints to stderr "Warning: line
<n> of <log> has the event <event>, which this version of sliceflow
does not know, so the state cannot be worked out.", reads no further
events to work out the state, and the state is unknown.
Check: unit, e2e

- **Example:** **Given** a slice `login`, made with `slice create`,
  whose log has a second line that can be read, with the event
  `spec.submitted`,
  **When** I run `sliceflow slice status login`,
  **Then** sliceflow prints to stderr:
  ```
  Warning: line 2 of docs/slices/login/.events.jsonl has the event spec.submitted, which this version of sliceflow does not know, so the state cannot be worked out.
  ```

### Rule `event-not-allowed`: an event that is not allowed in the slice's state

If a line of a slice's log can be read, and sliceflow knows its event,
but the event is not allowed in the state the slice is in at that line,
sliceflow prints to stderr "Warning: line <n> of <log> has the event
<event>, which is not allowed while the slice is in <state>, so the
state cannot be worked out.", reads no further events to work out the
state, and the state is unknown.
Check: unit, e2e

- **Example:** **Given** a slice `login`, made with `slice create`,
  whose log has a second line that can be read, with the event
  `slice.created`,
  **When** I run `sliceflow slice status login`,
  **Then** sliceflow prints to stderr:
  ```
  Warning: line 2 of docs/slices/login/.events.jsonl has the event slice.created, which is not allowed while the slice is in spec (draft), so the state cannot be worked out.
  ```

## Story 1: See where a slice is and what to do next

A developer, or an agent, asks where a slice is in its lifecycle and
what to do next.

Uses: damaged-line-skipped, unknown-event, event-not-allowed

### Rule `state-and-next`: the state, and the next step

When a user runs `sliceflow slice status <name>` and the slice's
worktree is found, sliceflow prints its lines as in slice-and-events,
with the line "  State:        <state>" right after the first line.
After its last line it prints an empty line, then "Next: write the spec
in <path>", where <path> is the spec's full path in the worktree, then
"Then run: sliceflow spec validate <name>", and exits with code 0. A
state is written as its phase, then the state inside the phase in
brackets, e.g. spec (draft). `sliceflow slice status --help` describes
the command as "Show one slice: its state, the next step, its title,
description, start, branch, worktree, and events."
Check: e2e

- **Example:** **Given** a slice `login`, made with
  `sliceflow slice create login --title "Log in" --description "Lets a user log in."`,
  with its worktree,
  **When** I run `sliceflow slice status login`,
  **Then** sliceflow prints:
  ```
  login: Log in
    State:        spec (draft)
    Description:  Lets a user log in.
    Created:      2026-11-18 10:00 UTC
    Started from: main, at a1b2c3d
    Branch:       slice/login
    Worktree:     /home/you/.sliceflow/worktrees/shop/login
    Events:       1 (last: slice.created)

  Next: write the spec in /home/you/.sliceflow/worktrees/shop/login/docs/slices/login/spec.md
  Then run: sliceflow spec validate login
  ```
  and exits with code 0.

### Rule `created-starts-in-draft`: a new slice is in spec (draft)

The event `slice.created` moves a slice into spec (draft). It is
allowed only as the first event of a log.
Check: unit

### Rule `edited-keeps-the-state`: an edit does not move the slice

The event `slice.edited` is allowed in every state, and leaves the
state as it is.
Check: unit

### Rule `no-next-without-worktree`: no next step while the worktree is gone

While the slice's worktree is not found, `slice status` prints no empty
line, and no Next: or Then run: line, after its last line.
Check: e2e

- **Example:** **Given** a slice `login`, made with
  `sliceflow slice create login --title "Log in"`, whose worktree was
  removed,
  **When** I run `sliceflow slice status login`,
  **Then** sliceflow prints:
  ```
  login: Log in
    State:        spec (draft)
    Description:  (none)
    Created:      2026-11-18 10:00 UTC
    Started from: main, at a1b2c3d
    Branch:       slice/login
    Worktree:     (not found)
    Events:       1 (last: slice.created)
    Read from:    branch slice/login
  ```

### Rule `status-of-unknown-state`: status still shows what it knows

If the state of a slice cannot be worked out, `slice status` prints
"  State:        unknown", prints its other lines as usual, prints no
empty line and no Next: or Then run: line after its last line, and
exits with code 0. The Events: line counts every event that could be
read, including the one that made the state unknown.
Check: e2e

- **Example:** **Given** a slice `login`, made with
  `sliceflow slice create login --title "Log in"`, with its worktree,
  whose log has a second line that can be read, with the event
  `spec.submitted`,
  **When** I run `sliceflow slice status login`,
  **Then** sliceflow prints to stderr the warning of `unknown-event`,
  then prints:
  ```
  login: Log in
    State:        unknown
    Description:  (none)
    Created:      2026-11-18 10:00 UTC
    Started from: main, at a1b2c3d
    Branch:       slice/login
    Worktree:     /home/you/.sliceflow/worktrees/shop/login
    Events:       2 (last: spec.submitted)
  ```
  and exits with code 0.

## Story 2: Print the spec template

A developer, or an agent, reads the rules for writing a spec again,
after they were deleted from the finished spec.

Uses: none

### Rule `template-printed`: the built-in template, as it is stored

When a user runs `sliceflow spec template`, sliceflow prints to stdout
the spec template built into this version of sliceflow, exactly as it
is stored, with its comments and its `{{...}}` placeholders, including
`{{slice}}` and `{{date}}`, and exits with code 0.
Check: e2e

- **Example:** **When** I run `sliceflow spec template`,
  **Then** sliceflow prints the whole template, which starts with:
  ```
  ---
  slice: {{slice}}
  created: {{date}}
  ---
  ```
  and exits with code 0.

### Rule `template-in-any-folder`: no project is needed

`sliceflow spec template` needs no project, no slice, and no git. It
works the same in a folder that is not inside a git repository.
Check: e2e

- **Example:** **Given** I am in a folder that is not inside a git
  repository,
  **When** I run `sliceflow spec template`,
  **Then** sliceflow prints the whole template, and exits with code 0.

### Rule `spec-help`: the spec commands in the help

`sliceflow --help` lists the command group `spec` with the line
"  spec   Print the spec template, and find problems in specs".
`sliceflow spec --help` lists the commands `template` and `validate`,
in the same layout as `sliceflow slice --help`.
Check: e2e

- **Example:** **When** I run `sliceflow spec --help`,
  **Then** sliceflow prints:
  ```
  Print the spec template, and find problems in specs.

  Usage:
    sliceflow spec <command> [options]

  Commands:
    template         Print the spec template
    validate <name>  Find the problems in a spec

  Options:
    -h, --help  Show this help.
  ```
  and exits with code 0.

## Story 3: Check that a spec is in the right shape

A developer, or an agent, finds the problems in a slice's spec that
would stop the later steps of sliceflow, before the spec is submitted.

Uses: none

### Rule `spec-has-no-problems`: a spec with no problems

When a user runs `sliceflow spec validate <name>` and the spec in the
slice's worktree, `docs/slices/<name>/spec.md`, has no problems,
sliceflow prints to stdout "<full path of the spec>: no problems
found." and exits with code 0. It reads the file as it is now,
including changes that are not committed yet.
Check: e2e

- **Example:** **Given** a slice `login` whose spec in its worktree has
  no problems,
  **When** I run `sliceflow spec validate login`,
  **Then** sliceflow prints to stdout:
  ```
  /home/you/.sliceflow/worktrees/shop/login/docs/slices/login/spec.md: no problems found.
  ```
  and exits with code 0.

### Rule `problems-listed`: each problem with its line and a hint

When the spec has one or more problems, sliceflow prints to stdout, for
each problem, the line "<full path>:<line>: error: <what is wrong>",
then a hint on how to fix it, indented by two spaces. The problems are
in the order of their line numbers. Then it prints "1 problem found."
or "<n> problems found.", and exits with code 1. It reports every
problem it finds, not only the first.
Check: e2e

- **Example:** **Given** a slice `login` whose spec has a shared rule
  `unknown-event` on line 12 that no story names in its `Uses:` line,
  and a rule `edit-empty-values` on line 40 with no `Check:` line,
  **When** I run `sliceflow spec validate login`,
  **Then** sliceflow prints to stdout:
  ```
  /home/you/.sliceflow/worktrees/shop/login/docs/slices/login/spec.md:12: error: shared rule unknown-event is not used by any story.
    Name it in a story's Uses: line, or delete it.
  /home/you/.sliceflow/worktrees/shop/login/docs/slices/login/spec.md:40: error: rule edit-empty-values has no Check: line.
    Add a line like "Check: e2e" under the rule's heading.
  2 problems found.
  ```
  and exits with code 1.
- **Example:** **Given** a slice `login` whose spec has only the second
  of those problems,
  **When** I run `sliceflow spec validate login`,
  **Then** the last line sliceflow prints is `1 problem found.`, and it
  exits with code 1.

### Rule `spec-needs-a-story`: a spec has at least one story

If the spec has no story, `spec validate` reports on line 1 "the spec
has no story." with the hint "Add one as \"## Story 1: <title>\", with
at least one rule under it.". No other section is required.
Check: unit

- **Example:** **Given** a spec with only a Summary,
  **When** I run `sliceflow spec validate login`,
  **Then** sliceflow reports on line 1:
  ```
  error: the spec has no story.
    Add one as "## Story 1: <title>", with at least one rule under it.
  ```
- **Example:** **Given** a spec with one story with one rule, and no
  other section,
  **When** I run `sliceflow spec validate login`,
  **Then** sliceflow reports no problem.

### Rule `story-heading-shape`: a story heading is `## Story <n>: <title>`

If a `##` heading starts with `Story` but is not in the shape
`## Story <n>: <title>`, where `<n>` is a number and the title is not
empty, `spec validate` reports on the heading's line "the heading
\"<heading>\" is not in the shape \"## Story <n>: <title>\"." with the
hint "Write it as \"## Story <n>: <title>\", e.g. \"## Story 2: Print
the template\".".
Check: unit

- **Example:** **Given** a spec whose line 30 is `## Story two: Print`,
  **When** I run `sliceflow spec validate login`,
  **Then** sliceflow reports on line 30:
  ```
  error: the heading "## Story two: Print" is not in the shape "## Story <n>: <title>".
    Write it as "## Story <n>: <title>", e.g. "## Story 2: Print the template".
  ```

### Rule `stories-numbered-in-order`: stories are numbered 1, 2, 3, ...

If a story's number is not the one after the story before it (1 for
the first story), `spec validate` reports on its heading's line
"Story <n> should be Story <m>: stories are numbered 1, 2, 3, ... in
order." with the hint "Renumber the stories from 1, in the order they
appear.".
Check: unit

- **Example:** **Given** a spec whose stories are Story 1 and, on line
  80, Story 3,
  **When** I run `sliceflow spec validate login`,
  **Then** sliceflow reports on line 80:
  ```
  error: Story 3 should be Story 2: stories are numbered 1, 2, 3, ... in order.
    Renumber the stories from 1, in the order they appear.
  ```
- **Example:** **Given** a spec whose stories are Story 1, Story 2, and,
  on line 120, Story 2 again,
  **When** I run `sliceflow spec validate login`,
  **Then** sliceflow reports on line 120 "Story 2 should be Story 3:
  stories are numbered 1, 2, 3, ... in order.", with the same hint.

### Rule `story-needs-a-rule`: a story has a rule of its own

If a story has no rule of its own, `spec validate` reports on its
heading's line "Story <n> has no rules." with the hint "Add at least
one rule under it, as \"### Rule `<name>`: <one line>\".". Shared
rules named in its `Uses:` line do not count.
Check: unit

- **Example:** **Given** a spec whose Story 2, on line 42, has the line
  `Uses: unknown-event` and no rule under it,
  **When** I run `sliceflow spec validate login`,
  **Then** sliceflow reports on line 42:
  ```
  error: Story 2 has no rules.
    Add at least one rule under it, as "### Rule `<name>`: <one line>".
  ```

### Rule `story-needs-uses`: every story has a `Uses:` line

If a story has no `Uses:` line, `spec validate` reports on its
heading's line "Story <n> has no Uses: line." with the hint "Add
\"Uses: none\", or name the shared rules it needs, e.g. \"Uses:
unknown-event\".".
Check: unit

- **Example:** **Given** a spec whose Story 1, on line 20, has no
  `Uses:` line,
  **When** I run `sliceflow spec validate login`,
  **Then** sliceflow reports on line 20:
  ```
  error: Story 1 has no Uses: line.
    Add "Uses: none", or name the shared rules it needs, e.g. "Uses: unknown-event".
  ```

### Rule `uses-names-shared-rules`: `Uses:` names only shared rules

`Uses:` is `none`, or rule names separated by commas. If a name in it
is not a rule in Shared rules, `spec validate` reports on the `Uses:`
line "Story <n> uses <name>, which is not a shared rule." with the hint
"Fix the name, or move the rule to Shared rules.".
Check: unit

- **Example:** **Given** a spec whose Story 1 has, on line 24, the line
  `Uses: unknwon-event`, and whose Shared rules has `unknown-event`,
  **When** I run `sliceflow spec validate login`,
  **Then** sliceflow reports on line 24:
  ```
  error: Story 1 uses unknwon-event, which is not a shared rule.
    Fix the name, or move the rule to Shared rules.
  ```

### Rule `shared-rule-used`: every shared rule is used by a story

If no story names a shared rule in its `Uses:` line, `spec validate`
reports on the rule's heading line "shared rule <name> is not used by
any story." with the hint "Name it in a story's Uses: line, or delete
it.".
Check: unit

- **Example:** **Given** a spec whose shared rule `unknown-event`, on
  line 12, is named in no `Uses:` line,
  **When** I run `sliceflow spec validate login`,
  **Then** sliceflow reports on line 12:
  ```
  error: shared rule unknown-event is not used by any story.
    Name it in a story's Uses: line, or delete it.
  ```

### Rule `rule-in-unknown-section`: a rule belongs to a story or Shared rules

A spec may have sections that are not in the template. If a rule is in
such a section, `spec validate` reports on the rule's heading line
"rule <name> is in the section \"<heading>\", which is not a story or
Shared rules." with the hint "If the section is a story, write its
heading as \"## Story <n>: <title>\". If not, move the rule into a
story or Shared rules.".
Check: unit

- **Example:** **Given** a spec with a section `## Stroy 2: Print`, and
  in it, on line 50, the rule `template-printed`,
  **When** I run `sliceflow spec validate login`,
  **Then** sliceflow reports on line 50:
  ```
  error: rule template-printed is in the section "Stroy 2: Print", which is not a story or Shared rules.
    If the section is a story, write its heading as "## Story <n>: <title>". If not, move the rule into a story or Shared rules.
  ```
- **Example:** **Given** a spec with a section `## Notes` that holds no
  rule,
  **When** I run `sliceflow spec validate login`,
  **Then** sliceflow reports no problem for it.

### Rule `rule-heading-shape`: a rule heading is ``### Rule `<name>`: <one line>``

If a `###` heading starts with `Rule` but is not in the shape
``### Rule `<name>`: <one line>``, with the name in backticks and a
line after the colon that is not empty, `spec validate` reports on the
heading's line "the heading \"<heading>\" is not in the shape \"###
Rule `<name>`: <one line>\"." with the hint "Write it as \"### Rule
`<name>`: <one line>\", e.g. \"### Rule `title-required`: a slice
needs a title\".".
Check: unit

- **Example:** **Given** a spec whose line 33 is
  `### Rule title-required: a slice needs a title`,
  **When** I run `sliceflow spec validate login`,
  **Then** sliceflow reports on line 33:
  ```
  error: the heading "### Rule title-required: a slice needs a title" is not in the shape "### Rule `<name>`: <one line>".
    Write it as "### Rule `<name>`: <one line>", e.g. "### Rule `title-required`: a slice needs a title".
  ```

### Rule `rule-name-shape`: a rule name is lowercase words joined by hyphens

A rule name is made of words of lowercase letters (a to z) and digits
(0 to 9), joined by single hyphens, with no hyphen at the start or the
end. If a name is not, `spec validate` reports on the heading's line
"rule name \"<name>\" is not lowercase words joined by hyphens." with
the hint "Use only a to z, 0 to 9, and single hyphens, e.g.
title-required.".
Check: unit

- **Example:** **Given** a spec whose line 33 is
  ``### Rule `Title Required`: a slice needs a title``,
  **When** I run `sliceflow spec validate login`,
  **Then** sliceflow reports on line 33:
  ```
  error: rule name "Title Required" is not lowercase words joined by hyphens.
    Use only a to z, 0 to 9, and single hyphens, e.g. title-required.
  ```
- **Example:** **Given** a rule named `utf8-text`,
  **When** I run `sliceflow spec validate login`,
  **Then** sliceflow reports no problem for its name.

### Rule `rule-name-unique`: a rule name is used once

If two rules have the same name, shared rules included, `spec validate`
reports on the second one's heading line "rule <name> is already used
on line <n>." with the hint "Give each rule its own name.".
Check: unit

- **Example:** **Given** a spec with the shared rule `unknown-event` on
  line 12, and a rule `unknown-event` in Story 1 on line 40,
  **When** I run `sliceflow spec validate login`,
  **Then** sliceflow reports on line 40:
  ```
  error: rule unknown-event is already used on line 12.
    Give each rule its own name.
  ```

### Rule `rule-has-one-check`: every rule has one `Check:` line

If a rule has no `Check:` line, `spec validate` reports on the rule's
heading line "rule <name> has no Check: line." with the hint "Add a
line like \"Check: e2e\" under the rule's heading.". A `Check:` line
with no value counts as no `Check:` line. If a rule has more than one,
it reports on the second one "rule <name> has more than one Check:
line." with the hint "Keep one Check: line, with every value on it,
e.g. \"Check: unit, e2e\".".
Check: unit

- **Example:** **Given** a spec whose rule `edit-empty-values`, on line
  88, has no `Check:` line,
  **When** I run `sliceflow spec validate login`,
  **Then** sliceflow reports on line 88:
  ```
  error: rule edit-empty-values has no Check: line.
    Add a line like "Check: e2e" under the rule's heading.
  ```
- **Example:** **Given** a spec whose rule `edit-empty-values` has the
  line `Check: unit`, and on line 92 the line `Check: e2e`,
  **When** I run `sliceflow spec validate login`,
  **Then** sliceflow reports on line 92:
  ```
  error: rule edit-empty-values has more than one Check: line.
    Keep one Check: line, with every value on it, e.g. "Check: unit, e2e".
  ```

### Rule `check-values`: a `Check:` value is unit, api, e2e, or manual

A `Check:` line holds one or more values, separated by commas. If a
value is not `unit`, `api`, `e2e`, or `manual`, `spec validate` reports
on the `Check:` line "rule <name> has the Check: value <value>, which
is not unit, api, e2e, or manual." with the hint "Fix the value.".
Check: unit

- **Example:** **Given** a spec whose rule `nothing-to-change` has, on
  line 95, the line `Check: unit, manaul`,
  **When** I run `sliceflow spec validate login`,
  **Then** sliceflow reports on line 95:
  ```
  error: rule nothing-to-change has the Check: value manaul, which is not unit, api, e2e, or manual.
    Fix the value.
  ```
- **Example:** **Given** a rule with the line `Check: unit, e2e`,
  **When** I run `sliceflow spec validate login`,
  **Then** sliceflow reports no problem for that line.

### Rule `rule-needs-an-example`: which rules need an example

An example is a list item that starts with `- **Example:**`, under a
rule. If a rule has no example, and its `Check:` line has a value other
than `unit`, `spec validate` reports on the rule's heading line "rule
<name> has no example." with the hint "Add one under its Check: line,
as \"- **Example:** **When** ... **Then** ...\".". If the first rule of
a story has no example, it reports on the rule's heading line "rule
<name> is the first rule of Story <n>, so it needs an example." with
the same hint, whatever its `Check:` line says.
Check: unit

- **Example:** **Given** a spec whose rule `nothing-to-change`, on line
  95, has `Check: e2e` and no example,
  **When** I run `sliceflow spec validate login`,
  **Then** sliceflow reports on line 95:
  ```
  error: rule nothing-to-change has no example.
    Add one under its Check: line, as "- **Example:** **When** ... **Then** ...".
  ```
- **Example:** **Given** a rule that is not the first of its story, with
  `Check: unit` and no example,
  **When** I run `sliceflow spec validate login`,
  **Then** sliceflow reports no problem for it.
- **Example:** **Given** a spec whose Story 2 has as its first rule, on
  line 60, `state-replayed`, with `Check: unit` and no example,
  **When** I run `sliceflow spec validate login`,
  **Then** sliceflow reports on line 60:
  ```
  error: rule state-replayed is the first rule of Story 2, so it needs an example.
    Add one under its Check: line, as "- **Example:** **When** ... **Then** ...".
  ```

### Rule `example-has-when-then`: an example has When and Then

If an example has no `**When**`, or no `**Then**`, `spec validate`
reports on the example's first line "an example of rule <name> has no
**When**." with the hint "Say what is done, starting with **When**.",
or "an example of rule <name> has no **Then**." with the hint "Say what
happens, starting with **Then**.". `**Given**` is not needed.
Check: unit

- **Example:** **Given** a spec whose rule `template-printed` has, on
  line 374, an example with **When** and no **Then**,
  **When** I run `sliceflow spec validate login`,
  **Then** sliceflow reports on line 374:
  ```
  error: an example of rule template-printed has no **Then**.
    Say what happens, starting with **Then**.
  ```
- **Example:** **Given** an example with **When** and **Then**, and no
  **Given**,
  **When** I run `sliceflow spec validate login`,
  **Then** sliceflow reports no problem for it.

### Rule `placeholder-left`: no placeholder from the template is left

If `{{` is outside backticks, `spec validate` reports on its line "a
placeholder from the template is still here: <placeholder>." with the
hint "Write the real text in its place. If you mean the characters
themselves, put them in backticks.". The placeholder is the text from
`{{` to the next `}}` on the same line, or to the end of the line. It
reports each placeholder once.
Check: unit

- **Example:** **Given** a spec whose line 30 is
  `## Story 1: {{short title}}`,
  **When** I run `sliceflow spec validate login`,
  **Then** sliceflow reports on line 30:
  ```
  error: a placeholder from the template is still here: {{short title}}.
    Write the real text in its place. If you mean the characters themselves, put them in backticks.
  ```
- **Example:** **Given** a rule whose text says
  ``including `{{slice}}` and `{{date}}` ``,
  **When** I run `sliceflow spec validate login`,
  **Then** sliceflow reports no problem for it.

### Rule `comment-left`: no HTML comment is left

If an HTML comment (`<!--` to `-->`) is outside backticks,
`spec validate` reports on the line where it starts "an HTML comment
is still here. It is hidden when the spec is shown, but an agent reads
it." with the hint "Delete it. Write what the reader needs as plain
text.". It reports each comment once, and does not report a
placeholder inside it.
Check: unit

- **Example:** **Given** a spec with the template's comment under
  `## Open questions`, from line 120 to line 124, with `{{` inside it,
  **When** I run `sliceflow spec validate login`,
  **Then** sliceflow reports one problem, on line 120:
  ```
  error: an HTML comment is still here. It is hidden when the spec is shown, but an agent reads it.
    Delete it. Write what the reader needs as plain text.
  ```

### Rule `open-questions-left`: no open question is left

A question is a list item that starts with `- **Q<n>.**` in the Open
questions section. If the section has one or more questions,
`spec validate` reports on the section's heading line "<n> questions
are still open." ("1 question is still open." for one) with the hint
"Answer each one in the spec, then remove it. When none are left, write
None.". If the section holds text that is neither `None` nor a
question, it reports on the same line "Open questions must say None, or
list questions as \"- **Q1.** ...\"." with the hint "Write each open
question as a Q item, or write None.". A spec with no Open questions
section has no such problem.
Check: unit

- **Example:** **Given** a spec whose `## Open questions`, on line 118,
  lists Q1, Q2, and Q3,
  **When** I run `sliceflow spec validate login`,
  **Then** sliceflow reports on line 118:
  ```
  error: 3 questions are still open.
    Answer each one in the spec, then remove it. When none are left, write None.
  ```
- **Example:** **Given** a spec whose `## Open questions`, on line 118,
  holds only the line `Ask about the title length.`,
  **When** I run `sliceflow spec validate login`,
  **Then** sliceflow reports on line 118:
  ```
  error: Open questions must say None, or list questions as "- **Q1.** ...".
    Write each open question as a Q item, or write None.
  ```
- **Example:** **Given** a spec whose Open questions says `None`,
  **When** I run `sliceflow spec validate login`,
  **Then** sliceflow reports no problem for it.

### Rule `decisions-numbered-in-order`: decisions are numbered SD1, SD2, SD3, ...

A decision is a list item that starts with `- **SD<n>.**` in the
Decisions section. If a decision's number is not the one after the
decision before it (SD1 for the first), `spec validate` reports on its
line "SD<n> should be SD<m>: decisions are numbered SD1, SD2, SD3, ...
in order." with the hint "Renumber the decisions from SD1, in the order
they appear.".
Check: unit

- **Example:** **Given** a spec whose decisions are SD1, SD2, and, on
  line 700, SD4,
  **When** I run `sliceflow spec validate login`,
  **Then** sliceflow reports on line 700:
  ```
  error: SD4 should be SD3: decisions are numbered SD1, SD2, SD3, ... in order.
    Renumber the decisions from SD1, in the order they appear.
  ```

### Rule `decision-needs-rule-line`: every decision has a `Rule:` line

If a decision has no `Rule:` line, `spec validate` reports on the
decision's first line "SD<n> has no Rule: line." with the hint "Add
\"Rule: <rule names>\", or \"Rule: none\".".
Check: unit

- **Example:** **Given** a spec whose SD5, on line 760, has no `Rule:`
  line,
  **When** I run `sliceflow spec validate login`,
  **Then** sliceflow reports on line 760:
  ```
  error: SD5 has no Rule: line.
    Add "Rule: <rule names>", or "Rule: none".
  ```

### Rule `decision-rule-names-exist`: `Rule:` names rules in this spec

`Rule:` is `none`, or rule names separated by commas. If a name in it
is not a rule in this spec, `spec validate` reports on the `Rule:` line
"SD<n> names rule <name>, which is not in this spec." with the hint
"Fix the name, or write none.".
Check: unit

- **Example:** **Given** a spec whose SD5 has, on line 790, the line
  `Rule: edit-needs-worktree`, and whose rule is named
  `edit-needs-the-worktree`,
  **When** I run `sliceflow spec validate login`,
  **Then** sliceflow reports on line 790:
  ```
  error: SD5 names rule edit-needs-worktree, which is not in this spec.
    Fix the name, or write none.
  ```

### Rule `code-blocks-skipped`: no check looks inside a code block

Every check of `spec validate` skips the lines inside a code block:
from a line that starts with ```, after any spaces, to the next such
line. A code block inside a list item, e.g. in an example, is indented,
so the spaces are allowed.
Check: unit

- **Example:** **Given** a rule whose example holds a code block with
  the lines `## Story 1: {{short title}}` and `<!-- note -->`,
  **When** I run `sliceflow spec validate login`,
  **Then** sliceflow reports no problem for those lines.

### Rule `validate-errors-as-before`: the same mistakes, the same messages

`spec validate` reports a missing name, an invalid name, and a name
with no slice exactly as `slice status` does, to stderr (exit codes 2,
2, and 1).
Check: e2e

- **Example:** **When** I run `sliceflow spec validate a.b`,
  **Then** sliceflow prints to stderr:
  ```
  Error: invalid slice name "a.b". Nothing was changed.
  Allowed characters: letters, digits, hyphens, underscores. Start with a letter or a digit.
  ```
  and exits with code 2.

### Rule `validate-needs-the-worktree`: no worktree, no spec to read

If the slice's worktree is not found, `spec validate` prints to stderr
"Error: the worktree of slice <name> was not found, so its spec cannot
be validated. Nothing was changed.", and exits with code 1. It never
reads the spec from the slice's branch instead.
Check: e2e

- **Example:** **Given** a slice `login` whose worktree was removed,
  **When** I run `sliceflow spec validate login`,
  **Then** sliceflow prints to stderr:
  ```
  Error: the worktree of slice login was not found, so its spec cannot be validated. Nothing was changed.
  ```
  and exits with code 1.

### Rule `validate-needs-the-spec`: no spec file, nothing to check

If the worktree has no file `docs/slices/<name>/spec.md`,
`spec validate` prints to stderr "Error: the spec of slice <name> was
not found at <full path>. Nothing was changed.", then "To start from
the template, run: sliceflow spec template", and exits with code 1.
Check: e2e

- **Example:** **Given** a slice `login` whose spec was deleted from its
  worktree,
  **When** I run `sliceflow spec validate login`,
  **Then** sliceflow prints to stderr:
  ```
  Error: the spec of slice login was not found at /home/you/.sliceflow/worktrees/shop/login/docs/slices/login/spec.md. Nothing was changed.
  To start from the template, run: sliceflow spec template
  ```
  and exits with code 1.

### Rule `validate-in-any-state`: the state of the slice does not matter

`spec validate` works in every state of the slice, including an unknown
one. It does not work out the state, so it prints no warning about the
log.
Check: e2e

- **Example:** **Given** a slice `login` whose log has a second line
  with the event `spec.submitted`, which this version does not know,
  and whose spec has no problems,
  **When** I run `sliceflow spec validate login`,
  **Then** sliceflow prints only:
  ```
  /home/you/.sliceflow/worktrees/shop/login/docs/slices/login/spec.md: no problems found.
  ```
  and exits with code 0.

### Rule `validate-help`: the help of spec validate

`sliceflow spec validate --help` prints the help of the command, in the
same layout as `sliceflow slice status --help`, and exits with code 0.
Check: e2e

- **Example:** **When** I run `sliceflow spec validate --help`,
  **Then** sliceflow prints:
  ```
  Find the problems in a slice's spec, each with its line and a hint on how to fix it.

  Usage:
    sliceflow spec validate <name> [options]

  Arguments:
    <name>  The name of the slice whose spec to read.

  Options:
    -h, --help  Show this help.
  ```
  and exits with code 0.

## Story 4: Change a slice's title and description

A developer, or an agent, changes a slice's title or description after
the slice was made.

Uses: damaged-line-skipped, unknown-event, event-not-allowed

### Rule `edit-changes-the-slice`: the title, the description, or both are changed

When a user runs `sliceflow slice edit <name>` with `--title`,
`--description`, or both, and a value differs from the slice's current
one, sliceflow adds one event to the log in the slice's worktree, makes
no commit, prints "Edited slice <name>.", then the lines
"  Title:        <title>" and "  Description:  <description>" with the
values the slice has now (the description as (none) when it has none),
and exits with code 0. A slice's current title and description are the
ones from its `slice.created` event, changed by each `slice.edited`
event after it, in order.
Check: e2e

- **Example:** **Given** a slice `login` with the title "Log in" and the
  description "Lets a user log in.",
  **When** I run `sliceflow slice edit login --title "Sign in"`,
  **Then** sliceflow prints:
  ```
  Edited slice login.
    Title:        Sign in
    Description:  Lets a user log in.
  ```
  and exits with code 0.
- **Example:** **Given** the same slice,
  **When** I run
  `sliceflow slice edit login --title "Sign in" --description "Lets a user sign in with a password."`,
  **Then** sliceflow prints:
  ```
  Edited slice login.
    Title:        Sign in
    Description:  Lets a user sign in with a password.
  ```

### Rule `edited-event`: the event holds only what changed

The event `slice edit` adds is `slice.edited`. Its data holds `title`
when the title differs from the current one, and `description` when
the description differs, and nothing else.
Check: e2e

- **Example:** **Given** a slice `login` with the title "Log in" and the
  description "Lets a user log in.",
  **When** I run
  `sliceflow slice edit login --title "Sign in" --description "Lets a user log in."`,
  **Then** the last line of its log has `"event":"slice.edited"` and
  `"data":{"title":"Sign in"}`.

### Rule `edit-empty-values`: an empty title or description

When `--title` is empty, the new title is the slice's name. When
`--description` is empty, the slice has no description.
Check: e2e

- **Example:** **Given** a slice `login` with the title "Sign in",
  **When** I run `sliceflow slice edit login --title ""`,
  **Then** sliceflow prints the line `  Title:        login`, and the
  event's data is `{"title":"login"}`.
- **Example:** **Given** a slice `login` with the description "Lets a
  user log in.",
  **When** I run `sliceflow slice edit login --description ""`,
  **Then** sliceflow prints the line `  Description:  (none)`, and the
  event's data is `{"description":""}`.

### Rule `nothing-to-change`: the same values add no event

If no value given to `slice edit` differs from the slice's current one,
after `edit-empty-values`, sliceflow adds no event, prints "Nothing to
change: slice <name> already has this title and description.", and
exits with code 0.
Check: e2e

- **Example:** **Given** a slice `login` with the title "Sign in",
  **When** I run `sliceflow slice edit login --title "Sign in"`,
  **Then** sliceflow prints:
  ```
  Nothing to change: slice login already has this title and description.
  ```
  adds no line to the log, and exits with code 0.

### Rule `edit-needs-an-option`: nothing to edit

If `slice edit` gets neither `--title` nor `--description`, sliceflow
changes nothing, prints "Error: slice edit needs --title,
--description, or both. Nothing was changed.", then "To see the
options, run: sliceflow slice edit --help", and exits with code 2.
Check: e2e

- **Example:** **When** I run `sliceflow slice edit login`,
  **Then** sliceflow prints:
  ```
  Error: slice edit needs --title, --description, or both. Nothing was changed.
  To see the options, run: sliceflow slice edit --help
  ```
  and exits with code 2.

### Rule `edit-needs-the-worktree`: no worktree, no edit

If the slice's worktree is not found, `slice edit` changes nothing,
prints "Error: the worktree of slice <name> was not found, so it cannot
be edited. Nothing was changed.", and exits with code 1.
Check: e2e

- **Example:** **Given** a slice `login` whose worktree was removed,
  **When** I run `sliceflow slice edit login --title "Sign in"`,
  **Then** sliceflow prints:
  ```
  Error: the worktree of slice login was not found, so it cannot be edited. Nothing was changed.
  ```
  and exits with code 1.

### Rule `no-edit-without-state`: an unknown state blocks an edit

If the state of the slice cannot be worked out, `slice edit` changes
nothing, prints after the warning "Error: the state of slice <name>
cannot be worked out, so it cannot be changed. Nothing was changed.",
then a hint, and exits with code 1. When the state is unknown because
of an unknown event, the hint is "Update sliceflow, then try again."
When it is unknown because of an event that is not allowed, the hint
is "To see how the log changed, run: git log -p -- <log>".
Check: e2e

- **Example:** **Given** a slice `login` whose log has a second line
  that can be read, with the event `spec.submitted`,
  **When** I run `sliceflow slice edit login --title "Sign in"`,
  **Then** sliceflow prints the warning of `unknown-event`, then:
  ```
  Error: the state of slice login cannot be worked out, so it cannot be changed. Nothing was changed.
  Update sliceflow, then try again.
  ```
  adds no line to the log, and exits with code 1.
- **Example:** **Given** a slice `login` whose log has a second line
  that can be read, with the event `slice.created`,
  **When** I run `sliceflow slice edit login --title "Sign in"`,
  **Then** sliceflow prints the warning of `event-not-allowed`, then:
  ```
  Error: the state of slice login cannot be worked out, so it cannot be changed. Nothing was changed.
  To see how the log changed, run: git log -p -- docs/slices/login/.events.jsonl
  ```
  adds no line to the log, and exits with code 1.

### Rule `edit-errors-as-before`: the same mistakes, the same messages

`slice edit` reports a title or description with a control character
exactly as `slice create` does (exit code 2), and an invalid name or a
name with no slice exactly as `slice status` does (exit codes 2 and 1).
Check: e2e

- **Example:** **When** I run `sliceflow slice edit a.b --title "x"`,
  **Then** sliceflow prints:
  ```
  Error: invalid slice name "a.b". Nothing was changed.
  Allowed characters: letters, digits, hyphens, underscores. Start with a letter or a digit.
  ```
  and exits with code 2.

### Rule `list-shows-edits`: slice list shows the current title and description

`slice list` shows each slice's current title and description, after
every `slice.edited` event in its log.
Check: e2e

- **Example:** **Given** a slice `login` made with the title "Log in",
  then edited with `--title "Sign in"`,
  **When** I run `sliceflow slice list`,
  **Then** sliceflow prints:
  ```
  NAME   TITLE    DESCRIPTION
  login  Sign in
  ```

### Rule `slice-help-lists-edit`: edit in the help

`sliceflow --help` describes the group `slice` as "Make, see, and
change slices". `sliceflow slice --help` lists
"  edit <name>    Change a slice's title or description" between
`create` and `list`.
Check: e2e

- **Example:** **When** I run `sliceflow slice --help`,
  **Then** its Commands part is:
  ```
  Commands:
    create <name>  Make a new slice
    edit <name>    Change a slice's title or description
    list           Show all slices
    status <name>  Show one slice
  ```

## Out of scope

- Submitting a spec, and the spec review: 0003.
- Approval: 0004.
- How `spec validate` treats a rule or decision marked removed after the first approval.
- Warnings in `spec validate`, e.g. for a rule with only one example.

## Decisions

- **SD1.** One command, `sliceflow slice edit <name>`, changes a slice's
  title, its description, or both, with the options `--title` and
  `--description`.
  Why: one command for one job, and the options mean the same as in
    `slice create` (design 1.18). GitHub's CLI works the same way, e.g.
    `gh repo edit --description`.
  Not chosen: two commands, `slice title` and `slice description`, as
    design 1.7 sketched them: two commands for one job.
  Cost: `slice edit login --title "Log in"` is longer to type than
    `slice title login "Log in"`.
  Rule: edit-changes-the-slice, edited-event, slice-help-lists-edit

- **SD2.** `slice edit` writes one event, `slice.edited`, whose data
  holds only the fields the user changed: the title, the description,
  or both.
  Why: one command writes one event, so a change is never half done.
    The log shows exactly what was changed. A reader starts from
    `slice.created`, and each `slice.edited` replaces the fields it
    holds.
  Not chosen: one event per field (`slice.titled`, `slice.described`),
    because changing both writes two events, and a failure between them
    leaves the change half done; one event with both fields every time,
    because it hides which field was changed.
  Cost: to know the current title, a reader must go through every
    `slice.edited` event, not only the last one.
  Rule: edited-event

- **SD3.** In `slice edit`, an empty `--title` sets the title back to
  the slice name, and an empty `--description` removes the description,
  so `slice status` shows it as (none).
  Why: an option means the same in every command (design 1.18), and in
    `slice create` an empty title already means the slice name, and an
    empty description means none. It also gives a way to undo a custom
    title, or to remove a description.
  Not chosen: refusing an empty value as an error, because then a
    description could never be removed once set.
  Cost: a title set to "" by mistake silently becomes the slice name.
  Rule: edit-empty-values

- **SD4.** `slice edit` compares each value it is given, after SD3,
  with the slice's current value: the one from `slice.created`, changed
  by each `slice.edited` after it, in order. The event holds only the
  values that differ. When none differ, it writes no event, prints
  "Nothing to change: slice <name> already has this title and
  description.", and exits with code 0.
  Why: the log records changes, and an event that changes nothing is
    noise. The user still got what they asked for, so it is not an
    error.
  Not chosen: writing the event anyway, because the log would fill with
    events that change nothing.
  Cost: a script cannot tell from the exit code whether an event was
    written.
  Rule: nothing-to-change

- **SD5.** `slice edit` adds its event to the log in the slice's
  worktree, and makes no commit. If the worktree is not found, it
  changes nothing, prints "Error: the worktree of slice <name> was not
  found, so it cannot be edited. Nothing was changed.", and exits with
  code 1.
  Why: design 1.13 writes each event to the file at once, and commits
    the log only when a step ends. Without a worktree there is no file
    to write to, and changing the branch would need a commit.
  Not chosen: writing the event to the branch with a commit when the
    worktree is gone, because `slice edit` would then commit in one
    case and not in the other.
  Cost: a slice whose worktree was removed cannot be edited until the
    worktree is back.
  Rule: edit-needs-the-worktree

- **SD6.** `slice list` reads each slice's whole log, and shows the
  current title and description: the ones from `slice.created`, changed
  by each `slice.edited` after it.
  Why: after `slice edit`, `slice status` shows the new title, so
    `slice list` must too. Two commands with two titles for one slice
    would look like a bug.
  Not chosen: reading only the first line of each log, as in
    slice-and-events, because `list` would keep showing the title from
    when the slice was made.
  Cost: `list` reads more of each log. Logs are small, so this is
    cheap.
  Rule: list-shows-edits
  Replaces: slice-and-events/SD16 (only its first part: a damaged line
    after the first still gives no warning in `list`)

- **SD7.** While the worktree is not found, `slice status` prints no
  next step.
  Why: the spec cannot be written in a worktree that does not exist,
    and the Worktree: (not found) line already shows the problem.
  Not chosen: a next step that says how to bring the worktree back with
    git, because sliceflow would tell the user to run a git command it
    does not manage. Bringing a worktree back is a feature of its own,
    in the backlog.
  Cost: the user must find out alone how to get the worktree back.
  Rule: no-next-without-worktree

- **SD8.** A line of the log that cannot be read is skipped with the
  warning from slice-and-events, and the state is worked out from the
  other events.
  Why: such a line almost always comes from a crash while an event was
    written, so the command that wrote it failed and the change never
    happened. sliceflow never repairs a line (slice-and-events/SD20),
    so if a damaged line made the state unknown, one crash would block
    the slice forever.
  Not chosen: an unknown state for any damaged line, because one crash
    would block the slice forever.
  Cost: if someone breaks by hand a line that was a real change, the
    state shown is wrong. The warning still names the line.
  Rule: damaged-line-skipped

- **SD9.** If a line of the log can be read, but its event is unknown,
  or not allowed in the state the slice is in, sliceflow prints a
  warning, reads no further events for the state, and the state is
  unknown. `slice status` then shows State: unknown and no next step,
  still prints its other lines, and exits with code 0. This works the
  same in every command that reads the state.
  Why: skipping the event could show a wrong state, e.g. spec (draft)
    for a slice that a newer sliceflow already moved on. Later commands
    decide what is allowed from the state, so a wrong state could let a
    wrong move through.
  Not chosen: skipping the event, because the state could be wrong;
    stopping `status` with an error, because it would show nothing just
    when it is needed most.
  Cost: one unknown event hides the state of the whole slice.
  Rule: unknown-event, event-not-allowed, status-of-unknown-state

- **SD10.** While the state is unknown, `slice edit` changes nothing
  and exits with code 1. Commands that only read still work.
  Why: the engine checks that each change is allowed in the current
    state. With no state, it cannot check, and a change that should
    never happen could enter the log.
  Not chosen: allowing harmless changes such as a new title, because a
    newer sliceflow may have a state where even that is not allowed,
    and an older one cannot know it.
  Cost: in a team where people update at different times, an older
    sliceflow can read a slice but not change it.
  Rule: no-edit-without-state, validate-in-any-state

- **SD11.** `spec validate` reads the spec in the slice's worktree, and
  prints its full path.
  Why: the worktree is where the spec is edited, so the check sees the
  latest changes, even ones not committed yet. A full path can be
  clicked in an editor or a terminal, from any folder.
  Not chosen: reading the last committed spec on the slice's branch,
  because it misses the changes the person just made; falling back to
  it when the worktree is gone, because "no problems found" for an old
  spec looks like the latest edits passed.
  Cost: with no worktree, there is no spec to read, so it is an error.
  Rule: spec-has-no-problems, validate-needs-the-worktree

- **SD12.** `spec validate` prints every problem it finds, each as a
  line `<path>:<line>: error: <what is wrong>` and an indented hint on
  how to fix it under it. Then a line counts the problems, and it exits
  with code 1.
  Why: `path:line` can be clicked in an editor to jump to the problem.
  Showing all problems at once lets the person or the agent fix them in
  one go. The hint saves asking how to fix it. Editors use the word
  `error:` to mark the line, and it leaves room for warnings later. The
  exit code lets a script or an agent check the result without reading
  the text.
  Not chosen: stopping at the first problem, because fixing a spec
  would then take one run per problem; leaving out `error:`, because
  editors could not mark the line, and warnings later would need a new
  format; a code of its own for a spec with problems, e.g. 3, because
  it would be a new convention for one command, and stdout and stderr
  already tell the two failures apart.
  Cost: one mistake can cause several lines, e.g. a broken heading.
  Rule: problems-listed

- **SD13.** `spec validate` finds errors only: every problem it prints
  blocks the spec. It has no warnings.
  Why: an agent that sees exit code 0 moves on, so warnings are often
  not read. Warnings need their own format and exit code, and adding
  them later breaks nothing.
  Not chosen: warnings printed with exit code 0, e.g. for a rule with
  only one example.
  Cost: hints about a weak spec, such as a thin rule, are not given.
  Rule: problems-listed

- **SD14.** `spec validate` checks only what later sliceflow code
  depends on, e.g. that rule names are unique, because the plan and the
  review point to rules by name.
  Why: input is limited only where allowing it would break the system.
  Style and taste, e.g. a long sentence, are left to the reviewer agent.
  Not chosen: style checks, because they block a spec that works, and
  people disagree about style.
  Cost: a spec can pass and still be weak.
  Rule: none

- **SD15.** A spec must have at least one story. Every other section
  of the template (Summary, Shared rules, Out of scope, Decisions, Open
  questions) is optional.
  Why: the plan and the review are built from the stories and their
  rules. No code reads the other sections, and a small slice may have
  nothing to put in them.
  Not chosen: requiring the Summary too, or all six sections, because
  that only adds empty headings.
  Cost: a spec with no summary passes, so a reader may have to read the
  stories to learn what the slice is about.
  Rule: spec-needs-a-story

- **SD16.** A spec may have sections that are not in the template,
  e.g. `## Notes`. A rule inside such a section is an error.
  Why: no code reads extra sections, so they break nothing. But a story
  with a typo in its heading, e.g. `## Stroy 2`, becomes an unknown
  section, and its rules would be lost from the plan and the review.
  It still has rules under it, so the error catches it.
  Not chosen: allowing extra sections with no check, because a typo
  hides rules; allowing no extra sections, because they do no harm.
  Cost: a person who wants rules in a section of their own must make
  it a story or put them in Shared rules.
  Rule: rule-in-unknown-section

- **SD17.** The sections of a spec may come in any order.
  Why: sliceflow finds each section by its heading, not by its place,
  so a different order breaks nothing.
  Not chosen: the template's order as a rule, because it would block
  a spec that works.
  Cost: specs may be laid out differently, so they are a little harder
  to compare. The template shows the usual order.
  Rule: none

- **SD18.** A story heading must be `## Story <n>: <title>`, and the
  stories are numbered 1, 2, 3, ... in file order, with no gap and no
  repeat.
  Why: sliceflow finds a story by the shape of its heading. No code
  reads the number, but people point to a story by it, e.g. "0002,
  story 1" in the backlog, and a repeated number makes that unclear.
  Not chosen: checking only the shape, because a repeated or skipped
  number would pass.
  Cost: this goes a little past SD14, since it guards a person's
  pointer, not code. Moving a story means renumbering the others.
  Rule: story-heading-shape, stories-numbered-in-order

- **SD19.** Every story has a `Uses:` line, which may say `none`. Each
  name in it must be a rule in Shared rules.
  Why: a name that points at nothing, e.g. a typo, drops that shared
  rule from the story's plan and review. A missing line cannot be told
  apart from a forgotten one, so `none` makes the choice clear.
  Not chosen: an optional line where missing means none, because a
  forgotten line would not be noticed.
  Cost: a story with no shared rules still needs the line `Uses: none`.
  Rule: story-needs-uses, uses-names-shared-rules

- **SD20.** A shared rule that no story names in its `Uses:` line is an
  error.
  Why: the plan and the review work story by story, so a shared rule
  with no story would never be built or checked. It is usually a
  forgotten `Uses:` name or a leftover.
  Not chosen: allowing it, because the rule would be lost silently.
  Cost: a shared rule kept for a later story must wait, or be moved to
  the backlog.
  Rule: shared-rule-used

- **SD21.** A rule heading must be ``### Rule `<name>`: <one line>``.
  The name is lowercase words joined by hyphens, and is used only once
  in the whole spec, shared rules included.
  Why: sliceflow finds a rule by its heading. `Uses:` separates names
  with commas, so a space or a comma in a name breaks the list. The
  plan, the review, and `Uses:` point to rules by name, so a repeated
  name makes them unclear.
  Not chosen: free names, e.g. with capitals or spaces, because they
  break the `Uses:` list.
  Cost: names like `R12` or `Edit Values` are refused.
  Rule: rule-heading-shape, rule-name-shape, rule-name-unique

- **SD22.** Every rule has exactly one `Check:` line. It holds one or
  more values, separated by commas, e.g. `Check: unit, e2e`, and each
  value is `unit`, `api`, `e2e`, or `manual`.
  Why: the checklist for trying the slice by hand takes every
  `Check: manual` rule, so a missing line or a typo, e.g. `manaul`,
  drops a rule from it without a word. A rule may need more than one
  kind of test, e.g. the engine in a unit test and the message in an
  e2e test, as the spec template already allows.
  Not chosen: allowing other values, because no code would know what
  they mean; one value only, because a rule that needs two kinds of
  test would lose one.
  Cost: a new kind of test needs a change to sliceflow first.
  Rule: rule-has-one-check, check-values

- **SD23.** If the Open questions section holds anything but `None`,
  it is one error that counts the questions, e.g. "3 questions are
  still open.", on the line of the section's heading. A spec with no
  Open questions section has no such error. A question is an item
  `- **Q<n>.** ...`. If the section holds other text and no such item,
  that is an error too, which asks for Q items or `None`.
  Why: design 2.2 says an open question is an error, because a spec
  with open questions is not ready to submit. The questions sit
  together, so one line that points to the section is enough.
  Not chosen: one error per question, because it repeats the same hint.
  Cost: the error does not point to each question.
  Rule: open-questions-left

- **SD24.** `{{` outside backticks and code blocks is an error: a
  placeholder from the template that was never filled in.
  Why: some leftovers pass every other check, e.g.
  `## Story 1: {{short title}}`. Text that is about placeholders puts
  them in backticks, as code in a spec already is.
  Not chosen: matching only the template's exact placeholders, because
  that breaks each time the template changes; no check, because some
  leftovers would pass.
  Cost: a spec that means the characters `{{` in plain text must put
  them in backticks.
  Rule: placeholder-left

- **SD25.** An HTML comment (`<!-- ... -->`) outside backticks and
  code blocks is an error, one per comment, on the line where it
  starts. A placeholder inside a comment is not reported again.
  Why: a comment is hidden when the Markdown is shown, but an agent
  reads the raw file. So a comment can give an agent an instruction
  the human never sees. A finished spec has no hidden text, so the
  template's guidance comments are deleted before the spec is
  submitted.
  Not chosen: allowing comments and skipping what is inside them,
  because the hidden text would still reach the agent.
  Cost: notes to oneself while writing must be deleted, or written as
  plain text, before the spec passes.
  Rule: comment-left

- **SD26.** Every rule has at least one example, except a rule whose
  `Check:` is only `unit`. A story's first rule always has one.
  Why: the checklist for trying the slice by hand takes the first
  example of each story's first rule. Design 2.2 says a rule the user
  can see has an example; sliceflow cannot tell which rules the user
  sees, and `Check: unit` alone is the closest sign of an inner rule.
  Not chosen: only the first rule of each story, because most rules
  would need none; every rule, because an inner rule has no output to
  show.
  Cost: a rule the user can see, but marked `Check: unit` by mistake,
  passes with no example.
  Rule: rule-needs-an-example

- **SD27.** Every example has `**When**` and `**Then**`. `**Given**`
  is optional. What is inside each part is not checked.
  Why: the checklist shows an example so the human can try it: what to
  do, and what to see. Some examples need no setup. How well an example
  is written is the reviewer's job.
  Not chosen: checking the inside, e.g. that the command is in
  backticks, because that is style, not something code needs.
  Cost: an example with a vague **When** passes.
  Rule: example-has-when-then

- **SD28.** Every decision has a `Rule:` line, which may say `none`.
  Each name in it must be a rule in this spec.
  Why: the line points from a decision to the rules that carry it out.
  A name that points at nothing, e.g. a typo, breaks that link, and a
  missing line cannot be told apart from a forgotten one, as with
  `Uses:`.
  Not chosen: an optional line where missing means none, because a
  forgotten line would not be noticed.
  Cost: a decision that no rule states still needs `Rule: none`.
  Rule: decision-needs-rule-line, decision-rule-names-exist

- **SD29.** Decisions are numbered SD1, SD2, SD3, ... in file order,
  with no gap and no repeat.
  Why: other specs point to a decision by its number, e.g.
  `Replaces: slice-and-events/SD16`. Design 2.2 says that after the
  first approval a number is never reused: a dropped decision stays in
  the list, marked removed, so the numbers stay whole. Before that, the
  decisions can be renumbered freely.
  Not chosen: free numbers, because a repeated number makes a pointer
  unclear.
  Cost: after the first approval, a dropped decision cannot simply be
  deleted.
  Rule: decisions-numbered-in-order

- **SD30.** The four stories stay in one slice, with 50 rules, far
  more than the template's limit of about 25-30.
  Why: the human's choice.
  Not chosen: moving Story 3 (`spec validate`) into a slice of its own;
  packing several checks into one rule, which breaks "one behavior
  each" (design 2.2).
  Cost: a longer plan, a longer review, and a longer wait before the
  slice is merged.
  Rule: none

- **SD31.** `spec validate` does not check a decision's `Why:`,
  `Not chosen:`, and `Cost:` lines.
  Why: no code reads them (SD14). They record the human's thinking,
  and the reviewer agent catches a missing one.
  Not chosen: checking that all three are there, because it adds
  rules to a slice that is already big (SD30).
  Cost: a decision with no `Why:` passes `spec validate`.
  Rule: none

- **SD32.** The problems in a spec, the count line, and "no problems
  found" go to stdout. An error that stops `spec validate` from reading
  the spec, e.g. no slice or no worktree, goes to stderr, as in every
  other command.
  Why: the problems are the result that was asked for, not a failure of
  sliceflow, so a script or an agent can save them with `>`. When the
  spec cannot be read, sliceflow really failed.
  Not chosen: the problems on stderr, as some compilers do, because
  then `>` saves nothing.
  Cost: a script that reads only stderr for errors misses the problems.
  Rule: spec-has-no-problems, problems-listed, validate-errors-as-before

- **SD33.** Every check of `spec validate` skips what is inside a code
  block (between ``` lines).
  Why: a spec often shows text as an example, e.g. the start of the
  template, or a broken heading in a rule about headings. Read as real
  spec text, it would make the spec fail.
  Not chosen: skipping code blocks only in the placeholder and comment
  checks, because a heading in a code block would then count as a story
  or a rule.
  Cost: a real story or rule written inside a code block by mistake is
  not seen.
  Rule: code-blocks-skipped

- **SD34.** `spec validate` has no check of its own for the front
  matter (`slice:` and `created:`).
  Why: no code reads it (SD14). sliceflow knows the slice from the
  spec's folder. A `{{slice}}` left in it is caught as a placeholder.
  Not chosen: checking that `slice:` matches the folder, because
  nothing breaks when it does not.
  Cost: a wrong `slice:` or `created:` passes.
  Rule: none

## Open questions

None
