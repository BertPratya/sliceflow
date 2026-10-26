package slice

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BertPratya/sliceflow/internal/gittest"
)

const notChanged = ". Nothing was changed."

// createErr runs Create and returns the text of its error.
func createErr(t *testing.T, dir, name, from string) string {
	t.Helper()
	_, err := Create(CreateOptions{Dir: dir, Name: name, From: from})
	if err == nil {
		t.Fatal("Create error = nil, want an error")
	}
	return err.Error()
}

func TestCreateNotRepo(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(dir))

	// R12 comes before R14: the bad name is not reported.
	want := "not inside a git repository" + notChanged
	if got := createErr(t, dir, "My Slice", ""); got != want {
		t.Errorf("error = %q, want %q", got, want)
	}
}

func TestCreateNoCommits(t *testing.T) {
	repo := gittest.Init(t)

	want := "the repository has no commits yet. Make a first commit, then try again" + notChanged
	if got := createErr(t, repo, "My Slice", ""); got != want {
		t.Errorf("error = %q, want %q", got, want)
	}
}

func TestCreateInvalidName(t *testing.T) {
	repo := gittest.NewRepo(t)

	// R14 comes before R16: the missing base branch is not reported.
	want := `invalid slice name "My Slice": use 1-64 lowercase letters, digits, and dashes, starting with a letter or digit` + notChanged
	if got := createErr(t, repo, "My Slice", "nope"); got != want {
		t.Errorf("error = %q, want %q", got, want)
	}

	want = `invalid slice name "con": it is a reserved name on Windows` + notChanged
	if got := createErr(t, repo, "con", ""); got != want {
		t.Errorf("error = %q, want %q", got, want)
	}
}

func TestCreateMissingBase(t *testing.T) {
	repo := gittest.NewRepo(t)
	gittest.Run(t, repo, "branch", "slice/x")

	// R16 comes before R15: the existing branch is not reported.
	want := "base branch develop does not exist (from --from)" + notChanged
	if got := createErr(t, repo, "x", "develop"); got != want {
		t.Errorf("error = %q, want %q", got, want)
	}
}

func TestCreateNameExists(t *testing.T) {
	repo := gittest.NewRepo(t)
	head := gittest.Run(t, repo, "rev-parse", "HEAD")

	if err := os.MkdirAll(filepath.Join(repo, "docs", "slices", "x"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "docs", "slices", "x", "spec.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	gittest.Run(t, repo, "add", ".")
	gittest.Run(t, repo, "commit", "-m", "add x")

	// Only the folder: the third place.
	want := "slice x already exists: folder docs/slices/x on main" + notChanged
	if got := createErr(t, repo, "x", ""); got != want {
		t.Errorf("folder: error = %q, want %q", got, want)
	}

	// A remote branch comes before the folder.
	gittest.Run(t, repo, "remote", "add", "origin", "https://example.com/o.git")
	gittest.Run(t, repo, "update-ref", "refs/remotes/origin/slice/x", head)
	want = "slice x already exists: branch origin/slice/x" + notChanged
	if got := createErr(t, repo, "x", ""); got != want {
		t.Errorf("remote: error = %q, want %q", got, want)
	}

	// A local branch comes before both.
	gittest.Run(t, repo, "branch", "slice/x")
	want = "slice x already exists: branch slice/x" + notChanged
	if got := createErr(t, repo, "x", ""); got != want {
		t.Errorf("local: error = %q, want %q", got, want)
	}
}

func TestCreateOldWorktreeFolder(t *testing.T) {
	repo := gittest.NewRepo(t)
	if err := os.MkdirAll(filepath.Join(repo, ".sliceflow", "worktrees", "x"), 0o755); err != nil {
		t.Fatal(err)
	}

	want := "folder .sliceflow/worktrees/x already exists. Remove it, or pick another name" + notChanged
	if got := createErr(t, repo, "x", ""); got != want {
		t.Errorf("error = %q, want %q", got, want)
	}
}

func TestCreateChecksPass(t *testing.T) {
	repo := gittest.NewRepo(t)

	// For now, all checks passing ends in "not built yet" (chunk 4).
	want := "creating a slice is not built yet"
	if got := createErr(t, repo, "x", ""); got != want {
		t.Errorf("error = %q, want %q", got, want)
	}
}

// TestCreateOrderPairs puts two problems in one repo and checks that the
// earlier check of R11 wins. R12 vs R13 cannot be tested: a folder that is
// not a repo cannot also be a repo with no commits.
func TestCreateOrderPairs(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T) string // makes the repo with both problems
		slice string
		from  string
		want  string // the start of the error that must win
	}{
		{
			name:  "R13 before R14",
			setup: gittest.Init,
			slice: "My Slice",
			want:  "the repository has no commits yet",
		},
		{
			name:  "R14 before R16",
			setup: gittest.NewRepo,
			slice: "My Slice",
			from:  "nope",
			want:  `invalid slice name "My Slice"`,
		},
		{
			name: "R16 before R15",
			setup: func(t *testing.T) string {
				repo := gittest.NewRepo(t)
				gittest.Run(t, repo, "branch", "slice/x")
				return repo
			},
			slice: "x",
			from:  "nope",
			want:  "base branch nope does not exist",
		},
		{
			name: "R15 before R17",
			setup: func(t *testing.T) string {
				repo := gittest.NewRepo(t)
				gittest.Run(t, repo, "branch", "slice/x")
				if err := os.MkdirAll(filepath.Join(repo, ".sliceflow", "worktrees", "x"), 0o755); err != nil {
					t.Fatal(err)
				}
				return repo
			},
			slice: "x",
			want:  "slice x already exists",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := createErr(t, tt.setup(t), tt.slice, tt.from)
			if !strings.HasPrefix(got, tt.want) {
				t.Errorf("error = %q, want it to start with %q", got, tt.want)
			}
		})
	}
}
