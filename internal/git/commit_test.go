package git

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test 1: a repo with no commits has none.
func TestHasCommitsWithNoCommits(t *testing.T) {
	repo := newRepo(t)

	got, err := HasCommits(context.Background(), repo)
	if err != nil {
		t.Fatalf("HasCommits returned the error %v, want none", err)
	}
	if got {
		t.Errorf("HasCommits returned true, want false")
	}
}

// Test 2: a repo with one commit has one.
func TestHasCommitsWithACommit(t *testing.T) {
	repo := newRepoWithCommit(t)

	got, err := HasCommits(context.Background(), repo)
	if err != nil {
		t.Fatalf("HasCommits returned the error %v, want none", err)
	}
	if !got {
		t.Errorf("HasCommits returned false, want true")
	}
}

// Test 3: on an orphan branch, the current branch has no commit, but
// main does. R15 asks about the repo, so HasCommits returns true.
func TestHasCommitsOnAnOrphanBranch(t *testing.T) {
	repo := newRepoWithCommit(t)
	gitIn(t, repo, "checkout", "--quiet", "--orphan", "other")

	got, err := HasCommits(context.Background(), repo)
	if err != nil {
		t.Fatalf("HasCommits returned the error %v, want none", err)
	}
	if !got {
		t.Errorf("HasCommits returned false, want true: main has a commit")
	}
}

// Test 4: outside any repo, HasCommits returns false and git's error as
// it is: a *Error with exit code 128.
func TestHasCommitsOutsideARepo(t *testing.T) {
	dir := t.TempDir()
	// Stop git from looking for a repo in the folders above dir.
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(dir))

	got, err := HasCommits(context.Background(), dir)
	if got {
		t.Errorf("HasCommits returned true with its error, want false")
	}
	gitErr := wantGitError(t, err)
	if gitErr.ExitCode != 128 {
		t.Errorf("the exit code is %d, want 128", gitErr.ExitCode)
	}
}

// IsIgnored follows the .gitignore files of the repo, also one in a
// subfolder, for files that exist or not.
func TestIsIgnored(t *testing.T) {
	repo := newRepoWithCommit(t)
	writeText(t, filepath.Join(repo, ".gitignore"), "*.jsonl\n")
	writeText(t, filepath.Join(repo, "docs", ".gitignore"), "secret.md\n")
	writeFile(t, filepath.Join(repo, "docs", "spec.md"))

	tests := []struct {
		path string
		want bool
	}{
		{"events.jsonl", true},
		{"docs/slices/login/.events.jsonl", true},
		{"docs/secret.md", true},
		{"docs/spec.md", false},
		{"secret.md", false},
	}
	for _, tt := range tests {
		got, err := IsIgnored(context.Background(), repo, tt.path)
		if err != nil {
			t.Fatalf("IsIgnored(%q) returned the error %v, want none", tt.path, err)
		}
		if got != tt.want {
			t.Errorf("IsIgnored(%q) returned %v, want %v", tt.path, got, tt.want)
		}
	}
}

// Commit adds the two files, also one that an ignore rule matches, and
// only them, with the given message.
func TestCommitAddsOnlyTheGivenFiles(t *testing.T) {
	useTestIdentity(t)
	repo := newRepoWithCommit(t)
	writeText(t, filepath.Join(repo, ".gitignore"), ".events.jsonl\n")
	writeFile(t, filepath.Join(repo, "docs", "slices", "login", "spec.md"))
	writeFile(t, filepath.Join(repo, "docs", "slices", "login", ".events.jsonl"))
	writeFile(t, filepath.Join(repo, "other.txt"))

	err := Commit(context.Background(), repo, "login: add the spec and the event log",
		"docs/slices/login/spec.md", "docs/slices/login/.events.jsonl")
	if err != nil {
		t.Fatalf("Commit returned the error %v, want none", err)
	}

	files := gitOutput(t, repo, "show", "--name-only", "--format=", "HEAD")
	want := "docs/slices/login/.events.jsonl\ndocs/slices/login/spec.md"
	if files != want {
		t.Errorf("the commit has the files:\n%s\nwant:\n%s", files, want)
	}
	message := gitOutput(t, repo, "log", "-1", "--format=%B")
	if message != "login: add the spec and the event log\n" {
		t.Errorf("the commit message is %q, want %q", message, "login: add the spec and the event log\n")
	}
}

