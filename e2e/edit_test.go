package e2e

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// newLoginSlice makes a project with the slice login, titled "Log in",
// with the description "Lets a user log in.", on main, and returns the
// project's folder.
func newLoginSlice(t *testing.T, home string) string {
	t.Helper()
	repo := newProject(t, home)
	createSlice(t, repo, home, "login", "--title", "Log in", "--description", "Lets a user log in.")
	return repo
}

// lastLine returns the last line of the log of login in repo.
func lastLine(t *testing.T, repo string) string {
	t.Helper()
	data, err := os.ReadFile(sliceLog(repo, "login"))
	if err != nil {
		t.Fatalf("could not read the log: %v", err)
	}
	lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	return lines[len(lines)-1]
}

// readLog returns the log of login in repo.
func readLog(t *testing.T, repo string) string {
	t.Helper()
	data, err := os.ReadFile(sliceLog(repo, "login"))
	if err != nil {
		t.Fatalf("could not read the log: %v", err)
	}
	return string(data)
}

// slice edit changes the title, or the title and the description, and
// prints the values the slice has now.
func TestEditChangesTheSlice(t *testing.T) {
	t.Run("the title", func(t *testing.T) {
		home := newHome(t)
		repo := newLoginSlice(t, home)

		got := runSliceflow(t, repo, home, "slice", "edit", "login", "--title", "Sign in")

		wantOutput(t, got, "Edited slice login.\n"+
			"  Title:        Sign in\n"+
			"  Description:  Lets a user log in.\n", "")
		status := runSliceflow(t, repo, home, "slice", "status", "login")
		if !strings.HasPrefix(status.stdout, "login: Sign in\n") {
			t.Errorf("slice status printed:\n%s\nwant the new title", status.stdout)
		}
	})

	t.Run("both", func(t *testing.T) {
		home := newHome(t)
		repo := newLoginSlice(t, home)

		got := runSliceflow(t, repo, home, "slice", "edit", "login", "--title", "Sign in", "--description", "Lets a user sign in with a password.")

		wantOutput(t, got, "Edited slice login.\n"+
			"  Title:        Sign in\n"+
			"  Description:  Lets a user sign in with a password.\n", "")
		if msg := gitOutput(t, repo, home, "log", "-1", "--format=%s"); msg != "login: edit the title and description" {
			t.Errorf("the commit message is %q, want %q", msg, "login: edit the title and description")
		}
	})
}

// The event holds only the values that differ.
func TestEditedEvent(t *testing.T) {
	home := newHome(t)
	repo := newLoginSlice(t, home)

	got := runSliceflow(t, repo, home, "slice", "edit", "login", "--title", "Sign in", "--description", "Lets a user log in.")

	wantCode(t, got, 0)
	line := lastLine(t, repo)
	if !strings.Contains(line, `"event":"slice.edited"`) || !strings.Contains(line, `"data":{"title":"Sign in"}`) {
		t.Errorf("the last line of the log is %s, want slice.edited with the data {\"title\":\"Sign in\"}", line)
	}
	if !strings.Contains(line, `"seq":2`) {
		t.Errorf("the last line of the log is %s, want seq 2", line)
	}
}

// An empty title sets the title back to the name, and an empty
// description removes it.
func TestEditEmptyValues(t *testing.T) {
	t.Run("title", func(t *testing.T) {
		home := newHome(t)
		repo := newLoginSlice(t, home)
		wantCode(t, runSliceflow(t, repo, home, "slice", "edit", "login", "--title", "Sign in"), 0)

		got := runSliceflow(t, repo, home, "slice", "edit", "login", "--title", "")

		wantCode(t, got, 0)
		if !strings.Contains(got.stdout, "\n  Title:        login\n") {
			t.Errorf("stdout is:\n%s\nwant the title login", got.stdout)
		}
		if line := lastLine(t, repo); !strings.Contains(line, `"data":{"title":"login"}`) {
			t.Errorf("the last line of the log is %s, want the data {\"title\":\"login\"}", line)
		}
	})

	t.Run("description", func(t *testing.T) {
		home := newHome(t)
		repo := newLoginSlice(t, home)

		got := runSliceflow(t, repo, home, "slice", "edit", "login", "--description", "")

		wantCode(t, got, 0)
		if !strings.HasSuffix(got.stdout, "\n  Description:  (none)\n") {
			t.Errorf("stdout is:\n%s\nwant the description (none)", got.stdout)
		}
		if line := lastLine(t, repo); !strings.Contains(line, `"data":{"description":""}`) {
			t.Errorf("the last line of the log is %s, want the data {\"description\":\"\"}", line)
		}
		if msg := gitOutput(t, repo, home, "log", "-1", "--format=%s"); msg != "login: edit the description" {
			t.Errorf("the commit message is %q, want %q", msg, "login: edit the description")
		}
	})
}

