---
slice: 0001-slice-and-events
---

<!-- HOW TO FILL THIS TEMPLATE

     The preview file says how the human tries the slice at the preview.
     The lead creates it when the plan starts, right after the spec
     review passes, and fills it along with the plan. The human approves
     it with the spec, plan, and tasks, in one approval.

     sliceflow reads this file. Keep every marker exactly as shown: the
     headings, the check numbers (### P1.), and the labeled lines
     (Covers:, Expect:). Don't rename or reorder them. A labeled line
     may go on over more lines: indent the extra lines 2 spaces more than
     the label.

     The fixed part comes from the spec alone, so it is written as soon
     as the spec review passes: the lead writes one ### P check, its
     number, title, and Covers: line, for each of these:
     - the first scenario of each story that is not removed, its main
       success case, e.g. Covers: S1. Tests prove the code does what the
       tasks say; this check lets the human see it does what they meant.
     - each requirement with Check: manual.
     - each success criterion (SC1, ...).
     When the spec changes later, the lead adds checks for new items,
     keeps the text already written, and tells the human about a check
     whose item was removed.

     The lead writes the setup, the steps, and the Expect: lines while
     writing the plan. More checks may be added; number them after the
     last.

     - Replace every {{...}} with real text. The lead fills slice when
       it copies this template.
     - Write plain English, short sentences. The human follows these
       steps by hand, so each step is one thing to do.
     - Results are not written here. The human tells the lead, who
       writes them in status.md → Preview.
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

     1. Run `sliceflow slice new checkout` in the test repo.
        Expect: "Created slice checkout." and the path to its spec.md.

     ### P2. A plan runs with no lead action

     Covers: SC1

     1. Run `sliceflow run` on the test repo.
        Expect: all 5 tasks finish, and the lead is never called.
     2. Run `sliceflow status`.
        Expect: the slice is in the preview phase. -->

### P1. Create a slice

Covers: S1

1. {{A step the human does.}}
   Expect: {{what they should see}}

### P2. See one slice

Covers: S13

1. {{A step the human does.}}
   Expect: {{what they should see}}

### P3. See all slices

Covers: S20

1. {{A step the human does.}}
   Expect: {{what they should see}}

### P4. Get help

Covers: S26

1. {{A step the human does.}}
   Expect: {{what they should see}}