// Commit runs the user's pre-commit hook. When the hook fails, Commit
// returns git's error with what the hook printed, and makes no commit.
func TestCommitRunsAFailingHook(t *testing.T) {
	useTestIdentity(t)
	repo := newRepoWithCommit(t)
	before := gitOutput(t, repo, "rev-parse", "HEAD")
	writeText(t, filepath.Join(repo, ".git", "hooks", "pre-commit"), "#!/bin/sh\necho 'the hook said no' >&2\nexit 1\n")
	if err := os.Chmod(filepath.Join(repo, ".git", "hooks", "pre-commit"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(repo, "spec.md"))

	err := Commit(context.Background(), repo, "a message", "spec.md")
	gitErr := wantGitError(t, err)
	if !strings.Contains(gitErr.Stderr, "the hook said no") {
		t.Errorf("git's error output is %q, want it to have the hook's message", gitErr.Stderr)
	}
	if after := gitOutput(t, repo, "rev-parse", "HEAD"); after != before {
		t.Errorf("HEAD moved from %s to %s, want no new commit", before, after)
	}
}

// Commit returns git's error when a file does not exist, and makes no
// commit.
func TestCommitAFileThatDoesNotExist(t *testing.T) {
	useTestIdentity(t)
	repo := newRepoWithCommit(t)
	before := gitOutput(t, repo, "rev-parse", "HEAD")

	err := Commit(context.Background(), repo, "a message", "nosuch.md")
	wantGitError(t, err)
	if after := gitOutput(t, repo, "rev-parse", "HEAD"); after != before {
		t.Errorf("HEAD moved from %s to %s, want no new commit", before, after)
	}
}

// Commit commits only its own files. A change the user staged before
// stays staged and is not in the commit.
func TestCommitLeavesOtherStagedChanges(t *testing.T) {
	useTestIdentity(t)
	repo := newRepoWithCommit(t)
	writeFile(t, filepath.Join(repo, "app.go"))
	gitIn(t, repo, "add", "app.go")
	writeFile(t, filepath.Join(repo, "spec.md"))

	if err := Commit(context.Background(), repo, "a message", "spec.md"); err != nil {
		t.Fatalf("Commit returned the error %v, want none", err)
	}

	if files := gitOutput(t, repo, "show", "--name-only", "--format=", "HEAD"); files != "spec.md" {
		t.Errorf("the commit has the files:\n%s\nwant only spec.md", files)
	}
	if staged := gitOutput(t, repo, "diff", "--cached", "--name-only"); staged != "app.go" {
		t.Errorf("the staged files are:\n%s\nwant only app.go", staged)
	}
}

// During a merge, git refuses to commit only some files. Commit returns
// git's error, makes no commit, and the merge is still going on.
func TestCommitDuringAMerge(t *testing.T) {
	useTestIdentity(t)
	repo := newRepoWithCommit(t)
	gitIn(t, repo, "checkout", "--quiet", "-b", "other")
	writeFile(t, filepath.Join(repo, "other.txt"))
	gitIn(t, repo, "add", "other.txt")
	gitIn(t, repo, "commit", "--quiet", "-m", "other")
	gitIn(t, repo, "checkout", "--quiet", "-")
	gitIn(t, repo, "merge", "--quiet", "--no-ff", "--no-commit", "other")
	before := gitOutput(t, repo, "rev-parse", "HEAD")
	writeFile(t, filepath.Join(repo, "spec.md"))

	err := Commit(context.Background(), repo, "a message", "spec.md")
	wantGitError(t, err)
	if after := gitOutput(t, repo, "rev-parse", "HEAD"); after != before {
		t.Errorf("HEAD moved from %s to %s, want no new commit", before, after)
	}
	if _, err := os.Stat(filepath.Join(repo, ".git", "MERGE_HEAD")); err != nil {
		t.Errorf("the merge is not going on any more: %v", err)
	}
}

// useTestIdentity gives git a name and an email through the environment,
// and hides the user's own git settings, such as commit signing, for
// this test only, so Commit works the same on every machine.
func useTestIdentity(t *testing.T) {
	t.Helper()
	useTestConfig(t)
	t.Setenv("GIT_AUTHOR_NAME", "sliceflow test")
	t.Setenv("GIT_AUTHOR_EMAIL", "test@example.com")
	t.Setenv("GIT_COMMITTER_NAME", "sliceflow test")
	t.Setenv("GIT_COMMITTER_EMAIL", "test@example.com")
}

// writeText makes the file path, and the folders above it, with text in
// it. It fails the test now if it cannot.
func writeText(t *testing.T, path, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("could not make the folder of %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatalf("could not write %s: %v", path, err)
	}
}