// The same values add no event and no commit.
func TestEditNothingToChange(t *testing.T) {
	home := newHome(t)
	repo := newLoginSlice(t, home)
	wantCode(t, runSliceflow(t, repo, home, "slice", "edit", "login", "--title", "Sign in"), 0)
	before := gitState(t, repo, home)
	log := readLog(t, repo)

	got := runSliceflow(t, repo, home, "slice", "edit", "login", "--title", "Sign in")

	wantOutput(t, got, "Nothing to change: slice login already has this title and description.\n", "")
	wantSameState(t, repo, home, before)
	if after := readLog(t, repo); after != log {
		t.Errorf("the log changed from:\n%s\nto:\n%s", log, after)
	}

	t.Run("an empty title when the title is the name", func(t *testing.T) {
		wantCode(t, runSliceflow(t, repo, home, "slice", "edit", "login", "--title", ""), 0)
		got := runSliceflow(t, repo, home, "slice", "edit", "login", "--title", "")
		wantOutput(t, got, "Nothing to change: slice login already has this title and description.\n", "")
	})
}

// With neither option, slice edit changes nothing, and exits with 2.
func TestEditNeedsAnOption(t *testing.T) {
	home := newHome(t)
	repo := newLoginSlice(t, home)
	before := gitState(t, repo, home)

	got := runSliceflow(t, repo, home, "slice", "edit", "login")

	wantError(t, got, 2, "Error: slice edit needs --title, --description, or both. Nothing was changed.\n"+
		"To see the options, run: sliceflow slice edit --help\n")
	wantSameState(t, repo, home, before)
}

// slice edit makes one commit, on the base branch, of only the log. The
// user's other changes, in the slice's folder too, and what is staged
// stay as they were.
func TestEditCommitsTheLog(t *testing.T) {
	home := newHome(t)
	repo := newLoginSlice(t, home)
	base := gitOutput(t, repo, home, "rev-parse", "main")
	writeText(t, filepath.Join(sliceFolder(repo, "login"), "spec.md"), "my draft\n")
	writeText(t, filepath.Join(repo, "notes.md"), "staged\n")
	gitOutput(t, repo, home, "add", "notes.md")

	got := runSliceflow(t, repo, home, "slice", "edit", "login", "--title", "Sign in")

	wantCode(t, got, 0)
	if parent := gitOutput(t, repo, home, "rev-parse", "main^"); parent != base {
		t.Errorf("the commit before main is %s, want the old main %s: one new commit", parent, base)
	}
	if msg := gitOutput(t, repo, home, "log", "-1", "--format=%s"); msg != "login: edit the title" {
		t.Errorf("the commit message is %q, want %q", msg, "login: edit the title")
	}
	if files := gitOutput(t, repo, home, "show", "--name-only", "--format=", "main"); files != "docs/slices/login/.events.jsonl" {
		t.Errorf("the commit holds:\n%s\nwant only the log", files)
	}
	if status := gitOutput(t, repo, home, "status", "--porcelain"); status != "A  notes.md\n M docs/slices/login/spec.md" &&
		status != " M docs/slices/login/spec.md\nA  notes.md" {
		t.Errorf("git status is:\n%s\nwant notes.md staged and spec.md changed, not staged", status)
	}
	if branch := gitOutput(t, repo, home, "branch", "--show-current"); branch != "main" {
		t.Errorf("the checkout is on %s, want main", branch)
	}
}

