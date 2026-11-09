package cli

import (
	"errors"
	"testing"
)

// TestExitCode checks R76 for nil and for an error that is not wrong
// use.
func TestExitCode(t *testing.T) {
	if code := exitCode(nil); code != 0 {
		t.Errorf("exitCode(nil) = %d, want 0", code)
	}
	if code := exitCode(errors.New("x")); code != 1 {
		t.Errorf("exitCode(plain error) = %d, want 1", code)
	}
}
