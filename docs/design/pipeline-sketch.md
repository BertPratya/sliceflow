# Pipeline sketch (DRAFT)

> **Status: brainstorm, not approved.** This is a working sketch of the whole
> agent-flow pipeline, from a real user's point of view, start to end.
> It may be messy. When it is settled, we decide what it becomes and where it lives.
>
> It summarizes and points. The detailed rules stay where they are
> (`AGENTS.md`, `workflow/formats/lifecycle.md`, the skills, the roles).
> Unclear points are marked **⚠ H1, H2, …** and listed at the end.
>
> Last update: 2026-10-17

---

## 1. The whole picture

The phases run left to right. Each row is one actor.

```
             SETUP       SPEC                PLAN                IMPLEMENTING           INTEGRATION        PREVIEW         DONE
             (once)      (per slice)         (per slice)         (per slice)            (per slice)        (per slice)

You          choose      answer questions    decide hard         only when stuck        only when stuck    try it,         -
             fixes       approve spec        choices,                                                      approve
                                             approve plan

Lead         check repo  create slice        write plan,         start the run,         handle stops       show what       merge
(your chat)  set up      interview, write    contracts, tasks    handle stops                              to try
                         spec

flow         doctor      slice new,          validate,           pick tasks, start      update from main,  -               slice done
(the CLI)    init        validate, record    record approval     workers, check,        slice checks,
                         approval                                merge tasks, record    start review

Workers      -           -                   -                   build one task,        fix tasks          -               -
                                                                 run checks

Reviewers    -           spec check          plan review         task review            integration        -               -
                         (questions)                                                    review
```

What moves a slice from one phase to the next is in section 4 (gates).

---

## 2. Use cases

Real things a user wants to do. Each one is a path through the steps in section 3.

| # | Use case | Steps |
|---|---|---|
| UC0 | Brainstorm: explore an idea, or the whole system (like this sketch). Outside `flow`: no slice yet. Results go to `backlog.md` or the project's own docs | before step 1 |
| UC1 | Set up a repo for the workflow (once) | S1-S3 |
| UC2 | **Build a feature, from idea to merged** (the main path) | 1-37 |
| UC3 | Something goes wrong during the run (blocked task, failing checks, review asks for changes, conflict) | 21-27 |
| UC4 | Change your mind (spec change, plan change, cancel a slice) | side paths in section 3 |
| UC5 | Plan the next feature while this one is being built (stacking) | 2 with `--after` |
| UC6 | Two features at the same time (parallel slices) | 28, 35 |
| UC7 | See how things are going | `flow status`, dashboard |
| UC8 | Update a project to a newer workflow version | S1-S3 again |

---

## 2b. Reviews: the shared model ✅ settled (2026-10-17)

Every review (spec check, plan review, task review, integration review) works the same way:

| Part | How |
|---|---|
| Who starts the reviewer | `flow`, in a fresh context, with the tool from `flow.yaml` |
| File | `reviews/<kind>-<target>-<round>.md`, created by flow with its header |
| What the reviewer writes | Summary, earlier findings (verified or reopened), new findings F1… (high, medium, low), "Needs human decision" |
| Hand-in | The reviewer runs `flow review submit`. flow checks only what it reads (finding headings with a level, the sections, every marked finding verified or reopened, the integration requirements check) and sets the result. A wrong format: flow says which line to fix; after 2 failed tries (setting) it stops and reports. |
| Result | anything under "Needs human decision" → needs you; else any high or medium (new or reopened) → changes; else → pass |
| A finding's life | open → fixed, wontfix, or deferred (by the fixer, with `flow finding …`) → verified or reopened (by the reviewer, next round). Closed = verified. A deferred finding stays visible ("deferred → task 05") until that task's review or the integration review verifies it. |
| Findings list | One list per slice in `status.md`, grouped by review, generated from the events; never edited by hand |
| Unmarked lows | Marking is optional. When the slice is done they close as "left as is" (listed once in the final summary); lows titled "later" go to `backlog.md` |
| Later rounds | verify fixes; a new finding only if high or caused by a fix |
| Low findings | never block |
| Limit | round limit from `flow.yaml` (default 3) without a pass → blocked, you decide |

What differs by kind:

