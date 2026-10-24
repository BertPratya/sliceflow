package slice

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/BertPratya/sliceflow/internal/git"
)

// undoItem is one thing Create made in this run: a file or folder (path), a
// worktree (worktree), or a branch (branch).
type undoItem struct {
	// A file or folder made by this run. A folder is removed only if empty.
	path  string
	isDir bool
	// A worktree made by "git worktree add".
	worktree string
	// A branch made by "git branch". commit is the commit it started at.
	branch  string
	commit  string
	subject string // the subject of the commit Create makes on the branch
}

// undoList is what Create made in this run, oldest first.
type undoList struct {
	root  string // the main checkout, in the native form
	items []undoItem
}

func (u *undoList) addFile(p string)   { u.items = append(u.items, undoItem{path: p}) }
func (u *undoList) addFolder(p string) { u.items = append(u.items, undoItem{path: p, isDir: true}) }
func (u *undoList) addBranch(branch, commit, subject string) {
	u.items = append(u.items, undoItem{branch: branch, commit: commit, subject: subject})
}
func (u *undoList) addWorktree(p string) { u.items = append(u.items, undoItem{worktree: p}) }

// insertFolder puts a folder into the list at index i, so the undo removes it
// after the items that come later.
func (u *undoList) insertFolder(i int, p string) {
	u.items = append(u.items, undoItem{})
	copy(u.items[i+1:], u.items[i:])
	u.items[i] = undoItem{path: p, isDir: true}
}

// rel writes p relative to the main checkout, with "/".
func (u *undoList) rel(p string) string {
	if r, err := filepath.Rel(u.root, p); err == nil {
		return filepath.ToSlash(r)
	}
	return filepath.ToSlash(p)
}

// run removes the items, newest first, and returns what it could not remove,
// with "/" in paths relative to the main checkout. It never touches a thing
// that is not on the list. It does not use a done context.
func (u *undoList) run(ctx context.Context) []string {
	ctx = context.WithoutCancel(ctx)
	var failed []string
	for i := len(u.items) - 1; i >= 0; i-- {
		it := u.items[i]
		switch {
		case it.worktree != "":
			if !u.removeWorktree(ctx, it.worktree) {
				failed = append(failed, u.rel(it.worktree))
			}
		case it.branch != "":
			if !u.removeBranch(ctx, it) {
				failed = append(failed, "branch "+it.branch)
			}
		case it.isDir:
			if !removeEmptyDir(it.path) {
				failed = append(failed, u.rel(it.path))
			}
		default:
			if err := os.Remove(it.path); err != nil && !errors.Is(err, fs.ErrNotExist) {
				failed = append(failed, u.rel(it.path))
			}
		}
	}
	return failed
}

// listedWorktree returns the path git lists for the worktree at wtDir, and
// whether there is one. It matches the path alone. ok is false when the list
// could not be read.
func (u *undoList) listedWorktree(ctx context.Context, wtDir string) (listed string, found, ok bool) {
	out, err := git.Run(ctx, u.root, "worktree", "list", "--porcelain")
	if err != nil {
		return "", false, false
	}
	want := normPath(wtDir)
	for _, line := range strings.Split(out, "\n") {
		p, cut := strings.CutPrefix(strings.TrimRight(line, "\r"), "worktree ")
		if cut && normPath(p) == want {
			return p, true, true
		}
	}
	return "", false, true
}

// cleanFailedAdd removes what a failed or stopped "git worktree add" left at
// wtDir: the worktree git registered there, then the folder. It returns what
// it could not remove. R17 proved the folder was not there before this run.
// listedBefore says git already listed a worktree at wtDir before the
// command: that entry is not this run's, so it stays.
func (u *undoList) cleanFailedAdd(ctx context.Context, wtDir string, listedBefore bool) []string {
	ctx = context.WithoutCancel(ctx)
	var failed []string
	if !listedBefore {
		if p, found, _ := u.listedWorktree(ctx, wtDir); found {
			_, _ = git.Run(ctx, u.root, "worktree", "remove", "--force", "--force", p)
		}
	}
	if exists(wtDir) {
		if err := os.RemoveAll(wtDir); err != nil && exists(wtDir) {
			failed = append(failed, u.rel(wtDir))
		}
	}
	return failed
}

func exists(p string) bool {
	_, err := os.Lstat(p)
	return err == nil
}

// normPath makes two spellings of one path equal: it resolves links in the
// folder part (macOS /var, Windows short names), and ignores case on Windows
// and macOS.
func normPath(p string) string {
	p = filepath.Clean(filepath.FromSlash(p))
	dir, base := filepath.Split(p)
	if r, err := filepath.EvalSymlinks(filepath.Clean(dir)); err == nil {
		p = filepath.Join(r, base)
	}
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		p = strings.ToLower(p)
	}
	return p
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

// removeWorktree removes a worktree. It reports false when the folder stays.
// When the folder is gone but git still lists this run's entry, e.g. a locked
// one, it removes only that entry. It never prunes: that would clear the
// other stale entries in the repo, which are not ours.
func (u *undoList) removeWorktree(ctx context.Context, wt string) bool {
	if _, err := git.Run(ctx, u.root, "worktree", "remove", "--force", wt); err != nil {
		if exists(wt) {
			return false
		}
		_, _ = git.Run(ctx, u.root, "worktree", "remove", "--force", "--force", wt)
	}
	return true
}

// removeBranch deletes the branch, only if it still points at the commit
// Create started from, or at Create's own commit on top of it. It reports
// false when a branch that is ours could not be deleted.
func (u *undoList) removeBranch(ctx context.Context, it undoItem) bool {
	ref := "refs/heads/" + it.branch
	tip, err := git.Run(ctx, u.root, "rev-parse", "--verify", "--quiet", ref+"^{commit}")
	if err != nil {
		var ge *git.Error
		return errors.As(err, &ge) && ge.ExitCode == 1 // no branch
	}
	tip = strings.TrimSpace(tip)
	if tip != it.commit {
		// Create's own commit may be on the branch: its parent is the start.
		parent, perr := git.Run(ctx, u.root, "rev-parse", "--verify", "--quiet", ref+"~1^{commit}")
		subject, serr := git.Run(ctx, u.root, "log", "-1", "--format=%s", ref)
		if perr != nil || serr != nil || strings.TrimSpace(parent) != it.commit || strings.TrimSpace(subject) != it.subject {
			return true // not ours any more
		}
	}
	_, err = git.Run(ctx, u.root, "branch", "-D", "--", it.branch)
	return err == nil
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
