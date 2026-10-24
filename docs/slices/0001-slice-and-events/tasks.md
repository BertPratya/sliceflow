---
slice: 0001-slice-and-events
---

# Tasks

## Task 01: git-runner

```yaml
covers: [R40]
depends_on: []
touches: [go.mod, internal/git/, internal/testrepo/]
follows: [TD1, TD3, TD8, Git arguments, No network, Tests, Testing]
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
   as its working directory, `GIT_TERMINAL_PROMPT=0` and
   `GIT_NO_LAZY_FETCH=1` added to the environment, and
   `cmd.WaitDelay` of 2 seconds, so a hook that outlives a stopped git
   cannot hold `Run` open. When `ctx` is done,
   return `ctx.Err()`, not a `*git.Error`.
4. Add `Error`: the arguments, the exit code, and git's stderr with
   spaces trimmed at both ends. Its `Error()` is that stderr text.
5. Add `RefExists` (`git rev-parse --verify --quiet <ref>^{commit}`)
   and `Show` (`git cat-file -e`, then `git show <ref>:<path>`). Both
   pass on `ctx.Err()`.
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
  creates. Returns stdout. A failed command returns a `*git.Error`; a
  done `ctx` returns `ctx.Err()`. Every later task uses it.
- `type Error struct{ Args []string; ExitCode int; Stderr string }`:
  creates. Task 09 builds R18's `<reason>` from it.
- `func RefExists(ctx context.Context, dir, ref string) (bool, error)`:
  creates. Tasks 04, 06, 08, and 10 use it.
- `func Show(ctx context.Context, dir, ref, path string) ([]byte, bool, error)`:
  creates. The bool is false when the file is not in that commit.
  Task 10 uses it.
- `func Setup(t testing.TB)`: creates. Sets `GIT_CONFIG_GLOBAL` to a
  temporary file with `user.name`, `user.email`, and
  `init.defaultBranch main`, and `GIT_CONFIG_NOSYSTEM=1`, with
  `t.Setenv`. Tasks 02 to 13 use it.
- `func New(t testing.TB) string`: creates. Calls `Setup`, makes a repo
  in `t.TempDir()` on `main` with one empty commit `first`, and returns
  its folder. Tasks 02 to 13 use it.
- `func Git(t testing.TB, dir string, args ...string) string`: creates.
  Runs git in `dir`, fails the test on an error, returns stdout. Tasks
  02 to 13 use it.

### Done when

- `TestRunReturnsOutput`: in a `testrepo.New` repo,
  `Run(ctx, dir, "rev-parse", "--abbrev-ref", "HEAD")` returns `main\n`.
  Proves: none (the runner itself)
- `TestRunErrorHasStderr`: `Run(ctx, dir, "rev-parse", "--verify", "nope")`
  returns a `*git.Error` with `ExitCode` 128, and a `Stderr` that is
  not empty and has no trailing newline.
  Proves: none (the runner itself)
- `TestRunSetsNoNetworkEnv`: in a repo whose local config has
  `alias.env-check` set to `!env`, the output of
  `Run(ctx, dir, "env-check")` holds the lines `GIT_TERMINAL_PROMPT=0`
  and `GIT_NO_LAZY_FETCH=1`.
  Proves: R40
- `TestRunCancelled`: `Run` with a cancelled `ctx` returns
  `context.Canceled`; `RefExists` with it returns `false` and
  `context.Canceled`, not `false, nil`.
  Proves: none (the runner itself)
- `TestRefExists`: `refs/heads/main` is true; `refs/heads/slice/none` is
  false with no error.
  Proves: none (the runner itself)
- `TestShowMissingFile`: `Show(ctx, dir, "main", "docs/none.md")`
  returns `nil, false, nil`; a committed `a.txt` with `hi\n` returns
  `hi\n, true, nil`.
  Proves: none (the runner itself)

### Out of scope

- Finding the repo and the main checkout. Done in: task 04
- The test that no `fetch`, `pull`, or `push` runs. Done in: task 13

## Task 02: event-log

```yaml
covers: [R19, R25, R26, R27]
depends_on: ["01"]
touches: [internal/eventlog/]
follows: [Events in the code, Reading a line, TD4, TD5, Tests, Testing]
risk: core
```

### Goal

The event log format exists as a package with no file or git access.
It writes one `slice.created` line, and reads a whole log into events
and warnings in R26's and R27's exact words. The hardest part is the
reader: every bad line is skipped with the right reason, including a
field that appears twice. Its brief is about 2,100 words, over the usual 1,500, because it carries every warning text of R26 and R27, which the worker must copy exactly.

### Steps

1. Write the tests listed under Done when. Run them, and see them fail.
2. In `internal/eventlog/event.go`, add `Event`, `Created`, `From`, and
   `NewCreated`, with the fields declared in R19's order (TD4).
3. Add `Encode`: one line of JSON, ending in `\n`.
4. In `internal/eventlog/parse.go`, add `Parse` as Reading a line says:
   split on `\n`; for each line, check `blank line` and
   `not valid JSON`, then find a name that appears twice with
   `json.Decoder.Token`, then read the line into a map and check the
   other reasons in R26's order (TD5), as Reading a line says.
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
  Tasks 06 and 10 use it.
- `type Created struct{ From From; Title string }`: creates, JSON
  `from` and `title,omitempty`. Tasks 06 and 10 use it.
- `type From struct{ Branch string; Commit string }`: creates, JSON
  `branch` and `commit`. Tasks 06 and 10 use it.
- `func NewCreated(at time.Time, from From, title string) Event`:
  creates. `Version` 1, `Seq` 1, `At` in UTC as
  `2006-01-02T15:04:05Z`. Task 06 uses it.
- `func Encode(e Event) ([]byte, error)`: creates. Task 06 uses it.
- `func Parse(log []byte) ([]Event, []string)`: creates. The warnings
  are R26's and R27's text, e.g. `.events.jsonl line 2: not valid JSON; skipped`.
  Task 10 uses it.
- `func Phase(events []Event) string`: creates. `planning` or
  `unknown`. Task 10 uses it.

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
- `TestParseFirstFaultWins`: a line with an extra `"by":"x"` and no `at`
  gives only `unknown field "by"`.
  Proves: R26
- `TestParseSeqGap`: lines with `seq` 1, 2, 4, 5 give all four events
  and one warning `.events.jsonl line 3: seq 4, expected 3`.
  Proves: R27
- `TestPhaseUnknown`: `Phase(nil)` is `unknown`.
  Proves: R25

### Out of scope

- Reading the log from git, and the `.events.jsonl not found` warning.
  Done in: task 10
- Adding a second event to a log. Done in: not in this slice

## Task 03: cli-skeleton

```yaml
covers: [R28, R38, R39]
depends_on: ["01", "04"]
touches: [cmd/sliceflow/, internal/cli/cli.go, internal/cli/flags.go,
  internal/cli/output.go, internal/cli/cli_test.go,
  internal/cli/flags_test.go, internal/e2e/helpers_test.go,
  internal/e2e/cli_skeleton_test.go]
