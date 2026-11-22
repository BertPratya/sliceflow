package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/BertPratya/sliceflow/templates"
)

// slice create writes the spec and the event log in the user's own
// checkout, on the base branch, and commits only them there, with one
// commit. It makes no branch and no worktree, and prints where the spec
// is.
func TestCreateASlice(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)
	base := gitOutput(t, repo, home, "rev-parse", "main")

	got := runSliceflow(t, repo, home, "slice", "create", "login")

	spec := filepath.Join(topFolder(t, repo, home), "docs", "slices", "login", "spec.md")
	wantStdout := "Created slice login on branch main.\n" +
		"Next: write the spec in " + spec + "\n"
	wantCode(t, got, 0)
	if got.stderr != "" {
		t.Errorf("stderr is %q, want nothing", got.stderr)
	}
	if got.stdout != wantStdout {
		t.Errorf("stdout is:\n%s\nwant:\n%s", got.stdout, wantStdout)
	}

	// main has one new commit, on top of where it was.
	if parent := gitOutput(t, repo, home, "rev-parse", "main^"); parent != base {
		t.Errorf("the commit before main is %s, want the old main %s", parent, base)
	}

	// That commit has exactly the two files, and the expected message.
	files := gitOutput(t, repo, home, "show", "--name-only", "--format=", "main")
	wantFiles := "docs/slices/login/.events.jsonl\ndocs/slices/login/spec.md"
	if files != wantFiles {
		t.Errorf("the commit has the files:\n%s\nwant:\n%s", files, wantFiles)
	}
	message := gitOutput(t, repo, home, "log", "-1", "--format=%s", "main")
	if message != "login: add the spec and the event log" {
		t.Errorf("the commit message is %q, want %q", message, "login: add the spec and the event log")
	}

	// No branch and no worktree were made.
	wantNoSlice(t, repo, home)
	wantNoFolder(t, defaultWorktree(home, "login"))

	// The user's checkout is still on main, with no changes: the commit
	// holds the files, and no lock file is left in the checkout.
	if branch := gitOutput(t, repo, home, "branch", "--show-current"); branch != "main" {
		t.Errorf("the project has %q checked out, want main", branch)
	}
	if status := gitOutput(t, repo, home, "status", "--porcelain"); status != "" {
		t.Errorf("the project has changes:\n%s\nwant none", status)
	}
}

// The spec is the built-in template with a header of two lines: the
// slice's name, and the day it was created, in UTC. It has no title line.
func TestCreateWritesTheSpec(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)

	before := time.Now().UTC().Format("2006-01-02")
	got := runSliceflow(t, repo, home, "slice", "create", "login", "--title", "Log in")
	after := time.Now().UTC().Format("2006-01-02")
	wantCode(t, got, 0)

	data, err := os.ReadFile(filepath.Join(repo, "docs", "slices", "login", "spec.md"))
	if err != nil {
		t.Fatalf("could not read the spec: %v", err)
	}
	spec := string(data)

	// The test runs over midnight in UTC only rarely; then either day is
	// right.
	wantBefore := "---\nslice: login\ncreated: " + before + "\n---\n" + templateBody(t)
	wantAfter := "---\nslice: login\ncreated: " + after + "\n---\n" + templateBody(t)
	if spec != wantBefore && spec != wantAfter {
		t.Errorf("the spec starts with:\n%s\nwant:\n%s", firstLines(spec, 6), firstLines(wantBefore, 6))
	}
}

