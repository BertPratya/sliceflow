---
name: brainstorming
description: Explores an idea, a direction, or a design with the human, without a slice, a spec, or code. Use when the human wants to brainstorm, think something through, or compare options before deciding whether to build it.
---

# Brainstorming

The human wants to think, not to build. Help them find what they want and
why, with research and options, and keep a clear line between their
decisions and your proposals.

## The gate

While brainstorming, never:

- create a slice, write a spec, plan, or tasks, or write code
- write or change a file before the human agrees to it (see "How it ends")
- record your own suggestion as the human's decision

Reading the repo and researching the web are always allowed.

## How it goes

1. **The human talks first.** Let them describe the idea in their own
   words. If it is vague, ask one open question that gets them to describe
   a real case. Do not start with a list of questions: there is nothing to
   resolve yet.
2. **Research.** Look at how the repo does it today and what other tools
   do. Use real sources and link them. Never invent a source. Say when you
   could not check something.
3. **Think wide, in private.** Before you answer, come up with options from
   several points of view, for example the user, the maintainer, security,
   cost, and how other tools solve it. AI ideas tend to look alike; looking
   from different sides is what makes them differ. Drop the weak ones.
4. **Show 2-3 real options**, each with its trade-offs, then your
   recommendation and why. Not more options than help the decision.
5. **Talk** until it feels right. When the human pushes back, take it
   seriously: research again if needed, and say so plainly if they are
   right.
6. **Play it back** now and then: a short summary of what is settled and
   what is open. Mark your guesses as guesses. End by asking what is wrong
   or missing.
7. **Questions only for real open points**, one at a time. Each question
   says what is happening, why it matters, the options, and your
   recommendation, with a small example when it helps. Never send a batch
   of questions to answer at once.

## Decisions

- A decision is the human's only when they said it. Write it as
  "decided, the human, YYYY-MM-DD".
- Everything you suggest stays "the lead's proposal" until the human
  confirms it.
- Settle principles: who decides what, and how the parts fit together.
  Leave deep details (exact fields, flags, timers, messages) to the slice
  spec that builds it, and say so instead of asking about them.

## How it ends

When the topic is settled, or the human wants to stop, suggest one ending
and let the human choose:

- **Park it.** Add one line to `backlog.md` at the project root, in its
  form: `- <the idea> (from brainstorm, YYYY-MM-DD)`.
- **Keep it as a design doc.** Write `docs/design/<topic>.md`: the idea,
  the options looked at, the decisions (marked as above), what is still
  open, and the sources. Its first lines say
  `Status: brainstorm, not approved.`
- **Turn it into slices.** Propose how to cut it: what each slice lets
  someone try, in what order. Then stop. Starting a slice is a separate
  step that the human starts.
- **Nothing to keep.** The chat was enough.

Only the first two write a file, and only after the human chooses them.
Commit the file on a branch, as the project's git policy says.
