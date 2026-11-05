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
  Shows: R1, R2, R8, R10, R11, R16, R17
- **S2.** **When** I run `sliceflow slice create card.payment`,
  **Then** sliceflow changes nothing, prints this to the error output, and
  exits with code 2:
  ```
  Error: invalid slice name "card.payment". Nothing was changed.
  Allowed characters: letters, digits, hyphens, underscores. Start with a letter or a digit.
  ```
  Shows: R19, R20, R76
- **S3.** **Given** a slice named `card-payment` already exists, with the
  branch `slice/card-payment`,
  **When** I run `sliceflow slice create card-payment`,
  **Then** sliceflow changes nothing, prints this to the error output, and
  exits with code 1:
  ```
  Error: slice card-payment already exists (branch slice/card-payment). Nothing was changed.
  ```
  Shows: R24
- **S4.** **Given** a slice named `card-payment` was finished and merged into
  `main` earlier, so its branch is gone but the folder
  `docs/slices/card-payment/` is in `main`,
  **When** I run `sliceflow slice create card-payment`,
  **Then** sliceflow changes nothing, prints this to the error output, and
  exits with code 1:
  ```
  Error: slice card-payment already exists (folder docs/slices/card-payment/ in main). Nothing was changed.
  ```
  Shows: R23
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
  Shows: R8, R21, R22
- **S6.** **Given** I am in a folder that is not inside a git repository,
  **When** I run `sliceflow slice create card-payment`,
  **Then** sliceflow changes nothing, prints this to the error output, and
  exits with code 1:
  ```
  Error: this folder is not in a git repository. Nothing was changed.
  Run sliceflow inside your project.
  ```
  Shows: R13
- **S7.** **Given** I am in a git repository that has no commits yet,
  **When** I run `sliceflow slice create card-payment`,
  **Then** sliceflow changes nothing, prints this to the error output, and
  exits with code 1:
  ```
  Error: this repository has no commits yet. Nothing was changed.
  Make a first commit, then try again.
  ```
  Shows: R15
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
  Shows: R12
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
  Shows: R10, R11
- **S10.** **Given** the folder `/home/you/.sliceflow/worktrees/shop/card-payment`
  already exists and is not empty,
  **When** I run `sliceflow slice create card-payment`,
  **Then** sliceflow removes the branch it made, prints this to the error
  output, and exits with code 1:
  ```
  Error: git could not create the worktree. The slice was not created. Nothing was changed.
  git said:
    Preparing worktree (checking out 'slice/card-payment')
    fatal: '/home/you/.sliceflow/worktrees/shop/card-payment' already exists
  ```
  Shows: R25, R36, R37
- **S11.** **Given** git cannot work out a name and an email on this
  computer,
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

    to set your account's default identity.
    Omit --global to set the identity only in this repository.

    fatal: unable to auto-detect email address (got 'you@laptop.(none)')
  ```
  Shows: R36, R37
- **S12.** **Given** git cannot work out a name and an email on this
  computer, and
  the worktree sliceflow made cannot be removed (for example, sliceflow is not
  allowed to delete it),
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

    to set your account's default identity.
    Omit --global to set the identity only in this repository.

    fatal: unable to auto-detect email address (got 'you@laptop.(none)')
  Some of what sliceflow made could not be removed:
    worktree  /home/you/.sliceflow/worktrees/shop/card-payment
    branch    slice/card-payment
  To clean up, run:
    git worktree remove --force "/home/you/.sliceflow/worktrees/shop/card-payment"
    git branch -D "slice/card-payment"
  ```
  Shows: R41
- **S13.** **Given** I created the slice `card-payment` with a title and a
  description, as in S1, on 2 November 2026 at 09:15:30 UTC,
  **When** I open the event log in the slice's folder,
  **Then** it has exactly one line, and the same line is in the slice
  branch:
  ```
  {"version":1,"seq":1,"event":"slice.created","at":"2026-11-02T09:15:30Z","data":{"name":"card-payment","title":"Pay with a card","description":"Let a customer pay with a credit or debit card.","base":"main","base_commit":"3f9a1c2e8b7d4a6f0c1e5d9b2a7f3c8e4d6b1a09"}}
  ```
  Shows: R27
- **S14.** **Given** I created the slice `card-payment`, as in S1,
  **When** I look at the slice branch with
  `git show --name-only --format=%s slice/card-payment`,
  **Then** it has one new commit on top of `main`, and my own checkout has
  no new commit. The commit has this message, and it holds exactly these
  two files:
  ```
  card-payment: add the spec and the event log

  docs/slices/card-payment/.events.jsonl
  docs/slices/card-payment/spec.md
  ```
  Shows: R2, R4, R7
- **S15.** **Given** I am in the project `shop`, which has commits and a
  `main` branch,
  **When** I run `sliceflow slice create card-payment` with no `--title`
  and no `--description`, on 2 November 2026 at 09:20:05 UTC,
  **Then** sliceflow creates the slice, the title is the name, the
  description is empty, and the event log has this line:
  ```
  {"version":1,"seq":1,"event":"slice.created","at":"2026-11-02T09:20:05Z","data":{"name":"card-payment","title":"card-payment","description":"","base":"main","base_commit":"3f9a1c2e8b7d4a6f0c1e5d9b2a7f3c8e4d6b1a09"}}
  ```
  Shows: R27, R28
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
  Shows: R1, R16
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
  Shows: R12
- **S18.** **Given** I created the slice `card-payment` on 2 November 2026,
  **When** I open `spec.md` in the slice's folder,
  **Then** it starts with this header, followed by the rest of the built-in
  spec template, with its comments:
  ```
  ---
  slice: card-payment
  created: 2026-11-02
  ---
  ```
  Shows: R2, R3
