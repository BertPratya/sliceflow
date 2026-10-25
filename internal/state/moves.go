package state

import (
	"fmt"

	"github.com/BertPratya/sliceflow/internal/eventlog"
)

// move is one row of R2's table. Later slices add rows to moves.
type move struct {
	event  string
	result string   // the data's result, "" for any
	from   []string // spec states it is allowed in, nil for any
	to     string   // spec state after, "" for as it was
}

var moves = []move{
	{event: "slice.titled"},
	{event: "spec.submitted", result: "accepted", from: []string{"draft", "review_passed"}, to: "submitted"},
	{event: "spec.submitted", result: "refused", from: []string{"draft", "review_passed"}},
	{event: "review.started", from: []string{"submitted"}, to: "in_review"},
	{event: "review.finished", result: "passed", from: []string{"in_review"}, to: "review_passed"},
	{event: "review.finished", result: "changes", from: []string{"in_review"}, to: "draft"},
	{event: "review.finished", result: "needs_human", from: []string{"in_review"}, to: "draft"},
}

func resultOf(e eventlog.Event) string {
	switch d := e.Data.(type) {
	case eventlog.Submitted:
		return d.Result
	case eventlog.ReviewFinished:
		return d.Result
	}
	return ""
}

// find returns the row for an event, or nil.
func find(e eventlog.Event) *move {
	r := resultOf(e)
	for i := range moves {
		m := &moves[i]
		if m.event == e.Event && (m.result == "" || m.result == r) {
			return m
		}
	}
	return nil
}

func contains(list []string, s string) bool {
	if list == nil {
		return true
	}
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// Allowed returns "" when R2 allows the event in state s, else the reason,
// as R4 words it without the line number and the "skipped" ending.
func Allowed(s State, e eventlog.Event) string {
	if e.Event == "slice.created" {
		if s.Spec.Status == "" {
			return ""
		}
		return fmt.Sprintf("slice.created is not allowed when spec.md is %s", s.Spec.Status)
	}
	if s.Spec.Status == "" {
		return fmt.Sprintf("%s comes before slice.created", e.Event)
	}
	m := find(e)
	if m == nil || !contains(m.from, s.Spec.Status) {
		return fmt.Sprintf("%s is not allowed when spec.md is %s", e.Event, s.Spec.Status)
	}
	if d, ok := e.Data.(eventlog.ReviewFinished); ok {
		if r := s.Spec.Review; r == nil || d.Doc != "spec.md" || d.Round != r.Round {
			running := 0
			if r != nil {
				running = r.Round
			}
			return fmt.Sprintf("review.finished is for round %d, but review %d is running", d.Round, running)
		}
	}
	return ""
}

// apply makes the change of an allowed event.
func apply(s State, e eventlog.Event) State {
	if e.Event == "slice.created" {
		s.Phase = "planning"
		s.Spec.Status = "draft"
		if d, ok := e.Data.(eventlog.Created); ok && d.Title != "" {
			s.Title = d.Title
		}
		return s
	}
	m := find(e)
	if m == nil {
		return s
	}
	if m.to != "" {
		s.Spec.Status = m.to
	}
	switch d := e.Data.(type) {
	case eventlog.Titled:
		s.Title = d.Title
	case eventlog.Submitted:
		if d.Result == "refused" {
			s.Spec.Refused = &Counts{Errors: d.Errors, Warnings: d.Warnings}
		} else {
			s.Spec.Refused = nil
			s.Spec.SubmitSeq = e.Seq
		}
	case eventlog.ReviewStarted:
		s.Spec.Refused = nil
		s.Spec.Review = &Review{Round: d.Round, Reviewer: d.Reviewer, File: d.File, Commit: d.Commit}
		if d.Round > s.Spec.MaxRound {
			s.Spec.MaxRound = d.Round
		}
	case eventlog.ReviewFinished:
		if s.Spec.Review != nil {
			r := *s.Spec.Review
			r.Result = d.Result
			s.Spec.Review = &r
		}
	}
	return s
}
