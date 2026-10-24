package slice

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/BertPratya/sliceflow/internal/git"
	"github.com/BertPratya/sliceflow/internal/state"
)

// specChanged says whether spec.md in the worktree differs from spec.md in
// the reviewed commit (R30). git compares through its line-end rules. Any
// failure other than a difference, e.g. a commit that is not in the repo,
// or git that cannot start in the folder, means not compared: false, nil.
func specChanged(ctx context.Context, worktree, name, commit string) (bool, error) {
	if worktree == "" || commit == "" {
		return false, nil
	}
	_, err := git.Run(ctx, worktree, "diff", "--quiet", commit, "--", path.Join("docs", "slices", name, "spec.md"))
	if err == nil {
		return false, nil
	}
	var ge *git.Error
	if errors.As(err, &ge) {
		return ge.ExitCode == 1, nil
	}
	// Git could not start, e.g. the folder is gone: not compared.
	return false, nil
}

func plural(n int, word string) string {
	if n == 1 {
		return fmt.Sprintf("1 %s", word)
	}
	return fmt.Sprintf("%d %ss", n, word)
}

// specNext works out the Spec: and Next: texts of R29.
func specNext(st state.State, changed, merged bool) (spec, next string) {
	d := st.Spec
	if d.Status == "" {
		return "-", "-"
	}
	round, reviewer, file := 0, "", ""
	if d.Review != nil {
		round, reviewer, file = d.Review.Round, d.Review.Reviewer, d.Review.File
	}
	changed = changed && d.Status == "review_passed"
	const fix = "fix spec.md, then: sliceflow spec submit"
	switch d.Status {
	case "draft":
		switch {
		case d.Review != nil && d.Review.Result == "changes":
			spec = fmt.Sprintf("draft (review %d found changes: %s)", round, file)
			next = fix
		case d.Review != nil && d.Review.Result == "needs_human":
			spec = fmt.Sprintf("draft (review %d needs the human: %s)", round, file)
			next = fmt.Sprintf("settle %s with the human, then: sliceflow spec submit", file)
		default:
			spec = "draft"
			next = "write spec.md, then: sliceflow spec submit"
		}
	case "submitted":
		spec = fmt.Sprintf("submitted (review %d not started)", d.MaxRound+1)
		next = "sliceflow spec review"
	case "in_review":
		spec = fmt.Sprintf("in_review (review %d, %s: %s)", round, reviewer, file)
		next = fmt.Sprintf("wait for review %d", round)
	case "review_passed":
		if changed {
			spec = fmt.Sprintf("passed review %d, but spec.md was changed after it", round)
			next = "sliceflow spec submit"
		} else {
			spec = fmt.Sprintf("review_passed (review %d, %s)", round, reviewer)
			next = "write the plan"
		}
	}
	if c := d.Refused; c != nil && !merged && (d.Status == "draft" || changed) {
		if d.Status == "draft" {
			spec = "draft"
		}
		var parts []string
		if c.Errors > 0 {
			parts = append(parts, plural(c.Errors, "error"))
		}
		if c.Warnings > 0 {
			parts = append(parts, plural(c.Warnings, "warning"))
		}
		spec += " (submit refused: " + strings.Join(parts, ", ") + ")"
		next = fix
	}
	if merged {
		next = "-"
	}
	return spec, next
}
