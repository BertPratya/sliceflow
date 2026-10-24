package slice

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/BertPratya/sliceflow/internal/testrepo"
)

func checkMsg(t *testing.T, r Repo, name, base, source string) string {
	t.Helper()
	err := checkCreate(context.Background(), r, CreateOptions{Name: name}, base, source)
	if err == nil {
		return ""
	}
	se, ok := err.(*Error)
	if !ok {
		t.Fatalf("want *Error, got %T: %v", err, err)
	}
	return se.Message
}

func addFolderOnMain(t *testing.T, r Repo, name string) {
	t.Helper()
	d := filepath.Join(filepath.FromSlash(r.Root), "docs", "slices", name)
	if err := os.MkdirAll(d, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(d, "spec.md"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	testrepo.Git(t, r.Root, "add", ".")
	testrepo.Git(t, r.Root, "commit", "-m", "add folder")
}

func TestCheckExistsOrder(t *testing.T) {
	r := newRepo(t)
	testrepo.Git(t, r.Root, "remote", "add", "team/origin", "https://example.invalid/x.git")
	head := gitOut(t, r.Root, "rev-parse", "HEAD")
	addFolderOnMain(t, r, "x")
	testrepo.Git(t, r.Root, "update-ref", "refs/remotes/team/origin/slice/x", head)
	testrepo.Git(t, r.Root, "branch", "slice/x")

	pre := "slice x already exists: "
	post := ". Nothing was changed."
	if got := checkMsg(t, r, "x", "main", "the default"); got != pre+"branch slice/x"+post {
		t.Errorf("local: %q", got)
	}
	testrepo.Git(t, r.Root, "branch", "-D", "slice/x")
	if got := checkMsg(t, r, "x", "main", "the default"); got != pre+"branch team/origin/slice/x"+post {
		t.Errorf("remote: %q", got)
	}
	testrepo.Git(t, r.Root, "update-ref", "-d", "refs/remotes/team/origin/slice/x")
	if got := checkMsg(t, r, "x", "main", "the default"); got != pre+"folder docs/slices/x on main"+post {
		t.Errorf("folder: %q", got)
	}
}

func TestCheckOtherRemoteBranch(t *testing.T) {
	r := newRepo(t)
	testrepo.Git(t, r.Root, "remote", "add", "origin", "https://example.invalid/x.git")
	head := gitOut(t, r.Root, "rev-parse", "HEAD")
	testrepo.Git(t, r.Root, "update-ref", "refs/remotes/origin/user/slice/x", head)
	if got := checkMsg(t, r, "x", "main", "the default"); got != "" {
		t.Errorf("got %q", got)
	}
}

func TestCheckOrder(t *testing.T) {
	r := newRepo(t)
	bad := `invalid slice name "My Slice": use 1-64 lowercase letters, digits, and dashes, starting with a letter or digit. Nothing was changed.`
	if got := checkMsg(t, r, "My Slice", "nope", "from --from"); got != bad {
		t.Errorf("name before base: %q", got)
	}
	testrepo.Git(t, r.Root, "branch", "slice/x")
	want := "base branch nope does not exist (from --from). Nothing was changed."
	if got := checkMsg(t, r, "x", "nope", "from --from"); got != want {
		t.Errorf("base before exists: %q", got)
	}
}

func TestCheckFolderLast(t *testing.T) {
	r := newRepo(t)
	if err := os.MkdirAll(worktreeDir(r.Root, "x"), 0o755); err != nil {
		t.Fatal(err)
	}
	want := "folder .sliceflow/worktrees/x already exists. Remove it, or pick another name. Nothing was changed."
	if got := checkMsg(t, r, "x", "main", "the default"); got != want {
		t.Errorf("got %q", got)
	}
	testrepo.Git(t, r.Root, "branch", "slice/x")
	if got := checkMsg(t, r, "x", "main", "the default"); got != "slice x already exists: branch slice/x. Nothing was changed." {
		t.Errorf("exists before folder: %q", got)
	}
}
