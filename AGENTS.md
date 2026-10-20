# AGENTS.md

## 1. How we work

There are two ways to work. Pick one at the start.

- **Feature → slice.** New behavior, API or contract change, data model change, auth/permissions/payments, frontend + backend together, or a design decision. Use the slice workflow below.
- **Small change → normal chat.** Clear bug fix, text/style tweak, small refactor in one area, config or test fix. Work on a branch, run the checks, and merge following the project's git policy (section 10). No slice.

If unsure, ask the human. If a small change grows into a feature, stop and suggest a slice.

## 2. Slices and phases

This repo is built with flow's own workflow, run by hand until flow's first
version: **manual flow**. The lead follows `workflow/manual-flow.md`.

A slice is one feature in `slices/<name>/`, on its own branch and worktree (section 10). Read, search, and edit its files in that worktree, not in the main checkout. Use full paths into the worktree: tools may skip the git-ignored `.worktrees/` folder when they search.

```
spec → plan → ready → implementing → integration → preview → done
                       marks on any phase: (blocked), (cancelled)
```

```
slices/<name>/
  spec.md        what the feature does              human approves
  plan.md        how it will be built               human approves, with tasks.md and preview.md
  tasks.md       every task: structure and content
  preview.md     how the human tries the slice
  status.md      the only state of the slice, kept by the lead
  reviews/       review files
  notes.md       technical ideas from the spec talk, optional, not approved
```

Templates: `templates/`. Each template's comments hold its rules; read the template before writing or changing a document.

Every approval is an annotated git tag, e.g. `flow/<name>/spec/R001`, holding the approval record.

Ideas for later go to `backlog.md` at the project root when a slice is done (`templates/backlog.md`).

Contracts, the data model, and other living docs stay at the project level (see section 9). A slice edits them on its own branch and lists them under `touches` in `plan.md`. The human approves the slice's change to them, not the whole file.

### Reviews

| Review | When |
|---|---|
| Spec check | every real gap as a question, before the human approves the spec |
| Plan review | plan, tasks, and preview together, always in full |
| Task review | every task with risk `core` or `normal`; not `low` |
| Integration review | always |

## 3. Roles

| Role | Does | Instructions |
|---|---|---|
| Human | decides, approves, previews | — |
| Lead | the main chat agent: designs with the human, writes docs, runs the slice, keeps the status, reports | `workflow/roles/lead.md`, `workflow/manual-flow.md` |
| Worker | implements exactly one task, from its brief | `workflow/roles/worker.md` |
| Plan reviewer | asks questions about the spec before approval; checks plan, tasks, and preview together before coding | `workflow/roles/plan-reviewer.md` |
| Code reviewer | checks one task, or the whole slice | `workflow/roles/code-reviewer.md` |

If you are the main chat agent, you are the lead. Read your role file before starting slice work.

## 4. Hard rules

- Only a human approves. Never approve for them, even if asked to "just continue".
- Never change an approved spec, plan, or contract without the human (`workflow/manual-flow.md` → "Changing an approved spec or plan"). If it looks wrong, stop and ask. The one exception: fix tasks and a wider `touches` in an approved `tasks.md`, as `workflow/manual-flow.md` → "Changes to an approved tasks.md" says.
- Never move or delete an approval tag. Never edit a submitted review.
- Only the lead changes status, and only in the slice's `status.md`. File headers describe what a file is; they hold no status.
- Merge or push only as the project's git policy allows (section 10), and for a slice only after the human approves the result.
- Never force-push. Never rewrite shared history.

## 5. Stop rules

Stop, mark the task (for a plan or integration review: the slice) `blocked`, and tell the human in plain words when:

- a task fails its checks 3 times since its last review
- a review (plan, task, or integration) still does not pass after 3 rounds
- the task is unclear, or the spec, plan, or contract looks wrong
- you need a new dependency, a storage format change, secrets, or CI changes

## 6. Status

- **Read:** the slice's `status.md`. It is the overview and the state of the whole slice: what needs the human, documents, tasks, findings, preview results, and the log. Format and the meaning of each status word: `templates/status.md`.
- **Write:** only the lead, after every step: `status.md` with a Log line, and a commit.
- **Checks:** a task moves to review or done only after the checks pass (section 9: the auto-fix commands, then the checks). Only a full run counts, not single tests.

## 7. Workflow files

| File | Holds |
|---|---|
| `workflow/manual-flow.md` | the lead's steps for every phase, until flow's first version |
| `workflow/roles/` | what each role does |
| `templates/` | one template per document; its comments hold the document's rules |
| `workflow/adapters/` | how each agent tool finds the roles |
| `docs/flow-design.md` | the design and the human's decisions behind the rules |

## 8. Project

flow is a command-line tool that runs a spec-driven workflow with AI coding agents. It starts workers and reviewers, runs the checks, merges tasks, and keeps each slice's status, so the lead does not do each step by hand. It works on Windows, macOS, and Linux.

The first slices build the backend: the orchestration pipeline.

