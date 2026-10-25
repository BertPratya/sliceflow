package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BertPratya/sliceflow/internal/testrepo"
)

// s6Spec is the spec of S6: R7 has no Check: line, Open questions lists Q1
// and Q2, and no scenario shows R8.
func s6Spec(name string) string {
	return `---
slice: ` + name + `
created: 2026-10-24
ideas: []
---

# Spec
## Summary

Text.
` + strings.Repeat("More text.\n", 15) + `
## Why

Text.

## User stories

### Story 1: x

- **S1.** When x.
  Shows: R1, R7

## Requirements

- **R1.** Thing.
  Check: unit
- **R7.** Other thing.
  More about it.
- **R8.** More.
  Check: unit

## Out of scope

- Nothing. Goes to: the backlog

## Decisions

- **SD1.** Choice.
  Why: a
  Not chosen: b
  Rule: R1

## Assumptions

- A.
` + strings.Repeat("- A.\n", 79) + `## Open questions

- **Q1.** a?
- **Q2.** b?
`
}

func lineNo(t *testing.T, text, anchor string) int {
	t.Helper()
	for i, l := range strings.Split(text, "\n") {
		if strings.Contains(l, anchor) {
			return i + 1
		}
	}
	t.Fatalf("no %q", anchor)
	return 0
}

