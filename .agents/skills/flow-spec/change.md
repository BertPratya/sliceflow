# flow-spec: changing an approved spec

Use this when a spec is already approved and something in it must change:
a reviewer or worker found a gap, or the human changed their mind.

1. **Write down the reason** in one or two sentences. It goes with the approval.
   Set the spec's status to `changed` in `status.md`.
2. **Interview only what changes**, as in step 3 of `SKILL.md`. Include what
   the change affects: which requirements, flows, and tasks.
3. **Edit the spec** following the numbering rules in the spec template.
   Never renumber. Removed requirements stay as a "(removed in R…)" line.
4. **Run the spec checks** (`workflow/formats/document-checks.md`). Fix every error.
5. **Spec check**, if a requirement was added, changed, or removed: submit
   as in step 6 of `SKILL.md`, and tell the reviewer which requirements
   changed. It looks only at those and what they affect. Settle its
   questions with the human and edit the spec. A wording-only change (same
   meaning, no requirement added, changed, or removed) skips the check:
   set its status to `awaiting_approval` directly.
6. **Show the human:**
   - the reason
   - the diff against the last approved revision:
     `git diff --no-index history/spec/<last revision>.md spec.md`
   - the requirements added, changed, and removed, or "Wording only, no spec check"
   - the tasks that become stale
7. **The human approves**, as in step 7 of `SKILL.md`. Only when they have
   just told you to.
8. **Update what depends on it.** If the plan or tasks need changes, use
   `flow-plan`. They need a new approval too.

Done when: the new revision is approved, and the plan and tasks agree with it.
