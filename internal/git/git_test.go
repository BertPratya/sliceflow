package git

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
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

func TestCommandName(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{[]string{"commit", "-m", "x"}, "commit"},
		{[]string{"branch", "-D", "x"}, "branch"},
		{[]string{"worktree", "add", "p", "b"}, "worktree add"},
		{[]string{"worktree", "remove", "--force", "p"}, "worktree remove"},
	}
	for _, tt := range tests {
		if got := commandName(tt.args); got != tt.want {
			t.Errorf("commandName(%v) = %q, want %q", tt.args, got, tt.want)
		}
	}
}

func TestRemoveWorktreeAndDeleteBranch(t *testing.T) {
	repo := gittest.NewRepo(t)
	path := filepath.ToSlash(filepath.Join(t.TempDir(), "wt"))
	gittest.Run(t, repo, "branch", "slice/x")
	if err := AddWorktree(repo, path, "slice/x"); err != nil {
		t.Fatal(err)
	}
	// An untracked file does not stop --force.
	if err := os.WriteFile(filepath.Join(path, "junk"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := RemoveWorktree(repo, path); err != nil {
		t.Fatalf("RemoveWorktree: %v", err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("worktree folder still there: %v", err)
	}
	if err := DeleteBranch(repo, "slice/x"); err != nil {
		t.Fatalf("DeleteBranch: %v", err)
	}
	if ok, _ := BranchExists(repo, "slice/x"); ok {
		t.Error("branch still exists")
	}

	var ce *CommandError
	err := DeleteBranch(repo, "slice/x")
	if !errors.As(err, &ce) || ce.Command != "branch" {
		t.Errorf("DeleteBranch of a missing branch: error = %v, want a branch *CommandError", err)
	}
	err = RemoveWorktree(repo, path)
	if !errors.As(err, &ce) || ce.Command != "worktree remove" {
		t.Errorf("RemoveWorktree of a missing worktree: error = %v, want a worktree remove *CommandError", err)
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

func TestCurrentBranch(t *testing.T) {
	repo := gittest.NewRepo(t)

	got, err := CurrentBranch(repo)
	if err != nil || got != "main" {
		t.Fatalf("CurrentBranch = %q, %v, want main, nil", got, err)
	}

	gittest.Run(t, repo, "checkout", "--detach")
	got, err = CurrentBranch(repo)
	if err != nil || got != "" {
		t.Fatalf("CurrentBranch(detached) = %q, %v, want empty, nil", got, err)
	}
}

func TestShowFile(t *testing.T) {
	repo := gittest.NewRepo(t)
	file := filepath.Join(repo, "a.txt")
	if err := os.WriteFile(file, []byte("one\n\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gittest.Run(t, repo, "add", "a.txt")
	gittest.Run(t, repo, "commit", "-m", "add a")

	// The bytes come back as committed, with the new lines at the end.
	got, found, err := ShowFile(repo, "main", "a.txt")
	if err != nil || !found || string(got) != "one\n\n" {
		t.Fatalf("ShowFile = %q, %v, %v, want %q, true, nil", got, found, err, "one\n\n")
	}

	// An edit that is not committed is not read.
	if err := os.WriteFile(file, []byte("changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, _, _ = ShowFile(repo, "main", "a.txt")
	if string(got) != "one\n\n" {
		t.Errorf("ShowFile after an edit = %q, want the committed content", got)
	}

	got, found, err = ShowFile(repo, "main", "missing.txt")
	if err != nil || found || got != nil {
		t.Errorf("ShowFile(missing) = %q, %v, %v, want nil, false, nil", got, found, err)
	}
}

func TestCurrentBranchAmbiguousName(t *testing.T) {
	repo := gittest.NewRepo(t)
	gittest.Run(t, repo, "branch", "heads/main")
	gittest.Run(t, repo, "checkout", "heads/main")

	got, err := CurrentBranch(repo)
	if err != nil || got != "heads/main" {
		t.Fatalf("CurrentBranch = %q, %v, want heads/main, nil", got, err)
	}
}

func TestBranches(t *testing.T) {
	repo := gittest.NewRepo(t)
	for _, b := range []string{"slice/b", "slice/a", "slice/team/login", "other"} {
		gittest.Run(t, repo, "branch", b)
	}

	got, err := Branches(repo, "slice/")
	want := []string{"slice/a", "slice/b", "slice/team/login"}
	if err != nil || !slices.Equal(got, want) {
		t.Errorf("Branches(slice/) = %v, %v, want %v, nil", got, err, want)
	}

	got, err = Branches(repo, "nothing/")
	if err != nil || len(got) != 0 {
		t.Errorf("Branches(nothing/) = %v, %v, want empty, nil", got, err)
	}
}

func TestListDirs(t *testing.T) {
	repo := gittest.NewRepo(t)
	for _, p := range []string{"docs/slices/b/x.md", "docs/slices/a/y.md", "docs/slices/file.md"} {
		full := filepath.Join(repo, filepath.FromSlash(p))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	gittest.Run(t, repo, "add", "-A")
	gittest.Run(t, repo, "commit", "-m", "slices")

	got, err := ListDirs(repo, "refs/heads/main", "docs/slices")
	want := []string{"a", "b"}
	if err != nil || !slices.Equal(got, want) {
		t.Errorf("ListDirs = %v, %v, want %v, nil", got, err, want)
	}

	got, err = ListDirs(repo, "refs/heads/main", "docs/none")
	if err != nil || len(got) != 0 {
		t.Errorf("ListDirs(missing path) = %v, %v, want empty, nil", got, err)
	}
}

// TestNoNetworkCommands guards R40. It reads every non-test .go file under
// internal/ and fails if one has the text of a git command that talks to
// another computer, written as a string literal. It is a simple check on
// the text, not a proof, but it catches a command added by mistake.
func TestNoNetworkCommands(t *testing.T) {
	banned := []string{`"fetch"`, `"pull"`, `"push"`, `"clone"`, `"ls-remote"`}
	err := filepath.WalkDir("..", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, word := range banned {
			if strings.Contains(string(content), word) {
				t.Errorf("%s has %s: sliceflow must not use the network (R40)", path, word)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
