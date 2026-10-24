---
slice: 0002-spec-and-state
---

<!-- HOW TO FILL THIS TEMPLATE

     The preview file says how the human tries the slice at the preview.
     The lead creates it when the plan starts, right after the spec
     review passes, and fills it along with the plan. The human approves
     it with the spec, plan, and tasks, in one approval.

     sliceflow reads this file. Keep every marker exactly as shown: the
     headings, the check numbers (### P1.), and the labeled lines
     (Covers:, Expect:). Don't rename or reorder them. A labeled line
     may go on over more lines: indent the extra lines 2 spaces more than
     the label.

     The fixed part comes from the spec alone, so it is written as soon
     as the spec review passes: the lead writes one ### P check, its
     number, title, and Covers: line, for each of these:
     - the first scenario of each story that is not removed, its main
       success case, e.g. Covers: S1. Tests prove the code does what the
       tasks say; this check lets the human see it does what they meant.
     - each requirement with Check: manual.
     - each success criterion (SC1, ...).
     When the spec changes later, the lead adds checks for new items,
     keeps the text already written, and tells the human about a check
     whose item was removed.

     The lead writes the setup, the steps, and the Expect: lines while
     writing the plan. More checks may be added; number them after the
     last.

     - Replace every {{...}} with real text. The lead fills slice when
       it copies this template.
     - Write plain English, short sentences. The human follows these
       steps by hand, so each step is one thing to do.
     - Results are not written here. The human tells the lead, who
       writes them in status.md → Preview.
     - When you are done, delete every comment and check that no {{ is
       left. Every placeholder of this template counts, even inside
       code. A {{ or a comment marker you write yourself inside a code
       block or `inline code` is real content.
     - A section with nothing to say gets one line: None. E.g. no setup
       needed: None. under Setup. Checks is never empty while the spec
       has a story. -->

# Preview

## Setup

<!-- What the human does once before the checks: build the program,
     make a test folder or test repo, add test data. Numbered steps.
     Each step that gives a result has an Expect: line. -->

1. Build sliceflow from the slice worktree:
   `go build -o <a temp folder>/sliceflow ./cmd/sliceflow`.
   Use that file as `sliceflow` in every step below.
   Expect: no output, and the file exists.
2. Make a test repo in a new empty folder: `git init -b main`, add one
   file, and `git commit -m first`.
   Expect: a repo on `main` with one commit.
3. In the test repo, run
   `sliceflow slice create login-google --title "Login with Google"`,
   then go into `.sliceflow/worktrees/login-google`.
   Expect: `Created slice login-google.`, and you are on branch
   `slice/login-google`.

## Checks

<!-- One ### subsection per check, numbered P1, P2, ...

     - Covers: names what this check proves: scenarios (S1),
       requirements (R12), and success criteria (SC1). The first
       scenario of each story that is not removed, every requirement
       with Check: manual, and every success criterion are in at least
       one Covers: line.
     - For a scenario, the steps and Expect: lines follow the scenario's
       exact command and output.
     - Then numbered steps. Each step is one thing the human does, with
       an Expect: line saying what they should see.

     Example:
     ### P1. Start a new slice

     Covers: S1

     1. Run `sliceflow slice new checkout` in the test repo.
        Expect: "Created slice checkout." and the path to its spec.md.

     ### P2. A plan runs with no lead action

     Covers: SC1

     1. Run `sliceflow run` on the test repo.
        Expect: all 5 tasks finish, and the lead is never called.
     2. Run `sliceflow status`.
        Expect: the slice is in the preview phase. -->

### P1. See where a new spec stands

Covers: S1

1. In the worktree, run `sliceflow slice status`.
   Expect: the lines of S1, ending with `Spec:     draft` and
   `Next:     write spec.md, then: sliceflow spec submit`.
2. Run `sliceflow slice status --json`.
   Expect: one JSON line with `"docs":[{"doc":"spec.md","status":"draft"}]`
   and `"next":"write spec.md, then: sliceflow spec submit"`.

### P2. Check a spec with problems

Covers: S5

1. Copy `docs/slices/0001-slice-and-events/spec.md` from the sliceflow
   repo over `docs/slices/login-google/spec.md`, and change its front
   matter to `slice: login-google`.
2. Delete R7's `Check:` line, and write `- **Q1.** Is this right?`
   under `## Open questions`, in place of `None.`.
3. Run `sliceflow spec validate`.
   Expect: an error for R7's missing `Check:` line, the warning
   `1 open question left (Q1)`, notes for R11, R40, and R43, and the
   summary `1 error, 1 warning, 3 notes`; the exit code is 1
   (`echo $?`).
4. Put R7's `Check:` line back, and run `sliceflow spec validate --strict`.
   Expect: the summary `0 errors, 1 warning, 3 notes`, and the exit code
   is 1.

### P3. Submit a spec, refused and then accepted

Covers: S13, S14

1. With the spec of P2 (still with Q1), run `sliceflow spec submit`.
   Expect: the problem lines, then
   `Submit refused: spec.md is not finished (0 errors, 1 warning). It stays a draft.`
   and the exit code is 1.
2. Run `git log --oneline -1`.
   Expect: `login-google: spec submit refused`.
3. Write `None.` under `## Open questions` again, and run
   `sliceflow spec submit`.
   Expect: the 3 notes, then `Submitted spec.md.`,
   `Review 1 started: reviews/spec-1.md (reviewer: placeholder)`,
   `Review 1 passed.`, and `Next: write the plan.`; the exit code is 0.
4. Run `git log --oneline -3`.
   Expect: `review 1 passed`, `review 1 started`, and `spec submitted`,
   newest first.
5. Open `docs/slices/login-google/reviews/spec-1.md`.
   Expect: the front matter of R25, and the line
   `No review was done: placeholder reviewer.`
6. Run `sliceflow slice status`.
   Expect: `Spec:     review_passed (review 1, placeholder)` and
   `Next:     write the plan`.

### P4. Set a new title

Covers: S25

1. Run `sliceflow slice title login-google "Sign in with Google"`.
   Expect: `Title of login-google is now "Sign in with Google".`
2. Run the same command again.
   Expect: `Error: the title of login-google is already "Sign in with Google". Nothing was changed.`
   and the exit code is 1.
3. Run `sliceflow slice list` in the test repo.
   Expect: the row of login-google with the title `Sign in with Google`.

### P5. Read the spec template's rules

Covers: S29

1. Run `sliceflow template show spec`.
   Expect: the whole spec template with its rule comments, starting with
   `---` and `slice: {{slice}}`.
2. Run `sliceflow template show plan`.
   Expect: `Error: unknown template "plan". Known templates: spec.`
   and the exit code is 1.
