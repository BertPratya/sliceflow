---
slice: 0002-spec-and-state
title: The state machine, and the spec's steps
created: 2026-10-24
ideas: []
---

# Spec

## Summary

A lead checks a slice's spec with `sliceflow spec validate`, and hands it
in with `sliceflow spec submit`, which saves the step and runs a review.
sliceflow works out each slice's state by replaying its event log, so
`slice status` shows where the spec stands and what to do next.

## Why

In 0001, sliceflow only knows that a slice exists. The spec's state
(draft, in review, passed) is kept by hand in `status.md`, and the spec's
form is checked by eye. A lead can skip the check, or send an unfinished
spec to review, and the state can drift from what really happened. This
slice builds the state machine that every later step uses, and puts the
spec's first steps on it.

## User stories

### Story 1: See where the spec stands

A user, or an AI agent, sees the spec's state and the next step.

- **S1.** **Given** the slice `login-google` was just created with
  `--title "Login with Google"`, and I am in its worktree,
  **When** I run `sliceflow slice status`,
  **Then** sliceflow prints:
  ```
  $ sliceflow slice status
  Slice:    login-google
  Title:    Login with Google
  Phase:    planning
  From:     main @ 9dd7ebf
  Created:  2026-10-24 07:42 UTC
  Spec:     draft
  Next:     write spec.md, then: sliceflow spec submit
  ```
  Shows: R1, R2, R3, R29
- **S2.** **Given** the slice of S1,
  **When** I run `sliceflow slice status --json`,
  **Then** sliceflow prints one JSON object:
  ```
  {"slice":"login-google","title":"Login with Google","phase":"planning","from":{"branch":"main","commit":"<full commit id>"},"created":"2026-10-24T07:42:00Z","docs":[{"doc":"spec.md","status":"draft"}],"next":"write spec.md, then: sliceflow spec submit","warnings":[]}
  ```
  Shows: R31
- **S3.** **Given** line 4 of the log of `login-google` is a
  `review.finished`, but no review started,
  **When** I run `sliceflow slice status`,
  **Then** sliceflow prints the status without that line, with
  `Spec:     draft`, and prints on stderr
  `Warning: login-google: .events.jsonl line 4: review.finished is not allowed when spec.md is draft; skipped`
  and the exit code is 0.
  Shows: R4
- **S4.** **Given** the log of `login-google` ends with review 1, whose
  result is `changes`,
  **When** I run `sliceflow slice status`,
  **Then** the last two lines are:
  ```
  Spec:     draft (review 1 found changes: reviews/spec-1.md)
  Next:     fix spec.md, then: sliceflow spec submit
  ```
  Shows: R2, R29

### Story 2: Check a spec

The lead checks the spec's form as often as it likes, while writing it.

- **S5.** **Given** in `spec.md`, R7 has no `Check:` line, Open
  questions lists Q1 and Q2, and no scenario shows R8,
  **When** I run `sliceflow spec validate` in the worktree,
  **Then** sliceflow prints:
  ```
  $ sliceflow spec validate
  spec.md:42: error: R7 has no Check: line
  spec.md:44: note: no scenario shows R8
  spec.md:140: warning: 2 open questions left (Q1, Q2)
  1 error, 1 warning, 1 note
  ```
  and the exit code is 1, and no commit was made.
  Shows: R13, R14, R15, R16, R17, R37
- **S6.** **Given** the spec of S5, with R7's `Check:` line added,
  **When** I run `sliceflow spec validate`,
  **Then** sliceflow prints the note, the warning, and
  `0 errors, 1 warning, 1 note`, and the exit code is 0.
  Shows: R14
- **S7.** **Given** the spec of S6,
  **When** I run `sliceflow spec validate --strict`,
  **Then** sliceflow prints the same lines as in S6, and the exit code
  is 1.
  Shows: R14
- **S8.** **Given** a finished spec,
  **When** I run `sliceflow spec validate login-google` on `main`,
  **Then** sliceflow prints `0 errors, 0 warnings, 0 notes`, and the
  exit code is 0.
  Shows: R13
- **S9.** **Given** I am on `main`,
  **When** I run `sliceflow spec validate`,
  **Then** sliceflow prints on stderr:
  ```
  Error: not on a slice branch (you are on main).
  Give a name: sliceflow spec validate <name>
  See all slices: sliceflow slice list
  ```
  and the exit code is 1.
  Shows: R12
- **S10.** **Given** a code block in `spec.md` holds the line
  `name: {{slice}}`, and a sentence holds `<!--` as inline code,
  **When** I run `sliceflow spec validate`,
  **Then** neither is reported.
  Shows: R18
- **S11.** **Given** the worktree of `login-google` was removed,
  **When** I run `sliceflow spec validate login-google` on `main`,
  **Then** sliceflow prints on stderr
  `Error: slice login-google has no worktree: no worktree has the branch slice/login-google checked out. Nothing was changed.`
  and the exit code is 1.
  Shows: R11
- **S12.** **Given** the lead deleted `spec.md` in the worktree,
  **When** I run `sliceflow spec validate`,
  **Then** sliceflow prints on stderr
  `Error: .sliceflow/worktrees/login-google/docs/slices/login-google/spec.md not found. Nothing was changed.`
  and the exit code is 1.
  Shows: R19

### Story 3: Submit the spec

The lead hands in a finished spec, and it is reviewed at once.

