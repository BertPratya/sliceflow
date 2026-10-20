---
slice: {{slice}}
created: {{date}}
based_on: { spec: {{spec revision, e.g. R001}} }
touches:
  - {{a project file this slice changes, e.g. api/openapi.yaml}}
---

<!-- HOW TO FILL THIS TEMPLATE

     A plan says HOW the approved spec will be built: the parts that
     change, the flows, the decisions, and the tasks. Building details
     (exact code, file-by-file steps) go in the task files.

     The plan never adds scope. Anything not in the spec needs a spec
     change first. If the spec leaves out something the user sees (an
     output layout, an exit code, a message), don't decide it here: add
     an Open question, so the spec is changed.

     The human approves the plan. Write plain English, short sentences,
     one idea per sentence.

     flow reads this file. Keep every marker exactly as shown: the
     headings, the numbers (**TD1.**), the labeled lines (Covers:, Why:),
     and the YAML blocks. Don't rename or reorder them.

     - Replace every {{...}} with real text. The text inside says what
       goes there. flow fills slice and date itself. If they
       still show {{...}}, fill them yourself.
     - Each section has a comment like this one with its rules. Follow it.
     - Add ### subsections freely inside a section. The ## sections are
       fixed.
     - Repeat a block as often as needed: flows, rules, decisions, tasks.
     - A required section with nothing to say gets one line: None.
     - Two sections are optional: Risks and rollout, and Parked. Delete
       them when they don't apply.
     - When you are done, delete every comment and check that no {{ is
       left. A plan with a comment or a {{ left in it is not finished.
     - To change a plan later, read this template again first. The
       finished plan no longer has these rules in it.

     The front matter:
     - based_on: the approved spec revision this plan builds on.
     - A stacked slice (one that waits for another slice) adds a line:
       after: { slice: <that slice's name>, plan: <its revision> }
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
     failure paths. Each one starts with a Covers: line, then lists the
     steps. Add a Mermaid diagram for complex flows.

     Rules that apply across several flows (expiry, retries, locking,
     security) go under ### Shared rules, one #### subsection each, also
     with a Covers: line.

     - Number the flows Flow 1, Flow 2, ...
     - Covers: names spec requirements, e.g. Covers: R1, R5
     - Every requirement in the spec is covered by at least one flow or
       shared rule.

     Example:
     ### Flow 1: Add a task with a due date

     Covers: R1, R4

     1. The add command reads --due and checks the date.
     2. A bad date stops the command before anything is saved.
     3. The store saves the task with its due date. -->

### Flow 1: {{name}}

Covers: {{R1, R2}}

1. {{step}}

### Shared rules

#### {{rule name}}

Covers: {{R5}}

{{the rule}}

## Interfaces

<!-- Every way the outside world talks to the system that this slice
     touches. One ### subsection per kind: CLI, HTTP API, Events,
     Settings, UI routes.

     - If a contract file exists (e.g. OpenAPI), list the operations and
       link the file. Never copy request or response details.
     - If there is no contract file (e.g. a CLI or settings), describe it
       fully here: commands, flags, output, exit codes.
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
     test tools, fakes, and test data are used. manual checks are not
     listed here: they go in preview.md.

     Example:
     | Check | How |
     |---|---|
     | unit | Go tests; a fake clock for "today" |
     | e2e | build the binary, run it in a temporary folder | -->

| Check | How |
|---|---|
| {{unit, api, or e2e}} | {{test tools, fakes, test data}} |

## Decisions

<!-- Technical decisions, numbered TD1, TD2, ...

     - Each one has the choice, Why:, and Not chosen: (the main other
       option, and why not).
     - To replace a decision from an earlier slice, add Replaces: with
       that slice's name and the decision's number, e.g. checkout/TD2.
       Otherwise remove the whole Replaces: line.
     - A decision that goes against a project rule says so in Why:.
     - Only technical decisions. Don't repeat the spec's decisions (SD1,
       ...); they are already approved.

     Example:
     - **TD1.** Store the due date as text in the form YYYY-MM-DD.
       Why: it sorts correctly as text and needs no time zone.
       Not chosen: a timestamp, because a due date has no time of day. -->

- **TD1.** {{The choice.}}
  Why: {{the reason}}
  Not chosen: {{the main other option, and why not}}
  Replaces: {{an earlier decision, e.g. checkout/TD2}}

## Tasks

<!-- The task list. flow creates one task file per task from this block,
     and checks it before any file exists. This block is the only place
     the task structure lives: the task files hold only what to build
     and when it is done.

     One YAML block, with one entry per task:
     - id: two digits in quotes, "01", "02", ... in order. Quotes are
       needed: without them YAML reads 01 as a number.
     - name: short, lowercase, words joined by -. The task file is
       tasks/<id>-<name>.md.
     - covers: the spec requirements this task builds, e.g. [R1, R2].
       Every requirement appears in at least one task. A requirement may
       be in more than one task, when each task builds a part of it,
       e.g. the rule in one task and the command in another. Each task
       proves its own part.
     - depends_on: ids of the tasks whose result this one needs, e.g.
       ["01"]. No loops. Empty: []
     - touches: the files and folders this task may change. A folder
       ends with /. Keep it narrow. Two tasks that can run at the same
       time never touch the same file. Tasks that must change the same
       file: order them with depends_on.
     - risk: core, normal, or low. core and normal tasks are
       code-reviewed. low tasks are not: use it only for small, safe
       changes, e.g. docs or a rename.

     One task is one reviewable commit. Too big to review in one go?
     Split it.

     Example:
     ```yaml
     tasks:
       - id: "01"
         name: store-due-dates
         covers: [R1, R2]
         depends_on: []
         touches: [internal/store/]
         risk: core
       - id: "02"
         name: list-by-due-date
         covers: [R6, R7]
         depends_on: ["01"]
         touches: [cmd/todo/list.go, internal/list/]
         risk: normal
     ``` -->

```yaml
tasks:
  - id: "01"
    name: {{short-name}}
    covers: [{{R1, R2}}]
    depends_on: []
    touches: [{{folder/ or file}}]
    risk: {{core, normal, or low}}
```

## Risks and rollout

<!-- Optional. Delete this section if not needed.
     What could go wrong, and how it is reduced. Migration order, feature
     flags, compatibility with older versions. -->

{{A risk, and how it is reduced.}}

## Parked

<!-- Optional. Delete this section if not needed.
     Ideas found while planning that belong to a later slice.

     When the slice is done, flow copies each idea into backlog.md, word
     for word, with where it came from. So:
     - One - bullet per idea. Start each idea with what it is.
     - Lines indented under a bullet belong to that idea.
     - No sub-bullets, headings, or code blocks.

     Example:
     - Show overdue tasks in red.
     - A command that lists only overdue tasks.
       It could also take a number of days. -->

- {{An idea for later.}}

## Open questions

<!-- Questions for the human, numbered Q1, Q2, ... This section must say
     None. before the plan is submitted. -->

- **Q1.** {{A question for the human.}}
