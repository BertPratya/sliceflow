package slicecmd

import (
	"context"
	"errors"
	"io/fs"
	"strings"
	"testing"

	"github.com/BertPratya/sliceflow/internal/eventlog"
	"github.com/BertPratya/sliceflow/internal/git"
	"github.com/BertPratya/sliceflow/internal/slice"
)

// Each case prints an error that slice.Create can return, and checks the
// exact message and the exit code. These are the cases that are hard to
// make happen with a real repository; the others are tested in e2e.
func TestPrintCreateError(t *testing.T) {
	writeFailed := &fs.PathError{Op: "open", Path: "/wt/docs/slices/login/spec.md", Err: fs.ErrPermission}
	lockFailed := &fs.PathError{Op: "mkdir", Path: "/wt/.git/sliceflow/locks", Err: errors.New("not a directory")}
	folder := slice.Leftover{Kind: "folder", Name: "/home/you/shop/docs/slices/login"}
	staged := slice.Leftover{Kind: "staged", Name: "docs/slices/login/", Top: "/home/you/shop"}

	tests := []struct {
		about string
		err   error
		code  int
		want  string
	}{
		{"a file that cannot be written",
			&slice.StepError{Step: "write /wt/docs/slices/login/spec.md", Err: writeFailed}, 1,
			"Error: could not write /wt/docs/slices/login/spec.md. The slice was not created. Nothing was changed.\n" +
				"The system said:\n" +
				"  permission denied\n"},
		{"the lock of the log cannot be taken",
			&slice.StepError{Step: "write /wt/docs/slices/login/.events.jsonl",
				Err: &eventlog.LockError{Path: "/wt/.git/sliceflow/locks/login.lock", Err: lockFailed}}, 1,
			"Error: could not write /wt/.git/sliceflow/locks/login.lock. The slice was not created. Nothing was changed.\n" +
				"The system said:\n" +
				"  not a directory\n"},
		{"stopped while waiting for the lock",
			&slice.StepError{Step: "write /wt/docs/slices/login/.events.jsonl",
				Err: &eventlog.LockError{Path: "/wt/.git/sliceflow/locks/login.lock", Err: context.Canceled}}, 1,
			"Stopped. The slice was not created. Nothing was changed.\n"},
		{"git printed nothing",
			&slice.StepError{Step: "make the commit", Err: &git.Error{ExitCode: 1, Stderr: ""}}, 1,
			"Error: git could not make the commit. The slice was not created. Nothing was changed.\n" +
				"git printed no message.\n"},
		{"git's empty lines stay empty, and Windows line ends are kept out",
			&slice.StepError{Step: "make the commit", Err: &git.Error{ExitCode: 1, Stderr: "first\r\n\r\nthird\r\n"}}, 1,
			"Error: git could not make the commit. The slice was not created. Nothing was changed.\n" +
				"git said:\n" +
				"  first\n" +
				"\n" +
				"  third\n"},
		{"the folder comes before the staged files",
			&slice.StepError{Step: "make the commit", Err: &git.Error{ExitCode: 1, Stderr: "commit refused\n"},
				Left: []slice.Leftover{staged, folder}}, 1,
			"Error: git could not make the commit. The slice was not created.\n" +
				"git said:\n" +
				"  commit refused\n" +
				"Some of what sliceflow made could not be removed:\n" +
				"  folder    /home/you/shop/docs/slices/login\n" +
				"  staged    docs/slices/login/\n" +
				"To clean up:\n" +
				"  delete the folder /home/you/shop/docs/slices/login\n" +
				"  run: git -C \"/home/you/shop\" reset -- \"docs/slices/login\"\n"},
		{"stopped, with something left",
			&slice.StepError{Step: "make the commit", Err: context.Canceled, Left: []slice.Leftover{folder}}, 1,
			"Stopped. The slice was not created.\n" +
				"Some of what sliceflow made could not be removed:\n" +
				"  folder    /home/you/shop/docs/slices/login\n" +
				"To clean up:\n" +
				"  delete the folder /home/you/shop/docs/slices/login\n"},
		{"stopped after the commit",
			&slice.CommittedError{Name: "login", Base: "main", Err: context.Canceled}, 1,
			"Stopped after the commit. Slice login was created on branch main.\n"},
		{"git failed after the commit",
			&slice.CommittedError{Name: "login", Base: "main", Err: &git.Error{ExitCode: 1, Stderr: "something broke\n"}}, 1,
			"Error: git could not finish the commit. Slice login was created on branch main.\n" +
				"git said:\n" +
				"  something broke\n"},
		{"stopped during a check",
			&slice.CheckError{Check: "find the project", Err: context.Canceled}, 1,
			"Stopped. The slice was not created. Nothing was changed.\n"},
		{"git not found during a check",
			&slice.CheckError{Check: "find the project", Err: git.ErrNotFound}, 1,
			"Error: git was not found. Nothing was changed.\nInstall git, then try again.\n"},
		{"a check that fails without git",
			&slice.CheckError{Check: "find the home folder", Err: errors.New("$HOME is not defined")}, 1,
			"Error: could not find the home folder. Nothing was changed.\n" +
				"The system said:\n" +
				"  $HOME is not defined\n"},
		{"an error with no message of its own",
			errors.New("something new"), 1,
			"Error: something new\n"},
	}
	for _, tt := range tests {
		t.Run(tt.about, func(t *testing.T) {
			var out strings.Builder
			code := printCreateError(&out, tt.err)
			if code != tt.code {
				t.Errorf("printCreateError returned %d, want %d", code, tt.code)
			}
			if out.String() != tt.want {
				t.Errorf("printCreateError printed:\n%s\nwant:\n%s", out.String(), tt.want)
			}
		})
	}
}
