package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/BertPratya/sliceflow/internal/gittest"
)

// newShop makes the project `shop` of the scenarios: a git repo in the
// folder shop of a new temporary folder, with the branch main and one
// commit. It returns the real path of the project's folder, from
// filepath.EvalSymlinks, as git gives it. The test's own git commands
// run away from the user's settings (gittest.Isolate).
func newShop(t *testing.T) string {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("find the real path of the temporary folder: %v", err)
	}
	return newRepo(t, filepath.Join(root, "shop"))
}

// newRepo makes a git repo in the folder dir, which must not exist yet,
// with the branch main and one commit that adds README.md. It returns
// dir.
func newRepo(t *testing.T, dir string) string {
	t.Helper()
	gittest.Isolate(t)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("make the project folder: %v", err)
	}
	gittest.Git(t, dir, "init", "--quiet", "-b", "main")
	commitFile(t, dir, "README.md", "# shop\n")
	return dir
}

// commitFile writes content to the file path, a path inside the repo
// dir with "/" between its parts, commits it on the branch that is
// checked out, and returns the new commit's full hash.
func commitFile(t *testing.T, dir, path, content string) string {
	t.Helper()
	writeFile(t, filepath.Join(dir, filepath.FromSlash(path)), content)
	gittest.Git(t, dir, "add", "--", path)
	gittest.Git(t, dir, "commit", "--quiet", "-m", "add "+path)
	return gittest.Git(t, dir, "rev-parse", "HEAD")
}

// writeFile writes content to the file at the full path, and makes the
// folders it needs.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("make the folders of %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// readFile returns the text of the file at path.
func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

// create runs `sliceflow slice create` with args in the folder dir, with
// opts, and returns the result and the home folder.
func create(t *testing.T, dir string, opts createOpts, args ...string) (result, string) {
	t.Helper()
	cmd, home := createCmd(t, dir, opts, append([]string{"slice", "create"}, args...)...)
	return runCmd(t, cmd), home
}

// homeWorktree is the worktree of the slice name of the project shop,
// under home, as R17 makes it.
func homeWorktree(home, name string) string {
	return filepath.Join(home, ".sliceflow", "worktrees", "shop", name)
}

// specIn is the full path of spec.md of the slice name in worktree.
func specIn(worktree, name string) string {
	return filepath.Join(worktree, "docs", "slices", name, "spec.md")
}

// createdLines are the lines of R8 for the slice name, from base at the
// commit hash, in worktree, with the Note lines of notes after the first
// line.
func createdLines(name, base, hash, worktree string, notes ...string) string {
	text := "Created slice " + name + ".\n"
	for _, note := range notes {
		text += "  Note:     " + note + "\n"
	}
	return text +
		"  Branch:   slice/" + name + " (from " + base + ", at " + hash[:7] + ")\n" +
		"  Worktree: " + worktree + "\n" +
		"Next: write the spec in " + specIn(worktree, name) + "\n"
}

// checkCreated checks that got is a slice create that worked: exit code
// 0, the lines want on the normal output, and nothing on the error
// output (R8).
func checkCreated(t *testing.T, got result, want string) {
	t.Helper()
	if got.code != 0 {
		t.Errorf("exit code %d, want 0", got.code)
	}
	if got.stdout != want {
		t.Errorf("normal output:\n%s\nwant:\n%s", got.stdout, want)
	}
	if got.stderr != "" {
		t.Errorf("error output:\n%s\nwant nothing", got.stderr)
	}
}

// committedFiles returns the files that the commit at rev changed, one
// per line, as git prints them, with "/".
func committedFiles(t *testing.T, dir, rev string) string {
	t.Helper()
	return gittest.Git(t, dir, "show", "--name-only", "--format=", rev)
}

// The two files of a slice's commit, as git lists them (R2).
func sliceFiles(name string) string {
	return "docs/slices/" + name + "/.events.jsonl\n" +
		"docs/slices/" + name + "/spec.md"
}

