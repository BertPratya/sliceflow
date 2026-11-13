package project

import (
	"os"
	"path/filepath"
)

// WorktreePath returns the folder for the worktree of the slice name, in
// the project whose main folder is root (R17, R18). setting is the value
// of the git setting sliceflow.worktree.dir, or "" when it is not set or
// is empty.
//
// Steps:
//  1. If setting is "", the folder is <home>/.sliceflow/worktrees (R17).
//  2. If setting is "~", or "~" then a folder separator of this system,
//     the folder is the home folder joined with the rest of setting.
//  3. If setting is a full path, the folder is setting.
//  4. Otherwise, the folder is setting joined to root.
//  5. Return the folder joined with the project's name (the last part of
//     root) and name. Join cleans the result, so it has no . or .. parts.
//
// It looks up the home folder only in steps 1 and 2. When the home folder
// is unknown there, it returns the error of os.UserHomeDir as it is.
func WorktreePath(root, setting, name string) (string, error) {
	dir, err := worktreesDir(root, setting)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, filepath.Base(root), name), nil
}

// worktreesDir does steps 1 to 4 of WorktreePath.
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