// The event log has one line: the event slice.created, with its keys in
// a fixed order. The title is --title, or the name when --title is left
// out or empty. The description is --description, or "". <, >, and & are
// written as they are.
func TestCreateWritesTheFirstEvent(t *testing.T) {
	tests := []struct {
		about           string
		options         []string
		wantTitle       string
		wantDescription string
	}{
		{"no options", nil, "login", ""},
		{"a title and a description", []string{"--title", "Log in", "--description", "Lets a user log in."},
			"Log in", "Lets a user log in."},
		{"an empty title", []string{"--title", ""}, "login", ""},
		{"<, >, and &", []string{"--title", "<b> & co", "--description", "a < b > c & d"},
			"<b> & co", "a < b > c & d"},
	}
	for _, tt := range tests {
		t.Run(tt.about, func(t *testing.T) {
			home := newHome(t)
			repo := newProject(t, home)
			base := gitOutput(t, repo, home, "rev-parse", "main")

			args := append([]string{"slice", "create", "login"}, tt.options...)
			wantCode(t, runSliceflow(t, repo, home, args...), 0)

			data, err := os.ReadFile(filepath.Join(repo, "docs", "slices", "login", ".events.jsonl"))
			if err != nil {
				t.Fatalf("could not read the event log: %v", err)
			}
			log := string(data)

			var event struct{ At string }
			if err := json.Unmarshal(data, &event); err != nil {
				t.Fatalf("the event log is not one JSON line: %v\n%s", err, log)
			}
			if _, err := time.Parse("2006-01-02T15:04:05Z", event.At); err != nil {
				t.Errorf("the time %q is not UTC written like 2026-11-02T09:15:30Z", event.At)
			}
			want := `{"version":1,"seq":1,"event":"slice.created","at":"` + event.At + `",` +
				`"data":{"name":"login","title":` + jsonText(t, tt.wantTitle) +
				`,"description":` + jsonText(t, tt.wantDescription) +
				`,"base":"main","base_commit":"` + base + `"}}` + "\n"
			if log != want {
				t.Errorf("the event log is:\n%s\nwant:\n%s", log, want)
			}
		})
	}
}

// --from chooses the base branch, before the git setting sliceflow.base.
// The commit goes on that branch, and no other branch moves.
func TestCreateFromAnotherBranch(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)
	oldMain := gitOutput(t, repo, home, "rev-parse", "main")
	gitOutput(t, repo, home, "config", "sliceflow.base", "dev/bert")
	gitOutput(t, repo, home, "switch", "--quiet", "-c", "release")
	gitOutput(t, repo, home, "commit", "--quiet", "--allow-empty", "-m", "on release")
	oldRelease := gitOutput(t, repo, home, "rev-parse", "release")

	got := runSliceflow(t, repo, home, "slice", "create", "login", "--from", "release")
	wantCode(t, got, 0)

	wantFirst := "Created slice login on branch release.\n"
	if !strings.HasPrefix(got.stdout, wantFirst) {
		t.Errorf("stdout is:\n%s\nwant it to start with:\n%s", got.stdout, wantFirst)
	}
	if parent := gitOutput(t, repo, home, "rev-parse", "release^"); parent != oldRelease {
		t.Errorf("the commit before release is %s, want the old release %s", parent, oldRelease)
	}
	if now := gitOutput(t, repo, home, "rev-parse", "main"); now != oldMain {
		t.Errorf("main is at %s, want it still at %s", now, oldMain)
	}
	if branch := gitOutput(t, repo, home, "branch", "--show-current"); branch != "release" {
		t.Errorf("the project has %q checked out, want release", branch)
	}
}

// An empty --from counts as left out: the base comes from the git
// setting sliceflow.base.
func TestCreateWithAnEmptyFrom(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)
	gitOutput(t, repo, home, "config", "sliceflow.base", "dev")
	gitOutput(t, repo, home, "switch", "--quiet", "-c", "dev")
	oldDev := gitOutput(t, repo, home, "rev-parse", "dev")

	got := runSliceflow(t, repo, home, "slice", "create", "login", "--from", "")
	wantCode(t, got, 0)

	wantFirst := "Created slice login on branch dev.\n"
	if !strings.HasPrefix(got.stdout, wantFirst) {
		t.Errorf("stdout is:\n%s\nwant it to start with:\n%s", got.stdout, wantFirst)
	}
	if parent := gitOutput(t, repo, home, "rev-parse", "dev^"); parent != oldDev {
		t.Errorf("the commit before dev is %s, want the old dev %s", parent, oldDev)
	}
}