| | Spec check | Plan review | Task review | Integration review |
|---|---|---|---|---|
| Reviewer | plan reviewer | plan reviewer | code reviewer | code reviewer |
| Reads | spec | spec, plan, contract changes, tasks | the task, its code, its requirements | the whole slice |
| Writes | questions only | findings, questions | findings, questions | findings, requirements check, questions |
| Who fixes | you and the lead | the lead (with you when a decision is needed) | the worker | workers, via fix tasks |
| After a pass | spec awaiting_approval | plan awaiting_approval | task is merged | preview |
| Simple track | yes | light: checks 1-4 | core tasks only | yes |

---

## 3. Every step, in order

Format of each step:

```
Who / Command / Reads → Writes / What it checks / If it fails / Needs
```

### Setup (UC1, once per repo)

**S1. Check the repo**
- Who: lead. Command: `flow doctor`
- Reads the repo → writes nothing. Reports each readiness item: ok, missing, out of date, or a known gap.
- Checks: git, workflow files, checks pass on main, test dependencies are local, tool adapters, hooks, CI.

**S2. Choose what to fix**
- Who: you. For each item that is not ok: fix now, skip (with its cost), or later (backlog).

**S3. Set up**
- Who: lead. Command: `flow init` (copies the workflow files), then by hand: `AGENTS.md` sections 8-10, adapters, hooks.
- Writes the workflow files and settings. Commits on a branch. Never changes code: code gaps go to `backlog.md`.

### Spec (per slice)

**1. Is it a slice?**
- A vague idea is explored first, outside `flow` (UC0). The slice is created only when it is one clear feature.
- Who: lead, with you. Small change → normal chat, no slice. Too big (more than ~25-30 requirements) → split.
- Also: does it need something from an unmerged slice? Then propose stacking (UC5). You confirm.

**2. Create the slice** ✅ settled (2026-10-17), except the open points below
- Who: lead (or you). Command: `flow slice new <name>`. Options: `--title "<title>"` (default: made from the name), `--after NNNN` (stacking, may come later).
- Slice ID = next free number + name, e.g. `0003-device-pairing`. The number is the next one after the slices on main and every slice branch flow can see. ⚠ H11
- Checks first, and changes nothing if one fails: repo is set up (else "run `flow init` first"), the name is valid (lowercase letters, digits, hyphens), the name is not taken, git user name is set.
- Creates: branch `slice/NNNN-<name>` from the base branch, worktree `.flow/worktrees/NNNN-<name>/`, and in it `slices/NNNN-<name>/spec.md` (from the template, header filled: slice, title, created; status draft) and `events.jsonl`. No empty folders: `tasks/`, `reviews/`, `history/` appear with their first file.
- Commits on the slice branch: `NNNN: new slice`.
- All or nothing: if a step fails, it undoes what it did and says why.
- Prints the slice ID, branch, worktree, files, phase, and the next step.
- Event: `slice.created` { slice, title, from, after, by, at }
- State after: the slice exists · phase **spec** · spec.md draft
- Waits on you: nothing

**3. Learn the facts**
- Who: lead. Reads the code, earlier specs, `backlog.md`, contracts. Asks you nothing it can look up.

**4. Interview**
- Who: lead and you. Rounds of up to ~5 questions, each with a recommendation.
- Writes the `spec.md` draft after each round, and `notes.md` for "how" ideas.

**5. Write the spec**
- Who: lead. Summary, walkthrough, requirements R1, R2, … each with a `Check:` line, out of scope, decisions.

