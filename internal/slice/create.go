package slice

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/BertPratya/sliceflow/internal/eventlog"
	"github.com/BertPratya/sliceflow/internal/git"
	"github.com/BertPratya/sliceflow/internal/template"
)

// CreateOptions are the inputs of Create.
type CreateOptions struct {
	Name, Title, From string
	skipChecks        bool // tests only: skip the checks of R14 to R17
}

// afterStep is a test seam: Create calls it after each step with the step's
// name. It is nil outside tests.
var afterStep func(step string)

// Created is the result of Create. Worktree is the full path with "/" on
// every system. Spec is the path inside the worktree.
type Created struct {
	Slice, Branch  string
	From           eventlog.From
	Worktree, Spec string
}

// interruptedError is R18's error when a signal stopped create.
func interruptedError(name string) *Error {
	return &Error{Message: "could not create slice " + name + ": interrupted. Everything sliceflow made was removed."}
}

// Interrupted is R18's error for a signal that arrives before Create starts.
func Interrupted(name string) *Error { return interruptedError(name) }

// Create makes the branch, the worktree, the spec and the event log, and
// the commit (Flow 1, steps 4 to 8). When a step fails, or ctx is done, it
// removes what this run made, newest first, and returns R18's *Error.
func Create(ctx context.Context, r Repo, opts CreateOptions) (Created, error) {
	if ctx.Err() != nil {
		return Created{}, interruptedError(opts.Name)
	}
	base, source, err := BaseBranch(ctx, r, opts.From)
	if err != nil {
		return Created{}, errBeforeStart(ctx, opts.Name, err)
	}
	if !opts.skipChecks {
		if err := checkCreate(ctx, r, opts, base, source); err != nil {
			return Created{}, errBeforeStart(ctx, opts.Name, err)
		}
	}
	root := filepath.FromSlash(r.Root)
	u := &undoList{root: root}
	// fail undoes this run's work and builds R18's error.
	fail := func(err error, gitCmd string) (Created, error) {
		failed := u.run(ctx)
		reason, extra := "interrupted", []string(nil)
		if ctx.Err() == nil {
			reason, extra = failureReason(err, gitCmd)
		}
		msg := "could not create slice " + opts.Name + ": " + reason + ". "
		if len(failed) > 0 {
			msg += "Could not remove: " + strings.Join(failed, ", ") + "."
		} else {
			msg += "Everything sliceflow made was removed."
		}
		if len(extra) > 0 {
			msg += "\n" + strings.Join(extra, "\n")
		}
		return Created{}, &Error{Message: msg}
	}
	// next marks the end of a step and checks ctx before the next one.
	next := func(step string) error {
		if afterStep != nil {
			afterStep(step)
		}
		return ctx.Err()
	}

	out, err := git.Run(ctx, r.Dir, "rev-parse", "--verify", "--quiet", "refs/heads/"+base+"^{commit}")
	if err != nil {
		return fail(err, "rev-parse")
	}
	commit := strings.TrimSpace(out)
	from := eventlog.From{Branch: base, Commit: commit}

	sliceflowDir := filepath.Join(root, ".sliceflow")
	switch err := os.Mkdir(sliceflowDir, 0o755); {
	case err == nil:
		u.addFolder(sliceflowDir)
	case errors.Is(err, fs.ErrExist):
		// Not made by this run: never removed.
	default:
		return fail(err, "")
	}
	if err := next("sliceflow-dir"); err != nil {
		return fail(err, "")
	}
	ignore := filepath.Join(sliceflowDir, ".gitignore")
	switch f, err := os.OpenFile(ignore, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644); {
	case err == nil:
		u.addFile(ignore)
		_, werr := f.WriteString("*\n")
		if cerr := f.Close(); werr == nil {
			werr = cerr
		}
		if werr != nil {
			return fail(&fs.PathError{Op: "write", Path: ignore, Err: werr}, "")
		}
	case errors.Is(err, fs.ErrExist):
		// Not made by this run: never removed.
	default:
		return fail(err, "")
	}
	if err := next("gitignore"); err != nil {
		return fail(err, "")
	}

	branch := "slice/" + opts.Name
	subject := opts.Name + ": slice created"
	wtDir := worktreeDir(r.Root, opts.Name)
	worktree := filepath.ToSlash(wtDir)
	worktreesDir := filepath.Join(sliceflowDir, "worktrees")
	_, statErr := os.Lstat(worktreesDir)
	hadWorktrees := statErr == nil
	_, wtErr := git.Run(ctx, root, "worktree", "add", "-b", branch, wtDir, commit)
	// git makes .sliceflow/worktrees/ when it is missing. It goes on the list
	// before the worktree, so the undo removes it after the worktree.
	if !hadWorktrees {
		if _, err := os.Lstat(worktreesDir); err == nil {
			u.addFolder(worktreesDir)
		}
	}
	if wtErr != nil {
		return fail(wtErr, "worktree add")
	}
	u.addSlice(wtDir, branch, commit, subject)
	if err := next("worktree"); err != nil {
		return fail(err, "")
	}

	now := r.Now()
	specRel := path.Join("docs", "slices", opts.Name, "spec.md")
	logRel := path.Join("docs", "slices", opts.Name, ".events.jsonl")
	line, err := eventlog.Encode(eventlog.NewCreated(now, from, opts.Title))
	if err != nil {
		return fail(err, "")
	}
	if err := os.MkdirAll(filepath.Join(wtDir, "docs", "slices", opts.Name), 0o755); err != nil {
		return fail(err, "")
	}
	if err := os.WriteFile(filepath.Join(wtDir, filepath.FromSlash(specRel)), template.Spec(opts.Name, now), 0o644); err != nil {
		return fail(err, "")
	}
	if err := os.WriteFile(filepath.Join(wtDir, filepath.FromSlash(logRel)), line, 0o644); err != nil {
		return fail(err, "")
	}
	if err := next("files"); err != nil {
		return fail(err, "")
	}

	if _, err := git.Run(ctx, wtDir, "add", "-f", "--", specRel, logRel); err != nil {
		return fail(err, "add")
	}
	if err := next("add"); err != nil {
		return fail(err, "")
	}
	if _, err := git.Run(ctx, wtDir, "commit", "-m", subject); err != nil {
		return fail(err, "commit")
	}
	return Created{
		Slice:    opts.Name,
		Branch:   branch,
		From:     from,
		Worktree: worktree,
		Spec:     specRel,
	}, nil
}

// errBeforeStart handles an error from the checks: nothing was made yet. A
// done ctx gives R18's interrupted error; the spec's own errors stay.
func errBeforeStart(ctx context.Context, name string, err error) error {
	var se *Error
	if ctx.Err() != nil && !errors.As(err, &se) {
		return interruptedError(name)
	}
	return err
}

// worktreeDir is the slice worktree folder, in the native form. It uses
// filepath.Join so a Windows network share root keeps its leading "//".
func worktreeDir(root, name string) string {
	return filepath.Join(filepath.FromSlash(root), ".sliceflow", "worktrees", name)
}
