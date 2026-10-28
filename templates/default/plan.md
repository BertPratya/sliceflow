---
slice: {{slice}}
created: {{date}}
touches:
  - {{a project file this slice changes, e.g. api/openapi.yaml}}
---

<!-- HOW TO FILL THIS TEMPLATE

     A plan says HOW the spec will be built: the parts that
     change, the flows, the interfaces, the data, the tests, and the
     decisions. The tasks go in tasks.md, written after this plan is
     designed and its Open questions say None. The human approves the
     spec, plan, and tasks together, in one approval.

     The plan never adds scope. Anything not in the spec needs a spec
     change first. If the spec leaves out something the user sees (an
     output layout, an exit code, a message), don't decide it here: add
     an Open question, so the spec is changed.

     Write plain English, short sentences, one idea per sentence. A
     point the lead decided alone ends with (lead's call); the human sees
     every one at approval.

     sliceflow reads this file. Keep every marker exactly as shown: the
     headings, the numbers (**TD1.**, **Q1.**), the labeled lines
     (Project rule:, Why:, Not chosen:, Cost:,
     Replaces:, Plan only:, Expect:), and the front matter. Don't rename or reorder
     them. A labeled line may go on over more lines:
     indent the extra lines 2 spaces more than the label.

     - Replace every {{...}} with real text. The text inside says what
       goes there. The lead fills slice and created when it copies
       this template.
     - Each section has a comment like this one with its rules. Follow it.
     - The ## sections are fixed. Add ### subsections freely, except in
       Flows and Interfaces, which have their own ### rules.
     - Repeat a block as often as needed: flows, rules, decisions.
     - A required section with nothing to say gets one line: None.
     - When you are done, delete every comment and check that no {{ is
       left. A plan with a comment or a {{ left in it is not finished.
       Every placeholder of this template counts, even inside code. A {{
       or a comment marker you write yourself inside a code block or
       `inline code` is real content.
     - To change a plan later, read this template again first. The
       finished plan no longer has these rules in it.

     The front matter:
     - A stacked slice (one that waits for another slice) adds a line:
       after: { slice: <that slice's name>, planning: <its approved revision> }
     - touches: the project-level files this slice changes, e.g. a
       contract or the data model doc. None? Write: touches: [] -->

# Plan

## Approach

<!-- The whole idea in a few sentences. A reader should understand the
     plan from this section alone.

     Then a table of the parts that change. A part is a module or
     package, e.g. internal/store. Change is one of: new, changed,
     removed.

     Then one line for new dependencies (libraries, tools, services).
     Each one needs the human's approval, so name it and say why it is
     needed. None? Write: New dependencies: None.

     Example:
     The due date is a new optional field on a task. Sorting and the
     overdue mark happen when the list is printed.

     | Part | Folder | Change |
     |---|---|---|
     | Task store | internal/store/ | changed |
     | List command | cmd/todo/ | changed |

     New dependencies: None. -->

{{The whole idea in a few sentences.}}

| Part | Folder | Change |
|---|---|---|
| {{part}} | {{folder}} | {{new, changed, or removed}} |

New dependencies: {{None, or each one with why it is needed}}

## Flows

<!-- One ### Flow subsection per main path, including the important
     failure paths. Every plan has at least one flow. Each one lists the
     steps in order. Add a Mermaid diagram for complex flows.

     Rules that apply across several flows (expiry, retries, locking,
     security) go under ### Shared rules, one #### subsection each. Give
     each rule a short heading that no other shared rule or interface
     uses: tasks name it in follows.

     Shared rules also list each project rule this slice must handle
     (from AGENTS.md or the project's rules), and how the slice handles
     it. Such a rule has a Project rule: line quoting where it comes
     from. List every project rule that applies: a missing one is a
     plan review finding.
     Tasks name these rules in follows like any other, so they reach the
     worker.

     - Number the flows Flow 1, Flow 2, ...
     - Which task builds which requirement is said in tasks.md, in each
       task's covers, not here.

     Example:
     ### Flow 1: Add a task with a due date

     1. The add command reads --due and checks the date.
     2. A bad date stops the command before anything is saved.
     3. The store saves the task with its due date.

     ### Shared rules

     #### Cross-platform paths

     Project rule: AGENTS.md, paths must work on Windows, macOS, and
       Linux.
     Paths are built with filepath.Join, never with "/". The tests run
     on all three systems in CI. -->

### Flow 1: {{name}}

1. {{step}}

### Shared rules

#### {{rule name}}

Project rule: {{where it comes from, for a project rule; otherwise delete this line}}

{{the rule}}

## Interfaces

<!-- Every way the outside world talks to the system that this slice
     touches. One ### subsection per kind: CLI, HTTP API, Events,
     Settings, UI routes. Tasks name a subsection in follows by its
     heading, so each heading is used once.

     - If a contract file exists (e.g. OpenAPI), list the operations and
       link the file. Never copy request or response details.
     - If there is no contract file (e.g. a CLI or settings), describe it
       fully here: commands, flags, output, exit codes. What the user
       sees comes from the spec. If the spec doesn't give it, don't
       decide it here: add an Open question.
     - None? Write: None. -->

### {{kind, e.g. CLI}}

{{What changes. Link the contract file if one exists.}}

## Data

<!-- What stored data changes: new or changed tables, fields, or local
     files. How existing data is moved over, if at all. Link the project
     data model doc for full details.
     None? Write: None. -->

{{The stored data that changes, and how old data is moved over.}}

## Testing

<!-- How the checks named in the spec will run. The spec says which kind
     of check each requirement gets. This section says how those checks
     are done, so every task uses the same helpers.

     One row per check kind the spec uses (unit, api, e2e). Say which
     test tools, fakes, and test data are used, and where the test files
     live. Each task puts its e2e tests in its own file, so say the
     folder and the file name form, e.g. e2e/<task_name>_test.go (the
     task's name with - written as _, as Go file names usually are).

     Then ### Trying it by hand: what the human needs at the preview.
     sliceflow builds the preview checklist itself, from the spec: the
     first scenario of each story, with its exact command and output,
     and each requirement with Check: manual. This part adds only what
     the spec cannot say, under two fixed #### headings:
     - Setup: numbered steps the human does once before trying, e.g.
       build the program, make a test repo. A step that gives a result
       has an Expect: line. Nothing needed? Write: None.
     - Manual checks: for each requirement with Check: manual, its
       number, then numbered steps, each one thing the human does, with
       an Expect: line. No manual requirement in the spec? Write: None.

     Example:
     | Check | How |
     |---|---|
     | unit | Go tests; a fake clock for "today" |
     | e2e | build the binary, run it in a temporary folder; one file per task: e2e/<task_name>_test.go |

     ### Trying it by hand

     #### Setup

     1. Run `go build -o todo ./cmd/todo`.
     2. Run `./todo add "first"`.
        Expect: "Added task 1."

     #### Manual checks

     - **R12.**
       1. Run `./todo list` in a terminal with colors.
          Expect: overdue tasks are shown in red. -->

| Check | How |
|---|---|
| {{unit, api, or e2e}} | {{test tools, fakes, test data}} |

### Trying it by hand

#### Setup

1. {{A setup step.}}
   Expect: {{what the human should see, or delete this line}}

#### Manual checks

- **{{R12}}.**
  1. {{A step the human does.}}
     Expect: {{what they should see}}

## Decisions

<!-- Technical decisions, numbered TD1, TD2, ...

     - Each one has the choice, Why:, and Not chosen: (the main other
       option, and why not).
     - Each one has a Cost: line: the downside we accept with this
       choice. With truly no downside, write Cost: None. A reviewer
       questions a None that looks wrong.
     - To replace a decision from an earlier slice, add Replaces: with
       that slice's name and the decision's number, e.g. checkout/TD2.
       Otherwise remove the whole Replaces: line.
     - A decision that goes against a project rule says so in Why:.
     - Only technical decisions. Don't repeat the spec's decisions (SD1,
       ...); they are already approved.
     - A worker sees a decision only when its task names it in follows
       (in tasks.md). So every decision is named by at least one task.
       A decision that no task's code depends on says why, in a
       Plan only: line, e.g. Plan only: it picks a dependency; no task's
       code depends on the choice. sliceflow warns about a decision that is
       neither followed nor plan only, and about one that is both.

     Example:
     - **TD1.** Store the due date as text in the form YYYY-MM-DD.
       Why: it sorts correctly as text and needs no time zone.
       Not chosen: a timestamp, because a due date has no time of day. -->

- **TD1.** {{The choice.}}
  Why: {{the reason}}
  Not chosen: {{the main other option, and why not}}
  Cost: {{the downside we accept, or None}}
  Plan only: {{why no task's code depends on it, or delete this line}}
  Replaces: {{an earlier decision, e.g. checkout/TD2}}

## Risks and rollout

<!-- What could go wrong, and how it is reduced: migration order,
     feature flags, compatibility with older versions. Always present.
     With truly no risk, write: None. A reviewer questions a None that
     looks wrong.

     Example:
     - Event logs written by older versions must still be read. The
       reader skips unknown fields, and a test reads an old log file. -->

{{A risk, and how it is reduced, or None.}}

## Open questions

<!-- Questions for the human, numbered Q1, Q2, ... An answered question
     is written into the plan and removed from here. A deferred question
     that does not block the build moves to docs/backlog.md.
     The tasks are written only when this section says None. -->

- **Q1.** {{A question for the human.}}
