# Change: plan on the base branch

Changes: slice-and-events (0001)

## Why

When `slice create` puts everything in a new worktree, it is hard to keep
track of slices: their files are spread over folders, and people delete
a worktree once a slice is done. Planning on the base branch keeps every
slice's documents in one place that is easy to check. A worktree is only
needed to write code.

## What changes

- `slice create` writes the slice's spec and log in the user's checkout,
  on the base branch, and commits only the slice's folder. It makes no
  branch and no worktree.
- The base comes from `--from`, then the git setting `sliceflow.base`,
  then `main`.
- `slice create` refuses when the user's checkout is not on the base
  branch.
- `slice status` and `slice list` read the log only from the user's
  checkout, and warn when it may be out of date.

## Changed rules

### Rule `create-in-your-folder`: a new slice is written in the user's checkout, on the base branch

Replaces: slice-and-events/R1, slice-and-events/R2,
slice-and-events/R11 (in part), slice-and-events/R16 (in part),
slice-and-events/R17 and R18 (moved to 0005)

The user's checkout is the top folder of the checkout the user runs the
command in, as git gives it (`git rev-parse --show-toplevel`): from any
subfolder, and also in a second worktree the user made. When a user runs `sliceflow slice create <name>` in a git repository
that has commits, with their folder on the base branch, and the name is
free, sliceflow writes `docs/slices/<name>/spec.md` and
`docs/slices/<name>/.events.jsonl` in the user's checkout, and makes one
commit of only those two files, on the base branch. It makes no branch
and no worktree. The command takes the options `--title`,
`--description`, and `--from`.
Check: e2e

- **Example:** **Given** my folder is on `main`, and there is no slice
  `login`,
  **When** I run `sliceflow slice create login`,
  **Then** my folder has `docs/slices/login/spec.md` and
  `docs/slices/login/.events.jsonl`, `main` has one new commit,
  "login: add the spec and the event log", that holds only those two
  files, and there is no branch `slice/login` and no worktree for it.

### Rule `create-touches-only-its-folder`: nothing else in the user's checkout changes

Replaces: slice-and-events/R7

sliceflow adds only the slice's folder and its one commit. It never
changes the user's other files, what is staged, or which branch is
checked out.
Check: e2e

- **Example:** **Given** my folder is on `main`, with a staged change to
  `app.go` and a new file `notes.txt` that is not staged,
  **When** I run `sliceflow slice create login`,
  **Then** the new commit does not hold `app.go` or `notes.txt`,
  `app.go` is still staged, `notes.txt` is still not staged, and `main`
  is still checked out.

### Rule `log-lock-in-git-folder`: the lock of a log is kept inside git's folder

sliceflow takes the lock of a slice's log in the file
`sliceflow/locks/<name>.lock` under the folder that
`git rev-parse --git-path` gives for the user's checkout, inside `.git`,
and never next to the log. So `git status` never shows it, and it can
never be committed. sliceflow makes the folder `sliceflow/locks/` when
it is missing. The file stays after the lock is released. In a second
worktree, git gives that worktree's own folder, so each checkout has
its own lock, as it has its own copy of the log. When the folder cannot
be made or the lock cannot be taken, slice-and-events/R38 applies, with
the lock's full path as `<path>`.
Check: e2e

- **Example:** **Given** my checkout is on `main`, and `git status`
  shows nothing,
  **When** I run `sliceflow slice create login`,
  **Then** `git status` still shows nothing, and there is no file
  `docs/slices/login/.events.jsonl.lock`.

### Rule `create-output`: what a new slice prints

Replaces: slice-and-events/R8, slice-and-events/R5 and R22 (in part:
the `Note:` padding)

When the slice is created, sliceflow prints these lines to the normal
output, and exits with code 0: `Created slice <name> on branch
<base>.`, then `Next: write the spec in <path>`, the full path of
`spec.md` in the user's checkout. The `Note:` lines of
slice-and-events/R22 and R5 come right after the first line, in that
order, when they apply, and start with `  Note: `, with one space after
the colon (the old padding lined up with labels that are gone). sliceflow prints nothing else: what git prints
when a step works is not shown.
Check: e2e

- **Example:** **Given** my folder `/home/you/projects/shop` is on
  `main`,
  **When** I run `sliceflow slice create login`,
  **Then** sliceflow prints:
  ```
  Created slice login on branch main.
  Next: write the spec in /home/you/projects/shop/docs/slices/login/spec.md
  ```
  and exits with code 0.

- **Example:** **Given** my checkout is on `main`,
  **When** I run `sliceflow slice create Login`,
  **Then** sliceflow prints:
  ```
  Created slice login on branch main.
    Note: "Login" was made lowercase.
  Next: write the spec in /home/you/projects/shop/docs/slices/login/spec.md
  ```

