package spec

import (
	"os"
	"testing"
)

// The 0002 spec is written by the rules it describes, so it has no
// problems.
func TestOwnSpecHasNoProblems(t *testing.T) {
	data, err := os.ReadFile("../../docs/slices/spec/spec.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range Check(Read(data)) {
		t.Errorf("line %d: %s\n  %s", p.Line, p.Message, p.Hint)
	}
}