**6. Spec self-check** ✅ settled for the spec (2026-10-17); plan checks are settled at step 15
- Who: lead. Command: `flow validate`
- Read-only: changes nothing, records no event.
- Arguments (settled): `flow validate` checks every slice in the folder it runs in; `flow validate 0003` checks one slice (the number is enough). It checks whichever documents exist.
- Levels (settled): **error** = the format is broken (validate fails); **warning** = not finished yet, e.g. open questions (validate passes). `--strict` also fails on warnings; `flow doc submit` and `flow approve` use it.
- Spec checks (settled, 2026-10-17):
  - errors: (1) header fields filled and valid (status: draft, in_review, awaiting_approval, approved, changed); (2) header `slice` matches the folder; (3) no `{{placeholder}}` left; (4) every template section present ("Key concepts" optional); (5) each requirement is `- **R<n>.** text` plus a `Check:` line (unit, api, e2e, manual); (6) a line that looks like a requirement but is formatted wrong; (7) no number used twice; (9) no approved number disappears (a removed one stays as a "(removed in …)" line); (10) a removed number never comes back with new text. Checks 9-10 apply only once an approved version exists in `history/spec/`. (13) no gaps in the numbers, e.g. R1-R10 without R5: if R5 was approved, check 9 reports it; if not, renumber the requirements not approved yet. Numbers out of order are allowed (a new requirement takes the next free number under any heading).
  - warnings: (8) a requirement without "shall"; (11) Open questions not empty; (12) a section with no content and no "(none)".
  - left to the spec check reviewer: does it make sense, is it testable, is every walkthrough step covered.
- Exit codes (settled): **0** valid (warnings allowed), **1** problems found, **2** could not check (not a repo, unknown slice, wrong command). All commands get one rule later (H13).
- Output (settled): one line per problem, `path:line: level: message with a fix hint`, then a count (`2 errors, 1 warning`); `OK: 3 slices checked` when nothing is wrong.
- If it fails: the lead fixes the spec and runs it again.
- When it runs: ⚠ H14 (settled later).

**7. Spec check (review)** ✅ settled (2026-10-17)
- Who: lead starts it. Command: `flow doc submit spec`. It: (1) runs `flow validate --strict` and refuses if the spec is not finished; (2) sets `in_review`; (3) creates `reviews/spec-R001-1.md` with its header; (4) starts the plan reviewer itself, in a fresh context, with the tool from `flow.yaml`; (5) waits while it works; (6) prints the result (e.g. "4 questions. Read the spec first."). If the reviewer ends without submitting, flow starts it once more, then reports. Long-running: Claude Code runs it in the background; Cursor uses `--detach` and `flow wait`.
- Settled: the reviewer lists **every real gap** as a question, most important first; each says what a worker would have to guess. The lead answers facts itself and asks you the decisions in rounds of at most 5, writing the answers into the spec. If the answers add requirements, the check runs again on those only.
- Settled: document status shows who has the ball: `draft` (lead) → `in_review` (reviewer) → `awaiting_approval` (you) → `approved`. `flow` sets it when the review is submitted: questions → back to `draft`; none → `awaiting_approval`. Submit again after answering: no new requirements → `awaiting_approval` directly; new ones → `in_review`. Approve only from `awaiting_approval`.
- Reads the spec → writes `reviews/spec-R001-1.md`.

**8. Answer the questions**
- Who: you and the lead. You read the spec first, then see the questions. The lead updates the spec and validates again.

**9. Approve the spec** (from existing rules; nothing new to decide)
- Who: you say yes. The lead runs `flow approve spec`, only right after you say so.
- Needs: spec is `awaiting_approval` and `flow validate --strict` passes.
- Writes: spec header (`approved`, revision R001, approved_at, approved_by = your git name), copy in `history/spec/R001.md`, commit `NNNN: spec R001 approved`.
- Phase → **plan**. `flow reject spec` instead: back to `draft`, with the reason.
- Event: `doc.approved` { doc: spec, revision, by, at } · Waits on you: nothing (the lead starts the plan).

> **Side path, change an approved spec (UC4):** reason → interview what changes → edit (numbering rules) → validate → spec check if a requirement changed (skip for wording only) → you approve again. Then the plan may need a change too.

### Plan (per slice)

**10. Learn the code**
- Who: lead. Parts this slice touches, how they are tested, contracts, earlier decisions. If stacked: the earlier slice's plan.

**11. Hard decisions**
- Who: you. Only decisions that are hard to reverse, surprising, and a real trade-off. The lead makes small choices itself and records them (TD1, TD2, …).

**12. Write the plan**
- Who: lead. Flows, interfaces, data, decisions, track (simple or complex), the plan's `touches` (contracts and other living docs only).

**13. Change the contracts**
- Who: lead only, only now. Edits the contract files on the slice branch.