- **S19.** **Given** I ran
  `git config --global sliceflow.worktree.dir ../shop-worktrees`, and the
  project `shop` is in `/home/you/projects/shop`,
  **When** I run `sliceflow slice create card-payment`,
  **Then** sliceflow makes the worktree next to the project, prints this,
  and exits with code 0:
  ```
  Created slice card-payment.
    Branch:   slice/card-payment (from main, at 3f9a1c2)
    Worktree: /home/you/projects/shop-worktrees/shop/card-payment
  Next: write the spec in /home/you/projects/shop-worktrees/shop/card-payment/docs/slices/card-payment/spec.md
  ```
  Shows: R18
- **S20.** **When** I run
  `sliceflow slice create card-payment --title "Pay with a card` and press
  Enter inside the quotes, then finish the title with `"` on the next line,
  **Then** sliceflow changes nothing, prints this to the error output, and
  exits with code 2:
  ```
  Error: --title has a control character. Nothing was changed.
  Received: "Pay with a card\n"
  Write it as plain text on one line, with no control characters.
  ```
  Shows: R29, R76
- **S21.** **Given** git is not installed, or is not on the path,
  **When** I run `sliceflow slice create card-payment`,
  **Then** sliceflow changes nothing, prints this to the error output, and
  exits with code 1:
  ```
  Error: git was not found. Nothing was changed.
  Install git, then try again.
  ```
  Shows: R14
- **S22.** **Given** the empty folder
  `/home/you/.sliceflow/worktrees/shop/card-payment` exists, and git cannot
  work out a name and an email on this computer,
  **When** I run `sliceflow slice create card-payment`,
  **Then** sliceflow removes the files, the worktree, and the branch it
  made, prints the same error as in S11, and exits with code 1, and the
  folder is still there, and still empty.
  Shows: R36
- **S23.** **Given** the `.gitignore` of the project `shop` has the line
  `*.jsonl`,
  **When** I run `sliceflow slice create card-payment`,
  **Then** sliceflow creates the slice with both files in its commit,
  prints this, and exits with code 0:
  ```
  Created slice card-payment.
    Note:     .events.jsonl matches a .gitignore rule, so sliceflow added it anyway.
    Branch:   slice/card-payment (from main, at 3f9a1c2)
    Worktree: /home/you/.sliceflow/worktrees/shop/card-payment
  Next: write the spec in /home/you/.sliceflow/worktrees/shop/card-payment/docs/slices/card-payment/spec.md
  ```
  Shows: R5, R8
- **S24.** **Given** a git hook makes the commit wait,
  **When** I run `sliceflow slice create card-payment` and press Ctrl+C
  while it waits,
  **Then** sliceflow removes the files, the worktree, and the branch it
  made, prints this to the error output, and exits with code 1:
  ```
  Stopped. The slice was not created. Nothing was changed.
  ```
  Shows: R39

### Story 2: See one slice

A developer checks one slice, and its history, even after it is merged.

- **S25.** **Given** I created the slice `card-payment` with a title and a
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
  Shows: R42, R43, R45, R46, R47, R48, R49, R51
- **S26.** **Given** I created the slice `card-payment`, and it was merged
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
  Shows: R42, R47, R48, R50, R51
- **S27.** **Given** no slice named `card-payment` exists, and no log for it
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
  Shows: R52
- **S28.** **Given** I created the slice `card-payment`, and the second line
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
  Shows: R53
- **S29.** **Given** the event log of `card-payment` has three lines, and
  lines 2 and 3 cannot be read,
  **When** I run `sliceflow slice status card-payment`,
  **Then** sliceflow skips those lines, prints one warning to the error
  output and the status to the normal output, changes nothing, and exits
  with code 0:
  ```
  Warning: lines 2, 3 of docs/slices/card-payment/.events.jsonl could not be read, so they were skipped.
  card-payment: Pay with a card
    Description:  Let a customer pay with a credit or debit card.
    Created:      2026-11-02 09:15 UTC
    Started from: main, at 3f9a1c2
    Branch:       slice/card-payment
    Worktree:     /home/you/.sliceflow/worktrees/shop/card-payment
    Events:       1 (last: slice.created)
  ```
  Shows: R49, R53
- **S30.** **Given** the event log of `card-payment` has 13 lines, and lines
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
  Shows: R49, R53, R54
- **S31.** **Given** the first line of the event log of `card-payment` is
  cut off, so no `slice.created` event can be read,
  **When** I run `sliceflow slice status card-payment`,
  **Then** sliceflow changes nothing, prints this to the error output, and
  exits with code 1:
  ```
  Warning: line 1 of docs/slices/card-payment/.events.jsonl could not be read, so it was skipped.
  Error: no slice.created event could be read for card-payment. Nothing was changed.
  Fix or restore the log, then try again.
  ```
  Shows: R53, R55
- **S32.** **Given** I created the slice `card-payment` with a title and no
  description, and I removed its worktree, but its branch
  `slice/card-payment` is still there,
  **When** I run `sliceflow slice status card-payment`,
  **Then** sliceflow reads the log from the branch, prints this to the
  normal output, changes nothing, and exits with code 0:
  ```
  card-payment: Pay with a card
    Description:  (none)
    Created:      2026-11-02 09:15 UTC
    Started from: main, at 3f9a1c2
    Branch:       slice/card-payment
    Worktree:     (not found)
    Events:       1 (last: slice.created)
    Read from:    branch slice/card-payment
  ```
  Shows: R42, R44, R48, R50, R51

### Story 3: See all slices

A developer sees all the slices of the project, including the merged ones,
in one table.

- **S33.** **Given** I am in the project `shop`, which has three slices:
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
  Shows: R56, R57, R58, R59, R60, R61
- **S34.** **Given** the project has the slice `card-payment`, and a branch
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
  Shows: R63
- **S35.** **Given** the project has the slice `card-payment`, and a folder
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
  Shows: R63
- **S36.** **Given** the project has the slice `card-payment`, and both a
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
  Shows: R63
- **S37.** **Given** the project has no slices,
  **When** I run `sliceflow slice list`,
  **Then** sliceflow prints this to the normal output, changes nothing, and
  exits with code 0:
  ```
  No slices yet. To make one, run: sliceflow slice create <name>
  ```
  Shows: R62
