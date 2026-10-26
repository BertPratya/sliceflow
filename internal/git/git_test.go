package git

import (
	"errors"
	"os"
	"path/filepath"
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
