package slice

import (
	"context"
	"errors"
	"slices"
	"testing"
)

// recordStep returns a step named name whose undo adds name to ran, and
// returns fail.
func recordStep(name string, ran *[]string, fail error) undoStep {
	return undoStep{
		left:    name,
		cleanup: "clean " + name,
		undo: func(context.Context) error {
			*ran = append(*ran, name)
			return fail
		},
	}
}

// leftNames returns the left of each step, in order.
func leftNames(steps []undoStep) []string {
	var names []string
	for _, step := range steps {
		names = append(names, step.left)
	}
	return names
}

// unwind runs the steps last pushed first (R36), and returns nothing when
// every undo works.
func TestUndoStackOrder(t *testing.T) {
	var ran []string
	var stack undoStack
	stack.push(recordStep("a", &ran, nil))
	stack.push(recordStep("b", &ran, nil))
	stack.push(recordStep("c", &ran, nil))

	left := stack.unwind(context.Background())

	if want := []string{"c", "b", "a"}; !slices.Equal(ran, want) {
		t.Errorf("unwind ran %q, want %q", ran, want)
	}
	if len(left) != 0 {
		t.Errorf("unwind returned %q, want nothing", leftNames(left))
	}
}

// When one undo fails, unwind still runs the steps after it, and returns
// only the step that failed (R36, R41).
func TestUndoStackKeepsGoing(t *testing.T) {
	var ran []string
	var stack undoStack
	stack.push(recordStep("a", &ran, nil))
	stack.push(recordStep("b", &ran, errors.New("could not remove b")))

	left := stack.unwind(context.Background())

	if want := []string{"b", "a"}; !slices.Equal(ran, want) {
		t.Errorf("unwind ran %q, want %q", ran, want)
	}
	if got, want := leftNames(left), []string{"b"}; !slices.Equal(got, want) {
		t.Errorf("unwind returned %q, want %q", got, want)
	}
}

// unwind returns the failed steps in the order it ran them, so the
// worktree, pushed after the branch, comes first (R41).
func TestUndoStackFailedOrder(t *testing.T) {
	var ran []string
	var stack undoStack
	failed := errors.New("could not remove")
	stack.push(recordStep("branch", &ran, failed))
	stack.push(recordStep("worktree", &ran, failed))
	stack.push(recordStep("file", &ran, nil))

	left := stack.unwind(context.Background())

	if got, want := leftNames(left), []string{"worktree", "branch"}; !slices.Equal(got, want) {
		t.Errorf("unwind returned %q, want %q", got, want)
	}
}

// unwind skips a step with no undo, and keeps going.
func TestUndoStackNilUndo(t *testing.T) {
	var ran []string
	var stack undoStack
	stack.push(recordStep("a", &ran, nil))
	stack.push(undoStep{left: "no undo"})

	left := stack.unwind(context.Background())

	if want := []string{"a"}; !slices.Equal(ran, want) {
		t.Errorf("unwind ran %q, want %q", ran, want)
	}
	if len(left) != 0 {
		t.Errorf("unwind returned %q, want nothing", leftNames(left))
	}
}

// unwind empties the stack, so a second call runs nothing.
func TestUndoStackUnwindOnce(t *testing.T) {
	var ran []string
	var stack undoStack
	stack.push(recordStep("a", &ran, nil))

	stack.unwind(context.Background())
	stack.unwind(context.Background())

	if want := []string{"a"}; !slices.Equal(ran, want) {
		t.Errorf("two unwinds ran %q, want %q", ran, want)
	}
}

// The constructors build the R41 lines of S12 (R41).
func TestStepConstructors(t *testing.T) {
	worktree := worktreeStep("/home/you/.sliceflow/worktrees/shop/card-payment", nil)
	if want := "worktree  /home/you/.sliceflow/worktrees/shop/card-payment"; worktree.left != want {
		t.Errorf("worktree left = %q, want %q", worktree.left, want)
	}
	if want := `git worktree remove --force --force "/home/you/.sliceflow/worktrees/shop/card-payment"`; worktree.cleanup != want {
		t.Errorf("worktree cleanup = %q, want %q", worktree.cleanup, want)
	}

	branch := branchStep("card-payment", nil)
	if want := "branch    slice/card-payment"; branch.left != want {
		t.Errorf("branch left = %q, want %q", branch.left, want)
	}
	if want := `git branch -D "slice/card-payment"`; branch.cleanup != want {
		t.Errorf("branch cleanup = %q, want %q", branch.cleanup, want)
	}
}