- **S38.** **Given** the slice `card-payment` was merged into `main` by hand,
  and its branch `slice/card-payment` was not deleted, so both the branch
  and the folder `docs/slices/card-payment/` in `main` have its log,
  **When** I run `sliceflow slice list`,
  **Then** sliceflow shows `card-payment` once, prints no warning, changes
  nothing, and exits with code 0:
  ```
  NAME          TITLE            DESCRIPTION
  card-payment  Pay with a card  Let a customer pay with a credit or debit card.
  ```
  Shows: R57, R58
- **S39.** **Given** the project has the slices `card-payment` and
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
  Shows: R65, R66
- **S40.** **Given** the project has the slice `card-payment`, and a branch
  `slice/Foo` that someone made by hand,
  **When** I run `sliceflow slice list`,
  **Then** sliceflow prints one warning to the error output and the table
  to the normal output, changes nothing, and exits with code 0:
  ```
  Warning: Foo is not listed as a slice. It is not a valid slice name.
  NAME          TITLE            DESCRIPTION
  card-payment  Pay with a card  Let a customer pay with a credit or debit card.
  ```
  Shows: R64

### Story 4: Get help, and use a command wrongly

A developer asks sliceflow for help, and learns what went wrong when a
command is used the wrong way.

- **S41.** **When** I run `sliceflow --help`,
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
  Shows: R67
- **S42.** **When** I run `sliceflow` with nothing after it,
  **Then** sliceflow prints the same guide as in S41 to the normal output,
  changes nothing, and exits with code 0.
  Shows: R67
- **S43.** **When** I run `sliceflow slice` with nothing after it,
  **Then** sliceflow prints the same guide as in S41 to the normal output,
  changes nothing, and exits with code 0.
  Shows: R67
- **S44.** **When** I run `sliceflow slice create --help`,
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
  Shows: R68
- **S45.** **When** I run `sliceflow slice status --help`,
  **Then** sliceflow prints this to the normal output, changes nothing, and
  exits with code 0:
  ```
  Show one slice: its title, description, start, branch, worktree, and events.

  Usage:
    sliceflow slice status <name> [options]

  Options:
    --help  Show this help.
  ```
  Shows: R68, R69
- **S46.** **When** I run `sliceflow slice delete card-payment`,
  **Then** sliceflow changes nothing, prints this to the error output, and
  exits with code 2:
  ```
  Error: unknown command "slice delete". Nothing was changed.
  To see the commands, run: sliceflow --help
  ```
  Shows: R70, R76
- **S47.** **When** I run `sliceflow slice create card-payment --titel "Pay with a card"`,
  **Then** sliceflow changes nothing, prints this to the error output, and
  exits with code 2:
  ```
  Error: unknown option "--titel". Nothing was changed.
  To see the options, run: sliceflow slice create --help
  ```
  Shows: R72, R76
- **S48.** **When** I run `sliceflow slice create` with no name,
  **Then** sliceflow changes nothing, prints this to the error output, and
  exits with code 2:
  ```
  Error: missing slice name. Nothing was changed.
  Usage: sliceflow slice create <name> [options]
  ```
  Shows: R73, R76
- **S49.** **When** I run `sliceflow slice create card-payment --title` with
  no text after `--title`,
  **Then** sliceflow changes nothing, prints this to the error output, and
  exits with code 2:
  ```
  Error: option "--title" needs a value. Nothing was changed.
  To see the options, run: sliceflow slice create --help
  ```
  Shows: R74, R76
- **S50.** **When** I run `sliceflow slice create card-payment extra`,
  **Then** sliceflow changes nothing, prints this to the error output, and
  exits with code 2:
  ```
  Error: unexpected argument "extra". Nothing was changed.
  Usage: sliceflow slice create <name> [options]
  ```
  Shows: R75, R76

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
  repository that has commits, and the name is free (R23, R24, R25),
  sliceflow makes the branch `slice/<name>`, a worktree for it, and one
  commit on that branch. The command takes the options `--title`,
  `--description`, and `--from`.
  Check: e2e
- **R2.** sliceflow adds exactly two files in that commit, both in
  `docs/slices/<name>/` of the worktree: `spec.md` and `.events.jsonl`, the
  event log.
  Check: e2e
- **R3.** The file `spec.md` is the built-in spec template, with its
  comments, and with this header: the line `slice: <name>`, and the line
  `created: <date>`, where `<date>` is the day the slice was created, in
  UTC, written like `2026-11-02`. The template has no title line.
  Check: unit, e2e
- **R4.** The message of that commit is `<name>: add the spec and the event
  log`.
  Check: e2e
- **R5.** sliceflow adds its two files to the commit even when a
  `.gitignore` rule matches them, and it does this only for those two
  files. When a rule matched one of them, the output of R8 has the line
  `  Note:     <file> matches a .gitignore rule, so sliceflow added it anyway.`
  for each such file, where `<file>` is `spec.md` or `.events.jsonl`.
  Check: e2e
- **R6.** sliceflow makes the commit with the user's own git settings, so
  git hooks and commit signing work as the user set them up.
  Check: e2e
- **R7.** sliceflow never changes the user's own checkout: no new commit,
  no changed file, and the same branch stays checked out. The one
  exception is a worktree folder that `sliceflow.worktree.dir` puts inside
  the project (R18).
  Check: e2e
- **R8.** When the slice is created, sliceflow prints these lines to the
  normal output, and exits with code 0: `Created slice <name>.`, then
  `  Branch:   slice/<name> (from <base>, at <commit>)`, where `<commit>` is
  the first seven characters of the base commit, then
  `  Worktree: <path>`, the full path of the worktree, then
  `Next: write the spec in <path>`, the full path of `spec.md` in the
  worktree. The `Note:` lines of R22 and R5 come right after the first
  line, in that order, when they apply. sliceflow prints nothing else:
  what git prints when a step works is not shown.
  Check: e2e
