package git_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/BertPratya/sliceflow/internal/git"
	"github.com/BertPratya/sliceflow/internal/gittest"
)

// Test 1: from the project's folder, a subfolder, a linked worktree, and
// a subfolder of the worktree, FindProject gives the main project's
// folder and name (R16, SD11). The worktree's folder has another name,
// so the test fails if the name comes from the folder FindProject was
// called in.
func TestFindProjectFromEverywhere(t *testing.T) {
	r := gittest.New(t)
	sub := mkdir(t, r.Dir, "src", "deep")
	worktree := r.AddWorktree("other")
	worktreeSub := mkdir(t, worktree, "src")

	for name, dir := range map[string]string{
		"project folder":      r.Dir,
		"subfolder":           sub,
		"linked worktree":     worktree,
		"worktree, subfolder": worktreeSub,
	} {
		t.Run(name, func(t *testing.T) {
			project, err := git.FindProject(context.Background(), dir)
			if err != nil {
				t.Fatalf("FindProject returned the error %v, want none", err)
			}
			wantSameFolder(t, project.Dir, r.Dir)
			if want := filepath.Base(realPath(t, r.Dir)); project.Name != want {
				t.Errorf("Name = %q, want %q", project.Name, want)
			}
			if !filepath.IsAbs(project.Dir) {
				t.Errorf("Dir = %q, want a full path", project.Dir)
			}
		})
	}
}

// Test 2: a folder that is not in a git repository, and a subfolder of
// it, give ErrNotARepo (R13).
func TestFindProjectNotARepo(t *testing.T) {
	gittest.Isolate(t)
	dir := t.TempDir()
	sub := mkdir(t, dir, "a", "b")

	for _, d := range []string{dir, sub} {
		_, err := git.FindProject(context.Background(), d)
		if !errors.Is(err, git.ErrNotARepo) {
			t.Errorf("FindProject(%s) returned the error %v, want ErrNotARepo", d, err)
		}
	}
}

// Test 3: in a repository that git refuses to read for another reason,
// FindProject returns git's own error, not ErrNotARepo: git exits with
// 128 here too, so this test fails if every 128 counts as "not a
// repository". The repository says it has a format version that git
// does not know.
func TestFindProjectOtherGitFailure(t *testing.T) {
	r := gittest.New(t)
	sub := mkdir(t, r.Dir, "src")
	r.SetConfig("core.repositoryformatversion", "99")

	for _, d := range []string{r.Dir, sub} {
		_, err := git.FindProject(context.Background(), d)
		wantGitErrorNotNotARepo(t, err)
	}
}

// Test 3b: a broken global config makes every git command fail, also in
// a real repository. FindProject returns git's own error, which names
// the broken file, not ErrNotARepo.
func TestFindProjectBrokenGlobalConfig(t *testing.T) {
	r := gittest.New(t)
	sub := mkdir(t, r.Dir, "src")
	gittest.BreakGlobalConfig(t)

	for _, d := range []string{r.Dir, sub} {
		_, err := git.FindProject(context.Background(), d)
		wantGitErrorNotNotARepo(t, err)
	}
}

// Test 3c: when GIT_DIR is set, git uses that repository. When it is not
// one, FindProject returns git's own error, which names it, not
// ErrNotARepo, even in a folder with no .git above it.
func TestFindProjectBadGitDir(t *testing.T) {
	gittest.Isolate(t)
	t.Setenv("GIT_DIR", filepath.Join(t.TempDir(), "missing"))

	_, err := git.FindProject(context.Background(), t.TempDir())
	wantGitErrorNotNotARepo(t, err)
}

// Test 3d: a .git file that points nowhere (a worktree whose main
// project was removed) gives git's own error, not ErrNotARepo. This is
// the known cost of looking for .git on disk.
func TestFindProjectBrokenGitFile(t *testing.T) {
	gittest.Isolate(t)
	dir := t.TempDir()
	gitFile := filepath.Join(dir, ".git")
	if err := os.WriteFile(gitFile, []byte("gitdir: "+filepath.Join(t.TempDir(), "gone")+"\n"), 0o644); err != nil {
		t.Fatalf("write %s: %v", gitFile, err)
	}

	_, err := git.FindProject(context.Background(), dir)
	wantGitErrorNotNotARepo(t, err)
}

// Test 4: a repository with an owner that git does not trust gives git's
// own error, not ErrNotARepo. GIT_TEST_ASSUME_DIFFERENT_OWNER is git's
// own switch for its tests; the test is skipped when this git does not
// know it.
func TestFindProjectUntrustedOwner(t *testing.T) {
	r := gittest.New(t)
	t.Setenv("GIT_TEST_ASSUME_DIFFERENT_OWNER", "1")

	_, err := git.FindProject(context.Background(), r.Dir)
	if err == nil {
		t.Skip("this git does not know GIT_TEST_ASSUME_DIFFERENT_OWNER")
	}
	wantGitErrorNotNotARepo(t, err)
}

