package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/BertPratya/sliceflow/internal/gittest"
)

// validSpec returns the lines of a spec of the slice name with no
// problems: four scenarios that each show R1, and three requirements.
// The summary text is on lines 10-12.
func validSpec(name string) []string {
	return []string{
		"---",
		"slice: " + name,
		"created: 2026-10-29",
		"---",
		"",
		"# Spec",
		"",
		"## Summary",
		"",
		"A user submits a spec for review, so that a reviewer can check it.",
		"The spec is read from the slice's worktree.",
		"Later steps read the spec as data.",
		"",
		"## User stories",
		"",
		"### Story 1: Submit a spec",
		"",
		"A user submits a spec before anyone builds from it.",
		"",
		"- **S1.** **When** I run `sliceflow spec submit`,",
		"  **Then** the spec is sent.",
		"  Shows: R1",
		"- **S2.** **When** I run it twice,",
		"  **Then** the spec is sent once.",
		"  Shows: R1",
		"- **S3.** **When** I run it with no spec,",
		"  **Then** sliceflow prints an error.",
		"  Shows: R1",
		"- **S4.** **When** I run it on main,",
		"  **Then** sliceflow prints an error.",
		"  Shows: R1",
		"",
		"## Requirements",
		"",
		"- **R1.** sliceflow sends the spec.",
		"  Check: e2e",
		"- **R2.** sliceflow sends the spec once.",
		"  Check: e2e",
		"- **R3.** sliceflow prints an error when there is no spec.",
		"  Check: e2e",
		"",
		"## Out of scope",
		"",
		"None.",
		"",
		"## Decisions",
		"",
		"None.",
		"",
		"## Open questions",
		"",
		"None.",
	}
}

// specS2 returns the spec of S2: R3 has no Check: line, and S4 shows R9,
// which does not exist.
func specS2(t *testing.T, name string) []string {
	t.Helper()
	lines := validSpec(name)
	r3 := lineOf(t, lines, "- **R3.** sliceflow prints an error when there is no spec.")
	lines = slices.Delete(lines, r3, r3+1) // the Check: line after R3
	s4 := lineOf(t, lines, "- **S4.** **When** I run it on main,")
	lines[s4+1] = "  Shows: R9" // the Shows: line of S4
	return lines
}

// lineOf returns the line number (from 1) of the first line that is text.
func lineOf(t *testing.T, lines []string, text string) int {
	t.Helper()
	i := slices.Index(lines, text)
	if i < 0 {
		t.Fatalf("no line %q in the spec", text)
	}
	return i + 1
}

// specFile returns the path of the slice's spec in its worktree.
func specFile(worktree, name string) string {
	return filepath.Join(worktree, "docs", "slices", name, "spec.md")
}

