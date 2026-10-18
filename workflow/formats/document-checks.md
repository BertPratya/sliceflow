# Document checks

Format checks for a slice's spec, plan, and task files. The lead runs them
after writing or changing a document, and again before submitting it.
The plan reviewer runs them again during a review.

They check form, not meaning. Whether a requirement makes sense, is
testable, or is really covered is the reviewer's job.

Two levels:

- **error:** something is broken, or points to something that does not exist.
  Fix it now.
- **warning:** not finished yet, or out of date. Normal while drafting.
  A document is submitted or approved only when it has no warnings left,
  except X1 once the human has been told (with a Log line).

Report each problem as one line: `path:line: level: what is wrong, and how to fix it`.

## Spec (`spec.md`)

| # | Check | Level |
|---|---|---|
| S1 | Header fields filled and valid (slice, title, created) | error |
| S2 | Header `slice` matches the folder name | error |
| S3 | No `{{placeholder}}` left from the template | error |
| S4 | Every template section is present ("Key concepts" may be deleted) | error |
| S5 | Each requirement is `- **R<n>.** text`, followed by a `Check:` line (unit, api, e2e, manual) | error |
| S6 | No line that looks like a requirement but has the wrong form (e.g. `**R7**` without the dot) | error |
| S7 | No number is used twice | error |
| S8 | No gaps in the numbers. If a missing number was approved before, see S9; otherwise renumber the requirements not approved yet. Numbers out of order are fine. | error |
| S9 | Once an approved version exists in `history/spec/`: no approved number disappears. A removed one stays as `- **R5.** (removed in R002: why)`. | error |
| S10 | A removed number never comes back with new text | error |
| S11 | Each requirement contains "shall" | warning |
| S12 | Open questions is empty | warning |
| S13 | Every section has content or "(none)" | warning |

## Plan (`plan.md`)

| # | Check | Level |
|---|---|---|
| P1 | Header valid (track: simple or complex) | error |
| P2 | Header `slice` matches the folder name | error |
| P3 | No `{{placeholder}}` left | error |
| P4 | Sections present: Approach, Flows, Interfaces, Data, Decisions, Open questions ("Risks and rollout" and "Parked" are optional) | error |
| P5 | Every `Covers:` line names a real requirement (it exists and is not removed) | error |
| P6 | Decisions are `- **TD<n>.**`, unique, with no gaps. After approval none disappears: a replaced one stays as "(replaced by TD9)". | error |
| P7 | Every requirement is covered by at least one flow or shared rule | warning |
| P8 | The plan is based on the latest approved spec revision | warning |
| P9 | Open questions is empty | warning |
| P10 | Every section has content or "(none)" | warning |

## Tasks (`tasks/NN-<name>.md`)

| # | Check | Level |
|---|---|---|
| T1 | File name `tasks/NN-<name>.md`, numbers unique, header `task` matches | error |
| T2 | Header valid (layer filled, risk, a non-empty `touches`) | error |
| T3 | No placeholder left, all sections present | error |
| T4 | `covers` names real requirements | error |
| T5 | The Requirements section copies each covered requirement word for word from the spec | error |
| T6 | `depends_on` names existing tasks, with no loops | error |
| T7 | `touches` never includes a file from the plan's `touches` (contracts and other living docs), even inside a folder it lists | error |
| T8 | Every requirement is covered by at least one task | warning |
| T9 | Every covered requirement has at least one row in "Done when" | warning |
| T10 | `touches` stays inside the project's parts (stack profile); outside them no checks would run | warning |
| T11 | The task is based on the current plan revision | warning |

## Across slices

| # | Check | Level |
|---|---|---|
| X1 | Another unmerged slice changes a file in this plan's `touches`, and neither is stacked on the other | warning |
| X2 | After approval, the slice's contract change no longer matches the approved patch in `history/contract/` | warning |
| X3 | Stacked slice: the earlier slice's plan changed after this plan was approved | warning |
