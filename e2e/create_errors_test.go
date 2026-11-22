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

// A base branch that does not exist: the hint depends on where the base
// came from, --from, the git setting sliceflow.base, or neither, so main.
// Nothing is changed.
func TestCreateWithNoBase(t *testing.T) {
	tests := []struct {
		about string
		setup func(t *testing.T, repo, home string)
		args  []string
		want  string
	}{
		{"main, by default",
			func(t *testing.T, repo, home string) {
				gitOutput(t, repo, home, "branch", "-M", "master")
			},
			[]string{"Login"},
			"Error: base branch main does not exist. Nothing was changed.\n" +
				"sliceflow starts from main unless you choose another branch. To choose one for this slice, run:\n" +
				"  sliceflow slice create login --from <branch>\n" +
				"To choose one for every slice, run:\n" +
				"  git config sliceflow.base <branch>\n"},
		{"chosen with --from",
			func(t *testing.T, repo, home string) {},
			[]string{"login", "--from", "release"},
			"Error: base branch release does not exist. Nothing was changed.\n" +
				"Check the name. To see your branches, run:\n" +
				"  git branch\n"},
		{"chosen with the git setting",
			func(t *testing.T, repo, home string) {
				gitOutput(t, repo, home, "config", "sliceflow.base", "dev/bert")
			},
			[]string{"login"},
			"Error: base branch dev/bert does not exist. Nothing was changed.\n" +
				"The base branch comes from the git setting sliceflow.base. To see it, run:\n" +
				"  git config sliceflow.base\n"},
	}
	for _, tt := range tests {
		t.Run(tt.about, func(t *testing.T) {
			home := newHome(t)
			repo := newProject(t, home)
			tt.setup(t, repo, home)
			before := gitState(t, repo, home)

			args := append([]string{"slice", "create"}, tt.args...)
			got := runSliceflow(t, repo, home, args...)
			wantError(t, got, 1, tt.want)
			wantSameState(t, repo, home, before)
			wantNoSlice(t, repo, home)
		})
	}
}

// When the user's checkout is not on the base branch, slice create
// refuses, and changes nothing. A checkout on no branch, e.g. after git
// checkout of a commit, is refused too.
func TestCreateOffTheBase(t *testing.T) {
	tests := []struct {
		about string
		setup func(t *testing.T, repo, home string)
		want  string
	}{
		{"on another branch",
			func(t *testing.T, repo, home string) {
				gitOutput(t, repo, home, "config", "sliceflow.base", "dev/bert")
				gitOutput(t, repo, home, "branch", "dev/bert")
				gitOutput(t, repo, home, "switch", "--quiet", "-c", "fix/typo")
			},
			"Error: the base branch is dev/bert, but your checkout is on fix/typo. Nothing was changed.\n" +
				"Switch to dev/bert, then try again.\n"},
		{"on no branch",
			func(t *testing.T, repo, home string) {
				gitOutput(t, repo, home, "checkout", "--quiet", "--detach", "main")
			},
			"Error: the base branch is main, but your checkout is not on a branch. Nothing was changed.\n" +
				"Switch to main, then try again.\n"},
	}
	for _, tt := range tests {
		t.Run(tt.about, func(t *testing.T) {
			home := newHome(t)
			repo := newProject(t, home)
			tt.setup(t, repo, home)
			before := gitState(t, repo, home)

			got := runSliceflow(t, repo, home, "slice", "create", "login")
			wantError(t, got, 1, tt.want)
			wantSameState(t, repo, home, before)
			wantNoFolder(t, sliceFolder(repo, "login"))
		})
	}
}

// When the checkout is not on the base branch and the name is taken too,
// only the wrong branch is reported: it is checked first.
func TestCreateErrorOrder(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)
	commitFolder(t, repo, home)
	gitOutput(t, repo, home, "switch", "--quiet", "-c", "fix/typo")

	got := runSliceflow(t, repo, home, "slice", "create", "login")
	wantError(t, got, 1, "Error: the base branch is main, but your checkout is on fix/typo. Nothing was changed.\n"+
		"Switch to main, then try again.\n")
}

