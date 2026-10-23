---
slice: 0001-slice-and-events
title: Slice and events
created: 2026-10-23
ideas: []
---

<!-- HOW TO FILL THIS TEMPLATE

     A spec says WHAT a feature does, never HOW it is built. No endpoints,
     tables, libraries, or file names. Those go in plan.md. What the
     user sees does belong here: commands, flags, screens, output layout,
     messages, and exit codes. The plan may not decide these.

     The human approves the spec. Write plain English, short sentences,
     one idea per sentence.

     sliceflow reads this file. Keep every marker exactly as shown: the
     headings, the numbers (**R1.**, **S1.**, **SD1.**, **SC1.**, **Q1.**), and the
     labeled lines (Check:, Shows:, Measured:, Why:, Not chosen:, Cost:, Rule:,
     Noticed:,
     Replaces:, Goes to:). Don't rename or reorder them. A labeled line
     may go on over more lines: indent the extra lines 2 spaces more than
     the label. Goes to: is the exception: it ends a one-line item.

     - Replace every {{...}} with real text. The text inside says what
       goes there. sliceflow fills slice, title, and date itself. If they
       still show {{...}}, fill them yourself.
     - ideas: the idea files from docs/ideas/ this slice builds on.
       Start the Summary and Why from them. None: ideas: []
     - The title of the slice lives only in this front matter. Don't
       repeat it in a heading or in other files.
     - Each section has a comment like this one with its rules. Follow it.
     - Keep the sections in this order. Don't rename them.
     - Repeat a block as often as needed: stories, scenarios,
       requirements, decisions.
     - A required section with nothing to say gets one line: None.
     - Two sections are optional: Success criteria and Key concepts.
       Delete them when they don't apply.
     - When you are done, delete every comment and check that no {{ is
       left. A spec with a comment or a {{ left in it is not finished.
       Every placeholder of this template counts, even inside code. A {{
       or a comment marker you write yourself inside a code block or
       `inline code` is real content.
     - To change a spec later, read this template again first. The
       finished spec no longer has these rules in it. -->

# Spec

## Summary

<!-- One or two sentences. What the feature lets someone do, not how. -->

{{One or two sentences: what this feature lets someone do.}}

## Why

<!-- The problem it solves, who has the problem, and why it matters now. -->

{{The problem it solves, who it is for, and why now.}}

## Success criteria

<!-- Optional. Most small slices have none. When in doubt, delete this
     section.

     Use it only when the slice has a goal for the whole result that no
     single requirement covers.
     - A requirement checks one part. One test of one behavior proves it.
     - A success criterion checks the whole result. You use the slice
       from start to end, and maybe count or compare something: time,
       tokens, steps. All requirements can pass while it fails.
     If it repeats a requirement, it is not a success criterion.

     - Number them SC1, SC2, ... One to three is enough.
     - Each one is measurable. "Easy to use" is not. "One command, no
       manual steps" is.
     - Each one says nothing about how it is built.
     - Each one has a Measured: line. The human checks them at the
       preview.

     Example:
     - **SC1.** A plan with 5 tasks runs to the end with no action from
       the lead, when no task fails.
       Measured: preview, on a test repo. -->

- **SC1.** {{A measurable result of the whole slice.}}
  Measured: {{how and when it is checked}}

## User stories

<!-- What the user does. This is the first thing a reviewer reads.

     - A story is one thing the user wants to do: a short title, and one
       sentence saying who wants to do what.
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

     - **S1.** **When** I run `sliceflow slice new checkout`,
       **Then** sliceflow creates the slice and says where to write the spec:
       ```
       $ sliceflow slice new checkout
       Created slice checkout.
       Next: write the spec in docs/slices/checkout/spec.md
       ```
       Shows: R1, R2
     - **S2.** **Given** a slice named `checkout` already exists,
       **When** I run `sliceflow slice new checkout`,
       **Then** sliceflow changes nothing and prints
       "Error: slice checkout already exists. Nothing was changed."
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

## Key concepts

<!-- Optional. Use it when the slice brings in terms that need one fixed
     meaning, e.g. "overdue". Otherwise delete this section.

     The important states, terms, and objects, so everyone uses the same
     words. One line each. -->

- **{{term}}:** {{what it means}}

## Requirements

<!-- The full rules, including the ones a user never sees.

     - Number them R1, R2, ... from top to bottom.
     - When you are done, renumber them so they read R1, R2, ... from
       top to bottom in this section, and update every link to match:
       the Shows: lines and the decisions' Rule: lines. You may renumber only before the spec is first approved.
     - Group them under small headings, e.g. ### Starting, ### Errors.
     - Write each one in one of these forms (EARS):
       - When <trigger>, the <system> shall <response>.
       - If <bad situation>, then the <system> shall <response>.
       - While <state>, the <system> shall <response>.
       - Where <feature or setting is on>, the <system> shall
         <response>.
       - The <system> shall <response>.
       Two keywords may be combined when a rule needs both, e.g.
       While <state>, when <trigger>, the <system> shall <response>.
       For <system>, use the product's name, e.g. sliceflow.
     - Name exact values: the message printed, the limit, the exit
       code. A worker sees this text, so "say so" is not enough.
     - One behavior per requirement.
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

     - **R3.** If a slice with that name already exists, then sliceflow shall
       change nothing and print "Error: slice <name> already exists.
       Nothing was changed."
       Check: e2e

     ### Settings

     - **R9.** Where task review is turned off, sliceflow shall merge a task
       without starting a reviewer.
       Check: e2e -->

### {{group name}}

- **R1.** {{When <trigger>, the <system> shall <response>.}}
  Check: {{unit, api, e2e, or manual}}

## Out of scope

<!-- What this slice does not do, so nobody builds it by mistake.
     For each item, say where it goes:
     - a later slice (name it if you know it)
     - the backlog
     - not planned: we decided not to do it

     One - bullet per item, on one line: the item, then Goes to: at the
     end of the line. No sub-bullets, headings, or code blocks.

     An item that goes to the backlog is also added to
     docs/backlog.md on main at once, as plan-slice says.

     Example:
     - Reminders and notifications. Goes to: the backlog
     - Repeating tasks. Goes to: a later slice (recurring-tasks) -->

- {{What this slice does not do.}} Goes to: {{a later slice, the backlog, or not planned}}

## Decisions

<!-- Product decisions made while writing the spec.

     - Number them SD1, SD2, ...
     - Each one has the choice, Why:, and Not chosen: (the main other
       option, and why not).
     - Optional: a Cost: line for the downside we accept with this
       choice. Leave it out when there is no real downside.
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
  Cost: {{optional: the downside we accept, or delete this line}}
  Rule: {{the requirement that states it, e.g. R11, or none}}
  Replaces: {{an earlier decision, e.g. device-pairing/SD3}}

## Assumptions

<!-- Things we believe are true but did not verify. If one turns out
     wrong, the spec may need to change.

     An assumption is a fact about the world: the environment, other
     systems, the users. A choice about what the feature does is never
     an assumption. It is a requirement, a decision, or an Open question
     for the human.

     Optional: add a Noticed: line saying how we would notice if it is
     wrong.

     Example:
     - Users run todo in one time zone.
       Noticed: a user across time zones sees overdue marks a day off.
     Not an assumption (a choice; ask it as a question instead):
     - Dates before today are refused when adding a task. -->

- {{A fact we believe is true but did not verify.}}
  Noticed: {{optional: how we would notice if it is wrong, or delete this line}}

## Open questions

<!-- Questions for the human, numbered Q1, Q2, ... An answered question
     is written into the spec and removed from here. A deferred question
     that does not block the build moves to Out of scope and the backlog.
     The spec review starts only when this section says None. -->

- **Q1.** {{A question for the human.}}
