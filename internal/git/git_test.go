package git

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/BertPratya/sliceflow/internal/gittest"
)

func TestRootFromSubfolder(t *testing.T) {
	repo := gittest.Init(t)

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
	repo := gittest.Init(t)

	if got, err := HasCommits(repo); err != nil || got {
		t.Fatalf("HasCommits(new repo) = %v, %v, want false, nil", got, err)
	}

	gittest.Commit(t, repo, "first")

	if got, err := HasCommits(repo); err != nil || !got {
		t.Errorf("HasCommits(repo with a commit) = %v, %v, want true, nil", got, err)
	}
}

func TestConfig(t *testing.T) {
	repo := gittest.NewRepo(t)

	if value, ok, err := Config(repo, "sliceflow.base"); err != nil || ok {
		t.Fatalf("Config(not set) = %q, %v, %v, want \"\", false, nil", value, ok, err)
	}

	gittest.Run(t, repo, "config", "sliceflow.base", "develop")

	if value, ok, err := Config(repo, "sliceflow.base"); err != nil || !ok || value != "develop" {
		t.Errorf("Config(set) = %q, %v, %v, want \"develop\", true, nil", value, ok, err)
	}
}

func TestBranchExists(t *testing.T) {
	repo := gittest.NewRepo(t)
	gittest.Run(t, repo, "branch", "develop")

	tests := []struct {
		branch string
		want   bool
	}{
		{"main", true},
		{"develop", true},
		{"nope", false},
	}
	for _, tt := range tests {
		got, err := BranchExists(repo, tt.branch)
		if err != nil || got != tt.want {
			t.Errorf("BranchExists(%q) = %v, %v, want %v, nil", tt.branch, got, err, tt.want)
		}
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

func TestCommandErrorText(t *testing.T) {
	base := errors.New("exit status 1")
	tests := []struct {
		name string
		err  *CommandError
		want string
	}{
		{"first line", &CommandError{Command: "commit", ExitCode: 1, Lines: []string{"first", "second"}, Err: base}, "first"},
		{"no lines", &CommandError{Command: "commit", ExitCode: 128, Err: base}, "git commit failed (exit code 128)"},
		{"cannot start", &CommandError{Command: "commit", ExitCode: -1, Err: base}, "run git: exit status 1"},
	}
	for _, tt := range tests {
		if got := tt.err.Error(); got != tt.want {
			t.Errorf("%s: Error() = %q, want %q", tt.name, got, tt.want)
		}
		if !errors.Is(tt.err, base) {
			t.Errorf("%s: errors.Is does not find the underlying error", tt.name)
		}
	}
}

func TestGitErrorKeepsLines(t *testing.T) {
	repo := gittest.NewRepo(t)
	_, err := PathExists(repo, "no-such-branch", "x")

	var ce *CommandError
	if !errors.As(err, &ce) {
		t.Fatalf("error = %v, want a *CommandError", err)
	}
	if ce.Command != "ls-tree" || ce.ExitCode == 0 || len(ce.Lines) == 0 {
		t.Errorf("CommandError = %+v, want ls-tree, a non-zero code, and lines", ce)
	}
}

func TestMainRoot(t *testing.T) {
	repo := gittest.NewRepo(t)
	want := realPath(t, repo)

	got, err := MainRoot(repo)
	if err != nil || got != want {
		t.Fatalf("MainRoot(repo) = %q, %v, want %q, nil", got, err, want)
	}

	// From inside a linked worktree, MainRoot still gives the main checkout.
	wt := filepath.Join(t.TempDir(), "wt")
	gittest.Run(t, repo, "worktree", "add", "-b", "other", wt)
	got, err = MainRoot(wt)
	if err != nil || got != want {
		t.Errorf("MainRoot(worktree) = %q, %v, want %q, nil", got, err, want)
	}
}

func TestRemoteBranches(t *testing.T) {
	repo := gittest.NewRepo(t)

	got, err := RemoteBranches(repo, "slice/x")
	if err != nil || len(got) != 0 {
		t.Fatalf("RemoteBranches(no remotes) = %v, %v, want none, nil", got, err)
	}

	// Make remote-tracking branches by hand, as a fetch would.
	head := gittest.Run(t, repo, "rev-parse", "HEAD")
	gittest.Run(t, repo, "remote", "add", "origin", "https://example.com/o.git")
	gittest.Run(t, repo, "remote", "add", "a/b", "https://example.com/ab.git")
	for _, ref := range []string{
		"refs/remotes/origin/slice/x",
		"refs/remotes/a/b/slice/x",
		"refs/remotes/origin/feature/slice/x",
		"refs/remotes/origin/slice/y",
		"refs/remotes/origin/other",
	} {
		gittest.Run(t, repo, "update-ref", ref, head)
	}

	got, err = RemoteBranches(repo, "slice/x")
	want := []string{"a/b/slice/x", "origin/slice/x"}
	if err != nil || !slices.Equal(got, want) {
		t.Errorf("RemoteBranches = %v, %v, want %v, nil", got, err, want)
	}
}

func TestPathExists(t *testing.T) {
	repo := gittest.NewRepo(t)
	if err := os.MkdirAll(filepath.Join(repo, "docs", "slices", "a"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "docs", "slices", "a", "spec.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	gittest.Run(t, repo, "add", ".")
	gittest.Run(t, repo, "commit", "-m", "add a")
	gittest.Run(t, repo, "branch", "old", "HEAD~1")

	tests := []struct {
		rev, path string
		want      bool
	}{
		{"main", "docs/slices/a", true},
		{"main", "docs/slices/a/spec.md", true},
		{"main", "docs/slices/b", false},
		{"old", "docs/slices/a", false},
	}
	for _, tt := range tests {
		got, err := PathExists(repo, tt.rev, tt.path)
		if err != nil || got != tt.want {
			t.Errorf("PathExists(%q, %q) = %v, %v, want %v, nil", tt.rev, tt.path, got, err, tt.want)
		}
	}
}

func TestCreateBranchAndWorktree(t *testing.T) {
	repo := gittest.NewRepo(t)

	id, err := CommitID(repo, "main")
	if err != nil || len(id) != 40 {
		t.Fatalf("CommitID(main) = %q, %v, want 40 characters, nil", id, err)
	}
	if _, err := CommitID(repo, "nope"); err == nil {
		t.Error("CommitID(nope) error = nil, want an error")
	}

	if err := CreateBranch(repo, "slice/x", id); err != nil {
		t.Fatal(err)
	}
	if got := gittest.Run(t, repo, "rev-parse", "slice/x"); got != id {
		t.Errorf("slice/x is at %q, want %q", got, id)
	}
	if err := CreateBranch(repo, "slice/x", id); err == nil {
		t.Error("CreateBranch twice: error = nil, want an error")
	}

	wt := filepath.ToSlash(filepath.Join(t.TempDir(), "wt"))
	if err := AddWorktree(repo, wt, "slice/x"); err != nil {
		t.Fatal(err)
	}
	if got := gittest.Run(t, wt, "branch", "--show-current"); got != "slice/x" {
		t.Errorf("worktree branch = %q, want slice/x", got)
	}
}

func TestAddAndCommit(t *testing.T) {
	repo := gittest.NewRepo(t)
	before := gittest.Run(t, repo, "rev-parse", "HEAD")

	// A user's .gitignore must not keep the file out.
	if err := os.WriteFile(filepath.Join(repo, ".gitignore"), []byte("*.log\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "a.log"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Add(repo, "a.log"); err != nil {
		t.Fatal(err)
	}
	if err := Commit(repo, "x: add a.log"); err != nil {
		t.Fatal(err)
	}

	if got := gittest.Run(t, repo, "show", "--name-only", "--format=%s", "HEAD"); got != "x: add a.log\n\na.log" {
		t.Errorf("last commit = %q", got)
	}
	if got := gittest.Run(t, repo, "rev-parse", "HEAD~1"); got != before {
		t.Errorf("parent = %q, want %q", got, before)
	}

	// Nothing staged: git refuses, and we get a *CommandError.
	var ce *CommandError
	if err := Commit(repo, "empty"); !errors.As(err, &ce) {
		t.Errorf("Commit with nothing staged: error = %v, want a *CommandError", err)
	}
}
