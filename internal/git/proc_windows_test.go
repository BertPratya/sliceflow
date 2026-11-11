//go:build windows

package git

import (
	"os/exec"
	"syscall"
	"testing"
)

// shield puts git in a new process group, which Windows keeps away from
// the console's Ctrl+C (R39).
func TestShieldSetsNewProcessGroup(t *testing.T) {
	cmd := exec.Command("git")
	shield(cmd)
	if cmd.SysProcAttr == nil || cmd.SysProcAttr.CreationFlags&syscall.CREATE_NEW_PROCESS_GROUP == 0 {
		t.Errorf("SysProcAttr = %+v, want CreationFlags with CREATE_NEW_PROCESS_GROUP", cmd.SysProcAttr)
	}
}

// interrupt cannot send a Ctrl+C to one process on Windows: it does
// nothing, and returns nil, so run returns the context's error.
func TestInterruptReturnsNil(t *testing.T) {
	cmd := exec.Command("git")
	if err := interrupt(cmd)(); err != nil {
		t.Errorf("interrupt(cmd)() = %v, want nil", err)
	}
}
