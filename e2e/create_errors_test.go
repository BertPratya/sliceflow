package e2e

import (
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// A title or a description with a control character is refused with
// exit code 2, and nothing is made. When both have one, the title is
// reported.
func TestCreateRefusesAControlCharacter(t *testing.T) {
	tests := []struct {
		about   string
		options []string
		want    string
	}{
		{"a line break in the title", []string{"--title", "Pay with a card\n"},
			"Error: --title has a control character. Nothing was changed.\n" +
				"Received: \"Pay with a card\\n\"\n" +
				"Write it as plain text on one line, with no control characters.\n"},
		{"a tab in the description", []string{"--description", "a\tb"},
			"Error: --description has a control character. Nothing was changed.\n" +
				"Received: \"a\\tb\"\n" +
				"Write it as plain text on one line, with no control characters.\n"},
		{"both", []string{"--title", "a\x1bb", "--description", "c\nd"},
			"Error: --title has a control character. Nothing was changed.\n" +
				"Received: \"a\\u001bb\"\n" +
				"Write it as plain text on one line, with no control characters.\n"},
	}
	for _, tt := range tests {
		t.Run(tt.about, func(t *testing.T) {
			home := newHome(t)
			repo := newProject(t, home)

			args := append([]string{"slice", "create", "login"}, tt.options...)
			got := runSliceflow(t, repo, home, args...)
			wantError(t, got, 2, tt.want)
			wantNoSlice(t, repo, home)
		})
	}
}

// A name that breaks the rules is refused with exit code 2. The message
// shows the name as it was given.
func TestCreateRefusesAnInvalidName(t *testing.T) {
	tests := []struct {
		about string
		name  string
		hint  string
	}{
		{"a dot", "card.payment", "Allowed characters: letters, digits, hyphens, underscores. Start with a letter or a digit."},
		{"too long", strings.Repeat("a", 41), "A name has at most 40 characters."},
		{"reserved by Windows", "CON", "Windows reserves this name. Choose another."},
	}
	for _, tt := range tests {
		t.Run(tt.about, func(t *testing.T) {
			home := newHome(t)
			repo := newProject(t, home)

			got := runSliceflow(t, repo, home, "slice", "create", tt.name)
			wantError(t, got, 2, "Error: invalid slice name \""+tt.name+"\". Nothing was changed.\n"+tt.hint+"\n")
			wantNoSlice(t, repo, home)
		})
	}
}

// When git cannot be found, slice create says so.
func TestCreateWhenGitIsNotFound(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)

	got := runSliceflowWith(t, repo, home, []string{"PATH=" + t.TempDir()}, "slice", "create", "login")
	wantError(t, got, 1, "Error: git was not found. Nothing was changed.\nInstall git, then try again.\n")
}

// Outside a git repository, slice create shows git's own message, after
// a line of its own.
func TestCreateOutsideARepository(t *testing.T) {
	home := newHome(t)
	dir := t.TempDir()
	// Stop git from finding a repository in the folders above dir.
	env := []string{"GIT_CEILING_DIRECTORIES=" + filepath.Dir(dir)}

	got := runSliceflowWith(t, dir, home, env, "slice", "create", "login")
	wantGitText(t, got, "Error: git could not find the project. Nothing was changed.\n")
}

