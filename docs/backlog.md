# Backlog

| Item | From | Added | Closed by |
|---|---|---|---|
| Text that is not valid UTF-8 passes the control-character check (choice 1 of step 3b). | 0001, PR #1 | 2026-11-10 | |
| A right-to-left override (U+202E) is not a control character, so it passes R29. | 0001, PR #1 | 2026-11-10 | |
| Words in wrong-use messages are printed as typed, with no escaping, e.g. a line break or ESC in a typed word. | 0001, PR #1 | 2026-11-10 | |
| From a linked worktree of a submodule, or of a project made with `--separate-git-dir`, `FindProject` gives the repository's folder, not the project's. | 0001, PR #1 | 2026-11-10 | |
| On Windows, a `sliceflow.worktree.dir` like `C:wt` (a drive, but not a full path) is not handled. On Linux and macOS, `~\a\b` gives a folder named `a\b`. | 0001, PR #1 | 2026-11-10 | |
| In a partial clone, `git ls-tree` may fetch from the remote, which breaks "no data out". `GIT_NO_LAZY_FETCH=1` would stop it, but needs git 2.44. | 0001, PR #1 | 2026-11-10 | |
| sliceflow needs git 2.31 or later (`--path-format`), but R14 says it does not check git's version. | 0001, PR #1 | 2026-11-10 | |
