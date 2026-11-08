package slice_test

import (
	"context"
	"errors"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/BertPratya/sliceflow/internal/git"
	"github.com/BertPratya/sliceflow/internal/gittest"
	"github.com/BertPratya/sliceflow/internal/slice"
)

// The expected paths below are built from project.Dir and r.Home, the
// same strings that WorktreePath starts from. So they also match on
// Windows, where a temporary folder may have a short name such as
// RUNNER~1 in one path and the long name in another.

// findProject returns the project of the folder dir, and fails the test
// on an error.
func findProject(t *testing.T, dir string) git.Project {
	t.Helper()
	project, err := git.FindProject(context.Background(), dir)
	if err != nil {
		t.Fatalf("FindProject returned the error %v, want none", err)
	}
	return project
}

// worktreePath calls WorktreePath for the slice login, and fails the
// test on an error.
func worktreePath(t *testing.T, project git.Project) string {
	t.Helper()
	got, err := slice.WorktreePath(context.Background(), project, "login")
	if err != nil {
		t.Fatalf("WorktreePath returned the error %v, want none", err)
	}
	return got
}

// Test 1: when sliceflow.worktree.dir is not set, or set to "", the
// worktree goes in <home>/.sliceflow/worktrees/<project>/<name> (R17,
// R18).
func TestWorktreePathDefault(t *testing.T) {
	r := gittest.New(t)
	project := findProject(t, r.Dir)
	want := filepath.Join(r.Home, ".sliceflow", "worktrees", project.Name, "login")

	if got := worktreePath(t, project); got != want {
		t.Errorf("not set: WorktreePath = %q, want %q", got, want)
	}

	r.SetConfig(slice.WorktreeDirKey, "")
	if got := worktreePath(t, project); got != want {
		t.Errorf("empty: WorktreePath = %q, want %q", got, want)
	}
}

// Test 2: each form of the value that R18 allows gives <dir>/<project>/
// <name>, with no "." or ".." parts, and with the system's separators
// (R9): the expected paths are built with filepath.Join, so a "/" left
// in the result fails on Windows.
func TestWorktreePathFromTheSetting(t *testing.T) {
	r := gittest.New(t)
	project := findProject(t, r.Dir)
	full := t.TempDir()
	parent := filepath.Dir(project.Dir)

	tests := []struct {
		value string
		want  string
	}{
		{full, filepath.Join(full, project.Name, "login")},
		{full + "/a/./b/../c", filepath.Join(full, "a", "c", project.Name, "login")},
		{"~", filepath.Join(r.Home, project.Name, "login")},
		{"~/", filepath.Join(r.Home, project.Name, "login")},
		{"~/wt", filepath.Join(r.Home, "wt", project.Name, "login")},
		{`~\wt`, filepath.Join(r.Home, "wt", project.Name, "login")},
		{"~wt", filepath.Join(project.Dir, "~wt", project.Name, "login")},
		{"wt/~", filepath.Join(project.Dir, "wt", "~", project.Name, "login")},
		{".worktrees", filepath.Join(project.Dir, ".worktrees", project.Name, "login")},
		{"../shop-worktrees", filepath.Join(parent, "shop-worktrees", project.Name, "login")},
		{"./a/../b", filepath.Join(project.Dir, "b", project.Name, "login")},
	}
	if runtime.GOOS == "windows" {
		// A value that starts with a separator, with no drive, gets the
		// drive of the project's folder. It is not put under the
		// project's folder.
		volume := filepath.VolumeName(project.Dir)
		tests = append(tests,
			struct{ value, want string }{`\wt`, filepath.Join(volume+`\`, "wt", project.Name, "login")},
			struct{ value, want string }{"/wt", filepath.Join(volume+`\`, "wt", project.Name, "login")},
		)
	}
	for _, tt := range tests {
		r.SetConfig(slice.WorktreeDirKey, tt.value)
		got := worktreePath(t, project)
		if got != tt.want {
			t.Errorf("value %q: WorktreePath = %q, want %q", tt.value, got, tt.want)
		}
		if hasDotPart(got) {
			t.Errorf("value %q: WorktreePath = %q has a . or .. part", tt.value, got)
		}
		if runtime.GOOS == "windows" && strings.Contains(got, "/") {
			t.Errorf("value %q: WorktreePath = %q, want no /", tt.value, got)
		}
	}
}

// Test 3: a value set for the project wins over a global one, and a
// global one is used when the project has none (R18).
func TestWorktreePathProjectValueWins(t *testing.T) {
	r := gittest.New(t)
	project := findProject(t, r.Dir)
	global := t.TempDir()
	local := t.TempDir()

	gittest.SetGlobalConfig(t, slice.WorktreeDirKey, global)
	if got, want := worktreePath(t, project), filepath.Join(global, project.Name, "login"); got != want {
		t.Errorf("global only: WorktreePath = %q, want %q", got, want)
	}

	r.SetConfig(slice.WorktreeDirKey, local)
	if got, want := worktreePath(t, project), filepath.Join(local, project.Name, "login"); got != want {
		t.Errorf("project and global: WorktreePath = %q, want %q", got, want)
	}
}

// Test 4: a relative value starts from the main project's folder, also
// when the project was found from a linked worktree (R18, SD19).
func TestWorktreePathRelativeFromAWorktree(t *testing.T) {
	r := gittest.New(t)
	worktree := r.AddWorktree("other")
	r.SetConfig(slice.WorktreeDirKey, ".worktrees")
	main := findProject(t, r.Dir)

	got := worktreePath(t, findProject(t, worktree))
	want := filepath.Join(main.Dir, ".worktrees", main.Name, "login")
	if got != want {
		t.Errorf("WorktreePath = %q, want %q", got, want)
	}
}

// Test 5: when git cannot read the setting for a reason other than "not
// set", here a broken config file, WorktreePath returns git's error, and
// does not fall back to the default folder.
func TestWorktreePathGitFails(t *testing.T) {
	r := gittest.New(t)
	project := findProject(t, r.Dir)
	r.BreakConfig()

	got, err := slice.WorktreePath(context.Background(), project, "login")
	var gitErr *git.Error
	if !errors.As(err, &gitErr) {
		t.Errorf("WorktreePath = %q, %v (%T), want a *git.Error", got, err, err)
	}
}

// hasDotPart reports whether path has a part that is . or ..
func hasDotPart(path string) bool {
	for _, part := range strings.Split(filepath.ToSlash(path), "/") {
		if part == "." || part == ".." {
			return true
		}
	}
	return false
}
