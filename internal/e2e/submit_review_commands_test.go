package e2e

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/BertPratya/sliceflow/internal/testrepo"
)

const specName = "login-google"

// s14Text is the spec of S14 (S7 in the spec): R7 has its Check: line, so the
// note and the warning are on lines 44 and 140, as S14 shows.
func s14Text() string {
	return strings.Replace(s6Spec(specName), "  More about it.\n", "  Check: unit\n", 1)
}

// s15Text is the spec of S15: Q1 and Q2 are settled.
func s15Text() string {
	return strings.Replace(s14Text(), "- **Q1.** a?\n- **Q2.** b?\n", "None.\n", 1)
}

var (
	logAtRe = regexp.MustCompile(`"at":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z"`)
)

// maskLine replaces the time (which the real run cannot know) in a log line.
func maskLine(s string) string { return logAtRe.ReplaceAllString(s, `"at":"T"`) }

func wtOf(dir string) string { return sliceWT(dir, specName) }

func logLines(t *testing.T, dir string) []string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(wtOf(dir), "docs", "slices", specName, ".events.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	return strings.Split(strings.TrimRight(string(b), "\n"), "\n")
}

func headOf(t *testing.T, dir string) string {
	t.Helper()
	return strings.TrimSpace(testrepo.Git(t, wtOf(dir), "rev-parse", "HEAD"))
}

func subjects(t *testing.T, dir string, n string) []string {
	t.Helper()
	return strings.Split(strings.TrimSpace(testrepo.Git(t, wtOf(dir), "log", "-"+n, "--format=%s")), "\n")
}

func filesOf(t *testing.T, dir, rev string) string {
	t.Helper()
	return strings.TrimSpace(testrepo.Git(t, wtOf(dir), "show", "--name-only", "--format=", rev))
}

func statusTail(t *testing.T, dir string) string {
	t.Helper()
	out, e, c := run(t, wtOf(dir), "slice", "status")
	if c != 0 {
		t.Fatalf("status: code %d stderr %q", c, e)
	}
	return lastLines(out, 2)
}

func failsWith(t *testing.T, dir string, wantErr string, args ...string) {
	t.Helper()
	before := headOf(t, dir)
	out, errOut, code := run(t, wtOf(dir), args...)
	if code != 1 || out != "" || errOut != wantErr+"\n" {
		t.Fatalf("code %d\nstdout %q\nstderr %q\nwant   %q", code, out, errOut, wantErr+"\n")
	}
	if headOf(t, dir) != before {
		t.Fatal("a commit was made")
	}
}

func draftSlice(t *testing.T, text string) string {
	t.Helper()
	dir := testrepo.New(t)
	createSlice(t, dir, specName)
	writeSpec(t, dir, specName, text)
	return dir
}

// submittedSlice is the slice of S15, after the submit.
func submittedSlice(t *testing.T) string {
	t.Helper()
	dir := draftSlice(t, s15Text())
	if _, e, c := run(t, wtOf(dir), "spec", "submit"); c != 0 {
		t.Fatalf("submit: code %d stderr %q", c, e)
	}
	return dir
}

// reviewedSlice is the slice of S20, after the review.
func reviewedSlice(t *testing.T) string {
	t.Helper()
	dir := submittedSlice(t)
	if _, e, c := run(t, wtOf(dir), "spec", "review"); c != 0 {
		t.Fatalf("review: code %d stderr %q", c, e)
	}
	return dir
}

const (
	note44   = "spec.md:44: note: no scenario shows R8\n"
	warn140  = "spec.md:140: warning: 2 open questions left (Q1, Q2)\n"
	reviewed = "Review 1 started: reviews/spec-1.md (reviewer: placeholder)\nReview 1 passed.\nNext: write the plan.\n"
)

