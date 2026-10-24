package slice

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BertPratya/sliceflow/internal/eventlog"
	"github.com/BertPratya/sliceflow/internal/git"
	"github.com/BertPratya/sliceflow/internal/state"
)

// prepared is what a command's prepare step returns: the event's data, the
// files to write (by path in the slice folder, e.g. reviews/spec-1.md), and
// the commit message.
type prepared struct {
	Data    any
	Files   map[string][]byte
	Message string
}

// saveOp describes one event to save. Prepare runs under the lock, after the
// replay. Keep lists files committed as they are on disk (e.g. spec.md).
// Tail is the last sentence of R10's error; "" means "Nothing was changed.".
type saveOp struct {
	Event   string
	Keep    []string
	Prepare func(state.State) (prepared, error)
	Tail    string
}

// saveError is R10's error.
func saveError(reason, tail string, extra []string) *Error {
	if tail == "" {
		tail = "Nothing was changed."
	}
	msg := "could not save the event: " + reason + ". " + tail
	if len(extra) > 0 {
		msg += "\n" + strings.Join(extra, "\n")
	}
	return &Error{Message: msg}
}

// writeAtomic writes data to p through p+".tmp" and a rename, so p is never
// half written.
func writeAtomic(p string, data []byte) error {
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, p); err != nil {
		os.Remove(tmp)
		return err
	}
	return nil
}

