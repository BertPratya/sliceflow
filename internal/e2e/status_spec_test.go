package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BertPratya/sliceflow/internal/testrepo"
)

const (
	nextWrite  = "write spec.md, then: sliceflow spec submit"
	specAt     = "2026-10-24T09:00:00Z"
	fakeCommit = "3f2a1c9e5b7d4a6f8c0e2b4d6f8a0c2e4b6d8f0a"
)

// sliceWT returns the worktree folder of a slice.
func sliceWT(dir, name string) string {
	return filepath.Join(dir, ".sliceflow", "worktrees", name)
}

// evLine builds one log line.
func evLine(seq int, event, data string) string {
	return fmt.Sprintf(`{"version":1,"seq":%d,"event":%q,"at":%q,"data":%s}`, seq, event, specAt, data)
}

func submittedLine(seq int, result string, errs, warns int) string {
	return evLine(seq, "spec.submitted", fmt.Sprintf(`{"result":%q,"errors":%d,"warnings":%d}`, result, errs, warns))
}

func startedLine(seq, round int, commit string) string {
	return evLine(seq, "review.started", fmt.Sprintf(
		`{"doc":"spec.md","round":%d,"reviewer":"placeholder","file":"reviews/spec-%d.md","commit":%q}`, round, round, commit))
}

func finishedLine(seq, round int, result string) string {
	return evLine(seq, "review.finished", fmt.Sprintf(`{"doc":"spec.md","round":%d,"result":%q}`, round, result))
}

// appendLog adds lines to the log of a slice in its worktree and commits.
func appendLog(t *testing.T, dir, name string, lines ...string) {
	t.Helper()
	wt := sliceWT(dir, name)
	logFile := filepath.Join(wt, "docs", "slices", name, ".events.jsonl")
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(strings.Join(lines, "\n") + "\n"); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	testrepo.Git(t, wt, "commit", "-am", "log lines")
}

// commitSpec writes and commits spec.md, and returns the commit id.
func commitSpec(t *testing.T, dir, name, text string) string {
	t.Helper()
	wt := sliceWT(dir, name)
	p := filepath.Join(wt, "docs", "slices", name, "spec.md")
	if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	testrepo.Git(t, wt, "add", p)
	testrepo.Git(t, wt, "commit", "-m", "spec")
	return strings.TrimSpace(testrepo.Git(t, wt, "rev-parse", "HEAD"))
}

// passedSlice makes a slice whose review 1 passed. It returns the commit.
func passedSlice(t *testing.T, dir, name string) string {
	t.Helper()
	createSlice(t, dir, name)
	commit := commitSpec(t, dir, name, "# Spec\n")
	appendLog(t, dir, name, submittedLine(2, "accepted", 0, 0), startedLine(3, 1, commit), finishedLine(4, 1, "passed"))
	return commit
}

func lastLines(s string, n int) string {
	lines := strings.SplitAfter(s, "\n")
	lines = lines[:len(lines)-1]
	return strings.Join(lines[len(lines)-n:], "")
}

func statusJSON(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, errOut, code := run(t, dir, append([]string{"slice", "status"}, args...)...)
	if code != 0 || errOut != "" {
		t.Fatalf("status %v: code %d stderr %q", args, code, errOut)
	}
	return out
}

func TestStatusSpecDraft(t *testing.T) {
	dir := testrepo.New(t)
	commit := strings.TrimSpace(testrepo.Git(t, dir, "rev-parse", "HEAD"))
	createSlice(t, dir, "login-google", "--title", "Login with Google")
	setCreated(t, dir, "login-google", "2026-10-24T07:42:00Z")
	out, errOut, code := run(t, sliceWT(dir, "login-google"), "slice", "status")
	want := "Slice:    login-google\n" +
		"Title:    Login with Google\n" +
		"Phase:    planning\n" +
		"From:     main @ " + commit[:7] + "\n" +
		"Created:  2026-10-24 07:42 UTC\n" +
		"Spec:     draft\n" +
		"Next:     write spec.md, then: sliceflow spec submit\n"
	if code != 0 || errOut != "" || out != want {
		t.Fatalf("code %d\nstdout %q\nwant   %q\nstderr %q", code, out, want, errOut)
	}
}

