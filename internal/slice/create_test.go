package slice_test

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/BertPratya/sliceflow/internal/eventlog"
	"github.com/BertPratya/sliceflow/internal/git"
	"github.com/BertPratya/sliceflow/internal/gittest"
	"github.com/BertPratya/sliceflow/internal/slice"
	"github.com/BertPratya/sliceflow/templates"
)

// lockLine is the line that Create adds to info/exclude (R2, SD31).
const lockLine = "*.events.jsonl.lock"

// mustCreate calls Create in dir with in, and fails the test on an
// error.
func mustCreate(t *testing.T, dir string, in slice.CreateInput) slice.Created {
	t.Helper()
	got, err := slice.Create(context.Background(), dir, in)
	if err != nil {
		t.Fatalf("Create returned the error:\n%v\nwant none", err)
	}
	return got
}

// createFails calls Create in dir with in, and fails the test when it
// returns no error.
func createFails(t *testing.T, dir string, in slice.CreateInput) error {
	t.Helper()
	_, err := slice.Create(context.Background(), dir, in)
	if err == nil {
		t.Fatalf("Create returned no error, want one")
	}
	return err
}

// removeIdentity takes the name and the email out of the test's global
// git config, and sets user.useConfigOnly, so git cannot make a commit
// on any machine (TD8). Call it after gittest.New, which needs a commit.
func removeIdentity(t *testing.T, r *gittest.Repo) {
	t.Helper()
	r.Git("config", "--global", "--unset", "user.name")
	r.Git("config", "--global", "--unset", "user.email")
	gittest.SetGlobalConfig(t, "user.useConfigOnly", "true")
	t.Setenv("EMAIL", "")
	if err := os.Unsetenv("EMAIL"); err != nil {
		t.Fatalf("unset EMAIL: %v", err)
	}
}

// excludePath returns the path of the repo's info/exclude file.
func excludePath(r *gittest.Repo) string {
	return filepath.Join(r.Dir, ".git", "info", "exclude")
}