- **S13.** **Given** the spec of S6, which still lists Q1 and Q2,
  **When** I run `sliceflow spec submit`,
  **Then** sliceflow prints:
  ```
  $ sliceflow spec submit
  spec.md:44: note: no scenario shows R8
  spec.md:140: warning: 2 open questions left (Q1, Q2)
  Submit refused: spec.md is not finished (0 errors, 1 warning). It stays a draft.
  ```
  and the exit code is 1,
  and the branch has one new commit, `login-google: spec submit refused`,
  holding `spec.md` as it is on disk and the log, with the new line:
  ```
  {"version":1,"seq":2,"event":"spec.submitted","at":"2026-10-24T09:10:00Z","data":{"result":"refused","errors":0,"warnings":1}}
  ```
  and `slice status` then shows:
  ```
  Spec:     draft (submit refused: 1 warning)
  Next:     fix spec.md, then: sliceflow spec submit
  ```
  Shows: R8, R20, R21, R22, R29
- **S14.** **Given** the spec of S13, with Q1 and Q2 settled,
  **When** I run `sliceflow spec submit`,
  **Then** sliceflow prints:
  ```
  $ sliceflow spec submit
  spec.md:44: note: no scenario shows R8
  Submitted spec.md.
  Review 1 started: reviews/spec-1.md (reviewer: placeholder)
  Review 1 passed.
  Next: write the plan.
  ```
  and the exit code is 0,
  and the branch has three new commits, one per event:
  1. `login-google: spec submitted`, commit `3f2a1c9`, holding `spec.md`
     and the log,
  2. `login-google: review 1 started`, holding the new file
     `reviews/spec-1.md` with only its front matter, and the log,
  3. `login-google: review 1 passed`, holding the filled
     `reviews/spec-1.md`, and the log,

  and the log has three new lines:
  ```
  {"version":1,"seq":3,"event":"spec.submitted","at":"2026-10-24T10:00:00Z","data":{"result":"accepted","errors":0,"warnings":0}}
  {"version":1,"seq":4,"event":"review.started","at":"2026-10-24T10:00:01Z","data":{"doc":"spec.md","round":1,"reviewer":"placeholder","file":"reviews/spec-1.md","commit":"3f2a1c9e5b7d4a6f8c0e2b4d6f8a0c2e4b6d8f0a"}}
  {"version":1,"seq":5,"event":"review.finished","at":"2026-10-24T10:00:01Z","data":{"doc":"spec.md","round":1,"result":"passed"}}
  ```
  and `slice status` then shows:
  ```
  Spec:     review_passed (review 1, placeholder)
  Next:     write the plan
  ```
  Shows: R8, R21, R23, R25, R26, R27, R28, R29
- **S15.** **Given** review 1 of S14 passed, and `spec.md` has not
  changed since,
  **When** I run `sliceflow spec submit`,
  **Then** sliceflow prints on stderr
  `Error: spec.md has not changed since review 1, which passed. Nothing was changed.`
  and the exit code is 1, and no commit was made.
  Shows: R20
- **S16.** **Given** review 1 of S14 passed, and then the lead changed
  R7 in `spec.md`,
  **When** I run `sliceflow slice status`,
  **Then** the last two lines are:
  ```
  Spec:     passed review 1, but spec.md was changed after it
  Next:     sliceflow spec submit
  ```
  Shows: R29, R30
- **S17.** **Given** the changed spec of S16,
  **When** I run `sliceflow spec submit`,
  **Then** sliceflow prints `Submitted spec.md.`,
  `Review 2 started: reviews/spec-2.md (reviewer: placeholder)`,
  `Review 2 passed.`, and `Next: write the plan.`,
  and `slice status` then shows `Spec:     review_passed (review 2, placeholder)`.
  Shows: R20, R25, R30
- **S18.** **Given** the log of `login-google` ends with a
  `review.started`, so `spec.md` is `in_review`,
  **When** I run `sliceflow spec submit`,
  **Then** sliceflow prints on stderr
  `Error: spec.md is in_review: review 1 has started, and has not finished. Nothing was changed.`
  and the exit code is 1.
  Shows: R20
- **S19.** **Given** the lock file `.sliceflow/locks/login-google.lock`
  holds pid 4312 and the time 10:02 UTC, and stays there,
  **When** I run `sliceflow spec submit`,
  **Then** sliceflow waits 10 seconds, prints on stderr
  `Error: slice login-google is busy: another sliceflow command holds .sliceflow/locks/login-google.lock (pid 4312, since 10:02 UTC). If no sliceflow command is running, delete that file.`
  and the exit code is 1, and no commit was made.
  Shows: R6, R7
- **S20.** **Given** a git pre-commit hook that only runs `exit 1`,
  **When** I run `sliceflow spec submit` on a finished spec,
  **Then** sliceflow prints on stderr
  `Error: could not save the event: git commit failed (exit code 1). Nothing was changed.`
  and the exit code is 1, and the log on disk is as it was committed,
  and no commit was made.
  Shows: R9, R10
- **S21.** **Given** the file `reviews/spec-1.md` is in the slice
  folder, but no review started,
  **When** I run `sliceflow spec submit` on a finished spec,
  **Then** sliceflow prints on stderr
  `Error: docs/slices/login-google/reviews/spec-1.md already exists, but review 1 has not started. Move it away, then submit again. Nothing was changed.`
  and the exit code is 1, and no commit was made.
  Shows: R24