func TestStatusSpecJSON(t *testing.T) {
	dir := testrepo.New(t)
	commit := strings.TrimSpace(testrepo.Git(t, dir, "rev-parse", "HEAD"))
	createSlice(t, dir, "login-google", "--title", "Login with Google")
	setCreated(t, dir, "login-google", "2026-10-24T07:42:00Z")
	out, errOut, code := run(t, sliceWT(dir, "login-google"), "slice", "status", "--json")
	want := `{"slice":"login-google","title":"Login with Google","phase":"planning","from":{"branch":"main","commit":"` + commit + `"},` +
		`"created":"2026-10-24T07:42:00Z","docs":[{"doc":"spec.md","status":"draft"}],"next":"write spec.md, then: sliceflow spec submit","warnings":[]}` + "\n"
	if code != 0 || errOut != "" || out != want {
		t.Fatalf("code %d\nstdout %q\nwant   %q\nstderr %q", code, out, want, errOut)
	}
}

func TestStatusNotAllowedLine(t *testing.T) {
	dir := testrepo.New(t)
	createSlice(t, dir, "login-google")
	appendLog(t, dir, "login-google",
		evLine(2, "slice.titled", `{"title":"One"}`),
		evLine(3, "slice.titled", `{"title":"Two"}`),
		finishedLine(4, 1, "passed"))
	out, errOut, code := run(t, dir, "slice", "status", "login-google")
	wantErr := "Warning: login-google: .events.jsonl line 4: review.finished is not allowed when spec.md is draft; skipped\n"
	if code != 0 || errOut != wantErr || !strings.HasSuffix(out, "Spec:     draft\nNext:     "+nextWrite+"\n") {
		t.Fatalf("code %d\nstdout %q\nstderr %q\nwant   %q", code, out, errOut, wantErr)
	}
}

func TestStatusReviewChanges(t *testing.T) {
	dir := testrepo.New(t)
	createSlice(t, dir, "login-google")
	commit := commitSpec(t, dir, "login-google", "# Spec\n")
	appendLog(t, dir, "login-google", submittedLine(2, "accepted", 0, 0), startedLine(3, 1, commit), finishedLine(4, 1, "changes"))
	out, errOut, code := run(t, dir, "slice", "status", "login-google")
	want := "Spec:     draft (review 1 found changes: reviews/spec-1.md)\n" +
		"Next:     fix spec.md, then: sliceflow spec submit\n"
	if code != 0 || errOut != "" || lastLines(out, 2) != want {
		t.Fatalf("code %d\nstdout %q\nwant   %q\nstderr %q", code, out, want, errOut)
	}
}

func TestStatusMerged(t *testing.T) {
	dir := testrepo.New(t)
	passedSlice(t, dir, "login-google")
	mergeAndRemove(t, dir, "login-google")
	out, errOut, code := run(t, dir, "slice", "status", "login-google")
	want := "Spec:     review_passed (review 1, placeholder)\nNext:     -\n"
	if code != 0 || errOut != "" || lastLines(out, 2) != want {
		t.Fatalf("code %d\nstdout %q\nwant   %q\nstderr %q", code, out, want, errOut)
	}
}

