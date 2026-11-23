package slice

import (
	"errors"
	"fmt"
)

// ErrNoCommits is returned when the project's repository has no commit
// yet, so there is nothing to start a slice from.
var ErrNoCommits = errors.New("slice: the repository has no commits")

// ControlCharacterError is returned when the title or the description
// holds a control character, such as a line break.
type ControlCharacterError struct {
	Option string // "title" or "description"
	Text   string // the text as it was given
}

func (e *ControlCharacterError) Error() string {
	return fmt.Sprintf("slice: the %s has a control character", e.Option)
}

// InvalidNameError is returned when the given name cannot be a slice
// name. The text of Problem says which rule it breaks, as one sentence
// for the user.
type InvalidNameError struct {
	Given   string // the name as it was given
	Problem error
}

func (e *InvalidNameError) Error() string {
	return fmt.Sprintf("slice: invalid name %q: %v", e.Given, e.Problem)
}

// BaseSource says where the base branch of a new slice came from.
type BaseSource int

const (
	// BaseDefault is main, used when the user chose no base branch.
	BaseDefault BaseSource = iota
	// BaseFromFlag is the branch given with --from.
	BaseFromFlag
	// BaseFromSetting is the branch in the git setting sliceflow.base.
	BaseFromSetting
)

// NoBaseError is returned when the base branch is not a local branch.
type NoBaseError struct {
	Name   string     // the slice name
	Base   string     // the base branch that was looked for
	Source BaseSource // where Base came from
}

func (e *NoBaseError) Error() string {
	return fmt.Sprintf("slice: base branch %s does not exist", e.Base)
}

// NameTakenError is returned when a slice of that name exists. Err is
// ErrFolderTaken or ErrBranchTaken, which says where it was found.
type NameTakenError struct {
	Name string // the slice name
	Err  error
}

func (e *NameTakenError) Error() string {
	return fmt.Sprintf("slice %s already exists: %v", e.Name, e.Err)
}

// Unwrap returns ErrFolderTaken or ErrBranchTaken.
func (e *NameTakenError) Unwrap() error {
	return e.Err
}

// CheckError is returned when a check that comes before anything is made
// could not be done, e.g. because git failed. Nothing was made.
type CheckError struct {
	// Check says what could not be done, in words that fit
	// "could not ...", e.g. "find the project".
	Check string
	Err   error // why: often a *git.Error
}

func (e *CheckError) Error() string {
	return "could not " + e.Check + ": " + e.Err.Error()
}

// Unwrap returns why the check failed.
func (e *CheckError) Unwrap() error {
	return e.Err
}

// NoLogError is returned when the user's checkout has no log of the
// slice.
type NoLogError struct {
	Name   string // the slice name
	Branch string // the branch the checkout is on, or "" when it is on no branch
}

func (e *NoLogError) Error() string {
	return "slice: no slice named " + e.Name
}

// NoCreatedEventError is returned when the log of a slice is found, but
// it is empty, its first line cannot be read, or its first line is not a
// slice.created event.
type NoCreatedEventError struct {
	Name     string // the slice name
	LogPath  string // where the log is in the project, written with /
	BadLines []int  // the numbers of the lines that could not be read, from 1
}

func (e *NoCreatedEventError) Error() string {
	return "slice: no slice.created event could be read for " + e.Name
}

// WrongBranchError is returned when the user's checkout is not on the
// base branch, so a slice cannot be made there.
type WrongBranchError struct {
	Base   string // the base branch
	Branch string // the branch the checkout is on, or "" when it is on no branch
}

func (e *WrongBranchError) Error() string {
	return fmt.Sprintf("slice: the checkout is on %q, not on the base branch %q", e.Branch, e.Base)
}

// CommittedError is returned when the slice's commit was made, but git
// did not end well afterwards, e.g. because the user stopped sliceflow
// while a post-commit hook ran. The slice exists, so nothing was
// removed. Err says why git did not end well, e.g. ctx's error.
type CommittedError struct {
	Name string // the slice name
	Base string // the branch that has the commit
	Err  error
}

func (e *CommittedError) Error() string {
	return "slice: " + e.Name + " was committed, but then: " + e.Err.Error()
}

// Unwrap returns why git did not end well.
func (e *CommittedError) Unwrap() error {
	return e.Err
}
