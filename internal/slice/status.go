package slice

import (
	"context"

	"github.com/BertPratya/sliceflow/internal/eventlog"
	"github.com/BertPratya/sliceflow/internal/git"
	"github.com/BertPratya/sliceflow/internal/project"
)

// worktreeSetting is the git setting that chose where worktrees went.
const worktreeSetting = "sliceflow.worktree.dir"

// StatusResult is what Status found about one slice.
type StatusResult struct {
	Name        string                // the slice name, in lowercase
	Created     eventlog.SliceCreated // the data of the slice.created event
	CreatedAt   string                // when it happened, in UTC, written like 2026-11-02T09:15:30Z
	BranchFound bool                  // whether the branch slice/<name> exists
	Worktree    string                // the full path of the slice's worktree, or "" when there is none
	Events      int                   // the number of lines of the log that can be read
	LastEvent   string                // the event name of the last of them
	ReadFrom    string                // "worktree", "branch", or "main": where the log was found
	LogPath     string                // where the log is in the project, written with /
	BadLines    []int                 // the numbers of the lines that could not be read, from 1
}

// Status finds the slice with the given name in the project that dir is
// in, and returns what its log says, and where its branch and worktree
// are. It changes nothing.
//
// It checks, in this order, and stops at the first problem: the name
// (*InvalidNameError), and the project (ErrNoCommits, or a *CheckError
// when git fails, which holds git.ErrNotFound when git is not found).
// When no log is found, the error is a *NoLogError; when the log has no
// slice.created event as its first line, a *NoCreatedEventError.
func Status(ctx context.Context, dir, given string) (StatusResult, error) {
	name, err := ParseName(given)
	if err != nil {
		return StatusResult{}, &InvalidNameError{Given: given, Problem: err}
	}
	root, err := findProject(ctx, dir)
	if err != nil {
		return StatusResult{}, err
	}

	log, found, err := findLog(ctx, root, name)
	if err != nil {
		return StatusResult{}, &CheckError{Check: "find the log of " + name, Err: err}
	}
	if !found {
		setting, err := git.Config(ctx, root, worktreeSetting)
		if err != nil {
			return StatusResult{}, &CheckError{Check: "read the setting " + worktreeSetting, Err: err}
		}
		worktree, err := project.WorktreePath(root, setting, name)
		if err != nil {
			return StatusResult{}, &CheckError{Check: "find the home folder", Err: err}
		}
		return StatusResult{}, &NoLogError{Name: name, Worktree: worktree}
	}

	created, at, events, bad, ok, err := readCreated(log.data)
	if err != nil {
		return StatusResult{}, &CheckError{Check: "read the log of " + name, Err: err}
	}
	badLines := lineNumbers(bad)
	if !ok {
		return StatusResult{}, &NoCreatedEventError{Name: name, LogPath: logPath(name), BadLines: badLines}
	}
	branchFound, err := git.BranchExists(ctx, root, "slice/"+name)
	if err != nil {
		return StatusResult{}, &CheckError{Check: "read the branch slice/" + name, Err: err}
	}

	return StatusResult{
		Name:        name,
		Created:     created,
		CreatedAt:   at,
		BranchFound: branchFound,
		Worktree:    log.worktree,
		Events:      len(events),
		LastEvent:   events[len(events)-1].Event,
		ReadFrom:    log.source,
		LogPath:     logPath(name),
		BadLines:    badLines,
	}, nil
}

// findProject returns the main folder of the project that dir is in,
// once it knows the project has a commit.
func findProject(ctx context.Context, dir string) (string, error) {
	root, err := git.RepoRoot(ctx, dir)
	if err != nil {
		return "", &CheckError{Check: "find the project", Err: err}
	}
	hasCommits, err := git.HasCommits(ctx, root)
	if err != nil {
		return "", &CheckError{Check: "read the commits", Err: err}
	}
	if !hasCommits {
		return "", ErrNoCommits
	}
	return root, nil
}

// lineNumbers returns the numbers of the bad lines, in order.
func lineNumbers(bad []eventlog.BadLine) []int {
	var numbers []int
	for _, line := range bad {
		numbers = append(numbers, line.Line)
	}
	return numbers
}