// readText returns the text of the file at path, and fails the test on
// an error.
func readText(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

// countLines returns how many lines of text are exactly line.
func countLines(text, line string) int {
	n := 0
	for _, l := range strings.Split(text, "\n") {
		if strings.TrimSuffix(l, "\r") == line {
			n++
		}
	}
	return n
}

// exists reports whether something is at path, and fails the test when
// it cannot tell.
func exists(t *testing.T, path string) bool {
	t.Helper()
	_, err := os.Lstat(path)
	if err == nil {
		return true
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false
	}
	t.Fatalf("look at %s: %v", path, err)
	return false
}

// checkGone fails the test when the branch slice/<name> or the worktree
// at path is still there: in git, or on disk.
func checkGone(t *testing.T, r *gittest.Repo, name, path string) {
	t.Helper()
	ctx := context.Background()
	branch, err := git.BranchExists(ctx, r.Dir, "slice/"+name)
	if err != nil {
		t.Fatalf("BranchExists: %v", err)
	}
	if branch {
		t.Errorf("the branch slice/%s is still there", name)
	}
	registered, err := git.WorktreeRegistered(ctx, r.Dir, path)
	if err != nil {
		t.Fatalf("WorktreeRegistered: %v", err)
	}
	if registered {
		t.Errorf("git still has a worktree at %s", path)
	}
	if exists(t, path) {
		t.Errorf("the worktree folder %s is still there", path)
	}
}

// R1, R2, R3, R4, R7: Create makes the branch, the worktree, and one
// commit with the two files, and leaves the user's checkout as it was.
func TestCreateMakesSlice(t *testing.T) {
	r := gittest.New(t)
	base := r.Git("rev-parse", "main")
	project := findProject(t, r.Dir)
	path := worktreePath(t, project)

	got := mustCreate(t, r.Dir, slice.CreateInput{GivenName: "Login"})

	if got.Name != "login" || !got.Lowered || got.Base != "main" || got.BaseCommit != base {
		t.Errorf("Create = %+v, want Name login, Lowered true, Base main, BaseCommit %s", got, base)
	}
	if got.Worktree != path {
		t.Errorf("Worktree = %q, want %q", got.Worktree, path)
	}
	wantSpec := filepath.Join(path, "docs", "slices", "login", "spec.md")
	if got.SpecPath != wantSpec {
		t.Errorf("SpecPath = %q, want %q", got.SpecPath, wantSpec)
	}
	if len(got.IgnoredFiles) != 0 {
		t.Errorf("IgnoredFiles = %q, want none", got.IgnoredFiles)
	}

	// One commit on top of the base, with exactly the two files.
	if parent := r.Git("rev-parse", "slice/login^"); parent != base {
		t.Errorf("the parent of slice/login is %s, want the base %s", parent, base)
	}
	if count := r.Git("rev-list", "--count", "main..slice/login"); count != "1" {
		t.Errorf("slice/login has %s commits on top of main, want 1", count)
	}
	changed := r.Git("diff-tree", "--no-commit-id", "--name-only", "-r", "slice/login")
	wantFiles := "docs/slices/login/.events.jsonl\ndocs/slices/login/spec.md"
	if changed != wantFiles {
		t.Errorf("the commit holds:\n%s\nwant:\n%s", changed, wantFiles)
	}
	if msg := r.Git("log", "-1", "--format=%s", "slice/login"); msg != "login: add the spec and the event log" {
		t.Errorf("the commit message is %q", msg)
	}

	// spec.md is the template as it is, on disk and in the commit.
	if spec := readText(t, got.SpecPath); spec != templates.Spec {
		t.Errorf("spec.md on disk differs from templates.Spec:\n%s", spec)
	}
	blob := r.Git("rev-parse", "slice/login:docs/slices/login/spec.md")
	if want := r.Git("hash-object", got.SpecPath); blob != want {
		t.Errorf("spec.md in the commit differs from the file on disk")
	}

	// The user's checkout is as it was.
	if branch := r.Git("symbolic-ref", "--short", "HEAD"); branch != "main" {
		t.Errorf("the checked-out branch is %q, want main", branch)
	}
	if head := r.Git("rev-parse", "HEAD"); head != base {
		t.Errorf("HEAD is %s, want %s", head, base)
	}
	if status := r.Git("status", "--porcelain"); status != "" {
		t.Errorf("the user's checkout has changes:\n%s", status)
	}
}

// R2: the lock file of the log is ignored through info/exclude, whose
// own lines stay as they are.
func TestCreateExcludesLock(t *testing.T) {
	t.Run("line added once", func(t *testing.T) {
		r := gittest.New(t)
		got := mustCreate(t, r.Dir, slice.CreateInput{GivenName: "login"})
		if status := gittest.Git(t, got.Worktree, "status", "--porcelain"); status != "" {
			t.Errorf("git status in the worktree:\n%s\nwant nothing", status)
		}
		if n := countLines(readText(t, excludePath(r)), lockLine); n != 1 {
			t.Errorf("info/exclude has the line %d times, want 1", n)
		}
		mustCreate(t, r.Dir, slice.CreateInput{GivenName: "signup"})
		if n := countLines(readText(t, excludePath(r)), lockLine); n != 1 {
			t.Errorf("after a second slice, info/exclude has the line %d times, want 1", n)
		}
	})

	tests := []struct {
		name   string
		before string
		want   string
	}{
		{"no line break at the end", "# mine\n*.tmp", "# mine\n*.tmp\n" + lockLine + "\n"},
		{"Windows line ends", lockLine + "\r\n", lockLine + "\r\n"},
		{"empty", "", lockLine + "\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gittest.New(t)
			if err := os.WriteFile(excludePath(r), []byte(tt.before), 0o644); err != nil {
				t.Fatal(err)
			}
			mustCreate(t, r.Dir, slice.CreateInput{GivenName: "login"})
			if got := readText(t, excludePath(r)); got != tt.want {
				t.Errorf("info/exclude = %q, want %q", got, tt.want)
			}
		})
	}

	t.Run("file and folder missing", func(t *testing.T) {
		r := gittest.New(t)
		if err := os.RemoveAll(filepath.Dir(excludePath(r))); err != nil {
			t.Fatal(err)
		}
		mustCreate(t, r.Dir, slice.CreateInput{GivenName: "login"})
		if got := readText(t, excludePath(r)); got != lockLine+"\n" {
			t.Errorf("info/exclude = %q, want %q", got, lockLine+"\n")
		}
	})
}

