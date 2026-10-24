---
slice: 0001-slice-and-events
---

# Tasks

## Task 01: git-runner

```yaml
covers: [R40]
depends_on: []
touches: [go.mod, internal/git/, internal/testrepo/]
follows: [TD1, TD3, TD8, Git arguments, No network, Testing]
risk: core
```

### Goal

The Go module exists, and every later task runs git through one
function. It runs one git command as its own process, with the
environment of No network, and returns its output or an error with
git's message. A test helper makes temporary repos with a test git
config, for every package's tests.

### Steps

1. Write the tests listed under Done when. Run them, and see them fail.
2. Create `go.mod`: module `github.com/BertPratya/sliceflow`, `go 1.25`
   (TD1, TD3). No `require` lines.
3. In `internal/git/git.go`, add `Run`: find `git` once with
   `exec.LookPath`, run it with `exec.CommandContext`, the given folder
   as its working directory, and `GIT_TERMINAL_PROMPT=0` and
   `GIT_NO_LAZY_FETCH=1` added to the environment.
4. Add `Error`: the arguments, and git's stderr with spaces trimmed.
   Its `Error()` is that stderr text.
5. Add `RefExists` (`git rev-parse --verify --quiet <ref>^{commit}`)
   and `Show` (`git cat-file -e`, then `git show <ref>:<path>`).
6. In `internal/testrepo/testrepo.go`, add `Setup`, `New`, and `Git`.
7. Run the tests until they pass.

### Files

- `go.mod`: new. The module line and `go 1.25`.
- `internal/git/git.go`: new. `Run`, `Error`, `RefExists`, `Show`.
- `internal/git/git_test.go`: new. The unit tests below.
- `internal/testrepo/testrepo.go`: new. Test helpers; only test files
  import this package.

### Interfaces

- `func Run(ctx context.Context, dir string, args ...string) (string, error)`:
  creates. Returns stdout. A failed command returns a `*git.Error`.
  Every later task uses it.
- `type Error struct{ Args []string; Stderr string }`: creates. Task 08
  puts `Stderr` into R18's `<reason>`.
- `func RefExists(ctx context.Context, dir, ref string) (bool, error)`:
  creates. Tasks 04, 06, 07, and 09 use it.
- `func Show(ctx context.Context, dir, ref, path string) ([]byte, bool, error)`:
  creates. The bool is false when the file is not in that commit.
  Task 09 uses it.
- `func Setup(t testing.TB)`: creates. Sets `GIT_CONFIG_GLOBAL` to a
  temporary file with `user.name`, `user.email`, and
  `init.defaultBranch main`, and `GIT_CONFIG_NOSYSTEM=1`, with
  `t.Setenv`. Tasks 03 to 11 use it.
- `func New(t testing.TB) string`: creates. Calls `Setup`, makes a repo
  in `t.TempDir()` on `main` with one empty commit `first`, and returns
  its folder. Tasks 03 to 11 use it.
- `func Git(t testing.TB, dir string, args ...string) string`: creates.
  Runs git in `dir`, fails the test on an error, returns stdout. Tasks
  03 to 11 use it.

### Done when

- `TestRunReturnsOutput`: in a `testrepo.New` repo,
  `Run(ctx, dir, "rev-parse", "--abbrev-ref", "HEAD")` returns `main\n`.
  Proves: R40
- `TestRunErrorHasStderr`: `Run(ctx, dir, "rev-parse", "--verify", "nope")`
  returns a `*git.Error` whose `Stderr` is not empty and has no
  trailing newline.
  Proves: R40
- `TestRunSetsNoNetworkEnv`: in a repo whose local config has
  `alias.env-check` set to `!env`, the output of
  `Run(ctx, dir, "env-check")` holds the lines `GIT_TERMINAL_PROMPT=0`
  and `GIT_NO_LAZY_FETCH=1`.
  Proves: R40
- `TestRefExists`: `refs/heads/main` is true; `refs/heads/slice/none` is
  false with no error.
  Proves: R40
- `TestShowMissingFile`: `Show(ctx, dir, "main", "docs/none.md")`
  returns `nil, false, nil`; a committed `a.txt` with `hi\n` returns
  `hi\n, true, nil`.
  Proves: R40

### Out of scope

- Finding the repo and the main checkout. Done in: task 04
- The test that no `fetch`, `pull`, or `push` runs. Done in: task 11

## Task 02: event-log

```yaml
covers: [R19, R25, R26, R27]
depends_on: []
touches: [internal/eventlog/]
follows: [Events in the code, Reading a line, TD4, TD5, Testing]
risk: core
```

### Goal

