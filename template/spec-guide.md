# How to write a spec

This guide explains how to fill `spec.md`. The template shows the shape.
This guide holds the rules.

A spec says **what** a feature does, never **how** it is built. No
endpoints, tables, libraries, or file names. Those go in `plan.md`.

The human approves the spec. Write it so they can read it quickly: plain
English, short sentences, one idea per sentence.

## Filling the template

- Replace every `{{...}}` with real text. The text inside says what goes
  there.
- flow fills `{{slice}}`, `{{title}}`, and `{{date}}` itself, when it
  creates the slice.
- A spec that still contains `{{` is not finished.
- Keep the sections in the template's order. Don't rename them.
- Repeat a block as often as needed: more stories, scenarios,
  requirements, decisions.
- A required section with nothing to say gets one line: `None.`
- Two sections are optional: Success criteria and Key concepts. Delete
  them when they don't apply.

## Summary

One or two sentences. What the feature lets someone do, not how.

## Why

The problem it solves, who has the problem, and why it matters now.

## Success criteria (optional)

Use this section only when the slice has a goal for the whole result
that no single requirement covers.

The difference from a requirement:

- **A requirement checks one part.** One test of one behavior proves it.
- **A success criterion checks the whole result.** You use the slice from
  start to end, and maybe count or compare something: time, tokens,
  steps. All requirements can pass while a success criterion fails.

Rules:

- Number them SC1, SC2, ... One to three is enough.
- Each one is measurable. "Easy to use" is not. "One command, no manual
  steps" is.
- Each one says nothing about how it is built.
- Each one has a `Measured:` line. The human checks them at the preview.

Example:

```markdown
- **SC1.** A plan with 5 tasks runs to the end with no action from the
  lead, when no task fails.
  Measured: preview, on a test repo.
```

## User stories

What the user does. This is the first thing a reviewer reads.

- A story is one thing the user wants to do. Give it a short title and
  one sentence: who wants to do what.
- Under it, a numbered list of scenarios, one per case. The main path
  comes first, then each important failure.
- Each scenario has three parts:
  - **Given:** the starting state. Leave it out when there is nothing
    special.
  - **When:** what the user does. Only one per scenario. If you need two,
    split it into two scenarios.
  - **Then:** what the user sees. Show the real screen, message, or
    terminal output, in a code block when it is longer than one line.
- Use "and" for more than one condition or result.
- End each scenario with the requirements it shows, e.g. `(R3)`.
- Every requirement a user can see appears in at least one scenario.

A scenario is one example, with real values. The requirements are the
full rules. Both are needed.

Example:

````markdown
### Story 1: Start a new slice

A developer starts work on a new feature with one command.

1. **When** I run `flow slice new checkout`,
   **Then** flow creates the slice and says where to write the spec:
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
````

## Key concepts (optional)

The important states, terms, and objects, so everyone uses the same
words. One line each. Skip it when the slice is simple.

## Requirements

The full rules, including the ones a user never sees.

- Number them R1, R2, ... The numbering rules are in the document checks.
- Group them under small headings, e.g. `### Starting`, `### Errors`.
- Write each one in one of these forms:
  - When `<trigger>`, the `<system>` shall `<response>`.
  - If `<bad situation>`, then the `<system>` shall `<response>`.
  - While `<state>`, the `<system>` shall `<response>`.
  - The `<system>` shall `<response>`.
- One behavior per requirement.
- Each one has a `Check:` line saying how it is tested: `unit`, `api`,
  `e2e`, or `manual`. More than one is allowed, e.g. `Check: api, e2e`.
- `manual` requirements go into the preview checklist.
- More than about 25-30 requirements? The slice is too big. Split it.

Example:

```markdown
### Errors

- **R3.** If a slice with that name already exists, then flow shall
  change nothing and say so.
  Check: e2e
```

## Out of scope

What this slice does not do, so nobody builds it by mistake. For each
item, say where it goes: a later slice, or the backlog.

## Decisions

Product decisions made while writing the spec.

- Number them SD1, SD2, ...
- Each one has the choice, `Why:`, and `Not chosen:` (the main other
  option, and why not).
- To replace a decision from an earlier slice, add `Replaces:` with that
  slice's name and number, e.g. `device-pairing/SD3`. Otherwise delete
  the line.

## Assumptions

Things we believe are true but did not verify. If one turns out wrong,
the spec may need to change.

## Open questions

Questions for the human, numbered Q1, Q2, ... This section must say
`None.` before the spec check starts.
