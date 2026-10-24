package slice

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BertPratya/sliceflow/internal/eventlog"
	"github.com/BertPratya/sliceflow/internal/spec"
	"github.com/BertPratya/sliceflow/internal/state"
)

// Submission is the result of Submit. Problems are the lines to print
// (without the summary line). Refused is nil when the submit was accepted.
// Before is the spec's state before a refused submit, for R22's last
// sentence.
type Submission struct {
	Problems []spec.Problem
	Refused  *state.Counts
	Before   string
}

// Submit runs Flows 3 and 4: with the lock held, it checks the state (R20),
// reads spec.md (R19), checks it as spec validate --strict does, and saves
// spec.submitted (R21). It never starts the review. An empty name means the
// slice of the current branch (R12).
func Submit(ctx context.Context, r Repo, name string) (Submission, error) {
	if name == "" {
		n, err := currentSliceName(ctx, r, "spec submit")
		if err != nil {
			if ctx.Err() != nil {
				return Submission{}, InterruptedSave()
			}
			return Submission{}, err
		}
		name = n
	}
	wt, err := openSlice(ctx, r, name)
	if err != nil {
		if ctx.Err() != nil {
			return Submission{}, InterruptedSave()
		}
		return Submission{}, err
	}
	var sub Submission
	_, _, err = saveEvent(ctx, r, name, wt, saveOp{
		Event: "spec.submitted",
		Keep:  []string{"spec.md"},
		Prepare: func(st state.State) (prepared, error) {
			if st.Phase == "unknown" {
				return prepared{}, &Error{Message: fmt.Sprintf("slice %s has no slice.created event, so its state is unknown. Nothing was changed.", name)}
			}
			round := 0
			if st.Spec.Review != nil {
				round = st.Spec.Review.Round
			}
			switch st.Spec.Status {
			case "draft":
			case "submitted":
				return prepared{}, &Error{Message: "spec.md is submitted; review it with: sliceflow spec review. Nothing was changed."}
			case "in_review":
				return prepared{}, &Error{Message: fmt.Sprintf("spec.md is in_review: review %d has started, and has not finished. Nothing was changed.", round)}
			case "review_passed":
				changed, err := specChanged(ctx, wt, name, st.Spec.Review.Commit)
				if err != nil {
					return prepared{}, InterruptedSave()
				}
				if !changed {
					return prepared{}, &Error{Message: fmt.Sprintf("spec.md has not changed since review %d, which passed. Nothing was changed.", round)}
				}
			}
			data, err := os.ReadFile(filepath.Join(wt, "docs", "slices", name, "spec.md"))
			if os.IsNotExist(err) {
				return prepared{}, specNotFoundError(r, wt, name)
			}
			if err != nil {
				return prepared{}, err
			}
			problems := spec.Check(spec.Parse(data), name)
			c := state.Counts{}
			for _, p := range problems {
				switch p.Level {
				case spec.LevelError:
					c.Errors++
				case spec.LevelWarning:
					c.Warnings++
				}
			}
			sub = Submission{Problems: problems, Before: st.Spec.Status}
			result, msg := "accepted", name+": spec submitted"
			if c.Errors > 0 || c.Warnings > 0 {
				result, msg = "refused", name+": spec submit refused"
				sub.Refused = &c
			}
			return prepared{
				Data:    eventlog.Submitted{Result: result, Errors: c.Errors, Warnings: c.Warnings},
				Message: msg,
			}, nil
		},
	})
	if err != nil {
		return Submission{}, err
	}
	return sub, nil
}
