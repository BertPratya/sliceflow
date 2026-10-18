---
slice: {{slice}}
title: {{title}}
created: {{date}}
---

# Spec: {{title}}

<!-- A spec says WHAT the feature does, never HOW it is built.
     No endpoints, tables, libraries, or file names. Those go in plan.md. -->

## Summary
<!-- One or two sentences. What this feature lets someone do. -->

## Why
<!-- The problem it solves, who it is for, and why now. -->

## What the user does
<!-- A step-by-step walkthrough of the main path, with real screens,
     messages, or terminal output. Add one or two important failure paths
     if they matter. This is the first thing a reviewer reads. -->

## Key concepts
<!-- Optional. The important states, terms, and objects, so everyone uses
     the same words. A state diagram or a small table helps.
     Delete this section if the slice is simple. -->

## Requirements
<!-- Numbered R1, R2, ... Group them under small headings (### Starting,
     ### Errors, ### Security, ...).

     Write each one in one of these forms:
     - When <trigger>, the <system> shall <response>.
     - If <bad situation>, then the <system> shall <response>.
     - While <state>, the <system> shall <response>.
     - The <system> shall <response>.        (always true, incl. speed and limits)

     End each one with how it is checked: unit, api, e2e, or manual.
     Requirements checked "manual" go into the preview checklist.

     Example:
     - **R7.** When Accept is clicked on a pairing that is waiting for approval,
       the server shall show a pairing code in the browser.
       Check: api, e2e

     Numbering rules. Plans, tasks, tests, and reviews refer to these
     numbers, so they must stay stable:
     1. A number is never reused or renumbered.
     2. A new requirement takes the next free number, under any heading.
        If the last one is R19, the new one is R20.
     3. A removed requirement stays as a line:
        - **R5.** (removed in R002: replaced by R20)
     4. Changed wording keeps the number if the intent is the same.
        If the intent changes, remove the old one and add a new number.
     5. Use exactly the format above. The document checks
        (`workflow/formats/document-checks.md`) refuse a spec that breaks these rules.

     More than about 25-30 requirements? Consider splitting the slice. -->

## Out of scope
<!-- What this slice does NOT do. For each item, say where it goes:
     a later slice, or the backlog. -->

## Decisions
<!-- Product decisions, numbered SD1, SD2, ... Each one: the choice, then why.
     To replace a decision from an earlier slice, write it here and say which
     one it replaces, e.g. "Replaces 0001/SD11". Never edit a finished slice. -->

## Assumptions
<!-- Things we believe are true but did not verify. -->

## Open questions
<!-- Numbered Q1, Q2, ... Must be empty before submitting for review. -->