func TestSubmitRefusedCmd(t *testing.T) {
	dir := draftSlice(t, s14Text())
	before := headOf(t, dir)
	out, errOut, code := run(t, wtOf(dir), "spec", "submit")
	want := note44 + warn140 +
		"Submit refused: spec.md is not finished (0 errors, 1 warning). It stays a draft.\n"
	if code != 1 || out != want || errOut != "" {
		t.Fatalf("code %d\nstdout %q\nwant   %q\nstderr %q", code, out, want, errOut)
	}
	if n := testrepo.Git(t, wtOf(dir), "rev-list", "--count", before+"..HEAD"); strings.TrimSpace(n) != "1" {
		t.Fatalf("commits: %s", n)
	}
	if s := subjects(t, dir, "1"); strings.Join(s, "") != "login-google: spec submit refused" {
		t.Fatalf("subject %v", s)
	}
	wantFiles := "docs/slices/login-google/.events.jsonl\ndocs/slices/login-google/spec.md"
	if f := filesOf(t, dir, "HEAD"); f != wantFiles {
		t.Fatalf("files %q", f)
	}
	if got := testrepo.Git(t, wtOf(dir), "show", "HEAD:docs/slices/login-google/spec.md"); got != s14Text() {
		t.Fatal("spec.md in the commit is not the one on disk")
	}
	ls := logLines(t, dir)
	wantLine := `{"version":1,"seq":2,"event":"spec.submitted","at":"T","data":{"result":"refused","errors":0,"warnings":1}}`
	if len(ls) != 2 || maskLine(ls[1]) != wantLine {
		t.Fatalf("log %q", ls)
	}
	wantStatus := "Spec:     draft (submit refused: 1 warning)\nNext:     fix spec.md, then: sliceflow spec submit\n"
	if got := statusTail(t, dir); got != wantStatus {
		t.Fatalf("status %q", got)
	}
}

func TestSubmitAcceptedCmd(t *testing.T) {
	dir := draftSlice(t, s15Text())
	before := headOf(t, dir)
	out, errOut, code := run(t, wtOf(dir), "spec", "submit")
	want := note44 + "Submitted spec.md.\nNext: sliceflow spec review\n"
	if code != 0 || out != want || errOut != "" {
		t.Fatalf("code %d\nstdout %q\nwant   %q\nstderr %q", code, out, want, errOut)
	}
	if n := testrepo.Git(t, wtOf(dir), "rev-list", "--count", before+"..HEAD"); strings.TrimSpace(n) != "1" {
		t.Fatalf("commits: %s", n)
	}
	if s := subjects(t, dir, "1"); strings.Join(s, "") != "login-google: spec submitted" {
		t.Fatalf("subject %v", s)
	}
	ls := logLines(t, dir)
	wantLine := `{"version":1,"seq":2,"event":"spec.submitted","at":"T","data":{"result":"accepted","errors":0,"warnings":0}}`
	if len(ls) != 2 || maskLine(ls[1]) != wantLine {
		t.Fatalf("log %q", ls)
	}
	wantStatus := "Spec:     submitted (review 1 not started)\nNext:     sliceflow spec review\n"
	if got := statusTail(t, dir); got != wantStatus {
		t.Fatalf("status %q", got)
	}
}

func TestSubmitTwice(t *testing.T) {
	dir := submittedSlice(t)
	failsWith(t, dir, "Error: spec.md is submitted; review it with: sliceflow spec review. Nothing was changed.", "spec", "submit")
}

func TestReviewCmd(t *testing.T) {
	dir := submittedSlice(t)
	submitCommit := headOf(t, dir)
	out, errOut, code := run(t, wtOf(dir), "spec", "review")
	if code != 0 || out != reviewed || errOut != "" {
		t.Fatalf("code %d\nstdout %q\nwant   %q\nstderr %q", code, out, reviewed, errOut)
	}
	if n := testrepo.Git(t, wtOf(dir), "rev-list", "--count", submitCommit+"..HEAD"); strings.TrimSpace(n) != "2" {
		t.Fatalf("commits: %s", n)
	}
	s := subjects(t, dir, "2")
	if len(s) != 2 || s[0] != "login-google: review 1 passed" || s[1] != "login-google: review 1 started" {
		t.Fatalf("subjects %v", s)
	}
	wantFiles := "docs/slices/login-google/.events.jsonl\ndocs/slices/login-google/reviews/spec-1.md"
	if f := filesOf(t, dir, "HEAD"); f != wantFiles {
		t.Fatalf("files %q", f)
	}
	if f := filesOf(t, dir, "HEAD~1"); f != wantFiles {
		t.Fatalf("files %q", f)
	}
	front := "---\nreview: spec-1\nslice: login-google\nkind: spec\ntarget: spec.md\nround: 1\nreviewer: placeholder\n---\n"
	if got := testrepo.Git(t, wtOf(dir), "show", "HEAD~1:docs/slices/login-google/reviews/spec-1.md"); got != front {
		t.Fatalf("first review file %q", got)
	}
	body := "# Review\n\n## Summary\n\nNo review was done: placeholder reviewer.\n\n## Checked\n\nNone.\n\n## Tests\n\nNone.\n\n" +
		"## Previous findings\n\nNone.\n\n## Findings\n\nNone.\n\n## Requirements check\n\nNone.\n\n## Needs human decision\n\nNone.\n"
	if got := testrepo.Git(t, wtOf(dir), "show", "HEAD:docs/slices/login-google/reviews/spec-1.md"); got != front+body {
		t.Fatalf("review file %q", got)
	}
	ls := logLines(t, dir)
	l1 := `{"version":1,"seq":3,"event":"review.started","at":"T","data":{"doc":"spec.md","round":1,"reviewer":"placeholder","file":"reviews/spec-1.md","commit":"` + submitCommit + `"}}`
	l2 := `{"version":1,"seq":4,"event":"review.finished","at":"T","data":{"doc":"spec.md","round":1,"result":"passed"}}`
	if len(ls) != 4 || maskLine(ls[2]) != l1 || maskLine(ls[3]) != l2 {
		t.Fatalf("log %q", ls)
	}
	wantStatus := "Spec:     review_passed (review 1, placeholder)\nNext:     write the plan\n"
	if got := statusTail(t, dir); got != wantStatus {
		t.Fatalf("status %q", got)
	}
}

