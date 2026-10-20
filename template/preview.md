---
slice: {{slice}}
based_on: { spec: {{spec revision, e.g. R001}} }
---

<!-- HOW TO FILL THIS TEMPLATE

     The preview file says how the human tries the slice at the preview.
     It is written with the plan, and approved together with it.

     flow reads this file. Keep every marker exactly as shown: the
     headings, the check numbers (### P1.), and the labeled lines
     (Covers:, Expect:). Don't rename or reorder them. A labeled line
     may go on over more lines: indent the extra lines 2 spaces more than
     the label.

     flow writes the fixed part: one ### P check for each requirement
     with Check: manual, and one for each success criterion (SC1, ...),
     each with its Covers: line. You write the setup, the steps, and the
     Expect: lines. You may add more checks; number them after the last.

     - Replace every {{...}} with real text. flow fills slice and
       based_on itself. If they still show {{...}}, fill them
       yourself.
     - Write plain English, short sentences. The human follows these
       steps by hand, so each step is one thing to do.
     - Results are not written here. The human reports them through flow,
       e.g. flow preview fail P2 "the date shows wrong".
     - When you are done, delete every comment and check that no {{ is
       left. Every placeholder of this template counts, even inside
       code. A {{ or a comment marker you write yourself inside a code
       block or `inline code` is real content.
     - A section with nothing to say gets one line: None. E.g. nothing to
       check by hand: None. under both Setup and Checks. -->

# Preview

## Setup

<!-- What the human does once before the checks: build the program,
     make a test folder or test repo, add test data. Numbered steps.
     Each step that gives a result has an Expect: line. -->

1. {{A setup step.}}

## Checks

<!-- One ### subsection per check, numbered P1, P2, ...

     - Covers: names what this check proves: requirements (R12) and
       success criteria (SC1). Every requirement with Check: manual and
       every success criterion is in at least one Covers: line.
     - Then numbered steps. Each step is one thing the human does, with
       an Expect: line saying what they should see.

     Example:
     ### P1. A plan runs with no lead action

     Covers: SC1

     1. Run `flow run` on the test repo.
        Expect: all 5 tasks finish, and the lead is never called.
     2. Run `flow status`.
        Expect: the slice is in the preview phase. -->

### P1. {{short title}}

Covers: {{SC1, R12}}

1. {{A step the human does.}}
   Expect: {{what they should see}}
