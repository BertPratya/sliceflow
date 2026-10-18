---
task: "{{task}}"
slice: {{slice}}
title: {{title}}
layer: {{layer}}              # default: setup | contract | backend | frontend | integration | docs (a project rule may name others)
depends_on: []                # tasks whose result this one needs, e.g. ["01", "03"]
covers: []                    # spec requirements, e.g. [R5, R6]
risk: normal                  # core | normal | low  (core = always code-reviewed)
based_on: { plan: {{plan_revision}} }
touches:                      # files and folders this task may change. Keep it narrow.
  # - internal/server/pairing/                  folder (ends with /)
  # - internal/server/server.go                 one file
  # - internal/server/db/migrations/0002_*.sql  pattern
  # Generated code and lock files listed in the stack profile are always allowed.
  # Unrelated areas? Split the task. More split rules: flow-plan step 7 and the project's rules.
---

# Task {{task}}: {{title}}

<!-- A task must make sense on its own. The worker reads only this file,
     spec.md, plan.md, the project files it links to, and the code.
     If it needs earlier conversation to understand, this file is incomplete.

     One task = one reviewable commit. Size: the default in flow-plan step 7,
     unless a project rule replaces it. -->

## Goal
<!-- 2-4 sentences. What exists after this task that did not exist before.
     Mention the hardest part, if there is one. -->

## Requirements
<!-- The spec requirements listed in "covers", copied word for word.
     If this copy and the spec differ, the spec wins.
     If the task covers only part of a requirement, say which part:
     *This task provides the link and the time left.* -->

## Context
<!-- Everything the worker needs that is not obvious from the code.
     Use ### subsections freely, for example:
     ### What exists            names, types, helpers from earlier tasks
     ### Rules                  from the plan: expiry, secrets, error format
     ### Where the code goes
     ### Tests and test data    how to run them, fixtures, database setup -->

## Done when
<!-- The tests that prove the task works. One row per test.
     Every requirement in "covers" needs at least one row.

     | Test | Proves |
     |---|---|
     | TestStartReturnsKeyAndLink: 201, 43-char key, link, 600s left | R5 |
     | TestSimultaneousClicks: 50 rounds, exactly one click wins | R10 |

     Also: the project checks pass (see the stack profile in AGENTS.md). -->

## Out of scope
<!-- What this task must NOT do, and which task does it instead.
     Example: the approval page (task 03), the worker (tasks 05-07). -->