// writeSpec writes lines as the spec of the slice name in its worktree. It
// is not committed.
func writeSpec(t *testing.T, worktree, name string, lines []string) {
	t.Helper()
	content := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(specFile(worktree, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// s2Output returns what S2 prints, with path as the spec's path, for the
// spec of specS2.
func s2Output(t *testing.T, path string) string {
	t.Helper()
	lines := specS2(t, "spec-submit")
	s4 := lineOf(t, lines, "  Shows: R9")
	r3 := lineOf(t, lines, "- **R3.** sliceflow prints an error when there is no spec.")
	return fmt.Sprintf("%s:%d: error: S4 shows R9, but there is no R9\n", path, s4) +
		"  hint: name only requirements that exist, e.g. \"Shows: R1, R2\".\n" +
		fmt.Sprintf("%s:%d: error: R3 has no Check: line\n", path, r3) +
		"  hint: add a line \"Check: e2e\" under R3. The kinds are unit, api, e2e, and manual.\n" +
		"2 errors.\n"
}

// wantRun checks the exit code and both outputs of a run.
func wantRun(t *testing.T, r result, code int, stdout, stderr string) {
	t.Helper()
	if r.code != code || r.stdout != stdout || r.stderr != stderr {
		t.Errorf("code = %d, stdout = %q, stderr = %q\nwant %d, %q, %q", r.code, r.stdout, r.stderr, code, stdout, stderr)
	}
}

// mainSpecPath is the path of spec-submit's spec, from the main checkout.
const mainSpecPath = ".sliceflow/worktrees/spec-submit/docs/slices/spec-submit/spec.md"

// TestValidateS1: a spec with no problems, not committed (0002 R1, R3).
func TestValidateS1(t *testing.T) {
	repo := gittest.NewRepo(t)
	worktree := createSlice(t, repo, "spec-submit")
	writeSpec(t, worktree, "spec-submit", validSpec("spec-submit"))

	r := run(t, repo, "spec", "validate", "spec-submit")
	wantRun(t, r, 0, "No problems in "+mainSpecPath+".\n", "")
}

// TestValidateS2: two problems, each with its line and hint, then the
// count (R2, R3). The lines are where specS2 puts them, not 48 and 61.
// The spec is not committed, so this also shows that sliceflow reads the
// file on disk (R1).
func TestValidateS2(t *testing.T) {
	repo := gittest.NewRepo(t)
	worktree := createSlice(t, repo, "spec-submit")
	writeSpec(t, worktree, "spec-submit", specS2(t, "spec-submit"))
	if got := gittest.Run(t, worktree, "status", "--porcelain"); got != "M docs/slices/spec-submit/spec.md" {
		t.Fatalf("git status = %q, want spec.md changed and not committed", got)
	}

	r := run(t, repo, "spec", "validate", "spec-submit")
	wantRun(t, r, 1, s2Output(t, mainSpecPath), "")
}

// TestValidateS3: one open question on line 210 gives "1 error." (R3,
// R22).
func TestValidateS3(t *testing.T) {
	repo := gittest.NewRepo(t)
	worktree := createSlice(t, repo, "spec-submit")
	lines := validSpec("spec-submit")
	lines[len(lines)-1] = "- **Q1.** Who reviews the spec?"
	// Fill the summary so that Q1 is on line 210.
	filler := make([]string, 210-len(lines))
	for i := range filler {
		filler[i] = "More text."
	}
	lines = slices.Insert(lines, 12, filler...)
	if got := lineOf(t, lines, "- **Q1.** Who reviews the spec?"); got != 210 {
		t.Fatalf("Q1 is on line %d, want 210", got)
	}
	writeSpec(t, worktree, "spec-submit", lines)

	r := run(t, repo, "spec", "validate", "spec-submit")
	want := mainSpecPath + ":210: error: Q1 is not answered\n" +
		"  hint: write the answer into the spec, and remove Q1. When no question is left, write None.\n" +
		"1 error.\n"
	wantRun(t, r, 1, want, "")
}

// TestValidateS4: with no name, inside the slice's worktree, the path is
// relative to the worktree (R1, R2).
func TestValidateS4(t *testing.T) {
	repo := gittest.NewRepo(t)
	worktree := createSlice(t, repo, "spec-submit")
	writeSpec(t, worktree, "spec-submit", specS2(t, "spec-submit"))

	r := run(t, worktree, "spec", "validate")
	wantRun(t, r, 1, s2Output(t, "docs/slices/spec-submit/spec.md"), "")
}

// TestValidateFromSubfolder: the path is relative to the current folder,
// not to the repo's root (R2).
func TestValidateFromSubfolder(t *testing.T) {
	repo := gittest.NewRepo(t)
	worktree := createSlice(t, repo, "spec-submit")
	writeSpec(t, worktree, "spec-submit", validSpec("spec-submit"))
	sub := filepath.Join(repo, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	r := run(t, sub, "spec", "validate", "spec-submit")
	wantRun(t, r, 0, "No problems in ../"+mainSpecPath+".\n", "")
}

// TestValidateS5: a placeholder left on line 12 is cut after 40
// characters (R14).
func TestValidateS5(t *testing.T) {
	repo := gittest.NewRepo(t)
	worktree := createSlice(t, repo, "spec-submit")
	lines := validSpec("spec-submit")
	lines[11] = "{{Two or three sentences: what this feature lets someone do, and why.}}"
	writeSpec(t, worktree, "spec-submit", lines)

	r := run(t, repo, "spec", "validate", "spec-submit")
	want := mainSpecPath + ":12: error: placeholder left from the template: {{Two or three sentences: what this feat…\n" +
		"  hint: replace it with real text.\n"
	if r.code != 1 || r.stderr != "" || !strings.HasPrefix(r.stdout, want) {
		t.Errorf("code = %d, stderr = %q, stdout = %q\nwant 1, empty, starting with %q", r.code, r.stderr, r.stdout, want)
	}
}

// TestValidateS6: the problems of S2 as one JSON object, with the full
// path of the file (R4).
func TestValidateS6(t *testing.T) {
	repo := gittest.NewRepo(t)
	worktree := createSlice(t, repo, "spec-submit")
	lines := specS2(t, "spec-submit")
	writeSpec(t, worktree, "spec-submit", lines)

	r := run(t, repo, "spec", "validate", "spec-submit", "--json")

	// The file is the full path, with links resolved (macOS /tmp).
	resolvedWorktree, err := filepath.EvalSymlinks(worktree)
	if err != nil {
		t.Fatal(err)
	}
	file, err := json.Marshal(specFile(resolvedWorktree, "spec-submit"))
	if err != nil {
		t.Fatal(err)
	}
	want := `{"slice":"spec-submit","file":` + string(file) + `,"problems":[` +
		fmt.Sprintf(`{"line":%d,"level":"error","message":"S4 shows R9, but there is no R9",`, lineOf(t, lines, "  Shows: R9")) +
		`"hint":"name only requirements that exist, e.g. \"Shows: R1, R2\"."},` +
		fmt.Sprintf(`{"line":%d,"level":"error","message":"R3 has no Check: line",`, lineOf(t, lines, "- **R3.** sliceflow prints an error when there is no spec.")) +
		`"hint":"add a line \"Check: e2e\" under R3. The kinds are unit, api, e2e, and manual."}],"errors":2}` + "\n"
	wantRun(t, r, 1, want, "")
}

// TestValidateJSONNoProblems: with no problems, the list is [] and the
// exit code is 0 (R3, R4).
func TestValidateJSONNoProblems(t *testing.T) {
	repo := gittest.NewRepo(t)
	worktree := createSlice(t, repo, "spec-submit")
	writeSpec(t, worktree, "spec-submit", validSpec("spec-submit"))

	r := run(t, repo, "spec", "validate", "spec-submit", "--json")
	if r.code != 0 || r.stderr != "" || !strings.HasSuffix(r.stdout, `"problems":[],"errors":0}`+"\n") {
		t.Errorf("code = %d, stderr = %q, stdout = %q, want 0, empty, no problems", r.code, r.stderr, r.stdout)
	}
}

// TestValidateNoLine: a problem with no line prints without ":<line>",
// and in JSON without "line" (R2, R4).
func TestValidateNoLine(t *testing.T) {
	repo := gittest.NewRepo(t)
	worktree := createSlice(t, repo, "spec-submit")
	lines := validSpec("spec-submit")
	decisions := lineOf(t, lines, "## Decisions")
	lines = slices.Delete(lines, decisions-1, decisions+3) // the heading, its blank line, None., and the blank line after
	writeSpec(t, worktree, "spec-submit", lines)

	r := run(t, repo, "spec", "validate", "spec-submit")
	want := mainSpecPath + ": error: missing section \"## Decisions\"\n" +
		"  hint: add it after \"## Out of scope\". With nothing to say, write None.\n" +
		"1 error.\n"
	wantRun(t, r, 1, want, "")

	r = run(t, repo, "spec", "validate", "spec-submit", "--json")
	wantProblems := `"problems":[{"level":"error","message":"missing section \"## Decisions\"",` +
		`"hint":"add it after \"## Out of scope\". With nothing to say, write None."}],"errors":1}` + "\n"
	if r.code != 1 || !strings.HasSuffix(r.stdout, wantProblems) {
		t.Errorf("code = %d, stdout = %q, want 1, ending with %q", r.code, r.stdout, wantProblems)
	}
}

// TestValidateControlCharacter: a control character from the spec does
// not reach the terminal.
func TestValidateControlCharacter(t *testing.T) {
	repo := gittest.NewRepo(t)
	worktree := createSlice(t, repo, "spec-submit")
	lines := validSpec("spec-submit")
	lines[11] = "{{\x1b[31mred}}"
	writeSpec(t, worktree, "spec-submit", lines)

	r := run(t, repo, "spec", "validate", "spec-submit")
	want := mainSpecPath + ":12: error: placeholder left from the template: {{\\u001b[31mred}}\n"
	if r.code != 1 || !strings.HasPrefix(r.stdout, want) || strings.Contains(r.stdout, "\x1b") {
		t.Errorf("code = %d, stdout = %q, want 1, starting with %q", r.code, r.stdout, want)
	}
}

// TestValidateS7: no name on main gives R5's three lines.
func TestValidateS7(t *testing.T) {
	repo := gittest.NewRepo(t)

	r := run(t, repo, "spec", "validate")
	want := "Error: not on a slice branch (you are on main).\n" +
		"Give a name: sliceflow spec validate <name>\n" +
		"See all slices: sliceflow slice list\n"
	wantRun(t, r, 1, "", want)
}

// TestValidateNoBranch: no name on a detached HEAD (R5).
func TestValidateNoBranch(t *testing.T) {
	repo := gittest.NewRepo(t)
	gittest.Run(t, repo, "checkout", "--detach")

	r := run(t, repo, "spec", "validate")
	want := "Error: not on a slice branch (you are on no branch).\n" +
		"Give a name: sliceflow spec validate <name>\n" +
		"See all slices: sliceflow slice list\n"
	wantRun(t, r, 1, "", want)
}

// TestValidateOutsideRepo: outside a git repo, 0001's R12 error (R5).
func TestValidateOutsideRepo(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(dir))

	r := run(t, dir, "spec", "validate", "spec-submit")
	wantRun(t, r, 1, "", "Error: not inside a git repository. Nothing was changed.\n")
}

// TestValidateNoCommits: in a repo with no commits, 0001's R13 error (R5).
func TestValidateNoCommits(t *testing.T) {
	repo := gittest.Init(t)

	r := run(t, repo, "spec", "validate", "spec-submit")
	wantRun(t, r, 1, "", "Error: the repository has no commits yet. Make a first commit, then try again. Nothing was changed.\n")
}

// TestValidateS8: the branch exists, but no worktree has it (R7).
func TestValidateS8(t *testing.T) {
	repo := gittest.NewRepo(t)
	createSlice(t, repo, "old-idea")
	gittest.Run(t, repo, "worktree", "remove", ".sliceflow/worktrees/old-idea")

	r := run(t, repo, "spec", "validate", "old-idea")
	wantRun(t, r, 1, "", "Error: slice old-idea has no worktree: branch slice/old-idea is not checked out anywhere.\n")
}

// TestValidateS9: no slice with that name (R6).
func TestValidateS9(t *testing.T) {
	repo := gittest.NewRepo(t)

	r := run(t, repo, "spec", "validate", "no-such-slice")
	wantRun(t, r, 1, "", "Error: no slice named no-such-slice. See all slices: sliceflow slice list\n")
}

// TestValidateS10: the spec file was deleted in the worktree (R8).
func TestValidateS10(t *testing.T) {
	repo := gittest.NewRepo(t)
	worktree := createSlice(t, repo, "spec-submit")
	if err := os.Remove(specFile(worktree, "spec-submit")); err != nil {
		t.Fatal(err)
	}

	r := run(t, repo, "spec", "validate", "spec-submit")
	wantRun(t, r, 1, "", "Error: spec not found: "+mainSpecPath+".\n")
}

// TestValidateWrongUse: more than one name is wrong use.
func TestValidateWrongUse(t *testing.T) {
	r := run(t, t.TempDir(), "spec", "validate", "a", "b")
	wantRun(t, r, 2, "", "Error: too many arguments. Run sliceflow --help.\n")
}

// repoState is what R36 says must not change: every file's git status in
// each worktree, every ref, the HEAD of every worktree, and the event log.
func repoState(t *testing.T, repo, worktree string) string {
	t.Helper()
	return strings.Join([]string{
		// -uall and --ignored=matching list every file, also each one
		// inside the ignored .sliceflow folder.
		gittest.Run(t, repo, "status", "--porcelain", "--ignored=matching", "-uall"),
		gittest.Run(t, worktree, "status", "--porcelain", "--ignored=matching", "-uall"),
		gittest.Run(t, repo, "for-each-ref"),
		gittest.Run(t, repo, "worktree", "list", "--porcelain"),
		readFile(t, filepath.ToSlash(logPath(worktree, "spec-submit"))),
	}, "\n---\n")
}

// TestValidateChangesNothing: validate never changes a file, a branch, or
// the event log, with or without problems (R36).
func TestValidateChangesNothing(t *testing.T) {
	repo := gittest.NewRepo(t)
	worktree := createSlice(t, repo, "spec-submit")

	writeSpec(t, worktree, "spec-submit", validSpec("spec-submit"))
	before := repoState(t, repo, worktree)
	run(t, repo, "spec", "validate", "spec-submit")
	run(t, worktree, "spec", "validate", "--json")
	run(t, repo, "spec", "validate", "no-such-slice")
	if after := repoState(t, repo, worktree); after != before {
		t.Errorf("state changed\nbefore:\n%s\nafter:\n%s", before, after)
	}

	writeSpec(t, worktree, "spec-submit", specS2(t, "spec-submit"))
	before = repoState(t, repo, worktree)
	run(t, repo, "spec", "validate", "spec-submit")
	run(t, repo, "spec", "validate", "spec-submit", "--json")
	if after := repoState(t, repo, worktree); after != before {
		t.Errorf("state changed\nbefore:\n%s\nafter:\n%s", before, after)
	}
}
