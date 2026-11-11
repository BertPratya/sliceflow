//go:build !windows

package e2e

// The tests here need a process group (Ctrl+C is sent to the whole group,
// as a terminal does) or a fake git. Windows has neither (plan, Risks).

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"syscall"
	"testing"
	"time"

	"github.com/BertPratya/sliceflow/internal/gittest"
)

// TestCreateCannotRemove checks S12 and R41: git cannot make the commit,
// and the worktree and the branch cannot be removed either. A fake git
// refuses both removals, and runs the real git for every other command.
// sliceflow says what is left, and how to remove it, with the
// --force --force of the worktree.
func TestCreateCannotRemove(t *testing.T) {
	shop := newShop(t)
	fake := t.TempDir()
	writeFakeGit(t, fake)
	path := fake + string(os.PathListSeparator) + os.Getenv("PATH")
	cmd, home := createCmd(t, shop, createOpts{NoIdentity: true, Path: path}, "slice", "create", "card-payment")
	worktree := homeWorktree(home, "card-payment")

	got := runCmd(t, cmd)

	tail := "Some of what sliceflow made could not be removed:\n" +
		"  worktree  " + worktree + "\n" +
		"  branch    slice/card-payment\n" +
		"To clean up, run:\n" +
		"  git worktree remove --force --force \"" + worktree + "\"\n" +
		"  git branch -D \"slice/card-payment\"\n"
	checkGitStep(t, got, commitFailedWithLeft, tail)

	// R41 tells the truth: the branch and the worktree are still there.
	if branches := gittest.Git(t, shop, "branch", "--list", "slice/card-payment"); branches == "" {
		t.Errorf("the branch slice/card-payment is gone, want it left")
	}
	if !slices.Contains(worktreePaths(t, shop), worktree) {
		t.Errorf("git does not list the worktree %s, want it left", worktree)
	}
	if _, err := os.Stat(worktree); err != nil {
		t.Errorf("the worktree folder %s is gone, want it left: %v", worktree, err)
	}
	// Not checked: the slice's files and its lock file stay in the
	// worktree folder, because git keeps the worktree, and the files go
	// only with it. R41 does not list them, and S12 says "removes what it
	// can". The spec and the code disagree here, so this test does not
	// decide it.
}

// TestCreateCtrlC checks S24 and R39: a pre-commit hook makes the commit
// wait, and one Ctrl+C, sent to the whole process group, stops the slice.
// sliceflow removes what it made, prints R39's line, and exits with 1.
func TestCreateCtrlC(t *testing.T) {
	shop := newShop(t)
	mainTip := gittest.Git(t, shop, "rev-parse", "main")
	cmd, home := startWaitingCreate(t, shop)

	interrupt(t, cmd)
	got := runCmd(t, cmd)

	checkStopped(t, got)
	checkNothingLeft(t, shop, home, "card-payment")
	checkTip(t, shop, "main", mainTip)
}

// TestCreateCtrlCMany checks R36 and R39: the terminal sends Ctrl+C five
// times, 2 ms apart, while the commit waits. The first one stops the
// slice. The others change nothing: the clean-up runs in process groups of
// its own, and sliceflow keeps catching Ctrl+C until it has finished. The
// short gap makes more of the signals land while the undo runs.
func TestCreateCtrlCMany(t *testing.T) {
	shop := newShop(t)
	mainTip := gittest.Git(t, shop, "rev-parse", "main")
	cmd, home := startWaitingCreate(t, shop)

	for i := range 5 {
		if i > 0 {
			time.Sleep(2 * time.Millisecond)
		}
		err := syscall.Kill(-cmd.Process.Pid, syscall.SIGINT)
		switch {
		case err == nil:
		case i > 0 && (errors.Is(err, syscall.ESRCH) || errors.Is(err, syscall.EPERM)):
			// sliceflow may have exited already: no process is left in
			// the group (ESRCH), or macOS answers EPERM for it. Later
			// signals may find nothing, which is not a failure.
		default:
			t.Fatalf("send SIGINT %d to the process group: %v", i+1, err)
		}
	}
	got := runCmd(t, cmd)

	checkStopped(t, got)
	checkNothingLeft(t, shop, home, "card-payment")
	checkTip(t, shop, "main", mainTip)
}

// checkStopped checks a run that Ctrl+C stopped (R39): exit code 1,
// nothing on the normal output, and R39's line on the error output.
func checkStopped(t *testing.T, got result) {
	t.Helper()
	checkRefused(t, got, result{stderr: stoppedLine, code: 1})
}

// startWaitingCreate prepares `sliceflow slice create card-payment` in
// the repo shop, as createCmd does. A pre-commit hook writes a file and
// then waits, so the commit waits too. The command is started in a
// process group of its own (Setpgid), as a shell starts a job, and the
// test waits until the hook has written its file. It returns the command
// and the home folder.
func startWaitingCreate(t *testing.T, shop string) (*exec.Cmd, string) {
	t.Helper()
	waiting := filepath.Join(t.TempDir(), "waiting")
	writeHook(t, shop, "pre-commit", waitingHookScript(waiting))
	cmd, home := createCmd(t, shop, createOpts{}, "slice", "create", "card-payment")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		t.Fatalf("start sliceflow: %v", err)
	}
	// If the test stops before runCmd waits for sliceflow, kill its whole
	// group, so no process is left running. The cleanups run in the
	// test's goroutine, after the test, so ProcessState tells whether
	// runCmd has waited already.
	t.Cleanup(func() {
		if cmd.ProcessState != nil {
			return
		}
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		_ = cmd.Wait()
	})
	waitForFile(t, waiting)
	return cmd, home
}

// waitingHookScript is a pre-commit hook script that writes the file
// waiting, then sleeps for 10 seconds, and exits with 0. The test
// interrupts sliceflow before the sleep ends, so the test never hangs.
func waitingHookScript(waiting string) string {
	return "#!/bin/sh\n" +
		"touch '" + waiting + "'\n" +
		"sleep 10\n" +
		"exit 0\n"
}

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

// interrupt sends SIGINT to the whole process group of cmd, as a
// terminal does for Ctrl+C. Every process of the group gets it: sliceflow,
// git, and the hook.
func interrupt(t *testing.T, cmd *exec.Cmd) {
	t.Helper()
	if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGINT); err != nil {
		t.Fatalf("send SIGINT to the process group: %v", err)
	}
}

// writeFakeGit writes in dir a script named git. The script refuses
// `git worktree remove` and `git branch -D` with exit code 1, and runs the
// real git, found before PATH changes, for every other command, with all
// its arguments. The script uses only the shell's own commands, so the
// rest of PATH is kept after dir.
func writeFakeGit(t *testing.T, dir string) {
	t.Helper()
	gitPath, err := exec.LookPath("git")
	if err != nil {
		t.Fatalf("find git: %v", err)
	}
	script := "#!/bin/sh\n" +
		"if [ \"$1\" = worktree ] && [ \"$2\" = remove ]; then\n" +
		"  echo 'simulated: the worktree cannot be removed' >&2\n" +
		"  exit 1\n" +
		"fi\n" +
		"if [ \"$1\" = branch ] && [ \"$2\" = -D ]; then\n" +
		"  echo 'simulated: the branch cannot be deleted' >&2\n" +
		"  exit 1\n" +
		"fi\n" +
		"exec '" + gitPath + "' \"$@\"\n"
	if err := os.WriteFile(filepath.Join(dir, "git"), []byte(script), 0o755); err != nil {
		t.Fatalf("write the fake git: %v", err)
	}
}