The event log format exists as a package with no file or git access.
It writes one `slice.created` line, and reads a whole log into events
and warnings in R26's and R27's exact words. The hardest part is the
reader: every bad line is skipped with the right reason, including a
field that appears twice.

### Steps

1. Write the tests listed under Done when. Run them, and see them fail.
2. In `internal/eventlog/event.go`, add `Event`, `Created`, `From`, and
   `NewCreated`, with the fields declared in R19's order (TD4).
3. Add `Encode`: one line of JSON, ending in `\n`.
4. In `internal/eventlog/parse.go`, add `Parse` as Reading a line says:
   split on `\n`, find a name that appears twice with
   `json.Decoder.Token`, then read the line into a map and check each
   field (TD5).
5. Add `Phase`.
6. Run the tests until they pass.

### Files

- `internal/eventlog/event.go`: new. The types, `NewCreated`, `Encode`.
- `internal/eventlog/parse.go`: new. `Parse`, `Phase`.
- `internal/eventlog/event_test.go`: new. Writer tests.
- `internal/eventlog/parse_test.go`: new. Reader tests.

### Interfaces

- `type Event struct{ Version int; Seq int; Event string; At string; Data any }`:
  creates, with JSON names `version`, `seq`, `event`, `at`, `data`.
  After `Parse`, `Data` of a `slice.created` event is a `Created`.
  Tasks 06 and 09 use it.
- `type Created struct{ From From; Title string }`: creates, JSON
  `from` and `title,omitempty`. Tasks 06 and 09 use it.
- `type From struct{ Branch string; Commit string }`: creates, JSON
  `branch` and `commit`. Tasks 06 and 09 use it.
- `func NewCreated(at time.Time, from From, title string) Event`:
  creates. `Version` 1, `Seq` 1, `At` in UTC as
  `2006-01-02T15:04:05Z`. Task 06 uses it.
- `func Encode(e Event) ([]byte, error)`: creates. Task 06 uses it.
- `func Parse(log []byte) ([]Event, []string)`: creates. The warnings
  are R26's and R27's text, e.g. `.events.jsonl line 2: not valid JSON; skipped`.
  Task 09 uses it.
- `func Phase(events []Event) string`: creates. `planning` or
  `unknown`. Task 09 uses it.

### Done when

- `TestEncodeCreated`: `NewCreated` at 2026-10-24 07:42:00 UTC, from
  `main` and commit `9dd7ebf4c2a1e8d3b5f6a7c8d9e0f1a2b3c4d5e6`, no
  title, encodes to exactly
  `{"version":1,"seq":1,"event":"slice.created","at":"2026-10-24T07:42:00Z","data":{"from":{"branch":"main","commit":"9dd7ebf4c2a1e8d3b5f6a7c8d9e0f1a2b3c4d5e6"}}}`
  and `\n`; with title `Login with Google`, `data` ends with
  `,"title":"Login with Google"}`.
  Proves: R19
- `TestEncodeLocalTimeIsUTC`: a time at 09:42 in UTC+2 gives
  `"at":"2026-10-24T07:42:00Z"`.
  Proves: R19
- `TestParseGoodLog`: the line of `TestEncodeCreated` gives one event,
  no warnings, and `Phase` `planning`; the same line ending in `\r\n`,
  or with no `\n` at the end, gives the same.
  Proves: R25, R26
- `TestParseBadLines`: one table row per reason, each a log of a good
  line 1 and a bad line 2, each giving one warning
  `.events.jsonl line 2: <reason>; skipped`: a line of spaces
  (`blank line`), `{` (`not valid JSON`), `"version":2`
  (`unknown version 2`), `"event":"slice.moved"`
  (`unknown event "slice.moved"`), an extra `"by":"x"`
  (`unknown field "by"`), `"seq":2,"seq":2`
  (`field "seq" appears twice`), a `data.from` with `branch` twice
  (`field "data.from.branch" appears twice`), no `at`
  (`missing field "at"`), `"seq":"2"` and `"seq":2.0`
  (`field "seq" has the wrong type`), `"at":"2026-10-24 07:42"`,
  a 7-character `commit`, and `"branch":""`
  (`field "<f>" has a wrong value`, with `at`, `data.from.commit`, and
  `data.from.branch`).
  Proves: R26
- `TestParseSeqGap`: lines with `seq` 1, 2, 4, 5 give all four events
  and one warning `.events.jsonl line 3: seq 4, expected 3`.
  Proves: R27
- `TestPhaseUnknown`: `Phase(nil)` is `unknown`.
  Proves: R25

### Out of scope

- Reading the log from git, and the `.events.jsonl not found` warning.
  Done in: task 09
- Adding a second event to a log. Done in: not in this slice

