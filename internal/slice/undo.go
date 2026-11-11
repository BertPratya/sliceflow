package slice

import "context"

// undoStep is one thing that slice create made, and how to remove it
// (R36).
type undoStep struct {
	// left is the R41 line of the thing, without its indent, e.g.
	// "worktree  <full path>". When it is "", the thing gets no R41
	// line, e.g. a file inside the worktree.
	left string
	// cleanup is the command that removes the thing by hand, for R41,
	// e.g. `git branch -D "slice/<name>"`.
	cleanup string
	// undo removes the thing.
	undo func(ctx context.Context) error
}

// worktreeStep is the step of the worktree at path, the full path, with
// its R41 lines: `worktree  <path>`, and
// `git worktree remove --force --force "<path>"`.
func worktreeStep(path string, undo func(context.Context) error) undoStep {
	return undoStep{
		left:    "worktree  " + path,
		cleanup: `git worktree remove --force --force "` + path + `"`,
		undo:    undo,
	}
}

// branchStep is the step of the branch of the slice name, slice/<name>,
// with its R41 lines: `branch    slice/<name>`, and
// `git branch -D "slice/<name>"`.
func branchStep(name string, undo func(context.Context) error) undoStep {
	return undoStep{
		left:    "branch    slice/" + name,
		cleanup: `git branch -D "slice/` + name + `"`,
		undo:    undo,
	}
}

// undoStack holds what slice create made, in the order it made it.
type undoStack []undoStep

// push adds step to the top of the stack.
func (s *undoStack) push(step undoStep) {
	*s = append(*s, step)
}

// unwind removes what was made, and returns the steps whose undo failed
// (R36, R41).
//
// Steps:
//  1. Go through the steps from the last pushed to the first.
//  2. Skip a step whose undo is nil. Run each other step's undo with
//     ctx. When it fails, keep the step, and go on with the next one.
//  3. Empty the stack, so a second call does nothing.
//  4. Return the kept steps, in the order they ran, or nil when every
//     undo worked.
//
// The errors of the undos are not returned: R41 does not show them.
func (s *undoStack) unwind(ctx context.Context) []undoStep {
	var failed []undoStep
	for i := len(*s) - 1; i >= 0; i-- {
		step := (*s)[i]
		if step.undo == nil {
			continue
		}
		if err := step.undo(ctx); err != nil {
			failed = append(failed, step)
		}
	}
	*s = nil
	return failed
}