// Test 4b: in a submodule, the project is the submodule's own folder,
// not the folder of its repository inside .git/modules (R16).
func TestFindProjectInASubmodule(t *testing.T) {
	lib := gittest.New(t)
	super := gittest.New(t)
	super.Git("-c", "protocol.file.allow=always", "submodule", "--quiet", "add", lib.Dir, "sub")
	sub := filepath.Join(super.Dir, "sub")
	subSub := mkdir(t, sub, "src")

	for _, d := range []string{sub, subSub} {
		project, err := git.FindProject(context.Background(), d)
		if err != nil {
			t.Fatalf("FindProject(%s) returned the error %v, want none", d, err)
		}
		wantSameFolder(t, project.Dir, sub)
		if project.Name != "sub" {
			t.Errorf("Name = %q, want %q", project.Name, "sub")
		}
	}

	project, err := git.FindProject(context.Background(), super.Dir)
	if err != nil {
		t.Fatalf("FindProject in the outer project returned the error %v, want none", err)
	}
	wantSameFolder(t, project.Dir, super.Dir)
}

// Test 4c: in a project made with git init --separate-git-dir, the
// repository is in another folder. The project is still the folder
// with the files (R16).
func TestFindProjectSeparateGitDir(t *testing.T) {
	gittest.Isolate(t)
	dir := mkdir(t, t.TempDir(), "shop")
	store := filepath.Join(t.TempDir(), "store.git")
	gittest.Git(t, dir, "init", "--quiet", "-b", "main", "--separate-git-dir", store)
	sub := mkdir(t, dir, "src")

	for _, d := range []string{dir, sub} {
		project, err := git.FindProject(context.Background(), d)
		if err != nil {
			t.Fatalf("FindProject(%s) returned the error %v, want none", d, err)
		}
		wantSameFolder(t, project.Dir, dir)
		if project.Name != "shop" {
			t.Errorf("Name = %q, want %q", project.Name, "shop")
		}
	}
}

// Test 5: a folder that does not exist gives the error as it is, not
// ErrNotARepo: git never started.
func TestFindProjectMissingFolder(t *testing.T) {
	gittest.Isolate(t)
	missing := filepath.Join(t.TempDir(), "missing")

	_, err := git.FindProject(context.Background(), missing)
	if err == nil || errors.Is(err, git.ErrNotARepo) {
		t.Errorf("FindProject returned the error %v, want the error of the missing folder", err)
	}
}

// Test 6: when git cannot be run, FindProject returns ErrNotFound (R14).
func TestFindProjectGitNotFound(t *testing.T) {
	r := gittest.New(t)
	t.Setenv("PATH", "")

	_, err := git.FindProject(context.Background(), r.Dir)
	if !errors.Is(err, git.ErrNotFound) {
		t.Errorf("FindProject returned the error %v, want ErrNotFound", err)
	}
}

// Test 7: HasCommits says no for a new repository, and yes after a
// commit (R15).
func TestHasCommits(t *testing.T) {
	r := gittest.NewEmpty(t)

	if got := hasCommits(t, r.Dir); got {
		t.Errorf("HasCommits of a new repo = true, want false")
	}
	r.Commit("a.txt", "a\n")
	if got := hasCommits(t, r.Dir); !got {
		t.Errorf("HasCommits after a commit = false, want true")
	}
}

// Test 8: HasCommits counts commits on any branch, not only on the
// branch that is checked out. Here the checked-out branch has no commit
// yet, but main has one: the repository has commits (R15).
func TestHasCommitsOnAnotherBranch(t *testing.T) {
	r := gittest.New(t)
	r.Git("checkout", "--quiet", "--orphan", "empty")

	if got := hasCommits(t, r.Dir); !got {
		t.Errorf("HasCommits = false, want true: main has a commit")
	}
}

// Test 9: outside a repository, HasCommits returns git's error, not
// "no commits".
func TestHasCommitsNotARepo(t *testing.T) {
	gittest.Isolate(t)

	got, err := git.HasCommits(context.Background(), t.TempDir())
	if err == nil {
		t.Errorf("HasCommits returned %v and no error, want an error", got)
	}
}

func hasCommits(t *testing.T, dir string) bool {
	t.Helper()
	got, err := git.HasCommits(context.Background(), dir)
	if err != nil {
		t.Fatalf("HasCommits returned the error %v, want none", err)
	}
	return got
}

// mkdir makes the folder dir/parts..., and returns its path.
func mkdir(t *testing.T, dir string, parts ...string) string {
	t.Helper()
	path := filepath.Join(append([]string{dir}, parts...)...)
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("make %s: %v", path, err)
	}
	return path
}

// realPath returns path with its links followed. git gives real paths:
// on macOS, for example, a temporary folder in /var is really in
// /private/var.
func realPath(t *testing.T, path string) string {
	t.Helper()
	real, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatalf("follow the links of %s: %v", path, err)
	}
	return real
}

// wantSameFolder fails the test unless got and want are the same folder.
func wantSameFolder(t *testing.T, got, want string) {
	t.Helper()
	if realPath(t, got) != realPath(t, want) {
		t.Errorf("Dir = %q, want %q", got, want)
	}
}

// wantGitErrorNotNotARepo fails the test unless err is git's own error
// (*git.Error), and not ErrNotARepo.
func wantGitErrorNotNotARepo(t *testing.T, err error) {
	t.Helper()
	if errors.Is(err, git.ErrNotARepo) {
		t.Fatalf("FindProject returned ErrNotARepo, want git's own error")
	}
	var gitErr *git.Error
	if !errors.As(err, &gitErr) {
		t.Fatalf("FindProject returned the error %v (%T), want a *git.Error", err, err)
	}
	if gitErr.Stderr == "" {
		t.Errorf("Stderr is empty, want git's message")
	}
}