// TestCreateSlice checks S1: the slice is made from main, under the home
// folder, with the lines of R8.
func TestCreateSlice(t *testing.T) {
	shop := newShop(t)
	mainCommit := gittest.Git(t, shop, "rev-parse", "main")

	got, home := create(t, shop, createOpts{}, "card-payment",
		"--title", "Pay with a card",
		"--description", "Let a customer pay with a credit or debit card.")

	worktree := homeWorktree(home, "card-payment")
	checkCreated(t, got, createdLines("card-payment", "main", mainCommit, worktree))

	// R1, R11: the branch is one commit on top of main.
	if parent := gittest.Git(t, shop, "rev-parse", "slice/card-payment^"); parent != mainCommit {
		t.Errorf("the slice's commit has the parent %s, want main at %s", parent, mainCommit)
	}
	// R1, R17: the worktree is there, on the branch.
	if branch := gittest.Git(t, worktree, "branch", "--show-current"); branch != "slice/card-payment" {
		t.Errorf("the worktree is on %q, want slice/card-payment", branch)
	}
	// R2: exactly the two files, and the lock file is ignored.
	if files := committedFiles(t, shop, "slice/card-payment"); files != sliceFiles("card-payment") {
		t.Errorf("the commit holds:\n%s\nwant:\n%s", files, sliceFiles("card-payment"))
	}
	exclude := readFile(t, filepath.Join(shop, ".git", "info", "exclude"))
	if !slices.Contains(strings.Split(exclude, "\n"), "*.events.jsonl.lock") {
		t.Errorf("info/exclude has no line *.events.jsonl.lock:\n%s", exclude)
	}
	if status := gittest.Git(t, worktree, "status", "--porcelain"); status != "" {
		t.Errorf("git status in the worktree:\n%s\nwant nothing", status)
	}
}

// TestCreateLowercase checks S5: a name with capitals is made lowercase,
// with a Note line (R21, R22).
func TestCreateLowercase(t *testing.T) {
	shop := newShop(t)
	mainCommit := gittest.Git(t, shop, "rev-parse", "main")

	got, home := create(t, shop, createOpts{}, "Card-Payment")

	checkCreated(t, got, createdLines("card-payment", "main", mainCommit,
		homeWorktree(home, "card-payment"), `"Card-Payment" was made lowercase.`))
	gittest.Git(t, shop, "rev-parse", "--verify", "--quiet", "refs/heads/slice/card-payment")
}

// TestCreateFrom checks S9: --from starts the slice at the tip of that
// branch (R10, R11).
func TestCreateFrom(t *testing.T) {
	shop := newShop(t)
	gittest.Git(t, shop, "switch", "--quiet", "-c", "release")
	releaseCommit := commitFile(t, shop, "release.txt", "1.0\n")
	gittest.Git(t, shop, "switch", "--quiet", "main")

	got, home := create(t, shop, createOpts{}, "card-payment", "--from", "release")

	checkCreated(t, got, createdLines("card-payment", "release", releaseCommit, homeWorktree(home, "card-payment")))
	if parent := gittest.Git(t, shop, "rev-parse", "slice/card-payment^"); parent != releaseCommit {
		t.Errorf("the slice's commit has the parent %s, want release at %s", parent, releaseCommit)
	}
}

// eventAt finds the "at" of an event in UTC, as R27 writes it, with the
// time in its group 1.
var eventAt = regexp.MustCompile(`"at":"(\d{4}-\d\d-\d\dT\d\d:\d\d:\d\dZ)"`)

