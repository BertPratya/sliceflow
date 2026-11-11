package slice

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/BertPratya/sliceflow/internal/eventlog"
	"github.com/BertPratya/sliceflow/internal/git"
	"github.com/BertPratya/sliceflow/templates"
)

// The files of a slice, in its folder docs/slices/<name> (R2), in the
// order of IgnoredFiles.
const (
	specFile = "spec.md"
	logFile  = ".events.jsonl"
)

// excludeLine is the line that makes git ignore the lock file of every
// log, in the repository's info/exclude file (R2, SD31).
const excludeLine = "*.events.jsonl.lock"

// CreateInput is what the user gave to slice create.
type CreateInput struct {
	// GivenName is the name as the user typed it (R21).
	GivenName string
	// Title is the text of --title, or "" when it was left out (R28).
	Title string
	// Description is the text of --description, or "" when it was left
	// out (R28).
	Description string
	// From is the base branch of --from, or "" when it was left out:
	// then the base is main (R10). internal/cli refuses an empty
	// --from= (SD29).
	From string
}

// Created is what slice create made, for the lines of R8.
type Created struct {
	// Name is the slice name, in lowercase (R21).
	Name string
	// Lowered is true when Name differs from the given name (R22).
	Lowered bool
	// Base is the base branch, and BaseCommit the full hash of the
	// commit the slice started from (R10, R11).
	Base, BaseCommit string
	// Worktree is the full path of the worktree, and SpecPath the full
	// path of spec.md in it (R8, R9).
	Worktree, SpecPath string
	// IgnoredFiles are the files that a .gitignore rule matched, and
	// that were added anyway: spec.md, then .events.jsonl (R5).
	IgnoredFiles []string
}

