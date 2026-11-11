//go:build !windows

package git

// These tests send signals and read process groups, which Windows does
// not have. On Windows, run cannot send git an interrupt, and the
// process group of runShielded is set with another flag (proc_windows.go).

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// The files that the fake git in the mode "trap" writes in its folder.
const (
	readyFile        = "ready"
	gotInterruptFile = "got-interrupt"
)

// The file that the fake git in the mode "pgrp-file" writes in its
// folder.
const pgrpFile = "pgrp"

func init() {
	// "trap": catch the interrupt, write readyFile, and wait. On the
	// interrupt, write gotInterruptFile and exit with 0. Without one,
	// exit with 96 after a minute.
	fakeModes["trap"] = func() {
		interrupts := make(chan os.Signal, 1)
		signal.Notify(interrupts, os.Interrupt)
		if err := os.WriteFile(readyFile, nil, 0o644); err != nil {
			os.Exit(97)
		}
		select {
		case <-interrupts:
			if err := os.WriteFile(gotInterruptFile, nil, 0o644); err != nil {
				os.Exit(98)
			}
			os.Exit(0)
		case <-time.After(time.Minute):
			os.Exit(96)
		}
	}
	// "pgrp-file": write the fake git's process group to the file
	// pgrpFile in its folder, and exit with 0. It is for the functions
	// that do not return git's output.
	fakeModes["pgrp-file"] = func() {
		if err := os.WriteFile(pgrpFile, []byte(strconv.Itoa(syscall.Getpgrp())), 0o644); err != nil {
			os.Exit(97)
		}
		os.Exit(0)
	}
	// "pgrp": print the fake git's process group, and exit with 0.
	fakeModes["pgrp"] = func() {
		os.Stdout.WriteString(strconv.Itoa(syscall.Getpgrp()))
		os.Exit(0)
	}
}

// When the context ends while git runs, run sends git an interrupt, not
// a kill, so git can clean up after itself (R39). The fake git exits
// with 0 after the interrupt, and run still returns the context's error.
func TestRunInterruptsOnCancel(t *testing.T) {
	useFakeGit(t, "trap")
	dir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		_, err := run(ctx, dir, "anything")
		done <- err
	}()

	// Wait until the fake git has set its trap, then cancel.
	waitForFile(t, filepath.Join(dir, readyFile), done)
	cancel()

	var err error
	select {
	case err = <-done:
	case <-time.After(30 * time.Second):
		t.Fatalf("run did not return 30 seconds after the cancel")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("run returned the error %v (%T), want context.Canceled", err, err)
	}
	if _, statErr := os.Stat(filepath.Join(dir, gotInterruptFile)); statErr != nil {
		t.Errorf("the fake git did not get the interrupt: %v", statErr)
	}
}

// runShielded starts git in a process group of its own, so a Ctrl+C in
// the terminal does not reach it (R39). run keeps git in the test's own
// group.
func TestRunShieldedOwnGroup(t *testing.T) {
	useFakeGit(t, "pgrp")
	own := syscall.Getpgrp()

	out, err := runShielded(context.Background(), t.TempDir(), "anything")
	if err != nil {
		t.Fatalf("runShielded returned the error %v, want none", err)
	}
	if got := readGroup(t, out); got == own {
		t.Errorf("with runShielded, git's process group is %d, the test's own; want another", got)
	}

	out, err = run(context.Background(), t.TempDir(), "anything")
	if err != nil {
		t.Fatalf("run returned the error %v, want none", err)
	}
	if got := readGroup(t, out); got != own {
		t.Errorf("with run, git's process group is %d, want the test's own, %d", got, own)
	}
}

// The functions of the undo run git shielded from Ctrl+C, in a process
// group of their own (R36, R39).
func TestUndoCallsAreShielded(t *testing.T) {
	useFakeGit(t, "pgrp-file")
	ctx := context.Background()
	own := syscall.Getpgrp()

	calls := map[string]func(dir string) error{
		"DeleteBranch": func(dir string) error {
			return DeleteBranch(ctx, dir, "slice/login")
		},
		"RemoveWorktree": func(dir string) error {
			return RemoveWorktree(ctx, dir, filepath.Join(dir, "login"))
		},
		"WorktreeRegistered": func(dir string) error {
			_, err := WorktreeRegistered(ctx, dir, filepath.Join(dir, "login"))
			return err
		},
		"BranchExistsShielded": func(dir string) error {
			_, err := BranchExistsShielded(ctx, dir, "slice/login")
			return err
		},
	}
	for name, call := range calls {
		dir := t.TempDir()
		if err := call(dir); err != nil {
			t.Errorf("%s returned the error %v, want none", name, err)
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, pgrpFile))
		if err != nil {
			t.Errorf("%s: the fake git wrote no process group: %v", name, err)
			continue
		}
		if got := readGroup(t, string(data)); got == own {
			t.Errorf("%s ran git in the test's own process group, %d; want another", name, got)
		}
	}
}

// waitForFile waits until the file at path exists, for at most 10
// seconds. It fails the test at once when run returns first, on done.
func waitForFile(t *testing.T, path string, done <-chan error) {
	t.Helper()
	limit := time.After(10 * time.Second)
	for {
		if _, err := os.Stat(path); err == nil {
			return
		}
		select {
		case err := <-done:
			t.Fatalf("run returned %v before %s was written", err, path)
		case <-limit:
			t.Fatalf("%s was not written within 10 seconds", path)
		case <-time.After(10 * time.Millisecond):
		}
	}
}

// readGroup reads the process group that the fake git printed.
func readGroup(t *testing.T, out string) int {
	t.Helper()
	group, err := strconv.Atoi(strings.TrimSpace(out))
	if err != nil {
		t.Fatalf("read the process group %q: %v", out, err)
	}
	return group
}
