---
name: flow-plan
description: Turns an approved spec into an approved plan, contract changes, and tasks, then gets them reviewed and approved together. Use when a slice's spec is approved, when a slice is in plan, or when an approved plan must change.
---

# flow-plan

You are the lead. Follow these steps in order. Each step ends when its
"Done when" is true. If `flow` does not exist yet, use `bootstrap.md` in
this folder wherever a step names a `flow` command.
To change a plan that is already approved, use `change.md` instead.

## 1. Check it is ready

Read the approved spec. Read the slice's `notes.md`, if it exists:
technical ideas from the spec interview. They are input, not decisions.

Done when: the spec is approved, and you have read it and `notes.md`.

## 2. Learn the code

Facts are your job. Look up:
- the modules and patterns in the parts this slice touches
- how those parts are tested today
- the project's contracts and data model
- earlier slices' technical decisions (TD…), so you do not contradict them by accident
- if the slice is stacked: the earlier slice's plan and contracts on this
  branch. Record its plan revision in `based_on` (`after`).
- the test dependencies and known gaps in the stack profile

Done when: you can name every part that will change, and how each one is tested today.

## 3. Settle the hard decisions with the human

Make small technical choices yourself, and record each one as a TD.
The human decides how involved they are: if they want to settle more of
the plan with you, play back the approach, the flows, and the task split
for their agreement before you write the tasks. By default, ask the human
only about a decision that is all three:
- **hard to reverse:** changing it later is expensive
- **surprising:** a future reader would ask "why this way?"
- **a real trade-off:** there are genuine alternatives

Ask in numbered rounds of at most about 5 questions, each with your
recommendation, worded so that "yes" accepts it.

Also propose the track, using the criteria in `AGENTS.md` section 2.

Done when: the human has answered every decision that passes the test,
and agreed on the track.

## 4. Try out risky unknowns (optional)

If something is truly unknown (how a library behaves, whether something is
fast enough), build a throwaway prototype outside the slice branch.
Keep only the result, written as a TD.

Done when: every risky unknown has a result written as a TD.

## 5. Write plan.md

Fill `plan.md` from the template, following the rules in its comments.
When you fill a section, delete its guidance comment. The template keeps it.
The ones most often missed:
- The plan never adds scope. Anything not in the spec needs a spec change (`flow-spec`).
- Every flow and shared rule starts with `Covers: R…`.
- Interfaces list and link the contract files. They never copy them.
- Every decision says why, and the main option rejected.
- `touches` lists every contract, data model, or other living doc the slice
  changes (`AGENTS.md` section 2). Code goes in the tasks' `touches`.
- If the slice depends on a known gap (e.g. database tests are skipped), say how
  the affected requirements are checked instead, usually in the preview.

Done when: every requirement is covered by a flow or a shared rule,
and Open questions is empty.

## 6. Change the contracts

Edit the project's contract files on the slice branch, before writing
tasks. Tasks build against them. List each one in the plan's `touches`.
Only you edit these files, and only now. No task's `touches` may include them.

Done when: every Interface in the plan matches the contract files.

## 7. Split into tasks

Write one file per task, `tasks/NN-<short-name>.md`, from the task template.
Delete each guidance comment once its section is filled: workers and
reviewers read task files many times.

- **Order:** setup first (checks for any new part, see `workflow/formats/lifecycle.md`),
  then contract, backend, frontend, integration, docs. A `contract` task
  builds what comes from the contract files (generated code, contract tests),
  never the files themselves.
- **Size:** one task is one reviewable commit, at most about 15-20 tests.
- **Self-contained:** a worker reads only the task file and what it links to.
  If it would need this conversation, the task is incomplete.
- **Coverage:** every requirement is covered by at least one task. Every
  requirement a task covers has at least one row in its "Done when".
- **Copies:** requirements in a task are copied word for word from the spec.
- **`touches`:** narrow. On the complex track, backend and frontend go in separate tasks.
- **`risk: core`:** anything touching auth, stored data, money, concurrency, or a contract.
- **`depends_on`:** only real dependencies: tasks whose result this task needs.
  `flow` may run tasks with no dependency between them at the same time,
  unless their `touches` overlap. `depends_on` also marks stale tasks.

Done when: every requirement is covered by a task, and every task follows the rules above.

## 8. Self-check

Run `flow validate`. Fix every error it reports.
Then read each task once as a worker would: only the task and what it links to.
Anything a worker would have to guess, fix now.

Done when: `flow validate` passes and no task needs guessing.

## 9. Plan review

Run `flow doc submit plan`. This starts the plan review: a fresh plan
reviewer checks the spec, plan, contracts, and tasks together.

You wrote the plan, so you fix its findings. Fix a finding alone when it
needs no decision from the human. Take it to the human, with your
recommendation, when it is under "Needs human decision", when the fix is a
hard decision (the test in step 3), or when you want to mark a high finding
wontfix or defer it out of the slice. A spec problem needs `flow-spec` (`change.md`).

For each finding run `flow finding fixed | deferred | wontfix` with a note.
Then submit again for the next round. In round 2 and later, the reviewer
checks your fixes, and adds a new finding only if it is high or was caused
by a fix.

Done when: the review passes.

## 10. Approval

Show the human a short summary:
- the approach in a few sentences, and the track
- the tasks, in order, with what each one builds
- the hard decisions
- what changes in the contracts
- the risks

Show the full documents or the diff when asked.
The human approves with `flow approve plan`. It approves the plan, the tasks,
and the contract changes together. Run it only when they have just told you to.

Done when: the plan is approved. Continue with `flow-implement`.
