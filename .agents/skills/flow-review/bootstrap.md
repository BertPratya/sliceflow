# flow-review: bootstrap mode

Use this while `flow` does not exist yet. After each step, add a dated
Log line to the slice's `status.md` (format: `workflow/formats/status.md`).

## `flow review start <kind> <slice>/<target>`

1. Find the round: 1 + the number of earlier reviews of the same kind and
   target in `reviews/`.
2. Copy `workflow/templates/review.md` to `reviews/<kind>-<target>-<round>.md`
   (integration: `reviews/integration-<round>.md`).
3. Fill the header: review, slice, kind, target, round, reviewer
   (role and tool), and the title. For a task, put the commits to review
   in `target`.

## The reviewer prompt

`flow brief` does not exist yet. Tell the reviewer what to read instead:

    You are the <role>. Read workflow/roles/<role>.md and follow it.
    Work only in <absolute path of the slice worktree>. All paths are relative to it.
    This is a <kind> review. Your review file: <path>.
    <task review: Review commits <hashes> of task <NN>.>
    <integration review: Review the diff of <slice branch> against <base branch>.>
    <plan review: The track is <track>.>
    <spec check of a change: Only R<n>, R<n> changed. Check them and what they affect.>

## After the reviewer ends

1. It should end with `RESULT: submitted`. Check the review file is filled.
2. Set `verdict` and `submitted_at` (now, UTC) in the header, as the template says.
3. Record the verdict and every finding in `status.md`.
4. Spec or plan review: set the document's status, as `workflow/formats/lifecycle.md`
   → Documents says: no questions or findings → `awaiting_approval`; otherwise → `draft`.
