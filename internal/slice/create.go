package slice

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BertPratya/sliceflow/internal/eventlog"
	"github.com/BertPratya/sliceflow/internal/git"
	"github.com/BertPratya/sliceflow/templates"
)

// baseSetting is the git setting that chooses the base branch of every
// new slice.
const baseSetting = "sliceflow.base"

// CreateOptions is what a new slice is made from.
type CreateOptions struct {
	// Name is the name as the user gave it. It may have capital letters,
	// which Create makes lowercase.
	Name string
	// Title is the slice's title, or "" for the default: the slice name.
	Title string
	// Description is one or two sentences about the slice, or "".
	Description string
	// From is the base branch, a local branch, or "" for the git setting
	// sliceflow.base, else main.
	From string
}

// CreateResult says what Create made, and where.
type CreateResult struct {
	Name    string   // the slice name, in lowercase
	Base    string   // the base branch, where the slice's commit was made
	Spec    string   // the full path of spec.md in the user's checkout
	Ignored []string // the slice's files that an ignore rule matched, e.g. "spec.md"
}

// Create makes a new slice in the user's checkout, the checkout that dir
// is in, which must be on the base branch: it writes the spec and the
// event log in docs/slices/<name>/, and makes one commit of only those
// two files on the base branch. It makes no branch and no worktree, and
// leaves the user's other files, what is staged, and the checked-out
// branch as they were.
//
// It checks, in this order, and stops at the first problem: a control
// character in the title or the description (*ControlCharacterError),
// the name (*InvalidNameError), the project (ErrNoCommits), the base
// branch (*NoBaseError), the user's checkout on the base branch
// (*WrongBranchError), and a name that is taken (*NameTakenError).
// When a check cannot be done, e.g. because git fails, the error is a
// *CheckError. Nothing is made before all the checks pass. When a step
// that makes the slice fails, what was made is removed, and the error is
// a *StepError. A failure because git is not found holds
// git.ErrNotFound, and a stop through ctx holds ctx's error.
func Create(ctx context.Context, dir string, opts CreateOptions) (CreateResult, error) {
	if HasControlCharacter(opts.Title) {
		return CreateResult{}, &ControlCharacterError{Option: "title", Text: opts.Title}
	}
	if HasControlCharacter(opts.Description) {
		return CreateResult{}, &ControlCharacterError{Option: "description", Text: opts.Description}
	}
	name, err := ParseName(opts.Name)
	if err != nil {
		return CreateResult{}, &InvalidNameError{Given: opts.Name, Problem: err}
	}

	root, err := findProject(ctx, dir)
	if err != nil {
		return CreateResult{}, err
	}
	top, err := git.CheckoutTop(ctx, dir)
	if err != nil {
		return CreateResult{}, &CheckError{Check: "find the top of the checkout", Err: err}
	}

	base, source, err := chooseBase(ctx, root, opts.From)
	if err != nil {
		return CreateResult{}, err
	}
	baseCommit, err := git.BranchCommit(ctx, root, base)
	if errors.Is(err, git.ErrNoBranch) {
		return CreateResult{}, &NoBaseError{Name: name, Base: base, Source: source}
	}
	if err != nil {
		return CreateResult{}, &CheckError{Check: "read the branch " + base, Err: err}
	}

	checkedOut, err := git.CheckedOutBranch(ctx, top)
	if err != nil {
		return CreateResult{}, &CheckError{Check: "read the current branch", Err: err}
	}
	if checkedOut != base {
		return CreateResult{}, &WrongBranchError{Base: base, Branch: checkedOut}
	}

	err = CheckNameFree(ctx, root, top, baseCommit, name)
	if errors.Is(err, ErrFolderTaken) || errors.Is(err, ErrBranchTaken) {
		return CreateResult{}, &NameTakenError{Name: name, Err: err}
	}
	if err != nil {
		return CreateResult{}, &CheckError{Check: "check the name", Err: err}
	}

	title := opts.Title
	if title == "" {
		title = name
	}
	result := CreateResult{
		Name: name,
		Base: base,
		Spec: filepath.Join(top, "docs", "slices", name, "spec.md"),
	}
	created := eventlog.SliceCreated{
		Name:        name,
		Title:       title,
		Description: opts.Description,
		Base:        base,
		BaseCommit:  baseCommit,
	}

	steps := createSteps(top, created, &result)
	if err := runSteps(ctx, steps); err != nil {
		return CreateResult{}, err
	}
	return result, nil
}

// chooseBase returns the base branch of a new slice in the project at
// root, and where it came from: from when it is not "", else the git
// setting sliceflow.base when it is not "", else main. It reads the
// setting only when from is "". It does not check that the branch
// exists. When the setting cannot be read, the error is a *CheckError.
func chooseBase(ctx context.Context, root, from string) (string, BaseSource, error) {
	if from != "" {
		return from, BaseFromFlag, nil
	}
	setting, err := git.Config(ctx, root, baseSetting)
	if err != nil {
		return "", BaseDefault, &CheckError{Check: "read the setting " + baseSetting, Err: err}
	}
	if setting != "" {
		return setting, BaseFromSetting, nil
	}
	return "main", BaseDefault, nil
}

// createSteps returns the steps that make the slice described by created
// in the checkout whose top folder is top, in order: the spec, the event
// log, and the commit. The commit step records in result which files an
// ignore rule matched.
func createSteps(top string, created eventlog.SliceCreated, result *CreateResult) []Step {
	name := created.Name
	folder := "docs/slices/" + name // inside the project, always with /
	specPath := filepath.Join(top, filepath.FromSlash(folder), "spec.md")
	logPath := filepath.Join(top, filepath.FromSlash(folder), ".events.jsonl")

	return []Step{
		{
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
					ignored, err := git.IsIgnored(ctx, top, file)
					if err != nil {
						return err
					}
					if ignored {
						result.Ignored = append(result.Ignored, filepath.Base(file))
					}
				}
				message := name + ": add the spec and the event log"
				return git.Commit(ctx, top, message, files...)
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
