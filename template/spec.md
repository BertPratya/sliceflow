---
slice: {{slice}}
title: {{title}}
created: {{date}}
---

# Spec: {{title}}

<!-- A spec says WHAT the feature does, never HOW it is built.
     No endpoints, tables, libraries, or file names. Those go in plan.md.
     When you fill a section, delete its comment. -->

## Summary
<!-- One or two sentences. What this feature lets someone do. -->

## Why
<!-- The problem it solves, who it is for, and why now. -->

## Success criteria
<!-- Optional. Use it only when the slice has a goal for the whole result
     that no single requirement covers. Otherwise delete this section.

     A requirement checks one part: one test of one behavior proves it.
     A success criterion checks the whole result: you use the slice from
     start to end, and maybe count or compare something (time, tokens,
     steps). All requirements can pass while a success criterion fails.

     Numbered SC1, SC2, ... Each one is measurable, says nothing about
     how it is built, and says how it is measured. The human checks them
     at the preview.

     - **SC1.** A plan with 5 tasks runs to the end with no action from
       the lead, when no task fails.
       Measured: preview, on a test repo.
     - **SC2.** The lead uses fewer tokens per task than when running the
       same slice by hand.
       Measured: compare the usage of one run with a manual run.

     One to three is enough. -->

## User stories
<!-- What the user does, as stories. This is the first thing a reviewer
     reads.

     A story is one thing the user wants to do. Under it, a numbered list
     of scenarios: one scenario per case. The main path comes first, then
     each important failure.

     - Given: the starting state. Leave it out when there is nothing
       special.
     - When: what the user does. Only one per scenario; if you need two,
       split the scenario.
     - Then: what the user sees. Show real screens, messages, or terminal
       output.
     - Use "and" for more than one condition or result.
     - At the end, name the requirements the scenario shows, e.g. (R3).

     ### Story 1: Start a new slice

     A developer starts work on a new feature with one command.

     1. **When** I run `flow slice new checkout`,
        **Then** flow creates the slice and tells me where to write the
        spec:
        ```
        $ flow slice new checkout
        Created slice checkout.
        Next: write the spec in slices/checkout/spec.md
        ```
        (R1, R2)
     2. **Given** a slice named `checkout` already exists,
        **When** I run `flow slice new checkout`,
        **Then** flow changes nothing and prints
        "Error: slice checkout already exists. Nothing was changed." (R3)

     A scenario is one example. The requirements below are the full rules.
     Every requirement a user can see should appear in at least one
     scenario. -->

## Key concepts
<!-- Optional. The important states, terms, and objects, so everyone uses
     the same words. Delete this section if the slice is simple. -->

## Requirements
<!-- The full rules, including the ones a user never sees.
     Numbered R1, R2, ... Group them under small headings (### Starting,
     ### Errors, ...).

     Each one in one of these forms, then a Check: line
     (unit, api, e2e, or manual):
     - When <trigger>, the <system> shall <response>.
     - If <bad situation>, then the <system> shall <response>.
     - While <state>, the <system> shall <response>.
     - The <system> shall <response>.

     - **R3.** If a slice with that name already exists, then flow shall
       change nothing and say so.
       Check: e2e

     "manual" requirements go into the preview checklist.
     The numbering rules are in the document checks.
     More than about 25-30 requirements? Split the slice. -->

## Out of scope
<!-- What this slice does NOT do. For each item, say where it goes:
     a later slice, or the backlog. -->

## Decisions
<!-- Product decisions, numbered SD1, SD2, ... Each one: the choice, why,
     and the main option not chosen. To replace a decision from an earlier
     slice, say which one, e.g. "Replaces device-pairing/SD3". -->

## Assumptions
<!-- Things we believe are true but did not verify. -->

## Open questions
<!-- Numbered Q1, Q2, ... Must be empty before the spec check. -->
