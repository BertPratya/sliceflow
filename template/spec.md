---
slice: {{slice}}
title: {{title}}
created: {{date}}
---

<!-- HOW TO FILL THIS TEMPLATE

     A spec says WHAT a feature does, never HOW it is built. No endpoints,
     tables, libraries, or file names. Those go in plan.md. Commands,
     flags, screens, and messages the user sees do belong here.

     The human approves the spec. Write plain English, short sentences,
     one idea per sentence.

     - Replace every {{...}} with real text. The text inside says what
       goes there. flow fills slice, title, and date itself. If they
       still show {{...}}, fill them yourself.
     - Each section has a comment like this one with its rules. Follow it.
     - Keep the sections in this order. Don't rename them.
     - Repeat a block as often as needed: stories, scenarios,
       requirements, decisions.
     - A required section with nothing to say gets one line: None.
     - Two sections are optional: Success criteria and Key concepts.
       Delete them when they don't apply.
     - When you are done, delete every comment and check that no {{ is
       left. A spec with a comment or a {{ left in it is not finished.
     - To change a spec later, read this template again first. The
       finished spec no longer has these rules in it. -->

# Spec: {{title}}

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
     - Under it, a numbered list of scenarios, one per case. The success
       cases come first, then each important failure.
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

     1. **When** I run `flow slice new checkout`,
        **Then** flow creates the slice and says where to write the spec:
        ```
        $ flow slice new checkout
        Created slice checkout.
        Next: write the spec in slices/checkout/spec.md
        ```
        Shows: R1, R2
     2. **Given** a slice named `checkout` already exists,
        **When** I run `flow slice new checkout`,
        **Then** flow changes nothing and prints
        "Error: slice checkout already exists. Nothing was changed."
        Shows: R3 -->

### Story 1: {{short title}}

{{One sentence: who wants to do what.}}

1. **When** {{what the user does}},
   **Then** {{what the user sees}}:
   ```
   {{real output}}
   ```
   Shows: {{the requirements this scenario shows, e.g. R1, R2}}
2. **Given** {{the starting state}},
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
       top to bottom in this section, and update the scenario links to
       match. You may renumber only before the spec is first approved.
     - Group them under small headings, e.g. ### Starting, ### Errors.
     - Write each one in one of these forms:
       - When <trigger>, the <system> shall <response>.
       - If <bad situation>, then the <system> shall <response>.
       - While <state>, the <system> shall <response>.
       - The <system> shall <response>.
       For <system>, use the product's name, e.g. flow.
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

     - **R3.** If a slice with that name already exists, then flow shall
       change nothing and say so.
       Check: e2e -->

### {{group name}}

- **R1.** {{When <trigger>, the <system> shall <response>.}}
  Check: {{unit, api, e2e, or manual}}

## Out of scope

<!-- What this slice does not do, so nobody builds it by mistake.
     For each item, say where it goes:
     - a later slice (name it if you know it)
     - the backlog
     - not planned: we decided not to do it -->

- {{What this slice does not do.}} Goes to: {{a later slice, the backlog, or not planned}}

## Decisions

<!-- Product decisions made while writing the spec.

     - Number them SD1, SD2, ...
     - Each one has the choice, Why:, and Not chosen: (the main other
       option, and why not).
     - To replace a decision from an earlier slice, add Replaces: with
       that slice's name and the decision's number, e.g. device-pairing/SD3.
       Otherwise remove the whole Replaces: line. -->

- **SD1.** {{The choice.}}
  Why: {{the reason}}
  Not chosen: {{the main other option, and why not}}
  Replaces: {{an earlier decision, e.g. device-pairing/SD3}}

## Assumptions

<!-- Things we believe are true but did not verify. If one turns out
     wrong, the spec may need to change. -->

- {{Something we believe is true but did not verify.}}

## Open questions

<!-- Questions for the human, numbered Q1, Q2, ... This section must say
     None. before the spec check starts. -->

- **Q1.** {{A question for the human.}}
