package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

// TestCatchCrashExitsWith1 checks TD4 and R76: a crash prints Go's
// "panic:" line, a blank line, and the stack, and gives exit code 1.
func TestCatchCrashExitsWith1(t *testing.T) {
	var errOut bytes.Buffer
	code := catchCrash(&errOut, func() int { panic("boom") })

	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	got := errOut.String()
	if !strings.HasPrefix(got, "panic: boom\n\ngoroutine ") {
		t.Errorf("error output does not start with the panic line, a blank line, and the stack:\n%s", got)
	}
	// The stack names the function that crashed.
	if !strings.Contains(got, "TestCatchCrashExitsWith1") {
		t.Errorf("stack does not name the function that crashed:\n%s", got)
	}
}

// TestCatchCrashError checks that a crash with an error prints the
// error's text, as Go does.
func TestCatchCrashError(t *testing.T) {
	var errOut bytes.Buffer
	code := catchCrash(&errOut, func() int { panic(errors.New("disk on fire")) })

	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if got := errOut.String(); !strings.HasPrefix(got, "panic: disk on fire\n\n") {
		t.Errorf("error output = %q, want it to start with %q", got, "panic: disk on fire\n\n")
	}
}

// TestCatchCrashNoCrash checks that the exit code of work is kept, and
// nothing is printed, when work does not crash.
func TestCatchCrashNoCrash(t *testing.T) {
	for _, want := range []int{0, 1, 2} {
		var errOut bytes.Buffer
		code := catchCrash(&errOut, func() int { return want })

		if code != want {
			t.Errorf("exit code = %d, want %d", code, want)
		}
		if errOut.Len() != 0 {
			t.Errorf("error output = %q, want nothing", errOut.String())
		}
	}
}