- **S22.** **Given** the lead has a staged change to `src/app.go` in the
  worktree,
  **When** I run `sliceflow spec submit`,
  **Then** none of sliceflow's commits holds `src/app.go`, and the change
  is still staged afterwards.
  Shows: R8
- **S23.** **Given** a git pre-commit hook that refuses every commit
  whose message has `review`,
  **When** I run `sliceflow spec submit` on a finished spec,
  **Then** sliceflow prints `Submitted spec.md.`, then prints on stderr
  `Error: could not save the event: git commit failed (exit code 1). The steps before it were saved; spec.md is submitted.`
  and the exit code is 1, and `slice status` then shows:
  ```
  Spec:     submitted (review 1 not started)
  Next:     sliceflow spec submit, to start review 1
  ```
  Shows: R10, R29
- **S24.** **Given** the slice of S23, with that hook removed,
  **When** I run `sliceflow spec submit`,
  **Then** sliceflow does not check or submit `spec.md` again, and
  prints:
  ```
  $ sliceflow spec submit
  Review 1 started: reviews/spec-1.md (reviewer: placeholder)
  Review 1 passed.
  Next: write the plan.
  ```
  and the exit code is 0, and the branch has two new commits,
  `login-google: review 1 started` and `login-google: review 1 passed`.
  Shows: R20

### Story 4: Set the title

A user gives a slice a new title, at any time.

- **S25.** **Given** the slice of S1,
  **When** I run `sliceflow slice title login-google "Sign in with Google"`,
  **Then** sliceflow prints
  `Title of login-google is now "Sign in with Google".`
  and the exit code is 0,
  and the branch has one new commit, `login-google: title set`, holding
  only the log, with the new line:
  ```
  {"version":1,"seq":2,"event":"slice.titled","at":"2026-10-24T07:50:00Z","data":{"title":"Sign in with Google"}}
  ```
  and `slice list` then shows the row:
  ```
  login-google  Sign in with Google  planning  2026-10-24 07:42
  ```
  Shows: R32, R34
- **S26.** **Given** the title of S25,
  **When** I run `sliceflow slice title login-google "Sign in with Google"`,
  **Then** sliceflow prints on stderr
  `Error: the title of login-google is already "Sign in with Google". Nothing was changed.`
  and the exit code is 1.
  Shows: R33
- **S27.** **When** I run `sliceflow slice title login-google ""`,
  **Then** sliceflow prints on stderr
  `Error: the title must not be empty. Nothing was changed.`
  and the exit code is 1.
  Shows: R33
- **S28.** **When** I run `sliceflow slice title login-google`,
  **Then** sliceflow prints on stderr
  `Error: missing title. Run sliceflow --help.`
  and the exit code is 2.
  Shows: R36

### Story 5: Read the spec's rules

The lead reads the template's rules, which a filled spec no longer holds.

- **S29.** **When** I run `sliceflow template show spec`,
  **Then** sliceflow prints the built-in spec template with all its rule
  comments, starting with:
  ```
  ---
  slice: {{slice}}
  created: {{date}}
  ideas:
    - {{an idea file this slice builds on, e.g. docs/ideas/2026-10-23-due-dates.md; none: ideas: []}}
  ---
  ```
  and the exit code is 0.
  Shows: R35
- **S30.** **When** I run `sliceflow template show plan`,
  **Then** sliceflow prints on stderr
  `Error: unknown template "plan". Known templates: spec.`
  and the exit code is 1.
  Shows: R35
- **S31.** **When** I run `sliceflow spec validate --json`,
  **Then** sliceflow prints on stderr
  `Error: unknown flag --json. Run sliceflow --help.`
  and the exit code is 2.
  Shows: R36

## Key concepts

- **Spec state:** where `spec.md` stands: `draft`, `submitted`,
  `in_review`, or `review_passed`.
- **Move:** an event that R2 allows in the state at that point.
- **Replay:** applying the good events of the committed log, in order,
  from the first, to get the state.
- **Round:** the number of one review of `spec.md`: 1 for the first, then
  1 more for each new review. A round is never reused.
- **Reviewed commit:** the commit named by `commit` in a `review.started`.
  It holds the exact `spec.md` the reviewer read.
- **Changed after its review:** a `spec.md` that passed review, but whose
  file on disk now differs from the reviewed commit (R30).
- **The lock:** the file `.sliceflow/locks/<name>.lock`, which says that
  one command is writing an event of that slice (R6).

## Requirements

### The state

- **R1.** Every time a command reads or writes a slice, sliceflow shall
  work out the slice's state by replaying the good events of its
  committed log, read as R24, R26, and R27 of 0001-slice-and-events say.
  It shall keep the state nowhere else.
  Check: unit
- **R2.** sliceflow shall change the state only by these moves:

  | Event | Allowed when `spec.md` is | `spec.md` becomes |
  |---|---|---|
  | `slice.created` | only when no good `slice.created` came before it | `draft` |
  | `slice.titled` | any state | as it was; the title becomes the event's `title` |
  | `spec.submitted`, `accepted` | `draft` or `review_passed` | `submitted` |
  | `spec.submitted`, `refused` | `draft` or `review_passed` | as it was |
  | `review.started` | `submitted` | `in_review` |
  | `review.finished`, `passed` | `in_review` | `review_passed` |
  | `review.finished`, `changes` or `needs_human` | `in_review` | `draft` |

  Every event other than `slice.created` is allowed only after a good
  `slice.created`. A `review.finished` is allowed only when its `doc` and
  `round` are those of the last `review.started`.
  Check: unit
