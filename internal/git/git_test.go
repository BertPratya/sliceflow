package git

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestRootFromSubfolder(t *testing.T) {
	repo := t.TempDir()
	runGit(t, repo, "init")

	sub := filepath.Join(repo, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := Root(sub)
	if err != nil {
		t.Fatalf("Root(%q) error = %v, want nil", sub, err)
	}

	if want := realPath(t, repo); got != want {
		t.Errorf("Root(%q) = %q, want %q", sub, got, want)
	}
}

func TestHasCommits(t *testing.T) {
	repo := t.TempDir()
	runGit(t, repo, "init", "-b", "main")

	if got, err := HasCommits(repo); err != nil || got {
		t.Fatalf("HasCommits(new repo) = %v, %v, want false, nil", got, err)
	}

	emptyCommit(t, repo)

	if got, err := HasCommits(repo); err != nil || !got {
		t.Errorf("HasCommits(repo with a commit) = %v, %v, want true, nil", got, err)
	}
}

// emptyCommit makes a commit with no files in the repo at dir. It sets its
// own name and email, and turns off signing, so it works on any computer.
func emptyCommit(t *testing.T, dir string) {
	t.Helper()
	runGit(t, dir,
		"-c", "user.name=sliceflow test", "-c", "user.email=test@example.com",
		"-c", "commit.gpgsign=false",
		"commit", "--allow-empty", "-m", "first")
}

// runGit runs git with args in dir, and stops the test if git fails.
func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	if out, err := command(dir, args...).CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// realPath returns path the way git prints it: with links resolved (on
// macOS the temp folder /var is a link to /private/var), and with / between
// the parts, also on Windows.
func realPath(t *testing.T, path string) string {
	t.Helper()
	real, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatal(err)
	}
	return filepath.ToSlash(real)
}

func TestRootNotRepo(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(dir))
	root, err := Root(dir)

	if !errors.Is(err, ErrNotRepo) {
		t.Fatalf("Root(%q) error = %v, want ErrNotRepo", dir, err)
	}

	if root != "" {
		t.Fatalf("Root(%q) = %q, want an empty root", dir, root)
	}
}

func TestHasCommitsNotRepo(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(dir))

	got, err := HasCommits(dir)

	if got {
		t.Errorf("expected false, got %v", got)
	}

	if err == nil {
		t.Fatal("expected an error, got nil")
	}

	t.Logf("error: %v", err)
}