### Rule `base-branch`: where the base comes from

Replaces: slice-and-events/R10

The base branch of a new slice is the local branch given with `--from`.
When `--from` is left out or empty, it is the git setting
`sliceflow.base`. When that is not set or empty, it is `main`. A tag, a
commit, or a remote branch is not a local branch, so `base-missing`
applies to it.
Check: e2e

- **Example:** **Given** the git setting `sliceflow.base` is `dev/bert`,
  and my folder is on `dev/bert`,
  **When** I run `sliceflow slice create login`,
  **Then** sliceflow prints `Created slice login on branch dev/bert.`
  first, and the commit is on `dev/bert`.
- **Example:** **Given** the git setting `sliceflow.base` is `dev/bert`,
  and my folder is on `release`,
  **When** I run `sliceflow slice create login --from release`,
  **Then** sliceflow prints `Created slice login on branch release.`
  first.

### Rule `base-missing`: the base branch does not exist

Replaces: slice-and-events/R12

If the base branch does not exist, sliceflow changes nothing and prints
to the error output the line "Error: base branch <base> does not exist.
Nothing was changed.", then lines that depend on where the base came
from:
- from `--from`: "Check the name. To see your branches, run:" and the
  command `git branch`;
- from `sliceflow.base`: "The base branch comes from the git setting
  sliceflow.base. To see it, run:" and the command
  `git config sliceflow.base`;
- from neither, so `main`: "sliceflow starts from main unless you choose
  another branch. To choose one for this slice, run:", the command
  `sliceflow slice create <name> --from <branch>` with the slice name
  filled in, "To choose one for every slice, run:", and the command
  `git config sliceflow.base <branch>`.

Each command is on its own line, indented by two spaces. It exits with
code 1.
Check: e2e

- **Example:** **Given** the git setting `sliceflow.base` is `dev/bert`,
  and there is no branch `dev/bert`,
  **When** I run `sliceflow slice create login`,
  **Then** sliceflow prints to the error output:
  ```
  Error: base branch dev/bert does not exist. Nothing was changed.
  The base branch comes from the git setting sliceflow.base. To see it, run:
    git config sliceflow.base
  ```
  and exits with code 1.
- **Example:** **Given** no `sliceflow.base` setting, and no branch
  `main`,
  **When** I run `sliceflow slice create login`,
  **Then** sliceflow prints to the error output:
  ```
  Error: base branch main does not exist. Nothing was changed.
  sliceflow starts from main unless you choose another branch. To choose one for this slice, run:
    sliceflow slice create login --from <branch>
  To choose one for every slice, run:
    git config sliceflow.base <branch>
  ```
  and exits with code 1.

### Rule `folder-taken`: the slice's folder already exists

Replaces: slice-and-events/R23, slice-and-events/R25

If `docs/slices/<name>/` exists and is not empty in the user's
checkout, or exists in the base commit, sliceflow changes nothing,
prints "Error: slice <name> already exists (folder docs/slices/<name>/).
Nothing was changed." to the error output, and exits with code 1. It
checks both, so it never writes over a folder that is not committed,
nor over a slice whose folder was deleted but whose deletion is not
committed. An empty folder in the checkout is not taken: sliceflow
writes into it. `create` makes no worktree, so a worktree folder is not
checked; the check of the branch `slice/<name>` (slice-and-events/R24)
stays.
Check: e2e

- **Example:** **Given** my folder has a folder `docs/slices/login/`
  that was never committed,
  **When** I run `sliceflow slice create login`,
  **Then** sliceflow prints to the error output:
  ```
  Error: slice login already exists (folder docs/slices/login/). Nothing was changed.
  ```
  leaves the folder as it was, and exits with code 1.
- **Example:** **Given** the slice `login` is committed on `main`, and I
  deleted `docs/slices/login/` in my folder without committing that,
  **When** I run `sliceflow slice create login`,
  **Then** sliceflow prints the same error, and exits with code 1.
- **Example:** **Given** my checkout has an empty folder
  `docs/slices/login/`,
  **When** I run `sliceflow slice create login`,
  **Then** sliceflow creates the slice in that folder.

### Rule `create-undo`: a failed create leaves nothing behind

Replaces: slice-and-events/R36, slice-and-events/R38 and R39 (in
part: they point to the undo)

If a step fails after sliceflow made something, or the user stops
`slice create` with Ctrl+C, before the commit is made, sliceflow
removes what it made in this run, and only that: the files it wrote in
`docs/slices/<name>/`, the folder when it made it (an empty folder that
was there before stays), and those files in git's staging area, which
it puts back as they were. Folders it made above it, e.g.
`docs/slices/`, stay.

