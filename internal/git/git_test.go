package git_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BertPratya/sliceflow/internal/git"
	"github.com/BertPratya/sliceflow/internal/testrepo"
)

func TestRunReturnsOutput(t *testing.T) {
	dir := testrepo.New(t)
	out, err := git.Run(context.Background(), dir, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	if out != "main\n" {
		t.Fatalf("got %q", out)
	}
}

func TestRunErrorHasStderr(t *testing.T) {
	dir := testrepo.New(t)
	_, err := git.Run(context.Background(), dir, "rev-parse", "--verify", "nope")
	var ge *git.Error
	if !errors.As(err, &ge) {
		t.Fatalf("want *git.Error, got %T %v", err, err)
	}
	if ge.ExitCode != 128 {
		t.Fatalf("exit code %d", ge.ExitCode)
	}
	if ge.Stderr == "" || ge.Stderr != strings.TrimSpace(ge.Stderr) {
		t.Fatalf("bad stderr %q", ge.Stderr)
	}
	if ge.Error() != ge.Stderr {
		t.Fatalf("Error() = %q", ge.Error())
	}
}

func TestRunSetsNoNetworkEnv(t *testing.T) {
	dir := testrepo.New(t)
	testrepo.Git(t, dir, "config", "alias.env-check", "!env")
	out, err := git.Run(context.Background(), dir, "env-check")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.ReplaceAll(out, "\r\n", "\n"), "\n")
	has := func(want string) bool {
		for _, l := range lines {
			if l == want {
				return true
			}
		}
		return false
	}
	if !has("GIT_TERMINAL_PROMPT=0") || !has("GIT_NO_LAZY_FETCH=1") {
		t.Fatalf("env missing:\n%s", out)
	}
}

func TestRunCancelled(t *testing.T) {
	dir := testrepo.New(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := git.Run(ctx, dir, "status"); !errors.Is(err, context.Canceled) {
		t.Fatalf("Run err = %v", err)
	}
	ok, err := git.RefExists(ctx, dir, "refs/heads/main")
	if ok || !errors.Is(err, context.Canceled) {
		t.Fatalf("RefExists = %v, %v", ok, err)
	}
}

func TestRefExists(t *testing.T) {
	dir := testrepo.New(t)
	ctx := context.Background()
	ok, err := git.RefExists(ctx, dir, "refs/heads/main")
	if err != nil || !ok {
		t.Fatalf("main: %v %v", ok, err)
	}
	ok, err = git.RefExists(ctx, dir, "refs/heads/slice/none")
	if err != nil || ok {
		t.Fatalf("none: %v %v", ok, err)
	}
}

func TestShowMissingFile(t *testing.T) {
	dir := testrepo.New(t)
	ctx := context.Background()
	b, ok, err := git.Show(ctx, dir, "main", "docs/none.md")
	if b != nil || ok || err != nil {
		t.Fatalf("missing: %v %v %v", b, ok, err)
	}
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hi\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	testrepo.Git(t, dir, "add", "a.txt")
	testrepo.Git(t, dir, "commit", "-m", "add a")
	b, ok, err = git.Show(ctx, dir, "main", "a.txt")
	if err != nil || !ok || string(b) != "hi\n" {
		t.Fatalf("present: %q %v %v", b, ok, err)
	}
}