## Task 03: cli-skeleton

```yaml
covers: [R38, R39]
depends_on: ["01", "04"]
touches: [cmd/sliceflow/, internal/cli/cli.go, internal/cli/flags.go,
  internal/cli/output.go, internal/cli/cli_test.go,
  internal/cli/flags_test.go, e2e/helpers_test.go, e2e/cli_skeleton_test.go]
follows: [TD2, TD7, Shared logic in internal, CLI, Core API, Testing]
risk: core
```

### Goal

`sliceflow` runs, shows help, and stops wrong use with exit code 2.
Commands register themselves from their own files, so tasks 06, 09,
and 10 each add a file and never edit `cli.go`. The e2e helpers exist.
The hardest part is the flag helper of TD2.

### Steps

1. Write the tests listed under Done when. Run them, and see them fail.
2. `cmd/sliceflow/main.go`: `os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr))`.
3. `internal/cli/cli.go`: `Run`, `command`, `register`. `Run` finds
   `slice <name>` in the registry; with no command, `--help`, or `-h`
   it prints the list of commands on stdout and returns 0.
4. `internal/cli/flags.go`: `parseFlags` (TD2). `-h` and `--help`
   print the command's usage on stdout and return `errHelp`.
5. `internal/cli/output.go`: `usageError`, `fail`, `printWarnings`.
6. `e2e/helpers_test.go`: `TestMain` builds the binary into a temporary
   folder with `go build`; `run` runs it.
7. Run the tests until they pass.

### Files

- `cmd/sliceflow/main.go`: new. Only the line of step 2.
- `internal/cli/cli.go`: new. `Run`, `command`, `register`, top usage.
- `internal/cli/flags.go`: new. `parseFlags`, `errHelp`.
- `internal/cli/output.go`: new. Error and warning output.
- `internal/cli/cli_test.go`, `internal/cli/flags_test.go`: new.
- `e2e/helpers_test.go`: new. `TestMain`, `run`.
- `e2e/cli_skeleton_test.go`: new. The e2e tests below.

### Interfaces

- `func Run(args []string, stdout, stderr io.Writer) int`: creates.
- `type command struct{ Name, Usage string; Run func(ctx context.Context, args []string, stdout, stderr io.Writer) int }`:
  creates. `Name` is e.g. `slice create`. Tasks 06, 09, 10 use it.
- `func register(c command)`: creates. Tasks 06, 09, 10 call it from
  `init()`.
- `func parseFlags(fs *flag.FlagSet, usage string, args []string, stdout io.Writer) ([]string, error)`:
  creates. Returns the positional arguments. Tasks 06, 09, 10 use it.
- `var errHelp`: creates. A command returns 0 on it.
- `func usageError(stderr io.Writer, msg string) int`: creates. Prints
  `Error: <msg>. Run sliceflow --help.`, returns 2. Tasks 06, 09, 10
  use it.
- `func fail(stderr io.Writer, err error) int`: creates. Prints
  `Error: ` and the message, returns 1. Tasks 06, 07, 08, 09, 10 use it.
- `func printWarnings(stderr io.Writer, ws []slice.Warning)`: creates.
  R28's form. Tasks 09 and 10 use it.
- `type slice.Warning`: uses, from task 04.
- `func run(t *testing.T, dir string, args ...string) (stdout, stderr string, code int)`:
  creates, in package `e2e`. Every later e2e test uses it.
- `testrepo.Setup`, `testrepo.New`: uses, from task 01.

### Done when

- `TestParseFlagsAfterName`: `create login-google --title "Login with Google" --json`
  gives positional `[login-google]`, title `Login with Google`, json
  true; `--title=X a` gives `[a]` and `X`; `-- --x` gives `[--x]`.
  Proves: R38
- `TestParseFlagsUnknown`: `--colour` returns an error whose text is
  `unknown flag --colour`; `--title` with no value returns
  `flag needs a value: --title`.
  Proves: R38
- `TestPrintWarnings`: a warning `{Slice: "a", Text: "x"}` prints
  `Warning: a: x`; `{Slice: "", Text: "y"}` prints `Warning: y`.
  Proves: R38
- `TestHelpTop`: `sliceflow --help` prints a usage that names
  `slice create`, `slice status`, and `slice list` on stdout, nothing
  on stderr, exit code 0.
  Proves: R39
- `TestUnknownCommand`: `sliceflow slice nope` prints
  `Error: unknown command "slice nope". Run sliceflow --help.` on
  stderr, nothing on stdout, exit code 2.
  Proves: R38

### Out of scope

- The usage text of each command, and its flags. Done in: task 06,
  task 09, task 10
- Catching Ctrl-C. Done in: task 08

