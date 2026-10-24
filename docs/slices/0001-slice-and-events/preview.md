---
slice: 0001-slice-and-events
---

# Preview

## Setup

These steps use a terminal with `sh`, like bash or Git Bash. `$T` is
a new empty test folder, e.g. `T=$(mktemp -d)`.

1. In the slice worktree, run `go build -o "$T/sliceflow" ./cmd/sliceflow`.
   Expect: no output, and `$T/sliceflow` exists (`sliceflow.exe` on
   Windows).
2. Run `git init -b main "$T/repo" && cd "$T/repo" && git commit --allow-empty -m first`.
   Expect: a repo on `main` with one commit.
3. Run `alias sf="$T/sliceflow"`.
   Expect: no output.

## Checks

### P1. Create a slice

Covers: S1

1. Run `sf slice create spec-submit`.
   Expect:
   ```
   Created slice spec-submit.
     Branch:   slice/spec-submit (from main @ <short commit>)
     Worktree: .sliceflow/worktrees/spec-submit
     Spec:     docs/slices/spec-submit/spec.md
   Next: write the spec in .sliceflow/worktrees/spec-submit/docs/slices/spec-submit/spec.md
   ```
   where `<short commit>` is what `git log -1 --format=%h main` shows.
2. Run `echo $?`.
   Expect: `0`.
3. Run `cat .sliceflow/.gitignore`.
   Expect: the one line `*`.
4. Run `git log -1 --format=%s slice/spec-submit`.
   Expect: `spec-submit: slice created`.
5. Run `git show --name-only --format= slice/spec-submit`.
   Expect: `docs/slices/spec-submit/.events.jsonl` and
   `docs/slices/spec-submit/spec.md`, and nothing else.
6. Run `cat .sliceflow/worktrees/spec-submit/docs/slices/spec-submit/.events.jsonl`.
   Expect: one line, like
   `{"version":1,"seq":1,"event":"slice.created","at":"<now, UTC>","data":{"from":{"branch":"main","commit":"<full commit>"}}}`.
7. Run `git status --short`.
   Expect: no output. The main checkout is not changed.

### P2. See one slice

Covers: S13

1. Run `sf slice status spec-submit`.
   Expect:
   ```
   Slice:    spec-submit
   Title:    spec-submit
   Phase:    planning
   From:     main @ <short commit>
   Created:  <the time of P1, UTC, to the minute> UTC
   ```

### P3. See all slices

Covers: S20

1. Run `sf slice create login-google --title "Login with Google"`.
   Expect: `Created slice login-google.` and its lines, as in P1.
2. Run `sf slice create first-slice`, then
   `git merge --no-ff -m "merge first-slice" slice/first-slice`.
   Expect: `main` now has `docs/slices/first-slice/`.
3. Run `sf slice list`.
   Expect:
   ```
   NAME          TITLE              PHASE     CREATED (UTC)
   spec-submit   spec-submit        planning  <time of P1>
   login-google  Login with Google  planning  <time of step 1>
   ```
   and no `first-slice`.
4. Run `sf slice list --all`.
   Expect: `first-slice` is in the list too, last.

### P4. Get help

Covers: S26

1. Run `sf slice create --help`, then `echo $?`.
   Expect: the usage of `slice create` with `--title`, `--from`, and
   `--json`, then `0`.

### P5. Stop a create with Ctrl-C

Covers: R18

1. Run `printf '#!/bin/sh\nsleep 30\n' > .git/hooks/pre-commit && chmod +x .git/hooks/pre-commit`.
   Expect: no output.
2. Run `sf slice create stop-me`, and press Ctrl-C within 30 seconds.
   Expect:
   `Error: could not create slice stop-me: interrupted. Everything sliceflow made was removed.`
3. Run `git branch --list slice/stop-me`, `git worktree list`, and
   `ls .sliceflow/worktrees`.
   Expect: no `stop-me` in any of them.
4. Run `rm .git/hooks/pre-commit`.
   Expect: no output.
5. On Windows, make the hook as in step 1 in Git Bash. Then run step 2
   in PowerShell or cmd, as `<test folder>\sliceflow.exe slice create stop-me`:
   Git Bash's own window does not pass Ctrl-C on to Windows programs.
   Then do steps 3 and 4.
   Expect: the same results as steps 2 to 4.
