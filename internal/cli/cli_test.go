package cli

import (
	"bytes"
	"errors"
	"testing"
)

// TestHelpWorksWithoutFolder checks that help does not need the current
// folder, and that workDir gives the error back.
func TestHelpWorksWithoutFolder(t *testing.T) {
	var stdout, stderr bytes.Buffer
	e := &env{stdout: &stdout, stderr: &stderr, dirErr: errors.New("no folder")}

	if code := rootCommands.route(e, []string{"slice", "--help"}); code != exitOK {
		t.Errorf("code = %d, want 0", code)
	}
	if stdout.Len() == 0 || stderr.Len() != 0 {
		t.Errorf("stdout = %q, stderr = %q", stdout.String(), stderr.String())
	}
	if _, err := e.workDir(); err == nil {
		t.Error("workDir() gave no error")
	}
}
