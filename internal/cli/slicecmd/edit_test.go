package slicecmd

import (
	"context"
	"strings"
	"testing"

	"github.com/BertPratya/sliceflow/internal/git"
	"github.com/BertPratya/sliceflow/internal/slice"
)

// Each case prints an error of slice.Edit that is hard to make happen
// with a real repository, and checks the exact message and exit code.
func TestPrintEditError(t *testing.T) {
	log := slice.Leftover{Kind: "log", Name: "/home/you/shop/docs/slices/login/.events.jsonl", Top: "/home/you/shop"}
	staged := slice.Leftover{Kind: "staged", Name: "docs/slices/login/.events.jsonl", Top: "/home/you/shop"}
	refused := &git.Error{ExitCode: 1, Stderr: "commit refused\n"}

	tests := []struct {
		about string
		err   error
		want  string
	}{
		{"the log and the staging area could not be put back",
			&slice.StepError{Step: "make the commit", Err: refused, Left: []slice.Leftover{staged, log}},
			"Error: git could not make the commit.\n" +
				"git said:\n" +
				"  commit refused\n" +
				"Some of what sliceflow made could not be removed:\n" +
				"  log       /home/you/shop/docs/slices/login/.events.jsonl\n" +
				"  staged    docs/slices/login/.events.jsonl\n" +
				"To clean up:\n" +
				"  run: git -C \"/home/you/shop\" restore --staged --worktree -- \"docs/slices/login/.events.jsonl\"\n"},
		{"only the log could not be put back",
			&slice.StepError{Step: "make the commit", Err: context.Canceled, Left: []slice.Leftover{log}},
			"Stopped.\n" +
				"Some of what sliceflow made could not be removed:\n" +
				"  log       /home/you/shop/docs/slices/login/.events.jsonl\n" +
				"To clean up:\n" +
				"  run: git -C \"/home/you/shop\" restore --staged --worktree -- \"docs/slices/login/.events.jsonl\"\n"},
		{"git failed after the commit",
			&slice.CommittedError{Name: "login", Base: "main", Done: "edited", Err: refused},
			"Error: git could not finish the commit. Slice login was edited on branch main.\n" +
				"git said:\n" +
				"  commit refused\n"},
		{"stopped during a check",
			&slice.CheckError{Check: "find the project", Err: context.Canceled},
			"Stopped. Nothing was changed.\n"},
	}
	for _, tt := range tests {
		t.Run(tt.about, func(t *testing.T) {
			var out strings.Builder
			if code := printEditError(&out, tt.err); code != 1 {
				t.Errorf("printEditError returned %d, want 1", code)
			}
			if out.String() != tt.want {
				t.Errorf("printEditError printed:\n%s\nwant:\n%s", out.String(), tt.want)
			}
		})
	}
}