## Task 04: repo-and-names

```yaml
covers: [R2, R3, R12, R13, R14, R21, R22]
depends_on: ["01"]
touches: [internal/slice/repo.go, internal/slice/repo_test.go,
  internal/slice/errors.go, internal/slice/name.go,
  internal/slice/name_test.go, internal/slice/base.go,
  internal/slice/base_test.go]
follows: [Core API, TD7, Git arguments, Cross-platform paths, Testing]
risk: core
```

### Goal

The shared base of the core exists: finding the repo and the main
checkout, the error and warning types, the name rule, and picking the
base branch. Every command uses these.

### Steps

1. Write the tests listed under Done when. Run them, and see them fail.
2. `internal/slice/errors.go`: `Error`, `Warning`.
3. `internal/slice/repo.go`: `Repo`, `FindRepo`, as Flow 1, step 2 and
   Flow 2, steps 1 and 2 say. `Root` uses `/` (Cross-platform paths).
4. `internal/slice/name.go`: `CheckName`, with R14's two messages.
5. `internal/slice/base.go`: `BaseBranch`.
6. Run the tests until they pass.

### Files

- `internal/slice/errors.go`: new. `Error`, `Warning`.
- `internal/slice/repo.go`: new. `Repo`, `FindRepo`.
- `internal/slice/name.go`: new. `CheckName`.
- `internal/slice/base.go`: new. `BaseBranch`.
- `internal/slice/repo_test.go`, `name_test.go`, `base_test.go`: new.

### Interfaces

- `type Error struct{ Message string }`, with `func (e *Error) Error() string`:
  creates. Every later core task returns it.
- `type Warning struct{ Slice, Text string }`: creates. Tasks 03, 09,
  10 use it.
- `type Repo struct{ Dir, Root string; Now func() time.Time }`:
  creates. `Dir` is where the command runs; `Root` is the main
  checkout's root. Tasks 06 to 10 use it.
- `func FindRepo(ctx context.Context, dir string) (Repo, error)`:
  creates. Returns R12's or R13's `*Error`. `Now` is `time.Now`.
  Tasks 06, 09, 10 use it.
- `func CheckName(name string) error`: creates. Tasks 07 and 09 use it.
- `func BaseBranch(ctx context.Context, r Repo, from string) (branch, source string, err error)`:
  creates. `from` is `--from` or empty. `source` is `from --from`,
  `from git config sliceflow.base`, or `the default`. It does not check
  that the branch exists. Tasks 06, 07, 09 use it.
- `git.Run`, `testrepo.New`, `testrepo.Git`: uses, from task 01.

### Done when

- `TestCheckName`: `spec-submit`, `a`, `0001-x`, and 64 `a`s pass;
  `My Slice`, `-a`, `a_b`, `a/b`, 65 `a`s, and the empty name return
  `invalid slice name "<name>": use 1-64 lowercase letters, digits, and dashes, starting with a letter or digit. Nothing was changed.`;
  `con`, `nul`, `com1`, and `lpt9` return
  `invalid slice name "<name>": it is a reserved name on Windows. Nothing was changed.`;
  `com10` passes.
  Proves: R2, R14
- `TestFindRepoNotARepo`: an empty `t.TempDir()` returns
  `not inside a git repository. Nothing was changed.`
  Proves: R12, R22
- `TestFindRepoNoCommits`: a `git init` repo with no commit returns
  `the repository has no commits yet. Make a first commit, then try again. Nothing was changed.`
  Proves: R13, R22
- `TestFindRepoFromWorktree`: from a subfolder of a second worktree,
  `Root` is the main checkout's folder, with `/`.
  Proves: R12
- `TestBaseBranch`: no config gives `main`, `the default`;
  `sliceflow.base develop` gives `develop`,
  `from git config sliceflow.base`; `from` `hotfix` with that config
  gives `hotfix`, `from --from`.
  Proves: R3, R21

### Out of scope

- Checking that the base branch exists (R16). Done in: task 07
- The R23 warning. Done in: task 09

## Task 05: spec-template

```yaml
covers: [R7]
depends_on: []
touches: [internal/template/]
follows: [Cross-platform paths, Testing]
risk: normal
```

### Goal

The spec template is built into the binary with `embed`, with its
front matter filled for a new slice.

### Steps

1. Write the tests listed under Done when. Run them, and see them fail.
2. Copy `docs/templates/spec.md` to `internal/template/spec.md`. Remove
   the `title:` line of the front matter, and the template's rules
   about the title.
3. In `internal/template/template.go`, embed it, and add `Spec`: it
   fills `slice`, `created`, and `ideas: []` in the front matter.
