//go:build !windows

package git

import (
	"os"
	"os/exec"
	"syscall"
)

// interrupt returns the function that run calls when its context ends
// (exec.Cmd.Cancel). It sends git an interrupt, as Ctrl+C does, so git
// can clean up after itself. When git has exited already, the error is
// os.ErrProcessDone, and run returns what git did.
//
// It signals only git, not the programs git started, such as a hook. A
// Ctrl+C in the terminal reaches them all; otherwise WaitDelay closes
// their pipes 5 seconds later.
func interrupt(cmd *exec.Cmd) func() error {
	return func() error {
		return cmd.Process.Signal(os.Interrupt)
	}
}

// shield makes cmd start in a new process group of its own (Setpgid). A
// Ctrl+C in the terminal goes only to the terminal's process group, so
// it does not reach cmd.
func shield(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}
