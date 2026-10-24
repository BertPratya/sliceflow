package slice

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BertPratya/sliceflow/internal/git"
)

// undoItem is one thing Create made in this run.
type undoItem struct {
	// A file or folder made by this run. A folder is removed only if empty.
	path  string
	isDir bool
	// A branch and its worktree, made by one "git worktree add". commit is
	// the commit the branch started at.
	worktree string
	branch   string
	commit   string
	subject  string // the subject of the commit Create makes on the branch
}

// undoList is what Create made in this run, oldest first.
type undoList struct {
	root  string // the main checkout, in the native form
	items []undoItem
}

func (u *undoList) addFile(p string)   { u.items = append(u.items, undoItem{path: p}) }
func (u *undoList) addFolder(p string) { u.items = append(u.items, undoItem{path: p, isDir: true}) }
func (u *undoList) addSlice(worktree, branch, commit, subject string) {
	u.items = append(u.items, undoItem{worktree: worktree, branch: branch, commit: commit, subject: subject})
}

// run removes the items, newest first, and returns what it could not remove,
// with "/" in paths. It never touches a thing that is not on the list. It
// does not use a done context.
func (u *undoList) run(ctx context.Context) []string {
	ctx = context.WithoutCancel(ctx)
	var failed []string
	for i := len(u.items) - 1; i >= 0; i-- {
		it := u.items[i]
		switch {
		case it.worktree != "":
			failed = append(failed, u.removeSlice(ctx, it)...)
		case it.isDir:
			if !removeEmptyDir(it.path) {
				failed = append(failed, filepath.ToSlash(it.path))
			}
		default:
			if err := os.Remove(it.path); err != nil && !errors.Is(err, fs.ErrNotExist) {
				failed = append(failed, filepath.ToSlash(it.path))
			}
		}
	}
	return failed
}

// removeEmptyDir removes dir only if it is empty. A folder with files in it
// is not ours to remove, so that is not a failure. It reports false only when
// an empty folder could not be removed.
func removeEmptyDir(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return errors.Is(err, fs.ErrNotExist)
	}
	if len(entries) > 0 {
		return true
	}
	err = os.Remove(dir)
	return err == nil || errors.Is(err, fs.ErrNotExist)
}

func (u *undoList) removeSlice(ctx context.Context, it undoItem) []string {
	var failed []string
	if _, err := git.Run(ctx, u.root, "worktree", "remove", "--force", it.worktree); err != nil {
		if _, statErr := os.Lstat(it.worktree); statErr == nil {
			// The folder is still there, and it holds the branch.
			return []string{filepath.ToSlash(it.worktree)}
		}
		_, _ = git.Run(ctx, u.root, "worktree", "prune")
	}
	ref := "refs/heads/" + it.branch
	tip, err := git.Run(ctx, u.root, "rev-parse", "--verify", "--quiet", ref+"^{commit}")
	if err != nil {
		var ge *git.Error
		if errors.As(err, &ge) && ge.ExitCode == 1 {
			return failed // no branch
		}
		return append(failed, it.branch)
	}
	tip = strings.TrimSpace(tip)
	if tip != it.commit {
		// Create's own commit may be on the branch: its parent is the start.
		parent, perr := git.Run(ctx, u.root, "rev-parse", "--verify", "--quiet", ref+"~1^{commit}")
		subject, serr := git.Run(ctx, u.root, "log", "-1", "--format=%s", ref)
		if perr != nil || serr != nil || strings.TrimSpace(parent) != it.commit || strings.TrimSpace(subject) != it.subject {
			return failed // not ours any more
		}
	}
	if _, err := git.Run(ctx, u.root, "branch", "-D", "--", it.branch); err != nil {
		failed = append(failed, it.branch)
	}
	return failed
}

// failureReason returns the reason of R18 and the lines that follow it.
// gitCmd is the git command of the failed step, e.g. "commit".
func failureReason(err error, gitCmd string) (string, []string) {
	var ge *git.Error
	if errors.As(err, &ge) {
		var lines []string
		for _, l := range strings.Split(ge.Stderr, "\n") {
			l = strings.TrimRight(l, " \t\r")
			if l == "" && len(lines) == 0 {
				continue
			}
			lines = append(lines, l)
		}
		if len(lines) == 0 {
			return "git " + gitCmd + " failed (exit code " + strconv.Itoa(ge.ExitCode) + ")", nil
		}
		return trimDot(lines[0]), lines[1:]
	}
	var pe *fs.PathError
	if errors.As(err, &pe) {
		return trimDot(pe.Op + " " + filepath.ToSlash(pe.Path) + ": " + pe.Err.Error()), nil
	}
	return trimDot(err.Error()), nil
}

func trimDot(s string) string { return strings.TrimSuffix(s, ".") }