func writeSpec(t *testing.T, dir, name, text string) {
	t.Helper()
	p := filepath.Join(sliceWT(dir, name), "docs", "slices", name, "spec.md")
	if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

// s6Lines are the three lines of S6, with the line numbers S6 shows, for
// the spec of s6Spec. S7 and S8 add R7's Check: line, which moves R8 and
// Open questions down by one line.
func s6Lines(shift int) (errLine, note, warn string) {
	return "spec.md:42: error: R7 has no Check: line\n",
		fmt.Sprintf("spec.md:%d: note: no scenario shows R8\n", 44+shift),
		fmt.Sprintf("spec.md:%d: warning: 2 open questions left (Q1, Q2)\n", 140+shift)
}

func TestValidateProblems(t *testing.T) {
	dir := testrepo.New(t)
	createSlice(t, dir, "login-google")
	text := s6Spec("login-google")
	writeSpec(t, dir, "login-google", text)
	e, n, w := s6Lines(0)
	if lineNo(t, text, "**R7.**") != 42 || lineNo(t, text, "**R8.**") != 44 || lineNo(t, text, "## Open questions") != 140 {
		t.Fatal("the test spec does not match S6's line numbers")
	}
	want := e + n + w + "1 error, 1 warning, 1 note\n"
	head := testrepo.Git(t, sliceWT(dir, "login-google"), "rev-parse", "HEAD")
	out, errOut, code := run(t, sliceWT(dir, "login-google"), "spec", "validate")
	if code != 1 || out != want || errOut != "" {
		t.Fatalf("code %d\nstdout %q\nwant   %q\nstderr %q", code, out, want, errOut)
	}
	if testrepo.Git(t, sliceWT(dir, "login-google"), "rev-parse", "HEAD") != head {
		t.Fatal("a commit was made")
	}
}

func s7Text() string {
	return strings.Replace(s6Spec("login-google"), "- **R7.** Other thing.\n", "- **R7.** Other thing.\n  Check: unit\n", 1)
}

func TestValidateWarningOnly(t *testing.T) {
	dir := testrepo.New(t)
	createSlice(t, dir, "login-google")
	text := s7Text()
	writeSpec(t, dir, "login-google", text)
	_, n, w := s6Lines(1)
	want := n + w + "0 errors, 1 warning, 1 note\n"
	out, errOut, code := run(t, sliceWT(dir, "login-google"), "spec", "validate")
	if code != 0 || out != want || errOut != "" {
		t.Fatalf("code %d\nstdout %q\nwant   %q\nstderr %q", code, out, want, errOut)
	}
}

func TestValidateStrict(t *testing.T) {
	dir := testrepo.New(t)
	createSlice(t, dir, "login-google")
	text := s7Text()
	writeSpec(t, dir, "login-google", text)
	_, n, w := s6Lines(1)
	want := n + w + "0 errors, 1 warning, 1 note\n"
	out, errOut, code := run(t, sliceWT(dir, "login-google"), "spec", "validate", "--strict")
	if code != 1 || out != want || errOut != "" {
		t.Fatalf("code %d\nstdout %q\nwant   %q\nstderr %q", code, out, want, errOut)
	}
	// A note alone never changes the exit code.
	text = strings.Replace(text, "- **Q1.** a?\n- **Q2.** b?\n", "None.\n", 1)
	writeSpec(t, dir, "login-google", text)
	out, _, code = run(t, sliceWT(dir, "login-google"), "spec", "validate", "--strict")
	if code != 0 || !strings.HasSuffix(out, "0 errors, 0 warnings, 1 note\n") {
		t.Fatalf("code %d stdout %q", code, out)
	}
}

func TestValidateByName(t *testing.T) {
	dir := testrepo.New(t)
	createSlice(t, dir, "login-google")
	text := strings.Replace(s7Text(), "- **Q1.** a?\n- **Q2.** b?\n", "None.\n", 1)
	text = strings.Replace(text, "  Shows: R1, R7\n", "  Shows: R1, R7, R8\n", 1)
	writeSpec(t, dir, "login-google", text)
	out, errOut, code := run(t, dir, "spec", "validate", "login-google")
	if code != 0 || out != "0 errors, 0 warnings, 0 notes\n" || errOut != "" {
		t.Fatalf("code %d stdout %q stderr %q", code, out, errOut)
	}
}

func TestValidateNoBranch(t *testing.T) {
	dir := testrepo.New(t)
	out, errOut, code := run(t, dir, "spec", "validate")
	want := "Error: not on a slice branch (you are on main).\n" +
		"Give a name: sliceflow spec validate <name>\n" +
		"See all slices: sliceflow slice list\n"
	if code != 1 || out != "" || errOut != want {
		t.Fatalf("code %d stdout %q stderr %q", code, out, errOut)
	}
}

func TestValidateSkipsCode(t *testing.T) {
	dir := testrepo.New(t)
	createSlice(t, dir, "login-google")
	text := strings.Replace(s7Text(), "- **Q1.** a?\n- **Q2.** b?\n", "None.\n", 1)
	text = strings.Replace(text, "  Shows: R1, R7\n", "  Shows: R1, R7, R8\n", 1)
	text = strings.Replace(text, "## Summary\n\nText.\n",
		"## Summary\n\nText with `<!--` here.\n\n```\nname: {{slice}}\n```\n", 1)
	writeSpec(t, dir, "login-google", text)
	out, errOut, code := run(t, sliceWT(dir, "login-google"), "spec", "validate")
	if code != 0 || out != "0 errors, 0 warnings, 0 notes\n" || errOut != "" {
		t.Fatalf("code %d stdout %q stderr %q", code, out, errOut)
	}
}

func TestValidateNoWorktree(t *testing.T) {
	dir := testrepo.New(t)
	createSlice(t, dir, "login-google")
	testrepo.Git(t, dir, "worktree", "remove", "--force", sliceWT(dir, "login-google"))
	root, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := "Error: slice login-google has no worktree: no worktree has the branch slice/login-google checked out. Nothing was changed.\n" +
		"Make one: git worktree add " + filepath.ToSlash(root) + "/.sliceflow/worktrees/login-google slice/login-google\n"
	out, errOut, code := run(t, dir, "spec", "validate", "login-google")
	if code != 1 || out != "" || errOut != want {
		t.Fatalf("code %d stdout %q\nstderr %q\nwant   %q", code, out, errOut, want)
	}
}

func TestValidateNoSpec(t *testing.T) {
	dir := testrepo.New(t)
	createSlice(t, dir, "login-google")
	p := filepath.Join(sliceWT(dir, "login-google"), "docs", "slices", "login-google", "spec.md")
	if err := os.Remove(p); err != nil {
		t.Fatal(err)
	}
	out, errOut, code := run(t, sliceWT(dir, "login-google"), "spec", "validate")
	want := "Error: .sliceflow/worktrees/login-google/docs/slices/login-google/spec.md not found. Nothing was changed.\n"
	if code != 1 || out != "" || errOut != want {
		t.Fatalf("code %d stdout %q stderr %q", code, out, errOut)
	}
}

func TestValidateJSON(t *testing.T) {
	dir := testrepo.New(t)
	createSlice(t, dir, "login-google")
	out, errOut, code := run(t, sliceWT(dir, "login-google"), "spec", "validate", "--json")
	if code != 2 || out != "" || errOut != "Error: unknown flag --json. Run sliceflow --help.\n" {
		t.Fatalf("code %d stdout %q stderr %q", code, out, errOut)
	}
}

func TestValidateMerged(t *testing.T) {
	dir := testrepo.New(t)
	createSlice(t, dir, "login-google")
	mergeAndRemove(t, dir, "login-google")
	out, errOut, code := run(t, dir, "spec", "validate", "login-google")
	want := "Error: slice login-google is merged into main, so it has no worktree. Nothing was changed.\n"
	if code != 1 || out != "" || errOut != want {
		t.Fatalf("code %d stdout %q stderr %q", code, out, errOut)
	}
}

func TestOpenSliceOrder(t *testing.T) {
	dir := testrepo.New(t)
	createSlice(t, dir, "has-branch")
	testrepo.Git(t, dir, "worktree", "remove", "--force", sliceWT(dir, "has-branch"))
	cases := []struct{ name, want string }{
		{"Bad_Name", `Error: invalid slice name "Bad_Name": use 1-64 lowercase letters, digits, and dashes, starting with a letter or digit. Nothing was changed.` + "\n"},
		{"no-such-slice", "Error: no slice named no-such-slice. See all slices: sliceflow slice list\n"},
	}
	for _, c := range cases {
		out, errOut, code := run(t, dir, "spec", "validate", c.name)
		if code != 1 || out != "" || errOut != c.want {
			t.Errorf("%s: code %d stdout %q stderr %q", c.name, code, out, errOut)
		}
	}
	_, errOut, code := run(t, dir, "spec", "validate", "has-branch")
	if code != 1 || strings.Count(errOut, "\n") != 2 || !strings.HasPrefix(errOut, "Error: slice has-branch has no worktree: ") {
		t.Errorf("no worktree: code %d stderr %q", code, errOut)
	}
}