// checkFirstEvent checks that the log of the slice name in worktree is
// one line, the same as in the slice branch, and is want once its "at"
// is set to the time in the line. The time must be in UTC, as in R27,
// and between before and after, which are truncated to the second.
func checkFirstEvent(t *testing.T, worktree, name, want string, before, after time.Time) {
	t.Helper()
	log := readFile(t, filepath.Join(worktree, "docs", "slices", name, ".events.jsonl"))
	if strings.Count(log, "\n") != 1 || !strings.HasSuffix(log, "\n") {
		t.Fatalf("the log is not exactly one line:\n%s", log)
	}
	line := strings.TrimSuffix(log, "\n")
	match := eventAt.FindStringSubmatch(line)
	if match == nil {
		t.Fatalf("the log line has no UTC time in \"at\":\n%s", line)
	}
	at, err := time.Parse(time.RFC3339, match[1])
	if err != nil {
		t.Fatalf("read the time %q: %v", match[1], err)
	}
	if at.Before(before) || at.After(after) {
		t.Errorf("the time %s is not between %s and %s", match[1], before.Format(time.RFC3339), after.Format(time.RFC3339))
	}
	want = strings.Replace(want, "<at>", match[1], 1)
	if line != want {
		t.Errorf("the log line:\n%s\nwant:\n%s", line, want)
	}
	inBranch := gittest.Git(t, worktree, "show", "slice/"+name+":docs/slices/"+name+"/.events.jsonl")
	if inBranch != line {
		t.Errorf("the log in the branch:\n%s\nwant the same line as the file:\n%s", inBranch, line)
	}
}

// TestCreateFirstEvent checks S13: the first event, with the title and
// the description (R27).
func TestCreateFirstEvent(t *testing.T) {
	shop := newShop(t)
	mainCommit := gittest.Git(t, shop, "rev-parse", "main")

	before := time.Now().UTC().Truncate(time.Second)
	got, home := create(t, shop, createOpts{}, "card-payment",
		"--title", "Pay with a card",
		"--description", "Let a customer pay with a credit or debit card.")
	after := time.Now().UTC().Truncate(time.Second)
	if got.code != 0 {
		t.Fatalf("exit code %d, error output:\n%s", got.code, got.stderr)
	}

	checkFirstEvent(t, homeWorktree(home, "card-payment"), "card-payment",
		`{"version":1,"seq":1,"event":"slice.created","at":"<at>","data":{"name":"card-payment","title":"Pay with a card","description":"Let a customer pay with a credit or debit card.","base":"main","base_commit":"`+mainCommit+`"}}`, before, after)
}

// TestCreateCommit checks S14: one commit with its message and the two
// files, and the user's checkout is not changed (R2, R4, R7).
func TestCreateCommit(t *testing.T) {
	shop := newShop(t)
	mainCommit := gittest.Git(t, shop, "rev-parse", "main")

	got, _ := create(t, shop, createOpts{}, "card-payment")
	if got.code != 0 {
		t.Fatalf("exit code %d, error output:\n%s", got.code, got.stderr)
	}

	want := "card-payment: add the spec and the event log\n\n" + sliceFiles("card-payment")
	if show := gittest.Git(t, shop, "show", "--name-only", "--format=%s", "slice/card-payment"); show != want {
		t.Errorf("git show:\n%s\nwant:\n%s", show, want)
	}
	if count := gittest.Git(t, shop, "rev-list", "--count", "main..slice/card-payment"); count != "1" {
		t.Errorf("the slice branch has %s commits on top of main, want 1", count)
	}

	// R7: the user's own checkout.
	if head := gittest.Git(t, shop, "rev-parse", "HEAD"); head != mainCommit {
		t.Errorf("HEAD is %s, want %s", head, mainCommit)
	}
	if branch := gittest.Git(t, shop, "branch", "--show-current"); branch != "main" {
		t.Errorf("the checkout is on %q, want main", branch)
	}
	if status := gittest.Git(t, shop, "status", "--porcelain"); status != "" {
		t.Errorf("git status in the project:\n%s\nwant nothing", status)
	}
}

