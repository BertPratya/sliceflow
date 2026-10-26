# Plan: 0001 slice and events

How the rest of `docs/specs/0001-slice-and-events.md` is built. The spec
says what; this plan says how, and splits the work into chunks. Each chunk
is one commit. Already built: R2 (`slice.ValidateName`), and the git parts
of R12, R13, R16.

## Packages

| Package | Job | Never |
|---|---|---|
| `cmd/sliceflow` | `main`: calls `cli.Run(os.Args[1:], os.Stdout, os.Stderr)`, then `os.Exit` | logic |
| `internal/cli` | reads words and flags, calls `slice`, prints, picks the exit code | runs git |
| `internal/slice` | the rules of slices: checks, creating, status, list | prints |
| `internal/eventlog` | writes and reads `.events.jsonl` lines | runs git, prints |
| `internal/git` | runs git, returns plain answers | knows what a slice is |
| `internal/gittest` | temporary repos for tests | used outside tests |
| `templates` | embeds `templates/default/spec.md` with `go:embed` | logic |
| `internal/e2e` | tests that run the built `sliceflow` program | |

## Plan decisions

- **PD1. Errors leave `slice` as finished sentences.** Every error that
  `slice` returns to `cli` is the user's message without `Error: `, e.g.
  `invalid slice name "x": use ... Nothing was changed.`. `cli` prints
  `Error: ` + the message on stderr. A message may have more lines (S15,
  R18). The check errors of R12-R17 end with `. Nothing was changed.`;
  `slice` adds it in one helper. Wrong-use errors (R38) are made in `cli`.
  Why: one rule, so no command can forget the suffix or add it twice.
- **PD2. cli writes to writers it is given.** `cli.Run(args []string,
  stdout, stderr io.Writer) int`. Commands get an `*env` with `stdout`,
  `stderr`, and `dir` (the current folder). Why: tests can read the output.
- **PD3. Our own flag parser, not the `flag` package.** `internal/cli/flags.go`
  parses `--name value`, `--name=value`, bool flags, and `-h`/`--help`,
  before and after the positional words. Why: R38 needs exact error texts,
  and S2 puts flags after the name; `flag` stops at the first word that is
  not a flag, and prints its own texts.
