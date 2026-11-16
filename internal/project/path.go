package project

import (
	"os"
	"path/filepath"
)

// WorktreePath returns the folder for the worktree of the slice name, in
// the project whose main folder is root:
// <worktrees>/<project>/<name>, where <project> is the last part of root.
// The path has no . or .. parts.
//
// setting is the value of the git setting sliceflow.worktree.dir, or ""
// when it is not set or is empty. <worktrees> is:
//   - <home>/.sliceflow/worktrees when setting is "";
//   - in the home folder when setting is "~", or starts with "~" and a
//     folder separator, e.g. ~/wt;
//   - setting itself when it is a full path;
//   - setting joined to root otherwise, e.g. ../wt.
//
// It needs the home folder only for the first two. When the home folder
// is unknown then, it returns the error of os.UserHomeDir as it is.
func WorktreePath(root, setting, name string) (string, error) {
	dir, err := worktreesDir(root, setting)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, filepath.Base(root), name), nil
}

// worktreesDir returns the <worktrees> folder of WorktreePath for
// setting.
func worktreesDir(root, setting string) (string, error) {
	if setting == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".sliceflow", "worktrees"), nil
	}
	if startsWithHome(setting) {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, setting[1:]), nil
	}
	if filepath.IsAbs(setting) {
		return setting, nil
	}
	return filepath.Join(root, setting), nil
}

// startsWithHome reports whether setting is "~", or "~" then a folder
// separator: "/" on every system, and also "\" on Windows.
func startsWithHome(setting string) bool {
	if setting == "~" {
		return true
	}
	return len(setting) >= 2 && setting[0] == '~' && os.IsPathSeparator(setting[1])
}
