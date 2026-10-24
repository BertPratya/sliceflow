package e2e

import (
	"strings"
	"testing"

	"github.com/BertPratya/sliceflow/internal/testrepo"
)

func TestHelpTop(t *testing.T) {
	dir := testrepo.New(t)
	for _, args := range [][]string{{"--help"}, {}, {"slice"}} {
		out, errOut, code := run(t, dir, args...)
		if code != 0 || errOut != "" {
			t.Fatalf("%v: code %d stderr %q", args, code, errOut)
		}
		for _, n := range []string{"slice create", "slice status", "slice list"} {
			if !strings.Contains(out, n) {
				t.Errorf("%v: usage lacks %q", args, n)
			}
		}
	}
}

func TestUnknownCommand(t *testing.T) {
	dir := testrepo.New(t)
	out, errOut, code := run(t, dir, "slice", "nope")
	want := "Error: unknown command \"slice nope\". Run sliceflow --help.\n"
	if code != 2 || out != "" || errOut != want {
		t.Fatalf("code %d out %q err %q", code, out, errOut)
	}
}

func TestUnknownFlagAfterSlice(t *testing.T) {
	dir := testrepo.New(t)
	out, errOut, code := run(t, dir, "slice", "--colour")
	want := "Error: unknown flag --colour. Run sliceflow --help.\n"
	if code != 2 || out != "" || errOut != want {
		t.Fatalf("code %d out %q err %q", code, out, errOut)
	}
}