4. Run the tests until they pass.

### Files

- `internal/template/spec.md`: new. The template.
- `internal/template/template.go`: new. `Spec`.
- `internal/template/template_test.go`: new.

### Interfaces

- `func Spec(name string, created time.Time) []byte`: creates.
  `created` is written in UTC as `YYYY-MM-DD`. Task 06 uses it.

### Done when

- `TestSpecFrontMatter`: `Spec("spec-submit", <2026-10-24 23:30 in UTC-5>)`
  starts with `---\nslice: spec-submit\ncreated: 2026-10-25\nideas: []\n`.
  Proves: R7
- `TestSpecNoTitle`: the output has no line starting with `title:`, and
  no `{{slice}}` or `{{created}}` placeholder; the rest of the body is
  the same as `docs/templates/spec.md` after its front matter.
  Proves: R7
- `TestSpecLineEndings`: the output has no `\r`.
  Proves: R7

### Out of scope

- Templates a project can change. Done in: not in this slice

## Task 06: slice-create

```yaml
covers: [R1, R3, R4, R5, R6, R7, R8, R9, R10, R19, R20, R39, R41]
depends_on: ["02", "03", "04", "05"]
touches: [internal/slice/create.go, internal/slice/create_test.go,
  internal/cli/create.go, e2e/slice_create_test.go]
follows: [Flow 1, Cross-platform paths, Git arguments, Core API, CLI, TD7, Testing]
risk: core
```

### Goal

`sliceflow slice create <name>` makes the branch, the worktree, the two
files, and the commit, and prints R9's lines or R10's JSON. The checks
before creating and the undo come in tasks 07 and 08; this task leaves
a clear place for them in `Create`.

### Steps

1. Write the tests listed under Done when. Run them, and see them fail.
2. `internal/slice/create.go`: `CreateOptions`, `Created`, `Create`, as
   Flow 1 steps 1, 2, 4 to 8 say. A check step that calls a
   `checkCreate` function, which only returns nil in this task.
3. `internal/cli/create.go`: register `slice create` with flags
   `--title`, `--from`, `--json`, and its usage. Print R9's lines with
   `Worktree:` and `Next:` relative to `Repo.Root`, or R10's JSON.
4. Run the tests until they pass.

### Files

- `internal/slice/create.go`: new. `Create` and its types.
- `internal/slice/create_test.go`: new.
- `internal/cli/create.go`: new. The `slice create` command.
- `e2e/slice_create_test.go`: new.

### Interfaces

- `type CreateOptions struct{ Name, Title, From string }`: creates.
- `type Created struct{ Slice, Branch string; From eventlog.From; Worktree, Spec string }`:
  creates. `Worktree` is the full path with `/`; `Spec` is the path
  inside the worktree.
- `func Create(ctx context.Context, r Repo, opts CreateOptions) (Created, error)`:
  creates. Task 07 adds the checks; task 08 adds the undo.
- `func checkCreate(ctx context.Context, r Repo, opts CreateOptions, base, source string) error`:
  creates, returning nil. Task 07 fills it.
- `eventlog.NewCreated`, `eventlog.Encode`, `template.Spec`,
  `FindRepo`, `BaseBranch`, `git.Run`, `register`, `parseFlags`,
  `fail`, `run`: uses.

### Done when

- `TestCreateSpecSubmit` (e2e): S1 with its exact values: a repo on
  `main`, `sliceflow slice create spec-submit` prints S1's five lines
  with the real short commit, exit code 0; `.sliceflow/.gitignore` is
  `*\n`; `slice/spec-submit` has one commit `spec-submit: slice created`
  with exactly the two files; the log is one line like S1's.
  Proves: R1, R4, R5, R6, R7, R8, R9, R19, R20
  Scenario: S1
- `TestCreateJSON` (e2e): S2: with `sliceflow.base develop`,
  `slice create login-google --title "Login with Google" --json` prints
  one JSON object with `slice`, `branch`, `from.branch` `develop`, the
  full commit, `worktree` (full path with `/`), and `spec`; the log's
  `data` has `"title":"Login with Google"`.
  Proves: R3, R10, R20, R41
  Scenario: S2
- `TestCreateFromFlag` (e2e): S3: with `sliceflow.base develop`,
  `slice create hotfix-ui --from main` starts from `main`.
  Proves: R3
  Scenario: S3
- `TestCreateInsideWorktree` (e2e): S4: run inside another slice's
  worktree, the new worktree is `.sliceflow/worktrees/spec-submit` of
  the main checkout, and the printed paths are as in S1.
  Proves: R5, R9
  Scenario: S4
