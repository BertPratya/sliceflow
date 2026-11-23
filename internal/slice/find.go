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

// logPath returns where the log of the slice name is in the project,
// written with /.
func logPath(name string) string {
	return "docs/slices/" + name + "/.events.jsonl"
}

// readLog returns the whole log of the slice name in the checkout whose
// top folder is top. found is false when there is no file there. It
// never reads another branch, another worktree, or a commit.
func readLog(top, name string) (data []byte, found bool, err error) {
	data, err = os.ReadFile(filepath.Join(top, filepath.FromSlash(logPath(name))))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return data, true, nil
}

// findCheckout returns the main folder of the project that dir is in,
// once it knows the project has a commit, and the top folder of the
// checkout that dir is in.
func findCheckout(ctx context.Context, dir string) (root, top string, err error) {
	root, err = findProject(ctx, dir)
	if err != nil {
		return "", "", err
	}
	top, err = git.CheckoutTop(ctx, dir)
	if err != nil {
		return "", "", &CheckError{Check: "find the top of the checkout", Err: err}
	}
	return root, top, nil
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
