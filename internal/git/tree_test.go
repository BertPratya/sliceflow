package git

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// Tests of DirectoryExists. Each one uses a repo made by
// newRepoWithSlices: a folder docs/slices/login, and a file
// docs/slices/notes.

// Test 1: a folder in the commit exists.
func TestDirectoryExistsForAFolder(t *testing.T) {
	repo, commit := newRepoWithSlices(t)
	wantDirectoryExists(t, repo, commit, "docs/slices/login", true)
}

// Test 2: a path that is not in the commit does not exist.
func TestDirectoryExistsForNoSuchPath(t *testing.T) {
	repo, commit := newRepoWithSlices(t)
	wantDirectoryExists(t, repo, commit, "docs/slices/nosuch", false)
}

// Test 3: a file is not a folder.
func TestDirectoryExistsForAFile(t *testing.T) {
	repo, commit := newRepoWithSlices(t)
	wantDirectoryExists(t, repo, commit, "docs/slices/notes", false)
}

// Test 4: DirectoryExists asks about the commit, not the disk. A folder
// that is on the disk, but not committed, does not exist.
func TestDirectoryExistsNotCommitted(t *testing.T) {
	repo, commit := newRepoWithSlices(t)
	writeFile(t, filepath.Join(repo, "docs", "slices", "draft", "spec.md"))
	wantDirectoryExists(t, repo, commit, "docs/slices/draft", false)
}

// Test 5: a folder added in a later commit does not exist in an earlier
// one.
func TestDirectoryExistsInAnEarlierCommit(t *testing.T) {
	repo, earlier := newRepoWithSlices(t)
	writeFile(t, filepath.Join(repo, "docs", "slices", "later", "spec.md"))
	gitIn(t, repo, "add", ".")
	gitIn(t, repo, "commit", "--quiet", "-m", "later")
	wantDirectoryExists(t, repo, earlier, "docs/slices/later", false)
}

// Test 6: path is from the repo's top folder, even when dir is a
// subfolder.
func TestDirectoryExistsFromASubfolder(t *testing.T) {
	repo, commit := newRepoWithSlices(t)
	sub := filepath.Join(repo, "docs")
	wantDirectoryExists(t, sub, commit, "docs/slices/login", true)
}

// Test 7: outside any repo, DirectoryExists returns false and git's
// error as it is: a *Error with exit code 128.
func TestDirectoryExistsOutsideARepo(t *testing.T) {
	dir := t.TempDir()
	// Stop git from looking for a repo in the folders above dir.
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(dir))

	got, err := DirectoryExists(context.Background(), dir, "HEAD", "docs/slices/login")
	if got {
		t.Errorf("DirectoryExists returned true with its error, want false")
	}
	gitErr := wantGitError(t, err)
	if gitErr.ExitCode != 128 {
		t.Errorf("the exit code is %d, want 128", gitErr.ExitCode)
	}
}

// newRepoWithSlices makes a repo with one commit that has a folder
// docs/slices/login and a file docs/slices/notes. It returns the repo's
// folder and the commit's ID.
func newRepoWithSlices(t *testing.T) (repo, commit string) {
	t.Helper()
	repo = newRepo(t)
	writeFile(t, filepath.Join(repo, "docs", "slices", "login", "spec.md"))
	writeFile(t, filepath.Join(repo, "docs", "slices", "notes"))
	gitIn(t, repo, "add", ".")
	gitIn(t, repo, "commit", "--quiet", "-m", "slices")
	return repo, gitOutput(t, repo, "rev-parse", "HEAD")
}

// writeFile makes the file path, and the folders above it, with a short
// text in it. It fails the test now if it cannot.
func writeFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("could not make the folder of %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte("text\n"), 0o644); err != nil {
		t.Fatalf("could not write %s: %v", path, err)
	}
}

// wantDirectoryExists fails the test unless DirectoryExists gives want for
// path in commit, with no error.
func wantDirectoryExists(t *testing.T, dir, commit, path string, want bool) {
	t.Helper()
	got, err := DirectoryExists(context.Background(), dir, commit, path)
	if err != nil {
		t.Fatalf("DirectoryExists(%q) returned the error %v, want none", path, err)
	}
	if got != want {
		t.Errorf("DirectoryExists(%q) returned %v, want %v", path, got, want)
	}
}

// FoldersAt lists the folders right inside a path, and no files; a path
// that is not there gives none.
func TestFoldersAt(t *testing.T) {
	repo, _ := newRepoWithSlices(t)
	gitIn(t, repo, "branch", "-M", "main")
	writeFile(t, filepath.Join(repo, "docs", "slices", "signup", "spec.md"))
	gitIn(t, repo, "add", ".")
	gitIn(t, repo, "commit", "--quiet", "-m", "signup")

	got, err := FoldersAt(context.Background(), repo, "main", "docs/slices")
	if err != nil {
		t.Fatalf("FoldersAt returned the error %v, want none", err)
	}
	if !slices.Equal(got, []string{"login", "signup"}) {
		t.Errorf("FoldersAt returned %q, want [login signup]", got)
	}

	got, err = FoldersAt(context.Background(), repo, "main", "nosuch")
	if err != nil || len(got) != 0 {
		t.Errorf("FoldersAt(nosuch) returned %q and the error %v, want none and none", got, err)
	}
}
