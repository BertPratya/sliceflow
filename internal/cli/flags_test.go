package cli

import (
	"bytes"
	"errors"
	"flag"
	"reflect"
	"testing"
)

func newFS() (*flag.FlagSet, *string, *bool) {
	fs := flag.NewFlagSet("t", flag.ContinueOnError)
	title := fs.String("title", "", "")
	js := fs.Bool("json", false, "")
	return fs, title, js
}

func TestParseFlagsAfterName(t *testing.T) {
	fs, title, js := newFS()
	var out bytes.Buffer
	pos, err := parseFlags(fs, "usage", []string{"login-google", "--title", "Login with Google", "--json"}, &out)
	if err != nil || !reflect.DeepEqual(pos, []string{"login-google"}) || *title != "Login with Google" || !*js {
		t.Fatalf("got %v %q %v %v", pos, *title, *js, err)
	}

	fs, title, _ = newFS()
	pos, err = parseFlags(fs, "usage", []string{"--title=X", "a"}, &out)
	if err != nil || !reflect.DeepEqual(pos, []string{"a"}) || *title != "X" {
		t.Fatalf("got %v %q %v", pos, *title, err)
	}

	fs, _, _ = newFS()
	pos, err = parseFlags(fs, "usage", []string{"--", "--x"}, &out)
	if err != nil || !reflect.DeepEqual(pos, []string{"--x"}) {
		t.Fatalf("got %v %v", pos, err)
	}
}

func TestParseFlagsWrongUse(t *testing.T) {
	var out bytes.Buffer
	fs, _, _ := newFS()
	_, err := parseFlags(fs, "usage", []string{"--colour"}, &out)
	if err == nil || err.Error() != "unknown flag --colour" {
		t.Fatalf("got %v", err)
	}
	fs, _, _ = newFS()
	_, err = parseFlags(fs, "usage", []string{"a", "--title"}, &out)
	if err == nil || err.Error() != "flag needs a value: --title" {
		t.Fatalf("got %v", err)
	}
}

func TestParseFlagsHelp(t *testing.T) {
	for _, h := range []string{"-h", "--help"} {
		fs, _, _ := newFS()
		var out bytes.Buffer
		_, err := parseFlags(fs, "the usage\n", []string{"a", h}, &out)
		if !errors.Is(err, errHelp) || out.String() != "the usage\n" {
			t.Fatalf("%s: got %v %q", h, err, out.String())
		}
	}
}