- `TestCreateHelp` (e2e): S26: `slice create --help` prints a usage
  with `--title`, `--from`, `--json` on stdout, exit code 0.
  Proves: R39
  Scenario: S26
- `TestCreateKeepsGitignore` (unit): an existing `.sliceflow/.gitignore`
  with `x\n` is not changed.
  Proves: R6
- `TestCreateHookRuns` (unit): a `pre-commit` hook that writes a file
  `ran` in the repo's git folder runs once.
  Proves: R8

### Out of scope

- The checks before creating (R11 to R17). Done in: task 07
- The undo and Ctrl-C (R18). Done in: task 08

## Task 07: create-checks

```yaml
covers: [R2, R11, R12, R13, R14, R15, R16, R17, R38, R42]
depends_on: ["06"]
touches: [internal/slice/create.go, internal/slice/check.go,
  internal/slice/check_test.go, e2e/create_checks_test.go]
follows: [Flow 2, Git arguments, Core API, Testing]
risk: core
```

### Goal

`create` runs every check of R11, in R11's order, before it makes
anything, and stops at the first that fails with the spec's words.

### Steps

1. Write the tests listed under Done when. Run them, and see them fail.
2. `internal/slice/check.go`: fill `checkCreate` with R14, R16, R15,
   R17, as Flow 2 says. (R12 and R13 come from `FindRepo`.)
3. In `internal/slice/create.go`, only move `checkCreate` out to
   `check.go`.
4. Run the tests until they pass.

### Files

- `internal/slice/check.go`: new. `checkCreate`.
- `internal/slice/create.go`: changed. The stub of `checkCreate` is
  removed.
- `internal/slice/check_test.go`: new.
- `e2e/create_checks_test.go`: new.

### Interfaces

- `func checkCreate(ctx context.Context, r Repo, opts CreateOptions, base, source string) error`:
  changes. Task 06 created it.
- `CheckName`, `git.RefExists`, `git.Run`: uses.

### Done when

Each e2e test also checks that stdout is empty, the exit code is 1,
and no branch, worktree, or `.sliceflow/` was made.

- `TestCreateBadName` (e2e): S5: `slice create "My Slice"` prints S5's
  error.
  Proves: R2, R14, R38
  Scenario: S5
- `TestCreateReservedName` (e2e): S6: `slice create con`.
  Proves: R14
  Scenario: S6
- `TestCreateRemoteExists` (e2e): S7: `refs/remotes/origin/slice/spec-submit`
  made with `git update-ref` gives
  `Error: slice spec-submit already exists: branch origin/slice/spec-submit. Nothing was changed.`
  Proves: R15
  Scenario: S7
- `TestCreateBaseMissing` (e2e): S8.
  Proves: R16
  Scenario: S8
- `TestCreateFolderExists` (e2e): S9.
  Proves: R17
  Scenario: S9
- `TestCreateNotARepo` (e2e): S10.
  Proves: R12
  Scenario: S10
- `TestCreateNoCommits` (e2e): S11.
  Proves: R13
  Scenario: S11
- `TestCreateBadNameJSON` (e2e): S29.
  Proves: R42
  Scenario: S29
- `TestCheckExistsOrder` (unit): a local branch, a remote branch
  `team/origin`'s `slice/x`, and `docs/slices/x` on the base all exist:
  `<where>` is `branch slice/x`; without the local branch it is
  `branch team/origin/slice/x`; with only the folder it is
  `folder docs/slices/x on main`.
  Proves: R15
- `TestCheckOrder` (unit): a bad name and a missing base give the name
  error (R14 before R16); a missing base and an existing branch give
  the base error (R16 before R15).
  Proves: R11

### Out of scope

- Undo after a step fails. Done in: task 08

## Task 08: create-undo

```yaml
covers: [R8, R18, R38]
depends_on: ["07"]
touches: [internal/slice/create.go, internal/slice/undo.go,
  internal/slice/undo_test.go, internal/cli/create.go,
  e2e/create_undo_test.go]
follows: [Flow 3, Undo on failure, Interrupts, Testing]
risk: core
```

### Goal

When a step of `create` fails, or Ctrl-C or `SIGTERM` arrives,
sliceflow removes what it made in this run, newest first, and prints
R18's error. The hardest part is to never remove what this run did not
make.

### Steps

1. Write the tests listed under Done when. Run them, and see them fail.
2. `internal/slice/undo.go`: the undo list, as Undo on failure says.
3. In `internal/slice/create.go`: add each made thing to the list;
   check `ctx` between steps; on an error, run the undo and return
   R18's `*Error`, with `git.Error.Stderr` or `interrupted` as the
   reason.
4. In `internal/cli/create.go`: make the `ctx` with
   `signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)`.
