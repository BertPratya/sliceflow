package slice

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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

// checkCreate runs the checks R12, R13, R14, R16, R15, R17, in this order.
// It changes nothing. It returns what the later steps need.
func checkCreate(opts CreateOptions) (checked, error) {
	root, err := openRepo(opts.Dir)
	if err != nil {
		return checked{}, err
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

// openRepo finds the top folder of the repository that dir is in, and
// checks that it has a commit (R12, R13). Its errors end with "Nothing was
// changed.". Create, Status and List all start with it.
func openRepo(dir string) (root string, err error) {
	root, err = git.Root(dir)
	if err != nil {
		return "", unchanged(err)
	}
	hasCommits, err := git.HasCommits(root)
	if err != nil {
		return "", unchanged(err)
	}
	if !hasCommits {
		return "", unchanged(errors.New("the repository has no commits yet. Make a first commit, then try again"))
	}
	return root, nil
}
