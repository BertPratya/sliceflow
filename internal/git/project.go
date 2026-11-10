package git

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// ErrNotARepo means the folder is not inside a git repository (R13).
var ErrNotARepo = errors.New("not in a git repository")

// Project is the main project that a folder belongs to (R16).
type Project struct {
	// Dir is the full path of the main project's folder, with the
	// system's separators (R9). From a linked worktree, it is still the
	// main project's folder, not the worktree's.
	Dir string
	// Name is the name of the main project's folder (R16, SD11).
	Name string
}

// FindProject asks git for the main project of the folder dir. dir may
// be the project's folder, any folder inside it, or any of its linked
// worktrees (R16).
//
// Steps:
//  1. Ask git for the repository of dir and for the main repository:
//     git rev-parse --path-format=absolute --git-dir --git-common-dir.
//  2. If git failed with an exit code, and isInRepo says dir is not in a
//     repository, return ErrNotARepo. Otherwise return the error as it
//     is, so the caller shows git's own words (design 1.17).
//  3. Check that git gave two full paths, with absPaths.
//  4. The two paths are the same: dir is in the main worktree. Ask git
//     rev-parse --path-format=absolute --show-toplevel, and use its
//     answer. This also holds in a submodule, and in a project made with
//     git init --separate-git-dir, where the repository is not in the
//     project's folder. In a bare repository, git fails here, and its
//     error is returned as it is.
//  5. The two paths differ: dir is in a linked worktree. The main
//     repository ends with .git, and the project's folder is the folder
//     that holds it. If it does not end with .git, the path itself is
//     used, as git worktree list does.
//
// Known limit (backlog): from a linked worktree of a submodule, or of a
// project made with --separate-git-dir, step 5 gives the folder of the
// repository, not of the project.
func FindProject(ctx context.Context, dir string) (Project, error) {
	out, err := run(ctx, dir, "rev-parse", "--path-format=absolute", "--git-dir", "--git-common-dir")
	var gitErr *Error
	if errors.As(err, &gitErr) {
		inRepo, confirmErr := isInRepo(dir)
		if confirmErr != nil {
			return Project{}, confirmErr
		}
		if !inRepo {
			return Project{}, ErrNotARepo
		}
		return Project{}, err
	}
	if err != nil {
		return Project{}, err
	}
	paths, err := absPaths(out, 2)
	if err != nil {
		return Project{}, err
	}
	gitDir, commonDir := paths[0], paths[1]

	var projectDir string
	if gitDir == commonDir {
		out, err := run(ctx, dir, "rev-parse", "--path-format=absolute", "--show-toplevel")
		if err != nil {
			return Project{}, err
		}
		top, err := absPaths(out, 1)
		if err != nil {
			return Project{}, err
		}
		projectDir = top[0]
	} else {
		projectDir = commonDir
		if filepath.Base(commonDir) == ".git" {
			projectDir = filepath.Dir(commonDir)
		}
	}
	return Project{Dir: projectDir, Name: filepath.Base(projectDir)}, nil
}

// absPaths reads git's answer to rev-parse: exactly n lines, each a full
// path. It returns them as clean paths of the system (R9).
//
// A git older than 2.31 does not know --path-format=absolute. It does
// not fail: it prints the option back as a line of its answer. So an
// answer with another number of lines, or with a line that is not a
// full path, is an error that quotes git's answer.
func absPaths(out string, n int) ([]string, error) {
	lines := strings.Split(trimNewline(out), "\n")
	if len(lines) != n {
		return nil, unexpectedAnswer(out)
	}
	paths := make([]string, n)
	for i, line := range lines {
		path := filepath.Clean(filepath.FromSlash(strings.TrimSuffix(line, "\r")))
		if !filepath.IsAbs(path) {
			return nil, unexpectedAnswer(out)
		}
		paths[i] = path
	}
	return paths, nil
}

// unexpectedAnswer is the error for an answer of git that absPaths
// cannot read.
func unexpectedAnswer(out string) error {
	return fmt.Errorf("git gave an unexpected answer to rev-parse: %q; sliceflow needs git 2.31 or later", out)
}

// isInRepo reports whether dir could be in a repository. FindProject
// calls it only after git has failed, to tell "not a repository" (R13)
// from other failures, without reading git's error text (design 1.17):
// git exits with 128 for both.
//
// It returns true when GIT_DIR is set (git then uses that repository,
// and its error says why it cannot), or when dir, or a folder above it,
// has a .git entry: a folder, or a file in a linked worktree or a
// submodule. An entry that cannot be checked, for example without
// permission, counts as there.
//
// It looks on disk with os.Lstat, and does not ask git again: whatever
// broke the first git call, such as a broken global config, would break
// a second one too, and hide git's message behind R13's.
//
// Cost: a .git file that points nowhere, for example in a worktree whose
// main project was removed, shows git's own text, not R13's.
func isInRepo(dir string) (bool, error) {
	if _, ok := os.LookupEnv("GIT_DIR"); ok {
		return true, nil
	}
	folder, err := filepath.Abs(dir)
	if err != nil {
		return false, err
	}
	for {
		_, err := os.Lstat(filepath.Join(folder, ".git"))
		if err == nil || !errors.Is(err, fs.ErrNotExist) {
			return true, nil
		}
		parent := filepath.Dir(folder)
		if parent == folder {
			return false, nil
		}
		folder = parent
	}
}

// HasCommits reports whether the repository of the folder dir has any
// commit at all, on any branch (R15). It asks git for one commit of all
// its refs: git rev-list -n 1 --all prints nothing, and exits with 0,
// when there is none.
func HasCommits(ctx context.Context, dir string) (bool, error) {
	out, err := run(ctx, dir, "rev-list", "-n", "1", "--all")
	if err != nil {
		return false, err
	}
	return trimNewline(out) != "", nil
}

// trimNewline removes the line break at the end of a line that git
// printed.
func trimNewline(s string) string {
	s = strings.TrimSuffix(s, "\n")
	return strings.TrimSuffix(s, "\r")
}
