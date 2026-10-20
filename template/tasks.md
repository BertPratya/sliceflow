---
slice: {{slice}}
based_on: { spec: {{spec revision, e.g. R001}} }
---

<!-- HOW TO FILL THIS TEMPLATE

     tasks.md holds every task of the slice: one section per task. Write
     it after plan.md is designed and its Open questions say None. The
     human approves plan.md and tasks.md together.

     Each task is built by one worker: an agent with a fresh context.
     flow gives each worker only its own task's section, plus the exact
     text of the requirements that task covers. So each section must make
     sense on its own. Don't copy requirements here; name them, e.g. R5.

     flow reads this file. Keep every marker exactly as shown: the task
     headings (## Task 01: name), the ### headings inside a task, the
     YAML block, and the labeled lines (Proves:, Done in:). Don't rename
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
     - touches: the files and folders this task may change. A folder
       ends with /. Keep it narrow. flow rejects a change outside it.
     - risk: core, normal, or low. core and normal tasks are
       code-reviewed. low tasks are not: use it only for small, safe
       changes, e.g. docs or a rename.

     Example:
     ## Task 02: list-by-due-date

     ```yaml
     covers: [R6, R7]
     depends_on: ["01"]
     touches: [cmd/todo/list.go, internal/list/]
     risk: normal
     ``` -->

```yaml
covers: [{{R1, R2}}]
depends_on: []
touches: [{{folder/ or file}}]
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
     All of them must be inside the task's touches.

     Example:
     - `internal/task/due.go`: new. parseDue() and its tests' helpers.
     - `cmd/todo/add.go`: changed. The --due flag. -->

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

<!-- Optional. Rules from the plan or the project that apply to this
     task and are easy to miss: error format, locking, security, limits.

     Example:
     - Errors start with "Error: " and end with "Nothing was changed."
       when nothing was saved. -->

- {{A rule that applies to this task.}}

### Done when

<!-- The tests that prove the task works. One bullet per test: its name,
     what it checks, then a Proves: line naming the requirements it
     proves. The plan's Testing section says which helpers and fakes to
     use.

     Every requirement this task covers is in at least one Proves: line.
     The only exception: a requirement whose Check: is manual and
     nothing else. preview.md proves that one. With Check: e2e, manual,
     the e2e part still needs a test here.

     flow also runs all project checks before the task is handed in, so
     don't list them here.

     Existing tests must keep passing. Change an existing test only when
     this task changes what it tests, and list its file under Files.

     Example:
     - `TestParseDueRejectsBadDate`: "25/10/2026" and "2026-02-30"
       return an error.
       Proves: R3 -->

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
