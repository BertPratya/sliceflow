# flow-setup: readiness items

What setup checks, how important each item is, and the usual
options when it is not ok. Adapt the options to what you find.

Levels:
- **required:** slices cannot run without it. No skip.
- **recommended:** works without it, but worse. Skip allowed, with the cost shown.
- **optional:** nice to have.

## Repo and git

| Item | Level | If not ok |
|---|---|---|
| Git repository with a base branch | required | Create it, or the human names the base branch. |
| `git config user.name` and `user.email` set | required | The human sets them. Approvals and commits need a name. |
| A remote | recommended | Add one. If skipped: no push, no CI, slices merge locally only. |
| `slices/`, `backlog.md`, `.gitattributes`, and `.worktrees/` in `.gitignore` | required | Create them. `.worktrees/` holds the slice, task, and update worktrees. |
| Test runners and linters skip `.worktrees/` | recommended | Add `.worktrees/` to their ignore settings. The `go` command skips it already; `gofmt -l .` does not. If skipped: running them in the main checkout also runs the copies in every worktree. |

## Workflow files

| Item | Level | If not ok |
|---|---|---|
| `AGENTS.md` block, `workflow/`, `.agents/skills/` | required | Copy them from the source. |
| Same version as the source (`workflow/VERSION`) | recommended | Update the block, `workflow/`, and skills. Sections 8-10 stay. If skipped: newer rules and fixes are missing. |
| `AGENTS.md` sections 8-10 filled, from `workflow/templates/project.md` | required | Fill them (steps 3-4). |

## Checks

| Item | Level | If not ok |
|---|---|---|
| Every part has checks for format, lint, types (where the language has them), test, build | recommended | 1. Use the existing tool (Makefile, package script). 2. Add a standard tool: a new dependency, so the human must agree. 3. Skip: reviewers look for those mistakes by hand, which is less reliable. |
| Auto-fix commands for every formatter and linter that can fix (e.g. `gofmt -w`, `eslint --fix`, `prettier --write`) | recommended | Add them to the stack profile. They run before the checks. If skipped: workers fix formatting by hand, and checks fail more often. |
| Checks pass on the base branch | required | Parts that do not exist yet are skipped (`workflow/formats/lifecycle.md` → Checks). Otherwise the human fixes it, as a small change or a slice. Slices cannot start on a failing base: every task would fail for reasons it did not cause. |
| Checks are safe to run at the same time in two worktrees: free ports, unique container names, own temp folders | recommended | Fix the checks. Test it by running the task checks in two worktrees at once. If skipped: tasks run in parallel, but the lead runs the full checks one task at a time, so workers may still see random failures from each other. |
| Slow tests (end-to-end, browser) are in the `slice` tier | recommended | Move them. If skipped: every task waits for them. |

## Test dependencies

For every service the tests use (database, cache, queue, file storage,
external API, login provider, browser):

| Item | Level | If not ok |
|---|---|---|
| Tests use a local stand-in, never a shared service | recommended | See the options below. |
| Tests refuse non-local addresses | recommended | Add the guard in the slice that sets up the stand-in. If skipped: a wrong environment variable can point tests at shared data. |
| No real secrets needed to run checks | required | Replace them with test keys or fakes. Workers must never get real secrets. |

Usual options when tests use a shared service, or cannot reach one:

| Situation | Options |
|---|---|
| Tests use a shared service (e.g. the dev database) | 1. Move to a temporary container or an embedded version, as the first slice (recommended). 2. Keep it, but run those tests one at a time and never in parallel slices: still touches shared data, slower, can fail randomly. 3. Skip those tests in checks: the bugs they would catch reach review or production; requirements that need them become "manual", tested in the preview. |
| Tests cannot reach the service, or there is no Docker | 1. An embedded version that needs no Docker (recommended). 2. Run those tests only in CI, where Docker exists: problems show up later. 3. A fake of that layer: fast, but it does not catch real query or protocol bugs. |
| No tests for that service at all | 1. Add them as the first slice (recommended). 2. Skip: those bugs are found only in the preview. |
| An external API (payment, email, SMS, maps) | 1. A fake server in the tests with prepared responses (recommended). Never the real service or its sandbox in checks. A sandbox test can exist as a separate check a human runs on purpose. 2. Skip: error handling for that API is untested. |

## Tool adapters and hooks

| Item | Level | If not ok |
|---|---|---|
| Adapters for each tool the team uses | required | Create them. Without them, that tool does not load the workflow. |
| Every role has a tested "How to start" in the stack profile | recommended | Find it and try it once (setup step 3). If left empty: the lead finds and tries it the first time it needs that role, and writes it in. |
| Pre-commit hook | recommended | Add it. If skipped: format errors and failing tests are found later, in CI. |
| Stop hooks running the task checks | optional | Add them for tools that support them. If skipped: the role files and CI still require checks. |
| CI job running the task checks | recommended | Add it, with the human's OK. If skipped: a skipped hook lets errors reach the base branch. |
