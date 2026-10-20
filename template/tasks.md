---
slice: {{slice}}
based_on: { spec: {{spec revision, e.g. R001}} }
---

<!-- HOW TO FILL THIS TEMPLATE

     tasks.md holds every task of the slice: one section per task. Write
     it after plan.md is designed and its Open questions say None. The
     human approves plan.md and tasks.md together.

     Each task is built by one worker: an agent with a fresh context.
     flow gives each worker a brief, built from the links in this file:
     - its own task's section
     - the exact text of the requirements in covers
     - the spec's scenarios whose Shows: names one of those requirements
     - the spec's Key concepts
     - the exact text of the plan items in follows
     The worker may also open the full spec and plan if something is
     unclear. So don't copy any of these here; name them, e.g. R5, TD2.

     flow reads this file. Keep every marker exactly as shown: the task
     headings (## Task 01: name), the ### headings inside a task, the
     YAML block, and the labeled lines (Proves:, Scenario:, Done in:). Don't rename
     or reorder them. A task's section runs from its ## Task heading to
     the next one. In Files, Interfaces, and Done when, put the path,
     name, or test name in `backticks`, then a colon. A labeled line may
     go on over more lines: indent the extra lines 2 spaces more than the
     label.

     What to write here, and what to leave to the worker:
     - Fix what others rely on: the files, the names and signatures
       other tasks use, the order of the work, the rules, and the tests.
       Tasks may run at the same time, so these must agree.
     - Leave the code inside the functions to the worker: private
       helpers, local names, loops, data structures, the test code. The
       worker sees the real code; this file is written before it.
     - Don't write code in this file. A signature is fine; a function
       body is not.
     - Be concrete. Name real files, functions, types, and commands.
       "Add validation" is not enough; "add parseDue() in
       internal/task/due.go, which returns an error for a bad date" is.

     The tasks together:
     - Every requirement in the spec is in at least one task's covers.
       A requirement may be in more than one task, when each task builds
       a part of it, e.g. the rule in one task and the command in
       another. Each task proves its own part.
     - A requirement with Check: manual is in a task's covers too: the
       task builds it, and preview.md proves it.
     - Every scenario in the spec (S1, S2, ...) becomes at least one e2e
       test in some task's Done when, with the scenario's exact values
       and a Scenario: line. flow checks this.
     - No dependency loops.
     - Two tasks that can run at the same time never touch the same
       file. Tasks that must change the same file: order them with
       depends_on.
     - One task is one reviewable commit: about an hour of work, or a
       few hundred lines. Bigger? Split it into two tasks.

     Changing tasks:
     - Task ids are two digits: 01, 02, ... Before the first approval,
       number them top to bottom. After it, never renumber and never
       reuse an id: a removed task leaves its number unused, and a new
       task goes at the end with the next free number. The order of work
       comes from depends_on, not from the order in this file.
     - To add, remove, or rename tasks, edit this file. There is nothing
       else to update.

     Filling the template:
     - Replace every {{...}} with real text. flow fills slice and
       based_on itself. If they still show {{...}}, fill them yourself.
     - Copy the whole ## Task block once per task.
     - Each part has a comment like this one with its rules. Follow it.
     - Three parts of a task are optional: Interfaces, Follow, and
       Rules. Delete them when they don't apply. The other parts are
       required; one with nothing to say gets one line: None.
     - When you are done, delete every comment and check that nothing of
       this template is left: no comment, and no {{...}} placeholder,
       even inside code. A {{ or a comment marker that you write yourself
       inside a code block or `inline code` is real content. -->

# Tasks

## Task 01: {{short-name}}

<!-- The heading: ## Task, the id, a colon, and a short name: lowercase,
     words joined by -.

     Then one YAML block with the task's structure:
     - covers: the spec requirements this task builds, e.g. [R1, R2].
     - depends_on: ids of the tasks whose result this one needs, in
       quotes, e.g. ["01"]. Quotes are needed: without them YAML reads
       01 as a number. None: []
     - touches: the files and folders this task may change, its test
       files included. A folder ends with /. Keep it narrow. flow
       rejects a change outside it.
     - follows: the plan items this task must follow, by name: a
       decision (TD2), a shared rule or an interface by its heading
       (Locking, CLI), or Testing. flow checks each name exists, and puts
       its exact text into the worker's brief. None: []
     - risk: core, normal, or low.
       - core: other tasks build on it, it changes stored data or a file
         format, it touches security, or it changes a contract or a
         shared interface.
       - normal: any other code change.
       - low: no change in behavior, e.g. docs, a rename, or comments.
       core and normal tasks are code-reviewed; low tasks are not.

     Example:
     ## Task 02: list-by-due-date

     ```yaml
     covers: [R6, R7]
     depends_on: ["01"]
     touches: [cmd/todo/list.go, cmd/todo/list_test.go, internal/list/,
       e2e/list_by_due_date_test.go]
     follows: [TD1, Testing]
     risk: normal
     ``` -->

```yaml
covers: [{{R1, R2}}]
depends_on: []
touches: [{{folder/ or file}}]
follows: [{{TD1, a shared rule, Testing}}]
risk: {{core, normal, or low}}
```

### Goal

<!-- 2-4 sentences. What exists after this task that did not exist
     before. Name the hardest part, if there is one. -->

{{What exists after this task that did not exist before. The hardest part, if there is one.}}

### Steps

<!-- The approach, in the order the worker should do it. One step per
     line, each naming the file, function, or type it changes. The worker
     still writes the code; the steps say what to build, not every line.

     Example:
     1. Add a Due field (string, YYYY-MM-DD, empty means none) to Task in
        internal/task/task.go.
     2. Add parseDue() in internal/task/due.go. It accepts only a real
        date in the form YYYY-MM-DD and returns an error otherwise.
     3. In cmd/todo/add.go, read --due, call parseDue(), and stop before
        saving when it fails.
     4. Write the tests listed under Done when. -->

1. {{A step of the approach, naming the file or function.}}

### Files

<!-- Every file this task creates or changes, with what goes in it.
     All of them must be inside the task's touches. Test files count:
     list them too. Each task puts its e2e tests in its own file, where
     the plan's Testing section says e2e tests live, so tasks that run at
     the same time never share a test file.

     Example:
     - `internal/task/due.go`: new. parseDue().
     - `internal/task/due_test.go`: new. Unit tests for parseDue().
     - `cmd/todo/add.go`: changed. The --due flag.
     - `e2e/add_due_test.go`: new. The e2e tests for S1 and S2. -->

- `{{path}}`: {{new or changed, and what goes in it}}

### Interfaces

<!-- Optional. The exact names and signatures this task creates that
     other tasks or other code will use. Other tasks may be built at the
     same time, so they rely on these names being exactly right. Also
     list names from earlier tasks that this task must use.

     Start each one with what this task does with it:
     - creates: a new name.
     - changes: an existing name gets a new signature or behavior.
     - uses: a name from earlier tasks or existing code.

     Example:
     - `func parseDue(s string) (string, error)`: creates. Task 02 uses
       it.
     - `task.Task.Due`: creates. Task 02 sorts by it. -->

- `{{name or signature}}`: {{creates, changes, or uses; what it does; who else uses it}}

### Follow

<!-- Optional. Existing code to copy the style from, so the new code
     looks like the code around it.

     Example:
     - Read flags the same way as cmd/todo/list.go does. -->

- {{Existing code to copy, and what to copy from it.}}

### Rules

<!-- Optional. Rules that apply to this task and exist nowhere else.
     A rule already in the plan goes in follows instead: flow gives the
     worker its exact text. Don't copy it here.

     Example:
     - Errors start with "Error: " and end with "Nothing was changed."
       when nothing was saved. -->

- {{A rule that applies to this task.}}

### Done when

<!-- The tests that prove the task works. One bullet per test: its name,
     what it checks, then a Proves: line naming the requirements it
     proves. Name Testing in follows, so the worker gets the plan's
     helpers and fakes.

     Each test states its input and its expected result, with real
     values. "Tests the date check" is not enough; "'2026-02-30' returns
     an error" is.

     Which level:
     - A scenario from the spec becomes an e2e test with the scenario's
       exact values. Add a Scenario: line naming it, under Proves:, e.g.
       Scenario: S2
     - Every other case, e.g. edge cases and rule details, goes in a
       unit test. Keep e2e tests few; put tests as low as they can go.

     Every requirement this task covers is in at least one Proves: line.
     The only exception: a requirement whose Check: is manual and
     nothing else. preview.md proves that one. With Check: e2e, manual,
     the e2e part still needs a test here.

     flow also runs all project checks before the task is handed in, so
     don't list them here. flow checks that each test named here exists
     in a file inside the task's touches, and, where the project's test
     tool reports test names, that it ran and passed.

     Existing tests must keep passing. Change an existing test only when
     this task changes what it tests, and list its file under Files.

     Example:
     - `TestParseDueRejectsBadDate`: "25/10/2026" and "2026-02-30"
       return an error; "2026-10-25" returns no error.
       Proves: R3
     - `TestDueUnknownTask`: `todo due 9 2026-11-01` with no task 9
       prints "Error: task 9 not found." and changes nothing.
       Proves: R5
       Scenario: S2 -->

- `{{TestName}}`: {{what the test checks}}
  Proves: {{R5}}

### Out of scope

<!-- What this task must not do, even if it looks related, and where it
     is done instead. Done in: is one or more tasks, e.g. task 02, or
     task 02, task 04. Anything not done in this slice: not in this
     slice.

     Example:
     - Sorting by due date. Done in: task 02
     - Due times. Done in: not in this slice -->

- {{What this task must not do.}} Done in: {{task 03, or: not in this slice}}