- **R9.** Full paths that sliceflow prints use the separator of the
  system, for example `\` on Windows. Paths inside the project, such as
  `docs/slices/<name>/.events.jsonl`, always use `/`.
  Check: e2e

### The base

- **R10.** sliceflow starts the slice from the branch given with `--from`,
  which takes the name of a local branch. When `--from` is left out, it
  starts from `main`. A tag, a commit, or a remote branch is not a local
  branch, so R12 applies to it.
  Check: e2e
- **R11.** sliceflow starts the slice from the commit at the tip of the local
  base branch at that moment. It never fetches. Changes that are not
  committed on the base branch are not in the slice.
  Check: e2e
- **R12.** If the base branch does not exist, sliceflow changes nothing and
  prints this to the error output: the line "Error: base branch <base>
  does not exist. Nothing was changed." Then, when `--from` was left out,
  the line "sliceflow starts from main unless you choose another branch.
  To choose one, run:" and the command `sliceflow slice create <name>
  --from <branch>`, with the slice name filled in. When `--from` was
  given, it prints the line "Check the name. To see your branches, run:"
  and the command `git branch`. Each command is on its own line, indented
  by two spaces.
  Check: e2e

### The repository

- **R13.** If the folder is not inside a git repository, sliceflow changes
  nothing and prints these two lines to the error output: "Error: this
  folder is not in a git repository. Nothing was changed." and "Run
  sliceflow inside your project." This holds for `slice create`,
  `slice status`, and `slice list`.
  Check: e2e
- **R14.** If git cannot be run, because it is not installed or not on the
  path, sliceflow changes nothing and prints these two lines to the error
  output: "Error: git was not found. Nothing was changed." and "Install
  git, then try again." It exits with code 1. sliceflow does not check
  the version of git. This holds for `slice create`, `slice status`, and
  `slice list`.
  Check: e2e
- **R15.** If the repository has no commits, sliceflow changes nothing and
  prints these two lines to the error output: "Error: this repository has
  no commits yet. Nothing was changed." and "Make a first commit, then try
  again." This holds for `slice create`, `slice status`, and
  `slice list`.
  Check: e2e
- **R16.** sliceflow finds the project by asking git, from any subfolder of
  it and from inside any of its worktrees. The project name used in a
  worktree's path is the name of the main project's folder.
  Check: e2e
- **R17.** When the git setting `sliceflow.worktree.dir` is not set,
  sliceflow makes the worktree in
  `<home>/.sliceflow/worktrees/<project>/<name>`, where `<home>` is the
  user's home folder.
  Check: e2e
- **R18.** When the git setting `sliceflow.worktree.dir` is set, sliceflow
  makes the worktree in `<dir>/<project>/<name>`, where `<dir>` is its
  value. sliceflow reads the value through git, so a value set for the
  project wins over a global one. `<dir>` may be a full path, a path that
  starts with `~` (the user's home folder), or a path relative to the main
  project's folder. `~` means the home folder only at the start of the
  value, as `~`, `~/`, or `~\`. An empty value counts as not set. The path
  that sliceflow prints has no `.` or `..` parts. sliceflow does not check
  the folder beforehand. If git cannot make the worktree there, R37
  applies.
  Check: e2e

### The name

- **R19.** A letter is one of `a` to `z` or `A` to `Z`. Each character of a
  slice name is a letter, a digit, a hyphen, or an underscore. The first
  character is a letter or a digit. The name has at most 40 characters.
  It is not `con`, `prn`, `aux`, `nul`, `com1` to `com9`, or `lpt1` to
  `lpt9`, which Windows reserves.
  Check: unit
- **R20.** If the name breaks R19, sliceflow changes nothing and prints two
  lines to the error output. The first is "Error: invalid slice name
  "<name>". Nothing was changed.", where `<name>` is the given name. The
  second depends on the problem: for a character or a start that is not
  allowed, "Allowed characters: letters, digits, hyphens, underscores.
  Start with a letter or a digit."; for a name that is too long, "A name
  has at most 40 characters."; for a reserved name, "Windows reserves this
  name. Choose another." When a name has more than one problem, the second
  line is for the first one in that order. This holds for `slice create`
  and `slice status`.
  Check: unit, e2e
- **R21.** The name the user gives, the given name, may have capital
  letters. In `slice create` and `slice status`, sliceflow first makes the
  letters `A` to `Z` lowercase, then checks the name with R19, and from then on uses the lowercase
  name, the slice name.
  Check: unit, e2e
- **R22.** When `slice create` made the given name lowercase, it prints the
  line `  Note:     "<given name>" was made lowercase.` right after the
  `Created` line. Otherwise it prints no `Note:` line. `slice status`
  prints no `Note:` line.
  Check: e2e

### A name that is not free

- **R23.** If `docs/slices/<name>/` exists in the base commit, sliceflow
  changes nothing and prints "Error: slice <name> already exists (folder
  docs/slices/<name>/ in <base>). Nothing was changed." to the error
  output.
  Check: e2e
- **R24.** If the branch `slice/<name>` exists, sliceflow changes nothing
  and prints "Error: slice <name> already exists (branch slice/<name>).
  Nothing was changed." to the error output.
  Check: e2e
- **R25.** If the worktree folder exists and is not empty, git refuses to
  make the worktree, and sliceflow shows git's message, as in R37. If the
  folder exists and is empty, sliceflow uses it.
  Check: e2e
- **R26.** sliceflow never looks at branches on a remote when it checks a
  name.
  Check: e2e

### The event log

- **R27.** The first line of the event log is the event `slice.created`, as
  in S13: `version` 1, `seq` 1, `event`, `at` (UTC, written like
  `2026-11-02T09:15:30Z`), and `data` with `name`, `title`, `description`,
  `base` (the name of the base branch), and `base_commit` (the full hash
  of the commit the slice started from). The keys are in the order shown
  in S13. The line is UTF-8 JSON, and the characters `<`, `>`, and `&` are
  written as they are.
  Check: unit, e2e
- **R28.** The title is the text of `--title`, or the name when `--title`
  is left out or empty. The description is the text of `--description`,
  or an empty string when it is left out. sliceflow stores them only in
  the event log.
  Check: unit, e2e
- **R29.** If `--title` or `--description` holds a control character,
  sliceflow changes nothing and prints these three lines to the error
  output: "Error: <option> has a control character. Nothing was changed.",
  then "Received: "<text>"", then "Write it as plain text on one line, with
  no control characters.", where `<option>` is `--title` or
  `--description`. A control character is a character with the code 0 to
  31, 127, or 128 to 159. A line break and a tab are control characters.
  In `<text>`, a line break is written `\n`, a carriage return `\r`, a tab
  `\t`, a backslash `\\`, and any other control character as `\u` with its
  code in four hex digits, such as `\u001b`. When both options have one,
  `--title` is reported.
  Check: unit, e2e
- **R30.** When two sliceflow processes add an event to the same log at the
  same time, on Windows, macOS, or Linux, both events are kept, each one
  whole and on its own line.
  Check: unit
- **R31.** When sliceflow adds an event to a log that has events, the `seq`
  of the new event is the highest `seq` that can be read in the log, plus
  one.
  Check: unit
- **R32.** A line of the log can be read when it is one JSON object with
  `version` 1, a text `event`, a whole-number `seq`, a text `at` that is
  a UTC time written like `2026-11-02T09:15:30Z`, and an object `data`.
  A line with an event name that sliceflow does not know can still be
  read. An empty line, or any other line, cannot be read. Lines are
  counted from 1, and every line of the file counts. The text after the
  last `\n` is a line only when it is not empty. A `slice.created` line
  counts as readable only when its `data` also has `name`, `title`,
  `description`, `base`, and `base_commit`.
  Check: unit
- **R33.** sliceflow ends every line it writes with `\n`. When it reads, a
  line may end with `\n` or with `\r\n`.
  Check: unit
- **R34.** Before sliceflow adds an event, if the log is not empty and does
  not end with `\n`, it first adds `\n`, so a cut-off last line stays a line of its
  own.
  Check: unit
- **R35.** sliceflow never deletes or changes a line of a log.
  Check: unit

### When something fails

- **R36.** If a step fails after sliceflow made something, sliceflow
  removes what it made in this run, and only that: the files, the
  worktree, the branch, and the empty folders that were made for the
  worktree. It never removes a branch, a worktree, a file, or a folder that
  was there before, even an empty one.
  Check: e2e
- **R37.** sliceflow takes these steps with git, in this order: create the
  branch, create the worktree, then make the commit, which includes adding
  the two files. When one of them fails and sliceflow has no message of
  its own for the failure, sliceflow prints this to the error output: a
  first line "Error: git could not <step>. The slice was not created.
  Nothing was changed.", where `<step>` is "create the branch", "create
  the worktree", or "make the commit", then "git said:", then everything
  git printed to its error output, word for word, each line indented by
  two spaces. An empty line in git's text stays empty, with no spaces.
  When git printed nothing, the line "git printed no message." takes the
  place of "git said:" and the text. sliceflow does not check beforehand
  for what git needs.
  Check: e2e
- **R38.** If sliceflow cannot write a file of the slice, or take the lock
  of the log, it prints this to the error output: the line "Error: could
  not write <path>. The slice was not created. Nothing was changed.", then
  "The system said:", then the error text of the system, indented by two
  spaces. R36 and R41 apply, as for a git step.
  Check: unit
- **R39.** If the user stops `slice create` with Ctrl+C, sliceflow removes
  what it made, as in R36, prints "Stopped. The slice was not created.
  Nothing was changed." to the error output, and exits with code 1. When
  something could not be removed, R41 applies.
  Check: e2e
- **R40.** When more than one problem applies, sliceflow reports only the
  first, in this order: the command line (R70 to R75), where `--help`
  comes before any error, the first wrong word from the left is reported,
  and a missing name is checked last; a control character (R29); the name
  (R20); git not found (R14); the repository (R13, R15); the base (R12); a
  name that is not free, the folder in the base commit before the branch
  (R23, R24); then the steps that make the slice (R37, R38).
  `slice status` and `slice list` use the same order, for the checks they
  have.
  Check: e2e
- **R41.** If sliceflow could not remove something it made, its first line
  does not say "Nothing was changed." After git's text, or the system's
  text, it prints the line
  "Some of what sliceflow made could not be removed:", then one line for
  each thing that is left: `  worktree  <full path>` or
  `  branch    slice/<name>`. Then it prints the line "To clean up, run:",
  and for each thing that is left, the worktree first, the command
  `  git worktree remove --force "<full path>"` or
  `  git branch -D "slice/<name>"`. Paths and branch names are in double
  quotes.
  Check: e2e

### Seeing one slice

- **R42.** When a user runs `sliceflow slice status <name>` and the log of
  the slice is found, sliceflow prints the status to the normal output,
  changes nothing, and exits with code 0.
  Check: e2e
- **R43.** The first line of the status is `<slice name>: <title>`. The
  next lines have the labels `Description:`, `Created:`, `Started from:`,
  `Branch:`, `Worktree:`, and `Events:`, in this order, with the values in
  one column. The title, the description, the base branch, and the base
  commit come from the `slice.created` event.
  Check: e2e
- **R44.** When the description is empty, the `Description:` line shows
  `(none)`.
  Check: unit
- **R45.** The `Created:` line shows the time of the `slice.created` event
  in UTC, written like `2026-11-02 09:15 UTC`.
  Check: unit
- **R46.** The `Started from:` line shows the base branch and the first
  seven characters of the base commit, written like `main, at 3f9a1c2`.
  Check: unit
- **R47.** The `Branch:` line shows `slice/<name>`. When that branch does
  not exist, it adds ` (not found)` after it.
  Check: e2e
- **R48.** The `Worktree:` line shows the full path of the slice's
  worktree. The slice's worktree is the worktree that git lists for the
  branch `slice/<name>`, when its folder exists. When there is none, the
  line shows `(not found)`.
  Check: e2e
- **R49.** The `Events:` line shows the number of lines of the log that
  can be read, then `(last: <event>)`, where `<event>` is the event name
  of the last line that can be read.
  Check: unit
- **R50.** When the log was not read from the worktree, the status ends with
  a `Read from:` line. It shows `branch slice/<name>` when the log came
  from that branch, and `main` when it came from `main`. Otherwise the
  status has no `Read from:` line.
  Check: e2e
- **R51.** sliceflow finds the log of a slice by looking in these places,
  in this order, and stops at the first place that has it: the log file in
  the slice's worktree (R48), the log committed at the tip of the branch
  `slice/<name>`, and the log committed at the tip of `main`. When `main`
  does not exist, sliceflow treats it as having no log.
  Check: e2e
- **R52.** If no log is found, sliceflow changes nothing and prints this to
  the error output: the line "Error: no slice named <name>. Nothing was
  changed.", the line "No log was found in any of these places:", then
  the three lines `  worktree  <path>`, where `<path>` is the full path
  that R17 or R18 gives, `  branch    slice/<name>`, and
  `  main      docs/slices/<name>/.events.jsonl`, then the line "Check
  the name, or run: sliceflow slice list". It exits with code 1.
  Check: e2e
- **R53.** In `slice status`, when a line of the log cannot be read,
  sliceflow skips it, and prints one warning for the log to the error output: for one line,
  "Warning: line <n> of <path> could not be read, so it was skipped."; for
  several lines, "Warning: lines <n>, <n>, ... of <path> could not be read,
  so they were skipped." Line numbers are counted as in R32. The path is
  the path of the log in the project, for example
  `docs/slices/card-payment/.events.jsonl`. The command goes on. The
  warning alone does not change the exit code.
  Check: unit, e2e
- **R54.** When more than 10 lines cannot be read, the warning lists the
  first 10 line numbers, then `(<k> more not listed)`, where `<k>` is the
  number of the other lines.
  Check: unit
- **R55.** If the log is empty, or its first line cannot be read, or its
  first line is not a `slice.created` event, sliceflow changes nothing and
  prints the warning of R53 when there is one, then the line "Error: no
  slice.created event could be read for <name>. Nothing was changed." and
  the line "Fix or restore the log, then try again." to the error output.
  It exits with code 1.
  Check: e2e

### Seeing all slices

- **R56.** When a user runs `sliceflow slice list`, sliceflow finds the
  slices of the project, prints them in a table to the normal output,
  changes nothing, and exits with code 0.
  Check: e2e
- **R57.** sliceflow takes the candidate names from the local branches
  called `slice/<name>` and from the folders `docs/slices/<name>/` at the
  tip of `main`. It joins the two lists, and a name that is in both is one
  candidate. When `main` does not exist, there are no folders in `main`.
  It never looks at branches on a remote.
  Check: e2e
- **R58.** For each candidate, sliceflow finds the log in the places of
  R51, in the same order. A candidate is a slice when its log is found.
  Check: e2e
- **R59.** The table has the header line `NAME`, `TITLE`, `DESCRIPTION`, then
  one row for each slice, sorted by name, by character code. The name column and the title
  column are padded with spaces to the width of their longest value, the
  header included, with two spaces between columns. Width is counted in
  characters. The description is the last column, and it is not padded. A
  line never ends with a space, so a row with no description ends after
  the title.
  Check: unit
- **R60.** The title and the description of a row come from the
  `slice.created` event.
  Check: unit
- **R61.** The list depends only on the first line of each log.
  Check: unit
- **R62.** If the project has no slices, sliceflow changes nothing and
  prints "No slices yet. To make one, run: sliceflow slice create <name>"
  to the normal output. It exits with code 0.
  Check: e2e
- **R63.** If a candidate has no log, sliceflow does not list it, and prints
  one warning for that name to the error output, and the exit code stays
  0. For a branch only: "Warning: <name> is not listed as a slice. The
  branch slice/<name> has no event log." For a folder in `main` only:
  "Warning: <name> is not listed as a slice. The folder docs/slices/<name>/
  in main has no event log." For both: "Warning: <name> is not listed as a
  slice. Neither the branch slice/<name> nor the folder docs/slices/<name>/
  in main has an event log." All warnings of `list` are sorted by name, by
  character code,
  and they come before the table, or before the line of R62.
  Check: e2e
- **R64.** A candidate whose name has a capital letter or breaks R19, for
  example from a branch
  `slice/Foo` or `slice/a/b`, is not listed. sliceflow prints one warning
  to the error output: "Warning: <name> is not listed as a slice. It is not
  a valid slice name.", where `<name>` is the candidate as it was found.
  The exit code stays 0.
  Check: e2e
- **R65.** If a log is empty, or its first line cannot be read, or its first
  line is not a `slice.created` event, sliceflow does not list the slice,
  and prints one warning to the error output, in place of the warning of
  R53: "Warning: <name> is not listed as a slice. No slice.created event could be read from <path>."
  The exit code stays 0.
  Check: e2e
- **R66.** sliceflow never prints a warning in the list about a damaged line
  of a log other than the first line.
  Check: e2e

### Help, and a command used the wrong way

- **R67.** When a user runs `sliceflow --help`, `sliceflow slice --help`,
  `sliceflow` with no command, or `sliceflow slice` with no command, sliceflow prints the guide shown in
  S41 to the normal output, changes nothing, and exits with code 0.
  Check: e2e
- **R68.** When a user runs a command with `--help`, sliceflow prints to the
  normal output one sentence about the command, its usage line, and its
  options with a short description of each, changes nothing, and exits
  with code 0. Every command has `--help`. The help of `slice create` is
  shown in S44.
  Check: e2e
- **R69.** The help of `slice status` is shown in S45. The help of
  `slice list` is the same shape, with the sentence "Show all slices in a
  table: name, title, and description." and the usage line
  `sliceflow slice list [options]`.
  Check: e2e
- **R70.** If a user runs a command that does not exist, sliceflow changes
  nothing and prints these two lines to the error output: "Error: unknown
  command "<words>". Nothing was changed." and "To see the commands, run:
  sliceflow --help". `<words>` is the first word after `sliceflow` when it
  is not `slice`, or `slice` and the word after it when that word is not a
  command of `slice`. sliceflow does not guess what the user meant.
  Check: e2e
- **R71.** A word that starts with `-` is an option. `-h` means the same as
  `--help`. Options may come before or after the name. `--title=<text>`
  means the same as `--title <text>`, and the same holds for every option
  with a value. The word after an option that needs a value is its value.
  When an option is given more than once, the last one counts.
  Check: unit
- **R72.** If a user gives an option that does not exist, sliceflow changes
  nothing and prints these two lines to the error output: "Error: unknown
  option "<option>". Nothing was changed." and "To see the options, run:
  sliceflow <command> --help", with the command the user ran. When no
  command was given, as in `sliceflow --version`, the second line is "To
  see the options, run: sliceflow --help".
  Check: e2e
- **R73.** If a user runs `slice create` or `slice status` with no name,
  sliceflow changes nothing and prints these two lines to the error
  output: "Error: missing slice name. Nothing was changed." and the usage
  line of that command: "Usage: sliceflow slice create <name> [options]"
  or "Usage: sliceflow slice status <name> [options]". An empty name, such
  as `""`, counts as a missing name.
  Check: e2e
- **R74.** If a user gives an option that needs a value, with no value
  after it, sliceflow changes nothing and prints these two lines to the
  error output: "Error: option "<option>" needs a value. Nothing was
  changed." and "To see the options, run: sliceflow <command> --help",
  with the command the user ran.
  Check: e2e
- **R75.** If a user gives a word that is not a name and not an option,
  sliceflow changes nothing and prints these two lines to the error
  output: "Error: unexpected argument "<word>". Nothing was changed." and
  the usage line of the command, as in R73, or "Usage: sliceflow slice
  list [options]" for `slice list`. `<word>` is the first such word.
  Check: e2e
- **R76.** When a command works, sliceflow exits with code 0. When the
  command line is wrong, it exits with code 2. That is the case for an
  unknown command (R70), an unknown option (R72), a missing name (R73), an
  option with no value (R74), an unexpected argument (R75), a control
  character in a title or a description (R29), and a name that breaks R19
  (R20). When a command fails for any other reason, including a crash, it
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
- Cancelling a slice, and cleaning up after a crash during `slice create` (0015).
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
  Rule: R11

- **SD3.** A slice's worktree is made outside the project, by default
  under `~/.sliceflow/worktrees/<project>/<slice>`. The git setting
  `sliceflow.worktree.dir` chooses another folder (SD19).
  Why: tools in other people's projects, such as ESLint, Jest, and
    `tsc`, can pick up a worktree nested in the project and give wrong
    results. The setting lets a person who prefers the worktree in their
    project choose that.
  Not chosen: a folder inside the project by default, as Claude Code does,
    because it can break those tools in projects we do not control.
  Cost: a person opens the worktree in a second editor window, unless
    they set the folder themselves.
  Rule: R17, R18

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
  Rule: R19

- **SD5.** sliceflow makes the name lowercase, in every command that
  takes a name. `slice create` says so when it changed the name.
  Why: Windows and macOS treat `Card` and `card` as one folder, while git
    treats them as two branches, so two slices differing only by case
    would clash for a teammate on another system.
  Not chosen: keeping the case and refusing a name that clashes in
    case, because it leaves two spellings of one name and needs a more
    complex check.
  Cost: a person or an agent that builds a path from the given name
    gets a wrong path on a case-sensitive system, unless it uses the
    path sliceflow prints.
  Rule: R21

- **SD6.** sliceflow refuses a slice name that is not free. A name is not
  free when `docs/slices/<name>/` exists in the base commit, when the
  branch `slice/<name>` exists, or when the worktree folder exists and
  is not empty. An empty worktree folder is free. For the branch and
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
  Rule: R23, R24, R25, R26

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
  Rule: R10

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
  Rule: R37, R41

- **SD9.** The title of a slice, and its description, live only in the
  event log. The spec file has no title line.
  Why: each fact has one copy (design 1.13), and the log is the only
    source of truth.
  Not chosen: the title in the spec's header as well, because two copies
    can disagree; the title in the spec's header only, because changing it
    would edit Markdown and not add an event.
  Cost: `spec.md` alone does not show its title. A readable view of the
    log, for example as Markdown, can be added later.
  Rule: R3, R28

- **SD10.** A slice created with no `--title` gets its name as the title.
  Why: nothing is ever blank, so `list` and `status` need no special case.
  Not chosen: an empty title, because it leaves a blank column; a
    required `--title`, because every create would need the flag.
  Cost: an untitled slice shows its name twice in the list.
  Rule: R28

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
    a new worktree folder for its next slices; two clones of different
    projects with the same folder name share one worktree folder, so a
    slice name used in both fails on the second `create` with git's
    "already exists" message.
  Rule: R16

- **SD12.** sliceflow adds every event to a log with one function, under a
  file lock, including the first event, `slice.created`.
  Why: two ways to write an event, one with the lock and one without,
    would leave a trap for later code. One rule has no exceptions.
  Not chosen: a first event written without the lock. `create` writes a
    new file that nothing else can write to yet, so the lock is not needed
    there. But a second way of writing an event is a risk, and the lock is
    needed in 0002 anyway.
  Cost: the lock, which must work on Windows too, is built in 0001 and
    not in 0002.
  Rule: R30

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
  Rule: R51

- **SD14.** The names for `list` come from the union of the local branches
  `slice/<name>` and the folders `docs/slices/<name>/` in `main`. A name is a
  slice only when its log is found in the places of SD13.
  Why: a merged slice often has no branch, and a slice being built has no
    folder in `main` yet, so each list alone would miss some. They are the
    same signs that `create` uses to refuse a taken name (SD6), so the two
    commands agree.
  Not chosen: only the branches, because it would hide every finished
    slice; also scanning the worktree folders, because every worktree is on
    a slice branch, so it adds nothing.
  Cost: a slice merged into a branch other than `main` is not listed.
  Rule: R57, R58

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
  Rule: R59

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
  Rule: R61, R66

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
  Rule: R76

- **SD18.** When a command is used the wrong way, sliceflow says which
  word or option it did not understand, and points to the help. It does
  not guess what the user meant.
  Why: with three commands, the help is short enough to read, and a guess
    can be wrong. A wrong-use error has the same shape as sliceflow's other
    errors: first what is wrong, with "Nothing was changed.", then what to
    do.
  Not chosen: a "did you mean" suggestion, because it needs more logic
    that can guess wrong.
  Cost: a typo costs the user one more command, to read the help.
  Rule: R70, R72, R73, R74, R75

- **SD19.** The git setting `sliceflow.worktree.dir` chooses the folder for
  worktrees. It may be a full path, a path that starts with `~`, or a path
  relative to the main project's folder. sliceflow reads it through git,
  so a value for one project wins over a global one.
  Why: it is a personal choice, and design 1.14 puts personal settings in
    git config. A relative path has one clear meaning when it starts from
    the main project's folder, and it lets a person keep worktrees next to
    the project (`../shop-worktrees`) or inside it (`.worktrees`).
  Not chosen: refusing a relative path, because nothing breaks with it;
    a path relative to the folder the person is in, because the same
    command would then put worktrees in different places.
  Cost: a worktree inside the project can confuse tools such as ESLint.
    That is the person's own choice.
  Rule: R18

- **SD20.** sliceflow writes each line of a log with `\n` at its end, reads
  lines that end with `\n` or `\r\n`, closes off a cut-off last line before
  it adds an event, and never deletes or changes a line.
  Why: the log is the only source of truth and is committed in git.
    Without the closing `\n`, a new event would be glued to a cut-off line
    and lost with it. Deleting a damaged line would rewrite history that
    nobody chose to change, and the line is evidence of what went wrong.
  Not chosen: cleaning damaged lines automatically, in this slice or in
    0002, because it changes history; a repair can come later as a command
    a person runs on purpose, for example `doctor`.
  Cost: damaged lines stay in the log, and every reader skips them.
  Rule: R33, R34, R35

- **SD21.** A title or a description is plain text on one line. sliceflow
  refuses any control character, such as a line break, a tab, or an
  escape, and the error shows the text it received, with each control
  character written as a visible code.
  Why: a control character is invisible, and a terminal obeys it when
    `status` or `list` prints it: it can break the layout, change colors,
    or clear the screen. The visible codes let the user see what got in,
    which the terminal would hide.
  Not chosen: storing the text and showing the codes only when it is
    printed, because then the log holds text that breaks other readers;
    replacing a line break with a space, because sliceflow would change
    the user's text without asking.
  Cost: a person who wants a longer description must write it on one
    line, and pasted text with hidden characters is refused.
  Rule: R29

- **SD22.** sliceflow makes its commit with the user's own git settings,
  including git hooks and commit signing.
  Why: git decides what a commit needs on this computer, as in SD8, and a
    team's hooks exist for a reason.
  Not chosen: skipping hooks and signing, because it would go around the
    rules the user or the team set up.
  Cost: a slow hook makes `create` slow, and a signing prompt can wait for
    the user.
  Rule: R6

- **SD23.** When more than one problem applies, sliceflow reports only the
  first, in a fixed order.
  Why: the same mistake always gives the same message, and the order puts
    the cheapest checks first, so nothing is made before a problem is
    found.
  Not chosen: reporting every problem at once, because one fix often
    removes the next problem, and a long list hides the first.
  Cost: a person with two problems sees the second only after fixing the
    first.
  Rule: R40

- **SD24.** A log is usable only when its first line is a readable
  `slice.created` event. `status` and `list` use the same rule.
  Why: the first line holds the title, the description, and the base.
    The two commands must agree on whether a slice can be shown.
  Not chosen: using any readable `slice.created` line in the log, because
    `list` reads only the first line, and the two commands would disagree.
  Cost: a log whose first line is damaged cannot be shown, even when a
    later line holds a copy of the event.
  Rule: R55, R65

- **SD25.** sliceflow adds its own two files to the commit even when a
  `.gitignore` rule matches them, and says so in a `Note:` line.
  Why: without the two files, no slice can be made in that project, and a
    rule such as `*.jsonl` was almost surely not written with sliceflow's
    log in mind. Once the files are in git, git keeps tracking them.
  Not chosen: showing git's error, because the user would have to change
    their `.gitignore` before any slice works.
  Cost: sliceflow goes around a rule the user wrote, for two files. The
    note makes it visible.
  Rule: R5

- **SD26.** Ctrl+C during `slice create` is a failure like any other:
  sliceflow cleans up and exits with code 1.
  Why: "Nothing was changed" must stay true, and a person presses Ctrl+C
    exactly when a step is slow, for example a hook or a signing prompt.
  Not chosen: exit code 130, which many Unix tools use for Ctrl+C,
    because it would be a fourth exit code to remember.
  Cost: a script cannot tell a stop by Ctrl+C from another failure by the
    exit code alone.
  Rule: R39

- **SD27.** When a step works, sliceflow does not show what git printed.
  It shows git's text only when a step fails.
  Why: lines such as "Preparing worktree" add noise to a short success
    screen, and git's text matters only when something went wrong.
  Not chosen: showing all of git's output, because the screen would
    differ between git versions and with every hook.
  Cost: a warning that a hook prints on success is not seen.
  Rule: R8

## Open questions

<!-- Questions for the human, numbered Q1, Q2, ... An answered question
     is written into the spec and removed from here. A deferred question
     that does not block the build moves to Out of scope and the backlog.
     The spec review starts only when this section says None. -->

None.
