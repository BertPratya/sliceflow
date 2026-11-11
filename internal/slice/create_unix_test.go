//go:build !windows

package slice_test

// This test uses a shell hook, and stops git with an interrupt, which
// Windows does not have (see interrupt in internal/git/proc_windows.go).

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/BertPratya/sliceflow/internal/gittest"
	"github.com/BertPratya/sliceflow/internal/slice"
)

// waitForFile waits until the file at path exists, for at most 10
// seconds, and fails the test when it does not.
func waitForFile(t *testing.T, path string) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("the file %s did not appear in 10 seconds", path)
}

// writeHook writes the git hook name, a shell script, in the repo's
// hooks folder, and makes it executable.
func writeHook(t *testing.T, r *gittest.Repo, name, script string) {
	t.Helper()
	r.WriteFile(".git/hooks/"+name, script)
	if err := os.Chmod(filepath.Join(r.Dir, ".git", "hooks", name), 0o755); err != nil {
		t.Fatal(err)
	}
}

// waitingHook returns a hook script that writes the file waiting, then
// waits for the file release, for at most 10 seconds, and exits with 0.
func waitingHook(waiting, release string) string {
	return "#!/bin/sh\n" +
		"touch '" + waiting + "'\n" +
		"i=0\n" +
		"while [ ! -f '" + release + "' ] && [ $i -lt 100 ]; do sleep 0.1; i=$((i+1)); done\n" +
		"exit 0\n"
}

// checkStoppedWhileHookWaits installs a hook, hookName, that waits, runs
// Create, cancels the context while the hook waits, then lets the hook
// go on. Create must return R39's line, and leave nothing it made.
func checkStoppedWhileHookWaits(t *testing.T, hookName string) {
	r := gittest.New(t)
	path := worktreePath(t, findProject(t, r.Dir))
	signals := t.TempDir()
	waiting := filepath.Join(signals, "waiting")
	release := filepath.Join(signals, "release")
	writeHook(t, r, hookName, waitingHook(waiting, release))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		_, err := slice.Create(ctx, r.Dir, slice.CreateInput{GivenName: "login"})
		done <- err
	}()

	waitForFile(t, waiting)
	cancel()
	if err := os.WriteFile(release, nil, 0o644); err != nil {
		t.Fatal(err)
	}

	var err error
	select {
	case err = <-done:
	case <-time.After(30 * time.Second):
		t.Fatalf("Create did not return 30 seconds after the cancel")
	}
	want := "Stopped. The slice was not created. Nothing was changed."
	if err == nil || err.Error() != want {
		t.Fatalf("error:\n%v\nwant:\n%s", err, want)
	}
	checkGone(t, r, "login", path)
	if made := filepath.Join(r.Home, ".sliceflow"); exists(t, made) {
		t.Errorf("the folder %s, made for the worktree, is still there", made)
	}
}

// R39: when the context ends while a pre-commit hook waits, Create
// returns R39's line, and nothing it made is left.
func TestCreateStopped(t *testing.T) {
	checkStoppedWhileHookWaits(t, "pre-commit")
}

// R36, R39: when the context ends while a post-checkout hook of git
// worktree add waits, git has made the worktree already. Create asks git
// whether it is there anyway, removes it, and returns R39's line.
func TestCreateStoppedDuringWorktree(t *testing.T) {
	checkStoppedWhileHookWaits(t, "post-checkout")
}

// R36, R37: a post-checkout hook that exits with 1 makes git worktree
// add fail after it made the worktree. Create asks git whether the
// worktree is there anyway, and removes it, the branch, and the folders
// made for it.
func TestCreateWorktreeThereAnyway(t *testing.T) {
	r := gittest.New(t)
	path := worktreePath(t, findProject(t, r.Dir))
	writeHook(t, r, "post-checkout", "#!/bin/sh\necho hook refused >&2\nexit 1\n")

	_, err := slice.Create(context.Background(), r.Dir, slice.CreateInput{GivenName: "login"})

	wantStart := "Error: git could not create the worktree. The slice was not created. Nothing was changed.\ngit said:\n"
	if err == nil || !strings.HasPrefix(err.Error(), wantStart) {
		t.Fatalf("error:\n%v\nwant it to start with:\n%s", err, wantStart)
	}
	checkGone(t, r, "login", path)
	if made := filepath.Join(r.Home, ".sliceflow"); exists(t, made) {
		t.Errorf("the folder %s, made for the worktree, is still there", made)
	}
}
