package slice

import (
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
// branch, the worktree, the two files, and the commit (R4-R8). A step that
// fails returns its error as it is.
func Create(opts CreateOptions) (Result, error) {
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
		return Result{}, err
	}
	from := eventlog.Ref{Branch: c.base, Commit: commit}

	// Build the texts first, so an error here leaves nothing on the disk.
	spec, err := specFile(templates.Spec, opts.Name, opts.Now)
	if err != nil {
		return Result{}, err
	}
	data, err := eventlog.Data(eventlog.SliceCreated{From: from, Title: opts.Title})
	if err != nil {
		return Result{}, err
	}
	logLine, err := eventlog.Line(eventlog.Event{
		Version: eventlog.Version,
		Seq:     1,
		Event:   eventlog.EventSliceCreated,
		At:      opts.Now,
		Data:    data,
	})
	if err != nil {
		return Result{}, err
	}

	// made records what sliceflow created, so chunk 5 can remove only that.
	var made struct {
		sliceflowDir bool // the folder .sliceflow was new
		gitignore    bool // the file .sliceflow/.gitignore was new
		branch       bool // the branch was created
		worktree     bool // the worktree was added
	}

	var madeDir, madeFile bool
	madeDir, madeFile, err = writeGitignore(c.mainRoot)
	made.sliceflowDir, made.gitignore = madeDir, madeFile
	if err != nil {
		// TODO(chunk 5): undo
		return Result{}, err
	}
	if err := git.CreateBranch(c.mainRoot, branch, commit); err != nil {
		// TODO(chunk 5): undo
		return Result{}, err
	}
	made.branch = true
	// git worktree add makes the folders .sliceflow/worktrees/ itself.
	if err := git.AddWorktree(c.mainRoot, worktree, branch); err != nil {
		// TODO(chunk 5): undo
		return Result{}, err
	}
	made.worktree = true

	specPath := filepath.Join(worktree, filepath.FromSlash(specRel))
	if err := os.MkdirAll(filepath.Dir(specPath), 0o755); err != nil {
		// TODO(chunk 5): undo
		return Result{}, err
	}
	if err := os.WriteFile(specPath, []byte(spec), 0o644); err != nil {
		// TODO(chunk 5): undo
		return Result{}, err
	}
	if err := os.WriteFile(filepath.Join(worktree, filepath.FromSlash(eventsRel)), logLine, 0o644); err != nil {
		// TODO(chunk 5): undo
		return Result{}, err
	}
	if err := git.Add(worktree, specRel, eventsRel); err != nil {
		// TODO(chunk 5): undo
		return Result{}, err
	}
	if err := git.Commit(worktree, opts.Name+": slice created"); err != nil {
		// TODO(chunk 5): undo
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