// A repository with no commits has nothing to start a slice from.
func TestCreateWithNoCommits(t *testing.T) {
	home := newHome(t)
	repo := filepath.Join(t.TempDir(), "shop")
	if err := os.Mkdir(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	gitOutput(t, repo, home, "init", "--quiet")

	got := runSliceflow(t, repo, home, "slice", "create", "login")
	wantError(t, got, 1, "Error: this repository has no commits yet. Nothing was changed.\nMake a first commit, then try again.\n")
}

// A base branch that does not exist: the hint depends on whether the
// user chose it.
func TestCreateWithNoBase(t *testing.T) {
	t.Run("main, by default", func(t *testing.T) {
		home := newHome(t)
		repo := newProject(t, home)
		gitOutput(t, repo, home, "branch", "-M", "master")

		got := runSliceflow(t, repo, home, "slice", "create", "Login")
		wantError(t, got, 1, "Error: base branch main does not exist. Nothing was changed.\n"+
			"sliceflow starts from main unless you choose another branch. To choose one, run:\n"+
			"  sliceflow slice create login --from <branch>\n")
		wantNoSlice(t, repo, home)
	})
	t.Run("chosen with --from", func(t *testing.T) {
		home := newHome(t)
		repo := newProject(t, home)

		got := runSliceflow(t, repo, home, "slice", "create", "login", "--from", "release")
		wantError(t, got, 1, "Error: base branch release does not exist. Nothing was changed.\n"+
			"Check the name. To see your branches, run:\n"+
			"  git branch\n")
		wantNoSlice(t, repo, home)
	})
}

// A slice whose folder is in the base, or whose branch exists, is not
// made again. The folder is checked first.
func TestCreateWhenTheNameIsTaken(t *testing.T) {
	folderTaken := "Error: slice login already exists (folder docs/slices/login/ in main). Nothing was changed.\n"
	branchTaken := "Error: slice login already exists (branch slice/login). Nothing was changed.\n"
	tests := []struct {
		about  string
		folder bool
		branch bool
		want   string
	}{
		{"the folder", true, false, folderTaken},
		{"the branch", false, true, branchTaken},
		{"both", true, true, folderTaken},
	}
	for _, tt := range tests {
		t.Run(tt.about, func(t *testing.T) {
			home := newHome(t)
			repo := newProject(t, home)
			if tt.folder {
				writeText(t, filepath.Join(repo, "docs", "slices", "login", "spec.md"), "done\n")
				gitOutput(t, repo, home, "add", ".")
				gitOutput(t, repo, home, "commit", "--quiet", "-m", "a finished slice")
			}
			if tt.branch {
				gitOutput(t, repo, home, "branch", "slice/login")
			}
			before := gitOutput(t, repo, home, "branch", "--list")

			got := runSliceflow(t, repo, home, "slice", "create", "Login")
			wantError(t, got, 1, tt.want)
			if after := gitOutput(t, repo, home, "branch", "--list"); after != before {
				t.Errorf("the branches changed from:\n%s\nto:\n%s", before, after)
			}
		})
	}
}

// When git cannot make the worktree, because its folder is not empty,
// slice create removes the branch it made, and shows git's message.
func TestCreateWhenTheWorktreeFolderIsNotEmpty(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)
	worktree := defaultWorktree(home, "login")
	writeText(t, filepath.Join(worktree, "notes.txt"), "mine\n")

	got := runSliceflow(t, repo, home, "slice", "create", "login")
	wantGitText(t, got, "Error: git could not create the worktree. The slice was not created. Nothing was changed.\n")
	wantNoSlice(t, repo, home)
	if _, err := os.Stat(filepath.Join(worktree, "notes.txt")); err != nil {
		t.Errorf("the user's file in the folder is gone: %v", err)
	}
}

// When the commit fails, slice create removes the worktree and the
// branch, and shows git's message, which has the hook's.
func TestCreateWhenTheCommitFails(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)
	base := gitOutput(t, repo, home, "rev-parse", "main")
	writeHook(t, repo, "#!/bin/sh\necho 'the hook said no' >&2\nexit 1\n")

	got := runSliceflow(t, repo, home, "slice", "create", "login")
	wantGitText(t, got, "Error: git could not make the commit. The slice was not created. Nothing was changed.\n")
	if !strings.Contains(got.stderr, "\n  the hook said no\n") {
		t.Errorf("stderr is:\n%s\nwant the hook's message in git's text", got.stderr)
	}
	wantNoSlice(t, repo, home)
	wantNoFolder(t, defaultWorktree(home, "login"))
	wantUntouched(t, repo, home, base)
}

// When the worktree cannot be removed, slice create says what is left,
// and how to remove it. The first line does not say "Nothing was
// changed." Here, the hook locks the worktree, which makes git refuse to
// remove it, and so to delete its branch.
func TestCreateWhenSomethingCannotBeRemoved(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)
	writeHook(t, repo, "#!/bin/sh\ngit worktree lock .\necho 'the hook said no' >&2\nexit 1\n")
	worktree := defaultWorktree(home, "login")

	got := runSliceflow(t, repo, home, "slice", "create", "login")
	wantGitText(t, got, "Error: git could not make the commit. The slice was not created.\n")
	wantEnd := "Some of what sliceflow made could not be removed:\n" +
		"  worktree  " + worktree + "\n" +
		"  branch    slice/login\n" +
		"To clean up, run:\n" +
		"  git worktree remove --force \"" + worktree + "\"\n" +
		"  git branch -D \"slice/login\"\n"
	if !strings.HasSuffix(got.stderr, wantEnd) {
		t.Errorf("stderr is:\n%s\nwant it to end with:\n%s", got.stderr, wantEnd)
	}
}