**14. Split into tasks**
- Who: lead. One file per task in `tasks/`: covers (requirements), touches (code), depends_on, risk, layer, "Done when" tests.

**15. Plan self-check** ✅ settled (2026-10-17)
- Who: lead. Command: `flow validate` (same command, output, and exit codes as step 6; checks the plan once `plan.md` exists, the tasks once task files exist). Same level rule: broken or invalid → error; incomplete or outdated → warning (`--strict`, submit, and approve refuse it).
- plan.md: errors: (P1) header valid (track simple|complex, …); (P2) `slice` matches the folder; (P3) no placeholder; (P4) fixed sections present (Approach, Flows, Interfaces, Data, Decisions, Open questions; Risks and Parked optional); (P5) every `Covers:` names a real requirement; (P6) decisions `- **TD<n>.**`, unique, no gaps, none disappears after approval ("(replaced by TD9)"). Warnings: (P7) every requirement covered by a flow or shared rule; (P8) based on the latest approved spec; (P9) Open questions empty; (P10) every section has content or "(none)".
- Task files: errors: (T1) `tasks/NN-name.md`, unique numbers, header `task` matches; (T2) header valid (layer, risk, non-empty touches); (T3) no placeholder, all sections present; (T4) `covers` names real requirements; (T5) the Requirements section copies each covered requirement word for word; (T6) `depends_on` names real tasks, no loops; (T7) `touches` never includes a file from the plan's `touches`. Warnings: (T8) every requirement covered by a task; (T9) every covered requirement has a "Done when" row; (T10) `touches` inside the project's parts; (T11) based on the current plan revision.
- Across slices (warnings): (X1) another unmerged slice changes the same shared file, neither stacked; (X2) a contract change no longer matches the approved patch; (X3) stacked: the earlier slice's plan changed since this plan was approved.
- Left to the plan reviewer: real coverage, contradictions, tasks buildable alone, tests really prove requirements, right track.

**16. Plan review** ✅ settled (2026-10-17)
- Who: lead starts it (`flow doc submit plan`), same mechanism as the spec check: strict validate, `in_review`, review file, flow starts the plan reviewer, waits, reports. The reviewer checks spec, plan, contracts, and tasks together. Light on simple (checks 1-4), full on complex.
- Who fixes: the lead, alone when no decision from you is needed; with you when a finding is under "Needs human decision", the fix is a hard decision, or a high finding should be wontfix or deferred out of the slice. A spec problem → spec change.
- The lead marks each finding fixed, deferred, or wontfix with a note, then submits again → next round (always `in_review`: the reviewer verifies the fixes).
- Round 2 and later: the reviewer verifies or reopens each marked finding, and adds a new finding only if it is high or was caused by a fix.
- Low findings never block. The fixer may fix or leave them; a low titled "later" goes to the backlog when the slice is done.
- Result: pass → plan `awaiting_approval`; changes or questions → `draft`.
- Round limit (`flow.yaml`, default 3) without a pass → blocked; you decide: accept as it is, change the plan, or allow one more round.

**17. Approve the plan** ✅ settled (2026-10-17), same as the spec
- Who: you say yes (one yes for plan, tasks, and contract changes). The lead runs `flow approve plan`, only right after you say so.
- Needs: plan is `awaiting_approval` and `flow validate --strict` passes.
- Writes: plan header (approved, revision R001), `history/plan/R001.md`, a copy of the tasks, `history/contract/R001.patch`, commit `NNNN: plan R001 approved`.
- Phase → **ready**. `flow reject plan` instead: back to `draft`, with the reason.
- Event: `doc.approved` { doc: plan, revision, by, at } · Waits on you: nothing (the lead starts the run).

> **Side path, change an approved plan (UC4):** reason → change only what must change → validate → plan review of the change → you approve again.

### Implementing (per slice, mostly automatic)

