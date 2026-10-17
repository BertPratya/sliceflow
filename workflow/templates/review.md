---
review: {{review}}            # <kind>-<target>-<round> (integration: integration-<round>), same as the file name
slice: {{slice}}
kind: {{kind}}                # spec | plan | task | integration
target: {{target}}            # e.g. { spec: R001 }, { plan: R001 }, { task: "02", commit: a1b2c3d }
round: {{round}}
reviewer: {{reviewer}}        # role (tool), e.g. code-reviewer (codex)
# --- managed by flow, do not edit ---
verdict: none                 # pass | changes | needs_human, set on submit
submitted_at: none
---

# Review: {{title}}

<!-- flow creates this file and fills the header. You write the sections below.
     When done, run: flow review submit {{slice}}/{{review}}
     After submit the file is locked. Never edit it again.

     flow sets the verdict from what you write:
     - "Needs human decision" is not empty     → needs_human
     - else any high or medium finding,
       new or reopened                         → changes
     - else                                    → pass  (low findings alone pass)

     Spec reviews are short: fill only Summary and "Needs human decision"
     (at most 5 questions). Write "(none)" in the other sections.

     Report real problems only. Do not invent issues to look thorough.
     If everything is fine, say so and leave Findings as "(none)".
     Sections that do not apply: write "(none)". -->

## Summary
<!-- 1-3 sentences. Overall impression, and the main problem if there is one. -->

## Previous findings
<!-- Round 2 and later. Every finding from the last round that the worker
     marked fixed, wontfix, or deferred. For each: verified, or reopened with why.

     - task-02-1/F1: verified
     - task-02-1/F2: reopened. Still returns 500 when id is not a number.

     Round 1: write "(none)". -->

## Findings
<!-- New findings only. Number them F1, F2, ... in this file.
     Outside this file they are called <review>/F1, e.g. task-02-1/F1.

     ### F1 [high] Missing 404 when order is not found
     - where: internal/orders/handler.go:42
     - why: the contract says 404, the code returns 500
     - fix: (optional, one or two lines) return ErrNotFound from the store

     Levels:
     - high:   wrong behavior, breaks the spec, plan, or contract, security, data loss.
               Blocks the task. Only a human can mark it wontfix,
               or defer it out of the slice.
     - medium: a real problem that does not break anything yet.
               Must be fixed, or the worker gives a reason.
     - low:    nice to have. Optional. -->

## Requirements check
<!-- Integration reviews only. Other kinds: write "(none)".
     Every requirement in spec.md, checked against the finished slice.

     - R1: ok
     - R2: ok
     - R3: fails, see F2
     - R4: manual, goes to the preview checklist -->

## Needs human decision
<!-- Questions only the human can answer, and why each one matters.
     Mostly for plan reviews. Anything here makes the verdict needs_human. -->
