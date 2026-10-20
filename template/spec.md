---
slice: {{slice}}
title: {{title}}
created: {{date}}
---

# Spec: {{title}}

## Summary

{{One or two sentences: what this feature lets someone do.}}

## Why

{{The problem it solves, who it is for, and why now.}}

## Success criteria

{{Optional. Delete this section if no goal for the whole result goes beyond the requirements.}}

- **SC1.** {{A measurable result of the whole slice.}}
  Measured: {{how and when it is checked}}

## User stories

### Story 1: {{short title}}

{{One sentence: who wants to do what.}}

1. **When** {{what the user does}},
   **Then** {{what the user sees}}:
   ```
   {{real output}}
   ```
   ({{requirements this shows, e.g. R1, R2}})
2. **Given** {{the starting state}},
   **When** {{what the user does}},
   **Then** {{what the user sees}}. ({{R3}})

## Key concepts

{{Optional. Delete this section if the slice is simple.}}

- **{{term}}:** {{what it means}}

## Requirements

### {{group name}}

- **R1.** {{When <trigger>, the <system> shall <response>.}}
  Check: {{unit, api, e2e, or manual}}

## Out of scope

- {{What this slice does not do.}} Goes to: {{a later slice, or the backlog}}

## Decisions

- **SD1.** {{The choice.}}
  Why: {{the reason}}
  Not chosen: {{the main other option, and why not}}
  Replaces: {{an earlier decision, e.g. device-pairing/SD3, or delete this line}}

## Assumptions

- {{Something we believe is true but did not verify.}}

## Open questions

- **Q1.** {{A question for the human.}}
