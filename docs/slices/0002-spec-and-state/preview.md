---
slice: 0002-spec-and-state
---

# Preview

## Setup

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

### P1. See where a new spec stands

Covers: S1

1. In the worktree, run `sliceflow slice status`.
   Expect: the lines of S1, ending with `Spec:     draft` and
   `Next:     write spec.md, then: sliceflow spec submit`.
2. Run `sliceflow slice status --json`.
   Expect: one JSON line with `"docs":[{"doc":"spec.md","status":"draft"}]`
   and `"next":"write spec.md, then: sliceflow spec submit"`.

### P2. Check a spec with problems

Covers: S6

1. Copy `docs/slices/0001-slice-and-events/spec.md` from the sliceflow
   repo over `docs/slices/login-google/spec.md`, and change its front
   matter to `slice: login-google`.
2. Delete R7's `Check:` line, and write `- **Q1.** Is this right?`
   under `## Open questions`, in place of `None.`.
3. Run `sliceflow spec validate`.
   Expect: the line `spec.md:<R7's line>: error: R7 has no Check: line`, the warning
   `1 open question left (Q1)`, notes for R11, R40, and R43, and the
   summary `1 error, 1 warning, 3 notes`; the exit code is 1
   (`echo $?`).
4. Put R7's `Check:` line back, and run `sliceflow spec validate --strict`.
   Expect: the summary `0 errors, 1 warning, 3 notes`, and the exit code
   is 1.

### P3. Submit a spec, refused and then accepted

Covers: S14, S15

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

Covers: S26

1. Run `sliceflow slice title login-google "Sign in with Google"`.
   Expect: `Title of login-google is now "Sign in with Google".`
2. Run the same command again.
   Expect: `Error: the title of login-google is already "Sign in with Google". Nothing was changed.`
   and the exit code is 1.
3. Run `sliceflow slice list` in the test repo.
   Expect: the row of login-google with the title `Sign in with Google`.

### P5. Read the spec template's rules

Covers: S30

1. Run `sliceflow template show spec`.
   Expect: the whole spec template with its rule comments, starting with
   `---` and `slice: {{slice}}`.
2. Run `sliceflow template show plan`.
   Expect: `Error: unknown template "plan". Known templates: spec.`
   and the exit code is 1.
