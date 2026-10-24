package slice

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/BertPratya/sliceflow/internal/eventlog"
	"github.com/BertPratya/sliceflow/internal/testrepo"
)

var fixedNow = time.Date(2026, 10, 24, 7, 42, 0, 0, time.UTC)

func newRepo(t *testing.T) Repo {
	t.Helper()
	dir := testrepo.New(t)
	r, err := FindRepo(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	r.Now = func() time.Time { return fixedNow }
	return r
}

func gitOut(t *testing.T, dir string, args ...string) string {
	t.Helper()
	return strings.TrimSpace(testrepo.Git(t, dir, args...))
}

func sameFile(t *testing.T, a, b string) bool {
	t.Helper()
	ea, err := filepath.EvalSymlinks(filepath.FromSlash(a))
	if err != nil {
		t.Fatal(err)
	}
	eb, err := filepath.EvalSymlinks(filepath.FromSlash(b))
	if err != nil {
		t.Fatal(err)
	}
	return ea == eb
}

func TestCreateMakesSlice(t *testing.T) {
	r := newRepo(t)
	mainCommit := gitOut(t, r.Dir, "rev-parse", "main")
	got, err := Create(context.Background(), r, CreateOptions{Name: "spec-submit"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Slice != "spec-submit" || got.Branch != "slice/spec-submit" {
		t.Errorf("got %+v", got)
	}
	if got.From != (eventlog.From{Branch: "main", Commit: mainCommit}) {
		t.Errorf("From = %+v", got.From)
	}
	if got.Spec != "docs/slices/spec-submit/spec.md" {
		t.Errorf("Spec = %q", got.Spec)
	}
	wantWT := r.Root + "/.sliceflow/worktrees/spec-submit"
	if strings.Contains(got.Worktree, `\`) || !sameFile(t, got.Worktree, wantWT) {
		t.Errorf("Worktree = %q, want %q", got.Worktree, wantWT)
	}
	if gitOut(t, r.Dir, "rev-parse", "slice/spec-submit~1") != mainCommit {
		t.Error("branch is not main plus one commit")
	}
	if msg := gitOut(t, r.Dir, "log", "-1", "--format=%s", "slice/spec-submit"); msg != "spec-submit: slice created" {
		t.Errorf("message = %q", msg)
	}
	files := gitOut(t, r.Dir, "diff", "--name-only", "main", "slice/spec-submit")
	if files != "docs/slices/spec-submit/.events.jsonl\ndocs/slices/spec-submit/spec.md" {
		t.Errorf("files = %q", files)
	}
	log := gitOut(t, r.Dir, "show", "slice/spec-submit:docs/slices/spec-submit/.events.jsonl")
	enc, _ := eventlog.Encode(eventlog.NewCreated(fixedNow, eventlog.From{Branch: "main", Commit: mainCommit}, ""))
	if log+"\n" != string(enc) {
		t.Errorf("log = %q, want %q", log, enc)
	}
	spec := gitOut(t, r.Dir, "show", "slice/spec-submit:docs/slices/spec-submit/spec.md")
	if !strings.Contains(spec, "slice: spec-submit\ncreated: 2026-10-24\nideas: []\n") {
		t.Errorf("spec front matter wrong:\n%s", spec[:min(len(spec), 200)])
	}
	if s := gitOut(t, r.Dir, "status", "--short"); s != "" {
		t.Errorf("main status = %q", s)
	}
	if _, err := os.Stat(filepath.Join(filepath.FromSlash(got.Worktree), "docs", "slices", "spec-submit", "spec.md")); err != nil {
		t.Error(err)
	}
}

func TestCreateWritesGitignore(t *testing.T) {
	r := newRepo(t)
	if _, err := Create(context.Background(), r, CreateOptions{Name: "one"}); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(filepath.FromSlash(r.Root), ".sliceflow", ".gitignore")
	b, err := os.ReadFile(p)
	if err != nil || string(b) != "*\n" {
		t.Errorf("gitignore = %q, %v", b, err)
	}

	r2 := newRepo(t)
	dir := filepath.Join(filepath.FromSlash(r2.Root), ".sliceflow")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	p2 := filepath.Join(dir, ".gitignore")
	if err := os.WriteFile(p2, []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Create(context.Background(), r2, CreateOptions{Name: "two"}); err != nil {
		t.Fatal(err)
	}
	b, _ = os.ReadFile(p2)
	if string(b) != "x\n" {
		t.Errorf("existing gitignore changed: %q", b)
	}
}

func TestCreateTitle(t *testing.T) {
	r := newRepo(t)
	if _, err := Create(context.Background(), r, CreateOptions{Name: "with", Title: "Login with Google"}); err != nil {
		t.Fatal(err)
	}
	log := gitOut(t, r.Dir, "show", "slice/with:docs/slices/with/.events.jsonl")
	if !strings.Contains(log, `"title":"Login with Google"`) {
		t.Errorf("log = %s", log)
	}
	if _, err := Create(context.Background(), r, CreateOptions{Name: "without"}); err != nil {
		t.Fatal(err)
	}
	log = gitOut(t, r.Dir, "show", "slice/without:docs/slices/without/.events.jsonl")
	if strings.Contains(log, "title") {
		t.Errorf("log = %s", log)
	}
}

func TestCreateHookRuns(t *testing.T) {
	r := newRepo(t)
	hooks := gitOut(t, r.Dir, "rev-parse", "--git-path", "hooks")
	if !filepath.IsAbs(hooks) {
		hooks = filepath.Join(r.Dir, hooks)
	}
	if err := os.MkdirAll(hooks, 0o755); err != nil {
		t.Fatal(err)
	}
	gitDir := gitOut(t, r.Dir, "rev-parse", "--absolute-git-dir")
	script := "#!/bin/sh\necho x >> \"" + filepath.ToSlash(filepath.Join(gitDir, "ran")) + "\"\n"
	if err := os.WriteFile(filepath.Join(hooks, "pre-commit"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := Create(context.Background(), r, CreateOptions{Name: "hooked"}); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(gitDir, "ran"))
	if err != nil {
		t.Fatalf("hook did not run: %v", err)
	}
	if string(b) != "x\n" {
		t.Errorf("hook ran %q", b)
	}
}

func TestCreateIgnoredFiles(t *testing.T) {
	r := newRepo(t)
	p := filepath.Join(filepath.FromSlash(r.Root), ".gitignore")
	if err := os.WriteFile(p, []byte("*.jsonl\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	testrepo.Git(t, r.Dir, "add", ".gitignore")
	testrepo.Git(t, r.Dir, "commit", "-m", "ignore")
	if _, err := Create(context.Background(), r, CreateOptions{Name: "ign"}); err != nil {
		t.Fatal(err)
	}
	files := gitOut(t, r.Dir, "diff", "--name-only", "main", "slice/ign")
	if files != "docs/slices/ign/.events.jsonl\ndocs/slices/ign/spec.md" {
		t.Errorf("files = %q", files)
	}
}

func TestWorktreeDirNetworkShare(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("a network share root only exists on Windows")
	}
	got := filepath.ToSlash(worktreeDir("//server/share/repo", "a"))
	want := "//server/share/repo/.sliceflow/worktrees/a"
	if got != want {
		t.Errorf("worktreeDir = %q, want %q", got, want)
	}
}