follows: [TD2, Shared logic in internal, "Flow 6: Help and wrong use", CLI, Tests, Testing]
risk: core
```

### Goal

`sliceflow` runs, shows help, and stops wrong use with exit code 2 and
R38's words. Commands register themselves from their own files, so
tasks 07, 11, and 12 each add a file and never edit `cli.go`. The e2e
helpers exist. The hardest part is the flag helper of TD2. Its brief is about 1,900 words, over the usual 1,500, because R38 brings the wrong-use scenarios, whose texts the CLI must print exactly.

### Steps

1. Write the tests listed under Done when. Run them, and see them fail.
2. `cmd/sliceflow/main.go`: `os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr))`.
3. `internal/cli/cli.go`: `Run`, `command`, `register`. `Run` finds
   `slice <name>` in the registry. `sliceflow` alone, `sliceflow slice`
   alone, `--help`, and `-h` print the list of commands on stdout and
   return 0.
4. `internal/cli/flags.go`: `parseFlags` (TD2), with R38's texts. `-h`
   and `--help` print the command's usage on stdout and return
   `errHelp`.
5. `internal/cli/output.go`: `usageError`, `fail`, `printWarnings`.
   `fail` prints the first line of the message after `Error: `, and
   each further line indented by 2 spaces.
6. `internal/e2e/helpers_test.go`: `TestMain` builds the binary into a
   temporary folder with `go build`; `run` runs it.
7. Run the tests until they pass.

### Files

- `cmd/sliceflow/main.go`: new. Only the line of step 2.
- `internal/cli/cli.go`: new. `Run`, `command`, `register`, top usage.
- `internal/cli/flags.go`: new. `parseFlags`, `errHelp`.
- `internal/cli/output.go`: new. Error and warning output.
- `internal/cli/cli_test.go`, `internal/cli/flags_test.go`: new.
- `internal/e2e/helpers_test.go`: new. `TestMain`, `run`.
- `internal/e2e/cli_skeleton_test.go`: new. The e2e tests below.

### Interfaces

- `func Run(args []string, stdout, stderr io.Writer) int`: creates.
- `type command struct{ Name, Usage string; Run func(ctx context.Context, args []string, stdout, stderr io.Writer) int }`:
  creates. `Name` is e.g. `slice create`. Tasks 07, 11, 12 use it.
- `func register(c command)`: creates. Tasks 07, 11, 12 call it from
  `init()`.
- `func parseFlags(fs *flag.FlagSet, usage string, args []string, stdout io.Writer) ([]string, error)`:
  creates. Returns the positional arguments. Tasks 07, 11, 12 use it.
- `var errHelp`: creates. A command returns 0 on it.
- `func usageError(stderr io.Writer, msg string) int`: creates. Prints
  `Error: <msg>. Run sliceflow --help.`, returns 2. Tasks 07, 11, 12
  use it, e.g. with `missing slice name` and `too many arguments`.
- `func fail(stderr io.Writer, err error) int`: creates. Prints
  `Error: ` and the message, returns 1. Tasks 07, 09, 11, 12 use it.
- `func printWarnings(stderr io.Writer, ws []slice.Warning)`: creates.
  R28's form. Tasks 11 and 12 use it.
- `type slice.Warning`: uses, from task 04.
- `func run(t *testing.T, dir string, args ...string) (stdout, stderr string, code int)`:
  creates, in package `e2e`. Every later e2e test uses it.
- `testrepo.Setup`, `testrepo.New`: uses, from task 01.

### Done when

- `TestParseFlagsAfterName`: `login-google --title "Login with Google" --json`
  gives positional `[login-google]`, title `Login with Google`, json
  true; `--title=X a` gives `[a]` and `X`; `-- --x` gives `[--x]`.
  Proves: R38
- `TestParseFlagsWrongUse`: `--colour` returns an error whose text is
  `unknown flag --colour`; `--title` with no value returns
  `flag needs a value: --title`.
  Proves: R38
- `TestPrintWarnings`: a warning `{Slice: "a", Text: "x"}` prints
  `Warning: a: x`; `{Slice: "", Text: "y"}` prints `Warning: y`.
  Proves: R28
- `TestFailIndentsLines`: `fail` with the message `a\nb\nc` prints
  `Error: a`, `  b`, `  c`, and returns 1.
  Proves: R38
- `TestHelpTop`: `sliceflow --help`, `sliceflow` alone, and
  `sliceflow slice` alone each print a usage that names `slice create`,
  `slice status`, and `slice list` on stdout, nothing on stderr, exit
  code 0.
  Proves: R38, R39
- `TestUnknownCommand`: `sliceflow slice nope` prints
  `Error: unknown command "slice nope". Run sliceflow --help.` on
  stderr, nothing on stdout, exit code 2.
  Proves: R38

### Out of scope

- The usage text of each command, and its flags. Done in: task 07,
  task 11, task 12
- Catching Ctrl-C. Done in: task 09

## Task 04: repo-and-names

```yaml
covers: [R2, R3, R12, R13, R14, R21, R22]
depends_on: ["01"]
touches: [internal/slice/repo.go, internal/slice/repo_test.go,
  internal/slice/errors.go, internal/slice/name.go,
  internal/slice/name_test.go, internal/slice/base.go,
  internal/slice/base_test.go]