- **R3.** sliceflow shall show the phase `planning` once a good
  `slice.created` is replayed, and `unknown` before that. This replaces
  R25 of 0001-slice-and-events.
  Check: unit
- **R4.** If an event is not allowed by R2, then sliceflow shall skip it,
  keep the state as it was, and add the warning
  `.events.jsonl line <n>: <event> is not allowed when spec.md is <state>; skipped`.
  Before a good `slice.created`, the warning is
  `.events.jsonl line <n>: <event> comes before slice.created; skipped`.
  For a `review.finished` whose round does not match, it is
  `.events.jsonl line <n>: review.finished is for round <x>, but review <y> is running; skipped`.
  Warnings print as R28 of 0001-slice-and-events says.
  Check: unit, e2e
- **R5.** sliceflow shall read the new events with these fields in
  `data`, and check them as R26 of 0001-slice-and-events says:

  | Event | Fields in `data` |
  |---|---|
  | `slice.titled` | `title`: a string, not empty |
  | `spec.submitted` | `result`: `accepted` or `refused`; `errors`, `warnings`: whole numbers, 0 or more |
  | `review.started` | `doc`: `spec.md`; `round`: a whole number, 1 or more; `reviewer`: a string, not empty; `file`: a string, not empty; `commit`: 40 or 64 lowercase hex characters |
  | `review.finished` | `doc`: `spec.md`; `round`: a whole number, 1 or more; `result`: `passed`, `changes`, or `needs_human` |

  A value outside these is a wrong value.
  Check: unit

### Writing an event

- **R6.** Before a command writes an event, sliceflow shall take the
  slice's lock: it creates the file `.sliceflow/locks/<name>.lock` at the
  root of the main checkout, holding its process id and the time in UTC.
  With the lock held, it replays the committed log, checks the move by
  R2, and saves the event. Then it removes the lock, also when the
  command failed or was stopped with Ctrl-C. It holds the lock for one
  event only, and never while a reviewer runs. If a later event of the
  same command is no longer allowed by R2, because another command moved
  the state first, then sliceflow shall save no event, print
  `Error: another sliceflow command changed slice <name> first; spec.md is now <state>. The steps before it were saved.`
  and exit with code 1.
  Check: unit, e2e
- **R7.** If the lock file already exists, then sliceflow shall try again
  for up to 10 seconds, and then print
  `Error: slice <name> is busy: another sliceflow command holds .sliceflow/locks/<name>.lock (pid <pid>, since <HH:MM> UTC). If no sliceflow command is running, delete that file.`
  and exit with code 1. When the lock file cannot be read, `<pid>` and
  `<HH:MM>` are `unknown`.
  Check: e2e
- **R8.** sliceflow shall save each event as its own commit, on the branch
  `slice/<name>`, in the slice's worktree. The commit holds the log and
  the files that the event's requirement names, and nothing else. Other
  changes in the worktree, staged or not, stay as they were. It is a
  normal git commit, as R8 of 0001-slice-and-events says. sliceflow shall
  never make an empty commit.
  Check: e2e
- **R9.** sliceflow shall write the new log as the committed log plus the
  one new line. When the committed log does not end with `\n`, sliceflow
  first adds a `\n`, so the new event starts on a line of its own. The
  last line is kept as it is; when it is bad, it is skipped with a
  warning, as R26 of 0001-slice-and-events says. The new event's `seq` is
  1 more than the `seq` of the last line read without error, also when R4
  skipped that line, and 1 when there is none. The log file on disk shall
  never be half written, also when sliceflow stops in the middle of
  writing it.
  Check: unit
- **R10.** If the commit fails, then sliceflow shall put the log file back
  as it was committed, remove the files it created for that event, and
  unstage what it staged for it. It never changes `spec.md`. Then it
  prints
  `Error: could not save the event: <reason>. Nothing was changed.`
  with `<reason>`, and git's other lines, as R18 of 0001-slice-and-events
  says. Ctrl-C, or a stop signal from the system, before the commit is
  made counts as a failed commit, with the reason `interrupted`. When an
  earlier event of the same command was already saved, the last sentence
  is `The steps before it were saved; spec.md is <state>.` instead. The
  exit code is 1.
  Check: e2e
- **R11.** If no worktree has the branch `slice/<name>` checked out, then
  `spec validate`, `spec submit`, and `slice title` shall print
  `Error: slice <name> has no worktree: no worktree has the branch slice/<name> checked out. Nothing was changed.`
  and exit with code 1. These three commands check in this order, and
  stop at the first that fails: R12 and R13 of 0001-slice-and-events (a
  git repo, with commits), R14 of 0001-slice-and-events (the name), R33
  of 0001-slice-and-events (the slice exists), this requirement, then the
  lock (R6), then the command's own checks.
  Check: e2e

### Which slice

- **R12.** When `spec validate` or `spec submit` runs with no name,
  sliceflow shall use the slice of the current branch, as R31 and R32 of
  0001-slice-and-events say. In R32's second line, the command is the one
  that ran, e.g. `Give a name: sliceflow spec validate <name>`.
  Check: e2e

### Checking a spec

- **R13.** When the user runs `sliceflow spec validate [<name>]`,
  sliceflow shall read `docs/slices/<name>/spec.md` in the slice's
  worktree, as it is on disk, and print on stdout one line per problem,
  `spec.md:<line>: <level>: <message>`, in line order. Then it prints one
  summary line, `<n> errors, <n> warnings, <n> notes`, with `error`,
  `warning`, or `note` when the number is 1. It writes nothing.
  Check: e2e