// R27, R28: the first event holds the name, the title, the
// description, the base, and the full base commit.
func TestCreateFirstEventData(t *testing.T) {
	tests := []struct {
		name  string
		in    slice.CreateInput
		title string
		desc  string
	}{
		{"no title", slice.CreateInput{GivenName: "login"}, "login", ""},
		{"title and description", slice.CreateInput{GivenName: "login", Title: "Card payment", Description: "Pay by card."}, "Card payment", "Pay by card."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gittest.New(t)
			base := r.Git("rev-parse", "main")
			got := mustCreate(t, r.Dir, tt.in)

			logPath := filepath.Join(filepath.Dir(got.SpecPath), ".events.jsonl")
			events, bad, err := eventlog.Read(logPath)
			if err != nil {
				t.Fatalf("Read: %v", err)
			}
			if len(events) != 1 || len(bad) != 0 {
				t.Fatalf("the log has %d events and %d bad lines, want 1 and 0", len(events), len(bad))
			}
			if events[0].Event != eventlog.SliceCreatedEvent || events[0].Seq != 1 {
				t.Errorf("first event = %s seq %d, want %s seq 1", events[0].Event, events[0].Seq, eventlog.SliceCreatedEvent)
			}
			var data eventlog.SliceCreated
			if err := json.Unmarshal(events[0].Data, &data); err != nil {
				t.Fatalf("read the data: %v", err)
			}
			want := eventlog.SliceCreated{Name: "login", Title: tt.title, Description: tt.desc, Base: "main", BaseCommit: base}
			if data != want {
				t.Errorf("data = %+v, want %+v", data, want)
			}
			if len(base) != 40 {
				t.Errorf("the base commit %q is not a full hash", base)
			}
		})
	}
}

// R5: a file that a .gitignore rule matches is still in the commit, and
// is named in IgnoredFiles.
func TestCreateIgnoredFiles(t *testing.T) {
	r := gittest.New(t)
	r.Commit(".gitignore", "*.jsonl\n")

	got := mustCreate(t, r.Dir, slice.CreateInput{GivenName: "login"})

	if !slices.Equal(got.IgnoredFiles, []string{".events.jsonl"}) {
		t.Errorf("IgnoredFiles = %q, want [.events.jsonl]", got.IgnoredFiles)
	}
	files := r.Git("ls-tree", "-r", "--name-only", "slice/login", "--", "docs/slices/login")
	want := "docs/slices/login/.events.jsonl\ndocs/slices/login/spec.md"
	if files != want {
		t.Errorf("the commit holds:\n%s\nwant:\n%s", files, want)
	}
}

// R40: when two problems apply, the first in R40's order is reported.
func TestCreateCheckOrder(t *testing.T) {
	tests := []struct {
		name  string
		setup func(r *gittest.Repo)
		in    slice.CreateInput
		want  string
	}{
		{
			"control character before a bad name",
			func(r *gittest.Repo) {},
			slice.CreateInput{GivenName: "bad.name", Title: "a\tb"},
			"Error: --title has a control character. Nothing was changed.\n" +
				`Received: "a\tb"` + "\n" +
				"Write it as plain text on one line, with no control characters.",
		},
		{
			"missing base before a taken branch",
			func(r *gittest.Repo) { r.Branch("slice/login") },
			slice.CreateInput{GivenName: "login", From: "nope"},
			"Error: base branch nope does not exist. Nothing was changed.\n" +
				"Check the name. To see your branches, run:\n" +
				"  git branch",
		},
		{
			"folder in the base before a taken branch",
			func(r *gittest.Repo) {
				r.Commit("docs/slices/login/spec.md", "x\n")
				r.Branch("slice/login")
			},
			slice.CreateInput{GivenName: "login"},
			"Error: slice login already exists (folder docs/slices/login/ in main). Nothing was changed.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gittest.New(t)
			tt.setup(r)
			err := createFails(t, r.Dir, tt.in)
			if err.Error() != tt.want {
				t.Errorf("error:\n%s\nwant:\n%s", err, tt.want)
			}
		})
	}
}

