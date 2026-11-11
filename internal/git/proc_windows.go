package git

import (
	"os/exec"
	"syscall"
)

// interrupt returns the function that run calls when its context ends
// (exec.Cmd.Cancel). On Windows, one process cannot be sent a Ctrl+C, so
// the function does nothing, and returns nil. A git in the console got
// the user's Ctrl+C from the console already. If git does not exit, Go
// kills it when WaitDelay runs out, 5 seconds later.
//
// Because the function returns nil, run returns ctx.Err(), even when git
// exits with 0. The caller then asks git what is there.
func interrupt(cmd *exec.Cmd) func() error {
	return func() error {
		return nil
	}
}

// shield makes cmd start in a new process group of its own
// (CREATE_NEW_PROCESS_GROUP). Windows turns Ctrl+C off for the processes
// of such a group, so a Ctrl+C in the console does not reach cmd.
func shield(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}
}