- **R14.** `spec validate` shall exit with code 1 when it found an error,
  and 0 otherwise. Where `--strict` is given, it shall exit with code 1
  when it found an error or a warning. A note never changes the exit
  code.
  Check: e2e
- **R15.** sliceflow shall report these errors, each at the line named:

  | Check | Line | Message |
  |---|---|---|
  | The file starts with front matter between two `---` lines | 1 | `no front matter: the file must start with a line ---` |
  | It has `slice`, `created`, and `ideas` | 1 | `front matter has no "<key>"` |
  | `slice` is the slice's name | the key's | `front matter slice is "<x>", but the slice is <name>` |
  | `created` is a date like `2026-10-24` | the key's | `front matter created is "<x>", not a date like 2026-10-24` |
  | `ideas` is a list | the key's | `front matter ideas is not a list, e.g. ideas: []` |
  | Each `##` section of the template is there | the next section's, or the last line | `missing section "## <section>"` |
  | No other `##` section | the heading's | `unknown section "## <section>"` |
  | No section twice | the second heading's | `section "## <section>" appears twice (also line <n>)` |
  | The sections are in the template's order | the heading's | `section "## <a>" must come before "## <b>"` |
  | Each marker has the form `**R7.**`; checked on a list item that starts with `**` and then `R`, `S`, `SD`, `SC`, or `Q`, and a digit or a space | the item's | `"<text>" is not a marker; write it like **R7.**` |
  | No marker twice | the second one's | `<id> is used twice (also line <n>)` |
  | Each marker is in its section: R in Requirements, S in User stories, SC in Success criteria, SD in Decisions, Q in Open questions | the marker's | `<id> is outside "## <section>"` |
  | Each requirement has a `Check:` line | the marker's | `<id> has no Check: line` |
  | Each kind in `Check:` is `unit`, `api`, `e2e`, or `manual` | the `Check:` line | `<id>: unknown check "<x>"; use unit, api, e2e, or manual` |
  | Each success criterion has a `Measured:` line | the marker's | `<id> has no Measured: line` |
  | Each scenario has a `Shows:` line | the marker's | `<id> has no Shows: line` |
  | Each name in `Shows:` is a requirement of this spec | the `Shows:` line | `<id> shows <R>, which does not exist`, or `<id> shows <R>, which was removed` |
  | Each decision has `Why:`, `Not chosen:`, and `Rule:` | the marker's | `<id> has no <label> line` |
  | Each name in `Rule:` is a requirement of this spec, or `Rule:` says `none` | the `Rule:` line | `<id>: Rule: names <R>, which does not exist` |
  | Each Out of scope item ends with `Goes to:` | the item's | `out-of-scope item has no Goes to:` |
  | `Goes to:` names `the backlog`, `not planned`, or starts with `a later slice` | the item's | `Goes to: "<x>" is not a place; use the backlog, a later slice (<name>), or not planned` |

  The template's sections, in order, are: Summary, Why, Success criteria
  (optional), User stories, Key concepts (optional), Requirements, Out of
  scope, Decisions, Assumptions, Open questions. Keys other than
  `slice`, `created`, and `ideas` in the front matter are not checked.
  `ideas` is a list when it is `[]`, a list in brackets like
  `[docs/ideas/a.md]`, or is followed by lines that start with `- `.
  A `Rule:` that names a removed requirement gets
  `<id>: Rule: names <R>, which was removed`.
  Check: unit
- **R16.** sliceflow shall report these warnings:

  | Check | Line | Message |
  |---|---|---|
  | Open questions says `None.` | its heading's | `<n> open questions left (<Q1, Q2>)`, or `1 open question left (Q1)`, or, with no Q marker, `"## Open questions" is not None.` |
  | No `{{` is left | the `{{`'s | `placeholder left: <text from {{ to }}>` |
  | No template comment is left | the `<!--`'s | `template comment left` |
  | No section other than Success criteria and Key concepts is empty | the heading's | `"## <section>" is empty; write None. if there is nothing to say` |

  A section is empty when it holds nothing but blank lines and comments.
  When a `{{` has no `}}` after it on its line, the placeholder's text
  runs to the end of the line.
  Check: unit
- **R17.** sliceflow shall report the note `no scenario shows <id>`, at
  the requirement's line, for each requirement that no `Shows:` line
  names.
  Check: unit
- **R18.** When checking, sliceflow shall read `\r\n` line ends as `\n`,
  and skip everything inside fenced code blocks and `inline code`: no
  marker, `{{`, or `<!--` there is reported. It shall also skip
  everything inside a template comment: no marker or `{{` there is
  reported, and the comment gets only the warning `template comment left`.
  A labeled line belongs to the
  item above it. A removed item, e.g. `**R4.** (removed in R002: why)`,
  needs no labeled lines, and gets no note.
  Check: unit
- **R19.** If `spec.md` does not exist, then `spec validate` and
  `spec submit` shall print
  `Error: <path> not found. Nothing was changed.`
  with the path from the root of the main checkout, e.g.
  `.sliceflow/worktrees/login-google/docs/slices/login-google/spec.md`,
  and exit with code 1.
  Check: e2e

### Submitting the spec