// TestCreateTitle checks S15: with no --title and no --description, the
// title is the name and the description is empty (R27, R28).
func TestCreateTitle(t *testing.T) {
	shop := newShop(t)
	mainCommit := gittest.Git(t, shop, "rev-parse", "main")

	before := time.Now().UTC().Truncate(time.Second)
	got, home := create(t, shop, createOpts{}, "card-payment")
	after := time.Now().UTC().Truncate(time.Second)
	if got.code != 0 {
		t.Fatalf("exit code %d, error output:\n%s", got.code, got.stderr)
	}

	checkFirstEvent(t, homeWorktree(home, "card-payment"), "card-payment",
		`{"version":1,"seq":1,"event":"slice.created","at":"<at>","data":{"name":"card-payment","title":"card-payment","description":"","base":"main","base_commit":"`+mainCommit+`"}}`, before, after)
}

// TestCreateFromSubfolder checks S16: run from shop/src/api, the
// worktree is under the name shop, as in S1 (R1, R16).
func TestCreateFromSubfolder(t *testing.T) {
	shop := newShop(t)
	mainCommit := gittest.Git(t, shop, "rev-parse", "main")
	api := filepath.Join(shop, "src", "api")
	if err := os.MkdirAll(api, 0o755); err != nil {
		t.Fatalf("make %s: %v", api, err)
	}

	got, home := create(t, api, createOpts{}, "card-payment")

	checkCreated(t, got, createdLines("card-payment", "main", mainCommit, homeWorktree(home, "card-payment")))
}

// TestCreateSpecFile checks S18: spec.md is the built-in template as it
// is, with no header (R2, R3). The template is read from the repo, not
// from the program, so a change in how it is built in is caught.
func TestCreateSpecFile(t *testing.T) {
	shop := newShop(t)

	got, home := create(t, shop, createOpts{}, "card-payment")
	if got.code != 0 {
		t.Fatalf("exit code %d, error output:\n%s", got.code, got.stderr)
	}

	want := readFile(t, filepath.Join("..", "templates", "default", "spec.md"))
	if spec := readFile(t, specIn(homeWorktree(home, "card-payment"), "card-payment")); spec != want {
		t.Errorf("spec.md:\n%s\nwant the template:\n%s", spec, want)
	}
}

// TestCreateWorktreeDir checks S19: sliceflow.worktree.dir, relative to
// the project, puts the worktree next to it (R18).
func TestCreateWorktreeDir(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("find the real path of the temporary folder: %v", err)
	}
	shop := newRepo(t, filepath.Join(root, "projects", "shop"))
	mainCommit := gittest.Git(t, shop, "rev-parse", "main")

	got, _ := create(t, shop, createOpts{Config: map[string]string{"sliceflow.worktree.dir": "../shop-worktrees"}}, "card-payment")

	worktree := filepath.Join(root, "projects", "shop-worktrees", "shop", "card-payment")
	checkCreated(t, got, createdLines("card-payment", "main", mainCommit, worktree))
	if _, err := os.Stat(specIn(worktree, "card-payment")); err != nil {
		t.Errorf("spec.md is not in the worktree: %v", err)
	}
}

// TestCreateIgnored checks S23: a .gitignore rule that matches the log
// gives a Note line, and the log is in the commit anyway (R5, R8).
func TestCreateIgnored(t *testing.T) {
	shop := newShop(t)
	mainCommit := commitFile(t, shop, ".gitignore", "*.jsonl\n")

	got, home := create(t, shop, createOpts{}, "card-payment")

	checkCreated(t, got, createdLines("card-payment", "main", mainCommit, homeWorktree(home, "card-payment"),
		".events.jsonl matches a .gitignore rule, so sliceflow added it anyway."))
	if files := committedFiles(t, shop, "slice/card-payment"); files != sliceFiles("card-payment") {
		t.Errorf("the commit holds:\n%s\nwant:\n%s", files, sliceFiles("card-payment"))
	}
}

