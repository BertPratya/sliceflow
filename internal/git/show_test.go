package git

import (
	"context"
	"path/filepath"
	"testing"
)

// FileAt returns a committed file as it is at the branch's tip.
func TestFileAtACommittedFile(t *testing.T) {
	repo := newRepoWithMain(t)
	writeText(t, filepath.Join(repo, "docs", "log.jsonl"), "one\ntwo\n")
	gitIn(t, repo, "add", ".")
	gitIn(t, repo, "commit", "--quiet", "-m", "log")

	data, found, err := FileAt(context.Background(), repo, "main", "docs/log.jsonl")
	if err != nil || !found {
		t.Fatalf("FileAt returned found %v and the error %v, want true and none", found, err)
	}
	if string(data) != "one\ntwo\n" {
		t.Errorf("FileAt returned %q, want %q", data, "one\ntwo\n")
	}
}

// FileAt reads the branch, not the files on disk.
func TestFileAtIgnoresChangesNotCommitted(t *testing.T) {
	repo := newRepoWithMain(t)
	writeText(t, filepath.Join(repo, "log.jsonl"), "committed\n")
	gitIn(t, repo, "add", ".")
	gitIn(t, repo, "commit", "--quiet", "-m", "log")
	writeText(t, filepath.Join(repo, "log.jsonl"), "changed\n")

	data, _, err := FileAt(context.Background(), repo, "main", "log.jsonl")
	if err != nil {
		t.Fatalf("FileAt returned the error %v, want none", err)
	}
	if string(data) != "committed\n" {
		t.Errorf("FileAt returned %q, want %q", data, "committed\n")
	}
}

// FileAt reports not found, with no error, for a missing file or a
// missing branch.
func TestFileAtNotFound(t *testing.T) {
	repo := newRepoWithMain(t)
	tests := []struct{ about, branch, path string }{
		{"no such file", "main", "nosuch.jsonl"},
		{"no such branch", "nosuch", "nosuch.jsonl"},
	}
	for _, tt := range tests {
		t.Run(tt.about, func(t *testing.T) {
			data, found, err := FileAt(context.Background(), repo, tt.branch, tt.path)
			if err != nil || found || data != nil {
				t.Errorf("FileAt returned %q, found %v, and the error %v; want nothing, false, and none", data, found, err)
			}
		})
	}
}
