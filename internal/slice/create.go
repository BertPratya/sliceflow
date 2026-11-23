package slice

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path"
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
// that makes the slice fails before the commit is made, what was made
// is removed, and the error is a *StepError. When the commit was made,
// but git or the user stopped it afterwards, nothing is removed, and the
// error is a *CommittedError. A failure because git is not found holds
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

	lockPath, err := git.GitPath(ctx, top, "sliceflow/locks/"+name+".lock")
	if err != nil {
		return CreateResult{}, &CheckError{Check: "find git's folder", Err: err}
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

	steps := createSteps(top, lockPath, created)
	if err := runSteps(ctx, steps); err != nil {
		return CreateResult{}, err
	}
	if err := commitSlice(ctx, top, &result); err != nil {
		return CreateResult{}, commitFailed(top, name, base, steps, err)
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

// createSteps returns the steps that write the slice described by
// created in the checkout whose top folder is top: the spec, then the
// event log, written while holding the lock at lockPath. Their Undo
// removes the files they wrote, and the slice's folder when the spec's
// step made it. A folder that was there before stays, and so do the
// folders above the slice's folder.
func createSteps(top, lockPath string, created eventlog.SliceCreated) []Step {
	name := created.Name
	folder := filepath.Join(top, "docs", "slices", name)
	specPath := filepath.Join(folder, "spec.md")
	logPath := filepath.Join(folder, ".events.jsonl")
	left := Leftover{Kind: "folder", Name: folder}
	madeFolder := false

	return []Step{
		{
			Name: "write " + specPath,
			Do: func(ctx context.Context) error {
				_, err := os.Lstat(folder)
				if err != nil && !errors.Is(err, fs.ErrNotExist) {
					return err
				}
				madeFolder = err != nil
				return writeSpecFile(specPath, name, time.Now())
			},
			Undo: func() error {
				if err := removeFile(specPath); err != nil {
					return err
				}
				if madeFolder {
					return removeEmptyFolder(folder)
				}
				return nil
			},
			Left: left,
		},
		{
			Name: "write " + logPath,
			Do: func(ctx context.Context) error {
				return eventlog.Create(ctx, logPath, lockPath, created)
			},
			Undo: func() error {
				return removeFile(logPath)
			},
			Left: left,
		},
	}
}

// commitSlice commits the spec and the event log of the slice in result,
// and only them, on the branch checked out in the checkout whose top
// folder is top. It records in result which of the two files an ignore
// rule matched; they are committed anyway.
func commitSlice(ctx context.Context, top string, result *CreateResult) error {
	folder := "docs/slices/" + result.Name // inside the project, always with /
	files := []string{folder + "/spec.md", folder + "/.events.jsonl"}
	for _, file := range files {
		ignored, err := git.IsIgnored(ctx, top, file)
		if err != nil {
			return err
		}
		if ignored {
			result.Ignored = append(result.Ignored, path.Base(file))
		}
	}
	message := result.Name + ": add the spec and the event log"
	return git.Commit(ctx, top, message, files...)
}

// commitFailed returns the error of Create when the commit of the slice
// name failed with err, in the checkout whose top folder is top, after
// steps wrote the slice's files.
//
// When the commit was made anyway, the slice's event log is in the
// commit checked out now, and commitFailed removes nothing and returns a
// *CommittedError. Otherwise, it unstages the slice's folder, undoes
// steps, and returns a *StepError for the step "make the commit". When
// it cannot tell whether the commit was made, it undoes, as if it was
// not: the files of a commit that was made can be brought back with git.
// It never waits for the user: ctx may already be cancelled.
func commitFailed(top, name, base string, steps []Step, err error) error {
	ctx := context.Background()
	folder := "docs/slices/" + name // inside the project, always with /
	made, checkErr := git.PathExists(ctx, top, "HEAD", folder+"/.events.jsonl")
	if checkErr == nil && made {
		return &CommittedError{Name: name, Base: base, Err: err}
	}

	var left []Leftover
	if unstageErr := git.Unstage(ctx, top, folder); unstageErr != nil {
		left = append(left, Leftover{Kind: "staged", Name: folder + "/", Top: top})
	}
	for _, thing := range undoSteps(steps) {
		left = addLeftover(left, thing)
	}
	return &StepError{Step: "make the commit", Err: err, Left: left}
}

// removeFile removes the file at path. A file that is not there is not
// an error.
func removeFile(path string) error {
	err := os.Remove(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return err
}

// removeEmptyFolder removes the folder at path, which must be empty. A
// folder that is not there is not an error, and neither is an empty
// folder that cannot be removed: nothing in it is left behind.
func removeEmptyFolder(path string) error {
	err := os.Remove(path)
	if err == nil || errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if entries, readErr := os.ReadDir(path); readErr == nil && len(entries) == 0 {
		return nil
	}
	return err
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
