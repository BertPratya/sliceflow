package cli

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/BertPratya/sliceflow/internal/slice"
)

func TestPrintWarnings(t *testing.T) {
	var b bytes.Buffer
	printWarnings(&b, []slice.Warning{{Slice: "a", Text: "x"}, {Slice: "", Text: "y"}})
	if want := "Warning: a: x\nWarning: y\n"; b.String() != want {
		t.Fatalf("got %q want %q", b.String(), want)
	}
}

func TestFailIndentsLines(t *testing.T) {
	var b bytes.Buffer
	if code := fail(&b, errors.New("a\nb\nc")); code != 1 {
		t.Fatalf("code %d", code)
	}
	if want := "Error: a\n  b\n  c\n"; b.String() != want {
		t.Fatalf("got %q want %q", b.String(), want)
	}
}

func TestUsageError(t *testing.T) {
	var b bytes.Buffer
	if code := usageError(&b, "missing slice name"); code != 2 {
		t.Fatalf("code %d", code)
	}
	if want := "Error: missing slice name. Run sliceflow --help.\n"; b.String() != want {
		t.Fatalf("got %q", b.String())
	}
}

func TestRunTopHelp(t *testing.T) {
	for _, args := range [][]string{nil, {"slice"}, {"--help"}, {"-h"}, {"slice", "--help"}, {"slice", "-h"}, {"spec"}, {"template"}, {"spec", "--help"}} {
		var o, e bytes.Buffer
		if code := Run(args, &o, &e); code != 0 || e.Len() != 0 {
			t.Fatalf("%v: code %d stderr %q", args, code, e.String())
		}
		for _, n := range []string{"slice create", "slice status", "slice list", "slice title <name> <title>", "spec validate [<name>] [--strict]", "spec submit [<name>]", "spec review [<name>]", "template show <template>"} {
			if !strings.Contains(o.String(), n) {
				t.Fatalf("%v: usage lacks %q", args, n)
			}
		}
	}
}

func TestRunUnknown(t *testing.T) {
	var o, e bytes.Buffer
	if code := Run([]string{"slice", "nope"}, &o, &e); code != 2 || o.Len() != 0 ||
		e.String() != "Error: unknown command \"slice nope\". Run sliceflow --help.\n" {
		t.Fatalf("code %d out %q err %q", code, o.String(), e.String())
	}
	o.Reset()
	e.Reset()
	if code := Run([]string{"--colour"}, &o, &e); code != 2 ||
		e.String() != "Error: unknown flag --colour. Run sliceflow --help.\n" {
		t.Fatalf("code %d err %q", code, e.String())
	}
	for _, flag := range []string{"--colour", "--json", "-x"} {
		o.Reset()
		e.Reset()
		want := "Error: unknown flag " + flag + ". Run sliceflow --help.\n"
		if code := Run([]string{"slice", flag}, &o, &e); code != 2 || o.Len() != 0 || e.String() != want {
			t.Fatalf("slice %s: code %d out %q err %q", flag, code, o.String(), e.String())
		}
	}
}