5. Run the tests until they pass.

### Files

- `internal/slice/undo.go`: new. The undo list.
- `internal/slice/create.go`: changed. Steps add to the list; errors
  undo.
- `internal/slice/undo_test.go`: new.
- `internal/cli/create.go`: changed. The signal `ctx`.
- `e2e/create_undo_test.go`: new.

### Interfaces

- `Create`: changes. Errors after the checks are R18's `*Error`.
- `git.Error`: uses, from task 01.

### Done when

- `TestCreateHookRefuses` (e2e): S12: a `pre-commit` hook that exits
  with 1: stderr starts with
  `Error: could not create slice spec-submit: ` and ends with
  `. Everything sliceflow made was removed.`, stdout is empty, exit
  code 1; no branch `slice/spec-submit`, no worktree in
  `git worktree list --porcelain`, no `.sliceflow/`.
  Proves: R8, R18, R38
  Scenario: S12
- `TestCreateInterrupt` (e2e, skipped on Windows): a `pre-commit` hook
  that sleeps 30 seconds; after the hook starts, the test sends
  `os.Interrupt`: stderr is
  `Error: could not create slice spec-submit: interrupted. Everything sliceflow made was removed.`,
  exit code 1, nothing left.
  Proves: R18
- `TestUndoCancelledContext` (unit): `Create` with a `ctx` cancelled
  after the worktree step returns the `interrupted` error, and leaves
  no branch or worktree.
  Proves: R18
- `TestUndoKeepsOthers` (unit): with the checks skipped, a branch
  `slice/x` that already exists makes `git worktree add` fail; after
  the undo, `slice/x` still exists at its old commit.
  Proves: R18
- `TestUndoCouldNotRemove` (unit): when the worktree folder cannot be
  removed, the error ends with `Could not remove: <the folder>.`
  Proves: R18

### Out of scope

- Cleaning up after a hard kill. Done in: not in this slice

## Task 09: slice-status

```yaml
covers: [R21, R22, R23, R24, R25, R26, R28, R29, R30, R31, R32, R33, R38]
depends_on: ["02", "03", "04", "06"]
touches: [internal/slice/status.go, internal/slice/status_test.go,
  internal/cli/status.go, e2e/slice_status_test.go]
follows: [Flow 4, Core API, CLI, TD6, Testing]
risk: core
```

### Goal

`sliceflow slice status [<name>]` reads a slice's log from git and
prints its status, or R30's JSON, with its warnings. The reading part
is shared with `slice list` in task 10.

### Steps

1. Write the tests listed under Done when. Run them, and see them fail.
2. `internal/slice/status.go`: `Status` type, `Status` function, and
   `readSlice`, as Flow 4 says.
3. `internal/cli/status.go`: register `slice status` with `--json`,
   print R29's lines or R30's JSON, and the warnings.
4. Run the tests until they pass.

### Files

- `internal/slice/status.go`: new.
- `internal/slice/status_test.go`: new.
- `internal/cli/status.go`: new.
- `e2e/slice_status_test.go`: new.

### Interfaces

- `type Status struct{ Slice, Title, Phase string; From *eventlog.From; Created time.Time; Warnings []Warning }`:
  creates. `Title` is empty when none was saved; `Created` is zero when
  the log has none. Task 10 uses it.
- `func Status(ctx context.Context, r Repo, name string) (Status, error)`:
  creates.
- `func readSlice(ctx context.Context, r Repo, name, base string, baseExists bool) (Status, bool, error)`:
  creates. The bool is false when the slice does not exist. Task 10
  uses it.
- `func baseWarning(ctx context.Context, r Repo) (base string, w *Warning, err error)`:
  creates. R21 and R23. Task 10 uses it.
- `eventlog.Parse`, `eventlog.Phase`, `git.Show`, `printWarnings`:
  uses.

### Done when

- `TestStatusSpecSubmit` (e2e): S13 with its values, the created time
  read from the log.
  Proves: R24, R25, R29
  Scenario: S13
- `TestStatusCurrentBranch` (e2e): S14.
  Proves: R31
  Scenario: S14
- `TestStatusNotOnSlice` (e2e): S15's three lines, exit code 1; on a
  detached HEAD, `on no branch`.
  Proves: R32, R38
  Scenario: S15
- `TestStatusJSON` (e2e): S16.
  Proves: R30
  Scenario: S16
- `TestStatusBadLines` (e2e): S17, with the log changed by a commit on
  the slice branch.
  Proves: R26, R28
  Scenario: S17
- `TestStatusNoLog` (e2e): S18.
  Proves: R25, R26, R29
  Scenario: S18