// saveEvent is the one write path. With the lock held it replays the
// committed log, runs op.Prepare, checks the move, writes the log and the
// event's files, and commits only those. When that fails it undoes it all.
// It returns the new commit's full id and the state after the event.
func saveEvent(ctx context.Context, r Repo, name, worktree string, op saveOp) (string, state.State, error) {
	if ctx.Err() != nil {
		return "", state.State{}, saveError("interrupted", op.Tail, nil)
	}
	free, err := takeLock(ctx, r, name)
	if err != nil {
		if ctx.Err() != nil {
			return "", state.State{}, saveError("interrupted", op.Tail, nil)
		}
		return "", state.State{}, err
	}
	defer free()

	folder := path.Join("docs", "slices", name)
	logRel := path.Join(folder, ".events.jsonl")
	ref := "refs/heads/slice/" + name
	committed, hasLog, err := git.Show(ctx, worktree, ref, logRel)
	if err != nil {
		return "", state.State{}, failSave(ctx, err, "show", op.Tail, nil)
	}
	events, _ := eventlog.Parse(committed)
	st, _ := state.Replay(events)

	p, err := op.Prepare(st)
	if err != nil {
		return "", state.State{}, err
	}
	seq := 1
	if len(events) > 0 {
		seq = events[len(events)-1].Seq + 1
	}
	ev := eventlog.New(seq, r.Now(), op.Event, p.Data)
	if reason := state.Allowed(st, ev); reason != "" {
		return "", state.State{}, errors.New("internal error: the move is not allowed: " + reason)
	}
	line, err := eventlog.Encode(ev)
	if err != nil {
		return "", state.State{}, err
	}

	dir := filepath.Join(worktree, filepath.FromSlash(folder))
	logFile := filepath.Join(dir, ".events.jsonl")
	// A temporary file left by a crash.
	_ = os.Remove(logFile + ".tmp")

	newLog := append([]byte(nil), committed...)
	if len(newLog) > 0 && newLog[len(newLog)-1] != '\n' {
		newLog = append(newLog, '\n')
	}
	newLog = append(newLog, line...)

	// What the undo needs.
	var created []string // files and folders this run made, in order
	var changed []string // files that existed in the commit, to put back
	gitFiles := []string{logRel}
	undo := func() []string {
		uctx := context.WithoutCancel(ctx)
		var failed []string
		if hasLog {
			if err := writeAtomic(logFile, committed); err != nil {
				failed = append(failed, logRel)
			}
		} else if err := os.Remove(logFile); err != nil && !errors.Is(err, fs.ErrNotExist) {
			failed = append(failed, logRel)
		}
		_ = os.Remove(logFile + ".tmp")
		for i := len(created) - 1; i >= 0; i-- {
			c := created[i]
			if fi, err := os.Lstat(c); err == nil && fi.IsDir() {
				removeEmptyDir(c)
			} else if err := os.Remove(c); err != nil && !errors.Is(err, fs.ErrNotExist) {
				failed = append(failed, filepath.ToSlash(c))
			}
		}
		for _, rel := range changed {
			gp := path.Join(folder, rel)
			old, _, err := git.Show(uctx, worktree, ref, gp)
			if err == nil {
				err = os.WriteFile(filepath.Join(dir, filepath.FromSlash(rel)), old, 0o644)
			}
			if err != nil {
				failed = append(failed, gp)
			}
		}
		args := append([]string{"reset", "-q", "--"}, gitFiles...)
		if _, err := git.Run(uctx, worktree, args...); err != nil {
			// A git command that was killed by Ctrl-C leaves its
			// index.lock. It is dead (Run waited for it), so the lock is
			// stale: remove it, then try again.
			if ctx.Err() != nil {
				if lp, perr := git.Run(uctx, worktree, "rev-parse", "--path-format=absolute", "--git-path", "index.lock"); perr == nil {
					_ = os.Remove(strings.TrimSpace(lp))
				}
				_, err = git.Run(uctx, worktree, args...)
			}
			if err != nil {
				failed = append(failed, "index")
			}
		}
		return failed
	}
	fail := func(err error, gitCmd string) (string, state.State, error) {
		failed := undo()
		return "", state.State{}, failSave(ctx, err, gitCmd, op.Tail, failed)
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fail(err, "")
	}
	if err := writeAtomic(logFile, newLog); err != nil {
		return fail(err, "")
	}

	rels := make([]string, 0, len(p.Files))
	for rel := range p.Files {
		rels = append(rels, rel)
	}
	sort.Strings(rels)
	for _, rel := range rels {
		target := filepath.Join(dir, filepath.FromSlash(rel))
		gp := path.Join(folder, rel)
		gitFiles = append(gitFiles, gp)
		_, inCommit, err := git.Show(ctx, worktree, ref, gp)
		if err != nil {
			return fail(err, "show")
		}
		if inCommit {
			changed = append(changed, rel)
		} else {
			// Folders this run makes are removed too, if empty.
			var made []string
			for d := filepath.Dir(target); d != dir && !exists(d); d = filepath.Dir(d) {
				made = append(made, d)
			}
			for i := len(made) - 1; i >= 0; i-- {
				created = append(created, made[i])
			}
			created = append(created, target)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return fail(err, "")
		}
		if err := os.WriteFile(target, p.Files[rel], 0o644); err != nil {
			return fail(err, "")
		}
	}
	for _, k := range op.Keep {
		gitFiles = append(gitFiles, path.Join(folder, k))
	}

	if _, err := git.Run(ctx, worktree, append([]string{"add", "-f", "--"}, gitFiles...)...); err != nil {
		return fail(err, "add")
	}
	if _, err := git.Run(ctx, worktree, append([]string{"commit", "--only", "-m", p.Message, "--"}, gitFiles...)...); err != nil {
		return fail(err, "commit")
	}
	out, err := git.Run(ctx, worktree, "rev-parse", "HEAD")
	if err != nil {
		return "", state.State{}, err
	}
	newState, _ := state.Replay(append(events, ev))
	return strings.TrimSpace(out), newState, nil
}

// failSave builds R10's error for a failed step.
func failSave(ctx context.Context, err error, gitCmd, tail string, failed []string) error {
	reason, extra := "interrupted", []string(nil)
	if ctx.Err() == nil {
		reason, extra = failureReason(err, gitCmd)
	}
	if len(failed) > 0 {
		extra = append([]string{"Could not put back: " + strings.Join(failed, ", ") + "."}, extra...)
	}
	return saveError(reason, tail, extra)
}
