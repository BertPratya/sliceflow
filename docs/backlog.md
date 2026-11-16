# Backlog

| Item | From | Added | Closed by |
|---|---|---|---|
| Text that is not valid UTF-8 passes the control-character check (choice 1 of step 3b). | 0001, PR #1 | 2026-11-10 | |
| A right-to-left override (U+202E) is not a control character, so it passes R29. | 0001, PR #1 | 2026-11-10 | |
| Words in wrong-use messages are printed as typed, with no escaping, e.g. a line break or ESC in a typed word. | 0001, PR #1 | 2026-11-10 | |
| `RepoRoot` takes the parent folder of the shared `.git`. In a submodule, a project made with `--separate-git-dir`, or a bare repository used with worktrees, that is not the project's folder, so the folder and name are wrong (e.g. `vendor`, not `lib`). Rare, so handled later (decided, the human). | 0001, step 5 | 2026-11-12 | |
| On Windows, a `sliceflow.worktree.dir` like `\wt` (no drive) or `C:wt` (a drive, but not a full path) is not a full path for `filepath.IsAbs`, so `WorktreePath` joins it to the project's folder: `\wt` gives a folder inside the project, and `C:wt` gives a broken path, so git fails (R37). Rare, so handled later (decided, the human). | 0001, step 5 | 2026-11-13 | |
| When the project is at the top of the disk (`/`, or `C:\`), `filepath.Base` of it is not a name, so the worktree path has no project name in it. Rare, so handled later (decided, the human). | 0001, step 5 | 2026-11-13 | |
| The lock file `.events.jsonl.lock` stays next to the log after an event is written, so `git status` in a new worktree shows it as untracked. Deleting it after use is not safe: another process may be waiting on it. Options: keep it inside `.git`, or add `*.lock` to `.git/info/exclude`. Decide with the event log work of step 8 (decided, the human). | 0001, step 7 | 2026-11-16 | |
| `internal/gittest` (TD6 in the 0001 plan) was not built: each package keeps its own test helpers for temporary git repos (e.g. `gitIn`, `newRepoWithMain`, `writeText`). Build it when 0002's tests show which helpers are really shared (decided, the human). | 0001, step 9 | 2026-11-16 | |
| In a partial clone, `git ls-tree` may fetch from the remote, which breaks "no data out". `GIT_NO_LAZY_FETCH=1` would stop it, but needs git 2.44. | 0001, PR #1 | 2026-11-10 | |
| sliceflow needs git 2.31 or later (`--path-format`), but R14 says it does not check git's version. | 0001, PR #1 | 2026-11-10 | |
