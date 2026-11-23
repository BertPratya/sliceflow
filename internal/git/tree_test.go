package git

import (
	"context"
	"os"
	"path/filepath"
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

// PathExists finds a file and a folder in a commit, by name or by ID, and
// not a path that is only on the disk.
func TestPathExists(t *testing.T) {
	repo := newRepo(t)
	writeFile(t, filepath.Join(repo, "docs", "slices", "login", ".events.jsonl"))
	gitIn(t, repo, "add", ".")
	gitIn(t, repo, "commit", "--quiet", "-m", "login")
	writeFile(t, filepath.Join(repo, "docs", "slices", "signup", ".events.jsonl"))

	tests := []struct {
		path string
		want bool
	}{
		{"docs/slices/login/.events.jsonl", true},
		{"docs/slices/login", true},
		{"docs/slices/signup/.events.jsonl", false},
		{"docs/slices/nosuch", false},
	}
	for _, tt := range tests {
		got, err := PathExists(context.Background(), repo, "HEAD", tt.path)
		if err != nil {
			t.Fatalf("PathExists(%q) returned the error %v, want none", tt.path, err)
		}
		if got != tt.want {
			t.Errorf("PathExists(%q) returned %v, want %v", tt.path, got, tt.want)
		}
	}
}

// IsAncestor is true for an older commit of the same branch and for the
// commit itself, and false for a commit that the other does not hold.
func TestIsAncestor(t *testing.T) {
	repo := newRepoWithCommit(t)
	gitIn(t, repo, "branch", "old")
	gitIn(t, repo, "commit", "--quiet", "--allow-empty", "-m", "second")
	gitIn(t, repo, "switch", "--quiet", "-c", "side", "old")
	gitIn(t, repo, "commit", "--quiet", "--allow-empty", "-m", "on the side")
	main := gitOutput(t, repo, "rev-parse", "HEAD@{2}")

	tests := []struct {
		commit, of string
		want       bool
	}{
		{"refs/heads/old", "HEAD", true},
		{"HEAD", "HEAD", true},
		{main, "HEAD", false},
	}
	for _, tt := range tests {
		got, err := IsAncestor(context.Background(), repo, tt.commit, tt.of)
		if err != nil {
			t.Fatalf("IsAncestor(%q, %q) returned the error %v, want none", tt.commit, tt.of, err)
		}
		if got != tt.want {
			t.Errorf("IsAncestor(%q, %q) returned %v, want %v", tt.commit, tt.of, got, tt.want)
		}
	}
}

// IsAncestor returns git's error for a commit that does not exist.
func TestIsAncestorOfNoSuchCommit(t *testing.T) {
	repo := newRepoWithCommit(t)

	_, err := IsAncestor(context.Background(), repo, "refs/heads/nosuch", "HEAD")
	wantGitError(t, err)
}