Once the commit is made, the slice exists, and sliceflow removes
nothing, even when the user stops it, e.g. while a post-commit hook
runs. It knows the commit is made when `docs/slices/<name>/.events.jsonl`
is in the commit at the tip of the user's branch (`folder-taken` made
sure it was not there before), so a commit by another program in the
meantime does not fool it. It then prints to the error output "Stopped after the commit. Slice <name> was
created on branch <base>." and exits with code 1. It never resets or
rewrites the user's branch.
Check: e2e

- **Example:** **Given** a git hook that refuses every commit,
  **When** I run `sliceflow slice create login`,
  **Then** sliceflow reports the failed commit, my checkout has no
  `docs/slices/login/`, `git status` shows nothing for it, and
  sliceflow exits with code 1.
- **Example:** **Given** a post-commit hook that waits,
  **When** I run `sliceflow slice create login` and press Ctrl+C while
  the hook waits,
  **Then** sliceflow prints to the error output
  `Stopped after the commit. Slice login was created on branch main.`,
  the commit stays on `main`, `docs/slices/login/` stays, and it exits
  with code 1.

### Rule `create-git-step-fails`: git cannot make the commit

Replaces: slice-and-events/R37

sliceflow takes one step with git: it makes the commit, which includes
adding the two files. When it fails and sliceflow has no message of its
own for the failure, sliceflow prints this to the error output: a first
line "Error: git could not make the commit. The slice was not created.
Nothing was changed.", then "git said:", then everything git printed to
its error output, word for word, each line indented by two spaces. An
empty line in git's text stays empty, with no spaces. When git printed
nothing, the line "git printed no message." takes the place of "git
said:" and the text. sliceflow does not check beforehand for what git
needs.
Check: e2e

- **Example:** **Given** a git hook that refuses every commit and
  prints `commit refused`,
  **When** I run `sliceflow slice create login`,
  **Then** sliceflow prints to the error output:
  ```
  Error: git could not make the commit. The slice was not created. Nothing was changed.
  git said:
    commit refused
  ```
  and exits with code 1.
- **Example:** **Given** my checkout is in the middle of a merge,
  **When** I run `sliceflow slice create login`,
  **Then** git refuses the commit (it cannot commit only some files
  during a merge), and sliceflow prints its message in the same form,
  removes what it made, and exits with code 1.

### Rule `create-leftovers`: what could not be removed

Replaces: slice-and-events/R41

If sliceflow could not remove everything it made, its first line does
not say "Nothing was changed." After git's text, or the system's text,
it prints the line "Some of what sliceflow made could not be removed:",
then one line for each kind that is left: `  folder    <full path>`,
the full path of `docs/slices/<name>/`, when the folder is left, and
`  staged    docs/slices/<name>/`, when its files are still in git's
staging area. Then it prints "To clean up:", and for each kind that is
left, in the same order, `  delete the folder <full path>` and
`  run: git -C "<top folder>" reset -- "docs/slices/<name>"`, where
`<top folder>` is the full path of the user's checkout, so the command
works from any folder. An empty
folder that could not be removed is not listed
(slice-and-events/SD31), and `folder-taken` does not count it as
taken.
Check: e2e

- **Example:** **Given** a git hook that refuses every commit and
  prints `commit refused`, and a file in `docs/slices/login/` that the
  system will not let sliceflow delete,
  **When** I run `sliceflow slice create login`,
  **Then** sliceflow prints to the error output:
  ```
  Error: git could not make the commit. The slice was not created.
  git said:
    commit refused
  Some of what sliceflow made could not be removed:
    folder    /home/you/projects/shop/docs/slices/login
    staged    docs/slices/login/
  To clean up:
    delete the folder /home/you/projects/shop/docs/slices/login
    run: git -C "/home/you/projects/shop" reset -- "docs/slices/login"
  ```
  and exits with code 1.

### Rule `list-finds-slices`: the slices are the folders in the user's checkout

Replaces: slice-and-events/R57, slice-and-events/R58,
slice-and-events/R63, slice-and-events/R64 (in part: its example names
branches)

`slice list` takes the candidate names from the folders
`docs/slices/<name>/` in the user's checkout. A candidate is a slice
when its folder has the log `.events.jsonl`. It never looks at
branches, `main`, or a remote. For a folder with no log, it prints to
the error output "Warning: <name> is not listed as a slice. The folder
docs/slices/<name>/ has no event log.", and the exit code stays 0.
All warnings of `list` about a single name (this one, and those of
slice-and-events/R64 and R65) are sorted by name, by character code.
They come after the warning of `list-old-copies`, and before the table,
or before the line of slice-and-events/R62 when there are no slices.
Check: e2e