follows: [Core API, Git arguments, Cross-platform paths, Tests, Testing]
risk: core
```

### Goal

The shared base of the core exists: finding the repo and the main
checkout, the error and warning types, the name rule, and picking the
base branch. Every command uses these. Its brief is about 2,150 words, over the usual 1,500, because R3 and R21 bring the create and list scenarios that show the base branch; the task needs only their base-branch part.

### Steps

1. Write the tests listed under Done when. Run them, and see them fail.
2. `internal/slice/errors.go`: `Error`, `Warning`.
3. `internal/slice/repo.go`: `Repo`, `FindRepo`: the folder with
   `git rev-parse --show-toplevel`, the main checkout's root as the
   first `worktree` line of `git worktree list --porcelain`, and R13's
   check, `git for-each-ref --count=1 refs/heads/` prints nothing.
   `Root` uses `/` (Cross-platform paths).
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
- `type Warning struct{ Slice, Text string }`: creates. Tasks 03, 10,
  12 use it.
- `type Repo struct{ Dir, Root string; Now func() time.Time }`:
  creates. `Dir` is where the command runs; `Root` is the main
  checkout's root. Tasks 06 to 12 use it.
- `func FindRepo(ctx context.Context, dir string) (Repo, error)`:
  creates. Returns R12's or R13's `*Error`. `Now` is `time.Now`.
  Tasks 07, 11, 12 call it from the CLI.
- `func CheckName(name string) error`: creates. Tasks 08 and 10 use it.
- `func BaseBranch(ctx context.Context, r Repo, from string) (branch, source string, err error)`:
  creates. `from` is `--from` or empty. `source` is `from --from`,
  `from git config sliceflow.base`, or `the default`. It does not check
  that the branch exists. Tasks 06, 08, 10 use it.
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

- Checking that the base branch exists (R16). Done in: task 08
- The R23 warning. Done in: task 10

## Task 05: spec-template

```yaml
covers: [R7]
depends_on: ["01"]
touches: [internal/template/]
follows: [Cross-platform paths, Tests, Testing]
risk: normal
```

### Goal

The spec template is built into the binary with `embed`, with its
front matter filled for a new slice.

### Steps

1. Write the tests listed under Done when. Run them, and see them fail.
2. Copy `docs/templates/spec.md` to `internal/template/spec.md`. Remove
   the `title:` line of the front matter, and the template's rules
   about the title in its comments.
3. In `internal/template/template.go`, embed it, and add `Spec`: it
   fills `{{slice}}` and `{{date}}` in the front matter, and sets
   `ideas: []`.
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
  starts with exactly
  `---\nslice: spec-submit\ncreated: 2026-10-25\nideas: []\n---\n`,
  so no other front matter line, such as the template's example idea
  line, is left in.
  Proves: R7
- `TestSpecNoPlaceholdersInFrontMatter`: the front matter has no line
  starting with `title:`, and no `{{slice}}`, `{{title}}`, or
  `{{date}}`.
  Proves: R7
- `TestSpecBody`: the output after its front matter is the same as the
  embedded `internal/template/spec.md` after its front matter.
  Proves: R7
- `TestSpecLineEndings`: the output has no `\r`.
  Proves: R7

### Out of scope

- Templates a project can change. Done in: not in this slice

## Task 06: create-core

```yaml
covers: [R1, R4, R5, R6, R7, R8, R19, R20]
depends_on: ["02", "04", "05"]
touches: [internal/slice/create.go, internal/slice/create_test.go]
follows: ["Flow 1: Create a slice", Git arguments, Cross-platform paths, Tests, Testing]
risk: core
```

### Goal

`slice.Create` makes the branch, the worktree, the two files, and the
commit, from a `Repo` the CLI found. The checks before creating come in
task 08, the undo in task 09; this task leaves a clear place for the
checks. Its brief is about 2,000 words, over the usual 1,500, because Flow 1 and the create scenarios hold the exact files, commit, and log line.

### Steps

1. Write the tests listed under Done when. Run them, and see them fail.
2. `internal/slice/create.go`: `CreateOptions`, `Created`, `Create`, as
   Flow 1 steps 4 to 8 say. Call `checkCreate` right after picking the
   base; in this task it only returns nil.
3. Run the tests until they pass.

### Files

- `internal/slice/create.go`: new. `Create` and its types, and a
  `checkCreate` that returns nil.
- `internal/slice/create_test.go`: new.

### Interfaces

- `type CreateOptions struct{ Name, Title, From string }`: creates.
  Task 07 uses it.
- `type Created struct{ Slice, Branch string; From eventlog.From; Worktree, Spec string }`:
  creates. `Worktree` is the full path with `/`; `Spec` is the path
  inside the worktree. Task 07 prints it.
- `func Create(ctx context.Context, r Repo, opts CreateOptions) (Created, error)`:
  creates. Task 07 calls it; task 09 adds the undo.
- `func checkCreate(ctx context.Context, r Repo, opts CreateOptions, base, source string) error`:
  creates, returning nil. Task 08 fills it.
- `eventlog.NewCreated`, `eventlog.Encode`, `template.Spec`,
  `BaseBranch`, `git.Run`: uses.

### Done when

- `TestCreateMakesSlice`: in a `testrepo.New` repo with a fixed
  `Repo.Now` of 2026-10-24 07:42:00 UTC, `Create` with name
  `spec-submit` makes branch `slice/spec-submit` at `main`'s commit plus
  one commit `spec-submit: slice created` with exactly
  `docs/slices/spec-submit/spec.md` and
  `docs/slices/spec-submit/.events.jsonl`; the log is the one line of
  `eventlog.NewCreated` for that time and commit; the worktree is
  `<Root>/.sliceflow/worktrees/spec-submit`; the main checkout's
  `git status --short` is empty.
  Proves: R1, R4, R5, R7, R8, R19, R20
- `TestCreateWritesGitignore`: `.sliceflow/.gitignore` is `*\n`; an
  existing one with `x\n` is not changed.
  Proves: R6
- `TestCreateTitle`: with title `Login with Google`, the log's `data`
  has that `title`; without one, `data` has no `title`.
  Proves: R20
- `TestCreateHookRuns`: a `pre-commit` hook that writes a file `ran` in
  the repo's git folder runs once.
  Proves: R8
- `TestCreateIgnoredFiles`: with `*.jsonl` in the repo's `.gitignore`,
  `Create` still commits both files.
  Proves: R7

### Out of scope

- The CLI command and its output. Done in: task 07
- The checks before creating (R11 to R17). Done in: task 08
- The undo and Ctrl-C (R18). Done in: task 09

## Task 07: create-command

```yaml
covers: [R1, R3, R5, R9, R10, R38, R39, R41]
depends_on: ["03", "06"]
touches: [internal/cli/create.go, internal/e2e/create_command_test.go]
follows: [CLI, Cross-platform paths, Tests, Testing]
risk: normal
```

### Goal

`sliceflow slice create <name>` works from the terminal, and prints
R9's lines or R10's JSON, with paths relative to the main checkout. Its brief is about 1,850 words, over the usual 1,500, because the create scenarios hold the exact output the command prints.

### Steps

1. Write the tests listed under Done when. Run them, and see them fail.
2. `internal/cli/create.go`: register `slice create` with `--title`,
   `--from`, `--json`, and its usage. Wrong use: no name gives
   `missing slice name`, two names give `too many arguments`. Call
   `FindRepo`, then `Create`. Print R9's lines with `Worktree:` and
   `Next:` relative to `Repo.Root`, or R10's JSON.
3. Run the tests until they pass.

### Files

- `internal/cli/create.go`: new. The `slice create` command.
- `internal/e2e/create_command_test.go`: new.

### Interfaces

- `FindRepo`, `Create`, `CreateOptions`, `Created`, `register`,
  `parseFlags`, `usageError`, `fail`, `run`: uses.

### Done when

- `TestCreateSpecSubmit`: S1 with its exact values: a repo on `main`,
  `sliceflow slice create spec-submit` prints S1's five lines with the
  real short commit, exit code 0; `.sliceflow/.gitignore` is `*\n`; the
  branch has the one commit with exactly the two files; the log is one
  line like S1's.
  Proves: R1, R9
  Scenario: S1
- `TestCreateJSON`: S2: with `sliceflow.base develop`,
  `slice create login-google --title "Login with Google" --json` prints
  one JSON object with `slice`, `branch`, `from.branch` `develop`, the
  full commit, `worktree` (full path with `/`), and `spec`.
  Proves: R3, R10, R41
  Scenario: S2
- `TestCreateFromFlag`: S3: with `sliceflow.base develop`,
  `slice create hotfix-ui --from main` starts from `main`.
  Proves: R3
  Scenario: S3
- `TestCreateInsideWorktree`: S4: run inside another slice's worktree,
  the new worktree is `.sliceflow/worktrees/spec-submit` of the main
  checkout, and the printed paths are as in S1.
  Proves: R5, R9
  Scenario: S4
- `TestCreateHelp`: S26: `slice create --help` prints a usage with
  `--title`, `--from`, `--json` on stdout, exit code 0.
  Proves: R39
  Scenario: S26
- `TestCreateWrongUse`: `slice create` prints
  `Error: missing slice name. Run sliceflow --help.`; `slice create a b`
  prints `Error: too many arguments. Run sliceflow --help.`; both exit
  with 2 and print nothing on stdout.
  Proves: R38

### Out of scope

- The checks before creating. Done in: task 08
- Catching Ctrl-C. Done in: task 09

## Task 08: create-checks

```yaml
covers: [R2, R11, R12, R13, R14, R15, R16, R17, R42]
depends_on: ["07"]
touches: [internal/slice/create.go, internal/slice/check.go,
  internal/slice/check_test.go, internal/e2e/create_checks_test.go]