// A slice whose folder is taken, or whose branch exists, is not made
// again, and nothing is changed. The folder is taken when it is in the
// base commit, or is not empty in the user's checkout: also when it was
// never committed, or was deleted with no commit of that. The folder is
// checked before the branch.
func TestCreateWhenTheNameIsTaken(t *testing.T) {
	folderTaken := "Error: slice login already exists (folder docs/slices/login/). Nothing was changed.\n"
	branchTaken := "Error: slice login already exists (branch slice/login). Nothing was changed.\n"
	tests := []struct {
		about string
		setup func(t *testing.T, repo, home string)
		want  string
	}{
		{"a committed folder", commitFolder, folderTaken},
		{"a folder never committed",
			func(t *testing.T, repo, home string) {
				writeText(t, filepath.Join(sliceFolder(repo, "login"), "notes.md"), "mine\n")
			},
			folderTaken},
		{"a folder deleted with no commit",
			func(t *testing.T, repo, home string) {
				commitFolder(t, repo, home)
				if err := os.RemoveAll(sliceFolder(repo, "login")); err != nil {
					t.Fatal(err)
				}
			},
			folderTaken},
		{"the branch",
			func(t *testing.T, repo, home string) {
				gitOutput(t, repo, home, "branch", "slice/login")
			},
			branchTaken},
		{"the folder and the branch",
			func(t *testing.T, repo, home string) {
				commitFolder(t, repo, home)
				gitOutput(t, repo, home, "branch", "slice/login")
			},
			folderTaken},
	}
	for _, tt := range tests {
		t.Run(tt.about, func(t *testing.T) {
			home := newHome(t)
			repo := newProject(t, home)
			tt.setup(t, repo, home)
			before := gitState(t, repo, home)

			got := runSliceflow(t, repo, home, "slice", "create", "Login")
			wantError(t, got, 1, tt.want)
			wantSameState(t, repo, home, before)
		})
	}
}

// An empty folder docs/slices/<name>/ in the checkout is not taken: slice
// create writes the slice into it.
func TestCreateIntoAnEmptyFolder(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)
	if err := os.MkdirAll(sliceFolder(repo, "login"), 0o755); err != nil {
		t.Fatal(err)
	}

	wantCode(t, runSliceflow(t, repo, home, "slice", "create", "login"), 0)

	files := gitOutput(t, repo, home, "show", "--name-only", "--format=", "main")
	if files != "docs/slices/login/.events.jsonl\ndocs/slices/login/spec.md" {
		t.Errorf("the commit has the files:\n%s\nwant the spec and the event log", files)
	}
}

// When the commit fails, slice create removes the files it wrote and
// unstages them, and shows git's message, which has the hook's.
func TestCreateWhenTheCommitFails(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)
	base := gitOutput(t, repo, home, "rev-parse", "main")
	writeHook(t, repo, "pre-commit", "#!/bin/sh\necho 'the hook said no' >&2\nexit 1\n")

	got := runSliceflow(t, repo, home, "slice", "create", "login")
	wantGitText(t, got, "Error: git could not make the commit. The slice was not created. Nothing was changed.\n")
	if !strings.Contains(got.stderr, "\n  the hook said no\n") {
		t.Errorf("stderr is:\n%s\nwant the hook's message in git's text", got.stderr)
	}
	wantNoFolder(t, sliceFolder(repo, "login"))
	wantUntouched(t, repo, home, base)
}

// During a merge, git refuses a commit of only some files. slice create
// shows git's message, and removes what it made. The merge is still
// going on, as before.
func TestCreateDuringAMerge(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)
	gitOutput(t, repo, home, "switch", "--quiet", "-c", "other")
	writeText(t, filepath.Join(repo, "a.txt"), "a\n")
	gitOutput(t, repo, home, "add", "a.txt")
	gitOutput(t, repo, home, "commit", "--quiet", "-m", "add a.txt")
	gitOutput(t, repo, home, "switch", "--quiet", "main")
	gitOutput(t, repo, home, "merge", "--quiet", "--no-ff", "--no-commit", "other")
	before := gitState(t, repo, home)

	got := runSliceflow(t, repo, home, "slice", "create", "login")
	wantGitText(t, got, "Error: git could not make the commit. The slice was not created. Nothing was changed.\n")
	wantSameState(t, repo, home, before)
	wantNoFolder(t, sliceFolder(repo, "login"))
	if _, err := os.Stat(filepath.Join(repo, ".git", "MERGE_HEAD")); err != nil {
		t.Errorf("the merge is no longer going on: %v", err)
	}
}

// When the slice's folder cannot be removed, slice create says what is
// left, and how to remove it. The first line does not say "Nothing was
// changed." Here, the hook makes the folder read-only, so its files
// cannot be deleted. They can still be unstaged, so only the folder is
// left.
func TestCreateWhenSomethingCannotBeRemoved(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("on Windows, a read-only folder does not stop its files from being deleted")
	}
	if os.Geteuid() == 0 {
		t.Skip("root can delete files in a read-only folder")
	}
	home := newHome(t)
	repo := newProject(t, home)
	folder := sliceFolder(repo, "login")
	// Make the folder writable again, so the test's temp folder can be
	// deleted.
	t.Cleanup(func() { os.Chmod(folder, 0o755) })
	writeHook(t, repo, "pre-commit", "#!/bin/sh\nchmod a-w docs/slices/login\necho 'the hook said no' >&2\nexit 1\n")

	got := runSliceflow(t, repo, home, "slice", "create", "login")
	wantGitText(t, got, "Error: git could not make the commit. The slice was not created.\n")
	left := filepath.Join(topFolder(t, repo, home), "docs", "slices", "login")
	wantEnd := "Some of what sliceflow made could not be removed:\n" +
		"  folder    " + left + "\n" +
		"To clean up:\n" +
		"  delete the folder " + left + "\n"
	if !strings.HasSuffix(got.stderr, wantEnd) {
		t.Errorf("stderr is:\n%s\nwant it to end with:\n%s", got.stderr, wantEnd)
	}
	if staged := gitOutput(t, repo, home, "diff", "--cached", "--name-only"); staged != "" {
		t.Errorf("these files are still staged:\n%s\nwant none", staged)
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
	// The hook lets go of git's output first, so git ends at once when
	// it is stopped.
	writeHook(t, repo, "pre-commit", "#!/bin/sh\nexec >/dev/null 2>&1\ntouch \"$HOME/hook-started\"\nsleep 10\n")

	got := interruptWhen(t, repo, home, filepath.Join(home, "hook-started"), "slice", "create", "login")
	wantError(t, got, 1, "Stopped. The slice was not created. Nothing was changed.\n")
	wantNoFolder(t, sliceFolder(repo, "login"))
	wantUntouched(t, repo, home, base)
}

