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
	"time"

	"github.com/BertPratya/sliceflow/internal/eventlog"
	"github.com/BertPratya/sliceflow/internal/git"
	"github.com/BertPratya/sliceflow/templates"
)

// CreateOptions is what Create needs to make a slice.
type CreateOptions struct {
	Dir   string    // the folder the command runs in
	Name  string    // the slice name
	From  string    // the --from branch; empty if not given
	Title string    // the --title text; empty if not given
	Now   time.Time // the time of the command
}

// Result is what Create made. The cli prints it (R9, R10).
type Result struct {
	Name        string       // the slice name
	Branch      string       // slice/<name>
	From        eventlog.Ref // the base branch and its full commit id
	Worktree    string       // the full path of the worktree, with /
	WorktreeRel string       // .sliceflow/worktrees/<name>, from the main checkout
	Spec        string       // docs/slices/<name>/spec.md, inside the worktree
}

// checked is what the checks found out, for the steps after them.
type checked struct {
	mainRoot string // the main checkout, with / between the parts
	base     string // the base branch
}

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
		if errors.As(err, &ce) && len(ce.Lines) > 1 {
			details = ce.Lines[1:]
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

// Create makes a new slice. It first runs the checks of R11 in order, and
// stops at the first that fails. Every error of a check is a finished
// sentence that ends with "Nothing was changed." (PD1). Then it makes the
// branch, the worktree, the two files, and the commit (R4-R8). When one of
// those steps fails, or ctx is cancelled (Ctrl-C), it removes what this run
// made, newest first, and returns the message of R18.
func Create(ctx context.Context, opts CreateOptions) (Result, error) {
	c, err := checkCreate(opts)
	if err != nil {
		return Result{}, err
	}

	branch := "slice/" + opts.Name
	worktreeRel := ".sliceflow/worktrees/" + opts.Name
	worktree := c.mainRoot + "/" + worktreeRel
	specRel := "docs/slices/" + opts.Name + "/spec.md"
	eventsRel := "docs/slices/" + opts.Name + "/.events.jsonl"

	commit, err := git.CommitID(c.mainRoot, "refs/heads/"+c.base)
	if err != nil {
		return Result{}, unchanged(err)
	}
	from := eventlog.Ref{Branch: c.base, Commit: commit}

	// Build the texts first, so an error here leaves nothing on the disk.
	spec, err := specFile(templates.Spec, opts.Name, opts.Now)
	if err != nil {
		return Result{}, unchanged(err)
	}
	data, err := eventlog.Data(eventlog.SliceCreated{From: from, Title: opts.Title})
	if err != nil {
		return Result{}, unchanged(err)
	}
	logLine, err := eventlog.Line(eventlog.Event{
		Version: eventlog.Version,
		Seq:     1,
		Event:   eventlog.EventSliceCreated,
		At:      opts.Now,
		Data:    data,
	})
	if err != nil {
		return Result{}, unchanged(err)
	}

	// From here on, the disk changes. Each step that makes something adds
	// its undo to the list.
	cr := &creation{ctx: ctx, name: opts.Name}
	mainRoot := c.mainRoot
	if err := cr.check(nil); err != nil {
		return Result{}, err
	}

	madeDir, madeFile, err := writeGitignore(mainRoot)
	if madeDir {
		cr.add("folder .sliceflow", func() error { return os.Remove(filepath.Join(mainRoot, ".sliceflow")) })
	}
	if madeFile {
		cr.add("file .sliceflow/.gitignore", func() error {
			return os.Remove(filepath.Join(mainRoot, ".sliceflow", ".gitignore"))
		})
	}
	if err := cr.check(err); err != nil {
		return Result{}, err
	}

	// git can fail after it made something (e.g. a hook), so the undo is
	// added before the step, and it first looks whether there is anything.
	cr.add("branch "+branch, func() error {
		exists, err := git.BranchExists(mainRoot, branch)
		if err != nil || !exists {
			return err
		}
		return git.DeleteBranch(mainRoot, branch)
	})
	err = git.CreateBranch(mainRoot, branch, commit)
	if err := cr.check(err); err != nil {
		return Result{}, err
	}

	// git worktree add makes the folders .sliceflow/worktrees/ itself. If
	// that folder is new, we remove it too, but only when it is empty.
	worktreesDir := filepath.Join(mainRoot, ".sliceflow", "worktrees")
	_, statErr := os.Stat(worktreesDir)
	if errors.Is(statErr, os.ErrNotExist) {
		cr.add("folder .sliceflow/worktrees", func() error { return os.Remove(worktreesDir) })
	}
	// Added before the step: a failing post-checkout hook makes git exit
	// with an error but leave the worktree. If the folder is not there,
	// git may still remember it, so we prune. --force removes the files
	// and the staged changes too.
	cr.add("worktree "+worktreeRel, func() error {
		if _, err := os.Stat(filepath.FromSlash(worktree)); errors.Is(err, os.ErrNotExist) {
			return git.PruneWorktrees(mainRoot)
		}
		return git.RemoveWorktree(mainRoot, worktree)
	})
	err = git.AddWorktree(mainRoot, worktree, branch)
	if err := cr.check(err); err != nil {
		return Result{}, err
	}

	specPath := filepath.Join(worktree, filepath.FromSlash(specRel))
	if err := cr.check(os.MkdirAll(filepath.Dir(specPath), 0o755)); err != nil {
		return Result{}, err
	}
	if err := cr.check(os.WriteFile(specPath, []byte(spec), 0o644)); err != nil {
		return Result{}, err
	}
	err = os.WriteFile(filepath.Join(worktree, filepath.FromSlash(eventsRel)), logLine, 0o644)
	if err := cr.check(err); err != nil {
		return Result{}, err
	}
	if err := cr.check(git.Add(worktree, specRel, eventsRel)); err != nil {
		return Result{}, err
	}
	if err := cr.check(git.Commit(worktree, opts.Name+": slice created")); err != nil {
		return Result{}, err
	}

	return Result{
		Name:        opts.Name,
		Branch:      branch,
		From:        from,
		Worktree:    worktree,
		WorktreeRel: worktreeRel,
		Spec:        specRel,
	}, nil
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

// checkCreate runs the checks R12, R13, R14, R16, R15, R17, in this order.
// It changes nothing. It returns what the later steps need.
func checkCreate(opts CreateOptions) (checked, error) {
	root, err := git.Root(opts.Dir)
	if err != nil {
		return checked{}, unchanged(err)
	}

	hasCommits, err := git.HasCommits(root)
	if err != nil {
		return checked{}, unchanged(err)
	}
	if !hasCommits {
		return checked{}, unchanged(errors.New("the repository has no commits yet. Make a first commit, then try again"))
	}

	if err := ValidateName(opts.Name); err != nil {
		return checked{}, unchanged(err)
	}

	base, err := baseBranch(root, opts.From)
	if err != nil {
		return checked{}, unchanged(err)
	}

	if err := checkNameFree(root, opts.Name, base); err != nil {
		return checked{}, err
	}

	mainRoot, err := git.MainRoot(root)
	if err != nil {
		return checked{}, unchanged(err)
	}
	old := filepath.Join(mainRoot, ".sliceflow", "worktrees", opts.Name)
	_, err = os.Stat(old)
	if err == nil {
		return checked{}, unchanged(fmt.Errorf("folder .sliceflow/worktrees/%s already exists. Remove it, or pick another name", opts.Name))
	}
	if !errors.Is(err, os.ErrNotExist) {
		return checked{}, unchanged(err)
	}
	return checked{mainRoot: mainRoot, base: base}, nil
}

// checkNameFree returns the R15 error when the slice name is already used:
// by a local branch, by a branch of a remote, or by a folder on the base
// branch. It names the first place it finds, in that order.
func checkNameFree(root, name, base string) error {
	exists := func(where string) error {
		return unchanged(fmt.Errorf("slice %s already exists: %s", name, where))
	}
	branch := "slice/" + name

	local, err := git.BranchExists(root, branch)
	if err != nil {
		return unchanged(err)
	}
	if local {
		return exists("branch " + branch)
	}

	remotes, err := git.RemoteBranches(root, branch)
	if err != nil {
		return unchanged(err)
	}
	if len(remotes) > 0 {
		return exists("branch " + remotes[0])
	}

	folder := "docs/slices/" + name
	onBase, err := git.PathExists(root, "refs/heads/"+base, folder)
	if err != nil {
		return unchanged(err)
	}
	if onBase {
		return exists(fmt.Sprintf("folder %s on %s", folder, base))
	}
	return nil
}

// unchanged turns the error of a check into the user's message: the error
// text, then ". Nothing was changed." (PD1). git.ErrNotRepo already reads
// as R12 asks: "not inside a git repository".
func unchanged(err error) error {
	return fmt.Errorf("%w. Nothing was changed.", err)
}
