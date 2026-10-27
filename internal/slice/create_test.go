package slice

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/BertPratya/sliceflow/internal/gittest"
)

const notChanged = ". Nothing was changed."

// createErr runs Create and returns the text of its error.
func createErr(t *testing.T, dir, name, from string) string {
	t.Helper()
	_, err := Create(context.Background(), CreateOptions{Dir: dir, Name: name, From: from})
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

	// When every check passes, the slice is created.
	got, err := Create(context.Background(), CreateOptions{Dir: repo, Name: "x", Now: time.Date(2026, 10, 24, 7, 42, 0, 0, time.UTC)})
	if err != nil {
		t.Fatalf("Create error = %v, want nil", err)
	}
	if got.Name != "x" || got.Branch != "slice/x" || got.From.Branch != "main" || len(got.From.Commit) != 40 {
		t.Errorf("Result = %+v", got)
	}
	if got.WorktreeRel != ".sliceflow/worktrees/x" || got.Spec != "docs/slices/x/spec.md" {
		t.Errorf("Result paths = %q, %q", got.WorktreeRel, got.Spec)
	}
	if !strings.HasSuffix(got.Worktree, "/.sliceflow/worktrees/x") || strings.Contains(got.Worktree, `\`) {
		t.Errorf("Worktree = %q", got.Worktree)
	}
	if out := gittest.Run(t, repo, "log", "-1", "--format=%s", "slice/x"); out != "x: slice created" {
		t.Errorf("last commit on slice/x = %q", out)
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

// TestCreateTimeIsUTC: 23:30 at UTC-5 is already the next day in UTC, for
// the spec's date and for the event's time.
func TestCreateTimeIsUTC(t *testing.T) {
	repo := gittest.NewRepo(t)
	now := time.Date(2026, 10, 24, 23, 30, 0, 0, time.FixedZone("UTC-5", -5*3600))

	if _, err := Create(context.Background(), CreateOptions{Dir: repo, Name: "x", Now: now}); err != nil {
		t.Fatal(err)
	}

	spec := gittest.Run(t, repo, "show", "slice/x:docs/slices/x/spec.md")
	if !strings.HasPrefix(spec, "---\nslice: x\ncreated: 2026-10-25\n") {
		t.Errorf("spec starts with %q", spec[:60])
	}
	event := gittest.Run(t, repo, "show", "slice/x:docs/slices/x/.events.jsonl")
	if !strings.Contains(event, `"at":"2026-10-25T04:30:00Z"`) {
		t.Errorf("event = %s, want at 2026-10-25T04:30:00Z", event)
	}
}

// writeHook writes a pre-commit hook into the repo.
func writeHook(t *testing.T, repo, body string) {
	t.Helper()
	path := filepath.Join(repo, ".git", "hooks", "pre-commit")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
}

func TestCreateFailed(t *testing.T) {
	tests := []struct {
		name       string
		reason     string
		notRemoved []string
		details    []string
		want       string
	}{
		{"removed", "git commit failed (exit code 1)", nil, nil,
			"could not create slice x: git commit failed (exit code 1). Everything sliceflow made was removed."},
		{"not removed", "boom", []string{"branch slice/x", "folder .sliceflow"}, nil,
			"could not create slice x: boom. Could not remove: branch slice/x, folder .sliceflow."},
		{"details", "first", nil, []string{"second", "third"},
			"could not create slice x: first. Everything sliceflow made was removed.\n  second\n  third"},
		{"not removed and details", "first", []string{"worktree w"}, []string{"second"},
			"could not create slice x: first. Could not remove: worktree w.\n  second"},
	}
	for _, tt := range tests {
		got := createFailed("x", tt.reason, tt.notRemoved, tt.details).Error()
		if got != tt.want {
			t.Errorf("%s: got %q, want %q", tt.name, got, tt.want)
		}
	}
}

// TestCreateCommitFailsUndoesAll: a hook that refuses the commit leaves
// nothing behind, not even the new .sliceflow folder.
func TestCreateCommitFailsUndoesAll(t *testing.T) {
	repo := gittest.NewRepo(t)
	writeHook(t, repo, "exit 1")

	want := "could not create slice x: git commit failed (exit code 1). Everything sliceflow made was removed."
	if got := createErr(t, repo, "x", ""); got != want {
		t.Errorf("error = %q, want %q", got, want)
	}
	if out := gittest.Run(t, repo, "branch", "--list", "slice/x"); out != "" {
		t.Errorf("branch left: %q", out)
	}
	if out := gittest.Run(t, repo, "worktree", "list", "--porcelain"); strings.Count(out, "worktree ") != 1 {
		t.Errorf("worktrees left: %q", out)
	}
	if _, err := os.Stat(filepath.Join(repo, ".sliceflow")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf(".sliceflow left: %v", err)
	}
}

// TestCreateLaterFailureKeepsOlderSlice: the second create fails, and the
// first slice, .sliceflow, and .gitignore stay.
func TestCreateLaterFailureKeepsOlderSlice(t *testing.T) {
	repo := gittest.NewRepo(t)
	if _, err := Create(context.Background(), CreateOptions{Dir: repo, Name: "a", Now: time.Now()}); err != nil {
		t.Fatal(err)
	}
	writeHook(t, repo, "exit 1")

	createErr(t, repo, "b", "")
	for _, p := range []string{".sliceflow", ".sliceflow/.gitignore", ".sliceflow/worktrees/a"} {
		if _, err := os.Stat(filepath.Join(repo, filepath.FromSlash(p))); err != nil {
			t.Errorf("%s should stay: %v", p, err)
		}
	}
	if _, err := os.Stat(filepath.Join(repo, ".sliceflow", "worktrees", "b")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("worktree b left: %v", err)
	}
	if out := gittest.Run(t, repo, "branch", "--list", "slice/b"); out != "" {
		t.Errorf("branch b left: %q", out)
	}
}

// TestCreateCancelledContext: I call Create with a context that is already
// cancelled. The checks still run first (a bad name gives its normal
// error); a good name gives "interrupted" and nothing is made.
func TestCreateCancelledContext(t *testing.T) {
	repo := gittest.NewRepo(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := Create(ctx, CreateOptions{Dir: repo, Name: "My Slice"})
	if err == nil || !strings.HasSuffix(err.Error(), notChanged) {
		t.Errorf("bad name: error = %v, want the check error", err)
	}

	_, err = Create(ctx, CreateOptions{Dir: repo, Name: "x", Now: time.Now()})
	want := "could not create slice x: interrupted. Everything sliceflow made was removed."
	if err == nil || err.Error() != want {
		t.Errorf("error = %v, want %q", err, want)
	}
	if _, err := os.Stat(filepath.Join(repo, ".sliceflow")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf(".sliceflow made: %v", err)
	}
	if out := gittest.Run(t, repo, "branch", "--list", "slice/x"); out != "" {
		t.Errorf("branch made: %q", out)
	}
}

// TestCreatePostCheckoutHookFails: git worktree add fails but leaves the
// worktree. The undo must still remove it, and the branch.
func TestCreatePostCheckoutHookFails(t *testing.T) {
	repo := gittest.NewRepo(t)
	path := filepath.Join(repo, ".git", "hooks", "post-checkout")
	if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 1\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	got := createErr(t, repo, "x", "")
	if !strings.HasSuffix(got, "Everything sliceflow made was removed.") {
		t.Errorf("error = %q, want it to end with the removed sentence", got)
	}
	if _, err := os.Stat(filepath.Join(repo, ".sliceflow", "worktrees", "x")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("worktree folder left: %v", err)
	}
	if out := gittest.Run(t, repo, "branch", "--list", "slice/x"); out != "" {
		t.Errorf("branch left: %q", out)
	}
	if out := gittest.Run(t, repo, "worktree", "list", "--porcelain"); strings.Count(out, "worktree ") != 1 {
		t.Errorf("worktrees left: %q", out)
	}
}

// TestCreateCouldNotRemove: the hook puts a file in the new .sliceflow
// folder, so removing that folder fails and the message names it.
func TestCreateCouldNotRemove(t *testing.T) {
	repo := gittest.NewRepo(t)
	extra := filepath.ToSlash(filepath.Join(repo, ".sliceflow", "extra"))
	writeHook(t, repo, "touch '"+extra+"'\nexit 1")

	want := "could not create slice x: git commit failed (exit code 1). Could not remove: folder .sliceflow."
	if got := createErr(t, repo, "x", ""); got != want {
		t.Errorf("error = %q, want %q", got, want)
	}
	// The rest was removed.
	if out := gittest.Run(t, repo, "branch", "--list", "slice/x"); out != "" {
		t.Errorf("branch left: %q", out)
	}
	if _, err := os.Stat(filepath.Join(repo, ".sliceflow", "worktrees")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("worktrees folder left: %v", err)
	}
}
