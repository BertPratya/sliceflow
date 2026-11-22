package git

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Test 1: from the top folder of a repo, RepoRoot returns that folder.
func TestRepoRootFromTheTopFolder(t *testing.T) {
	repo := newRepoWithCommit(t)

	got, err := RepoRoot(context.Background(), repo)
	if err != nil {
		t.Fatalf("RepoRoot returned the error %v, want none", err)
	}
	wantSamePath(t, got, repo)
}

// Test 2: from a subfolder, RepoRoot returns the top folder of the repo.
func TestRepoRootFromASubfolder(t *testing.T) {
	repo := newRepoWithCommit(t)
	sub := filepath.Join(repo, "internal", "api")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatalf("make the subfolder: %v", err)
	}

	got, err := RepoRoot(context.Background(), sub)
	if err != nil {
		t.Fatalf("RepoRoot returned the error %v, want none", err)
	}
	wantSamePath(t, got, repo)
}

// Test 3: from inside a worktree, RepoRoot returns the main repo's
// folder, not the worktree's (R16). The worktree is in a separate temp
// folder, outside the repo, as real worktrees are, so the test cannot
// pass by finding a folder above the worktree.
func TestRepoRootFromAWorktree(t *testing.T) {
	repo := newRepoWithCommit(t)
	worktree := filepath.Join(t.TempDir(), "login")
	gitIn(t, repo, "worktree", "add", "--quiet", "-b", "slice/login", worktree)

	got, err := RepoRoot(context.Background(), worktree)
	if err != nil {
		t.Fatalf("RepoRoot returned the error %v, want none", err)
	}
	wantSamePath(t, got, repo)
}

// Test 4: a repo with no commits has a root too. R15 (no commits) is
// checked after the repo is found, so RepoRoot must not fail here.
func TestRepoRootWithNoCommits(t *testing.T) {
	repo := newRepo(t)

	got, err := RepoRoot(context.Background(), repo)
	if err != nil {
		t.Fatalf("RepoRoot returned the error %v, want none", err)
	}
	wantSamePath(t, got, repo)
}

// Test 5: outside any repo, RepoRoot returns "" and git's error as it
// is: a *Error with exit code 128. The text is not checked: git
// translates it, and it can change between versions.
func TestRepoRootOutsideARepo(t *testing.T) {
	dir := t.TempDir()
	// Stop git from looking for a repo in the folders above dir.
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(dir))

	got, err := RepoRoot(context.Background(), dir)
	if got != "" {
		t.Errorf("RepoRoot returned %q with its error, want \"\"", got)
	}
	gitErr := wantGitError(t, err)
	if gitErr.ExitCode != 128 {
		t.Errorf("the exit code is %d, want 128", gitErr.ExitCode)
	}
}

// Test 6: from a subfolder, CheckoutTop returns the top folder of the
// checkout.
func TestCheckoutTopFromASubfolder(t *testing.T) {
	repo := newRepoWithCommit(t)
	sub := filepath.Join(repo, "internal", "api")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatalf("make the subfolder: %v", err)
	}

	got, err := CheckoutTop(context.Background(), sub)
	if err != nil {
		t.Fatalf("CheckoutTop returned the error %v, want none", err)
	}
	wantSamePath(t, got, repo)
}

// Test 7: from inside a worktree, CheckoutTop returns the worktree's
// folder, not the main repo's: this is where it differs from RepoRoot.
func TestCheckoutTopFromAWorktree(t *testing.T) {
	repo := newRepoWithCommit(t)
	worktree := filepath.Join(t.TempDir(), "second")
	gitIn(t, repo, "worktree", "add", "--quiet", "-b", "second", worktree)

	got, err := CheckoutTop(context.Background(), worktree)
	if err != nil {
		t.Fatalf("CheckoutTop returned the error %v, want none", err)
	}
	wantSamePath(t, got, worktree)
}

// Test 8: in the main checkout, GitPath gives the path inside .git.
func TestGitPathInTheMainCheckout(t *testing.T) {
	repo := newRepoWithCommit(t)

	got, err := GitPath(context.Background(), repo, "sliceflow/locks/login.lock")
	if err != nil {
		t.Fatalf("GitPath returned the error %v, want none", err)
	}
	wantSameGitPath(t, got, filepath.Join(repo, ".git"), filepath.Join("sliceflow", "locks", "login.lock"))
}

// Test 9: in a worktree, GitPath gives the path inside the worktree's own
// folder in .git, so each checkout has its own.
func TestGitPathInAWorktree(t *testing.T) {
	repo := newRepoWithCommit(t)
	worktree := filepath.Join(t.TempDir(), "second")
	gitIn(t, repo, "worktree", "add", "--quiet", "-b", "second", worktree)

	got, err := GitPath(context.Background(), worktree, "sliceflow/locks/login.lock")
	if err != nil {
		t.Fatalf("GitPath returned the error %v, want none", err)
	}
	wantSameGitPath(t, got, filepath.Join(repo, ".git", "worktrees", "second"), filepath.Join("sliceflow", "locks", "login.lock"))
}

// wantSameGitPath fails the test unless got is the path rest inside the
// folder gitDir, once links and short names in gitDir are resolved, as
// in wantSamePath.
func wantSameGitPath(t *testing.T, got, gitDir, rest string) {
	t.Helper()
	real, err := filepath.EvalSymlinks(gitDir)
	if err != nil {
		t.Fatalf("resolve %s: %v", gitDir, err)
	}
	if want := filepath.Join(real, rest); got != want {
		t.Errorf("the path is %q, want %q", got, want)
	}
}

// newRepoWithCommit makes a repo in a temp folder, with one empty
// commit, and returns its folder.
func newRepoWithCommit(t *testing.T) string {
	t.Helper()
	repo := newRepo(t)
	gitIn(t, repo, "commit", "--quiet", "--allow-empty", "-m", "first")
	return repo
}

// gitIn runs real git in dir to set up a test, and fails the test now if
// git fails. The user's own git settings for a name, an email, and
// signing could make a commit fail, so they are set here.
func gitIn(t *testing.T, dir string, args ...string) {
	t.Helper()
	full := append([]string{
		"-c", "user.name=sliceflow test",
		"-c", "user.email=test@example.com",
		"-c", "commit.gpgsign=false",
	}, args...)
	cmd := exec.Command("git", full...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// wantSamePath fails the test unless got is the folder want, once links
// and short names are resolved. On macOS, /tmp is a link to
// /private/tmp, and git gives the real path. On Windows, a temp folder
// can have a short name such as RUNNER~1, and git gives the long one.
func wantSamePath(t *testing.T, got, want string) {
	t.Helper()
	real, err := filepath.EvalSymlinks(want)
	if err != nil {
		t.Fatalf("resolve %s: %v", want, err)
	}
	if got != real {
		t.Errorf("the path is %q, want %q", got, real)
	}
}
