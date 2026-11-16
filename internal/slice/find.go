package slice

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/BertPratya/sliceflow/internal/eventlog"
	"github.com/BertPratya/sliceflow/internal/git"
)

// Where a log was found.
const (
	fromWorktree = "worktree"
	fromBranch   = "branch"
	fromMain     = "main"
)

// foundLog is the log of a slice, and where it was found.
type foundLog struct {
	source   string // fromWorktree, fromBranch, or fromMain
	worktree string // the slice's worktree when its folder exists, or ""
	data     []byte // the whole log
}

// logPath returns where the log of the slice name is in the project,
// written with /.
func logPath(name string) string {
	return "docs/slices/" + name + "/.events.jsonl"
}

// findLog looks for the log of the slice name in the project whose main
// folder is root, in this order, and stops at the first place that has
// it: the slice's worktree, the tip of the branch slice/<name>, and the
// tip of main. found is false when none has it. The slice's worktree is
// returned in both cases, when git lists one for the branch and its
// folder exists.
func findLog(ctx context.Context, root, name string) (log foundLog, found bool, err error) {
	branch := "slice/" + name
	worktree, err := git.WorktreeOf(ctx, root, branch)
	if err != nil {
		return foundLog{}, false, err
	}
	if worktree != "" {
		if _, err := os.Stat(worktree); err != nil {
			worktree = ""
		}
	}
	log.worktree = worktree

	if worktree != "" {
		data, err := os.ReadFile(filepath.Join(worktree, filepath.FromSlash(logPath(name))))
		if err == nil {
			log.source, log.data = fromWorktree, data
			return log, true, nil
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return foundLog{}, false, err
		}
	}

	for _, place := range []struct{ source, branch string }{{fromBranch, branch}, {fromMain, "main"}} {
		data, ok, err := git.FileAt(ctx, root, place.branch, logPath(name))
		if err != nil {
			return foundLog{}, false, err
		}
		if ok {
			log.source, log.data = place.source, data
			return log, true, nil
		}
	}
	return log, false, nil
}

// readCreated reads data as a log, and returns its slice.created event
// and when it happened. ok is false when the log is empty, its first line
// cannot be read, or its first line is not a slice.created event. The
// readable events and the bad lines are returned in every case.
func readCreated(data []byte) (created eventlog.SliceCreated, at string, events []eventlog.Event, bad []eventlog.BadLine, ok bool, err error) {
	events, bad, err = eventlog.ReadFrom(bytes.NewReader(data))
	if err != nil {
		return created, "", nil, nil, false, err
	}
	firstIsBad := len(bad) > 0 && bad[0].Line == 1
	if len(events) == 0 || firstIsBad || events[0].Event != eventlog.SliceCreatedEvent {
		return created, "", events, bad, false, nil
	}
	if err := json.Unmarshal(events[0].Data, &created); err != nil {
		return created, "", events, bad, false, nil
	}
	return created, events[0].At, events, bad, true, nil
}
