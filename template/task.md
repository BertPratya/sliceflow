<!-- HOW TO FILL THIS TEMPLATE

     A task file tells one worker how to build one task. The worker is an
     agent with a fresh context. It must be able to build the task from
     this file alone, plus what flow gives it.

     flow creates this file from the plan's task list, as
     tasks/<id>-<name>.md. The task's structure lives only in that list:
     id, name, covers, depends_on, touches, risk. Don't repeat it here.
     When the worker starts, flow brief adds that entry and the exact text
     of the covered requirements from the approved spec. So don't copy
     requirements here either; name them, e.g. R5.

     flow reads this file. Keep every marker exactly as shown: the
     headings and the labeled lines (Proves:, Done in:). Don't rename or
     reorder them.

     - Be concrete. Name real files, functions, types, and commands.
       "Add validation" is not enough; "add parseDue() in
       internal/task/due.go, which returns an error for a bad date" is.
     - Replace every {{...}} with real text.
     - Each section has a comment like this one with its rules. Follow it.
     - Three sections are optional: Interfaces, Follow, and Rules. Delete
       them when they don't apply. The other sections are required; one
       with nothing to say gets one line: None.
     - When you are done, delete every comment and check that no {{ is
       left. A task with a comment or a {{ left in it is not finished.
     - One task is one reviewable commit: about an hour of work, or a few
       hundred lines. Bigger? Split it in the plan's task list. -->

# Task

## Goal

<!-- 2-4 sentences. What exists after this task that did not exist
     before. Name the hardest part, if there is one. -->

{{What exists after this task that did not exist before. The hardest part, if there is one.}}

## Steps

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

## Files

<!-- Every file this task creates or changes, with what goes in it.
     All of them must be inside the task's touches in the plan.

     Example:
     - internal/task/due.go: new. parseDue() and its tests' helpers.
     - cmd/todo/add.go: changed. The --due flag. -->

- `{{path}}`: {{new or changed, and what goes in it}}

## Interfaces

<!-- Optional. The exact names and signatures this task creates that
     other tasks or other code will use. Other tasks may be built at the
     same time, so they rely on these names being exactly right. Also
     list names from earlier tasks that this task must use.

     Example:
     - func parseDue(s string) (string, error): creates. Task 02 uses it.
     - task.Task.Due: creates. Task 02 sorts by it. -->

- `{{name or signature}}`: {{creates or uses; what it does; who else uses it}}

## Follow

<!-- Optional. Existing code to copy the style from, so the new code
     looks like the code around it.

     Example:
     - Read flags the same way as cmd/todo/list.go does. -->

- {{Existing code to copy, and what to copy from it.}}

## Rules

<!-- Optional. Rules from the plan or the project that apply to this
     task and are easy to miss: error format, locking, security, limits.

     Example:
     - Errors start with "Error: " and end with "Nothing was changed."
       when nothing was saved. -->

- {{A rule that applies to this task.}}

## Done when

<!-- The tests that prove the task works. One bullet per test: its name,
     what it checks, then a Proves: line naming the requirements it
     proves. Every requirement this task covers is in at least one
     Proves: line. The plan's Testing section says which helpers and
     fakes to use.

     flow also runs all project checks before the task is handed in, so
     don't list them here.

     Example:
     - TestParseDueRejectsBadDate: "25/10/2026" and "2026-02-30" return
       an error.
       Proves: R3 -->

- `{{TestName}}`: {{what the test checks}}
  Proves: {{R5}}

## Out of scope

<!-- What this task must not do, even if it looks related, and where it
     is done instead: another task, or a later slice.

     Example:
     - Sorting by due date. Done in: task 02 -->

- {{What this task must not do.}} Done in: {{task 03, or a later slice}}
