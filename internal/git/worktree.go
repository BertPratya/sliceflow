package git

import (
	"context"
	"path/filepath"
	"runtime"
	"strings"
)

// AddWorktree makes a linked worktree at path, with the local branch
// <branch> checked out (R37). If the folder at path exists and is not
// empty, git refuses; if it is empty, git uses it (R25).
//
// Steps:
//  1. Run git worktree add --end-of-options <path> <branch>.
//     --end-of-options makes git read a path or a name that starts with
//     - as a path or a name, not as an option.
//  2. Return git's error as it is.
func AddWorktree(ctx context.Context, dir, path, branch string) error {
	_, err := run(ctx, dir, "worktree", "add", "--end-of-options", path, branch)
	return err
}

// WorktreeRegistered reports whether git knows a worktree at path, even
// one whose folder is gone or half made (R36). The undo asks it, after a
// failure or a Ctrl+C, whether the worktree was made anyway, so git is
// shielded from Ctrl+C (runShielded). A relative path is read from dir.
//
// Steps:
//  1. Run git worktree list --porcelain. Each worktree has a line
//     "worktree <path>". There is no -z: it needs git 2.36.
//  2. Turn path into a full path, and resolve it with realPath.
//  3. Return true when the path of a "worktree" line, resolved with
//     realPath too, names the same folder, with samePath. Otherwise
//     return false.
func WorktreeRegistered(ctx context.Context, dir, path string) (bool, error) {
	out, err := runShielded(ctx, dir, "worktree", "list", "--porcelain")
	if err != nil {
		return false, err
	}

	if !filepath.IsAbs(path) {
		path = filepath.Join(dir, path)
	}
	full, err := filepath.Abs(path)
	if err != nil {
		return false, err
	}
	want := realPath(full)

	for _, line := range strings.Split(out, "\n") {
		listed, ok := strings.CutPrefix(strings.TrimSuffix(line, "\r"), "worktree ")
		if !ok {
			continue
		}
		if samePath(realPath(filepath.FromSlash(listed)), want) {
			return true, nil
		}
	}
	return false, nil
}

// RemoveWorktree removes the linked worktree at path, and its folder,
// even when it has changes or untracked files, or is locked (R36). The
// undo calls it, so git is shielded from Ctrl+C (runShielded).
//
// Steps:
//  1. Run git worktree remove -f -f --end-of-options <path>. One -f
//     removes a worktree with changes or untracked files; the second
//     also removes a locked one. A worktree that git add left half made
//     is locked.
//  2. Return git's error as it is.
func RemoveWorktree(ctx context.Context, dir, path string) error {
	_, err := runShielded(ctx, dir, "worktree", "remove", "-f", "-f", "--end-of-options", path)
	return err
}

// realPath returns the full path path with its links followed, so two
// names of one folder become the same path: macOS's /var is a link to
// /private/var, and Windows has short names such as RUNNER~1.
//
// Steps:
//  1. Find the deepest folder of path that can be resolved, with
//     filepath.EvalSymlinks: the folder of a half made worktree may not
//     exist. Any error means: go one folder up.
//  2. Add the rest of path to it, and return it clean (filepath.Clean).
//     When nothing can be resolved, return path clean.
func realPath(path string) string {
	path = filepath.Clean(path)
	rest := ""
	for {
		if real, err := filepath.EvalSymlinks(path); err == nil {
			return filepath.Clean(filepath.Join(real, rest))
		}
		parent := filepath.Dir(path)
		if parent == path {
			return filepath.Clean(filepath.Join(path, rest))
		}
		rest = filepath.Join(filepath.Base(path), rest)
		path = parent
	}
}

// samePath reports whether a and b are the same path. Windows ignores
// case in paths, so samePath ignores it there too.
func samePath(a, b string) bool {
	if runtime.GOOS == "windows" {
		return strings.EqualFold(a, b)
	}
	return a == b
}
