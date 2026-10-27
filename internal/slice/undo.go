package slice

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"strings"
	"syscall"

	"github.com/BertPratya/sliceflow/internal/git"
)

// undoStep is one thing sliceflow made, and how to remove it again.
type undoStep struct {
	what string       // short text for "Could not remove:", e.g. "branch slice/x"
	do   func() error // removes it
}

// creation keeps the undo steps of one run of Create (R18).
type creation struct {
	ctx  context.Context
	name string
	undo []undoStep // in the order they were made
}

// add records an undo step for something that was just made.
func (c *creation) add(what string, do func() error) {
	c.undo = append(c.undo, undoStep{what: what, do: do})
}

// check is called after each step with the step's error (nil if it
// worked). It returns nil when the step worked and no signal came. Else it
// undoes everything and returns the message of R18. A signal counts as a
// failed step with the reason "interrupted", even when the step worked.
func (c *creation) check(err error) error {
	if err == nil && c.ctx.Err() == nil {
		return nil
	}
	reason := "interrupted"
	var details []string
	if c.ctx.Err() == nil {
		reason = err.Error()
		var ce *git.CommandError
		if errors.As(err, &ce) {
			// Reason and details both come from git's error, so they always
			// match, even if the error was wrapped on the way.
			reason = ce.Error()
			details = ce.OtherLines()
		}
	}
	return createFailed(c.name, reason, c.rollback(), details)
}

// rollback runs the undo steps, newest first. It does not use the context,
// so a signal cannot cut it short. A step that fails does not stop the
// others. It returns the text of the steps that failed.
func (c *creation) rollback() []string {
	// A second Ctrl-C goes to the whole process group and could kill the
	// git commands of the undo. Ignored signals stay ignored in child
	// processes, so we ignore them from here on. A later long-running
	// server that calls Create must revisit this.
	signal.Ignore(os.Interrupt, syscall.SIGTERM)

	var notRemoved []string
	for _, u := range slices.Backward(c.undo) {
		// A thing that is already gone is not a failure.
		if err := u.do(); err != nil && !errors.Is(err, os.ErrNotExist) {
			notRemoved = append(notRemoved, u.what)
		}
	}
	return notRemoved
}

// createFailed builds the message of R18. The first line says why it
// failed, and either that everything was removed, or what could not be.
// Each detail follows on its own line, indented by 2 spaces.
func createFailed(name, reason string, notRemoved, details []string) error {
	msg := fmt.Sprintf("could not create slice %s: %s.", name, reason)
	if len(notRemoved) == 0 {
		msg += " Everything sliceflow made was removed."
	} else {
		msg += " Could not remove: " + strings.Join(notRemoved, ", ") + "."
	}
	for _, d := range details {
		msg += "\n  " + d
	}
	return errors.New(msg)
}

// writeGitignore makes the folder .sliceflow in the main checkout, and
// writes .sliceflow/.gitignore with the one line "*" (R6), unless the file
// is already there. An existing file is left as it is. It reports whether
// the folder and the file were new, so that an undo removes only those.
func writeGitignore(mainRoot string) (madeDir, madeFile bool, err error) {
	dir := filepath.Join(mainRoot, ".sliceflow")
	if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		madeDir = true
	} else if err != nil {
		return false, false, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return false, false, err
	}
	// O_EXCL makes the open fail when the file exists, so we never
	// overwrite it.
	f, err := os.OpenFile(filepath.Join(dir, ".gitignore"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if errors.Is(err, os.ErrExist) {
		return madeDir, false, nil
	}
	if err != nil {
		return madeDir, false, err
	}
	if _, err := f.WriteString("*\n"); err != nil {
		f.Close()
		return madeDir, true, err
	}
	return madeDir, true, f.Close()
}