**18. Start the run** ✅ settled (2026-10-17)
- Settled: `flow run [slice]`. Without a number: in a slice worktree, that slice; in the main checkout, the only runnable slice, or it lists them and asks for the number (exit 2, nothing starts). The lead always gives the number. `--detach` for tools without background commands. Worker slots from `flow.yaml`.
- Settled: start conditions, else nothing starts (exit 2, with the reason): (1) spec approved, not changed or in review; (2) plan approved; (3) plan based on the current spec revision; (4) contract changes match the approved patch; (5) checks pass on the base branch; (6) stacked: the earlier slice is merged; (7) no other run on this slice (`.flow/run.lock`); (8) no uncommitted changes in the slice worktree.
- Settled: runs until preview or a stop, then prints a report: each task done, blocked (with the reason), or waiting (with what it waits for), and the next step.
- Settled: exit codes for all commands (solves H13): 0 done, 1 stopped and needs attention, 2 could not run, 130 interrupted.
- Settled (also for every other command): `flow` commits automatically after every command that changes something, in the slice worktree, with `NNNN: <what happened>` (task work: `NNNN/NN: <what changed>`). Nobody commits slice files by hand; in bootstrap mode the lead does it at the same moments. `submit` commits first, so a review always reads an exact saved version. Solves F7.
- Settled: the lead starts `flow run` right after you approve the plan, without asking. It waits only when the slice is stacked (earlier slice not merged) or you said to hold. `flow approve` itself never starts the run.
- Who: lead. Command: `flow run 0003 --detach`, then `flow wait 0003` (every tool the same way; in Claude Code `flow wait` runs as a background command). `flow wait` returns when the run ends, or on an alert (e.g. a worker over its time limit).
- Needs: plan approved; base checks pass (`flow check --base`); if stacked, the earlier slice is merged.
- The lead does not watch. `flow` reports when the run ends (step 27).

**19. Pick the next task** ✅ settled
- Who: `flow`. Records nothing. A task can start when: its `depends_on` are merged, its `touches` do not overlap a running task, and a worker slot is free (1 in v1).

**20. Start the worker** ✅ settled (2026-10-17)
- Who: `flow`. (1) Creates `task/NNNN-NN` from the newest slice branch (so it contains every task merged so far) and its worktree. (2) Runs each touched part's `setup` command from `flow.yaml` (e.g. `npm ci`) in the new worktree; if setup fails, the task is blocked "setup failed" and it does not count against the worker. (3) Starts the worker tool from `flow.yaml` in the worktree with the same prompt every time: "You are the worker. Read `workflow/roles/worker.md` and follow it. Run `flow brief worker NNNN/NN`." (4) Task → `in_progress`; at the first worker, phase → **implementing**.
- `flow brief worker` shows in full: the task file, the requirements text from the approved spec, open findings to fix first, the lead's notes, conflict files, the files it may change, the checks, and how to finish. Plan sections and contract files only as paths.
- Time limit (`limits.worker_minutes`, default 60): flow does not stop the worker. `flow wait` wakes the lead with the time and the last activity ("last file change 1 min ago" or "silent for 25 min"). The lead lets it continue (`flow wait` again), stops it (`flow task stop NNNN/NN`: saves the work on the task branch, task blocked), or asks you.
- Event: `task.started` { task, tool, branch, at } · State after: task in_progress · phase implementing (first worker) · Waits on you: nothing (only if the lead asks about a slow worker).

**21. Build** ✅ settled (2026-10-17)
- Settled: `flow check` runs auto-fix, then the task checks, only for the touched parts; one line per command; for a failure only the last 40 lines plus the full log path (`.flow/logs/…`); after the first failure the rest are skipped. The worker runs it when it believes it is done and after each fix (single tests while working are free); flow runs it at hand-in (if files changed since the last pass), at merge, at run start (`--base`), at integration (`--slice`).
- Settled: `flow check` also fails when the worker changed a file outside `touches` (plus always-allowed files), and names the files.
- Settled: every failed `flow check` is one attempt. After `limits.check_failures` (default 3) since the last review, the task is blocked and the worker stops; the lead unblocks with a hint if it sees the cause, otherwise asks you.
- Settled: `touches` describes an area (folders, files, patterns), narrow but not exact. When unblocking, the lead may widen a task's `touches` without a new approval, if the path stays inside the project's parts, is not in the plan's `touches`, and adds no behavior; it tells you in one line. Anything else is a plan change.
- Settled: when a worker blocks with a question, the run continues with tasks that do not depend on it; it ends (exit 1, needs attention) only when nothing else can run.
- Who: worker. Command: `flow brief worker` (what to read), then code and tests inside `touches`.
- Command: `flow check` = auto-fix (formatting), then the task checks. Repeats until they pass.
- If checks fail 3 times since the last review → the task is blocked (UC3).
- If something is unclear and changes behavior → `flow task block` with a question (UC3).

