package slice

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/BertPratya/sliceflow/internal/eventlog"
	"github.com/BertPratya/sliceflow/internal/git"
	"github.com/BertPratya/sliceflow/internal/state"
)

// SliceStatus is what sliceflow knows about one slice. Title is empty when
// none was saved. Created is zero, and From is nil, when the log has no
// good slice.created event.
type SliceStatus struct {
	Slice, Title, Phase string
	From                *eventlog.From
	Created             time.Time
	Warnings            []Warning
	// State is the replayed state. SpecText and Next are R29's lines.
	// ChangedAfterReview is R30. Merged: the log was read from the base branch.
	State              state.State
	SpecText, Next     string
	ChangedAfterReview bool
	Merged             bool
}

func noSliceError(name string) error {
	return &Error{Message: fmt.Sprintf("no slice named %s. See all slices: sliceflow slice list", name)}
}

// Status returns the status of one slice. An empty name means the slice of
// the current branch.
func Status(ctx context.Context, r Repo, name string) (SliceStatus, error) {
	if name == "" {
		n, err := currentSliceName(ctx, r, "slice status")
		if err != nil {
			return SliceStatus{}, err
		}
		name = n
	}
	// R33: check the name before any git command uses it.
	if CheckName(name) != nil {
		return SliceStatus{}, noSliceError(name)
	}
	base, exists, w, err := baseWarning(ctx, r)
	if err != nil {
		return SliceStatus{}, err
	}
	st, found, err := readSlice(ctx, r, name, base, exists)
	if err != nil {
		return SliceStatus{}, err
	}
	if !found {
		return SliceStatus{}, noSliceError(name)
	}
	if w != nil {
		st.Warnings = append([]Warning{*w}, st.Warnings...)
	}
	return st, nil
}

// currentSliceName finds the name from the current branch (R31, R32).
func currentSliceName(ctx context.Context, r Repo, command string) (string, error) {
	out, err := git.Run(ctx, r.Dir, "symbolic-ref", "--quiet", "--short", "HEAD")
	branch := strings.TrimSpace(out)
	if err != nil {
		var ge *git.Error
		if !errors.As(err, &ge) || ge.ExitCode != 1 {
			return "", err
		}
		branch = ""
	}
	if name, ok := strings.CutPrefix(branch, "slice/"); ok && name != "" {
		return name, nil
	}
	where := "on no branch"
	if branch != "" {
		where = "on " + branch
	}
	return "", &Error{Message: fmt.Sprintf("not on a slice branch (you are %s).\nGive a name: sliceflow %s <name>\nSee all slices: sliceflow slice list", where, command)}
}

// baseWarning picks the base branch (R21) and says whether it exists. When
// it does not, w holds the warning of R23.
func baseWarning(ctx context.Context, r Repo) (base string, exists bool, w *Warning, err error) {
	base, source, err := BaseBranch(ctx, r, "")
	if err != nil {
		return "", false, nil, err
	}
	// Only an existing local branch is a base (as in checkCreate): check the
	// exact ref, so a tag or HEAD or main~1 never counts.
	exists, err = baseExists(ctx, r, base)
	if err != nil {
		return "", false, nil, err
	}
	if !exists {
		w = &Warning{Text: fmt.Sprintf("base branch %s does not exist (%s); merged slices are not shown", base, source)}
	}
	return base, exists, w, nil
}

// readSlice reads one slice from git (R24). name must pass CheckName. The
// bool is false when the slice does not exist.
func readSlice(ctx context.Context, r Repo, name, base string, baseExists bool) (SliceStatus, bool, error) {
	folder := path.Join("docs", "slices", name)
	logPath := path.Join(folder, ".events.jsonl")
	ref := ""
	merged := false
	branchRef := "refs/heads/slice/" + name
	ok, err := git.RefExists(ctx, r.Dir, branchRef)
	if err != nil {
		return SliceStatus{}, false, err
	}
	if ok {
		ref = branchRef
	} else if baseExists {
		_, err := git.Run(ctx, r.Dir, "cat-file", "-e", "refs/heads/"+base+":"+folder)
		var ge *git.Error
		switch {
		case err == nil:
			ref = "refs/heads/" + base
			merged = true
		case errors.As(err, &ge):
		default:
			return SliceStatus{}, false, err
		}
	}
	if ref == "" {
		return SliceStatus{}, false, nil
	}

	st := SliceStatus{Slice: name, Phase: "unknown"}
	data, found, err := git.Show(ctx, r.Dir, ref, logPath)
	if err != nil {
		return SliceStatus{}, false, err
	}
	if !found {
		st.Warnings = []Warning{{Slice: name, Text: ".events.jsonl not found"}}
		st.Merged = merged
		st.SpecText, st.Next = specNext(st.State, false, merged)
		return st, true, nil
	}
	events, warnings := eventlog.Parse(data)
	for _, text := range warnings {
		st.Warnings = append(st.Warnings, Warning{Slice: name, Text: text})
	}
	st.State, warnings = state.Replay(events)
	for _, text := range warnings {
		st.Warnings = append(st.Warnings, Warning{Slice: name, Text: text})
	}
	st.Phase = st.State.Phase
	st.Title = st.State.Title
	st.Merged = merged
	for _, e := range events {
		c, isCreated := e.Data.(eventlog.Created)
		if e.Event != "slice.created" || !isCreated {
			continue
		}
		from := c.From
		st.From = &from
		if t, err := time.Parse(eventlog.TimeLayout, e.At); err == nil {
			st.Created = t.UTC()
		}
		break
	}
	if !merged && st.State.Spec.Status == "review_passed" && st.State.Spec.Review != nil {
		wt, ok, err := findWorktree(ctx, r, name)
		if err != nil {
			return SliceStatus{}, false, err
		}
		if ok {
			st.ChangedAfterReview, _ = specChanged(ctx, wt, name, st.State.Spec.Review.Commit)
		}
	}
	st.SpecText, st.Next = specNext(st.State, st.ChangedAfterReview, merged)
	return st, true, nil
}