- **R20.** When the user runs `sliceflow spec submit [<name>]`, sliceflow
  shall go on only when `spec.md` is `draft`, `submitted`, or
  `review_passed` and changed after its review (R30). When it is
  `submitted`, sliceflow skips R21 and R22, and only starts the review,
  as R24 to R28 say. Otherwise it shall save no event, print one of
  these, and exit with code 1:
  - `review_passed`, not changed:
    `Error: spec.md has not changed since review <n>, which passed. Nothing was changed.`
  - `in_review`:
    `Error: spec.md is in_review: review <n> has started, and has not finished. Nothing was changed.`
  - no good `slice.created`:
    `Error: slice <name> has no slice.created event, so its state is unknown. Nothing was changed.`

  sliceflow checks R11 first, then takes the lock (R6), then checks
  this, then R19.
  Check: e2e
- **R21.** sliceflow shall then check `spec.md` as `spec validate --strict`
  does, print its problem lines without the summary line, and save
  `spec.submitted`. Its `data` holds `result` (`accepted` when there is no
  error and no warning, else `refused`), `errors`, and `warnings` (the
  numbers found). The commit holds `spec.md` as it is on disk, and the
  log. Its message is `<name>: spec submitted`, or
  `<name>: spec submit refused`.
  Check: e2e
- **R22.** When the submit is refused, sliceflow shall print
  `Submit refused: spec.md is not finished (<n> errors, <n> warnings). It stays a draft.`
  with `error` or `warning` when the number is 1, and exit with code 1.
  When `spec.md` was `review_passed`, the last sentence is
  `It keeps the result of review <n>.`
  Check: e2e
- **R23.** When the submit is accepted, sliceflow shall print
  `Submitted spec.md.`, and start the review at once, as R25 to R28 say.
  Check: e2e
- **R24.** If the review file of the next round already exists in the
  slice folder, then sliceflow shall save no event, print
  `Error: docs/slices/<name>/reviews/spec-<n>.md already exists, but review <n> has not started. Move it away, then submit again. Nothing was changed.`
  and exit with code 1. It checks this before it saves an accepted
  `spec.submitted`, or, when `spec.md` is `submitted`, before it saves
  `review.started`.
  Check: e2e

### The review

- **R25.** sliceflow shall save `review.started`, with `doc` `spec.md`,
  `round` (1 more than the highest round of a `review.started` for
  `spec.md` that the replay allowed by R2, else 1), `reviewer`, `file` (`reviews/spec-<round>.md`), and `commit` (the
  full id of the commit of the last accepted `spec.submitted`). The commit holds the new file
  `docs/slices/<name>/reviews/spec-<round>.md` and the log. The file holds
  only this front matter:
  ```
  ---
  review: spec-1
  slice: login-google
  kind: spec
  target: spec.md
  round: 1
  reviewer: placeholder
  ---
  ```
  The commit's message is `<name>: review <round> started`. sliceflow
  then prints
  `Review <round> started: reviews/spec-<round>.md (reviewer: <reviewer>)`.
  Check: e2e
- **R26.** The reviewer in this slice shall be the placeholder: its
  `reviewer` is `placeholder`, it reads nothing, and its result is
  `passed`. It adds this to the review file, after the front matter. It
  has every section of the review template, in the template's order:
  ```
  # Review

  ## Summary

  No review was done: placeholder reviewer.

  ## Checked

  None.

  ## Tests

  None.

  ## Previous findings

  None.

  ## Findings

  None.

  ## Requirements check

  None.

  ## Needs human decision

  None.
  ```
  Check: e2e
- **R27.** When the reviewer is done, sliceflow shall save
  `review.finished`, with `doc`, `round`, and `result`. The commit holds
  the filled review file and the log. Its message is
  `<name>: review <round> passed`, `<name>: review <round> found changes`,
  or `<name>: review <round> needs the human`.
  Check: e2e
- **R28.** When the review passed, sliceflow shall print
  `Review <round> passed.` and `Next: write the plan.`, and exit with
  code 0.
  Check: e2e

### Status

- **R29.** `slice status` shall print two more lines after `Created:`:
  `Spec:` and `Next:`, from the spec's state:

  | When | `Spec:` | `Next:` |
  |---|---|---|
  | `draft`, no review yet | `draft` | `write spec.md, then: sliceflow spec submit` |
  | `draft`, after review `<n>` found changes | `draft (review <n> found changes: reviews/spec-<n>.md)` | `fix spec.md, then: sliceflow spec submit` |
  | `draft`, after review `<n>` needs the human | `draft (review <n> needs the human: reviews/spec-<n>.md)` | `settle reviews/spec-<n>.md with the human, then: sliceflow spec submit` |
  | `submitted` | `submitted (review <n> not started)` | `sliceflow spec submit, to start review <n>` |
  | `in_review` | `in_review (review <n>, <reviewer>: reviews/spec-<n>.md)` | `wait for review <n>` |
  | `review_passed` | `review_passed (review <n>, <reviewer>)` | `write the plan` |
  | `review_passed`, changed after its review | `passed review <n>, but spec.md was changed after it` | `sliceflow spec submit` |
  | no good `slice.created` | `-` | `-` |

  When the last spec event is a refused `spec.submitted`, and `spec.md`
  is `draft` or changed after its review, `Spec:` instead shows `draft`,
  or `passed review <n>, but spec.md was changed after it`, followed by
  ` (submit refused: <counts>)`, and `Next:` shows
  `fix spec.md, then: sliceflow spec submit`. `<counts>` names only the
  numbers that are not 0, e.g. `1 warning`, or `2 errors, 1 warning`.
  Check: e2e