func TestReviewDraft(t *testing.T) {
	dir := draftSlice(t, s14Text())
	if _, _, c := run(t, wtOf(dir), "spec", "submit"); c != 1 {
		t.Fatal("submit should refuse")
	}
	failsWith(t, dir, "Error: spec.md is a draft; submit it first: sliceflow spec submit. Nothing was changed.", "spec", "review")
}

func TestReviewInReview(t *testing.T) {
	dir := submittedSlice(t)
	appendLog(t, dir, specName, startedLine(3, 1, headOf(t, dir)))
	failsWith(t, dir, "Error: spec.md is in_review: review 1 has started, and has not finished. Nothing was changed.", "spec", "review")
}

func TestReviewNothing(t *testing.T) {
	dir := reviewedSlice(t)
	failsWith(t, dir, "Error: spec.md passed review 1; there is nothing to review. Nothing was changed.", "spec", "review")
}

func TestSubmitUnchanged(t *testing.T) {
	dir := reviewedSlice(t)
	failsWith(t, dir, "Error: spec.md has not changed since review 1, which passed. Nothing was changed.", "spec", "submit")
}

func changeR7(t *testing.T, dir string) {
	t.Helper()
	writeSpec(t, dir, specName, strings.Replace(s15Text(), "- **R7.** Other thing.", "- **R7.** Other thing, changed.", 1))
}

func TestSubmitRefusedReviewed(t *testing.T) {
	dir := reviewedSlice(t)
	text := strings.Replace(s15Text(), "- **R7.** Other thing.", "- **R7.** Other thing, changed.", 1)
	text = strings.Replace(text, "## Open questions\n\nNone.\n", "## Open questions\n\n- **Q1.** a?\n", 1)
	if text == s15Text() {
		t.Fatal("no edit made")
	}
	writeSpec(t, dir, specName, text)
	out, errOut, code := run(t, wtOf(dir), "spec", "submit")
	last := "Submit refused: spec.md is not finished (0 errors, 1 warning). It keeps the result of review 1.\n"
	if code != 1 || !strings.HasSuffix(out, last) || errOut != "" {
		t.Fatalf("code %d stdout %q stderr %q", code, out, errOut)
	}
}

func TestSubmitChangedStatus(t *testing.T) {
	dir := reviewedSlice(t)
	changeR7(t, dir)
	want := "Spec:     passed review 1, but spec.md was changed after it\nNext:     sliceflow spec submit\n"
	if got := statusTail(t, dir); got != want {
		t.Fatalf("status %q", got)
	}
}

func TestReviewRound2Cmd(t *testing.T) {
	dir := reviewedSlice(t)
	changeR7(t, dir)
	out, errOut, code := run(t, wtOf(dir), "spec", "submit")
	want := note44 + "Submitted spec.md.\nNext: sliceflow spec review\n"
	if code != 0 || out != want || errOut != "" {
		t.Fatalf("submit: code %d stdout %q stderr %q", code, out, errOut)
	}
	out, errOut, code = run(t, wtOf(dir), "spec", "review")
	want = "Review 2 started: reviews/spec-2.md (reviewer: placeholder)\nReview 2 passed.\nNext: write the plan.\n"
	if code != 0 || out != want || errOut != "" {
		t.Fatalf("review: code %d stdout %q stderr %q", code, out, errOut)
	}
	if got := statusTail(t, dir); got != "Spec:     review_passed (review 2, placeholder)\nNext:     write the plan\n" {
		t.Fatalf("status %q", got)
	}
}
