package slice

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/BertPratya/sliceflow/internal/eventlog"
	"github.com/BertPratya/sliceflow/internal/git"
	"github.com/BertPratya/sliceflow/internal/review"
	"github.com/BertPratya/sliceflow/internal/state"
)

// reviewer is the one reviewer of this slice.
var reviewer = review.Placeholder()

// Step is one step of a review, for the caller to show. Kind is "started"
// or "finished". Result is set for "finished".
type Step struct {
	Kind                   string
	Round                  int
	Reviewer, File, Result string
}

// savedTail is R10's last sentence once review.started is saved.
const savedTail = "The steps before it were saved; spec.md is in_review."

// ReviewSpec runs Flow 5: it saves review.started, runs the reviewer with
// no lock held, then saves review.finished. It calls step after each save.
// An empty name means the slice of the current branch (R12).
func ReviewSpec(ctx context.Context, r Repo, name string, step func(Step)) error {
	if name == "" {
		n, err := currentSliceName(ctx, r, "spec review")
		if err != nil {
			if ctx.Err() != nil {
				return InterruptedSave()
			}
			return err
		}
		name = n
	}
	wt, err := openSlice(ctx, r, name)
	if err != nil {
		if ctx.Err() != nil {
			return InterruptedSave()
		}
		return err
	}

	var round int
	var commit string
	file := ""
	_, _, err = saveEvent(ctx, r, name, wt, saveOp{
		Event: "review.started",
		Prepare: func(st state.State) (prepared, error) {
			if st.Phase == "unknown" {
				return prepared{}, &Error{Message: fmt.Sprintf("slice %s has no slice.created event, so its state is unknown. Nothing was changed.", name)}
			}
			n := 0
			if st.Spec.Review != nil {
				n = st.Spec.Review.Round
			}
			switch st.Spec.Status {
			case "draft":
				return prepared{}, &Error{Message: "spec.md is a draft; submit it first: sliceflow spec submit. Nothing was changed."}
			case "in_review":
				return prepared{}, &Error{Message: fmt.Sprintf("spec.md is in_review: review %d has started, and has not finished. Nothing was changed.", n)}
			case "review_passed":
				return prepared{}, &Error{Message: fmt.Sprintf("spec.md passed review %d; there is nothing to review. Nothing was changed.", n)}
			}
			round = st.Spec.MaxRound + 1
			file = fmt.Sprintf("reviews/spec-%d.md", round)
			if exists(filepath.Join(wt, "docs", "slices", name, filepath.FromSlash(file))) {
				return prepared{}, &Error{Message: fmt.Sprintf("docs/slices/%s/%s already exists, but review %d has not started. Move it away, then review again. Nothing was changed.", name, file, round)}
			}
			c, err := submitCommit(ctx, wt, name, st.Spec.SubmitSeq)
			if err != nil {
				if ctx.Err() != nil {
					return prepared{}, InterruptedSave()
				}
				return prepared{}, err
			}
			commit = c
			return prepared{
				Data:    eventlog.ReviewStarted{Doc: "spec.md", Round: round, Reviewer: reviewer.Name(), File: file, Commit: commit},
				Files:   map[string][]byte{file: review.FrontMatter(name, round, reviewer.Name())},
				Message: fmt.Sprintf("%s: review %d started", name, round),
			}, nil
		},
	})
	if err != nil {
		return err
	}
	step(Step{Kind: "started", Round: round, Reviewer: reviewer.Name(), File: file})

	full := filepath.Join(wt, "docs", "slices", name, filepath.FromSlash(file))
	err = reviewer.Review(ctx, review.Request{Slice: name, Doc: "spec.md", File: full, Commit: commit, Worktree: wt, Round: round})
	if ctx.Err() != nil {
		return saveError("interrupted", savedTail, nil)
	}
	if err != nil {
		return &Error{Message: fmt.Sprintf("the reviewer failed: %v. %s", err, savedTail)}
	}
	text, err := os.ReadFile(full)
	if err != nil {
		return &Error{Message: fmt.Sprintf("could not read %s: %v. %s", file, err, savedTail)}
	}
	result := review.ReadResult(text)
	word := "passed"
	if result != "passed" {
		word = "needs the human"
	}

	_, _, err = saveEvent(ctx, r, name, wt, saveOp{
		Event: "review.finished",
		Tail:  savedTail,
		Prepare: func(st state.State) (prepared, error) {
			if st.Spec.Status != "in_review" || st.Spec.Review == nil || st.Spec.Review.Round != round {
				return prepared{}, &Error{Message: fmt.Sprintf("another sliceflow command changed slice %s first; spec.md is now %s. The steps before it were saved.", name, st.Spec.Status)}
			}
			return prepared{
				Data:    eventlog.ReviewFinished{Doc: "spec.md", Round: round, Result: result},
				Files:   map[string][]byte{file: text},
				Message: fmt.Sprintf("%s: review %d %s", name, round, word),
			}, nil
		},
	})
	if err != nil {
		return err
	}
	step(Step{Kind: "finished", Round: round, Reviewer: reviewer.Name(), File: file, Result: result})
	return nil
}

// submitCommit finds the commit that added the spec.submitted with this
// seq: it reads the log of each commit that changed it, newest first, and
// takes the oldest one whose log still holds that event.
func submitCommit(ctx context.Context, wt, name string, seq int) (string, error) {
	logRel := path.Join("docs", "slices", name, ".events.jsonl")
	out, err := git.Run(ctx, wt, "log", "--format=%H", "refs/heads/slice/"+name, "--", logRel)
	if err != nil {
		return "", err
	}
	found := ""
	for _, id := range splitLines(out) {
		data, ok, err := git.Show(ctx, wt, id, logRel)
		if err != nil {
			return "", err
		}
		if !ok || !holdsSubmit(data, seq) {
			break
		}
		found = id
	}
	if found == "" {
		return "", fmt.Errorf("internal error: no commit holds spec.submitted with seq %d", seq)
	}
	return found, nil
}

func holdsSubmit(data []byte, seq int) bool {
	events, _ := eventlog.Parse(data)
	for _, e := range events {
		if e.Seq == seq && e.Event == "spec.submitted" {
			return true
		}
	}
	return false
}

func splitLines(s string) []string {
	var out []string
	for _, l := range strings.Split(s, "\n") {
		if l = strings.TrimSpace(l); l != "" {
			out = append(out, l)
		}
	}
	return out
}
