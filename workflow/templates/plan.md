---
slice: {{slice}}
title: {{title}}
created: {{date}}
based_on: { spec: {{spec_revision}} }   # stacked slice, also: after: { slice: "0001", plan: R001 }
track: {{track}}              # simple | complex
touches:                      # project files this slice changes (contract, data model, ...)
  # - api/openapi.yaml
  # - design/data-model.md
---

# Plan: {{title}}

<!-- A plan says HOW the spec will be built: behavior and decisions.
     Building details (exact messages, file-by-file steps) go in tasks.
     The plan never adds scope. Anything not in the spec needs a spec change.

     The ## sections below are fixed. Add ### subsections freely inside them. -->

## Approach
<!-- The whole idea in a few sentences. A reader should understand the plan
     from this section alone. Then a small table of the parts that change:

     | Part | Folder | Change |
     |---|---|---|
     | Server pairing | internal/server/pairing | new | -->

## Flows
<!-- One ### subsection per main path, including important failure paths.
     Each one starts with "Covers: R…" and lists the steps.
     Add a diagram (Mermaid) for complex flows.

     ### Flow 1: Start pairing
     Covers: R1, R5
     1. …

     Rules that apply across several flows (expiry, retries, locking,
     security) go under "### Shared rules", one #### subsection each,
     also with "Covers: R…".

     Every requirement must be covered by at least one flow or shared rule. -->

## Interfaces
<!-- Every way the outside world talks to the system that this slice touches.
     One ### subsection per kind: HTTP API, Events, CLI, Settings, UI routes.

     If a contract file exists (e.g. OpenAPI), LIST the operations and LINK
     the file. Never copy request or response details.
     If there is no contract file (CLI, settings), describe it fully here. -->

## Data
<!-- What stored data changes: new or changed tables, fields, local files.
     How existing data is migrated, if any.
     Link to the project data model doc for full details. -->

## Decisions
<!-- Technical decisions, numbered TD1, TD2, ...
     Each one: the choice, why, and the main option rejected.

     - **TD1.** Proofs go in the Authorization header.
       Why: URLs end up in logs. Rejected: query parameters. -->

## Open questions
<!-- Numbered Q1, Q2, ... Must be empty before submitting for review. -->

## Risks and rollout
<!-- Optional. Delete if not needed.
     What could go wrong, and how it is reduced. Migration order, feature
     flags, compatibility with older versions. -->

## Parked
<!-- Optional. Delete if not needed.
     Ideas found while planning that belong to a later slice.
     Moved to the backlog when the slice is done. -->
