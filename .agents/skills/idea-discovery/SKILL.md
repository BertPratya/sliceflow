---
name: idea-discovery
description: Helps the human turn an early product idea into a shared overview, before any slice. Use when the human has an early idea, or wants to agree on a direction before choosing a slice.
---

# idea-discovery

The goal is a **shared overview**: you and the human see the same thing
to build, and why. It sets the direction. Approving a slice is a later,
separate step.

## 1. Look for related ideas

Read `docs/ideas/index.md`, if it exists. If an earlier idea is close to
this one, say so in one line, and read its file before you go on.

## 2. Let the human propose

Invite the human to describe what they want to build, in their own
words. One open invitation, e.g. "Tell me what you have in mind." Let
them talk first: the questions come out of what they say.

## 3. Brainstorm together

Talk naturally, like two engineers at a whiteboard:

- Talk about real cases: who uses it, what they do, what they see.
- Ask one open question at a time, while the idea is still forming.
- Research how the repo does it today and how other tools do it. Use
  real sources and link them. Say when you could not check something.
- Explain what is unclear in plain words, with a small example.
- When there is a real choice, show 2-3 options, each with its
  trade-offs, then your recommendation and why.
- Keep the line clear: a point is the human's decision only when they
  said it. Everything you suggest is "the lead's proposal" until they
  confirm it.
- Stay at the level of direction: what, for whom, why, and the main flow.
  Exact fields, flags, and messages belong to the slice spec later.

The step is done when the direction is clear: the human has no open
point they want to settle now.

## 4. Play the overview back

Summarize it in the chat, in plain words, so the human can check that
you think the same way:

- **What:** one or two sentences.
- **Who and why:** who has the problem, and why it matters.
- **How it works:** the workflow, as numbered steps the user goes
  through.
- **In and later:** what is in, and what waits for later.
- **Decisions:** the human's, and your proposals, marked as such.
- **Open points:** what is still unclear.

Then ask: "Does this overview match what you mean?" Say that a yes sets
the direction only, and approves no slice.

If they correct it, fix it and play it back again.

## 5. Save it when the human confirms

The human's yes is the moment to save. Save without asking again, and
tell them the path in one line.

1. Write `docs/ideas/YYYY-MM-DD-<topic>.md` from
   `docs/templates/idea.md`. `<topic>` is short: lowercase letters,
   digits, and hyphens.
2. Add one row to `docs/ideas/index.md`. If the file does not exist yet,
   create it in the form below.
3. Commit only these two files, in the main checkout, on `main`:
   `ideas: <topic>, direction agreed`.

If the human only wanted to think out loud and never confirmed an
overview, nothing is saved.

### The index

`docs/ideas/index.md` is a short table, so an agent can find an idea
without reading every file. One row per idea, newest first:

```markdown
# Ideas

| Date | Idea | Status | Summary | Slices |
|---|---|---|---|---|
| 2026-10-23 | [due-dates](2026-10-23-due-dates.md) | agreed | Tasks get a due date; overdue tasks stand out in the list. | - |
```

- Summary: one sentence, the same as the idea file's What.
- Status:
  - `agreed`: the direction is agreed; no slice yet.
  - `in slices`: slices were created from it; Slices names them.
  - `dropped`: the human dropped it; the idea file says why.
- When a status changes, update the row and the idea file's front matter
  together, in one commit.

## 6. Then

Suggest the next step in one line, e.g. "When you want, this can become a
slice." The human starts it, with `start-slice`.