// Ctrl+C while the commit waits on a hook: slice create removes what it
// made, and says it stopped.
func TestCreateStopsOnCtrlC(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Go cannot send Ctrl+C to one process on Windows")
	}
	home := newHome(t)
	repo := newProject(t, home)
	base := gitOutput(t, repo, home, "rev-parse", "main")
	started := filepath.Join(home, "hook-started")
	// The hook lets go of git's output first, so git ends at once when
	// it is stopped.
	writeHook(t, repo, "#!/bin/sh\nexec >/dev/null 2>&1\ntouch \"$HOME/hook-started\"\nsleep 10\n")

	cmd := exec.Command(program, "slice", "create", "login")
	cmd.Dir = repo
	cmd.Env = testEnv(home)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	waitForFile(t, started)
	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}
	err := cmd.Wait()

	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
		t.Errorf("sliceflow ended with %v, want exit code 1", err)
	}
	if want := "Stopped. The slice was not created. Nothing was changed.\n"; stderr.String() != want {
		t.Errorf("stderr is %q, want %q", stderr.String(), want)
	}
	if stdout.String() != "" {
		t.Errorf("stdout is %q, want nothing", stdout.String())
	}
	wantNoSlice(t, repo, home)
	wantNoFolder(t, defaultWorktree(home, "login"))
	wantUntouched(t, repo, home, base)
}

// wantError fails the test unless sliceflow exited with code, printed
// nothing to stdout, and printed exactly want to stderr.
func wantError(t *testing.T, got result, code int, want string) {
	t.Helper()
	if got.code != code {
		t.Errorf("exit code is %d, want %d", got.code, code)
	}
	if got.stdout != "" {
		t.Errorf("stdout is %q, want nothing", got.stdout)
	}
	if got.stderr != want {
		t.Errorf("stderr is:\n%s\nwant:\n%s", got.stderr, want)
	}
}

// wantGitText fails the test unless sliceflow exited with code 1, printed
// nothing to stdout, and printed first, then git's own text after "git
// said:", each line indented by two spaces or empty. Git's words are not
// compared: they differ between git versions and languages.
func wantGitText(t *testing.T, got result, first string) {
	t.Helper()
	if got.code != 1 {
		t.Errorf("exit code is %d, want 1", got.code)
	}
	if got.stdout != "" {
		t.Errorf("stdout is %q, want nothing", got.stdout)
	}
	rest, found := strings.CutPrefix(got.stderr, first+"git said:\n  ")
	if !found {
		t.Fatalf("stderr is:\n%s\nwant it to start with:\n%sgit said:\n  <git's text>", got.stderr, first)
	}
	text, _, _ := strings.Cut(rest, "Some of what sliceflow made")
	for _, line := range strings.Split(strings.TrimSuffix(text, "\n"), "\n")[1:] {
		if line != "" && !strings.HasPrefix(line, "  ") {
			t.Errorf("git's line %q is not indented by two spaces", line)
		}
	}
}

// wantNoSlice fails the test unless there is no branch slice/login, and
// no worktree but the project's own.
func wantNoSlice(t *testing.T, repo, home string) {
	t.Helper()
	if branches := gitOutput(t, repo, home, "branch", "--list", "slice/*"); branches != "" {
		t.Errorf("the slice's branch is there:\n%s", branches)
	}
	list := gitOutput(t, repo, home, "worktree", "list", "--porcelain")
	if n := strings.Count(list, "worktree "); n != 1 {
		t.Errorf("git lists %d worktrees, want 1, the project's own:\n%s", n, list)
	}
}

// wantNoFolder fails the test unless path does not exist.
func wantNoFolder(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("the folder %s is still there, or cannot be checked: %v", path, err)
	}
}

// wantUntouched fails the test unless the project still has main checked
// out, at base, with no changes.
func wantUntouched(t *testing.T, repo, home, base string) {
	t.Helper()
	if branch := gitOutput(t, repo, home, "branch", "--show-current"); branch != "main" {
		t.Errorf("the project has %q checked out, want main", branch)
	}
	if now := gitOutput(t, repo, home, "rev-parse", "main"); now != base {
		t.Errorf("main moved from %s to %s", base, now)
	}
	if status := gitOutput(t, repo, home, "status", "--porcelain"); status != "" {
		t.Errorf("the project has changes:\n%s", status)
	}
}

// writeHook makes text the pre-commit hook of the project in repo.
func writeHook(t *testing.T, repo, text string) {
	t.Helper()
	hook := filepath.Join(repo, ".git", "hooks", "pre-commit")
	writeText(t, hook, text)
	if err := os.Chmod(hook, 0o755); err != nil {
		t.Fatal(err)
	}
}

// waitForFile waits up to 10 seconds for path to exist, and fails the
// test now if it does not.
func waitForFile(t *testing.T, path string) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("%s did not appear within 10 seconds", path)
}
