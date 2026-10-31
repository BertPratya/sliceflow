package slice

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/BertPratya/sliceflow/internal/spec"
)

// SpecResult is what ValidateSpec found. The cli prints it (0002 R2-R4).
type SpecResult struct {
	Name string // the slice's name
	// File is the full path of the spec, with the separators of the
	// system (R4).
	File string
	// Path is the spec's path relative to the current folder, with "/"
	// (R2).
	Path     string
	Problems []spec.Problem // nil when the spec has no problem
}

// ValidateSpec checks the spec of the slice name (0002 R1, R5-R8). dir is
// the current folder. When name is empty, the slice is the one of the
// current branch slice/<name>. The spec is read from the disk of the
// slice's worktree, also when it is not committed. Nothing is written
// (R36). A spec with problems is a result, not an error.
func ValidateSpec(dir, name string, today time.Time) (SpecResult, error) {
	// R5-R8, in that order; the first that fails stops.
	root, err := openRepo(dir)
	if err != nil {
		return SpecResult{}, err
	}
	if name == "" {
		name, err = currentSlice(root, "sliceflow spec validate <name>")
		if err != nil {
			return SpecResult{}, err
		}
	}
	worktree, err := findWorktree(root, name)
	if err != nil {
		return SpecResult{}, err
	}

	file := filepath.Join(resolved(worktree), "docs", "slices", name, "spec.md")
	path := relPath(dir, file)
	content, err := os.ReadFile(file)
	if errors.Is(err, fs.ErrNotExist) {
		return SpecResult{}, fmt.Errorf("spec not found: %s.", path)
	}
	if err != nil {
		// Assumed: the spec has no text for a read error other than "not
		// found", e.g. no permission, or a folder named spec.md. The path
		// is printed once: the reason is only the system's words.
		var pe *fs.PathError
		if errors.As(err, &pe) {
			err = pe.Err
		}
		return SpecResult{}, fmt.Errorf("could not read %s: %v.", path, err)
	}

	return SpecResult{
		Name:     name,
		File:     file,
		Path:     path,
		Problems: spec.Check(content, name, today),
	}, nil
}

// relPath returns file relative to dir, with "/" (R2). Both are resolved
// first, so a link in one of them (e.g. /tmp on macOS) does not give a
// path full of "..". When there is no relative path, e.g. another drive on
// Windows, it returns the full path, with "/".
//
// Assumed: the spec does not say what to print when there is no relative
// path.
func relPath(dir, file string) string {
	rel, err := filepath.Rel(resolved(dir), file)
	if err != nil {
		return filepath.ToSlash(file)
	}
	return filepath.ToSlash(rel)
}

// resolved returns path with its links followed. When that fails, it
// returns path as it is.
func resolved(path string) string {
	resolvedPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return path
	}
	return resolvedPath
}
