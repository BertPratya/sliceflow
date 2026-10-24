package slice

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/BertPratya/sliceflow/internal/eventlog"
	"github.com/BertPratya/sliceflow/internal/git"
	"github.com/BertPratya/sliceflow/internal/template"
)

// CreateOptions are the inputs of Create.
type CreateOptions struct{ Name, Title, From string }

// Created is the result of Create. Worktree is the full path with "/" on
// every system. Spec is the path inside the worktree.
type Created struct {
	Slice, Branch  string
	From           eventlog.From
	Worktree, Spec string
}

// Create makes the branch, the worktree, the spec and the event log, and
// the commit (Flow 1, steps 4 to 8). It does not undo anything yet.
func Create(ctx context.Context, r Repo, opts CreateOptions) (Created, error) {
	base, source, err := BaseBranch(ctx, r, opts.From)
	if err != nil {
		return Created{}, err
	}
	if err := checkCreate(ctx, r, opts, base, source); err != nil {
		return Created{}, err
	}
	out, err := git.Run(ctx, r.Dir, "rev-parse", "--verify", "--quiet", "refs/heads/"+base+"^{commit}")
	if err != nil {
		return Created{}, err
	}
	commit := strings.TrimSpace(out)
	from := eventlog.From{Branch: base, Commit: commit}

	root := filepath.FromSlash(r.Root)
	sliceflowDir := filepath.Join(root, ".sliceflow")
	if err := os.MkdirAll(sliceflowDir, 0o755); err != nil {
		return Created{}, err
	}
	ignore := filepath.Join(sliceflowDir, ".gitignore")
	if _, err := os.Stat(ignore); os.IsNotExist(err) {
		if err := os.WriteFile(ignore, []byte("*\n"), 0o644); err != nil {
			return Created{}, err
		}
	} else if err != nil {
		return Created{}, err
	}

	branch := "slice/" + opts.Name
	wtDir := worktreeDir(r.Root, opts.Name)
	worktree := filepath.ToSlash(wtDir)
	if _, err := git.Run(ctx, root, "worktree", "add", "-b", branch, wtDir, commit); err != nil {
		return Created{}, err
	}

	now := r.Now()
	specRel := path.Join("docs", "slices", opts.Name, "spec.md")
	logRel := path.Join("docs", "slices", opts.Name, ".events.jsonl")
	line, err := eventlog.Encode(eventlog.NewCreated(now, from, opts.Title))
	if err != nil {
		return Created{}, err
	}
	if err := os.MkdirAll(filepath.Join(wtDir, "docs", "slices", opts.Name), 0o755); err != nil {
		return Created{}, err
	}
	if err := os.WriteFile(filepath.Join(wtDir, filepath.FromSlash(specRel)), template.Spec(opts.Name, now), 0o644); err != nil {
		return Created{}, err
	}
	if err := os.WriteFile(filepath.Join(wtDir, filepath.FromSlash(logRel)), line, 0o644); err != nil {
		return Created{}, err
	}

	if _, err := git.Run(ctx, wtDir, "add", "-f", "--", specRel, logRel); err != nil {
		return Created{}, err
	}
	if _, err := git.Run(ctx, wtDir, "commit", "-m", opts.Name+": slice created"); err != nil {
		return Created{}, err
	}
	return Created{
		Slice:    opts.Name,
		Branch:   branch,
		From:     from,
		Worktree: worktree,
		Spec:     specRel,
	}, nil
}

// checkCreate runs the checks before anything is made. Task 08 fills it.
func checkCreate(ctx context.Context, r Repo, opts CreateOptions, base, source string) error {
	return nil
}

// worktreeDir is the slice worktree folder, in the native form. It uses
// filepath.Join so a Windows network share root keeps its leading "//".
func worktreeDir(root, name string) string {
	return filepath.Join(filepath.FromSlash(root), ".sliceflow", "worktrees", name)
}