- **PD4. git errors keep all of git's lines.** `gitError` returns a
  `*git.CommandError` with `Command` (e.g. `commit`), `ExitCode`, and
  `Lines` (git's stderr lines that are not empty). `Error()` is the first
  line, or `git <command> failed (exit code <n>)`. R18 reads `Lines[1:]`.
- **PD5. The spec template has one source.** `templates/templates.go`
  embeds `default/spec.md`. `slice` makes the built-in template from it
  (R7): drops the `title:` line and the title rule, fills `slice` and
  `created`, and sets `ideas: []`. A test fails when one of these edits
  finds nothing to change, so a template edit cannot break it silently.
- **PD6. The main checkout's root** is the parent of
  `git rev-parse --path-format=absolute --git-common-dir`, so a command
  run inside a worktree still finds it (R5).
- **PD7. End-to-end tests run the real program.** `internal/e2e` builds
  `sliceflow` once in `TestMain`, and runs it in `gittest` repos with
  `exec.Command`, checking stdout, stderr, and the exit code. Every
  scenario S1-S29 has an e2e test named after it, e.g. `TestS5`.
- **PD8. Time comes in as a value.** `slice.Create` takes a `Now
  time.Time`; `eventlog` takes times as values. `cli` passes `time.Now()`.
  Tests pass a fixed time.

## Chunks

Each chunk: the requirements it covers, the files, the names others rely
on, and the tests. All chunks keep `go vet` and `go test ./...` green.

### Chunk 1: cli plumbing

- **Covers:** R38, R39, and PD2, PD3, PD7.
- **Files:** `cmd/sliceflow/main.go`, `internal/cli/*.go`,
  `internal/e2e/main_test.go`, `internal/e2e/cli_test.go`.
- **Names:** `cli.Run(args, stdout, stderr) int`; `type env struct
  { stdout, stderr io.Writer; dir string }`; `command.run` becomes
  `func(e *env, args []string) int`; a leaf command gets a `usage` text for
  `--help`; `parseFlags(args []string, spec flagSpec) (flags, positional,
  error)` in `flags.go`; `wrongUse(e, msg) int` prints
  `Error: <msg>. Run sliceflow --help.` and returns 2.
- **Behavior:** `-h`/`--help` at any level prints that level's usage on
  stdout, exit 0. Unknown command, unknown flag, missing value, missing
  name, and too many arguments print R38's texts on stderr, exit 2,
  nothing on stdout. `slice create` gets its flags (`--from`, `--title`,
  `--json`), and still only prints that the checks are not built yet.
- **Tests:** unit tests for `parseFlags`; e2e `TestS26`, plus e2e tests for
  each R38 text, using `slice create`.

### Chunk 2: the create checks

- **Covers:** R11-R17, R3 (`--from` reaches `baseBranch`), R14 printed,
  PD1, PD4, PD6.
- **Files:** `internal/git/git.go` (+test), `internal/slice/create.go`
  (+test), `internal/cli/slice.go`, `internal/e2e/create_checks_test.go`.
- **Names:** `git.CommandError`; `git.MainRoot(dir) (string, error)`;
  `git.RemoteBranches(dir, name) ([]string, error)` (all
  `refs/remotes/<remote>/<name>`); `git.PathExists(dir, rev, path) (bool,
  error)` (a path on a branch); `slice.CreateOptions{Dir, Name, From,
  Title string; Now time.Time}`; `slice.Create(opts) (Result, error)`.
- **Behavior:** checks in the R11 order, each error as PD1. When all pass,
  `Create` returns an error `creating a slice is not built yet` (until
  chunk 4).
- **Tests:** e2e `TestS5`-`TestS11`, and S7 with a local bare remote that
  the test pushes to.

### Chunk 3: writing an event

- **Covers:** R19, R20.
- **Files:** `internal/eventlog/event.go` (+test).
- **Names:** `eventlog.Event{Version, Seq int; Event string; At
  time.Time; Data json.RawMessage}`; `eventlog.SliceCreated{From
  Ref; Title string (omitempty)}`; `eventlog.Ref{Branch, Commit string}`;
  `eventlog.Line(e Event) ([]byte, error)`: one line, `\n` at the end,
  fields in R19's order, `at` as `2026-10-24T07:42:00Z`, no HTML escaping.
- **Tests:** unit tests with the exact line of S1 and with a title (S2).

### Chunk 4: creating the slice

- **Covers:** R1, R4-R10, R20 (title), R41, PD5.
- **Files:** `templates/templates.go`, `internal/slice/create.go`,
  `internal/slice/spec.go` (+tests), `internal/cli/slice.go`,
  `internal/e2e/create_test.go`.
- **Behavior:** after the checks: write `.sliceflow/.gitignore` if missing;
  `git branch --no-track slice/<name> <base commit>`; `git worktree add`;
  write `spec.md` and `.events.jsonl` with LF endings; `git add` both;
  `git commit -m "<name>: slice created"`, as a normal commit. `Result`
  holds what R9 and R10 print. Paths printed with `/` (R41).
- **Tests:** e2e `TestS1`-`TestS4`: the output (commit ids and dates
  matched by pattern), the files on the branch, the one commit, the log
  line.

### Chunk 5: removing what was made

- **Covers:** R18.
- **Files:** `internal/slice/create.go` (+test), `internal/e2e/create_undo_test.go`.
- **Behavior:** each step that makes something adds its undo to a list;
  on a failed step, undo newest first, and build R18's message, with the
  other lines of git's message indented 2 spaces, and
  `Could not remove: <list>.` when an undo fails. Ctrl-C or a stop signal
  while creating: the reason is `interrupted`. `cli` passes a `context`
  that `signal.NotifyContext` cancels.
- **Tests:** e2e `TestS12` (a pre-commit hook `exit 1`), and a hook that
  prints two lines.

### Chunk 6: reading the log

- **Covers:** R25-R27.
- **Files:** `internal/eventlog/parse.go` (+test).
- **Names:** `eventlog.Parse(content []byte) Log`; `Log{Events []Event;
  Warnings []string}`; `eventlog.Phase(log Log) string`.
- **Behavior:** every reason of R26, with field names like
  `data.from.branch`; duplicate fields found with `json.Decoder` tokens;
  the `seq` rule of R27; phase `planning` or `unknown`.
- **Tests:** a table test with one row per reason of R26, and the `seq`
  cases of R27.

### Chunk 7: slice status

- **Covers:** R21-R24, R28-R33.
- **Files:** `internal/git/git.go` (+test), `internal/slice/status.go`
  (+test), `internal/cli/slice.go`, `internal/e2e/status_test.go`.
- **Names:** `git.CurrentBranch(dir) (string, error)` (`""` when
  detached); `git.ShowFile(dir, rev, path) ([]byte, bool, error)`;
  `slice.Status(dir, name string) (Info, error)`; `slice.Info` holds what
  R29 and R30 show, and its warnings.
- **Tests:** e2e `TestS13`-`TestS19`, `TestS28`.

### Chunk 8: slice list

- **Covers:** R34-R37, R40, R43, and `--all`.
- **Files:** `internal/git/git.go` (+test), `internal/slice/list.go`
  (+test), `internal/cli/slice.go`, `internal/e2e/list_test.go`,
  `internal/e2e/network_test.go`.
- **Names:** `git.Branches(dir, prefix) ([]string, error)`;
  `git.ListDir(dir, rev, path) ([]string, error)`;
  `slice.List(dir string, all bool) (Listing, error)`.
- **Tests:** e2e `TestS20`-`TestS25`, `TestS27`, `TestS29`; R40: every
  command works with a remote whose URL cannot be reached.

## Assumed

Choices the spec does not settle. The human checks them.

- An empty `git config sliceflow.base` counts as not set (built).
- The worktree folder and `.sliceflow/` folders that `Create` makes count
  as "what sliceflow made" for R18, and are removed too.
