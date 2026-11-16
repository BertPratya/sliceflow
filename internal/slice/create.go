package slice

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BertPratya/sliceflow/internal/eventlog"
	"github.com/BertPratya/sliceflow/internal/git"
	"github.com/BertPratya/sliceflow/internal/project"
	"github.com/BertPratya/sliceflow/templates"
)

// ErrNoCommits is returned when the project's repository has no commit
// yet, so there is nothing to start a slice from.
var ErrNoCommits = errors.New("slice: the repository has no commits")

// worktreeSetting is the git setting that chooses where worktrees go.
const worktreeSetting = "sliceflow.worktree.dir"

// CreateOptions is what a new slice is made from.
type CreateOptions struct {
	// Name is the name as the user gave it. It may have capital letters,
	// which Create makes lowercase.
	Name string
	// Title is the slice's title, or "" for the default: the slice name.
	Title string
	// Description is one or two sentences about the slice, or "".
	Description string
	// From is the local branch to start from, or "" for main.
	From string
}

// CreateResult says what Create made, and where.
type CreateResult struct {
	Name       string   // the slice name, in lowercase
	Base       string   // the branch the slice started from
	BaseCommit string   // the full ID of the commit the slice started from
	Worktree   string   // the full path of the worktree
	Spec       string   // the full path of spec.md in the worktree
	Ignored    []string // the slice's files that an ignore rule matched, e.g. "spec.md"
}

// Create makes a new slice in the project that dir is in: the branch
// slice/<name> from the tip of the base branch, a worktree for it, and
// one commit on it with the spec and the event log. It never changes the
// user's own checkout.
//
// It checks, in this order, and stops at the first problem: a control
// character in the title or the description, the name, the project, the
// base branch, and a name that is taken. Nothing is made before all the
// checks pass. When a step that makes the slice fails, what was made is
// removed, and the error is a *StepError.
func Create(ctx context.Context, dir string, opts CreateOptions) (CreateResult, error) {
	if HasControlCharacter(opts.Title) {
		return CreateResult{}, fmt.Errorf("the title has a control character")
	}
	if HasControlCharacter(opts.Description) {
		return CreateResult{}, fmt.Errorf("the description has a control character")
	}
	name, err := ParseName(opts.Name)
	if err != nil {
		return CreateResult{}, err
	}

	root, err := git.RepoRoot(ctx, dir)
	if err != nil {
		return CreateResult{}, err
	}
	hasCommits, err := git.HasCommits(ctx, root)
	if err != nil {
		return CreateResult{}, err
	}
	if !hasCommits {
		return CreateResult{}, ErrNoCommits
	}

	base := opts.From
	if base == "" {
		base = "main"
	}
	baseCommit, err := git.BranchCommit(ctx, root, base)
	if err != nil {
		return CreateResult{}, err
	}
	if err := CheckNameFree(ctx, root, baseCommit, name); err != nil {
		return CreateResult{}, err
	}

	setting, err := git.Config(ctx, root, worktreeSetting)
	if err != nil {
		return CreateResult{}, err
	}
	worktree, err := project.WorktreePath(root, setting, name)
	if err != nil {
		return CreateResult{}, err
	}

	title := opts.Title
	if title == "" {
		title = name
	}
	result := CreateResult{
		Name:       name,
		Base:       base,
		BaseCommit: baseCommit,
		Worktree:   worktree,
		Spec:       filepath.Join(worktree, "docs", "slices", name, "spec.md"),
	}
	created := eventlog.SliceCreated{
		Name:        name,
		Title:       title,
		Description: opts.Description,
		Base:        base,
		BaseCommit:  baseCommit,
	}

	steps := createSteps(root, worktree, baseCommit, created, &result)
	if err := runSteps(ctx, steps); err != nil {
		return CreateResult{}, err
	}
	return result, nil
}

// createSteps returns the steps that make the slice described by created,
// in order: the branch at baseCommit, the worktree, the spec, the event
// log, and the commit. The commit step records in result which files an
// ignore rule matched.
func createSteps(root, worktree, baseCommit string, created eventlog.SliceCreated, result *CreateResult) []Step {
	name := created.Name
	branch := "slice/" + name
	folder := "docs/slices/" + name // inside the project, always with /
	specPath := filepath.Join(worktree, filepath.FromSlash(folder), "spec.md")
	logPath := filepath.Join(worktree, filepath.FromSlash(folder), ".events.jsonl")

	return []Step{
		{
			Name: "create the branch",
			Do: func(ctx context.Context) error {
				return git.CreateBranch(ctx, root, branch, baseCommit)
			},
			Undo: func() error {
				return git.DeleteBranch(context.Background(), root, branch)
			},
			Left: Leftover{Kind: "branch", Name: branch},
		},
		{
			Name: "create the worktree",
			Do: func(ctx context.Context) error {
				return git.AddWorktree(ctx, root, worktree, branch)
			},
			Undo: func() error {
				return git.RemoveWorktree(context.Background(), root, worktree)
			},
			Left: Leftover{Kind: "worktree", Name: worktree},
		},
		{
			// The spec and the log are inside the worktree, so removing
			// the worktree removes them: these steps need no Undo.
			Name: "write " + specPath,
			Do: func(ctx context.Context) error {
				return writeSpecFile(specPath, name, time.Now())
			},
		},
		{
			Name: "write " + logPath,
			Do: func(ctx context.Context) error {
				return eventlog.Create(ctx, logPath, created)
			},
		},
		{
			Name: "make the commit",
			Do: func(ctx context.Context) error {
				files := []string{folder + "/spec.md", folder + "/.events.jsonl"}
				for _, file := range files {
					ignored, err := git.IsIgnored(ctx, worktree, file)
					if err != nil {
						return err
					}
					if ignored {
						result.Ignored = append(result.Ignored, filepath.Base(file))
					}
				}
				message := name + ": add the spec and the event log"
				return git.Commit(ctx, worktree, message, files...)
			},
		},
	}
}

// writeSpecFile writes the spec of the slice name at path: the built-in
// spec template, with the slice's name and the day of now, in UTC,
// written like 2026-11-02. It makes the folders above path that are
// missing.
func writeSpecFile(path, name string, now time.Time) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	text := strings.NewReplacer(
		"{{slice}}", name,
		"{{date}}", now.UTC().Format("2006-01-02"),
	).Replace(templates.Spec)
	return os.WriteFile(path, []byte(text), 0o644)
}
