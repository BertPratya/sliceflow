package slice

import (
	"context"
	"errors"
	"slices"
	"testing"
)

// The tests of runSteps use fake steps. Each fake step writes down when
// its Do and its Undo run, so a test can check what ran, and in which
// order.

var (
	errDo   = errors.New("the step failed")
	errUndo = errors.New("the undo failed")
)

// All steps work: runSteps returns nil, and no Undo runs.
func TestRunStepsAllWork(t *testing.T) {
	var r recorder
	steps := []Step{r.step("a", nil, nil), r.step("b", nil, nil), r.step("c", nil, nil)}

	err := runSteps(context.Background(), steps)
	if err != nil {
		t.Fatalf("runSteps returned the error %v, want none", err)
	}
	r.want(t, "do a", "do b", "do c")
}

// The third step fails: it and the steps before it are undone, newest
// first. The failed step's own Undo runs too, because its Do may have
// made part of its thing. The steps after it do not run at all.
func TestRunStepsUndoesNewestFirst(t *testing.T) {
	var r recorder
	steps := []Step{
		r.step("a", nil, nil),
		r.step("b", nil, nil),
		r.step("c", errDo, nil),
		r.step("d", nil, nil),
	}

	err := runSteps(context.Background(), steps)
	stepErr := wantStepError(t, err)
	if stepErr.Step != "c" {
		t.Errorf("the error names the step %q, want %q", stepErr.Step, "c")
	}
	if len(stepErr.Left) != 0 {
		t.Errorf("the error lists the leftovers %v, want none", stepErr.Left)
	}
	r.want(t, "do a", "do b", "do c", "undo c", "undo b", "undo a")
}

// The first step fails: only its own Undo runs.
func TestRunStepsFirstStepFails(t *testing.T) {
	var r recorder
	steps := []Step{r.step("a", errDo, nil), r.step("b", nil, nil)}

	err := runSteps(context.Background(), steps)
	stepErr := wantStepError(t, err)
	if stepErr.Step != "a" {
		t.Errorf("the error names the step %q, want %q", stepErr.Step, "a")
	}
	if len(stepErr.Left) != 0 {
		t.Errorf("the error lists the leftovers %v, want none", stepErr.Left)
	}
	r.want(t, "do a", "undo a")
}

// A step with no Undo is skipped, and the steps before it are still
// undone.
func TestRunStepsSkipsANilUndo(t *testing.T) {
	var r recorder
	noUndo := r.step("b", nil, nil)
	noUndo.Undo = nil
	steps := []Step{r.step("a", nil, nil), noUndo, r.step("c", errDo, nil)}

	err := runSteps(context.Background(), steps)
	wantStepError(t, err)
	r.want(t, "do a", "do b", "do c", "undo c", "undo a")
}

// An Undo fails: the steps before it are still undone, and its Left is
// listed.
func TestRunStepsKeepsGoingWhenAnUndoFails(t *testing.T) {
	var r recorder
	steps := []Step{r.step("a", nil, nil), r.step("b", nil, errUndo), r.step("c", errDo, nil)}

	err := runSteps(context.Background(), steps)
	stepErr := wantStepError(t, err)
	want := []Leftover{{Kind: "folder", Name: "b"}}
	if !slices.Equal(stepErr.Left, want) {
		t.Errorf("the error lists the leftovers %v, want %v", stepErr.Left, want)
	}
	r.want(t, "do a", "do b", "do c", "undo c", "undo b", "undo a")
}

// Two Undos fail: both are listed, newest first.
func TestRunStepsListsEachLeftoverNewestFirst(t *testing.T) {
	var r recorder
	steps := []Step{r.step("a", nil, errUndo), r.step("b", nil, errUndo), r.step("c", errDo, nil)}

	err := runSteps(context.Background(), steps)
	stepErr := wantStepError(t, err)
	want := []Leftover{{Kind: "folder", Name: "b"}, {Kind: "folder", Name: "a"}}
	if !slices.Equal(stepErr.Left, want) {
		t.Errorf("the error lists the leftovers %v, want %v", stepErr.Left, want)
	}
}

// Two Undos fail with the same Left: it is listed once.
func TestRunStepsListsTheSameLeftoverOnce(t *testing.T) {
	var r recorder
	a := r.step("a", nil, errUndo)
	b := r.step("b", nil, errUndo)
	b.Left = a.Left
	steps := []Step{a, b, r.step("c", errDo, nil)}

	err := runSteps(context.Background(), steps)
	stepErr := wantStepError(t, err)
	want := []Leftover{{Kind: "folder", Name: "a"}}
	if !slices.Equal(stepErr.Left, want) {
		t.Errorf("the error lists the leftovers %v, want %v", stepErr.Left, want)
	}
}

// The error holds why the step failed, so errors.Is finds it.
func TestRunStepsErrorHoldsTheCause(t *testing.T) {
	var r recorder
	steps := []Step{r.step("a", nil, nil), r.step("b", errDo, nil)}

	err := runSteps(context.Background(), steps)
	if !errors.Is(err, errDo) {
		t.Errorf("runSteps returned %v, want an error that holds %v", err, errDo)
	}
}

// Each Do gets the context that runSteps was given.
func TestRunStepsPassesTheContext(t *testing.T) {
	type key struct{}
	ctx := context.WithValue(context.Background(), key{}, "this run")
	var got any
	steps := []Step{{
		Name: "a",
		Do: func(ctx context.Context) error {
			got = ctx.Value(key{})
			return nil
		},
	}}

	if err := runSteps(ctx, steps); err != nil {
		t.Fatalf("runSteps returned the error %v, want none", err)
	}
	if got != "this run" {
		t.Errorf("Do got a context with the value %v, want %q", got, "this run")
	}
}

// recorder writes down the calls of the fake steps it makes, in order.
type recorder struct {
	calls []string
}

// step returns a fake step called name. Its Do writes "do <name>" and
// returns doErr; its Undo writes "undo <name>" and returns undoErr. Its
// Left is a folder called name.
func (r *recorder) step(name string, doErr, undoErr error) Step {
	return Step{
		Name: name,
		Do: func(ctx context.Context) error {
			r.calls = append(r.calls, "do "+name)
			return doErr
		},
		Undo: func() error {
			r.calls = append(r.calls, "undo "+name)
			return undoErr
		},
		Left: Leftover{Kind: "folder", Name: name},
	}
}

// want fails the test unless the calls were exactly want, in that order.
func (r *recorder) want(t *testing.T, want ...string) {
	t.Helper()
	if !slices.Equal(r.calls, want) {
		t.Errorf("the calls were %q, want %q", r.calls, want)
	}
}

// wantStepError fails the test now unless err is a *StepError, and
// returns it.
func wantStepError(t *testing.T, err error) *StepError {
	t.Helper()
	var stepErr *StepError
	if !errors.As(err, &stepErr) {
		t.Fatalf("runSteps returned %v (%T), want a *StepError", err, err)
	}
	return stepErr
}