**22. Hand in** ✅ settled (2026-10-18)
- flow then: (1) reruns `flow check` if files changed since the last pass; (2) checks only allowed files changed; (3) commits on the task branch `NNNN/NN: <summary>` (note in the body); (4) task → `in_review` if it needs a review, else straight to merge (step 25).
- Event: `task.handed_in` { task, summary, note, commit, at }.
- Who: worker. Command: `flow task review NNNN/NN --summary "<one line>" [--note "Assumed: …"]` (settled: `--summary` is required and becomes the commit message `NNNN/NN: <summary>`; the note goes in the commit body and the handoff), or `flow task block`.

**23. Task review** ✅ settled (2026-10-18)
- Follows the shared review model (2b). Reviewed: every task on the complex track, core tasks on the simple track.
- `flow brief code-reviewer` shows: the task file, the requirements from the approved spec, the task's diff (round 2+: also only what changed since the last round), earlier findings with the fixer's notes, the handoff (hand-in note, commits, changed files), that `flow check` passed, and plan/contract parts as paths.
- After: pass → merge (25); changes → new worker, fix round (24); needs you → the task waits, the run continues with other tasks.
- Who: `flow` starts a code reviewer (simple track: core tasks only; complex: every task).
- Reviewer: `flow brief code-reviewer`, writes `reviews/task-NN-<round>.md`, `flow review submit`.
- Verdict, set by `flow`: **pass**, **changes** (any high or medium finding), or **needs_human**.

**24. Fix round** ✅ settled (2026-10-18)
- Settled: a fix round starts a **new worker** (fresh context) on the same task branch and worktree. `flow brief` shows the findings to fix first, plus a handoff flow assembles from round 1: the hand-in note, the commits, the changed files. Nobody writes the handoff separately.
- Who: worker. On **changes**: fixes findings first (`flow finding fixed | deferred | wontfix`), then step 21 again.
- 3 review rounds without a pass → the task is blocked.

**25. Merge the task** ✅ settled (2026-10-18)
- Events: `task.merged` { task, commit, at } → task done, task branch and worktree removed; `task.merge_failed` { task, reason, files, at } → back to a new worker (not counted against it). After `limits.merge_failures` (default 2) failed merges in a row → task blocked, the lead looks (usually wrong `touches` or `depends_on`).
- Who: `flow`. Rebase onto the newest slice branch, run the task checks, move the slice branch.
- If it conflicts or checks fail → back to the same worker (step 21). Does not count against them.

**26. Repeat** steps 19-25 until every task is merged. Phase → **integration**.

**27. Handle a stop (UC3)**
- `flow run` ends with an exit code: 0 done, 10 needs a human, 20 interrupted, 1 error.
- Who: lead. Answer in an approved document → `flow task unblock` with a note. A new decision → ask you. Failed checks with a visible cause → unblock with a hint. Then step 18 again.

### Integration (per slice)

**28. Update from main**
- Who: `flow`. Brings in what other slices merged meanwhile (UC6). A conflict → a fix task (a contract conflict → the lead, and you approve again).

**29. Slice checks**
- Who: `flow`. Command: `flow check --slice` (all task checks, then slow tests like end-to-end).
- A failure → a fix task.

**30. Integration review**
- Who: a code reviewer, on the whole slice. Checks every requirement, and problems no single task could see.
- High or medium findings → fix tasks → back to step 18. You are told about each fix task in one line (no approval needed if it adds nothing new).

**31.** All pass → phase **preview**.

### Preview (per slice)

**32. Show you the result**
- Who: lead. What it does now, how to run it, the list of things to try by hand, what was parked.

**33. You try it**
- Problems → fix tasks → back to step 18. Something new → a spec change (UC4), not a fix.

**34. You approve the result.**

### Done (per slice)

**35. Update from main again**
- Who: `flow`. If nothing came in → go on. If something came in → checks again; a conflict → fix task, integration review, and preview again.

