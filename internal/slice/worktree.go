package slice

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/BertPratya/sliceflow/internal/git"
)

// WorktreeDirKey is the git setting that chooses the folder for the
// worktrees (R18).
const WorktreeDirKey = "sliceflow.worktree.dir"

// WorktreePath returns the full path of the worktree of the slice name
// in project (R17, R18). It does not look at the folder on disk.
//
// Steps:
//  1. Read the setting WorktreeDirKey in the main project's folder, with
//     git.ConfigValue, so a value set for the project wins over a global
//     one. Any error is returned as it is.
//  2. Not set, or empty: the base folder is <home>/.sliceflow/worktrees
//     (R17), where <home> comes from os.UserHomeDir.
//  3. Set: the base folder is the value, made full by baseFolder (R18).
//  4. Return <base>/<project>/<name>, made clean by filepath.Join: no "."
//     or ".." parts, and the system's separators (R9).
func WorktreePath(ctx context.Context, project git.Project, name string) (string, error) {
	value, _, err := git.ConfigValue(ctx, project.Dir, WorktreeDirKey)
	if err != nil {
		return "", err
	}
	var base string
	if value == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".sliceflow", "worktrees")
	} else {
		base, err = baseFolder(value, project.Dir)
		if err != nil {
			return "", err
		}
	}
	return filepath.Join(base, project.Name, name), nil
}

// baseFolder turns the value of WorktreeDirKey into a full path (R18).
//
// Steps:
//  1. The value is ~, or starts with ~/ or ~\: replace the ~ and the
//     separator after it with the home folder. This holds on every
//     system, as R18 says. A ~ anywhere else is a plain character. R18
//     does not settle the rest of the value: on Linux and macOS, ~\a\b
//     gives a folder named a\b in the home folder, since \ is not a
//     separator there.
//  2. The value is a full path: use it as it is.
//  3. Windows only: the value has a drive but is not full, such as C:wt.
//     It is not handled (backlog): it is returned clean, and is never
//     put under the project's folder.
//  4. Windows only: the value starts with a separator but has no drive,
//     such as \wt or /wt. It gets the drive of the project's folder, as
//     Windows would give it the current drive.
//  5. Otherwise, the value is relative to the main project's folder,
//     projectDir.
func baseFolder(value, projectDir string) (string, error) {
	if value == "~" || strings.HasPrefix(value, "~/") || strings.HasPrefix(value, `~\`) {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		rest := ""
		if len(value) > 2 {
			rest = value[2:]
		}
		return filepath.Join(home, filepath.FromSlash(rest)), nil
	}
	path := filepath.FromSlash(value)
	switch {
	case filepath.IsAbs(path):
		return filepath.Clean(path), nil
	case filepath.VolumeName(path) != "":
		return filepath.Clean(path), nil
	case os.IsPathSeparator(path[0]):
		return filepath.Join(filepath.VolumeName(projectDir), path), nil
	default:
		return filepath.Join(projectDir, path), nil
	}
}
