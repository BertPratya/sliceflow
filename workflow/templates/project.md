<!-- Template for AGENTS.md sections 8-10: this project's own part.
     flow-setup fills it. Delete the comments once a section is filled.
     A section with nothing in it says "(none)". -->

## 8. Project

<!-- What the project is and who it is for, in 2-3 sentences. -->

## 9. Stack profile

Parts:

| Part | Folders | Tech |
|---|---|---|
<!-- | CLI | `cmd/`, `internal/` | Go |
     | Web | `web/` | React + TypeScript |
     A part with none of its folders yet has nothing to check. -->

Connections (contracts between parts):

| From → To | Contract |
|---|---|
<!-- | Web → API | `api/openapi.yaml` (OpenAPI) | -->

Tools per role (the lead is the tool you chat in):

| Role | Tool | How to start |
|---|---|---|
<!-- | Worker | Codex | `Get-Content <prompt file> | codex exec -C <worktree> --sandbox workspace-write -` |
     | Code reviewer | Claude Code | subagent `flow-code-reviewer` |
     | Plan reviewer | Claude Code | subagent `flow-plan-reviewer` |
     The prompt goes in a file, never as one quoted argument.
     flow-setup tries each one once. A row may stay empty: the lead fills it on first use. -->

Setup (run in each new task or update worktree, for the parts it touches):

| Part | Folder | Command |
|---|---|---|
<!-- | Web | `web/` | `npm ci` | -->

Always allowed in any task (no need to list in `touches`):

<!-- - Generated code: `internal/api/gen/`
     - Lock files: `go.sum`, `web/package-lock.json` -->

Auto-fix (runs before the checks; changes only layout, or fixes the tool marks as safe):

| Part | Folder | Command |
|---|---|---|
<!-- | Go | `.` | `gofmt -w cmd internal` | -->

Checks:

| Part | Tier | Folder | Command | Passes when |
|---|---|---|---|---|
<!-- | Go | task | `.` | `go vet ./...` | exit code 0 |
     | Web | slice | `web/` | `npx playwright test` | exit code 0 |
     Tiers: task runs after each task; slice runs once all tasks are done. -->

Run locally (for the preview; run in the slice worktree):

| Part | Folder | Command |
|---|---|---|
<!-- | API | `.` | `go run ./cmd/server` |
     | Web | `web/` | `npm run dev` | -->

Test dependencies (checks start these themselves; never shared services):

| Dependency | Used by | In checks | How |
|---|---|---|---|
<!-- | PostgreSQL | API | yes | temporary container per test run | -->

Known gaps (chosen in setup; revisit when possible):

<!-- - What is skipped, its effect on the workflow, who chose it, and the date. -->

## 10. Project rules

### Git policy

<!-- - Base branch: `main`
     - Slice branch: `slice/NNNN-<name>` from `main` (or from another slice's branch when stacked), worktree `.worktrees/NNNN-<name>/`
     - Task branch: `task/NNNN-NN` from the newest slice branch, worktree `.worktrees/task-NNNN-NN/`
     - Update branch: `update/NNNN` from the newest slice branch, worktree `.worktrees/update-NNNN/`
     - Small changes: `fix/<short-name>` from `main`
     - After the human approves the result: who merges, and whether to push
     - Slices update from `main` before the integration review, and again just before merging
     - Parallel slices: merge one at a time
     - Commits: the lead commits after every step: `NNNN: <what happened>`, task work `NNNN/NN: <summary>` -->

### Other rules

How work is done in this project. Each rule names who follows it and when.

<!-- The roles read the rules for themselves and follow them at that step.
     flow-setup asks the human for them. Examples:
     - Lead, planning: split tasks by layer: data, API, then screens.
     - Worker, UI tasks: use only the colors and spacing in `web/src/styles/tokens.css`.
     - Lead, planning: one task per endpoint, backend and frontend together.
     - Code reviewer: every new endpoint has an authorization test.
     - Everyone: names in English, comments only where the code is not obvious. -->