**36. Merge and finish** ⚠ H2
- Who: lead. Merge into `main`, push to `origin`. `flow slice done`: parked items go to `backlog.md`.

**37.** Phase → **done**. The slice worktree is removed.

---

## 4. What depends on what (gates)

| Gate | Opens when |
|---|---|
| Spec approved | validate passes + spec check answered + **you say yes** |
| Plan approved | spec approved + validate passes + plan review passes + **you say yes** |
| Run can start | plan approved + base checks pass + (stacked: the earlier slice is merged) |
| Task goes to review | `flow check` passes |
| Task is done | review passes (or checks pass, for a simple-track non-core task) + its merge succeeds |
| Integration starts | every task is done |
| Preview starts | update from main OK + slice checks pass + integration review passes |
| Merge | **you approve the result** + update from main OK + checks pass |

Only you open the three gates in **bold**. Everything else is checked by `flow`.

---

## 5. Commands, one line each

| Command | What it does | Who | Steps |
|---|---|---|---|
| `flow doctor` | Checks if the repo is ready | lead | S1 |
| `flow init` | Copies the workflow files into a repo | lead | S3 |
| `flow slice new` | Creates a slice: branch, worktree, files | lead | 2 |
| `flow slice move` | Moves a slice that has only documents onto another base | lead | UC5 |
| `flow status` | Shows slices, phases, what waits on you | anyone | any |
| `flow validate` | Checks the documents are written correctly | lead, CI | 6, 15 |
| `flow doc submit` | Sends a spec or plan to review | lead | 7, 16 |
| `flow approve` / `reject` | Records your yes or no | lead, only right after you say so | 9, 17 |
| `flow diff` | Shows a document's change since its last approval | anyone | UC4 |
| `flow run` | Runs the loop: tasks, checks, reviews, merges | lead | 18-31 |
| `flow wait` | Waits for a detached run (Cursor) | lead | 18 |
| `flow brief` | Tells a worker or reviewer what to read | worker, reviewers | 21, 23 |
| `flow check` | Auto-fix, then checks (`--slice`, `--base`) | worker, `flow` | 18, 21, 29 |
| `flow task review` / `block` | A worker hands in or asks a question | worker | 22 |
| `flow task unblock` | Answers a blocked task | lead or you | 27 |
| `flow task cancel` | Removes a task in a plan change | lead | UC4 |
| `flow task stop` | Stops a running worker, saves its work, marks the task blocked | lead | 20 |
| `flow finding fixed` / `deferred` / `wontfix` | Answers a review finding | worker (high wontfix: you) | 24 |
| `flow review start` / `submit` | Creates and hands in a review file | `flow`, reviewers | 7, 16, 23, 30 |
| `flow sync` | Rebuilds `status.md` (pre-commit) | hook | any |
| `flow slice done` | Moves parked items to the backlog | lead | 36 |

---

## 6. Where things live

```
AGENTS.md                      how we work + this project's profile (sections 8-10)
workflow/                      templates, formats, roles, adapters (the reusable workflow)
.agents/skills/                flow-spec, flow-plan, flow-implement, flow-review, flow-setup
backlog.md                     ideas for later
slices/NNNN-<name>/            one feature (lives on its slice branch until merged)
  spec.md  plan.md  notes.md
  tasks/  reviews/  history/
  status.md                    generated summary (never edit)
  events.jsonl                 the state: every change as one line (only flow writes)
.flow/                         local only: worktrees, run lock (not in git)
```

---

## 7. State model (filled step by step)

Rule (settled): for every step we settle **what changes** (event and status).
**What gets shown** (`flow status`, `status.md`) is designed at the end, from this list.