// Create makes the slice in.GivenName in the project of the folder dir:
// the branch slice/<name>, its worktree, and one commit with spec.md and
// the event log (R1 to R5, R37). It prints nothing. On an error, the
// error's text is the whole message of the spec. When ctx ends, it
// stops, and removes what it made (R39).
//
// The checks, in the order of R40. Each returns its error at once, with
// nothing to undo. When ctx has ended at a check, it returns stopped(nil).
//  1. A control character in in.Title, then in in.Description (R29).
//  2. The name, with ParseName (R19 to R21). Lowered is true when the
//     name differs from in.GivenName.
//  3. The project, with git.FindProject: git.ErrNotFound gives R14,
//     git.ErrNotARepo R13.
//  4. Commits, with git.HasCommits (R15).
//  5. The base: in.From, or main when in.From is "". Its commit, with
//     git.BranchCommit; git.ErrNoBranch gives R12, whose second line
//     depends on whether in.From was "".
//  6. The folder docs/slices/<name> in the base commit, with
//     git.FolderInCommit (R23), then the branch slice/<name>, with
//     git.BranchExists (R24).
//  7. The worktree path, with WorktreePath. Then record, before anything
//     is made: the folders of the path that do not exist yet; whether
//     the worktree folder exists; and whether git has a worktree
//     registered at the path already, with git.WorktreeRegistered.
//     Then the path of the info/exclude file, with git.ExcludeFile; a
//     failure is failedCheck("find the info/exclude file").
//
// The steps. A ctx that ended between two steps counts as Ctrl+C.
//  1. Add excludeLine to the info/exclude file (addExcludeLine, R2,
//     SD31). A failure is R38, with the file's full path; nothing is
//     on the stack yet. The line is never undone (SD31's cost).
//  2. git.CreateBranch at the base commit, in the project's folder.
//     Push the branch with branchStep: its undo is git.DeleteBranch.
//  3. Push the folders step (foldersStep). It comes before the worktree,
//     so its undo runs after the worktree's undo.
//  4. git.AddWorktree at the path, in the project's folder. Push the
//     worktree with worktreeStep: its undo is git.RemoveWorktree.
//  5. In the worktree, make docs/slices/<name>, write spec.md from
//     templates.Spec as it is (R3, SD30), then the log with
//     eventlog.Create (R27, R28).
//  6. git.IsIgnored for each file, before the add (R5); then
//     git.AddForce on both; then git.Commit with the message of R4. All
//     three run in the worktree, and each failure is the step "make the
//     commit".
//  7. Return Created. A ctx that ends after the commit no longer undoes
//     the finished slice.
//
// When a step fails, or ctx ends:
//  1. Make a new context with context.WithoutCancel (TD5).
//  2. When step 2 or 4 failed before its push, ask git with the new
//     context whether the thing is there anyway, and push it if so: the
//     branch with git.BranchExistsShielded; the worktree with
//     git.WorktreeRegistered, but only when it was not registered before
//     (check 7).
//  3. unwind with the new context.
//  4. Return stopped(left) when ctx has ended, even when the step's own
//     error is a git error. Otherwise failedStep after a git step, or
//     failedWrite after a file write or the lock. A ctx that ended
//     between two steps also gives stopped(left).
func Create(ctx context.Context, dir string, in CreateInput) (Created, error) {
	// checkFailed is the error of a check that failed: stopped(nil) when
	// ctx has ended, otherwise err.
	checkFailed := func(err error) (Created, error) {
		if ctx.Err() != nil {
			return Created{}, stopped(nil)
		}
		return Created{}, err
	}

	// Check 1.
	if HasControlCharacter(in.Title) {
		return checkFailed(errControl("--title", in.Title))
	}
	if HasControlCharacter(in.Description) {
		return checkFailed(errControl("--description", in.Description))
	}

	// Check 2.
	name, err := ParseName(in.GivenName)
	if err != nil {
		return checkFailed(errBadName(in.GivenName, err))
	}
	created := Created{Name: name, Lowered: name != in.GivenName}

	// Check 3.
	project, err := git.FindProject(ctx, dir)
	switch {
	case errors.Is(err, git.ErrNotFound):
		return checkFailed(errGitNotFound())
	case errors.Is(err, git.ErrNotARepo):
		return checkFailed(errNotARepo())
	case err != nil:
		return checkFailed(failedCheck("find the repository", err))
	}

	// Check 4.
	hasCommits, err := git.HasCommits(ctx, project.Dir)
	if err != nil {
		return checkFailed(failedCheck("check for commits", err))
	}
	if !hasCommits {
		return checkFailed(errNoCommits())
	}

	// Check 5.
	created.Base = in.From
	if created.Base == "" {
		created.Base = "main"
	}
	created.BaseCommit, err = git.BranchCommit(ctx, project.Dir, created.Base)
	if errors.Is(err, git.ErrNoBranch) {
		return checkFailed(errNoBase(name, created.Base, in.From != ""))
	}
	if err != nil {
		return checkFailed(failedCheck("find the base branch", err))
	}

	// Check 6.
	branch := "slice/" + name
	folderTaken, err := git.FolderInCommit(ctx, project.Dir, created.BaseCommit, "docs/slices/"+name)
	if err != nil {
		return checkFailed(failedCheck("check the name", err))
	}
	if folderTaken {
		return checkFailed(errFolderTaken(name, created.Base))
	}
	branchTaken, err := git.BranchExists(ctx, project.Dir, branch)
	if err != nil {
		return checkFailed(failedCheck("check the name", err))
	}
	if branchTaken {
		return checkFailed(errBranchTaken(name))
	}

	// Check 7.
	path, err := WorktreePath(ctx, project, name)
	if err != nil {
		return checkFailed(failedCheck("find the worktree folder", err))
	}
	missing := missingFolders(path)
	folderInfo, folderErr := os.Lstat(path)
	folderExisted := folderErr == nil && folderInfo.IsDir()
	registeredBefore, err := git.WorktreeRegistered(ctx, project.Dir, path)
	if err != nil {
		return checkFailed(failedCheck("find the worktree folder", err))
	}
	created.Worktree = path
	excludePath, err := git.ExcludeFile(ctx, project.Dir)
	if err != nil {
		return checkFailed(failedCheck("find the info/exclude file", err))
	}

	// The steps, and what to do when one fails.
	undoCtx := context.WithoutCancel(ctx)
	var stack undoStack
	// fail runs the undo, and returns the error: stopped when ctx has
	// ended, otherwise the error that stepErr makes from what is left.
	// A nil stepErr is used only when ctx has ended between two steps,
	// and also gives stopped.
	fail := func(stepErr func(left []undoStep) error) (Created, error) {
		left := stack.unwind(undoCtx)
		if ctx.Err() != nil || stepErr == nil {
			return Created{}, stopped(left)
		}
		return Created{}, stepErr(left)
	}
	gitStep := func(step string, err error) func([]undoStep) error {
		return func(left []undoStep) error { return failedStep(step, err, left) }
	}
	writeStep := func(path string, err error) func([]undoStep) error {
		return func(left []undoStep) error { return failedWrite(path, err, left) }
	}

	// Step 1.
	if ctx.Err() != nil {
		return fail(nil)
	}
	if err := addExcludeLine(excludePath); err != nil {
		return fail(writeStep(excludePath, err))
	}

	// Step 2.
	if ctx.Err() != nil {
		return fail(nil)
	}
	branchUndo := branchStep(name, func(ctx context.Context) error {
		return git.DeleteBranch(ctx, project.Dir, branch)
	})
	if err := git.CreateBranch(ctx, project.Dir, branch, created.BaseCommit); err != nil {
		// Failure 2: the branch may be there anyway. When git cannot
		// tell, it is pushed too, so R41 names it if it cannot be
		// removed.
		if there, askErr := git.BranchExistsShielded(undoCtx, project.Dir, branch); there || askErr != nil {
			stack.push(branchUndo)
		}
		return fail(gitStep("create the branch", err))
	}
	stack.push(branchUndo)

	// Step 3.
	stack.push(foldersStep(path, folderExisted, folderInfo, missing))

	// Step 4.
	if ctx.Err() != nil {
		return fail(nil)
	}
	worktreeUndo := worktreeStep(path, func(ctx context.Context) error {
		return git.RemoveWorktree(ctx, project.Dir, path)
	})
	if err := git.AddWorktree(ctx, project.Dir, path, branch); err != nil {
		// Failure 2: the worktree may be there anyway, but never undo
		// one that git had registered before. When git cannot tell, it
		// is pushed too, as for the branch.
		if !registeredBefore {
			if there, askErr := git.WorktreeRegistered(undoCtx, project.Dir, path); there || askErr != nil {
				stack.push(worktreeUndo)
			}
		}
		return fail(gitStep("create the worktree", err))
	}
	stack.push(worktreeUndo)

	// Step 5.
	if ctx.Err() != nil {
		return fail(nil)
	}
	sliceDir := filepath.Join(path, "docs", "slices", name)
	if err := os.MkdirAll(sliceDir, 0o755); err != nil {
		return fail(writeStep(sliceDir, err))
	}
	created.SpecPath = filepath.Join(sliceDir, specFile)
	if err := os.WriteFile(created.SpecPath, []byte(templates.Spec), 0o644); err != nil {
		return fail(writeStep(created.SpecPath, err))
	}
	logPath := filepath.Join(sliceDir, logFile)
	title := in.Title
	if title == "" {
		title = name
	}
	firstEvent := eventlog.SliceCreated{
		Name:        name,
		Title:       title,
		Description: in.Description,
		Base:        created.Base,
		BaseCommit:  created.BaseCommit,
	}
	if err := eventlog.Create(ctx, logPath, firstEvent); err != nil {
		return fail(writeStep(logPath, err))
	}

	// Step 6.
	if ctx.Err() != nil {
		return fail(nil)
	}
	var gitPaths []string
	for _, file := range []string{specFile, logFile} {
		gitPath := "docs/slices/" + name + "/" + file
		gitPaths = append(gitPaths, gitPath)
		ignored, err := git.IsIgnored(ctx, path, gitPath)
		if err != nil {
			return fail(gitStep("make the commit", err))
		}
		if ignored {
			created.IgnoredFiles = append(created.IgnoredFiles, file)
		}
	}
	if err := git.AddForce(ctx, path, gitPaths...); err != nil {
		return fail(gitStep("make the commit", err))
	}
	if err := git.Commit(ctx, path, name+": add the spec and the event log"); err != nil {
		return fail(gitStep("make the commit", err))
	}

	// Step 7.
	return created, nil
}

