---
slice: slice-and-events
created: 2026-11-02
---

<!-- HOW TO FILL THIS TEMPLATE

     A spec says WHAT a feature does, never HOW it is built. No endpoints,
     tables, libraries, or file names. Those go in plan.md. What the
     user sees does belong here: commands, flags, screens, output layout,
     messages, and exit codes. The plan may not decide these.

     The human approves the spec. Write plain English, short sentences,
     one idea per sentence.

     sliceflow reads this file. Keep every marker exactly as shown: the
     headings, the numbers (**R1.**, **S1.**, **SD1.**, **Q1.**), and the
     labeled lines (Check:, Shows:, Why:, Not chosen:, Cost:, Rule:,
     Replaces:). Don't rename or reorder them. A labeled line
     may go on over more lines: indent the extra lines 2 spaces more than
     the label.

     - Replace every {{...}} with real text. The text inside says what
       goes there. sliceflow fills slice, title, and date itself. If they
       still show {{...}}, fill them yourself.
     - The title of the slice lives only in this front matter. Don't
       repeat it in a heading or in other files.
     - Each section has a comment like this one with its rules. Follow it.
     - Keep the sections in this order. Don't rename them.
     - Repeat a block as often as needed: stories, scenarios,
       requirements, decisions.
     - A required section with nothing to say gets one line: None.
     - When you are done, delete every comment and check that no {{ is
       left. A spec with a comment or a {{ left in it is not finished.
       Every placeholder of this template counts, even inside code. A {{
       or a comment marker you write yourself inside a code block or
       `inline code` is real content.
     - To change a spec later, read this template again first. The
       finished spec no longer has these rules in it. -->

# Spec

## Summary

<!-- Two or three sentences: what the feature lets someone do, not how,
     and why it is needed. The why helps a worker guess right when a
     requirement leaves something open. -->

A developer starts a new slice with one command, and later sees one slice
or all slices. sliceflow makes the slice's branch, its worktree, its spec
from the template, and an event log that records that the slice was
created. The event log is the only source of truth, so every later slice
builds on it.

## User stories

<!-- What the user does. This is the first thing a reviewer reads.

     - Every scenario sits under a story, even when there is only one
       story. A story is one thing the user wants to do. It always has
       both: a short title, and one sentence saying who wants to do what.
     - Under it, a list of scenarios, one per case. The success cases
       come first, then each important failure.
     - Number scenarios S1, S2, ... across the whole spec, not per story.
       Same rules as requirement numbers: renumber freely only before the
       spec is first approved. After that, never reuse a number, and a
       removed scenario stays, as: **S4.** (removed in R002: why)
     - Each scenario has three parts:
       - Given: the starting state. Leave it out when there is nothing
         special.
       - When: what the user does. Only one per scenario. If you need
         two, split it into two scenarios.
       - Then: what the user sees. Show the real screen, message, or
         terminal output. Use a code block when it is longer than one
         line.
     - Use "and" for more than one condition or result.
     - End each scenario with a Shows: line naming the requirements it
       shows, e.g. Shows: R1, R2
     - Every requirement a user can see, including error messages,
       appears in at least one scenario.

     A scenario is one example, with real values. The requirements are
     the full rules. Both are needed.

     Example:
     ### Story 1: Start a new slice

     A developer starts work on a new feature with one command.

     - **S1.** **When** I run `sliceflow slice create checkout`,
       **Then** sliceflow creates the slice and says where to write the spec:
       ```
       $ sliceflow slice create checkout
       Created slice checkout.
       Next: write the spec in .sliceflow/worktrees/checkout/docs/slices/checkout/spec.md
       ```
       Shows: R1, R2
     - **S2.** **Given** a slice named `checkout` already exists,
       **When** I run `sliceflow slice create checkout`,
       **Then** sliceflow changes nothing and prints
       "Error: slice checkout already exists: branch slice/checkout.
       Nothing was changed."
       Shows: R3 -->

### Story 1: Create a slice

A developer starts a new slice with one command.

- **S1.** **Given** I am in the project `shop`, which has commits and a
  `main` branch,
  **When** I run `sliceflow slice create card-payment --title "Pay with a card" --description "Let a customer pay with a credit or debit card."`,
  **Then** sliceflow creates the slice, prints this, and exits with code 0:
  ```
  Created slice card-payment.
    Branch:   slice/card-payment (from main, at 3f9a1c2)
    Worktree: /home/you/.sliceflow/worktrees/shop/card-payment
  Next: write the spec in /home/you/.sliceflow/worktrees/shop/card-payment/docs/slices/card-payment/spec.md
  ```
  Shows: R1, R2, R4, R5, R6, R7, R11, R12
- **S2.** **When** I run `sliceflow slice create card.payment`,
  **Then** sliceflow changes nothing, prints this to the error output, and
  exits with code 2:
  ```
  Error: invalid slice name "card.payment". Nothing was changed.
  Allowed characters: letters, digits, hyphens, underscores. Start with a letter or a digit.
  ```
  Shows: R13, R14, R58
- **S3.** **Given** a slice named `card-payment` already exists, with the
  branch `slice/card-payment`,
  **When** I run `sliceflow slice create card-payment`,
  **Then** sliceflow changes nothing, prints this to the error output, and
  exits with code 1:
  ```
  Error: slice card-payment already exists (branch slice/card-payment). Nothing was changed.
  ```
  Shows: R18
- **S4.** **Given** a slice named `card-payment` was finished and merged into
  `main` earlier, so its branch is gone but the folder
  `docs/slices/card-payment/` is in `main`,
  **When** I run `sliceflow slice create card-payment`,
  **Then** sliceflow changes nothing, prints this to the error output, and
  exits with code 1:
  ```
  Error: slice card-payment already exists (folder docs/slices/card-payment/ in main). Nothing was changed.
  ```
  Shows: R17
- **S5.** **Given** I am in the project `shop`, which has commits and a
  `main` branch,
  **When** I run `sliceflow slice create Card-Payment`,
  **Then** sliceflow creates the slice `card-payment`, prints this, and
  exits with code 0:
  ```
  Created slice card-payment.
    Note:     "Card-Payment" was made lowercase.
    Branch:   slice/card-payment (from main, at 3f9a1c2)
    Worktree: /home/you/.sliceflow/worktrees/shop/card-payment
  Next: write the spec in /home/you/.sliceflow/worktrees/shop/card-payment/docs/slices/card-payment/spec.md
  ```
  Shows: R5, R15, R16
- **S6.** **Given** I am in a folder that is not inside a git repository,
  **When** I run `sliceflow slice create card-payment`,
  **Then** sliceflow changes nothing, prints this to the error output, and
  exits with code 1:
  ```
  Error: this folder is not in a git repository. Nothing was changed.
  Run sliceflow inside your project.
  ```
  Shows: R9
- **S7.** **Given** I am in a git repository that has no commits yet,
  **When** I run `sliceflow slice create card-payment`,
  **Then** sliceflow changes nothing, prints this to the error output, and
  exits with code 1:
  ```
  Error: this repository has no commits yet. Nothing was changed.
  Make a first commit, then try again.
  ```
  Shows: R10
- **S8.** **Given** I am in a git repository whose main branch is `master`,
  and no branch named `main` exists,
  **When** I run `sliceflow slice create card-payment`,
  **Then** sliceflow changes nothing, prints this to the error output, and
  exits with code 1:
  ```
  Error: base branch main does not exist. Nothing was changed.
  sliceflow starts from main unless you choose another branch. To choose one, run:
    sliceflow slice create card-payment --from <branch>
  ```
  Shows: R8
- **S9.** **Given** I am in the project `shop`, which has a `main` branch and
  a `release` branch,
  **When** I run `sliceflow slice create card-payment --from release`,
  **Then** sliceflow creates the slice from the tip of `release`, prints
  this, and exits with code 0:
  ```
  Created slice card-payment.
    Branch:   slice/card-payment (from release, at 8c41d07)
    Worktree: /home/you/.sliceflow/worktrees/shop/card-payment
  Next: write the spec in /home/you/.sliceflow/worktrees/shop/card-payment/docs/slices/card-payment/spec.md
  ```
  Shows: R6, R7
- **S10.** **Given** the folder `/home/you/.sliceflow/worktrees/shop/card-payment`
  already exists and has files in it,
  **When** I run `sliceflow slice create card-payment`,
  **Then** sliceflow removes anything it made, prints this to the error
  output, and exits with code 1:
  ```
  Error: git could not create the worktree. The slice was not created. Nothing was changed.
  git said:
    fatal: '/home/you/.sliceflow/worktrees/shop/card-payment' already exists
  ```
  Shows: R19, R25, R26
- **S11.** **Given** git has no name or email set on this computer,
  **When** I run `sliceflow slice create card-payment`,
  **Then** sliceflow removes the files, the worktree, and the branch it
  made, prints this to the error output, and exits with code 1:
  ```
  Error: git could not make the commit. The slice was not created. Nothing was changed.
  git said:
    Author identity unknown
    *** Please tell me who you are.
    Run
      git config --global user.email "you@example.com"
      git config --global user.name "Your Name"
  ```
  Shows: R25, R26
- **S12.** **Given** git has no name or email set on this computer, and
  the worktree sliceflow made cannot be removed (for example, a file in it
  is locked),
  **When** I run `sliceflow slice create card-payment`,
  **Then** sliceflow removes what it can, says what is left and how to
  remove it, prints this to the error output, and exits with code 1:
  ```
  Error: git could not make the commit. The slice was not created.
  git said:
    Author identity unknown
    *** Please tell me who you are.
    Run
      git config --global user.email "you@example.com"
      git config --global user.name "Your Name"
  Some of what sliceflow made could not be removed:
    worktree /home/you/.sliceflow/worktrees/shop/card-payment
    branch slice/card-payment
  To clean up, run:
    git worktree remove --force /home/you/.sliceflow/worktrees/shop/card-payment
    git branch -D slice/card-payment
  ```
  Shows: R27
- **S13.** **Given** I created the slice `card-payment` with a title and a
  description, as in S1, on 2 November 2026 at 09:15:30 UTC,
  **When** I open the event log in the slice's folder,
  **Then** it has exactly one line, and the same line is in the slice
  branch:
  ```
  {"version":1,"seq":1,"event":"slice.created","at":"2026-11-02T09:15:30Z","data":{"name":"card-payment","title":"Pay with a card","description":"Let a customer pay with a credit or debit card.","base":"main","base_commit":"3f9a1c2e8b7d4a6f0c1e5d9b2a7f3c8e4d6b1a09"}}
  ```
  Shows: R21
- **S14.** **Given** I created the slice `card-payment`, as in S1,
  **When** I look at the slice branch with `git log --stat slice/card-payment`,
  **Then** it has one new commit on top of `main`, and my own checkout has
  no new commit. The commit has this message, and it holds exactly two
  files:
  ```
  card-payment: add the spec and the event log

   docs/slices/card-payment/.events.jsonl | 1 +
   docs/slices/card-payment/spec.md       | 63 +++++++++
  ```
  Shows: R2, R3, R4
- **S15.** **Given** I am in the project `shop`, which has commits and a
  `main` branch,
  **When** I run `sliceflow slice create card-payment` with no `--title`
  and no `--description`, on 2 November 2026 at 09:20:05 UTC,
  **Then** sliceflow creates the slice, the title is the name, the
  description is empty, and the event log has this line:
  ```
  {"version":1,"seq":1,"event":"slice.created","at":"2026-11-02T09:20:05Z","data":{"name":"card-payment","title":"card-payment","description":"","base":"main","base_commit":"3f9a1c2e8b7d4a6f0c1e5d9b2a7f3c8e4d6b1a09"}}
  ```
  Shows: R21, R22
- **S16.** **Given** I am in the folder `shop/src/api`, inside the project
  `shop`, which has commits and a `main` branch,
  **When** I run `sliceflow slice create card-payment`,
  **Then** sliceflow creates the slice exactly as in S1, and the worktree is
  under `shop`, not under `api`:
  ```
  Created slice card-payment.
    Branch:   slice/card-payment (from main, at 3f9a1c2)
    Worktree: /home/you/.sliceflow/worktrees/shop/card-payment
  Next: write the spec in /home/you/.sliceflow/worktrees/shop/card-payment/docs/slices/card-payment/spec.md
  ```
  Shows: R1, R11
- **S17.** **Given** I am in a git repository that has no branch named
  `release`,
  **When** I run `sliceflow slice create card-payment --from release`,
  **Then** sliceflow changes nothing, prints this to the error output, and
  exits with code 1:
  ```
  Error: base branch release does not exist. Nothing was changed.
  Check the name. To see your branches, run:
    git branch
  ```
  Shows: R8

### Story 2: See one slice

A developer checks one slice, and its history, even after it is merged.

- **S18.** **Given** I created the slice `card-payment` with a title and a
  description, and its worktree is still there,
  **When** I run `sliceflow slice status card-payment`,
  **Then** sliceflow prints this to the normal output, changes nothing, and
  exits with code 0:
  ```
  card-payment: Pay with a card
    Description:  Let a customer pay with a credit or debit card.
    Created:      2026-11-02 09:15 UTC
    Started from: main, at 3f9a1c2
    Branch:       slice/card-payment
    Worktree:     /home/you/.sliceflow/worktrees/shop/card-payment
    Events:       1 (last: slice.created)
  ```
  Shows: R15, R28, R29, R30, R31, R32, R33, R34, R36
- **S19.** **Given** I created the slice `card-payment`, and it was merged
  into `main` by hand, and its branch and its worktree were deleted,
  **When** I run `sliceflow slice status card-payment`,
  **Then** sliceflow reads the log from `main`, prints this to the normal
  output, changes nothing, and exits with code 0:
  ```
  card-payment: Pay with a card
    Description:  Let a customer pay with a credit or debit card.
    Created:      2026-11-02 09:15 UTC
    Started from: main, at 3f9a1c2
    Branch:       slice/card-payment (not found)
    Worktree:     (not found)
    Events:       1 (last: slice.created)
    Read from:    main
  ```
  Shows: R28, R32, R33, R35, R36
- **S20.** **Given** no slice named `card-payment` exists, and no log for it
  is in a worktree, on a branch, or in `main`,
  **When** I run `sliceflow slice status card-payment`,
  **Then** sliceflow changes nothing, prints this to the error output, and
  exits with code 1:
  ```
  Error: no slice named card-payment. Nothing was changed.
  No log was found in any of these places:
    worktree  /home/you/.sliceflow/worktrees/shop/card-payment
    branch    slice/card-payment
    main      docs/slices/card-payment/.events.jsonl
  Check the name, or run: sliceflow slice list
  ```
  Shows: R37
- **S21.** **Given** I created the slice `card-payment`, and the second line
  of its event log is cut off, so it cannot be read,
  **When** I run `sliceflow slice status card-payment`,
  **Then** sliceflow skips that line, prints the warning to the error output
  and the status to the normal output, changes nothing, and exits with
  code 0:
  ```
  Warning: line 2 of docs/slices/card-payment/.events.jsonl could not be read, so it was skipped.
  card-payment: Pay with a card
    Description:  Let a customer pay with a credit or debit card.
    Created:      2026-11-02 09:15 UTC
    Started from: main, at 3f9a1c2
    Branch:       slice/card-payment
    Worktree:     /home/you/.sliceflow/worktrees/shop/card-payment
    Events:       1 (last: slice.created)
  ```
  Shows: R38
- **S22.** **Given** the event log of `card-payment` has five lines, and
  lines 2, 4, and 5 are cut off, so they cannot be read,
  **When** I run `sliceflow slice status card-payment`,
  **Then** sliceflow skips those lines, prints one warning to the error
  output and the status to the normal output, changes nothing, and exits
  with code 0:
  ```
  Warning: lines 2, 4, 5 of docs/slices/card-payment/.events.jsonl could not be read, so they were skipped.
  card-payment: Pay with a card
    Description:  Let a customer pay with a credit or debit card.
    Created:      2026-11-02 09:15 UTC
    Started from: main, at 3f9a1c2
    Branch:       slice/card-payment
    Worktree:     /home/you/.sliceflow/worktrees/shop/card-payment
    Events:       2 (last: slice.created)
  ```
  Shows: R34, R38
- **S23.** **Given** the event log of `card-payment` has 13 lines, and lines
  2 to 13 are cut off, so they cannot be read,
  **When** I run `sliceflow slice status card-payment`,
  **Then** sliceflow skips those lines, prints one warning that lists the
  first 10 line numbers and says how many more there are, prints the
  status, changes nothing, and exits with code 0:
  ```
  Warning: lines 2, 3, 4, 5, 6, 7, 8, 9, 10, 11 (2 more not listed) of docs/slices/card-payment/.events.jsonl could not be read, so they were skipped.
  card-payment: Pay with a card
    Description:  Let a customer pay with a credit or debit card.
    Created:      2026-11-02 09:15 UTC
    Started from: main, at 3f9a1c2
    Branch:       slice/card-payment
    Worktree:     /home/you/.sliceflow/worktrees/shop/card-payment
    Events:       1 (last: slice.created)
  ```
  Shows: R34, R38, R39
- **S24.** **Given** the first line of the event log of `card-payment` is
  cut off, so no `slice.created` event can be read,
  **When** I run `sliceflow slice status card-payment`,
  **Then** sliceflow changes nothing, prints this to the error output, and
  exits with code 1:
  ```
  Warning: line 1 of docs/slices/card-payment/.events.jsonl could not be read, so it was skipped.
  Error: no slice.created event could be read for card-payment. Nothing was changed.
  Fix or restore the log, then try again.
  ```
  Shows: R38, R40
- **S25.** **Given** I created the slice `card-payment`, and I removed its
  worktree, but its branch `slice/card-payment` is still there,
  **When** I run `sliceflow slice status card-payment`,
  **Then** sliceflow reads the log from the branch, prints this to the
  normal output, changes nothing, and exits with code 0:
  ```
  card-payment: Pay with a card
    Description:  Let a customer pay with a credit or debit card.
    Created:      2026-11-02 09:15 UTC
    Started from: main, at 3f9a1c2
    Branch:       slice/card-payment
    Worktree:     (not found)
    Events:       1 (last: slice.created)
    Read from:    branch slice/card-payment
  ```
  Shows: R28, R33, R35, R36

### Story 3: See all slices

A developer sees all the slices of the project, including the merged ones,
in one table.

- **S26.** **Given** I am in the project `shop`, which has three slices:
  `card-payment`, which still has its worktree; `login-v2`, which has its
  branch but no worktree; and `old-search`, which was merged into `main`,
  and its branch and its worktree are gone, and `old-search` has no
  description,
  **When** I run `sliceflow slice list`,
  **Then** sliceflow prints this table to the normal output, sorted by
  name, changes nothing, and exits with code 0:
  ```
  NAME          TITLE                 DESCRIPTION
  card-payment  Pay with a card       Let a customer pay with a credit or debit card.
  login-v2      Login with a passkey  Replace the password login with a passkey.
  old-search    Search by name
  ```
  Shows: R41, R42, R43, R44, R45, R46
- **S27.** **Given** the project has the slice `card-payment`, and a branch
  `slice/foo` that has no event log, and no folder `docs/slices/foo/` in
  `main`,
  **When** I run `sliceflow slice list`,
  **Then** sliceflow prints the warning to the error output and the table
  to the normal output, changes nothing, and exits with code 0:
  ```
  Warning: foo is not listed as a slice. The branch slice/foo has no event log.
  NAME          TITLE            DESCRIPTION
  card-payment  Pay with a card  Let a customer pay with a credit or debit card.
  ```
  Shows: R48
- **S28.** **Given** the project has the slice `card-payment`, and a folder
  `docs/slices/foo/` in `main` that has no event log, and no branch
  `slice/foo`,
  **When** I run `sliceflow slice list`,
  **Then** sliceflow prints the warning to the error output and the table
  to the normal output, changes nothing, and exits with code 0:
  ```
  Warning: foo is not listed as a slice. The folder docs/slices/foo/ in main has no event log.
  NAME          TITLE            DESCRIPTION
  card-payment  Pay with a card  Let a customer pay with a credit or debit card.
  ```
  Shows: R48
- **S29.** **Given** the project has the slice `card-payment`, and both a
  branch `slice/foo` and a folder `docs/slices/foo/` in `main`, and
  neither has an event log,
  **When** I run `sliceflow slice list`,
  **Then** sliceflow prints one warning to the error output and the table
  to the normal output, changes nothing, and exits with code 0:
  ```
  Warning: foo is not listed as a slice. Neither the branch slice/foo nor the folder docs/slices/foo/ in main has an event log.
  NAME          TITLE            DESCRIPTION
  card-payment  Pay with a card  Let a customer pay with a credit or debit card.
  ```
  Shows: R48
- **S30.** **Given** the project has no slices,
  **When** I run `sliceflow slice list`,
  **Then** sliceflow prints this to the normal output, changes nothing, and
  exits with code 0:
  ```
  No slices yet. To make one, run: sliceflow slice create <name>
  ```
  Shows: R47
- **S31.** **Given** the slice `card-payment` was merged into `main` by hand,
  and its branch `slice/card-payment` was not deleted, so both the branch
  and the folder `docs/slices/card-payment/` in `main` have its log,
  **When** I run `sliceflow slice list`,
  **Then** sliceflow shows `card-payment` once, prints no warning, changes
  nothing, and exits with code 0:
  ```
  NAME          TITLE            DESCRIPTION
  card-payment  Pay with a card  Let a customer pay with a credit or debit card.
  ```
  Shows: R42, R43
- **S32.** **Given** the project has the slices `card-payment` and
  `login-v2`, and the first line of the event log of `card-payment` cannot
  be read, and some later line of the log of `login-v2` cannot be read,
  **When** I run `sliceflow slice list`,
  **Then** sliceflow prints one warning to the error output, for
  `card-payment` only, and the table to the normal output, changes
  nothing, and exits with code 0:
  ```
  Warning: card-payment is not listed as a slice. No slice.created event could be read from docs/slices/card-payment/.events.jsonl.
  NAME      TITLE                 DESCRIPTION
  login-v2  Login with a passkey  Replace the password login with a passkey.
  ```
  Shows: R49, R50

### Story 4: Get help, and use a command wrongly

A developer asks sliceflow for help, and learns what went wrong when a
command is used the wrong way.

- **S33.** **When** I run `sliceflow --help`,
  **Then** sliceflow prints this to the normal output, changes nothing, and
  exits with code 0:
  ```
  sliceflow runs a spec-driven workflow with AI coding agents.

  Usage:
    sliceflow <command> [options]

  Commands:
    slice create <name>  Make a new slice
    slice status <name>  Show one slice
    slice list           Show all slices

  For help with a command, run: sliceflow <command> --help
  ```
  Shows: R51
- **S34.** **When** I run `sliceflow` with nothing after it,
  **Then** sliceflow prints the same guide as in S33 to the normal output,
  changes nothing, and exits with code 0.
  Shows: R51
- **S35.** **When** I run `sliceflow slice create --help`,
  **Then** sliceflow prints this to the normal output, changes nothing, and
  exits with code 0:
  ```
  Make a new slice: a branch, a worktree, a spec, and an event log.

  Usage:
    sliceflow slice create <name> [options]

  Options:
    --title <text>        A short title for the slice. Default: the name.
    --description <text>  One or two sentences about the slice.
    --from <branch>       The branch to start from. Default: main.
    --help                Show this help.
  ```
  Shows: R52
- **S36.** **When** I run `sliceflow slice delete card-payment`,
  **Then** sliceflow changes nothing, prints this to the error output, and
  exits with code 2:
  ```
  Error: unknown command "slice delete". Nothing was changed.
  To see the commands, run: sliceflow --help
  ```
  Shows: R53, R58
- **S37.** **When** I run `sliceflow slice create card-payment --titel "Pay with a card"`,
  **Then** sliceflow changes nothing, prints this to the error output, and
  exits with code 2:
  ```
  Error: unknown option "--titel". Nothing was changed.
  To see the options, run: sliceflow slice create --help
  ```
  Shows: R54, R58
- **S38.** **When** I run `sliceflow slice create` with no name,
  **Then** sliceflow changes nothing, prints this to the error output, and
  exits with code 2:
  ```
  Error: missing slice name. Nothing was changed.
  Usage: sliceflow slice create <name> [options]
  ```
  Shows: R55, R58
- **S39.** **When** I run `sliceflow slice create card-payment --title` with
  no text after `--title`,
  **Then** sliceflow changes nothing, prints this to the error output, and
  exits with code 2:
  ```
  Error: option "--title" needs a value. Nothing was changed.
  To see the options, run: sliceflow slice create --help
  ```
  Shows: R56, R58
- **S40.** **When** I run `sliceflow slice create card-payment extra`,
  **Then** sliceflow changes nothing, prints this to the error output, and
  exits with code 2:
  ```
  Error: unexpected argument "extra". Nothing was changed.
  Usage: sliceflow slice create <name> [options]
  ```
  Shows: R57, R58

## Requirements

<!-- The full rules, including the ones a user never sees.

     - Number them R1, R2, ... from top to bottom.
     - When you are done, renumber them so they read R1, R2, ... from
       top to bottom in this section, and update every link to match:
       the Shows: lines and the decisions' Rule: lines. You may renumber only before the spec is first approved.
     - Group them under small headings, e.g. ### Starting, ### Errors.
     - Write each one in one of these forms (EARS, without "shall"),
       in the present tense:
       - When <trigger>, <system> <response>.
       - If <bad situation>, <system> <response>.
       - While <state>, <system> <response>.
       - Where <feature or setting is on>, <system> <response>.
       - <system> <response>.
       Two keywords may be combined when a rule needs both, e.g.
       While <state>, when <trigger>, <system> <response>.
       For <system>, use the product's name, e.g. sliceflow. Write
       "sliceflow prints", not "sliceflow shall print" or "sliceflow
       must print". For something it must not do, write "never", e.g.
       "sliceflow never sends data over the network."
     - Name exact values: the message printed, the limit, the exit
       code. A worker sees this text, so "say so" is not enough.
     - One behavior per requirement.
     - A term that needs one fixed meaning is defined in the requirement
       where it first appears, e.g. "A task is overdue when its due date
       is before today."
     - Each one has a Check: line saying how it is tested. More than one
       is allowed, e.g. Check: unit, e2e.
       - unit: tests one piece of code on its own.
       - api: calls the program's interface directly, e.g. an HTTP
         endpoint. A plain command-line tool usually has none: use e2e.
       - e2e: runs the real program the way a user does.
       - manual: a human checks it at the preview. Use it only when no
         automatic test can check it.
     - More than about 25-30 requirements? The slice is too big. Split
       it.

     When the spec changes after it was approved:
     - Never reuse a number. A new requirement gets the next free number.
     - A removed requirement stays, as: **R4.** (removed in R002: why)

     Example:
     ### Errors

     - **R3.** If a slice with that name already exists, sliceflow changes
       nothing and prints "Error: slice <name> already exists:
       branch slice/<name>. Nothing was changed."
       Check: e2e

     ### Settings

     - **R9.** Where task review is turned off, sliceflow merges a task
       without starting a reviewer.
       Check: e2e -->

### Creating a slice

- **R1.** When a user runs `sliceflow slice create <name>` in a git
  repository that has commits, and the name is free, sliceflow makes the
  branch `slice/<name>`, a worktree for it, and one commit on that branch.
  The command takes the options `--title`, `--description`, and `--from`.
  Check: e2e
- **R2.** sliceflow adds exactly two files in that commit, both in
  `docs/slices/<name>/` of the worktree: `spec.md`, made from the built-in
  spec template, which has no title line, and `.events.jsonl`, the event
  log.
  Check: e2e
- **R3.** The message of that commit is `<name>: add the spec and the event
  log`.
  Check: e2e
- **R4.** sliceflow never changes the user's own checkout: no new commit,
  no changed file, and the same branch stays checked out.
  Check: e2e
- **R5.** When the slice is created, sliceflow prints the lines shown in S1
  to the normal output, and exits with code 0. The `Branch:` line names
  the base and the first seven characters of its commit. The `Worktree:`
  line is the full path of the worktree. The `Next:` line is the full path
  of the spec file.
  Check: e2e

### The base

- **R6.** sliceflow starts the slice from the branch given with `--from`.
  When `--from` is left out, it starts from `main`.
  Check: e2e
- **R7.** sliceflow starts the slice from the tip of the local base branch
  at that moment. It never fetches, and the slice has no uncommitted
  changes.
  Check: e2e
- **R8.** If the base branch does not exist, sliceflow changes nothing and
  prints this to the error output: the line "Error: base branch <base>
  does not exist. Nothing was changed." Then, when `--from` was left out,
  the line "sliceflow starts from main unless you choose another branch.
  To choose one, run:" and the command `sliceflow slice create <name>
  --from <branch>`, with the slice name filled in. When `--from` was
  given, it prints the line "Check the name. To see your branches, run:"
  and the command `git branch`.
  Check: e2e

### The repository

- **R9.** If the folder is not inside a git repository, sliceflow changes
  nothing and prints these two lines to the error output: "Error: this
  folder is not in a git repository. Nothing was changed." and "Run
  sliceflow inside your project."
  Check: e2e
- **R10.** If the repository has no commits, sliceflow changes nothing and
  prints these two lines to the error output: "Error: this repository has
  no commits yet. Nothing was changed." and "Make a first commit, then try
  again."
  Check: e2e
- **R11.** sliceflow finds the project by asking git, from any subfolder of
  it and from inside any of its worktrees. The project name used in a
  worktree's path is the name of the main project's folder.
  Check: e2e
- **R12.** sliceflow makes the worktree in
  `<home>/.sliceflow/worktrees/<project>/<name>`, where `<home>` is the
  user's home folder. Where the setting `worktree.dir` is set, sliceflow
  uses that folder in place of `<home>/.sliceflow/worktrees`.
  Check: e2e

### The name

- **R13.** Each character of a slice name is a letter, a digit, a hyphen,
  or an underscore. The first character is a letter or a digit. The name
  has at most 40 characters. It is not `con`, `prn`, `aux`, `nul`, `com1`
  to `com9`, or `lpt1` to `lpt9`, which Windows reserves.
  Check: unit
- **R14.** If the name breaks R13, sliceflow changes nothing and prints two
  lines to the error output. The first is "Error: invalid slice name
  "<name>". Nothing was changed." The second depends on the problem: for a
  character or a start that is not allowed, "Allowed characters: letters,
  digits, hyphens, underscores. Start with a letter or a digit."; for a name that
  is too long, "A name has at most 40 characters."; for a reserved name,
  "Windows reserves this name. Choose another."
  Check: unit, e2e
- **R15.** The name the user gives, the given name, may have capital
  letters. sliceflow makes it lowercase, in every command that takes a
  name. The lowercase name is the slice name.
  Check: unit, e2e
- **R16.** When `slice create` made the given name lowercase, it prints the
  line `  Note:     "<given name>" was made lowercase.` right after the
  `Created` line. Otherwise it prints no `Note:` line.
  Check: e2e

### A name that is not free

- **R17.** If `docs/slices/<name>/` exists in the base commit, sliceflow
  changes nothing and prints "Error: slice <name> already exists (folder
  docs/slices/<name>/ in <base>). Nothing was changed." to the error
  output.
  Check: e2e
- **R18.** If the branch `slice/<name>` exists, sliceflow changes nothing
  and prints "Error: slice <name> already exists (branch slice/<name>).
  Nothing was changed." to the error output.
  Check: e2e
- **R19.** If the worktree folder exists and has files in it, sliceflow
  shows git's message, as in R26. If the folder exists and is empty,
  sliceflow uses it.
  Check: e2e
- **R20.** sliceflow never looks at branches on a remote when it checks a
  name.
  Check: e2e

### The event log

- **R21.** The first line of the event log is the event `slice.created`, as
  in S13: `version` 1, `seq` 1, `event`, `at` (UTC, written like
  `2026-11-02T09:15:30Z`), and `data` with `name`, `title`, `description`,
  `base` (the name of the base branch), and `base_commit` (the full commit
  hash).
  Check: unit, e2e
- **R22.** The title is the text of `--title`, or the name when `--title`
  is left out. The description is the text of `--description`, or an empty
  string when it is left out. sliceflow stores them only in the event log.
  Check: unit, e2e
- **R23.** sliceflow adds every event to a log, including the first one,
  with one function, under a file lock that works on Windows, macOS, and
  Linux. When two sliceflow processes add an event at the same time, both
  events are kept, each on its own line.
  Check: unit
- **R24.** When sliceflow adds an event to a log that has events, the
  `seq` of the new event is the `seq` of the last event plus one.
  Check: unit

### When something fails

- **R25.** If a step fails after sliceflow made something, sliceflow
  removes what it made in this run, and only that. It never removes a
  branch, a worktree, or a file that was there before.
  Check: e2e
- **R26.** When git fails and sliceflow has no message of its own for the
  failure, sliceflow prints this to the error output: a first line
  "Error: git could not <step>. The slice was not created. Nothing was
  changed.", then "git said:", then git's text indented by two spaces.
  `<step>` is "create the branch", "create the worktree", or "make the
  commit". sliceflow does not check beforehand for what git needs.
  Check: e2e
- **R27.** If sliceflow could not remove something it made, its first line
  does not say "Nothing was changed." After git's text, it prints "Some of
  what sliceflow made could not be removed:", one line for each thing that
  is left, then "To clean up, run:" and the commands that remove them,
  the worktree first and then the branch.
  Check: e2e

### Seeing one slice

- **R28.** When a user runs `sliceflow slice status <name>` and the log of
  the slice is found, sliceflow prints the status to the normal output,
  changes nothing, and exits with code 0.
  Check: e2e
- **R29.** The first line of the status is `<slice name>: <title>`. The
  next lines have the labels `Description:`, `Created:`, `Started from:`,
  `Branch:`, `Worktree:`, and `Events:`, in this order, with the values in
  one column. The title, the description, the base branch, and the base
  commit come from the `slice.created` event.
  Check: e2e
- **R30.** The `Created:` line shows the time of the `slice.created` event
  in UTC, written like `2026-11-02 09:15 UTC`.
  Check: unit
- **R31.** The `Started from:` line shows the base branch and the first
  seven characters of the base commit, written like `main, at 3f9a1c2`.
  Check: unit
- **R32.** The `Branch:` line shows `slice/<name>`. When that branch does
  not exist, it adds ` (not found)` after it.
  Check: e2e
- **R33.** The `Worktree:` line shows the full path of the slice's
  worktree. When the worktree does not exist, it shows `(not found)`.
  Check: e2e
- **R34.** The `Events:` line shows the number of lines of the log that
  could be read, then in brackets `last:` and the event name of the last
  line that could be read.
  Check: unit
- **R35.** When the log was not read from the worktree, the status ends with
  a `Read from:` line. It shows `branch slice/<name>` when the log came
  from that branch, and `main` when it came from `main`. Otherwise the
  status has no `Read from:` line.
  Check: e2e
- **R36.** sliceflow finds the log of a slice by looking in this order, and
  stops at the first place that has it: the log file in the slice's
  worktree, the log committed on the branch `slice/<name>`, and the log
  committed in `main`.
  Check: e2e
- **R37.** If no log is found, sliceflow changes nothing and prints this to
  the error output: the line "Error: no slice named <name>. Nothing was
  changed.", the line "No log was found in any of these places:", one
  line for each place, with the label `worktree`, `branch`, or `main` and
  its path, then the line "Check the name, or run: sliceflow slice list".
  It exits with code 1.
  Check: e2e
- **R38.** When a line of the log cannot be read, sliceflow skips it, and
  prints one warning for the log to the error output: for one line,
  "Warning: line <n> of <path> could not be read, so it was skipped."; for
  several lines, "Warning: lines <n>, <n>, ... of <path> could not be read,
  so they were skipped." The path is the path of the log in the project,
  for example `docs/slices/card-payment/.events.jsonl`. The command goes
  on, and the exit code stays 0.
  Check: unit, e2e
- **R39.** When more than 10 lines cannot be read, the warning lists the
  first 10 line numbers, then `(<k> more not listed)`, where `<k>` is the
  number of the other lines.
  Check: unit
- **R40.** If no `slice.created` event can be read, sliceflow changes
  nothing and prints the warning, then the line "Error: no slice.created
  event could be read for <name>. Nothing was changed." and the line "Fix
  or restore the log, then try again." to the error output. It exits with
  code 1.
  Check: e2e

### Seeing all slices

- **R41.** When a user runs `sliceflow slice list`, sliceflow finds the
  slices of the project, prints them in a table to the normal output,
  changes nothing, and exits with code 0.
  Check: e2e
- **R42.** sliceflow takes the candidate names from the local branches
  called `slice/<name>` and from the folders `docs/slices/<name>/` in
  `main`. It joins the two lists, and a name that is in both is one
  candidate. It never looks at branches on a remote.
  Check: e2e
- **R43.** For each candidate, sliceflow finds the log as in R36. A candidate
  is a slice when its log is found.
  Check: e2e
- **R44.** The table has the header line `NAME`, `TITLE`, `DESCRIPTION`, then
  one row for each slice, sorted by name. The name column and the title
  column are padded to the width of their longest value, with two spaces
  between columns. The description is the last column, and it is not
  padded. It is empty when the slice has no description.
  Check: unit
- **R45.** The title and the description of a row come from the
  `slice.created` event.
  Check: unit
- **R46.** For the list, sliceflow reads only the first line of each log.
  Check: unit
- **R47.** If the project has no slices, sliceflow changes nothing and
  prints "No slices yet. To make one, run: sliceflow slice create <name>"
  to the normal output. It exits with code 0.
  Check: e2e
- **R48.** If a candidate has no log, sliceflow does not list it, and prints
  one warning for that name to the error output, and the exit code stays
  0. For a branch only: "Warning: <name> is not listed as a slice. The
  branch slice/<name> has no event log." For a folder in `main` only:
  "Warning: <name> is not listed as a slice. The folder docs/slices/<name>/
  in main has no event log." For both: "Warning: <name> is not listed as a
  slice. Neither the branch slice/<name> nor the folder docs/slices/<name>/
  in main has an event log."
  Check: e2e
- **R49.** If the first line of a log cannot be read, sliceflow does not list
  the slice, and prints one warning to the error output: "Warning: <name>
  is not listed as a slice. No slice.created event could be read from
  <path>." The exit code stays 0.
  Check: e2e
- **R50.** sliceflow never prints a warning in the list about a damaged line
  of a log other than the first line.
  Check: e2e

### Help, and a command used the wrong way

- **R51.** When a user runs `sliceflow --help`, or runs `sliceflow` with no
  command, sliceflow prints the guide shown in S33 to the normal output,
  changes nothing, and exits with code 0. The guide has one sentence about
  sliceflow, the usage line, the three commands with a short description
  each, and the line that points to the help of a command.
  Check: e2e
- **R52.** When a user runs a command with `--help`, sliceflow prints to the
  normal output one sentence about the command, its usage line, and its
  options with a short description of each, changes nothing, and exits
  with code 0. Every command has `--help`.
  Check: e2e
- **R53.** If a user runs a command that does not exist, sliceflow changes
  nothing and prints these two lines to the error output: "Error: unknown
  command "<words>". Nothing was changed." and "To see the commands, run:
  sliceflow --help". `<words>` are the words the user typed for the
  command. sliceflow does not guess what the user meant.
  Check: e2e
- **R54.** If a user gives an option that does not exist, sliceflow changes
  nothing and prints these two lines to the error output: "Error: unknown
  option "<option>". Nothing was changed." and "To see the options, run:
  sliceflow <command> --help", with the command the user ran.
  Check: e2e
- **R55.** If a user runs `slice create` or `slice status` with no name,
  sliceflow changes nothing and prints these two lines to the error
  output: "Error: missing slice name. Nothing was changed." and the usage
  line of that command, written like "Usage: sliceflow slice create <name>
  [options]".
  Check: e2e
- **R56.** If a user gives an option that needs a value, with no value
  after it, sliceflow changes nothing and prints these two lines to the
  error output: "Error: option "<option>" needs a value. Nothing was
  changed." and "To see the options, run: sliceflow <command> --help",
  with the command the user ran.
  Check: e2e
- **R57.** If a user gives a word that is not a name and not an option,
  sliceflow changes nothing and prints these two lines to the error
  output: "Error: unexpected argument "<word>". Nothing was changed." and
  the usage line of the command. `<word>` is the first such word.
  Check: e2e
- **R58.** When a command works, sliceflow exits with code 0. When the
  command line is wrong, it exits with code 2. That is the case for an
  unknown command (R53), an unknown option (R54), a missing name (R55), an
  option with no value (R56), an unexpected argument (R57), and a name
  that breaks R13 (R14). When a command fails for any other reason, it
  exits with code 1.
  Check: e2e

## Out of scope

<!-- What this slice does not do, even if it looks related, so nobody
     builds it by mistake. One - bullet per item, one line each. No
     sub-bullets, headings, or code blocks.

     Example:
     - Reminders and notifications.
     - Repeating tasks. -->

- Slice states and what comes next (0002).
- Checking a spec, and printing the template (0002).
- Changing a slice's title or description after it is created (0002).
- The team settings file, `sliceflow.yaml`.
- Worktree setup: copying files, running setup commands, ports (0005).
- Cancelling a slice, and cleaning up after a crash in the middle of
  `slice create` (0015).
- Seeing or checking slices that teammates made on remote branches.
- A git setting for the base branch (`sliceflow.base`).
- A readable view of the event log, for example as Markdown.
- JSON output (`--json`).

## Decisions

<!-- Product decisions made while writing the spec.

     - Number them SD1, SD2, ...
     - Each one has the choice, Why:, and Not chosen: (the main other
       option, and why not).
     - Each one has a Cost: line: the downside we accept with this
       choice. With truly no downside, write Cost: None. A reviewer
       questions a None that looks wrong.
     - Each one has a Rule: line. A worker sees requirements, not
       decisions. So a decision that changes what the feature does is
       also written as a requirement, and Rule: names it, e.g. Rule: R11.
       A decision that changes no behavior says Rule: none. sliceflow checks
       that every decision has this line, and that it names a real
       requirement.
     - To replace a decision from an earlier slice, add Replaces: with
       that slice's name and the decision's number, e.g. device-pairing/SD3.
       Otherwise remove the whole Replaces: line.

     Example:
     - **SD2.** A task due today is not overdue.
       Why: the user still has the whole day to finish it.
       Not chosen: overdue on the due date, because it warns people who
         still have time.
       Cost: a task finished late at night still shows as on time that
         day.
       Rule: R11 -->

- **SD1.** The slice name is an argument, not a flag: `sliceflow slice
  create <name>`. The rule for every command: the slice you act on is the
  argument, and everything else is a flag.
  Why: it is how `gh repo create` and `gh label create` work, it is short
    for a command typed often, and from 0003 the slice name is optional in
    every command, which fits an optional argument.
  Not chosen: a `--name` flag, because it is longer to type and breaks
    from `gh`, although it reads more clearly.
  Cost: the call is less self-describing, and a second required input
    later could be ambiguous.
  Rule: R1
- **SD2.** A new slice starts from the tip of the local base branch as it
  is at that moment. sliceflow never fetches. The `slice.created` event
  records the commit the slice started from.
  Why: it works offline, it is how `git worktree add` behaves, and the
    user or an agent decides when to fetch or pull. The recorded commit
    lets later slices tell exactly what changed since the slice started.
  Not chosen: fetching first and starting from the remote branch, because
    it needs the network and fails offline; starting from the checked-out
    commit, because it could carry unrelated work into the slice.
  Cost: a stale local branch gives a stale start, and uncommitted changes
    are not in the slice.
  Rule: R7
- **SD3.** A slice's worktree is made outside the project, by default
  under `~/.sliceflow/worktrees/<project>/<slice>`. The `worktree.dir`
  setting chooses another folder.
  Why: tools in other people's projects, such as ESLint, Jest, and
    `tsc`, can pick up a worktree nested in the project and give wrong
    results. The setting lets a person who prefers the worktree in their
    project choose that.
  Not chosen: a folder inside the project by default, as Claude Code does,
    because it can break those tools in projects we do not control.
  Cost: a person opens the worktree in a second editor window, unless
    they set the folder themselves.
  Rule: R12
- **SD4.** Each character of a slice name is a letter, a digit, a hyphen,
  or an underscore. The first character is a letter or a digit. The name
  has at most 40 characters. It is not a Windows reserved name such as
  `con`, `nul`, `prn`, `aux`, `com1`, or `lpt1`. Repeated hyphens and a
  trailing hyphen are allowed.
  Why: the name is also the branch `slice/<name>`, the folder, and the
    start of each commit message, and it must work on every system. A
    name that starts with a hyphen would read as a flag. Git and all
    three systems accept repeated hyphens and underscores.
  Not chosen: forbidding repeated hyphens or underscores, because nothing
    needs it; a setting for the allowed characters, because a name made
    on one machine could be refused on a teammate's.
  Cost: a Windows reserved name is refused on every system, even where
    it would work.
  Rule: R13
- **SD5.** sliceflow makes the name lowercase, in every command that
  takes a name, and says so when it changed the name.
  Why: Windows and macOS treat `Card` and `card` as one folder, while git
    treats them as two branches, so two slices differing only by case
    would clash for a teammate on another system.
  Not chosen: keeping the case and refusing a name that clashes in
    case, because it leaves two spellings of one name and needs a more
    complex check.
  Cost: a person or an agent that builds a path from the given name
    gets a wrong path on a case-sensitive system, unless it uses the
    path sliceflow prints.
  Rule: R15
- **SD6.** sliceflow refuses a slice name that is not free. A name is not
  free when `docs/slices/<name>/` exists in the base commit, when the
  branch `slice/<name>` exists, or when the worktree folder exists and
  has files in it. An empty worktree folder is free. For the branch and
  the folder in the base commit, the message says which one it found. For
  the worktree folder, sliceflow shows git's own message (SD8). sliceflow
  does not look at branches on a remote.
  Why: the name is never used again, even for a finished slice (design
    2.1), and git does not stop that: it let a finished slice's name be
    used again in a test. Reading the base commit, not the checkout, gives
    the same answer from any branch. Git accepts an empty folder, so
    refusing it would be a restriction that prevents nothing.
  Not chosen: checking the checked-out folder, because the answer would
    depend on which branch the person is on; checking remote branches,
    because seeing a teammate's slices is a later item (team view), and a
    clash is caught when the branches are merged.
  Cost: a slice a teammate has made on a remote branch can still get the
    same name here.
  Rule: R17, R18, R19, R20

- **SD7.** The base branch of a new slice comes from `--from <branch>`. When
  `--from` is left out, it is `main`. sliceflow has no git setting for the
  base branch in this slice.
  Why: it is the most flexible way, and a person can always name the base.
    A git setting for it can be added later without changing what
    `--from` does, because it would only fill the gap before `main`.
  Not chosen: a required `--from`, because the most common case would be
    longer in every repo; a git setting now, because it is not needed yet.
  Cost: a repo whose main branch is not `main` needs `--from <branch>` on
    every create, until a setting is added.
  Rule: R6

- **SD8.** When git fails in a way sliceflow has no message for, sliceflow
  prints one line of its own and then git's text, word for word. Its line
  says which step failed, that the slice was not created, and, when
  everything it made was removed, "Nothing was changed." Git's text comes
  next. When something could not be removed, the first line does not say
  "Nothing was changed." Instead, the last lines list what is left and
  the commands that remove it. sliceflow does not check beforehand for
  what git needs, such as a commit name and email.
  Why: git can fail in many ways we cannot list, and only git knows what
    it needs on this computer. Its message is right about what it saw, and
    it does not go out of date. sliceflow keeps its own messages for its
    own rules and for cases where git's text is hard to read.
  Not chosen: a message of our own for each git failure, because the list
    has no end and a wrong message is worse than git's.
  Cost: the user reads git's wording, which can differ between git
    versions and languages.
  Rule: R26

- **SD9.** The title of a slice, and its description, live only in the
  event log. The spec file has no title line.
  Why: each fact has one copy (design 1.13), and the log is the only
    source of truth.
  Not chosen: the title in the spec's header as well, because two copies
    can disagree; the title in the spec's header only, because changing it
    would edit Markdown and not add an event.
  Cost: `spec.md` alone does not show its title. A readable view of the
    log, for example as Markdown, can be added later.
  Rule: R22
- **SD10.** A slice created with no `--title` gets its name as the title.
  Why: nothing is ever blank, so `list` and `status` need no special case.
  Not chosen: an empty title, because it leaves a blank column; a
    required `--title`, because every create would need the flag.
  Cost: an untitled slice shows its name twice in the list.
  Rule: R22

- **SD11.** sliceflow finds the project by asking git, as git does: from any
  subfolder of the project, and from inside any worktree of it. The
  project name in a worktree's path is the name of the main project's
  folder, the same from everywhere.
  Why: git already knows where the project is, and it does not refuse to
    work from a subfolder or from another worktree. One project name keeps
    all the slices of a project in one folder, and keeps two projects with
    a slice of the same name apart.
  Not chosen: the name of the folder the person is in, because from inside
    another slice's worktree it would scatter one project's slices over
    several folders.
  Cost: one more question to git; a project whose folder is renamed gets
    a new worktree folder for its next slices.
  Rule: R11

- **SD12.** sliceflow adds every event to a log with one function, under a
  file lock, including the first event, `slice.created`.
  Why: two ways to write an event, one with the lock and one without,
    would leave a trap for later code. One rule has no exceptions.
  Not chosen: no lock for the first event, because `create` writes a new
    file that nothing else can be writing to yet, since a second way of
    writing an event is a risk, and the lock is needed in 0002 anyway.
  Cost: the lock, which must work on Windows too, is built in 0001 and
    not in 0002.
  Rule: R23

- **SD13.** sliceflow finds a slice's event log by looking in this order,
  and stops at the first place that has it: the log file in the slice's
  worktree, the log committed on the slice branch, and the log committed
  in `main`. One function does this, and `status`, `list`, and later the
  dashboard all use it. Whether a slice is merged is worked out from git,
  not stored in the log.
  Why: the log has the same path in every place, and the history of a
    slice must stay readable after it is merged and its branch and
    worktree are gone. One function means no command makes up its own
    rule.
  Not chosen: a local index of where each log is, because it can be added
    later as a speed-up; keeping the log in git refs, because the log
    would no longer be committed with the slice and visible in a pull
    request.
  Cost: three places to look; a slice merged into a branch other than
    `main` is not found.
  Rule: R36

- **SD14.** The names for `list` come from the union of the local branches
  `slice/<name>` and the folders `docs/slices/<name>/` in `main`. A name is a
  slice only when its log is found, by the gates of SD13.
  Why: a merged slice often has no branch, and a slice being built has no
    folder in `main` yet, so each list alone would miss some. They are the
    same signs that `create` uses to refuse a taken name (SD6), so the two
    commands agree.
  Not chosen: only the branches, because it would hide every finished
    slice; also scanning the worktree folders, because every worktree is on
    a slice branch, so it adds nothing.
  Cost: a slice merged into a branch other than `main` is not listed.
  Rule: R42, R43

- **SD15.** `list` prints a borderless table with a header row, sorted by
  name.
  Why: plain columns work in every terminal, are easy to scan, and are easy
    for a script or an agent to read. The name order is the same every
    time and needs no extra reading.
  Not chosen: a table drawn with box lines, because wrapped text breaks it
    and older Windows consoles can show the lines wrongly; one block for
    each slice, because the slices are hard to tell apart; an order by
    creation time, because every log would have to be read, and two slices
    made in the same minute have no clear order.
  Cost: a very long title pushes the description to the right. Nothing is
    cut off.
  Rule: R44

- **SD16.** In `list`, only the first line of each log counts, and only a
  damaged first line gives a warning.
  Why: the first line holds everything `list` shows, a warning for every
    damaged line would flood the screen, and reading one line is cheaper.
  Not chosen: a warning for any damaged line, as in `status`, because
    `status` shows the number of events and the last event, and `list` does
    not.
  Cost: damage after the first line is not seen in `list`; `status` shows
    it. It is worth a second look when `list` shows state from later
    events.
  Rule: R46, R50

- **SD17.** sliceflow exits with 0 when a command works, with 2 when the
  command line is wrong, and with 1 when a command fails for any other
  reason.
  Why: a script or an agent can tell "I called it wrong, so I should fix
    my call" from "I called it right and the work failed", which need
    different reactions. Many Unix tools use 2 for wrong use.
  Not chosen: one code for every failure, because the two cases look the
    same; a different code for each kind of error, because a long table
    must be remembered.
  Cost: two numbers to remember. A bad slice name counts as wrong use,
    which is a border case.
  Rule: R58

- **SD18.** When a command is used the wrong way, sliceflow says which
  word or option it did not understand, and points to the help. It does
  not guess what the user meant.
  Why: with three commands, the help is short enough to read, and a guess
    can be wrong. Every error of sliceflow's own has the same two lines:
    what is wrong, with "Nothing was changed.", then what to do.
  Not chosen: a "did you mean" suggestion, because it needs more logic
    that can guess wrong.
  Cost: a typo costs the user one more command, to read the help.
  Rule: R53, R54, R55, R56, R57

## Open questions

<!-- Questions for the human, numbered Q1, Q2, ... An answered question
     is written into the spec and removed from here. A deferred question
     that does not block the build moves to Out of scope and the backlog.
     The spec review starts only when this section says None. -->

None.
