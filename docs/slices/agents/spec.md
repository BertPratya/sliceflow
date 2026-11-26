---
slice: agents
created: 2026-11-25
---

<!-- HOW TO FILL THIS TEMPLATE

     A spec says WHAT a feature does, never HOW it is built. No endpoints,
     tables, libraries, or file names. Those go in plan.md. What the
     user sees does belong here: commands, flags, screens, output layout,
     messages, and exit codes. The plan may not decide these.

     The human approves the spec. Write plain English, short sentences,
     one idea per sentence.

     The spec is a list of stories. Each story holds its rules, and each
     rule holds its examples, so a reader finds a rule and its proof in
     one place. A rule that two or more stories need is written once, in
     Shared rules, and each of those stories names it in its Uses: line.

     sliceflow reads this file. Keep every marker exactly as shown: the
     headings (## Shared rules, ## Story 1: ..., ### Rule `name`: ...),
     the bold labels (**Example:**, **SD1.**, **Q1.**), and the labeled
     lines (Uses:, Check:, Why:, Not chosen:, Cost:, Rule:, Replaces:).
     Don't rename them. A labeled line may go on over more
     lines: indent the extra lines 2 spaces more than the label.

     - Replace every {{...}} with real text. The text inside says what
       goes there. sliceflow fills slice and date itself. If they still
       show {{...}}, fill them yourself.
     - The title of the slice lives only in its event log. Don't repeat
       it in a heading or in other files.
     - Each section has a comment like this one with its rules. Follow it.
     - The sections may come in any order. Don't rename them.
     - Repeat a block as often as needed: stories, rules, examples,
       decisions.
     - Only a story is required. A section you don't need can be left
       out, or say one line: None.
     - When you are done, delete every comment and check that no {{ is
       left. A spec with a comment or a {{ left in it is not finished.
       This template puts no placeholder inside a code block, so the check
       finds every one that is left. A {{ or a comment marker inside a
       code block or `inline code` is real content, and is not checked.
     - To change a spec later, read this template again first. The
       finished spec no longer has these rules in it. -->

# Spec

## Summary

<!-- Two or three sentences: what the feature lets someone do, not how,
     and why it is needed. The why helps a worker guess right when a
     rule leaves something open. -->

sliceflow starts a Claude Code agent on a slice with one command, and
the agent works in the background, even after the terminal is closed.
The user sees every agent run of a slice: whether it is running, done,
stuck, or stopped, its model, its time, its tokens, and how close the
user is to their plan's limits; they read what an agent said and did,
and stop it so that nothing it started keeps running. Today agents are
started by hand in separate chats, and nothing records which run
belongs to which slice, what it used, or whether it is still alive;
every later step (the spec review, the run, the dashboard) needs one
reliable way to start, watch, and stop an agent.

## Shared rules

<!-- The rules that two or more stories need, written once. A rule that
     only one story needs goes in that story instead. Write each one as
     a rule is written in a story: see the comment under Story 1.

     Typical shared rules: an error that every command can hit, the order
     in which problems are reported, how paths are printed, the exit
     codes. Each shared rule is named in the Uses: line of at least one
     story; a shared rule that no story uses is not needed, or belongs
     inside a story. If no rule is shared, write: None.

     Example:
     ### Rule `not-in-a-repo`: a folder outside a repository

     If the folder is not inside a git repository, sliceflow changes
     nothing, prints "Error: git could not find the project. Nothing was
     changed.", then git's own message, and exits with code 1.
     Check: e2e

     - **Example:** **Given** I am in a folder that is not inside a git
       repository,
       **When** I run `sliceflow slice list`,
       **Then** sliceflow prints:
       ```
       Error: git could not find the project. Nothing was changed.
       git said:
         fatal: not a git repository (or any of the parent directories): .git
       ``` -->

### Rule `{{rule-name}}`: {{one line: what the rule is about}}

{{The full rule.}}
Check: {{unit, api, e2e, or manual}}

- **Example:** **When** {{what the user does}},
  **Then** {{what the user sees}}.

## Story 1: {{short title}}

<!-- One thing the user wants to do. Every rule sits under a story or
     in Shared rules.

     The story:
     - A short title in the heading, then one sentence saying who wants
       to do what.
     - A Uses: line naming the shared rules this story needs, e.g.
       Uses: not-in-a-repo, exit-codes. Write Uses: none when it needs
       none.
     - Then its rules. A story has at least one rule of its own, and
       its first rule is the main success case: what happens when
       everything goes right. Then the other cases, then each important
       failure. If every rule of a story seems to be a shared one, the
       story is likely part of another story.

     A rule:
     - A ### heading: Rule, its name in backticks, a colon, and one line
       saying what it is about. The name is lowercase words joined by
       hyphens, e.g. `lowercase-name`, and says what the rule is about,
       so a reader knows it without looking it up. Each name is used
       once in the whole spec.
     - The full rule, in one of these forms (EARS, without "shall"), in
       the present tense:
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
     - Name exact values: the message printed, the limit, the exit code.
       A worker sees this text, so "say so" is not enough.
     - One behavior per rule. A term that needs one fixed meaning is
       defined in the rule where it first appears, e.g. "A task is
       overdue when its due date is before today."
     - A Check: line saying how it is tested. More than one is allowed,
       e.g. Check: unit, e2e.
       - unit: tests one piece of code on its own.
       - api: calls the program's interface directly, e.g. an HTTP
         endpoint. A plain command-line tool usually has none: use e2e.
       - e2e: runs the real program the way a user does.
       - manual: a human checks it at the preview. Use it only when no
         automatic test can check it.
     - Its examples, as a list under the Check: line. A rule the user
       can see, including an error message, has at least one example. A
       rule with several cases has one example per case. A rule the user
       cannot see, e.g. a file format, may have none.

     An example:
     - Starts with **Example:**, then has three parts:
       - Given: the starting state. Leave it out when there is nothing
         special.
       - When: what the user does. Only one per example. If you need
         two, split it into two examples.
       - Then: what the user sees. Show the real screen, message, or
         terminal output, with real values. Use a code block when it is
         longer than one line.
     - Use "and" for more than one condition or result.
     - An example shows the rule it sits under. When the output also
       shows other rules, e.g. the whole success output, it still sits
       under one rule only: the one it is meant to prove.

     A rule is the full rule; an example is one case of it, with real
     values. Both are needed.

     More than about 25-30 rules in the whole spec, shared ones
     included? The slice is too big. Split it.

     When the spec changes after it was approved:
     - Never rename a rule, and never reuse a name. A changed rule keeps
       its name. A new rule gets a new name.
     - A removed rule keeps its heading, with the reason in place of its
       text: ### Rule `name`: (removed in R002: why)

     Example:
     ## Story 1: Start a new slice

     A developer starts work on a new feature with one command.

     Uses: not-in-a-repo

     ### Rule `slice-created`: a new slice and where to write its spec

     When a user runs `sliceflow slice create <name>`, sliceflow makes
     the slice, prints "Created slice <name>.", then the line
     "Next: write the spec in <path>", and exits with code 0.
     Check: e2e

     - **Example:** **When** I run `sliceflow slice create checkout`,
       **Then** sliceflow prints:
       ```
       Created slice checkout.
       Next: write the spec in /home/you/.sliceflow/worktrees/shop/checkout/docs/slices/checkout/spec.md
       ```

     ### Rule `name-taken`: a slice that exists is not made again

     If the branch `slice/<name>` exists, sliceflow changes nothing,
     prints "Error: slice <name> already exists (branch slice/<name>).
     Nothing was changed.", and exits with code 1.
     Check: e2e

     - **Example:** **Given** a slice named `checkout` already exists,
       **When** I run `sliceflow slice create checkout`,
       **Then** sliceflow changes nothing and prints
       "Error: slice checkout already exists (branch slice/checkout).
       Nothing was changed." -->

{{One sentence: who wants to do what.}}

Uses: {{the shared rules this story needs, or none}}

### Rule `{{rule-name}}`: {{one line: what the rule is about}}

{{When <trigger>, <system> <response>.}}
Check: {{unit, api, e2e, or manual}}

- **Example:** **When** {{what the user does}},
  **Then** {{what the user sees}}:
  {{the exact output, in a code block}}
- **Example:** **Given** {{the starting state}},
  **When** {{what the user does}},
  **Then** {{what the user sees}}.

## Out of scope

<!-- What this slice does not do, even if it looks related, so nobody
     builds it by mistake. One - bullet per item, one line each. No
     sub-bullets, headings, or code blocks.

     Example:
     - Reminders and notifications.
     - Repeating tasks. -->

- A review of the spec by an agent, and answers in JSON: slice 0003a.
- A page in the browser that shows a running agent live: slice 0003b.
- Sending a message to an agent while it runs: the chat page.
- Codex, or any agent tool other than Claude Code: slice 0008.
- Settings per role, and a settings file: slice 0005.
- A worktree for an agent; an agent that may edit works in the user's checkout.
- Binding a chat or a lead to a slice, and listing a slice's chats: slice 0009a.
- Bringing back runs after sliceflow itself was killed, and retrying a run: slice 0010.
- Limits on tokens or cost per run or per slice: slice 0013.
- A copy of an agent's conversation saved by sliceflow.
- Bringing back a lost run by itself; the user continues its conversation instead: slice 0010.
- Agent commands on Windows, beyond an error that says they do not work there yet: slice 0013a.

## Decisions

<!-- Product decisions made while writing the spec.

     - Number them SD1, SD2, ...
     - Each one has the choice, Why:, and Not chosen: (the main other
       option, and why not).
     - Each one has a Cost: line: the downside we accept with this
       choice. With truly no downside, write Cost: None. A reviewer
       questions a None that looks wrong.
     - Each one has a Rule: line. A worker sees rules, not decisions. So
       a decision that changes what the feature does is also written as
       a rule, and Rule: names it, e.g. Rule: lowercase-name. A decision
       that changes no behavior says Rule: none. sliceflow checks that
       every decision has this line, and that each name in it is a rule
       of this spec.
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
       Rule: overdue-task -->

- **SD1.** A record of each agent run stays on the user's computer, inside
  git's own folder (`.git/sliceflow/`), and is never committed.
  Why: a run is like a process: it lives on one machine. Committing
  would add two bookkeeping commits per run to the base branch. Inside
  `.git/`, nothing shows in `git status`, and the user's `.gitignore`
  is never touched, as with the lock (plan-on-base SD10).
  Not chosen: committing each event, as `slice edit` does, because the
  history would fill with noise; a `.sliceflow/` folder in the project,
  because git would show it as untracked.
  Cost: teammates who pull the repo do not see the user's runs, and a
  fresh clone starts with none.
  Rule: {{named when this story's rules are drafted}}
- **SD2.** sliceflow saves no copy of an agent's conversation. It reads
  the agent tool's own session files.
  Why: one source of truth; a copy would only repeat what the tool keeps.
  Not chosen: a copy in sliceflow's own format, because it duplicates
  every conversation.
  Cost: when the tool deletes an old session (Claude Code does after 30
  days by default), its conversation can no longer be shown.
  Rule: {{named when this story's rules are drafted}}
- **SD3.** Agent commands work on Linux and macOS first. On Windows they
  print an error and change nothing, until slice 0013a.
  Why: the human develops on Linux, and Linux and macOS share one way to
  start and stop processes; Windows needs different work.
  Not chosen: all three systems now, because it doubles the hardest part
  of the slice.
  Cost: Windows users cannot use agents until 0013a.
  Rule: {{named when this story's rules are drafted}}
- **SD4.** If the part of sliceflow that watches a run stops, the agent
  stops too, and the run shows as lost.
  Why: an agent must never run with nobody enforcing its time limit or
  recording how it ended.
  Not chosen: the agent keeps running alone, because it could run all
  night, unwatched, using the user's plan limits.
  Cost: the work in progress stops; the user continues the conversation
  with `sliceflow agent continue`.
  Rule: {{named when this story's rules are drafted}}
- **SD5.** Continuing a conversation starts a new run on the same
  session. Each run is one start and one end of the agent.
  Why: Claude Code continues a session in a new process, with the same
  session, so one conversation can span several runs.
  Not chosen: reopening the old run, because a run that ended would
  start again and its times and usage would mix.
  Cost: one conversation shows as several lines in `agent list`.
  Rule: {{named when this story's rules are drafted}}
- **SD6.** A run the user starts has full capability by default: the
  user's own Claude Code setup, as when they run `claude` themselves.
  `--allow read` or `--allow edit` narrows it. Runs sliceflow starts for
  a role, e.g. the reviewer, get a fixed list that the user cannot widen.
  Why: who drives the agent decides its permissions; the user wants the
  same power as the official app.
  Not chosen: read-only by default, because it caps what the user can do
  with their own agent.
  Cost: a run with full capability works in the user's own checkout,
  limited only by their own settings.
  Rule: {{named when this story's rules are drafted}}

## Open questions

<!-- Questions for the human, numbered Q1, Q2, ... An answered question
     is written into the spec and removed from here. A deferred question
     that does not block the build moves to Out of scope and the backlog.
     The spec review starts only when this section says None. -->

- **Q1.** How is a run named in `agent show`, `agent stop`, and
  `agent continue`: a number per slice, or a short ID unique in the
  project? The human decides later, after more thought.