follows: ["Flow 2: A check fails before creating", Git arguments, Cross-platform paths, Tests, Testing]
risk: core
```

### Goal

`create` runs every check of R11, in R11's order, before it makes
anything, and stops at the first that fails with the spec's words. This
brief is longer than most because it carries the exact error text of
every check, which the worker must copy as written.

### Steps

1. Write the tests listed under Done when. Run them, and see them fail.
2. `internal/slice/check.go`: `checkCreate` with R14, R16, R15, R17, as
   Flow 2 says. R12 and R13 come from `FindRepo`. Remote names come
   from `git remote`.
3. In `internal/slice/create.go`, remove the stub of `checkCreate`.
4. Run the tests until they pass.

### Files

- `internal/slice/check.go`: new. `checkCreate`.
- `internal/slice/create.go`: changed. The stub is removed.
- `internal/slice/check_test.go`: new.
- `internal/e2e/create_checks_test.go`: new.

### Interfaces

- `func checkCreate(ctx context.Context, r Repo, opts CreateOptions, base, source string) error`:
  changes. Task 06 created it.
- `CheckName`, `git.RefExists`, `git.Run`, `run`: uses.

### Done when

Each e2e test also checks that stdout is empty, the exit code is 1,
and no branch, worktree, or `.sliceflow/` was made.

- `TestCreateBadName`: S5: `slice create "My Slice"` prints S5's error.
  Proves: R2, R14
  Scenario: S5
- `TestCreateReservedName`: S6: `slice create con`.
  Proves: R14
  Scenario: S6
- `TestCreateRemoteExists`: S7: `refs/remotes/origin/slice/spec-submit`
  made with `git update-ref`, and a remote `origin` in the config, gives
  `Error: slice spec-submit already exists: branch origin/slice/spec-submit. Nothing was changed.`
  Proves: R15
  Scenario: S7
- `TestCreateBaseMissing`: S8.
  Proves: R16
  Scenario: S8
- `TestCreateFolderExists`: S9.
  Proves: R17
  Scenario: S9
- `TestCreateNotARepo`: S10.
  Proves: R12
  Scenario: S10
- `TestCreateNoCommits`: S11.
  Proves: R13
  Scenario: S11
- `TestCreateBadNameJSON`: S29.
  Proves: R42
  Scenario: S29
- `TestCheckExistsOrder` (unit): with a remote `team/origin` in the
  config, a local branch `slice/x`, a ref
  `refs/remotes/team/origin/slice/x`, and `docs/slices/x` on the base:
  `<where>` is `branch slice/x`; without the local branch it is
  `branch team/origin/slice/x`; with only the folder it is
  `folder docs/slices/x on main`.
  Proves: R15
- `TestCheckOtherRemoteBranch` (unit): a ref
  `refs/remotes/origin/user/slice/x` alone does not block `x`.
  Proves: R15
- `TestCheckOrder` (unit): a bad name and a missing base give the name
  error (R14 before R16); a missing base and an existing branch give
  the base error (R16 before R15).
  Proves: R11

### Out of scope

- Undo after a step fails, and Ctrl-C. Done in: task 09

## Task 09: create-undo

```yaml
covers: [R8, R18]
depends_on: ["08"]
touches: [internal/slice/create.go, internal/slice/undo.go,
  internal/slice/undo_test.go, internal/cli/create.go,
  internal/e2e/create_undo_test.go]