// slice edit works only in a checkout on the slice's base branch.
func TestEditOnTheBase(t *testing.T) {
	t.Run("on another branch", func(t *testing.T) {
		home := newHome(t)
		repo := newProject(t, home)
		gitOutput(t, repo, home, "switch", "--quiet", "-c", "dev/bert")
		createSlice(t, repo, home, "login", "--from", "dev/bert")
		gitOutput(t, repo, home, "switch", "--quiet", "-c", "fix/typo")
		before := gitState(t, repo, home)

		got := runSliceflow(t, repo, home, "slice", "edit", "login", "--title", "Sign in")

		wantError(t, got, 1, "Error: slice login is planned on branch dev/bert, but your checkout is on fix/typo. Nothing was changed.\n"+
			"Switch to dev/bert, then try again.\n")
		wantSameState(t, repo, home, before)
	})

	t.Run("on no branch", func(t *testing.T) {
		home := newHome(t)
		repo := newLoginSlice(t, home)
		gitOutput(t, repo, home, "switch", "--quiet", "--detach")

		got := runSliceflow(t, repo, home, "slice", "edit", "login", "--title", "Sign in")

		wantError(t, got, 1, "Error: slice login is planned on branch main, but your checkout is not on a branch. Nothing was changed.\n"+
			"Switch to main, then try again.\n")
	})
}

// When the commit fails, slice edit puts the log back, unstages it, and
// shows git's text.
func TestEditUndo(t *testing.T) {
	home := newHome(t)
	repo := newLoginSlice(t, home)
	writeHook(t, repo, "pre-commit", "#!/bin/sh\necho 'commit refused' >&2\nexit 1\n")
	before := gitState(t, repo, home)
	log := readLog(t, repo)

	got := runSliceflow(t, repo, home, "slice", "edit", "login", "--title", "Sign in")

	wantError(t, got, 1, "Error: git could not make the commit. Nothing was changed.\n"+
		"git said:\n"+
		"  commit refused\n")
	wantSameState(t, repo, home, before)
	if after := readLog(t, repo); after != log {
		t.Errorf("the log changed from:\n%s\nto:\n%s", log, after)
	}
}

// Ctrl+C while the commit waits on a hook: slice edit puts the log back,
// and says it stopped.
func TestEditStopsOnCtrlC(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Go cannot send Ctrl+C to one process on Windows")
	}
	home := newHome(t)
	repo := newLoginSlice(t, home)
	writeHook(t, repo, "pre-commit", "#!/bin/sh\nexec >/dev/null 2>&1\ntouch \"$HOME/hook-started\"\nsleep 10\n")
	before := gitState(t, repo, home)
	log := readLog(t, repo)

	got := interruptWhen(t, repo, home, filepath.Join(home, "hook-started"), "slice", "edit", "login", "--title", "Sign in")

	wantError(t, got, 1, "Stopped. Nothing was changed.\n")
	wantSameState(t, repo, home, before)
	if after := readLog(t, repo); after != log {
		t.Errorf("the log changed from:\n%s\nto:\n%s", log, after)
	}
}

// Ctrl+C after the commit is made: slice edit puts nothing back.
func TestEditStopsAfterTheCommit(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Go cannot send Ctrl+C to one process on Windows")
	}
	home := newHome(t)
	repo := newLoginSlice(t, home)
	base := gitOutput(t, repo, home, "rev-parse", "main")
	writeHook(t, repo, "post-commit", "#!/bin/sh\nexec >/dev/null 2>&1\ntouch \"$HOME/hook-started\"\nsleep 10\n")

	got := interruptWhen(t, repo, home, filepath.Join(home, "hook-started"), "slice", "edit", "login", "--title", "Sign in")

	wantError(t, got, 1, "Stopped after the commit. Slice login was edited on branch main.\n")
	if parent := gitOutput(t, repo, home, "rev-parse", "main^"); parent != base {
		t.Errorf("the commit before main is %s, want the old main %s", parent, base)
	}
	if status := gitOutput(t, repo, home, "status", "--porcelain"); status != "" {
		t.Errorf("the project has changes:\n%s\nwant none", status)
	}
}

