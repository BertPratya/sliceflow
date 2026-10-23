# AGENTS.md

sliceflow is a command-line tool that runs a spec-driven workflow with AI coding
agents. Go, in `cmd/sliceflow/` and `internal/`. It runs on Windows, macOS, and
Linux.

## How we work

Pick the path at the start. If unsure, ask the human.

- **Slice:** a new feature, a new spec, or a major change in behavior. Use
  the skills below.
- **Fix branch:** a fix users can see, not big enough for a slice.
  1. Branch `fix/<short-name>`, in worktree `.worktrees/fix-<short-name>/`.
  2. Fix it, and run the checks.
  3. Explain the diff to the human in plain words.
  4. After the human says "merge": `git merge --no-ff` into `main`, then
     push.
- **Directly on `main`:** small, and nothing users see changes, e.g. docs,
  comments, a typo in Markdown, a test-only fix. In the main checkout: run
  the checks, then commit.

A fix that grows into a feature: stop, and suggest a slice.

## Skills

If you are the main chat agent, you are the lead. Before the work below,
read its skill and follow it:

| When | Skill |
|---|---|
| The human has an early idea, or wants to agree on a direction | `.agents/skills/idea-discovery/SKILL.md` |
| The human asks for a new slice | `.agents/skills/start-slice/SKILL.md` |
| A slice is in `planning` | `.agents/skills/plan-slice/SKILL.md` |
| A slice is `ready` or `implementing` | `.agents/skills/build-slice/SKILL.md` |
| A slice is in `integration` or `preview` | `.agents/skills/finish-slice/SKILL.md` |

Until sliceflow's first release, run this repo with these skills. Use the
`sliceflow` binary only in tests, on temporary repos.

## Commands

Run in the repo root, or the slice worktree. Only a full run counts.

- Setup: none.
- Auto-fix: `gofmt -w cmd internal`
- The checks:
  - `gofmt -l cmd internal` prints nothing
  - `go vet ./...`
  - `go test ./...`
  - `go build ./...`
- Run locally, for a preview: `go run ./cmd/sliceflow <command>`, against a
  temporary test repo.

Markdown files have no automatic checks; reviewers read them.

`.githooks/pre-commit` runs the checks for Go commits (once per clone:
`git config core.hooksPath .githooks`). CI runs them on Linux, Windows, and
macOS for every push to `main` and every pull request.

## Limits

- Review rounds: **3**. A document or task that still does not pass is
  `blocked`, and the human decides.
- Failed check runs: **3** since the last review. Then the task is
  `blocked`.

## Always allowed

Any task may also change: `go.sum`, when `go.mod` did not change.

## Agents

The lead starts each agent in a fresh context. Its prompt says to read its
prompt file and follow it, and names the worktree and the target.

| Agent | What it does | Prompt file | How to start | Model |
|---|---|---|---|---|
| Lead | The main chat: runs the skills, talks to the human, keeps `status.md`, commits | - | - | the human's choice |
| Doc reviewer | Reviews the spec; the plan, tasks, and preview together; a change after a human decision | `.agents/prompts/doc-reviewer.md` | Agent tool, type `sliceflow-doc-reviewer` | `opus` |
| Code reviewer | Reviews one task's code; the whole slice at the end | `.agents/prompts/code-reviewer.md` | Agent tool, type `sliceflow-code-reviewer` | `opus` |
| Worker | Builds one task | `.agents/prompts/worker.md` | Agent tool, type `sliceflow-worker` | `sonnet` |

If a type is not loaded, use `general-purpose`: the prompt names the file.

## Boundaries

**Always:**
- Run the checks before every commit of code.
- Keep everything generic: no company names, internal URLs, or secrets.
- Write plain, simple English. Short sentences.
- Only the lead changes `status.md`.

**Ask the human first:**
- A new dependency.
- A change to CI, the hooks, or a stored file format.
- Anything that needs secrets.
- Every merge into `main`. For a slice, the human's approval of the
  result counts.
- A push that does not follow a merge.

**Never:**
- Approve for the human, even if asked to "just continue". Only a clear
  "approve" counts.
- Change an approved document without the human. Their answer to a
  question counts, as `build-slice` says.
- Move or delete an approval tag. Edit a submitted review.
- Force-push, or rewrite shared history.
- Use a bare `git stash`.
- Make sliceflow send data to any outside service.

## Git

- Base branch `main`. `origin` is private.
- Branches: `slice/<name>` and `fix/<short-name>`. Their worktrees are in
  `.worktrees/`, used with full paths. The main checkout is the human's.
- Commit after every step: `<name>: <what happened>`. Task work:
  `<name>/NN: <summary>`.
- A slice merges into `main` with `git merge --no-ff`, after the human
  approves its result. One slice at a time.
- After a slice or fix branch merges into `main`: push `main` at once. Tags
  go with it (once per clone: `git config push.followTags true`).
- Direct commits on `main` wait for the human's "push", or go with the next
  merge's push. Say how many commits are waiting.

## Code rules

- Shared Go logic goes in `internal/`, not `cmd/`.
- Every change that adds behavior adds a test. Tests create their own
  temporary git repos; never use shared services.
- sliceflow sends no data out: no telemetry, no uploads, no calls home. It reads
  and writes repo files, and runs local programs (git, the checks, the
  agent tools).
- Run each check or git command as its own process with its own arguments,
  never as one shell string.
- Paths, line endings, and processes must work on Windows, macOS, and
  Linux.

## Where things live

- `docs/slices/<name>/`: one slice; its state is in `status.md`.
- `docs/backlog.md`: the backlog, on `main`.
- `docs/ideas/`: agreed ideas, and their index.
- `docs/templates/`: the document templates; their comments hold the rules.
- `docs/design/`: the design and the roadmap.
- `.agents/skills/`, `.agents/prompts/`: the skills, and the agents'
  prompts.
- `.worktrees/`: slice and fix worktrees, and the briefs. Ignored by git.