// A name with capital letters is made lowercase, and a Note says so.
// Files that a .gitignore rule matches are still in the commit, with a
// Note for each. The Notes come right after the first line: the name's
// first, then the files'.
func TestCreateNotes(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)
	writeText(t, filepath.Join(repo, ".gitignore"), "*.md\n*.jsonl\n")
	gitOutput(t, repo, home, "add", ".gitignore")
	gitOutput(t, repo, home, "commit", "--quiet", "-m", "ignore rules")

	got := runSliceflow(t, repo, home, "slice", "create", "Login")
	wantCode(t, got, 0)

	spec := filepath.Join(topFolder(t, repo, home), "docs", "slices", "login", "spec.md")
	wantStdout := "Created slice login on branch main.\n" +
		"  Note: \"Login\" was made lowercase.\n" +
		"  Note: spec.md matches a .gitignore rule, so sliceflow added it anyway.\n" +
		"  Note: .events.jsonl matches a .gitignore rule, so sliceflow added it anyway.\n" +
		"Next: write the spec in " + spec + "\n"
	if got.stdout != wantStdout {
		t.Errorf("stdout is:\n%s\nwant:\n%s", got.stdout, wantStdout)
	}
	files := gitOutput(t, repo, home, "show", "--name-only", "--format=", "main")
	if files != "docs/slices/login/.events.jsonl\ndocs/slices/login/spec.md" {
		t.Errorf("the commit has the files:\n%s\nwant the spec and the event log", files)
	}
}

// slice create works from a subfolder of the checkout. The slice is
// written at the top of the checkout, not in the subfolder.
func TestCreateFromASubfolder(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)
	sub := filepath.Join(repo, "src", "app")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	got := runSliceflow(t, sub, home, "slice", "create", "login")
	wantCode(t, got, 0)

	spec := filepath.Join(topFolder(t, repo, home), "docs", "slices", "login", "spec.md")
	wantLine := "Next: write the spec in " + spec + "\n"
	if !strings.Contains(got.stdout, wantLine) {
		t.Errorf("stdout is:\n%s\nwant it to have the line:\n%s", got.stdout, wantLine)
	}
	if _, err := os.Stat(filepath.Join(repo, "docs", "slices", "login", "spec.md")); err != nil {
		t.Errorf("the spec is not at the top of the checkout: %v", err)
	}
	wantNoFolder(t, filepath.Join(sub, "docs"))
}

// With no --from, the base branch is the git setting sliceflow.base, and
// the commit goes on that branch.
func TestCreateWithABaseSetting(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)
	oldMain := gitOutput(t, repo, home, "rev-parse", "main")
	gitOutput(t, repo, home, "config", "sliceflow.base", "dev/bert")
	gitOutput(t, repo, home, "switch", "--quiet", "-c", "dev/bert")

	got := runSliceflow(t, repo, home, "slice", "create", "login")
	wantCode(t, got, 0)

	wantFirst := "Created slice login on branch dev/bert.\n"
	if !strings.HasPrefix(got.stdout, wantFirst) {
		t.Errorf("stdout is:\n%s\nwant it to start with:\n%s", got.stdout, wantFirst)
	}
	if parent := gitOutput(t, repo, home, "rev-parse", "dev/bert^"); parent != oldMain {
		t.Errorf("the commit before dev/bert is %s, want %s, where dev/bert started", parent, oldMain)
	}
	if now := gitOutput(t, repo, home, "rev-parse", "main"); now != oldMain {
		t.Errorf("main is at %s, want it still at %s", now, oldMain)
	}
}

