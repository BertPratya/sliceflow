// Package review holds the Reviewer contract, the review file's front
// matter, and the reader that gets the result from a review file.
package review

import (
	"context"
	"fmt"
	"strings"
)

// Reviewer reviews a document and writes the review file.
type Reviewer interface {
	// Name is the event's reviewer, such as "placeholder".
	Name() string
	// Review writes the review file at req.File.
	Review(ctx context.Context, req Request) error
}

// Request says what to review. File is the review file's full path.
type Request struct {
	Slice, Doc, File, Commit, Worktree string
	Round                              int
}

// FrontMatter returns the front matter of a spec review file. Line ends
// are always "\n".
func FrontMatter(slice string, round int, reviewer string) []byte {
	return []byte(fmt.Sprintf("---\nreview: spec-%d\nslice: %s\nkind: spec\ntarget: spec.md\nround: %d\nreviewer: %s\n---\n",
		round, slice, round, reviewer))
}

const needsHeading = "## Needs human decision"

// ReadResult returns "needs_human" when the section "## Needs human
// decision" is missing or is not "None."; else "passed".
func ReadResult(file []byte) string {
	lines := strings.Split(strings.ReplaceAll(string(file), "\r\n", "\n"), "\n")
	in, found := false, false
	var body []string
	for _, l := range lines {
		if in {
			if strings.HasPrefix(l, "## ") {
				break
			}
			body = append(body, l)
			continue
		}
		if strings.TrimRight(l, " \t") == needsHeading {
			in, found = true, true
		}
	}
	if found && strings.TrimSpace(strings.Join(body, "\n")) == "None." {
		return "passed"
	}
	return "needs_human"
}