func TestStatusJSONReview(t *testing.T) {
	dir := testrepo.New(t)
	// A passed review.
	commit := passedSlice(t, dir, "passed-one")
	out := statusJSON(t, dir, "passed-one", "--json")
	wantDocs := `"docs":[{"doc":"spec.md","status":"review_passed","review":{"round":1,"reviewer":"placeholder","file":"reviews/spec-1.md","result":"passed"}}],"next":"write the plan"`
	if !strings.Contains(out, wantDocs) || strings.Contains(out, "changed_after_review") {
		t.Fatalf("passed: %q", out)
	}
	// The same slice, with spec.md edited on disk after the review.
	p := filepath.Join(sliceWT(dir, "passed-one"), "docs", "slices", "passed-one", "spec.md")
	if err := os.WriteFile(p, []byte("# Spec\nchanged\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out = statusJSON(t, dir, "passed-one", "--json")
	if !strings.Contains(out, `"changed_after_review":true`) || !strings.Contains(out, `"next":"sliceflow spec submit"`) {
		t.Fatalf("changed: %q", out)
	}
	// A refused submit.
	createSlice(t, dir, "refused-one")
	appendLog(t, dir, "refused-one", submittedLine(2, "refused", 0, 1))
	out = statusJSON(t, dir, "refused-one", "--json")
	if !strings.Contains(out, `"docs":[{"doc":"spec.md","status":"draft","submit_refused":{"errors":0,"warnings":1}}]`) ||
		!strings.Contains(out, `"next":"fix spec.md, then: sliceflow spec submit"`) {
		t.Fatalf("refused: %q", out)
	}
	// A running review has no result.
	createSlice(t, dir, "running-one")
	c2 := commitSpec(t, dir, "running-one", "# Spec\n")
	appendLog(t, dir, "running-one", submittedLine(2, "accepted", 0, 0), startedLine(3, 1, c2))
	out = statusJSON(t, dir, "running-one", "--json")
	if !strings.Contains(out, `"status":"in_review","review":{"round":1,"reviewer":"placeholder","file":"reviews/spec-1.md"}}]`) ||
		!strings.Contains(out, `"next":"wait for review 1"`) {
		t.Fatalf("running: %q", out)
	}
	_ = commit
	// No slice.created: docs is [], and next is left out.
	testrepo.Git(t, dir, "branch", "slice/old-idea")
	out, _, code := run(t, dir, "slice", "status", "old-idea", "--json")
	if code != 0 || !strings.Contains(out, `"docs":[]`) || strings.Contains(out, `"next"`) {
		t.Fatalf("no created: code %d %q", code, out)
	}
	// A merged slice has no next.
	passedSlice(t, dir, "merged-one")
	mergeAndRemove(t, dir, "merged-one")
	out = statusJSON(t, dir, "merged-one", "--json")
	if !strings.Contains(out, `"status":"review_passed"`) || strings.Contains(out, `"next"`) {
		t.Fatalf("merged: %q", out)
	}
	// slice list --json has the same docs and next for each slice.
	listOut, errOut, code := run(t, dir, "slice", "list", "--json", "--all")
	if code != 0 {
		t.Fatalf("list: code %d stderr %q", code, errOut)
	}
	var list struct {
		Slices []struct {
			Slice string          `json:"slice"`
			Docs  json.RawMessage `json:"docs"`
			Next  *string         `json:"next"`
		} `json:"slices"`
	}
	if err := json.Unmarshal([]byte(listOut), &list); err != nil {
		t.Fatalf("list json: %v: %q", err, listOut)
	}
	if len(list.Slices) != 5 {
		t.Fatalf("list has %d slices: %q", len(list.Slices), listOut)
	}
	for _, s := range list.Slices {
		var one struct {
			Docs json.RawMessage `json:"docs"`
			Next *string         `json:"next"`
		}
		stOut, _, _ := run(t, dir, "slice", "status", s.Slice, "--json")
		if err := json.Unmarshal([]byte(stOut), &one); err != nil {
			t.Fatal(err)
		}
		if string(s.Docs) != string(one.Docs) || (s.Next == nil) != (one.Next == nil) || (s.Next != nil && *s.Next != *one.Next) {
			t.Errorf("%s: list docs %s next %v; status docs %s next %v", s.Slice, s.Docs, s.Next, one.Docs, one.Next)
		}
	}
}