follows: ["Flow 3: A step fails, or the user stops it", Undo on failure, Interrupts, Cross-platform paths, Tests, Testing]
risk: core
```

### Goal

When a step of `create` fails, or Ctrl-C or `SIGTERM` arrives,
sliceflow removes what it made in this run, newest first, and prints
R18's error with the right reason. The hardest part is to never remove
what this run did not make. Its brief is about 1,950 words, over the usual 1,500, because the undo and interrupt rules must all be followed together.

### Steps

1. Write the tests listed under Done when. Run them, and see them fail.
2. `internal/slice/undo.go`: the undo list, as Undo on failure says;
   it runs with `context.WithoutCancel(ctx)`.
3. In `internal/slice/create.go`: add each made thing to the list;
   note whether `.sliceflow/worktrees/` exists before step 6; check
   `ctx` between steps; on an error, run the undo and return R18's
   `*Error`. The reason is `interrupted` when `ctx.Err()` is set;
   else, for a git step, the first line of `git.Error.Stderr` that is
   not empty, or `git <command> failed (exit code <n>)` with the
   command name the step passes (e.g. `commit`, `worktree add`), with
   the other lines after it; else, for a step that is not git, the
   error text with paths written with `/`, as Flow 3, step 4 says.
4. In `internal/cli/create.go`: make the `ctx` with
   `signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)` first,
   before `FindRepo`. When `FindRepo` returns because the `ctx` is done,
   print R18's error with `interrupted`, not `context canceled`.
5. Run the tests until they pass.

### Files

- `internal/slice/undo.go`: new. The undo list.
- `internal/slice/create.go`: changed. Steps add to the list; errors
  undo.
- `internal/slice/undo_test.go`: new.
- `internal/cli/create.go`: changed. The signal `ctx`.
- `internal/e2e/create_undo_test.go`: new.

### Interfaces

- `Create`: changes. Errors after the checks are R18's `*Error`.
- `git.Error`: uses, from task 01.

### Done when

- `TestCreateHookRefuses` (e2e): S12: a `pre-commit` hook that only
  runs `exit 1`: stderr is
  `Error: could not create slice spec-submit: git commit failed (exit code 1). Everything sliceflow made was removed.`,
  stdout is empty, exit code 1; no branch `slice/spec-submit`, no
  worktree in `git worktree list --porcelain`, no `.sliceflow/`.
  Proves: R8, R18
  Scenario: S12
- `TestCreateHookOutput` (e2e): a `pre-commit` hook that prints
  `lint: 2 errors` and `a.go:1: bad` on stderr and exits with 1:
  the error line is
  `Error: could not create slice spec-submit: lint: 2 errors. Everything sliceflow made was removed.`
  and the next line is `  a.go:1: bad`.
  Proves: R18
- `TestCreateInterrupt` (e2e, skipped on Windows): a `pre-commit` hook
  that sleeps 30 seconds; after the hook starts, the test sends
  `os.Interrupt`: within 10 seconds, stderr is
  `Error: could not create slice spec-submit: interrupted. Everything sliceflow made was removed.`,
  exit code 1, nothing left.
  Proves: R18
- `TestUndoCancelledContext` (unit): `Create` with a `ctx` cancelled
  after the worktree step returns the `interrupted` error, and leaves
  no branch, worktree, or `.sliceflow/worktrees/`.
  Proves: R18
- `TestUndoCancelledDuringChecks` (unit): `Create` with a `ctx`
  cancelled before it starts returns the `interrupted` error, not R16's.
  Proves: R18
- `TestUndoKeepsOthers` (unit): with the checks skipped, a branch
  `slice/x` that already exists makes `git worktree add` fail; after
  the undo, `slice/x` still exists at its old commit.
  Proves: R18
- `TestUndoCouldNotRemove` (unit): when the worktree folder cannot be
  removed, the error ends with `Could not remove: <the folder>.`
  Proves: R18
- `TestReasonFileError` (unit): the reason for a `*fs.PathError` with
  `Op` `open`, `Path` `filepath.FromSlash("C:/a/b/spec.md")` (so it has
  the system's own separator), and `Err` `Access is denied.` is
  `open C:/a/b/spec.md: Access is denied`, with no `.` at the end, on
  every system.
  Proves: R18
- `TestReasonGitCommand` (unit): a failed `worktree add` with an empty
  stderr and exit code 128 gives `git worktree add failed (exit code 128)`.
  Proves: R18

### Out of scope

- Cleaning up after a hard kill. Done in: not in this slice

## Task 10: status-core

```yaml
covers: [R21, R23, R24, R25, R26, R33]
depends_on: ["02", "04", "06"]
touches: [internal/slice/status.go, internal/slice/status_test.go]
follows: ["Flow 4: Status of one slice", TD6, Cross-platform paths, Tests, Testing]
risk: core
```

### Goal

The core reads a slice's log from git and works out its status, with
its warnings. `Status` uses it for one slice; task 12's `List` uses the
same reading for every slice. Its brief is about 1,800 words, over the usual 1,500, because R26 and the status scenarios hold the exact warnings and values.

### Steps

1. Write the tests listed under Done when. Run them, and see them fail.
2. `internal/slice/status.go`: `SliceStatus`, `Status`, `readSlice`,
   and `baseWarning`, as Flow 4 says. Git paths are built with `path`.
3. Run the tests until they pass.

### Files

- `internal/slice/status.go`: new.
- `internal/slice/status_test.go`: new.

### Interfaces

- `type SliceStatus struct{ Slice, Title, Phase string; From *eventlog.From; Created time.Time; Warnings []Warning }`:
  creates. `Title` is empty when none was saved; `Created` is zero when
  the log has none. Tasks 11 and 12 use it.
- `func Status(ctx context.Context, r Repo, name string) (SliceStatus, error)`:
  creates. `name` empty means the current branch. Task 11 calls it.
- `func readSlice(ctx context.Context, r Repo, name, base string, baseExists bool) (SliceStatus, bool, error)`:
  creates. The bool is false when the slice does not exist. Task 12
  uses it.
- `func baseWarning(ctx context.Context, r Repo) (base string, exists bool, w *Warning, err error)`:
  creates. R21 and R23. Task 12 uses it.
- `eventlog.Parse`, `eventlog.Phase`, `git.Show`, `git.RefExists`,
  `CheckName`, `BaseBranch`, `Create` (in tests): uses.

### Done when

- `TestStatusReadsBranch`: after `Create` of `spec-submit` at a fixed
  time, `Status(ctx, r, "spec-submit")` has phase `planning`, `From`
  `main` and its commit, `Created` that time, no warnings.
  Proves: R24, R25
- `TestStatusReadsCommitted`: an uncommitted change to the log in the
  worktree is not read.
  Proves: R24
- `TestStatusMergedFromBase`: with no local branch and the folder on
  `main`, the status is read from `main`.
  Proves: R24
- `TestStatusNoLog`: a branch `slice/old-idea` with no log has phase
  `unknown` and the warning `.events.jsonl not found`.
  Proves: R25, R26
- `TestStatusBadName`: `Status(ctx, r, "../x")` returns
  `no slice named ../x. See all slices: sliceflow slice list`, and runs
  no git command with that name.
  Proves: R33
- `TestStatusBaseMissing`: with no `main` and the default base,
  `Warnings` holds
  `base branch main does not exist (the default); merged slices are not shown`
  with `Slice` empty.
  Proves: R21, R23

### Out of scope

- The `slice status` command and its output. Done in: task 11
- The list of slices. Done in: task 12

## Task 11: status-command

```yaml
covers: [R22, R26, R28, R29, R30, R31, R32, R33, R42]
depends_on: ["03", "07", "10"]
touches: [internal/cli/status.go, internal/e2e/status_command_test.go]
follows: [CLI, Tests, Testing]
risk: normal
```

### Goal

`sliceflow slice status [<name>]` prints R29's lines or R30's JSON, and
the warnings on stderr. Its brief is about 1,750 words, over the usual 1,500, because the status scenarios hold the exact output the command prints.

### Steps

1. Write the tests listed under Done when. Run them, and see them fail.
2. `internal/cli/status.go`: register `slice status` with `--json` and
   its usage; more than one name gives `too many arguments`. Call
   `FindRepo`, then `Status`. Print R29's lines or R30's JSON, and the
   warnings with `printWarnings`.
3. Run the tests until they pass.

### Files

- `internal/cli/status.go`: new.
- `internal/e2e/status_command_test.go`: new.

### Interfaces

- `FindRepo`, `Status`, `SliceStatus`, `register`, `parseFlags`,
  `usageError`, `fail`, `printWarnings`, `run`: uses.

### Done when

The slices in these tests are made with `sliceflow slice create`; a
changed log is committed on the slice branch.

- `TestStatusSpecSubmit`: S13 with its values, the created time read
  from the log.
  Proves: R29
  Scenario: S13
- `TestStatusCurrentBranch`: S14.
  Proves: R31
  Scenario: S14
- `TestStatusNotOnSlice`: S15's three lines, exit code 1; on a
  detached HEAD, `on no branch`.
  Proves: R32
  Scenario: S15
- `TestStatusJSON`: S16.
  Proves: R30
  Scenario: S16
- `TestStatusBadLines`: S17.
  Proves: R26, R28
  Scenario: S17
- `TestStatusNoLog`: S18.
  Proves: R26, R29
  Scenario: S18
- `TestStatusNoSuchSlice`: S19.
  Proves: R33
  Scenario: S19
- `TestStatusErrorJSON`: S19 with `--json`: stdout is empty, stderr is
  S19's error, exit code 1.
  Proves: R42
- `TestStatusNotARepo`: in an empty folder, R12's error, exit code 1.
  Proves: R22

### Out of scope

- The list of slices. Done in: task 12

## Task 12: slice-list

```yaml
covers: [R21, R22, R23, R28, R34, R35, R36, R37, R38, R42, R43]
depends_on: ["11"]
touches: [internal/slice/list.go, internal/slice/list_test.go,
  internal/cli/list.go, internal/e2e/slice_list_test.go]
