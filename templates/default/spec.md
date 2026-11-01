---
slice: {{slice}}
title: {{title}}
created: {{date}}
---

<!-- HOW TO FILL THIS TEMPLATE

     A spec says WHAT a feature does, never HOW it is built. No endpoints,
     tables, libraries, or file names. Those go in plan.md. What the
     user sees does belong here: commands, flags, screens, output layout,
     messages, and exit codes. The plan may not decide these.

     The human approves the spec. Write plain English, short sentences,
     one idea per sentence.

     sliceflow reads this file. Keep every marker exactly as shown: the
     headings, the numbers (**R1.**, **S1.**, **SD1.**, **Q1.**), and the
     labeled lines (Check:, Shows:, Why:, Not chosen:, Cost:, Rule:,
     Replaces:). Don't rename or reorder them. A labeled line
     may go on over more lines: indent the extra lines 2 spaces more than
     the label.

     - Replace every {{...}} with real text. The text inside says what
       goes there. sliceflow fills slice, title, and date itself. If they
       still show {{...}}, fill them yourself.
     - The title of the slice lives only in this front matter. Don't
       repeat it in a heading or in other files.
     - Each section has a comment like this one with its rules. Follow it.
     - Keep the sections in this order. Don't rename them.
     - Repeat a block as often as needed: stories, scenarios,
       requirements, decisions.
     - A required section with nothing to say gets one line: None.
     - When you are done, delete every comment and check that no {{ is
       left. A spec with a comment or a {{ left in it is not finished.
       Every placeholder of this template counts, even inside code. A {{
       or a comment marker you write yourself inside a code block or
       `inline code` is real content.
     - To change a spec later, read this template again first. The
       finished spec no longer has these rules in it. -->

# Spec

## Summary

<!-- Two or three sentences: what the feature lets someone do, not how,
     and why it is needed. The why helps a worker guess right when a
     requirement leaves something open. -->

{{Two or three sentences: what this feature lets someone do, and why.}}

## User stories

<!-- What the user does. This is the first thing a reviewer reads.

     - Every scenario sits under a story, even when there is only one
       story. A story is one thing the user wants to do. It always has
       both: a short title, and one sentence saying who wants to do what.
     - Under it, a list of scenarios, one per case. The success cases
       come first, then each important failure.
     - Number scenarios S1, S2, ... across the whole spec, not per story.
       Same rules as requirement numbers: renumber freely only before the
       spec is first approved. After that, never reuse a number, and a
       removed scenario stays, as: **S4.** (removed in R002: why)
     - Each scenario has three parts:
       - Given: the starting state. Leave it out when there is nothing
         special.
       - When: what the user does. Only one per scenario. If you need
         two, split it into two scenarios.
       - Then: what the user sees. Show the real screen, message, or
         terminal output. Use a code block when it is longer than one
         line.
     - Use "and" for more than one condition or result.
     - End each scenario with a Shows: line naming the requirements it
       shows, e.g. Shows: R1, R2
     - Every requirement a user can see, including error messages,
       appears in at least one scenario.

     A scenario is one example, with real values. The requirements are
     the full rules. Both are needed.

     Example:
     ### Story 1: Start a new slice

     A developer starts work on a new feature with one command.

     - **S1.** **When** I run `sliceflow slice create checkout`,
       **Then** sliceflow creates the slice and says where to write the spec:
       ```
       $ sliceflow slice create checkout
       Created slice checkout.
       Next: write the spec in .sliceflow/worktrees/checkout/docs/slices/checkout/spec.md
       ```
       Shows: R1, R2
     - **S2.** **Given** a slice named `checkout` already exists,
       **When** I run `sliceflow slice create checkout`,
       **Then** sliceflow changes nothing and prints
       "Error: slice checkout already exists: branch slice/checkout.
       Nothing was changed."
       Shows: R3 -->

### Story 1: {{short title}}

{{One sentence: who wants to do what.}}

- **S1.** **When** {{what the user does}},
  **Then** {{what the user sees}}:
  ```
  {{real output}}
  ```
  Shows: {{the requirements this scenario shows, e.g. R1, R2}}
- **S2.** **Given** {{the starting state}},
  **When** {{what the user does}},
  **Then** {{what the user sees}}.
  Shows: {{R3}}

## Requirements

<!-- The full rules, including the ones a user never sees.

     - Number them R1, R2, ... from top to bottom.
     - When you are done, renumber them so they read R1, R2, ... from
       top to bottom in this section, and update every link to match:
       the Shows: lines and the decisions' Rule: lines. You may renumber only before the spec is first approved.
     - Group them under small headings, e.g. ### Starting, ### Errors.
     - Write each one in one of these forms (EARS, without "shall"),
       in the present tense:
       - When <trigger>, <system> <response>.
       - If <bad situation>, <system> <response>.
       - While <state>, <system> <response>.
       - Where <feature or setting is on>, <system> <response>.
       - <system> <response>.
       Two keywords may be combined when a rule needs both, e.g.
       While <state>, when <trigger>, <system> <response>.
       For <system>, use the product's name, e.g. sliceflow. Write
       "sliceflow prints", not "sliceflow shall print" or "sliceflow
       must print". For something it must not do, write "never", e.g.
       "sliceflow never sends data over the network."
     - Name exact values: the message printed, the limit, the exit
       code. A worker sees this text, so "say so" is not enough.
     - One behavior per requirement.
     - A term that needs one fixed meaning is defined in the requirement
       where it first appears, e.g. "A task is overdue when its due date
       is before today."
     - Each one has a Check: line saying how it is tested. More than one
       is allowed, e.g. Check: unit, e2e.
       - unit: tests one piece of code on its own.
       - api: calls the program's interface directly, e.g. an HTTP
         endpoint. A plain command-line tool usually has none: use e2e.
       - e2e: runs the real program the way a user does.
       - manual: a human checks it at the preview. Use it only when no
         automatic test can check it.
     - More than about 25-30 requirements? The slice is too big. Split
       it.

     When the spec changes after it was approved:
     - Never reuse a number. A new requirement gets the next free number.
     - A removed requirement stays, as: **R4.** (removed in R002: why)

     Example:
     ### Errors

     - **R3.** If a slice with that name already exists, sliceflow changes
       nothing and prints "Error: slice <name> already exists:
       branch slice/<name>. Nothing was changed."
       Check: e2e

     ### Settings

     - **R9.** Where task review is turned off, sliceflow merges a task
       without starting a reviewer.
       Check: e2e -->

### {{group name}}

- **R1.** {{When <trigger>, <system> <response>.}}
  Check: {{unit, api, e2e, or manual}}

## Out of scope

<!-- What this slice does not do, even if it looks related, so nobody
     builds it by mistake. One - bullet per item, one line each. No
     sub-bullets, headings, or code blocks.

     Example:
     - Reminders and notifications.
     - Repeating tasks. -->

- {{What this slice does not do.}}

## Decisions

<!-- Product decisions made while writing the spec.

     - Number them SD1, SD2, ...
     - Each one has the choice, Why:, and Not chosen: (the main other
       option, and why not).
     - Each one has a Cost: line: the downside we accept with this
       choice. With truly no downside, write Cost: None. A reviewer
       questions a None that looks wrong.
     - Each one has a Rule: line. A worker sees requirements, not
       decisions. So a decision that changes what the feature does is
       also written as a requirement, and Rule: names it, e.g. Rule: R11.
       A decision that changes no behavior says Rule: none. sliceflow checks
       that every decision has this line, and that it names a real
       requirement.
     - To replace a decision from an earlier slice, add Replaces: with
       that slice's name and the decision's number, e.g. device-pairing/SD3.
       Otherwise remove the whole Replaces: line.

     Example:
     - **SD2.** A task due today is not overdue.
       Why: the user still has the whole day to finish it.
       Not chosen: overdue on the due date, because it warns people who
         still have time.
       Cost: a task finished late at night still shows as on time that
         day.
       Rule: R11 -->

- **SD1.** {{The choice.}}
  Why: {{the reason}}
  Not chosen: {{the main other option, and why not}}
  Cost: {{the downside we accept, or None}}
  Rule: {{the requirement that states it, e.g. R11, or none}}
  Replaces: {{an earlier decision, e.g. device-pairing/SD3}}

## Open questions

<!-- Questions for the human, numbered Q1, Q2, ... An answered question
     is written into the spec and removed from here. A deferred question
     that does not block the build moves to Out of scope and the backlog.
     The spec review starts only when this section says None. -->

- **Q1.** {{A question for the human.}}