// While the state is unknown, slice edit changes nothing, after the
// warning, with a hint for the reason.
func TestEditWithoutState(t *testing.T) {
	t.Run("an unknown event", func(t *testing.T) {
		home := newHome(t)
		repo := newLoginSlice(t, home)
		appendText(t, sliceLog(repo, "login"), eventLine(2, "spec.submitted", `{}`))
		log := readLog(t, repo)

		got := runSliceflow(t, repo, home, "slice", "edit", "login", "--title", "Sign in")

		wantError(t, got, 1, unknownEventLine2+
			"Error: the state of slice login cannot be worked out, so it cannot be changed. Nothing was changed.\n"+
			"Update sliceflow, then try again.\n")
		if after := readLog(t, repo); after != log {
			t.Errorf("the log changed from:\n%s\nto:\n%s", log, after)
		}
	})

	t.Run("an event that is not allowed", func(t *testing.T) {
		home := newHome(t)
		repo := newLoginSlice(t, home)
		appendText(t, sliceLog(repo, "login"), secondCreatedLine(t, repo))

		got := runSliceflow(t, repo, home, "slice", "edit", "login", "--title", "Sign in")

		wantError(t, got, 1, notAllowedLine2+
			"Error: the state of slice login cannot be worked out, so it cannot be changed. Nothing was changed.\n"+
			"To see how the log changed, run: git log -p -- docs/slices/login/.events.jsonl\n")
	})
}

// slice edit reports the mistakes slice create and slice status report
// with the same messages, in its own order.
func TestEditErrorsAsBefore(t *testing.T) {
	home := newHome(t)
	repo := newLoginSlice(t, home)

	got := runSliceflow(t, repo, home, "slice", "edit", "a.b", "--title", "x")
	wantError(t, got, 2, "Error: invalid slice name \"a.b\". Nothing was changed.\n"+
		"Allowed characters: letters, digits, hyphens, underscores. Start with a letter or a digit.\n")

	got = runSliceflow(t, repo, home, "slice", "edit", "a.b", "--title", "a\tb")
	wantError(t, got, 2, "Error: --title has a control character. Nothing was changed.\n"+
		"Received: \"a\\tb\"\n"+
		"Write it as plain text on one line, with no control characters.\n")

	got = runSliceflow(t, repo, home, "slice", "edit", "nope", "--title", "x")
	wantError(t, got, 1, "Error: no slice named nope. Nothing was changed.\n"+
		"No log was found at docs/slices/nope/.events.jsonl.\n"+
		"Your checkout is on branch main. Check the name, or switch to the branch the slice is planned on.\n")

	writeText(t, sliceLog(repo, "login"), "cut off\n")
	got = runSliceflow(t, repo, home, "slice", "edit", "login", "--title", "x")
	wantError(t, got, 1, "Warning: line 1 of docs/slices/login/.events.jsonl could not be read, so it was skipped.\n"+
		"Error: no slice.created event could be read for login. Nothing was changed.\n"+
		"Fix or restore the log, then try again.\n")
}

// A damaged line is skipped with its warning, and the edit still works.
func TestEditSkipsADamagedLine(t *testing.T) {
	home := newHome(t)
	repo := newLoginSlice(t, home)
	appendText(t, sliceLog(repo, "login"), "cut off\n")
	gitOutput(t, repo, home, "commit", "--quiet", "-am", "a damaged line")

	got := runSliceflow(t, repo, home, "slice", "edit", "login", "--title", "Sign in")

	wantOutput(t, got, "Edited slice login.\n"+
		"  Title:        Sign in\n"+
		"  Description:  Lets a user log in.\n",
		"Warning: line 2 of docs/slices/login/.events.jsonl could not be read, so it was skipped.\n")
}

// slice edit --help prints its help, in the layout of slice create.
func TestEditHelp(t *testing.T) {
	got := runSliceflow(t, t.TempDir(), newHome(t), "slice", "edit", "--help")

	wantOutput(t, got, "Change a slice's title, its description, or both, in one commit on its base branch.\n"+
		"\n"+
		"Usage:\n"+
		"  sliceflow slice edit <name> [options]\n"+
		"\n"+
		"Arguments:\n"+
		"  <name>  The name of the slice to change.\n"+
		"\n"+
		"Options:\n"+
		"  --description <text>  The new description. Empty removes it.\n"+
		"  --title <text>        The new title. Empty sets it back to the name.\n"+
		"  -h, --help            Show this help.\n", "")
}
