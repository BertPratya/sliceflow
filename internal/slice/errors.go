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

// NoBaseError is returned when the base branch is not a local branch.
type NoBaseError struct {
	Name   string // the slice name
	Base   string // the base branch that was looked for
	Chosen bool   // true when the user chose the base; false for the default, main
}

func (e *NoBaseError) Error() string {
	return fmt.Sprintf("slice: base branch %s does not exist", e.Base)
}

// NameTakenError is returned when a slice of that name exists. Err is
// ErrFolderTaken or ErrBranchTaken, which says where it was found.
type NameTakenError struct {
	Name string // the slice name
	Base string // the base branch, whose tip was looked in for the folder
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