- **R30.** While `spec.md` is `review_passed`, sliceflow shall compare
  `spec.md` in the last reviewed commit with `spec.md` on disk in the
  slice's worktree. When git would show a change between them, `spec.md`
  is changed after its review. Line ends that git converts do not count
  as a change. When the slice has no worktree, sliceflow does not
  compare. This writes no event, and is no state.
  Check: e2e
- **R31.** Where `--json` is given, `slice status` shall add to R30 of
  0001-slice-and-events' object:
  - `docs`: a list with one object for `spec.md`: `doc`, `status` (its
    state), `review` (the last review's `round`, `reviewer`, `file`, and
    `result`; `result` is left out while it runs; `review` is left out
    when no review started), `submit_refused` (`errors` and `warnings`,
    only when R29's refused mark shows), and `changed_after_review`
    (`true`, only when R30 says so).
  - `next`: the text of R29's `Next:`.

  With no good `slice.created`, `docs` is `[]`, and `next` is left out.
  Each slice in `slice list --json` has these two fields too.
  Check: e2e

### The title

- **R32.** When the user runs `sliceflow slice title <name> <title>`,
  sliceflow shall save `slice.titled` with `title`, and print
  `Title of <name> is now "<title>".` The commit holds only the log. Its
  message is `<name>: title set`. This works in any phase. When the log
  has no good `slice.created`, sliceflow shall instead print
  `Error: slice <name> has no slice.created event, so its state is unknown. Nothing was changed.`
  and exit with code 1.
  Check: e2e
- **R33.** If `<title>` is the title the slice has now, then sliceflow
  shall print
  `Error: the title of <name> is already "<title>". Nothing was changed.`
  If `<title>` is empty, it shall print
  `Error: the title must not be empty. Nothing was changed.`
  Both save no event, and exit with code 1.
  Check: e2e
- **R34.** `slice status` and `slice list` shall show the title of the
  last good `slice.titled`, else the title of `slice.created`, else the
  slice's name, as R29 of 0001-slice-and-events says.
  Check: e2e

### The template

- **R35.** When the user runs `sliceflow template show spec`, sliceflow
  shall print on stdout the built-in spec template of R7 of
  0001-slice-and-events, before its front matter is filled, with all its
  rule comments, and exit with code 0. For any other template name, it
  shall print `Error: unknown template "<x>". Known templates: spec.`
  and exit with code 1.
  Check: e2e

### Every command

- **R36.** For wrong use, as R38 of 0001-slice-and-events says, sliceflow
  shall also know `missing title` (`slice title` with one argument) and
  `missing template name` (`template show` with none). `spec validate`,
  `spec submit`, `slice title`, and `template show` have no `--json`.
  `sliceflow spec` alone and `sliceflow template` alone print their usage
  on stdout, and exit with code 0.
  Check: e2e
- **R37.** `spec validate` and `spec submit` shall print their problem
  lines, the summary line, the `Submit refused:` line, and the lines of
  steps already done, e.g. `Submitted spec.md.`, on stdout, also
  when they exit with code 1. This is the one exception to R38 of
  0001-slice-and-events, which prints nothing on stdout when a command
  fails. Every `Error:` line goes on stderr.
  Check: e2e

## Out of scope

- The plan, the tasks, the preview, and the approval. Goes to: a later slice (0003)
- A number reused for something new after approval. Goes to: a later slice (0003)
- A real reviewer, and what `spec submit` prints when a review does not pass. Goes to: a later slice (0004)
- A limit on review rounds. Goes to: a later slice (0004)
- Going on after a submit that stopped while `spec.md` is `in_review`. Goes to: a later slice (0007)
- Who caused an event (a `by` field). Goes to: a later slice (0004)
- Turning the spec review off, and the `review.skipped` event. Goes to: a later slice (0012)
- The spec as JSON, and `--json` for `spec validate`, `spec submit`, `slice title`, and `template show`. Goes to: a later slice (0015)
- Templates other than the spec in `template show`. Goes to: a later slice (configurable templates)
- Commands that show the spec as approved, and the difference between two revisions. Goes to: a later slice (0003)
- Showing the title in `spec.md`, written from the log. Goes to: the backlog
- Removing a title. Goes to: not planned
- Checking how well the spec is written, e.g. plain English. Goes to: not planned
- Checking that numbers have no gaps, e.g. R1, R2, R4. Goes to: not planned

## Decisions

- **SD1.** sliceflow works out the state by replaying the committed log
  each time. It keeps no saved state.
  Why: one source of truth, as SD2 of 0001-slice-and-events says. A
    slice's log holds few events, so a replay is fast.
  Not chosen: a saved state, updated with each event, because it is a
    second copy that can drift from the log.
  Cost: every command reads the whole log.
  Rule: R1
- **SD2.** Every state change comes from exactly one event, by a fixed
  table. An event does not hold the new state, and its meaning never
  changes: a new behavior gets a new event name, or a new `version`.
  Why: the log tells what happened, and the same log always gives the
    same state.
  Not chosen: events that hold the new state, e.g. `doc.status_changed`
    with `to: draft`, because the same fact is then saved twice, and the
    two can disagree.
  Rule: R2
- **SD3.** `slice.created` starts the spec as `draft`, with no extra
  event.
  Why: creating a slice already means its spec is a draft, and 0001
    stays as it was approved.
  Not chosen: a second event after `slice.created`, because it repeats a
    fact, and changes 0001.
  Rule: R2, R3
- **SD4.** A submit from the wrong state is refused with no event. A
  submit refused because the spec is not finished is saved, as an event
  and a commit with that `spec.md`.
  Why: a move from the wrong state changes nothing. An unfinished submit
    is a step of the work, and the history then shows each version that
    was checked.
  Not chosen: saving only accepted submits, because the attempts, and
    the versions checked, would be lost.
  Cost: one commit for each refused submit.
  Rule: R20, R21
- **SD5.** Each event is its own commit, a normal git commit.
  Why: an event counts once it is committed, as SD4 of
    0001-slice-and-events says. With a real reviewer, a review's start
    and finish are minutes apart.
  Not chosen: one commit per command, because a long review would leave
    its start unsaved until it ends.
  Rule: R8
- **SD6.** `spec validate` has three levels: an error (the form is
  broken), a warning (not finished yet), and a note (worth a look).
  `--strict` fails on warnings. A note never fails.
  Why: a correct spec can have a requirement that no scenario shows,
    e.g. R40 of 0001-slice-and-events (no network). As a warning, it would
    force a filler scenario just to pass `--strict`.
  Not chosen: two levels, with "no scenario shows" as a warning,
    because a correct spec could then pass `--strict` only with a filler
    scenario.
  Rule: R14, R17
- **SD7.** The reviewer in this slice is a placeholder. It always passes,
  and says so in the event and in the review file.
  Why: the review steps are built and tested now, without an agent. The
    `sliceflow` binary is used only in tests until the first release.
    `reviewer` shows which reviews were real.
  Not chosen: no review step until 0004, because submit would then end
    in a state that 0004 has to change again.
  Cost: a spec can show `review_passed` with no real review.
  Rule: R26
- **SD8.** The review always runs after an accepted submit.
  Why: a switch needs a setting, which comes with 0012.
  Not chosen: a switch now, because there is no real reviewer to turn
    off yet.
  Rule: R23
- **SD9.** "Changed after its review" is worked out when it is shown, by
  comparing `spec.md` with the reviewed commit. It is no state, and has
  no event.
  Why: editing a file is not a sliceflow step, and the reviewed commit
    gives an exact answer.
  Not chosen: refusing every submit after a passed review, because the
    lead could then never fix a reviewed spec.
  Rule: R20, R30
- **SD10.** One lock file per slice, held only while one event is
  written.
  Why: it is simple, works the same on every system, and needs no new
    dependency. A long review does not block other commands.
  Not chosen: one lock for the whole command, because in 0004 a review
    would block `slice title` for minutes.
  Cost: after a crash, a lock file is left, and must be deleted by hand.
    The busy message says so.
  Rule: R6, R7
- **SD11.** A title can be set in any phase, and cannot be removed.
  Why: the title is for people. Changing it never changes the work.
  Not chosen: a way to remove the title, because the name already shows
    when no title was set.
  Rule: R32, R33
- **SD12.** `spec validate` and `spec submit` take the slice's name, or
  use the current branch's slice. `slice title` always takes the name.
  Why: `slice title` with one argument could be either the name or the
    title.
  Not chosen: an optional name for `slice title`, because the meaning of
    its one argument would depend on the branch.
  Rule: R12, R36
- **SD13.** A submit that stopped after `spec.submitted` goes on when
  `spec submit` runs again: it starts the review, and does not hand the
  spec in again.
  Why: no review started, so no reviewer can be running, and starting
    one now is safe.
  Not chosen: leaving it stuck until the retry slice (0007), because a
    refused commit is the most likely stop. Also going on from
    `in_review`, because from 0004 on a reviewer may still be running
    then, and a second one would start.
  Cost: a slice that stopped at `in_review` stays there until 0007.
  Rule: R20
- **SD14.** `spec validate` and `spec submit` print the spec's problems
  on stdout, also when they exit with code 1.
  Why: a list of problems is the command's answer, not a broken
    command, as in linters like ESLint. A pipe, e.g. to `grep`, sees it.
  Not chosen: stderr, as R38 of 0001-slice-and-events says for errors,
    because a pipe or a saved file then needs `2>&1`, which is easy to
    forget.
  Cost: one exception to R38 of 0001-slice-and-events.
  Rule: R37
- **SD15.** When the committed log's last line is bad or has no `\n`,
  sliceflow keeps it, and starts the new event on a new line.
  Why: nothing is deleted, and a bad line is skipped with a warning and
    never stops a command, as SD5 of 0001-slice-and-events says.
  Not chosen: cutting off the bad last line, because sliceflow would then
    delete a line from the log. Also refusing to write, because nobody
    edits the log by hand.
  Cost: the warning about that line stays for good.
  Rule: R9

## Assumptions

- Renaming a file over another replaces it in one step, on Windows,
  macOS, and Linux.
  Noticed: a test on one of the CI systems finds the log missing, or
  half written.
- All sliceflow commands for one repo run on one computer, so the process
  id in a lock file means something.
  Noticed: two computers share a repo folder over a network, and the busy
  message names a process that is not there.
- 0001 is built as its spec says: an event that sliceflow does not know is
  skipped with a warning, so an older sliceflow can still read a newer
  log.
  Noticed: the change review of 0002 after 0001 merges.

## Open questions

None.