// R36, R37: when the commit fails, the branch, the worktree, and the
// folders made for it are removed, and a folder that was there before
// stays.
func TestCreateUndoOnCommitFail(t *testing.T) {
	r := gittest.New(t)
	removeIdentity(t, r)
	project := findProject(t, r.Dir)
	path := worktreePath(t, project)
	before := filepath.Join(r.Home, ".sliceflow")
	if err := os.Mkdir(before, 0o755); err != nil {
		t.Fatal(err)
	}

	err := createFails(t, r.Dir, slice.CreateInput{GivenName: "login"})

	wantStart := "Error: git could not make the commit. The slice was not created. Nothing was changed.\ngit said:\n  "
	if !strings.HasPrefix(err.Error(), wantStart) {
		t.Errorf("error:\n%s\nwant it to start with:\n%s", err, wantStart)
	}
	checkGone(t, r, "login", path)
	if made := filepath.Join(before, "worktrees"); exists(t, made) {
		t.Errorf("the folder %s, made for the worktree, is still there", made)
	}
	if !exists(t, before) {
		t.Errorf("the folder %s, there before, was removed", before)
	}
}

// R36: a worktree folder that was there before, and empty, is there
// and empty afterwards.
func TestCreateKeepsEmptyFolder(t *testing.T) {
	r := gittest.New(t)
	removeIdentity(t, r)
	path := worktreePath(t, findProject(t, r.Dir))
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}

	err := createFails(t, r.Dir, slice.CreateInput{GivenName: "login"})

	if !strings.HasPrefix(err.Error(), "Error: git could not make the commit.") {
		t.Errorf("error:\n%s\nwant R37's make the commit", err)
	}
	entries, readErr := os.ReadDir(path)
	if readErr != nil {
		t.Fatalf("the folder %s is not there: %v", path, readErr)
	}
	if len(entries) != 0 {
		t.Errorf("the folder %s has %d entries, want none", path, len(entries))
	}
	registered, regErr := git.WorktreeRegistered(context.Background(), r.Dir, path)
	if regErr != nil {
		t.Fatalf("WorktreeRegistered: %v", regErr)
	}
	if registered {
		t.Errorf("git still has a worktree at %s", path)
	}
	if exists, _ := git.BranchExists(context.Background(), r.Dir, "slice/login"); exists {
		t.Errorf("the branch slice/login is still there")
	}
}

// R36: a worktree that git had registered before, at the path, stays
// registered.
func TestCreateKeepsRegisteredWorktree(t *testing.T) {
	r := gittest.New(t)
	path := worktreePath(t, findProject(t, r.Dir))
	r.Git("worktree", "add", "--quiet", "-b", "other", path)
	if err := os.RemoveAll(path); err != nil {
		t.Fatal(err)
	}

	err := createFails(t, r.Dir, slice.CreateInput{GivenName: "login"})

	if !strings.HasPrefix(err.Error(), "Error: git could not create the worktree. The slice was not created. Nothing was changed.\ngit said:\n") {
		t.Errorf("error:\n%s\nwant R37's create the worktree", err)
	}
	registered, regErr := git.WorktreeRegistered(context.Background(), r.Dir, path)
	if regErr != nil {
		t.Fatalf("WorktreeRegistered: %v", regErr)
	}
	if !registered {
		t.Errorf("the worktree registered at %s is gone", path)
	}
	if exists, _ := git.BranchExists(context.Background(), r.Dir, "slice/login"); exists {
		t.Errorf("the branch slice/login is still there")
	}
}

// R38: when the slice's folder cannot be made in the worktree, the
// error is R38's, and the branch and the worktree are removed.
func TestCreateWriteFails(t *testing.T) {
	r := gittest.New(t)
	r.Commit("docs", "a file, not a folder\n")
	path := worktreePath(t, findProject(t, r.Dir))

	err := createFails(t, r.Dir, slice.CreateInput{GivenName: "login"})

	folder := filepath.Join(path, "docs", "slices", "login")
	wantStart := "Error: could not write " + folder + ". The slice was not created. Nothing was changed.\nThe system said:\n  "
	if !strings.HasPrefix(err.Error(), wantStart) {
		t.Errorf("error:\n%s\nwant it to start with:\n%s", err, wantStart)
	}
	checkGone(t, r, "login", path)
}