## 9. Stack profile

Parts:

| Part | Folders | Tech |
|---|---|---|
| CLI | `cmd/flow/`, `internal/` | Go |
| Workflow files | `AGENTS.md`, `workflow/`, `templates/` | Markdown |

Connections: (none)

Tools per role (the lead is the tool you chat in):

| Role | Tool | How to start |
|---|---|---|
| Lead | Claude Code (Opus 5.5) | the chat |
| Worker | Claude Code (Sonnet 5.5, high effort) | subagent (Agent tool): type `flow-worker`, model `sonnet`, effort `high`, the prompt from `workflow/manual-flow.md` → 4. Implement. If the session did not start in this repo, `flow-worker` is not loaded: use type `general-purpose`; the prompt names the role file anyway. |
| Code reviewer | Claude Code (Opus 5.5, high effort) | subagent (Agent tool): type `flow-code-reviewer` or `flow-plan-reviewer`, model `opus`, effort `high`, the prompt from `workflow/manual-flow.md` → Reviews. Same fallback to `general-purpose`. |
| Plan reviewer | Claude Code (Opus 5.5, high effort) | subagent (Agent tool): type `flow-code-reviewer` or `flow-plan-reviewer`, model `opus`, effort `high`, the prompt from `workflow/manual-flow.md` → Reviews. Same fallback to `general-purpose`. |

Setup (run in each new task or update worktree, for the parts it touches): (none)

Always allowed in any task (no need to list in `touches`):

- Lock files: `go.sum`, only when `go.mod` did not change

Auto-fix (runs before the checks):

| Part | Folder | Command |
|---|---|---|
| CLI | `.` | `gofmt -w cmd internal` |

Checks:

| Part | Tier | Folder | Command | Passes when |
|---|---|---|---|---|
| CLI | task | `.` | `gofmt -l cmd internal` | it prints nothing |
| CLI | task | `.` | `go vet ./...` | exit code 0 |
| CLI | task | `.` | `go test ./...` | exit code 0 |
| CLI | task | `.` | `go build ./...` | exit code 0 |

Run locally (for the preview; run in the slice worktree):

| Part | Folder | Command |
|---|---|---|
| CLI | `.` | `go run ./cmd/flow <command>`, against a temporary test repo |

Test dependencies (checks start these themselves; never shared services):

| Dependency | Used by | In checks | How |
|---|---|---|---|
| Git repositories | CLI | yes | temporary repos created by the tests |

Known gaps (chosen in setup; revisit when possible):

- Workflow files (Markdown) have no automatic checks. Reviewers check them by hand. Chosen by bert-pratya, 2026-10-18.
- No Stop hooks: the role files and the pre-commit hook already require the checks. Chosen by bert-pratya, 2026-10-18.

Hooks and CI:

- Pre-commit hook: `.githooks/pre-commit` runs the Go checks when a commit includes Go files. Turn it on once per clone: `git config core.hooksPath .githooks`.
- CI: `.github/workflows/checks.yml` runs the task checks on Linux, Windows, and macOS, on every push to `main` and every pull request.

## 10. Project rules

### Git policy

- Base branch: `main`
- Slice branch: `slice/<name>` from `main`, worktree `.worktrees/<name>/`. Slices made before 2026-10-20 keep their numbered names (e.g. `slice/0001-slice-and-events`).
- Task branch: `task/<name>-NN` from the newest slice branch, worktree `.worktrees/task-<name>-NN/`
- Update branch: `update/<name>` from the newest slice branch, worktree `.worktrees/update-<name>/` (to update a slice from `main`)
- Small changes: `fix/<short-name>` from `main`
- After the human approves the result: the lead merges into `main` with a merge commit (`git merge --no-ff`) and pushes to `origin`.
- Approval tags travel with every push: run `git config push.followTags true` once per clone.
- Remote: `origin` is private and holds the full history.
- Slices update from `main` before the integration review, and again just before merging, in an update worktree.
- Parallel slices: merge one at a time.
- Commits: the lead commits after every step that changes something: `<name>: <what happened>`, and task work as `<name>/NN: <summary>`.

### Other rules

How work is done in this project. Each rule names who follows it and when.

- Everyone: write in plain, simple English. Short sentences.
- Everyone: keep everything generic. No company names, internal URLs, or secrets in this repo.
- Lead: until flow's first release, run this repo in manual flow (`workflow/manual-flow.md`). Use the `flow` binary only in tests, on temporary repos, never to manage this repo's slices.
- Lead, planning: shared Go logic goes in `internal/`, not `cmd/`.
- Worker: every change that adds behavior adds a test for it.
- Worker: `flow` itself sends no data to any outside service: no telemetry, no uploads, no calls home. It reads and writes repo files, and runs local programs: git, the check commands, and the agent tools. What those programs send is theirs, not `flow`'s.
- Worker: run each check or git command as a separate process with its own arguments, never as one shell string.
- Worker: paths, line endings, and process handling must work on Windows, macOS, and Linux.
