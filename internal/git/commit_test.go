package git_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/BertPratya/sliceflow/internal/git"
	"github.com/BertPratya/sliceflow/internal/gittest"
)

// Test 1: IsIgnored says yes for a file that a .gitignore rule matches,
// and no for one that no rule matches (R5).
func TestIsIgnored(t *testing.T) {
	r := gittest.New(t)
	r.WriteFile(".gitignore", "*.jsonl\n")

	for path, want := range map[string]bool{"a.jsonl": true, "a.md": false} {
		got, err := git.IsIgnored(context.Background(), r.Dir, path)
		if err != nil {
			t.Errorf("IsIgnored(%q) returned the error %v, want none", path, err)
			continue
		}
		if got != want {
			t.Errorf("IsIgnored(%q) = %v, want %v", path, got, want)
		}
	}
}

// Test 1b: outside a repository, git fails with a code that is not 1.
// IsIgnored returns git's error, not "no".
func TestIsIgnoredNotARepo(t *testing.T) {
	gittest.Isolate(t)
	dir := t.TempDir()
	// Stop git from looking for a repo in the folders above dir.
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(dir))

	got, err := git.IsIgnored(context.Background(), dir, "a.jsonl")

	var gitErr *git.Error
	if !errors.As(err, &gitErr) {
		t.Errorf("IsIgnored = %v, %v (%T), want a *git.Error", got, err, err)
	}
}

// Test 2: AddForce adds the files it is given, even one that .gitignore
// matches, and only those; Commit makes a commit with the message (R5).
func TestAddForceAndCommit(t *testing.T) {
	r := gittest.New(t)
	ctx := context.Background()
	r.WriteFile(".gitignore", "spec.md\n")
	r.WriteFile("spec.md", "spec\n")
	r.WriteFile("b.md", "b\n")
	message := "login: add the spec and the event log"

	if err := git.AddForce(ctx, r.Dir, "spec.md", "b.md"); err != nil {
		t.Fatalf("AddForce returned the error %v, want none", err)
	}
	if err := git.Commit(ctx, r.Dir, message); err != nil {
		t.Fatalf("Commit returned the error %v, want none", err)
	}

	if got := r.Git("log", "-1", "--format=%s"); got != message {
		t.Errorf("the commit's message is %q, want %q", got, message)
	}
	got := strings.Split(r.Git("ls-tree", "-r", "--name-only", "HEAD"), "\n")
	want := []string{"README.md", "b.md", "spec.md"}
	if !slices.Equal(got, want) {
		t.Errorf("the commit holds %q, want %q", got, want)
	}
}

// Test 3: Commit runs the user's hooks (R6). A pre-commit hook that
// fails stops the commit, and Commit returns git's error with the hook's
// text (R37).
func TestCommitRunsHooks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("a shell hook may not run on Windows")
	}
	r := gittest.New(t)
	ctx := context.Background()
	hook := filepath.Join(r.Dir, ".git", "hooks", "pre-commit")
	mkdir(t, filepath.Dir(hook))
	script := "#!/bin/sh\necho 'the hook said no' >&2\nexit 1\n"
	if err := os.WriteFile(hook, []byte(script), 0o755); err != nil {
		t.Fatalf("write the hook: %v", err)
	}
	r.WriteFile("spec.md", "spec\n")
	if err := git.AddForce(ctx, r.Dir, "spec.md"); err != nil {
		t.Fatalf("AddForce returned the error %v, want none", err)
	}

	err := git.Commit(ctx, r.Dir, "login: add the spec and the event log")

	var gitErr *git.Error
	if !errors.As(err, &gitErr) {
		t.Fatalf("Commit returned %v (%T), want a *git.Error", err, err)
	}
	if !strings.Contains(gitErr.Stderr, "the hook said no") {
		t.Errorf("Stderr = %q, want it to contain the hook's text", gitErr.Stderr)
	}
}

// Test 4: ExcludeFile gives the same full path, the main repository's
// info/exclude, from the main folder and from a linked worktree (R2).
func TestExcludeFile(t *testing.T) {
	r := gittest.New(t)
	worktree := r.AddWorktree("slice/login")

	fromMain, err := git.ExcludeFile(context.Background(), r.Dir)
	if err != nil {
		t.Fatalf("ExcludeFile from the main folder returned the error %v, want none", err)
	}
	fromWorktree, err := git.ExcludeFile(context.Background(), worktree)
	if err != nil {
		t.Fatalf("ExcludeFile from the worktree returned the error %v, want none", err)
	}

	if fromMain != fromWorktree {
		t.Errorf("ExcludeFile from the main folder = %q, from the worktree = %q, want the same", fromMain, fromWorktree)
	}
	if !filepath.IsAbs(fromMain) {
		t.Errorf("ExcludeFile = %q, want a full path", fromMain)
	}
	if !strings.HasSuffix(fromMain, filepath.Join("info", "exclude")) {
		t.Errorf("ExcludeFile = %q, want it to end in info/exclude", fromMain)
	}
	wantSameFolder(t, filepath.Dir(filepath.Dir(fromMain)), filepath.Join(r.Dir, ".git"))
}
