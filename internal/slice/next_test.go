package slice

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/BertPratya/sliceflow/internal/state"
	"github.com/BertPratya/sliceflow/internal/testrepo"
)

func TestSpecNext(t *testing.T) {
	rev := func(n int, result string) *state.Review {
		return &state.Review{Round: n, Reviewer: "placeholder", File: "reviews/spec-" + string(rune('0'+n)) + ".md", Result: result}
	}
	mk := func(status string, r *state.Review, refused *state.Counts, max int) state.State {
		return state.State{Phase: "planning", Spec: state.Doc{Status: status, Review: r, Refused: refused, MaxRound: max}}
	}
	fix := "fix spec.md, then: sliceflow spec submit"
	rows := []struct {
		name          string
		st            state.State
		changed, mrgd bool
		spec, next    string
	}{
		{"draft", mk("draft", nil, nil, 0), false, false, "draft", "write spec.md, then: sliceflow spec submit"},
		{"changes", mk("draft", rev(1, "changes"), nil, 1), false, false, "draft (review 1 found changes: reviews/spec-1.md)", fix},
		{"needs human", mk("draft", rev(2, "needs_human"), nil, 2), false, false, "draft (review 2 needs the human: reviews/spec-2.md)", "settle reviews/spec-2.md with the human, then: sliceflow spec submit"},
		{"submitted", mk("submitted", rev(1, "changes"), nil, 1), false, false, "submitted (review 2 not started)", "sliceflow spec review"},
		{"submitted first", mk("submitted", nil, nil, 0), false, false, "submitted (review 1 not started)", "sliceflow spec review"},
		{"in review", mk("in_review", rev(1, ""), nil, 1), false, false, "in_review (review 1, placeholder: reviews/spec-1.md)", "wait for review 1"},
		{"passed", mk("review_passed", rev(1, "passed"), nil, 1), false, false, "review_passed (review 1, placeholder)", "write the plan"},
		{"passed changed", mk("review_passed", rev(1, "passed"), nil, 1), true, false, "passed review 1, but spec.md was changed after it", "sliceflow spec submit"},
		{"no created", state.State{Phase: "unknown"}, false, false, "-", "-"},
		{"passed refused, unchanged", mk("review_passed", rev(1, "passed"), &state.Counts{Errors: 1}, 1), false, false, "review_passed (review 1, placeholder)", "write the plan"},
		{"merged", mk("review_passed", rev(1, "passed"), nil, 1), false, true, "review_passed (review 1, placeholder)", "-"},
		{"refused 1 warning", mk("draft", nil, &state.Counts{Warnings: 1}, 0), false, false, "draft (submit refused: 1 warning)", fix},
		{"refused 2 errors 1 warning", mk("draft", nil, &state.Counts{Errors: 2, Warnings: 1}, 0), false, false, "draft (submit refused: 2 errors, 1 warning)", fix},
		{"refused 1 error", mk("draft", rev(1, "changes"), &state.Counts{Errors: 1}, 1), false, false, "draft (submit refused: 1 error)", fix},
		{"refused after passed, changed", mk("review_passed", rev(1, "passed"), &state.Counts{Errors: 1}, 1), true, false, "passed review 1, but spec.md was changed after it (submit refused: 1 error)", fix},
		{"refused merged", mk("draft", nil, &state.Counts{Errors: 1}, 0), false, true, "draft", "-"},
	}
	for _, r := range rows {
		spec, next := specNext(r.st, r.changed, r.mrgd)
		if spec != r.spec || next != r.next {
			t.Errorf("%s: got %q / %q, want %q / %q", r.name, spec, next, r.spec, r.next)
		}
	}
}

func specCommit(t *testing.T, dir, text string) (file, commit string) {
	t.Helper()
	folder := filepath.Join(dir, "docs", "slices", "x")
	if err := os.MkdirAll(folder, 0o755); err != nil {
		t.Fatal(err)
	}
	file = filepath.Join(folder, "spec.md")
	if err := os.WriteFile(file, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	testrepo.Git(t, dir, "add", "--", "docs/slices/x/spec.md")
	testrepo.Git(t, dir, "commit", "-m", "spec")
	return file, gitOut(t, dir, "rev-parse", "HEAD")
}

func TestSpecChanged(t *testing.T) {
	ctx := context.Background()
	r := newRepo(t)
	file, commit := specCommit(t, r.Dir, "one\ntwo\n")
	got, err := specChanged(ctx, r.Dir, "x", commit)
	if err != nil || got {
		t.Errorf("no edit: %v %v", got, err)
	}
	if err := os.WriteFile(file, []byte("one\nthree\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err = specChanged(ctx, r.Dir, "x", commit)
	if err != nil || !got {
		t.Errorf("edit: %v %v", got, err)
	}
	got, err = specChanged(ctx, r.Dir, "x", "0123456789abcdef0123456789abcdef01234567")
	if err != nil || got {
		t.Errorf("missing commit: %v %v", got, err)
	}
}

func TestSpecChangedAutoCRLF(t *testing.T) {
	ctx := context.Background()
	r := newRepo(t)
	testrepo.Git(t, r.Dir, "config", "core.autocrlf", "true")
	file, commit := specCommit(t, r.Dir, "one\ntwo\n")
	if err := os.WriteFile(file, []byte("one\r\ntwo\r\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := specChanged(ctx, r.Dir, "x", commit)
	if err != nil || got {
		t.Errorf("crlf only: %v %v", got, err)
	}
}

func TestSpecChangedNoFolder(t *testing.T) {
	gone := filepath.Join(t.TempDir(), "gone")
	got, err := specChanged(context.Background(), gone, "x", "0123456789abcdef0123456789abcdef01234567")
	if err != nil || got {
		t.Errorf("no folder: %v %v", got, err)
	}
}