| Step | Event | State after | Waits on you |
|---|---|---|---|
| 2 | `slice.created` { slice, title, from, after, by, at } | slice exists · phase spec · spec draft | nothing |
| 6 | none (read-only) | unchanged | nothing |
| 7 | `doc.submitted` { doc, revision, by, at } | spec in_review | nothing |
| 7 | `review.started` { review, reviewer, at } | review running | nothing |
| 7 | `review.submitted` { review, verdict, questions, at } | spec draft (questions) or awaiting_approval (none) | answer the questions / approve |
| 8 | `doc.submitted` again | spec awaiting_approval (no new requirements) or in_review (new ones) | approve |
| 9 | `doc.approved` { doc, revision, by, at } | spec approved · phase plan | nothing |
| 15 | none (read-only) | unchanged | nothing |
| 16 | as in step 7: `doc.submitted`, `review.started`, `review.submitted`, `finding.marked` | plan in_review → draft or awaiting_approval | hard decisions on findings; approve |
| 17 | `doc.approved` { doc: plan, … } | plan approved · phase ready | nothing |
| 18 | `run.started` { slice, workers, by, at } / `run.ended` { result, done, blocked, at } | run going / over | depends on the result |
| 20 | `task.started` { task, tool, branch, at } | task in_progress · phase implementing | nothing |
| 20 | `task.stopped` { task, by, at } (only when the lead stops it) | task blocked | the lead decides, may ask you |
| any review | `finding.marked` { finding, as: fixed/deferred/wontfix, note, task?, by, at } | finding fixed / deferred / wontfix | high wontfix or defer out of slice: you |
| any review | (inside `review.submitted`) new findings open; earlier ones verified or reopened | findings list updated | — |

---

## 8. Settings: `flow.yaml` (filled step by step)

Settled: flow reads its settings from `flow.yaml` at the repo root, not from
the `AGENTS.md` tables. It comes from a template with comments
(`workflow/templates/flow.yaml`). `AGENTS.md` points to it, so no value is
written twice.

Settled: only the lead edits it, with the human's OK (like a contract);
workers never.

Settled: the template, this repo's `flow.yaml`, the move out of `AGENTS.md`,
and pointing the skills to it are done in one go, after the pipeline is
settled. Until then the `AGENTS.md` tables stay the source.

| Setting | Example | Used by |
|---|---|---|
| `base_branch` | `main` | step 2 (slice new), merge |
| `worktrees` | `.flow/worktrees` | step 2 |
| `branches.slice` | `slice/{id}` | step 2 |
| `limits.review_rounds` | `3` | every review (rounds without a pass before blocked) |
| `limits.check_failures` | `3` | step 21 (failed checks since the last review before blocked) |
| `limits.workers` | `1` | step 19 (tasks running at the same time) |
| `limits.merge_failures` | `2` | step 25 (failed merges in a row before blocked) |
| `limits.worker_minutes` | `60` | step 20 (alert the lead; the worker is not stopped) |
| `parts[].setup` | `npm ci` | step 20 (run in each new task worktree, for touched parts) |
| `roles.worker` (etc.) | `codex` | step 20 (which tool starts each role) |

Limits are settings because flow enforces them. The lead's interview round
size (about 5 questions) stays in the skills: it is guidance for a
conversation, not something flow can enforce. Raising a limit weakens a
guard, so it needs your OK, like any `flow.yaml` change. Until the
template exists, the docs still say "3".

---

## 9. Hotspots: unclear or open (⚠)

| # | What is unclear | Comes from |
|---|---|---|
| H1 | Which folders belong to which part, so checks know what to skip | review F1 |
| H2 | The order of the last steps by hand: backlog, done, merge, remove worktree | review F6 |
| H3 | ✅ solved: flow commits after every command; bootstrap: the lead, at the same moments | review F7 |
| H4 | "Always ask" settings vs workers that cannot answer a prompt | review F10 |
| H5 | A personal name in a file that will be public | review F18 |
| H6 | The list of events and their format (`events.jsonl`) is not designed yet | new |
| H7 | What exactly `flow doctor` and `flow init` do | new |
| H8 | Dashboard: what it shows, when we build it | backlog |
| H9 | UI workflow (sketch, tokens, screenshots) | backlog |
| H10 | Build order of the CLI: which steps first | open |
| H11 | Two people on different machines can get the same slice number (flow makes no network calls) | step 2 |
| H13 | ✅ solved: 0 done, 1 problems or needs attention, 2 could not run, 130 interrupted | step 6, 18 |
| H14 | When `flow validate` runs: by the lead, inside submit and approve, hooks, CI | step 6 |
| H12 | Document headers (`status: draft`) repeat state that is also in `events.jsonl`: which one wins? | step 2 |
