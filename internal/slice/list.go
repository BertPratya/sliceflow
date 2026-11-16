package slice

import (
	"bytes"
	"context"
	"maps"
	"slices"

	"github.com/BertPratya/sliceflow/internal/git"
)

// The reasons a candidate is not listed.
const (
	NotValidName   = "name"  // the name has a capital letter, or breaks the name rules
	NoLog          = "log"   // no log was found
	NoCreatedEvent = "event" // the log's first line is not a readable slice.created event
)

// ListResult is what List found: the slices, and the candidates it left
// out, both sorted by name, by character code.
type ListResult struct {
	Slices  []ListedSlice
	Skipped []SkippedCandidate
}

// ListedSlice is one slice of the list.
type ListedSlice struct {
	Name, Title, Description string
}

// SkippedCandidate is a name that looked like a slice, but is not listed.
type SkippedCandidate struct {
	Name     string // as it was found
	Reason   string // NotValidName, NoLog, or NoCreatedEvent
	OnBranch bool   // whether it came from a branch slice/<name>
	InMain   bool   // whether it came from a folder docs/slices/<name>/ in main
	LogPath  string // for NoCreatedEvent: where the log is in the project, written with /
}

// List finds the slices of the project that dir is in. The candidates
// are the names of the local branches slice/<name>, and of the folders
// docs/slices/<name>/ at the tip of main; a name in both is one
// candidate. A candidate is listed when its name is valid, its log is
// found as Status finds it, and the log's first line is a slice.created
// event. Only that line is read. It changes nothing.
//
// It checks the project first, and returns ErrNoCommits, or a
// *CheckError when git fails, which holds git.ErrNotFound when git is not
// found.
func List(ctx context.Context, dir string) (ListResult, error) {
	root, err := findProject(ctx, dir)
	if err != nil {
		return ListResult{}, err
	}
	candidates, err := findCandidates(ctx, root)
	if err != nil {
		return ListResult{}, err
	}

	var result ListResult
	for _, name := range slices.Sorted(maps.Keys(candidates)) {
		place := candidates[name]
		skipped := SkippedCandidate{Name: name, OnBranch: place.onBranch, InMain: place.inMain}

		if valid, err := ParseName(name); err != nil || valid != name {
			skipped.Reason = NotValidName
			result.Skipped = append(result.Skipped, skipped)
			continue
		}
		log, found, err := findLog(ctx, root, name)
		if err != nil {
			return ListResult{}, &CheckError{Check: "find the log of " + name, Err: err}
		}
		if !found {
			skipped.Reason = NoLog
			result.Skipped = append(result.Skipped, skipped)
			continue
		}
		first, _, _ := bytes.Cut(log.data, []byte("\n"))
		created, _, _, _, ok, err := readCreated(first)
		if err != nil || !ok {
			skipped.Reason, skipped.LogPath = NoCreatedEvent, logPath(name)
			result.Skipped = append(result.Skipped, skipped)
			continue
		}
		result.Slices = append(result.Slices, ListedSlice{
			Name:        name,
			Title:       created.Title,
			Description: created.Description,
		})
	}
	return result, nil
}

// place says where a candidate name came from.
type place struct {
	onBranch, inMain bool
}

// findCandidates returns the candidate names of the project whose main
// folder is root, with where each came from.
func findCandidates(ctx context.Context, root string) (map[string]place, error) {
	candidates := map[string]place{}
	branches, err := git.BranchesWithPrefix(ctx, root, "slice/")
	if err != nil {
		return nil, &CheckError{Check: "list the branches", Err: err}
	}
	for _, name := range branches {
		candidates[name] = place{onBranch: true}
	}

	hasMain, err := git.BranchExists(ctx, root, "main")
	if err != nil {
		return nil, &CheckError{Check: "read the branch main", Err: err}
	}
	if !hasMain {
		return candidates, nil
	}
	folders, err := git.FoldersAt(ctx, root, "main", "docs/slices")
	if err != nil {
		return nil, &CheckError{Check: "list the slices in main", Err: err}
	}
	for _, name := range folders {
		p := candidates[name]
		p.inMain = true
		candidates[name] = p
	}
	return candidates, nil
}
