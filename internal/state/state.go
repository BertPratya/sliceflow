// Package state replays slice events into a State through the table of
// moves. It touches no files and no git.
package state

// State is the state of a slice, worked out by replaying its log.
type State struct {
	// Phase is "planning" once a good slice.created is replayed, else "unknown".
	Phase string
	// Title is the newest title, "" when none was saved.
	Title string
	Spec  Doc
}

// Doc is where spec.md stands.
type Doc struct {
	// Status is "" before slice.created, then draft, submitted, in_review,
	// or review_passed.
	Status string
	// Review is the last started review.
	Review *Review
	// Refused is set while the last spec event is a refused spec.submitted.
	Refused *Counts
	// MaxRound is the highest round started.
	MaxRound int
	// SubmitSeq is the seq of the last accepted spec.submitted.
	SubmitSeq int
}

// Review is one review of spec.md. Result is "" while it runs.
type Review struct {
	Round                          int
	Reviewer, File, Commit, Result string
}

// Counts are the errors and warnings of a refused submit.
type Counts struct{ Errors, Warnings int }