- **Example:** **Given** my checkout has `docs/slices/login/` with a
  log, and `docs/slices/notes/` with no log,
  **When** I run `sliceflow slice list`,
  **Then** sliceflow prints to the error output:
  ```
  Warning: notes is not listed as a slice. The folder docs/slices/notes/ has no event log.
  ```
  then the table, with one row, for `login`, and exits with code 0.

### Rule `list-old-copies`: one warning for slices planned on other branches

When the user's checkout does not hold every commit of the base branch
of one or more slices in the list (as in `old-copy-warning`),
`slice list` prints to the error output, before the table, one line:
"Warning: your checkout is on branch <branch>, but <n> slices are
planned on other branches, so they may be out of date." For one slice,
the line ends "but 1 slice is planned on another branch, so it may be
out of date." When the checkout is on no branch, the line starts
"Warning: your checkout is not on a branch, but". The warning does not
change the exit code.
Check: e2e

- **Example:** **Given** my checkout is on `fix/new`, and has the slices
  `login` and `report`, both planned on `dev/bert`, and `dev/bert` got
  one more commit after `fix/new` was made from it,
  **When** I run `sliceflow slice list`,
  **Then** sliceflow prints to the error output:
  ```
  Warning: your checkout is on branch fix/new, but 2 slices are planned on other branches, so they may be out of date.
  ```
  then the table, and exits with code 0.

### Rule `create-on-the-base`: the user's checkout must be on the base branch

If the user's checkout is not on the base branch, `slice create` changes
nothing, prints to the error output "Error: the base branch is <base>,
but your checkout is on <branch>. Nothing was changed.", then "Switch to
<base>, then try again.", and exits with code 1. When the folder is on
no branch (a detached HEAD, e.g. during a rebase), the first line ends
"but your checkout is not on a branch. Nothing was changed." instead.
Check: e2e

- **Example:** **Given** the git setting `sliceflow.base` is `dev/bert`,
  and my folder is on `fix/typo`,
  **When** I run `sliceflow slice create login`,
  **Then** sliceflow prints to the error output:
  ```
  Error: the base branch is dev/bert, but your checkout is on fix/typo. Nothing was changed.
  Switch to dev/bert, then try again.
  ```
  makes no commit and no file, and exits with code 1.
- **Example:** **Given** my folder is on no branch, after
  `git checkout <commit>`,
  **When** I run `sliceflow slice create login`,
  **Then** sliceflow prints to the error output:
  ```
  Error: the base branch is main, but your checkout is not on a branch. Nothing was changed.
  Switch to main, then try again.
  ```
  and exits with code 1.

### Rule `create-error-order`: which error comes first

Replaces: slice-and-events/R40

When more than one problem applies, sliceflow checks them one by one,
in this order, and reports only the first that fails: the command line
(slice-and-events/R70 to R75), where `--help` comes before any error,
the first wrong word from the left is reported, and a missing name is
checked last; a control character; the name; git not found; the
repository; the base (`base-missing`); the user's checkout on the base
(`create-on-the-base`); a name that is not free, the folder
(`folder-taken`) before the branch (slice-and-events/R24); then the
steps that make the slice. `slice status` and `slice list` use the same
order, for the checks they have.
Check: e2e

- **Example:** **Given** my folder is on `fix/typo`, and `main` already
  has `docs/slices/login/`,
  **When** I run `sliceflow slice create login`,
  **Then** sliceflow reports only the error of `create-on-the-base`,
  because it comes before `folder-taken`.

### Rule `status-reads-your-checkout`: the log is read from the user's checkout only

Replaces: slice-and-events/R51

`slice status` and `slice list` read a slice's log only from the user's
checkout (see `create-in-your-folder`): the file
`docs/slices/<name>/.events.jsonl` under its top folder. They never
read another branch, another worktree, or a commit.
Check: e2e

- **Example:** **Given** the slice `login` on `main`, and I am in the
  subfolder `internal/cli` of my checkout, which is on `main`,
  **When** I run `sliceflow slice status login`,
  **Then** sliceflow prints the status of `login`, read from
  `docs/slices/login/.events.jsonl` in my checkout, and exits with
  code 0.
- **Example:** **Given** my main checkout is on `main`, and I made a
  second worktree of my own on `dev/bert`, which has the slice `login`,
  **When** I run `sliceflow slice status login` inside that second
  worktree,
  **Then** sliceflow reads the log from that worktree, not from the
  main checkout.

### Rule `no-slice`: no log in the user's checkout

Replaces: slice-and-events/R52

If the user's checkout has no log for the slice, sliceflow changes
nothing, prints to the error output "Error: no slice named <name>.
Nothing was changed.", then "No log was found at
docs/slices/<name>/.events.jsonl.", then "Your checkout is on branch
<branch>. Check the name, or switch to the branch the slice is planned
on.", and exits with code 1. When the checkout is on no branch, the
last line starts "Your checkout is not on a branch." instead.
Check: e2e