follows: ["Flow 5: List of slices", CLI, Cross-platform paths, Tests, Testing]
risk: normal
```

### Goal

`sliceflow slice list` shows the slices in progress, oldest first, or
all of them with `--all`, or R37's JSON, with ` !` for a slice with
warnings, and leaves out branches whose name breaks R2. Its brief is about 2,350 words, over the usual 1,500, because the list scenarios hold the exact table, JSON, and warnings, and the core and the command share them.

### Steps

1. Write the tests listed under Done when. Run them, and see them fail.
2. `internal/slice/list.go`: `SliceList` and `List`, as Flow 5 says.
3. `internal/cli/list.go`: register `slice list` with `--all` and
   `--json`; any name gives `too many arguments`. Print the table as
   R34 says, or R37's JSON.
4. Run the tests until they pass.

### Files

- `internal/slice/list.go`: new.
- `internal/slice/list_test.go`: new.
- `internal/cli/list.go`: new.
- `internal/e2e/slice_list_test.go`: new.

### Interfaces

- `type SliceList struct{ Slices []SliceStatus; Warnings []Warning }`:
  creates.
- `func List(ctx context.Context, r Repo, all bool) (SliceList, error)`:
  creates.
- `readSlice`, `baseWarning`, `SliceStatus`, `FindRepo`,
  `printWarnings`, `run`: uses.

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
- `TestListErrorJSON` (e2e): S28 with `--json`: stdout is empty, stderr
  is S28's error, exit code 1.
  Proves: R42
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

## Task 13: no-network-check

```yaml
covers: [R40]
depends_on: ["09", "12"]
touches: [internal/e2e/no_network_check_test.go]
follows: [No network, Tests, Testing]
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

- `internal/e2e/no_network_check_test.go`: new.

### Interfaces

- `run`, `testrepo.New`, `testrepo.Git`: uses.

### Done when

- `TestNoNetwork`: a repo with a remote `origin` at a folder that does
  not exist; with `GIT_TRACE` set to a file, run `slice create a`,
  `slice create b --json`, `slice status a`, `slice list --all`: every
  exit code is 0, and the trace has no `git fetch`, `git pull`,
  `git push`, or `remote-` line.
  Proves: R40

### Out of scope

- Network calls from the user's own git hooks. Done in: not in this
  slice