- `TestStatusNoSuchSlice` (e2e): S19; also `slice status ../x` gives
  the same error for `../x`.
  Proves: R33
  Scenario: S19
- `TestStatusReadsCommitted` (unit): an uncommitted change to the log in
  the worktree is not read.
  Proves: R24
- `TestStatusMergedFromBase` (unit): with no local branch and the folder
  on `main`, the status is read from `main`.
  Proves: R24
- `TestStatusBaseMissing` (unit): with no `main` and the default base,
  `Warnings` holds
  `base branch main does not exist (the default); merged slices are not shown`
  with `Slice` empty.
  Proves: R21, R23
- `TestStatusOutsideRepo` (unit): an empty folder returns R12's error.
  Proves: R22

### Out of scope

- The list of slices. Done in: task 10

## Task 10: slice-list

```yaml
covers: [R21, R22, R23, R28, R34, R35, R36, R37, R38, R43]
depends_on: ["09"]
touches: [internal/slice/list.go, internal/slice/list_test.go,
  internal/cli/list.go, e2e/slice_list_test.go]
follows: [Flow 5, Core API, CLI, Testing]
risk: normal
```

### Goal

`sliceflow slice list` shows the slices in progress, oldest first, or
all of them with `--all`, or R37's JSON, with ` !` for a slice with
warnings.

### Steps

1. Write the tests listed under Done when. Run them, and see them fail.
2. `internal/slice/list.go`: `List` type and function, as Flow 5 says.
3. `internal/cli/list.go`: register `slice list` with `--all` and
   `--json`; print the table as R34 says, or R37's JSON.
4. Run the tests until they pass.

### Files

- `internal/slice/list.go`: new.
- `internal/slice/list_test.go`: new.
- `internal/cli/list.go`: new.
- `e2e/slice_list_test.go`: new.

### Interfaces

- `type List struct{ Slices []Status; Warnings []Warning }`: creates.
- `func List(ctx context.Context, r Repo, all bool) (List, error)`:
  creates.
- `readSlice`, `baseWarning`, `Status`, `printWarnings`: uses.

### Done when

- `TestListInProgress` (e2e): S20 with its values; created times set by
  committing logs with S20's `at` values on each branch.
  Proves: R34, R35
  Scenario: S20
- `TestListAll` (e2e): S21.
  Proves: R35
  Scenario: S21
- `TestListBadLine` (e2e): S22.
  Proves: R36
  Scenario: S22
- `TestListJSON` (e2e): S23.
  Proves: R37
  Scenario: S23
- `TestListEmpty` (e2e): S24; with `--all`, `No slices.`; with
  `--json`, `{"slices":[],"warnings":[]}`.
  Proves: R34, R35, R37
  Scenario: S24
- `TestListBaseMissing` (e2e): S25.
  Proves: R21, R23, R28
  Scenario: S25
- `TestListUnknownFlag` (e2e): S27.
  Proves: R38
  Scenario: S27
- `TestListNotARepo` (e2e): S28.
  Proves: R22
  Scenario: S28
- `TestListBadBranchName` (e2e): branches `slice/My_Thing` and
  `slice/team/login` are not rows; stderr has
  `Warning: branch slice/My_Thing: not a valid slice name; skipped` and
  the same for `slice/team/login`; no row gets ` !`; `--json` has both
  in the top `warnings`.
  Proves: R36, R43
- `TestListOrder` (unit): two slices with the same created time are a
  to z; a slice with no created time is last.
  Proves: R34

### Out of scope

- Teammates' slices from remote branches. Done in: not in this slice

## Task 11: no-network-check

```yaml
covers: [R40]
depends_on: ["08", "10"]
touches: [e2e/no_network_check_test.go]
follows: [No network, Testing]
risk: normal
```

### Goal

One e2e test proves that no command of this slice runs `git fetch`,
`git pull`, or `git push`, also when a remote is set up.

### Steps

1. Write the test listed under Done when. It should pass at once; to
   see it fail, check that it finds a `fetch` when the test itself runs
   `git fetch` with the same trace file, then remove that line.
2. Run the test until it passes.

### Files

- `e2e/no_network_check_test.go`: new.

### Interfaces

- `run`, `testrepo.New`, `testrepo.Git`: uses.

### Done when

- `TestNoNetwork` (e2e): a repo with a remote `origin` at a folder that
  does not exist; with `GIT_TRACE` set to a file, run `slice create a`,
  `slice create b --json`, `slice status a`, `slice list --all`: every
  exit code is 0, and the trace has no `git fetch`, `git pull`,
  `git push`, or `remote-` line.
  Proves: R40

### Out of scope

- Network calls from the user's own git hooks. Done in: not in this
  slice