// TestCreateIgnoredFilesStayTracked checks R2 and R5: both files are
// added even when ignored, then stay tracked, so a plain git add works
// on them, while the .gitignore rules still work for other files.
func TestCreateIgnoredFilesStayTracked(t *testing.T) {
	shop := newShop(t)
	mainCommit := commitFile(t, shop, ".gitignore", "spec.md\n*.jsonl\n")

	got, home := create(t, shop, createOpts{}, "login")
	worktree := homeWorktree(home, "login")
	checkCreated(t, got, createdLines("login", "main", mainCommit, worktree,
		"spec.md matches a .gitignore rule, so sliceflow added it anyway.",
		".events.jsonl matches a .gitignore rule, so sliceflow added it anyway."))
	if files := committedFiles(t, shop, "slice/login"); files != sliceFiles("login") {
		t.Fatalf("the commit holds:\n%s\nwant:\n%s", files, sliceFiles("login"))
	}

	folder := filepath.Join(worktree, "docs", "slices", "login")
	for _, file := range []string{"spec.md", ".events.jsonl"} {
		path := filepath.Join(folder, file)
		writeFile(t, path, readFile(t, path)+"one more line\n")
	}
	gittest.Git(t, worktree, "add", "docs/slices/login/spec.md", "docs/slices/login/.events.jsonl")
	gittest.Git(t, worktree, "commit", "--quiet", "-m", "edit")
	if files := committedFiles(t, worktree, "HEAD"); files != sliceFiles("login") {
		t.Errorf("the edit commit holds:\n%s\nwant:\n%s", files, sliceFiles("login"))
	}
	// SD31: the lock file is ignored, so nothing is left to add.
	if status := gittest.Git(t, worktree, "status", "--porcelain"); status != "" {
		t.Errorf("git status in the worktree:\n%s\nwant nothing", status)
	}

	// The rules are active: a new file that matches one is refused.
	writeFile(t, filepath.Join(folder, "notes.jsonl"), "{}\n")
	add := exec.Command("git", "add", "docs/slices/login/notes.jsonl")
	add.Dir = worktree
	if out, err := add.CombinedOutput(); err == nil {
		t.Errorf("git add of an ignored new file worked, want it refused:\n%s", out)
	}
}

// TestCreatePaths checks R9: the printed paths use the separator of the
// system, and the paths in the commit use "/". It is the Windows check
// of R9: only there do the two separators differ.
func TestCreatePaths(t *testing.T) {
	shop := newShop(t)

	got, home := create(t, shop, createOpts{}, "login")
	if got.code != 0 {
		t.Fatalf("exit code %d, error output:\n%s", got.code, got.stderr)
	}

	worktree := homeWorktree(home, "login")
	for _, want := range []string{"  Worktree: " + worktree + "\n", "Next: write the spec in " + specIn(worktree, "login") + "\n"} {
		if !strings.Contains(got.stdout, want) {
			t.Errorf("normal output:\n%s\nwant the line %q", got.stdout, want)
		}
	}
	if runtime.GOOS == "windows" {
		for _, line := range strings.Split(got.stdout, "\n") {
			if (strings.HasPrefix(line, "  Worktree: ") || strings.HasPrefix(line, "Next: ")) && strings.Contains(line, "/") {
				t.Errorf("a printed path has a /: %q", line)
			}
		}
	}
	if files := committedFiles(t, shop, "slice/login"); files != sliceFiles("login") {
		t.Errorf("the commit holds:\n%s\nwant:\n%s", files, sliceFiles("login"))
	}
}

// TestCreateEmptyFromE2E checks R74 and SD29: an empty --from is an
// option with no value. It changes nothing, and exits with code 2.
func TestCreateEmptyFromE2E(t *testing.T) {
	want := result{
		stderr: "Error: option \"--from\" needs a value. Nothing was changed.\n" + createOptions,
		code:   2,
	}
	tests := []struct {
		name string
		args []string
	}{
		{"--from=", []string{"slice", "create", "x", "--from="}},
		{`--from ""`, []string{"slice", "create", "x", "--from", ""}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check(t, tt.args, want)
		})
	}
}