// An empty sliceflow.base counts as not set: the base branch is main.
func TestCreateWithAnEmptyBaseSetting(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)
	gitOutput(t, repo, home, "config", "sliceflow.base", "")

	got := runSliceflow(t, repo, home, "slice", "create", "login")
	wantCode(t, got, 0)

	wantFirst := "Created slice login on branch main.\n"
	if !strings.HasPrefix(got.stdout, wantFirst) {
		t.Errorf("stdout is:\n%s\nwant it to start with:\n%s", got.stdout, wantFirst)
	}
}

// slice create changes nothing in the checkout but the slice's folder: a
// staged change and a new file that is not staged are not in its commit,
// and stay as they were.
func TestCreateTouchesOnlyItsFolder(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)
	writeText(t, filepath.Join(repo, "app.go"), "package app\n")
	gitOutput(t, repo, home, "add", "app.go")
	gitOutput(t, repo, home, "commit", "--quiet", "-m", "add app.go")
	writeText(t, filepath.Join(repo, "app.go"), "package app\n\n// changed\n")
	gitOutput(t, repo, home, "add", "app.go")
	writeText(t, filepath.Join(repo, "notes.txt"), "mine\n")

	wantCode(t, runSliceflow(t, repo, home, "slice", "create", "login"), 0)

	files := gitOutput(t, repo, home, "show", "--name-only", "--format=", "main")
	if files != "docs/slices/login/.events.jsonl\ndocs/slices/login/spec.md" {
		t.Errorf("the commit has the files:\n%s\nwant only the spec and the event log", files)
	}
	// "M " is a staged change; "??" is a file git does not track.
	if status := gitOutput(t, repo, home, "status", "--porcelain"); status != "M  app.go\n?? notes.txt" {
		t.Errorf("git status is:\n%s\nwant app.go still staged, and notes.txt still not staged", status)
	}
	if branch := gitOutput(t, repo, home, "branch", "--show-current"); branch != "main" {
		t.Errorf("the project has %q checked out, want main", branch)
	}
}

// The lock of the log is kept inside git's folder, at
// sliceflow/locks/<name>.lock under git's own path for it, and stays
// there after the create. Nothing is left next to the log.
func TestCreateKeepsTheLockInGit(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)

	wantCode(t, runSliceflow(t, repo, home, "slice", "create", "login"), 0)

	if _, err := os.Stat(lockPath(t, repo, home, "login")); err != nil {
		t.Errorf("the lock file is not inside git's folder: %v", err)
	}
	wantNoFolder(t, filepath.Join(repo, "docs", "slices", "login", ".events.jsonl.lock"))
	if status := gitOutput(t, repo, home, "status", "--porcelain"); status != "" {
		t.Errorf("the project has changes:\n%s\nwant none", status)
	}
}

// In a second worktree that the user made, slice create writes the slice
// in that worktree, and commits on its branch. The main checkout does not
// change. The lock is in the second worktree's own git folder.
func TestCreateInASecondWorktree(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)
	oldMain := gitOutput(t, repo, home, "rev-parse", "main")
	second := addWorktree(t, repo, home, "dev/bert")
	gitOutput(t, repo, home, "config", "sliceflow.base", "dev/bert")

	got := runSliceflow(t, second, home, "slice", "create", "login")
	wantCode(t, got, 0)

	spec := filepath.Join(topFolder(t, second, home), "docs", "slices", "login", "spec.md")
	wantStdout := "Created slice login on branch dev/bert.\n" +
		"Next: write the spec in " + spec + "\n"
	if got.stdout != wantStdout {
		t.Errorf("stdout is:\n%s\nwant:\n%s", got.stdout, wantStdout)
	}
	if parent := gitOutput(t, repo, home, "rev-parse", "dev/bert^"); parent != oldMain {
		t.Errorf("the commit before dev/bert is %s, want %s, where dev/bert started", parent, oldMain)
	}
	if now := gitOutput(t, repo, home, "rev-parse", "main"); now != oldMain {
		t.Errorf("main is at %s, want it still at %s", now, oldMain)
	}
	wantNoFolder(t, filepath.Join(repo, "docs", "slices", "login"))
	if _, err := os.Stat(lockPath(t, second, home, "login")); err != nil {
		t.Errorf("the lock file is not in the second worktree's git folder: %v", err)
	}
	wantNoFolder(t, lockPath(t, repo, home, "login"))
}