// missingFolders returns the folders of path that do not exist, path
// itself included, deepest first. It stops at the first folder that
// exists, or that it cannot check: that one counts as there, so the undo
// never removes it.
func missingFolders(path string) []string {
	var missing []string
	for {
		if _, err := os.Lstat(path); !errors.Is(err, fs.ErrNotExist) {
			return missing
		}
		missing = append(missing, path)
		parent := filepath.Dir(path)
		if parent == path {
			return missing
		}
		path = parent
	}
}

// foldersStep is the undo step of the folders around the worktree at
// path (R36). folderExisted and info tell whether the worktree folder
// was there before; missing are the folders that were not, deepest
// first (missingFolders).
//
// Its undo:
//  1. When the worktree folder was there before and is gone, make it
//     again, empty, with its old permissions: git worktree remove
//     deletes it, and so does git when it is stopped during worktree add.
//  2. Remove the folders of missing, deepest first, with os.Remove,
//     which removes only an empty folder.
//
// It is best effort: it ignores every error of os.Mkdir and os.Remove,
// and always returns nil. Its left is empty, so it has no R41 line.
func foldersStep(path string, folderExisted bool, info fs.FileInfo, missing []string) undoStep {
	return undoStep{undo: func(context.Context) error {
		if folderExisted {
			if _, err := os.Lstat(path); errors.Is(err, fs.ErrNotExist) {
				_ = os.Mkdir(path, info.Mode().Perm())
			}
		}
		for _, folder := range missing {
			_ = os.Remove(folder)
		}
		return nil
	}}
}

// addExcludeLine adds excludeLine to the info/exclude file at path,
// unless it is there already (R2, SD31).
//
// Steps:
//  1. Read the file. When it is missing, its text is "".
//  2. If a line of the file is exactly excludeLine, with a "\r" at its
//     end cut off (Windows line ends), change nothing.
//  3. Make the file's folder when it is missing.
//  4. Add excludeLine and "\n" at the end of the file, making the file
//     when it is missing. When the file is not empty and does not end
//     with "\n", write "\n" first, so the user's last line stays as it
//     is.
//  5. Close the file. When nothing else failed, Close's error is the
//     result: the line may not have been saved.
//
// Known limits, low risk: there is no Sync, and two runs at once could
// add the line twice.
func addExcludeLine(path string) (err error) {
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	text := string(data)
	for _, line := range strings.Split(text, "\n") {
		if strings.TrimSuffix(line, "\r") == excludeLine {
			return nil
		}
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := f.Close(); err == nil {
			err = closeErr
		}
	}()
	add := excludeLine + "\n"
	if text != "" && !strings.HasSuffix(text, "\n") {
		add = "\n" + add
	}
	_, err = f.WriteString(add)
	return err
}
