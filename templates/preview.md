---
slice: {{slice}}
based_on: { spec: {{spec revision, e.g. R001}} }
---

<!-- HOW TO FILL THIS TEMPLATE

     The preview file says how the human tries the slice at the preview.
     flow creates it when the spec is approved (see below); the lead
     fills it during the plan phase; the human approves it together with
     the plan.

     flow reads this file. Keep every marker exactly as shown: the
     headings, the check numbers (### P1.), and the labeled lines
     (Covers:, Expect:). Don't rename or reorder them. A labeled line
     may go on over more lines: indent the extra lines 2 spaces more than
     the label.

     The fixed part comes from the spec alone, so it is written as soon
     as the spec is approved: flow creates this file then, with one
     ### P check, its number, title, and Covers: line, for each of these:
     - the first scenario of each story that is not removed, its main
       success case, e.g. Covers: S1. Tests prove the code does what the
       tasks say; this check lets the human see it does what they meant.
     - each requirement with Check: manual.
     - each success criterion (SC1, ...).
     When the spec is approved again after a change, flow adds checks for
     new items. It never deletes text you wrote, and it reports a check
     whose item was removed. (In manual flow, the lead writes these
     checks by hand, by the same rules, right after the spec is
     approved.)

     You write the setup, the steps, and the Expect: lines, during the
     plan phase. You may add more checks; number them after the last.

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
     - A section with nothing to say gets one line: None. E.g. no setup
       needed: None. under Setup. Checks is never empty while the spec
       has a story. -->

# Preview

## Setup

<!-- What the human does once before the checks: build the program,
     make a test folder or test repo, add test data. Numbered steps.
     Each step that gives a result has an Expect: line. -->

1. {{A setup step.}}

## Checks

<!-- One ### subsection per check, numbered P1, P2, ...

     - Covers: names what this check proves: scenarios (S1),
       requirements (R12), and success criteria (SC1). The first
       scenario of each story that is not removed, every requirement
       with Check: manual, and every success criterion are in at least
       one Covers: line.
     - For a scenario, the steps and Expect: lines follow the scenario's
       exact command and output.
     - Then numbered steps. Each step is one thing the human does, with
       an Expect: line saying what they should see.

     Example:
     ### P1. Start a new slice

     Covers: S1

     1. Run `flow slice new checkout` in the test repo.
        Expect: "Created slice checkout." and the path to its spec.md.

     ### P2. A plan runs with no lead action

     Covers: SC1

     1. Run `flow run` on the test repo.
        Expect: all 5 tasks finish, and the lead is never called.
     2. Run `flow status`.
        Expect: the slice is in the preview phase. -->

### P1. {{short title}}

Covers: {{S1, R12, SC1}}

1. {{A step the human does.}}
   Expect: {{what they should see}}