// Ctrl+C after the commit is made, while a post-commit hook runs: the
// slice exists, so slice create removes nothing, says so, and exits with
// code 1. main keeps the commit, and the checkout keeps the files.
func TestCreateStopsAfterTheCommit(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Go cannot send Ctrl+C to one process on Windows")
	}
	home := newHome(t)
	repo := newProject(t, home)
	base := gitOutput(t, repo, home, "rev-parse", "main")
	writeHook(t, repo, "post-commit", "#!/bin/sh\nexec >/dev/null 2>&1\ntouch \"$HOME/hook-started\"\nsleep 10\n")

	got := interruptWhen(t, repo, home, filepath.Join(home, "hook-started"), "slice", "create", "login")
	wantError(t, got, 1, "Stopped after the commit. Slice login was created on branch main.\n")
	if parent := gitOutput(t, repo, home, "rev-parse", "main^"); parent != base {
		t.Errorf("the commit before main is %s, want the old main %s", parent, base)
	}
	if _, err := os.Stat(filepath.Join(sliceFolder(repo, "login"), "spec.md")); err != nil {
		t.Errorf("the spec is gone from the checkout: %v", err)
	}
	if status := gitOutput(t, repo, home, "status", "--porcelain"); status != "" {
		t.Errorf("the project has changes:\n%s\nwant none", status)
	}
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

// gitState returns what a command that changes nothing must leave as it
// was in the checkout at repo: what is checked out, where every branch
// and tag points, and every staged, changed, and new file.
func gitState(t *testing.T, repo, home string) string {
	t.Helper()
	return "HEAD: " + gitOutput(t, repo, home, "rev-parse", "--abbrev-ref", "HEAD") +
		" at " + gitOutput(t, repo, home, "rev-parse", "HEAD") + "\n" +
		gitOutput(t, repo, home, "for-each-ref", "--format=%(refname) %(objectname)") + "\n" +
		gitOutput(t, repo, home, "status", "--porcelain", "--untracked-files=all")
}

// wantSameState fails the test unless gitState is still before.
func wantSameState(t *testing.T, repo, home, before string) {
	t.Helper()
	if after := gitState(t, repo, home); after != before {
		t.Errorf("the project changed from:\n%s\nto:\n%s", before, after)
	}
}

// sliceFolder returns the folder of the slice name in the checkout at
// repo.
func sliceFolder(repo, name string) string {
	return filepath.Join(repo, "docs", "slices", name)
}

// commitFolder commits a folder for the slice login on the branch the
// project has checked out, as if the slice was made before.
func commitFolder(t *testing.T, repo, home string) {
	t.Helper()
	writeText(t, filepath.Join(sliceFolder(repo, "login"), "spec.md"), "done\n")
	gitOutput(t, repo, home, "add", ".")
	gitOutput(t, repo, home, "commit", "--quiet", "-m", "a finished slice")
}

// writeHook makes text the git hook name, e.g. pre-commit, of the
// project in repo.
func writeHook(t *testing.T, repo, name, text string) {
	t.Helper()
	hook := filepath.Join(repo, ".git", "hooks", name)
	writeText(t, hook, text)
	if err := os.Chmod(hook, 0o755); err != nil {
		t.Fatal(err)
	}
}

// interruptWhen runs sliceflow with args in repo, waits for the file
// started to appear, then sends it Ctrl+C, and returns what it printed
// and its exit code.
func interruptWhen(t *testing.T, repo, home, started string, args ...string) result {
	t.Helper()
	cmd := exec.Command(program, args...)
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

	code := 0
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		code = exitErr.ExitCode()
	} else if err != nil {
		t.Fatalf("sliceflow did not end well: %v", err)
	}
	return result{stdout: stdout.String(), stderr: stderr.String(), code: code}
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