// The commit runs the user's own git hooks.
func TestCreateRunsTheUsersHooks(t *testing.T) {
	home := newHome(t)
	repo := newProject(t, home)
	hook := filepath.Join(repo, ".git", "hooks", "pre-commit")
	writeText(t, hook, "#!/bin/sh\necho ran > \"$HOME/hook-ran\"\n")
	if err := os.Chmod(hook, 0o755); err != nil {
		t.Fatal(err)
	}

	wantCode(t, runSliceflow(t, repo, home, "slice", "create", "login"), 0)

	if _, err := os.Stat(filepath.Join(home, "hook-ran")); err != nil {
		t.Errorf("the pre-commit hook did not run: %v", err)
	}
}

// topFolder returns the top folder of the checkout at dir, as git gives
// it, written the way this system writes paths. On macOS, the temp
// folder has two names, and git gives the one with its links resolved,
// so a test compares printed paths with this, not with the folder it
// made.
func topFolder(t *testing.T, dir, home string) string {
	t.Helper()
	return filepath.FromSlash(gitOutput(t, dir, home, "rev-parse", "--show-toplevel"))
}

// lockPath returns the full path of the lock of the slice name's log for
// the checkout at dir: sliceflow/locks/<name>.lock under the path git
// gives for it inside its own folder.
func lockPath(t *testing.T, dir, home, name string) string {
	t.Helper()
	path := gitOutput(t, dir, home, "rev-parse", "--path-format=absolute", "--git-path", "sliceflow/locks/"+name+".lock")
	return filepath.FromSlash(path)
}

// addWorktree makes a second worktree of the project in repo, the way a
// user would by hand, on a new branch made from main, in a folder next to
// the project. It returns the worktree's folder.
func addWorktree(t *testing.T, repo, home, branch string) string {
	t.Helper()
	dir := filepath.Join(filepath.Dir(repo), "second")
	gitOutput(t, repo, home, "worktree", "add", "--quiet", "-b", branch, dir)
	return dir
}

// defaultWorktree returns where the worktree of the slice name in the
// project shop goes when sliceflow.worktree.dir is not set.
func defaultWorktree(home, name string) string {
	return filepath.Join(home, ".sliceflow", "worktrees", "shop", name)
}

// wantCode fails the test now unless sliceflow exited with want. It shows
// what sliceflow printed to stderr, which says why.
func wantCode(t *testing.T, got result, want int) {
	t.Helper()
	if got.code != want {
		t.Fatalf("exit code is %d, want %d; stderr:\n%s", got.code, want, got.stderr)
	}
}

// templateBody returns the built-in spec template after its header: the
// part a new spec copies as it is.
func templateBody(t *testing.T) string {
	t.Helper()
	_, body, found := strings.Cut(strings.TrimPrefix(templates.Spec, "---\n"), "\n---\n")
	if !found {
		t.Fatal("the spec template has no header between --- lines")
	}
	return body
}

// firstLines returns the first n lines of text.
func firstLines(text string, n int) string {
	lines := strings.SplitAfterN(text, "\n", n+1)
	if len(lines) > n {
		lines = lines[:n]
	}
	return strings.Join(lines, "")
}

// jsonText returns text as a JSON string, with <, >, and & as they are.
func jsonText(t *testing.T, text string) string {
	t.Helper()
	var b strings.Builder
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(text); err != nil {
		t.Fatal(err)
	}
	return strings.TrimSuffix(b.String(), "\n")
}

// writeText makes the file path with text in it, and the folders above
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