- **Example:** **Given** the slice `login` is planned on `dev/bert`, and
  my checkout is on `fix/old`, a branch made before `login`,
  **When** I run `sliceflow slice status login`,
  **Then** sliceflow prints to the error output:
  ```
  Error: no slice named login. Nothing was changed.
  No log was found at docs/slices/login/.events.jsonl.
  Your checkout is on branch fix/old. Check the name, or switch to the branch the slice is planned on.
  ```
  and exits with code 1.

### Rule `old-copy-warning`: a log from another branch may be out of date

When the user's checkout does not hold every commit of the base branch
that the log's `slice.created` event names (git: `git merge-base
--is-ancestor <base> HEAD` fails), `slice status` prints to the error
output "Warning: your checkout is on branch <branch>, but slice <name>
is planned on branch <base>, so this may be out of date.", then the
status as usual. It comes before the warning about lines that cannot be
read (slice-and-events/R53). When the checkout is on no branch, the
warning starts "Warning: your checkout is not on a branch, but". When
the base branch does not exist, there is no warning. When git cannot
answer (an error, not "no"), sliceflow prints instead "Warning:
sliceflow could not check whether slice <name> is out of date.", then
"git said:" and git's text, indented by two spaces, and goes on; `list`
does the same once per base branch. The warning does not change the
exit code.
Check: e2e

- **Example:** **Given** the slice `login` is planned on `dev/bert`, and
  my checkout is on `fix/new`, a branch made from `dev/bert` after
  `login` was created, and `dev/bert` got one more commit after that,
  **When** I run `sliceflow slice status login`,
  **Then** sliceflow prints to the error output:
  ```
  Warning: your checkout is on branch fix/new, but slice login is planned on branch dev/bert, so this may be out of date.
  ```
  then prints the status, and exits with code 0.
- **Example:** **Given** the slice `login` was planned on `dev/bert`,
  and `dev/bert` was merged into `main`, and my checkout is on `main`,
  **When** I run `sliceflow slice status login`,
  **Then** sliceflow prints no warning, because `main` holds every
  commit of `dev/bert`.

### Rule `status-lines`: the lines of the status

Replaces: slice-and-events/R43, slice-and-events/R46,
slice-and-events/R47, slice-and-events/R48, slice-and-events/R50,
slice-and-events/R27 (in part)

The first line of the status is `<slice name>: <title>`. The next lines
have the labels `Description:`, `Created:`, `Base:`, and `Events:`, in
this order, with the values in one column. The title, the description,
and the base branch come from the `slice.created` event. `Base:` shows
only the base branch. The event still records the base commit: the tip
of the base branch just before the slice's own commit. The status does
not show it. The status has no `Branch:`, `Worktree:`, or `Read from:`
line: a slice has no branch or worktree while it is planned, and the
log is always read from the user's checkout.
Check: e2e

- **Example:** **Given** I created the slice `login` on `main`, with the
  title "Log in" and no description,
  **When** I run `sliceflow slice status login`,
  **Then** sliceflow prints:
  ```
  login: Log in
    Description:  (none)
    Created:      2026-11-19 09:15 UTC
    Base:         main
    Events:       1 (last: slice.created)
  ```
  and exits with code 0.

### Rule `help-text`: the help says what the commands do now

Replaces: slice-and-events/R68 and R69 (in part: the help of `create`
and `status` shown in S44 and S45)

`sliceflow slice create --help` starts with "Make a new slice on the
base branch: a spec and an event log, in one commit.", and describes
`--from` as "The base branch. Default: the sliceflow.base setting, else
main.". `sliceflow slice status --help` starts with "Show one slice: its
title, description, creation time, base, and events.". The help of
`slice list` does not change.
Check: e2e

- **Example:** **When** I run `sliceflow slice create --help`,
  **Then** its first line is
  `Make a new slice on the base branch: a spec and an event log, in one commit.`,
  and the line for `--from` ends with
  `The base branch. Default: the sliceflow.base setting, else main.`

## Out of scope

- Making the slice's branch and worktree when implementation starts:
  0005 (`sliceflow run`).
- `slice status` reading the worktree once implementation has started:
  0005.
- The wrong-branch check for other commands that change a slice: in
  each command's own slice, e.g. `slice edit` in 0002.
- Switching branches while a long command runs: 0003.
- Two people changing one slice on a shared base branch: later
  (backlog).
- The template and checks for `change.md` files (the light path):
  later.
- Where `sliceflow run` starts the slice's branch: from the base
  commit, or from the base's tip at that time: 0005.
- A teammate's slice with the same name, on the remote: sliceflow never
  fetches, so it is not seen, and the two folders conflict when pulled.
  Handled with the backlog row on two people and one slice.
- One lock for log commits and merges across worktrees (design 1.13):
  0005.
- Where a worktree is made (slice-and-events/R17, R18, SD3, SD19):
  moved to 0005; `create` no longer reads `sliceflow.worktree.dir`.

## Decisions

- **SD1.** Planning happens on the slice's base branch, in the user's
  own folder. The branch and worktree are made only when implementation
  starts.
  Why: with every slice in its own worktree, slices are hard to track,
  and a deleted worktree hides a finished slice. On the base branch,
  every slice's documents are in one place.
  Not chosen: keeping the worktree from the start (slice-and-events);
  a separate planning branch per slice.
  Cost: the user's checkout must be on the base branch while planning.
  Rule: create-in-your-folder, create-output, status-lines, help-text
  Replaces: slice-and-events/SD2 (in part), slice-and-events/SD3 and
  SD19 (in part: where the worktree goes moves to 0005)

- **SD2.** The base is `--from`, else the git setting `sliceflow.base`,
  else `main`.
  Why: tools that pick a base branch use a flag first, then a setting,
  then a default (GitHub CLI, spec-kitty, Git Town). The setting covers
  what a person usually does, e.g. `dev/<name>`; the flag covers one
  slice that goes elsewhere.
  Not chosen: dropping `--from`, because a one-off slice would need the
  setting changed and changed back.
  Cost: three places to look when the base is not what was expected.
  Rule: base-branch, base-missing
  Replaces: slice-and-events/SD7 (no setting for the base in that
  slice), slice-and-events/SD32 (an empty `--from` now falls to the
  setting)

- **SD3.** A command that changes a slice refuses, and changes nothing,
  when the user's checkout is not on the slice's base branch. Commands
  that only read are not checked.
  Why: otherwise sliceflow would commit the slice's files onto the wrong
  branch. Spec Kit, Lerna, and the pre-commit framework also refuse in
  this case rather than switch branches.
  Not chosen: switching the branch for the user, because it can fail or
  surprise them when they have uncommitted changes.
  Cost: the user must switch by hand.
  Rule: create-on-the-base, create-error-order

- **SD4.** sliceflow commits only the slice's own folder,
  `docs/slices/<name>/`.
  Why: the commit happens in the user's checkout, next to their own
  uncommitted work, which must never be swept in.
  Not chosen: committing everything that is staged.
  Cost: a git hook runs as the user set it up (slice-and-events/R6),
  and may change other files, e.g. a hook tool that stashes the user's
  changes and puts them back; sliceflow cannot prevent that.
  Rule: create-in-your-folder, create-touches-only-its-folder

- **SD5.** `slice status` and `slice list` read logs only from the
  user's checkout. They never search other branches, worktrees, or
  `main`.
  Why: spec-driven tools read specs from the folder you are in (Spec
  Kit, OpenSpec, Kiro, spec-kitty), so what sliceflow shows matches the
  files in front of the user. While planning, the guard keeps the user
  on the base branch anyway.
  Not chosen: falling back to the log committed on the base branch,
  because a log's base is written inside the log, so it cannot be found
  without the log; keeping logs in git refs outside the files, as
  git-bug does, because a normal clone does not copy them.
  Cost: on another branch, a slice is not found, or an old copy is
  shown, with a warning.
  Rule: status-reads-your-checkout, no-slice, list-finds-slices
  Replaces: slice-and-events/SD13, slice-and-events/SD14

- **SD6.** When the checkout does not hold every commit of a log's base
  branch, `status` warns once for the slice, and `list` warns once for
  all such slices.
  Why: a branch made after a slice holds an old copy of its log, which
  may be out of date. Comparing commits, not branch names, keeps the
  warning away after the base is merged, e.g. on `main` after
  `dev/bert` was merged into it.
  Not chosen: comparing branch names, because every finished slice
  would warn forever after a merge; a warning on every row of `list`,
  because it is noisy; no warning, because an old copy would show
  silently.
  Cost: one more git call per base branch.
  Rule: old-copy-warning, list-old-copies

- **SD7.** The slice's folder is taken when it is not empty in the
  checkout, or exists in the base commit. An empty folder is free.
  Why: the checkout covers a folder that was never committed; the base
  commit covers a slice whose deletion is not committed. An empty
  folder holds nothing to overwrite, and may be left by a failed
  create.
  Not chosen: checking only the checkout, because writing over a
  deleted but committed slice would replace it silently.
  Cost: two places to check.
  Rule: folder-taken
  Replaces: slice-and-events/SD6 (in part: the worktree folder becomes
  the slice's folder in the checkout)

- **SD8.** When `create` fails, sliceflow removes the slice's folder it
  made, and puts its files in git's staging area back as they were.
  Why: files left behind would be swept into the user's next commit,
  show up as a slice that was never created, and block a retry.
  Once the commit is made, the slice exists and nothing is removed,
  because the only way back would be to rewrite the user's branch.
  Not chosen: leaving the files for the user to clean up; undoing a
  commit that was already made, or listing it as a leftover to revert.
  Cost: an undo step in `create`, which must tell whether the commit was
  made.
  Rule: create-undo, create-git-step-fails, create-leftovers
  Replaces: slice-and-events/SD31 (the leftovers now name a folder),
  slice-and-events/SD26 (in part: after the commit, the slice is kept)

- **SD9.** The user's checkout is the checkout the user runs the
  command in (`git rev-parse --show-toplevel`), a second worktree
  included. The main project's folder is used only to place the slice's
  worktree (0005).
  Why: the wrong-branch check asks git for the branch of the checkout
  the user is in, so the files must be written and read there too.
  Otherwise, in a second worktree on `dev/bert`, the check would pass
  while the files went to a main checkout on `main`.
  Not chosen: refusing to run in a second worktree, because people who
  keep several checkouts could not plan from them.
  Cost: two ideas of "the project's folder" in the code.
  Rule: create-in-your-folder, status-reads-your-checkout
  Replaces: slice-and-events/SD11 (in part)

- **SD10.** The lock of a log is a file inside git's folder
  (`git rev-parse --git-path sliceflow/locks/<name>.lock`), not next to
  the log.
  Why: the lock file is never deleted, because another process may be
  waiting on it. Next to the log, it now lands in the user's checkout:
  `git status` shows it, and `git add -A` commits it. Git never shows or
  commits files inside `.git`.
  Not chosen: adding `*.lock` to `.git/info/exclude`, because sliceflow
  would edit the user's git files; leaving it untracked in the checkout.
  Cost: the lock is no longer beside the file it protects, so the code
  must pass its path around.
  Rule: log-lock-in-git-folder

## Notes for the design

- Each slice-and-events rule that a rule here replaces gets a line
  `Replaced by: plan-on-base` in slice-and-events/spec.md, when its new
  rule is written, so the old spec points to the change, as a
  superseded decision record does.

- Who writes the code (agreed): the human writes the wrong-branch check,
  the new undo step (removing the slice's folder when the commit fails),
  and reading the log only from the checkout. The agent writes the rest:
  removing the branch and worktree steps, the commit in the user's
  folder, messages, help, wiring, and reading `sliceflow.base` (the
  human words its error for a missing branch).

- How the undo is built is the human's design (task 6). The second
  review found that today's `runSteps` undoes only the steps before the
  one that failed, and that `git.Commit` stages and commits in one step,
  so nothing would unstage the files when the commit fails. Its options:
  (a) staging becomes its own step, whose undo is `git reset`; (b) the
  commit step cleans up after itself, and checks first whether the
  commit landed; (c) `runSteps` gets a point of no return, after which
  nothing is undone.

## Tasks

Existing tests that changed, because the behavior they test changed,
checked against `c5cc50f`, the commit before task 1. Read these
closely.

- Create writes in the user's checkout, with no branch or worktree:
  e2e `TestCreateASlice`, `TestCreateWritesTheSpec`,
  `TestCreateWritesTheFirstEvent`, `TestCreateFromAnotherBranch`,
  `TestCreateWithAnEmptyFrom`, `TestCreateNotes` (the `Note:` padding),
  `TestCreateFromASubfolder`.
- Removed, because create makes no worktree: e2e
  `TestCreateWithAWorktreeSetting`,
  `TestCreateWhenTheWorktreeFolderIsNotEmpty`, and the helper
  `realPath`.
- The base and its messages: e2e `TestCreateWithNoBase` (three
  messages, by where the base came from).
- The folder check also looks at the disk, and its message no longer
  names the base: e2e `TestCreateWhenTheNameIsTaken`; unit
  `TestCheckNameFreeFolderNotCommitted` (a folder never committed now
  takes the name), and `TestCheckNameFreeOutsideARepo` and
  `wantNameCheck` (the new `top` input).
- The undo removes files and unstages them, not a worktree and a
  branch: e2e `TestCreateWhenTheCommitFails`, `TestCreateStopsOnCtrlC`,
  `TestCreateWhenSomethingCannotBeRemoved` (the `folder` leftover), and
  the helper `writeHook`; unit `TestPrintCreateError` (the leftover
  lines), and `TestRunStepsUndoesNewestFirst`,
  `TestRunStepsFirstStepFails`, `TestRunStepsSkipsANilUndo`,
  `TestRunStepsKeepsGoingWhenAnUndoFails`, and
  `TestRunStepsListsEachLeftoverNewestFirst` (the failed step's own Undo
  now runs too, and the leftover kind is `folder`).
- The caller gives the lock's path: every Create and Append test in
  `internal/eventlog/append_test.go` passes one, and so do the helpers
  `holdLock`, `mustAppend`, and `runHelper`.
- Status and list read only the checkout: e2e `TestStatusOfNoSlice`,
  `TestStatusWithBadLines`, `TestStatusWithNoCreatedEvent`,
  `TestListSlices`, `TestListSkipsCandidates`, `TestListDamagedLogs`,
  and the helper `createdTime`; unit `TestStatusText` (the new lines).
- Removed, because nothing reads other branches: e2e
  `TestStatusWithItsWorktree`, `TestStatusFromTheBranch`,
  `TestStatusFromMain`, `TestListMergedWithItsBranch`, and the helpers
  `mergeIntoMain`, `removeWorktree`, and `worktreeLog`; unit
  `TestFileAtACommittedFile`, `TestFileAtIgnoresChangesNotCommitted`,
  `TestFileAtNotFound`, `TestFoldersAt`, and `TestBranchesWithPrefix`.
- The help text: the constants `wantCreateHelp` and `wantStatusHelp` in
  `internal/cli/run_test.go`.
- Only a failure message: the helper `wantSamePath` in
  `internal/git/repo_test.go` no longer names `RepoRoot`.

Listed here before, but not changed: `TestStatusOfAnInvalidName`,
`TestCreateRunsTheUsersHooks`, the helper `createSlice`,
`internal/cli/slice_list_test.go`, `internal/project/path_test.go`, and
`internal/git/commit_test.go`, which only got new tests.

Code that changes, besides the steps below: new git functions for the
current branch and the top of the current checkout, and `git reset --`
for the unstage; `Leftover` kinds and how leftovers print;
`NoLogError`, `StatusResult`, and `SkippedCandidate` lose their branch,
worktree, and main fields; the messages in `internal/cli/text.go`;
`create` stops reading `sliceflow.worktree.dir`; `NoBaseError` records
where the base came from (the flag, the setting, or the default), not
only whether it was chosen; a new git function for
`merge-base --is-ancestor`. `FileAt`, `FoldersAt`,
`BranchesWithPrefix`, and `WorktreeOf` in `internal/git` may become
unused. Kept for 0005, though unused: `CreateBranch`, `DeleteBranch`,
`AddWorktree`, `RemoveWorktree`, `WorktreeOf`, and
`project.WorktreePath`. Removed, with their tests, because sliceflow no
longer reads other branches (SD5): `FileAt`, `FoldersAt`, and
`BranchesWithPrefix`.

In order, from the command and its e2e tests down. "Human" means the
human writes the code, and the agent writes its tests from the human's
test design; "agent" means the agent writes the code from the human's
design. Every existing test that changes is listed, with its reason.

1. Change the e2e tests to the new behavior first; they fail until the
   work is done. Rules: all. Agent, from the human's test design.
2. `create`: pick the base (`--from`, then `sliceflow.base`, then
   `main`), and the three messages for a missing base. Rules:
   `base-branch`, `base-missing`. Agent; the human worded the messages.
3. `create`: the wrong-branch check, and its place in the error order.
   Rules: `create-on-the-base`, `create-error-order`. Human.
4. `create`: the folder check (not empty in the checkout, or in the
   base commit). Rules: `folder-taken`. Agent.
5. `create`: drop the branch and worktree steps, write in the checkout,
   commit only the slice's folder, and the new output. `git.Commit`
   today commits everything that is staged; change it to commit only
   its paths (`git commit --only -- <paths>`), with its test. During a
   merge, git refuses such a commit, and sliceflow shows git's message. Rules:
   `create-in-your-folder`, `create-touches-only-its-folder`,
   `create-output`. Agent.
5b. `eventlog`: take the lock at a path the caller gives, inside git's
   folder, not at the log's path plus `.lock`. Rules:
   `log-lock-in-git-folder`. The human designs the change to the
   functions' inputs (the event log is core); the agent writes it.
6. `create`: the undo (remove the folder, unstage its files) and the
   message for what is left. Rules: `create-undo`,
   `create-git-step-fails`, `create-leftovers`. Human.
7. `status` and `list`: read only the checkout, and the message when no
   log is there. Rules: `status-reads-your-checkout`, `no-slice`,
   `list-finds-slices`. Human.
8. `status` and `list`: the new lines, the two warnings for old
   copies, and the help text. Rules: `status-lines`,
   `old-copy-warning`, `list-old-copies`, `help-text`. Agent.
