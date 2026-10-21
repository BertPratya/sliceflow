# AGENTS.md

flow is a command-line tool that runs a spec-driven workflow with AI coding
agents. It starts workers and reviewers, runs the checks, merges tasks, and
keeps each slice's status. It runs on Windows, macOS, and Linux. Go, in
`cmd/flow/` and `internal/`.

## How we work

Pick one at the start. If unsure, ask the human.

- **Brainstorm:** the human wants to think, not build.
- **Slice:** new behavior, a contract or data change, or a design decision.
- **Small change:** a clear bug fix, a text tweak, a small refactor, a
  config or test fix. A `fix/<short-name>` branch, the checks, then merge.
  If it grows into a feature, stop and suggest a slice.

If you are the main chat agent, you are the lead. Before the work below,
read its skill and follow it:

| When | Read |
|---|---|
| The human wants to explore an idea | `.agents/skills/brainstorming/SKILL.md` |
| The human asks for a new slice | `.agents/skills/flow-slice/SKILL.md` |
| A slice is in `planning` | `.agents/skills/flow-planning/SKILL.md` |
| A slice is `ready` or `implementing` | `.agents/skills/flow-implementation/SKILL.md` |
| A slice is in `integration` or `preview` | `.agents/skills/flow-finish/SKILL.md` |

A slice lives in `docs/slices/<name>/`, on branch `slice/<name>`, in
worktree `.worktrees/<name>/`. Its state is in its `status.md`.
Phases: `planning → ready → implementing → integration → preview → done`,
with the marks `(blocked)` and `(cancelled)`. Templates: `docs/templates/`.
Their comments hold each document's rules. The design and its decisions:
`docs/design/flow-design.md`.

## Agents

The lead starts each agent in a fresh context. Its prompt says to read its
prompt file and follow it, and names the worktree and the target.

| Agent | Prompt file | Tool | How to start (Claude Code) |
|---|---|---|---|
| Worker | `.agents/prompts/worker.md` | Claude Code, Sonnet 5.5, high effort | Agent tool, type `flow-worker`, model `sonnet` |
| Code reviewer | `.agents/prompts/code-reviewer.md` | Claude Code, Opus 5.5, high effort | Agent tool, type `flow-code-reviewer`, model `opus` |
| Doc reviewer | `.agents/prompts/doc-reviewer.md` | Claude Code, Opus 5.5, high effort | Agent tool, type `flow-doc-reviewer`, model `opus` |

If a type is not loaded, use `general-purpose`: the prompt names the file.

## Hard rules

- Only the human approves. Never approve for them, even if asked to "just
  continue".
- Never change an approved document without the human. Their answer to a
  question counts, as `flow-implementation` says.
- Never move or delete an approval tag. Never edit a submitted review.
- Only the lead changes `status.md`.
- Never force-push. Never rewrite shared history.
- Merge or push only as the git policy says.

## Stop rules

Stop, mark the task or slice `blocked`, and tell the human when:

- the checks fail 3 times since the last review
- a review does not pass after **3 rounds** (the review round limit)
- something is unclear, or a document looks wrong
- you need a new dependency, a storage format change, secrets, or CI changes

## Checks

Run in the repo root (or the task's worktree). Only a full run counts.

1. Auto-fix: `gofmt -w cmd internal`
2. Checks:
   - `gofmt -l cmd internal` prints nothing
   - `go vet ./...`
   - `go test ./...`
   - `go build ./...`

Run locally, for the preview: `go run ./cmd/flow <command>` in the slice
worktree, against a temporary test repo. Tests create their own temporary
git repos; never use shared services.

Always allowed in any task: `go.sum`, when `go.mod` did not change.

Setup commands: none. Workflow files (Markdown) have no automatic checks;
reviewers read them by hand.

Hooks and CI: `.githooks/pre-commit` runs the checks for Go commits (turn on
once per clone: `git config core.hooksPath .githooks`). CI runs the checks
on Linux, Windows, and macOS for every push to `main` and every pull request.

## Git

- Base branch `main`. `origin` is private.
- Slice `slice/<name>`, task `task/<name>-NN`, update `update/<name>`, small
  change `fix/<short-name>`. Worktrees in `.worktrees/`.
- Slice, task, and update work happens in its worktree, with full paths.
  The main checkout is the human's.
- Commit after every step: `<name>: <what happened>`; task work
  `<name>/NN: <summary>`.
- Merge into `main` with `git merge --no-ff`, only after the human
  approves, one slice at a time. Then push to `origin`. Tags go with it:
  `git config push.followTags true` once per clone.

## Project rules

- Everyone: plain, simple English. Short sentences.
- Everyone: keep everything generic: no company names, internal URLs, or
  secrets.
- Lead: until flow's first release, run this repo with the skills above.
  Use the `flow` binary only in tests, on temporary repos.
- Lead, planning: shared Go logic goes in `internal/`, not `cmd/`.
- Worker: every change that adds behavior adds a test.
- Worker: `flow` sends no data to any outside service: no telemetry, no
  uploads, no calls home. It reads and writes repo files, and runs local
  programs (git, the checks, the agent tools).
- Worker: run each check or git command as its own process with its own
  arguments, never as one shell string.
- Worker: paths, line endings, and processes must work on Windows, macOS,
  and Linux.
