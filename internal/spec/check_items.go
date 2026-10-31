package spec

import (
	"regexp"
	"slices"
	"strings"
)

var (
	// reqIDRe is a requirement id in a Shows: or Rule: list, e.g. "R3".
	reqIDRe = regexp.MustCompile(`^R[1-9][0-9]*$`)
	// replacesRe is a Replaces: value, "<slice>/SD<n>" (R21). The slice
	// name follows the shape of 0001's slice names: 1 to 64 lowercase
	// letters, digits, and dashes, starting with a letter or a digit.
	replacesRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,63}/SD[1-9][0-9]*$`)
)

// checkKinds are the kinds a Check: line may list (R19).
var checkKinds = []string{"unit", "api", "e2e", "manual"}

// hintRuleExist is the hint of R21 for a Rule: id that is missing or
// removed.
const hintRuleExist = "name a requirement that exists, or write none"

// decisionLabels are the labels of a decision, in their order (R20).
// Replaces is optional; the others must be there.
var decisionLabels = []string{"Why", "Not chosen", "Cost", "Rule", "Replaces"}

// requirements returns each requirement id in the spec, and whether it is
// removed. Every R item counts, wherever it sits (R16 reports the section).
//
// Assumed: when one id appears twice (R15 reports it), it is removed only
// when every item with that id is removed.
func requirements(d doc) map[string]bool {
	reqs := map[string]bool{}
	for _, it := range d.items {
		if it.kind != "R" {
			continue
		}
		if removed, seen := reqs[it.id]; seen {
			reqs[it.id] = removed && it.removed
			continue
		}
		reqs[it.id] = it.removed
	}
	return reqs
}

// labelsNamed returns the item's labels with that name, in file order.
func labelsNamed(it item, name string) []label {
	var list []label
	for _, lb := range it.labels {
		if lb.name == name {
			list = append(list, lb)
		}
	}
	return list
}

// splitList splits a label's value by commas, and trims the spaces around
// each part. Each part is kept, also an empty one.
func splitList(value string) []string {
	parts := strings.Split(value, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

// requirementList reads a list of requirement ids, e.g. "R1, R2". ok is
// false when a part is not a requirement id, also an empty part, e.g.
// "R1,,R2" or "R1,". An id named twice is returned once.
func requirementList(value string) (ids []string, ok bool) {
	for _, part := range splitList(value) {
		if !reqIDRe.MatchString(part) {
			return nil, false
		}
		if !slices.Contains(ids, part) {
			ids = append(ids, part)
		}
	}
	return ids, true
}

// hasPlaceholder says whether a label's value holds a "{{" outside code
// and comments, on its line or its continuation lines. R14 reports it.
func hasPlaceholder(d doc, it item, lb label) bool {
	for num := lb.line; num <= it.end; num++ {
		l := d.lines[num-1]
		if num > lb.line && !isContinuation(l) {
			break
		}
		if strings.Contains(l.bare, "{{") {
			return true
		}
	}
	return false
}

// checkShows checks the Shows: line of each scenario (R18). A removed
// scenario is checked only for its number (R11), so it is skipped.
//
// Assumed: with two or more Shows: lines, the problem is on the second
// one, and only the first one's value is checked.
// Assumed: a value with a placeholder is not checked: R14 reports it, and
// one fault gives one problem.
// Assumed: a value with a part that is not a requirement id, also an
// empty part (e.g. "R1,,R2", "R1,", or no value), gives only "does not
// list requirements", and its ids are not looked up.
// Assumed: an id named twice in one line is looked up once.
func checkShows(in input) []Problem {
	reqs := requirements(in.doc)
	var problems []Problem
	for _, it := range in.doc.items {
		if it.kind != "S" || it.removed {
			continue
		}
		shows := labelsNamed(it, "Shows")
		if len(shows) == 0 {
			problems = append(problems, Problem{it.line, it.id + " has no Shows: line",
				`add a line "Shows: R1" under ` + it.id + `, naming the requirements it shows`})
			continue
		}
		if len(shows) > 1 {
			problems = append(problems, Problem{shows[1].line, it.id + " has two Shows: lines", "merge them into one"})
		}
		lb := shows[0]
		if hasPlaceholder(in.doc, it, lb) {
			continue
		}
		ids, ok := requirementList(lb.value)
		if !ok {
			problems = append(problems, Problem{lb.line, it.id + "'s Shows: line does not list requirements",
				`write it as e.g. "Shows: R1, R2"`})
			continue
		}
		for _, id := range ids {
			removed, exists := reqs[id]
			switch {
			case !exists:
				problems = append(problems, Problem{lb.line, it.id + " shows " + id + ", but there is no " + id,
					`name only requirements that exist, e.g. "Shows: R1, R2"`})
			case removed:
				problems = append(problems, Problem{lb.line, it.id + " shows " + id + ", but " + id + " is removed",
					"remove " + id + " from the Shows: line"})
			}
		}
	}
	return problems
}

// checkCheckKinds checks the Check: line of each requirement (R19). A
// removed requirement is skipped (R11).
//
// Assumed: with two or more Check: lines, the problem is on the second
// one, and only the first one's value is checked.
// Assumed: a value with a placeholder is not checked (R14 reports it).
// Assumed: kinds are exact and lowercase. An empty part, e.g. "unit,,e2e",
// "unit,", or no value at all, is the unknown kind "". Each unknown kind
// is reported once per line, in the order it is written.
func checkCheckKinds(in input) []Problem {
	var problems []Problem
	for _, it := range in.doc.items {
		if it.kind != "R" || it.removed {
			continue
		}
		checkLines := labelsNamed(it, "Check")
		if len(checkLines) == 0 {
			problems = append(problems, Problem{it.line, it.id + " has no Check: line",
				`add a line "Check: e2e" under ` + it.id + `. The kinds are unit, api, e2e, and manual`})
			continue
		}
		if len(checkLines) > 1 {
			problems = append(problems, Problem{checkLines[1].line, it.id + " has two Check: lines", "merge them into one"})
		}
		lb := checkLines[0]
		if hasPlaceholder(in.doc, it, lb) {
			continue
		}
		var unknown []string
		for _, kind := range splitList(lb.value) {
			if !slices.Contains(checkKinds, kind) && !slices.Contains(unknown, kind) {
				unknown = append(unknown, kind)
			}
		}
		for _, kind := range unknown {
			problems = append(problems, Problem{lb.line, it.id + ` has an unknown check kind "` + kind + `"`,
				"use unit, api, e2e, or manual"})
		}
	}
	return problems
}

// checkDecisions checks the labeled lines of each decision (R20). A
// removed decision is skipped (R11).
//
// Assumed: a label that is there two or more times gives one problem, on
// its second line; only its first line is checked for being empty.
// Replaces: may be there once, too.
// Assumed: "out of order" is checked only when no label is there twice,
// so a doubled label gives one problem. A missing label does not stop
// it: the labels that are there must still be in order.
// Assumed: an empty value is one with no text outside comments.
// Other labels on a decision, e.g. Check:, are not checked.
func checkDecisions(in input) []Problem {
	var problems []Problem
	for _, it := range in.doc.items {
		if it.kind != "SD" || it.removed {
			continue
		}
		twice := false
		for _, name := range decisionLabels {
			list := labelsNamed(it, name)
			switch {
			case len(list) == 0 && name == "Cost":
				problems = append(problems, Problem{it.line, it.id + " has no Cost: line",
					`add a line "Cost: ..." under ` + it.id + `: the downside we accept, or None`})
			case len(list) == 0 && name != "Replaces":
				problems = append(problems, Problem{it.line, it.id + " has no " + name + ": line",
					`add a line "` + name + `: ..." under ` + it.id})
			case len(list) > 1:
				twice = true
				problems = append(problems, Problem{list[1].line, it.id + " has two " + name + ": lines", "keep one"})
			}
			if len(list) > 0 && list[0].value == "" {
				problems = append(problems, Problem{list[0].line, it.id + "'s " + name + ": line is empty", "write its value"})
			}
		}
		if !twice && !decisionInOrder(it) {
			problems = append(problems, Problem{it.line, it.id + "'s lines are out of order",
				"put them in this order: Why:, Not chosen:, Cost:, Rule:, Replaces:"})
		}
	}
	return problems
}

// decisionInOrder says whether a decision's labels of decisionLabels come
// in that order. Other labels are skipped.
func decisionInOrder(it item) bool {
	last := -1
	for _, lb := range it.labels {
		rank := slices.Index(decisionLabels, lb.name)
		if rank < 0 {
			continue
		}
		if rank < last {
			return false
		}
		last = rank
	}
	return true
}

// checkRules checks the values of each decision's Rule: and Replaces:
// lines (R21). A removed decision is skipped (R11).
//
// Assumed: only the first Rule: and the first Replaces: line is checked;
// a line with an empty value is skipped (R20 reports it), and so is a
// value with a placeholder (R14 reports it).
// Assumed: "none" is exact and lowercase. A list with a part that is not
// a requirement id, also an empty part, gives only the "not "none" or a
// list" problem, and its ids are not looked up.
// Assumed: the slice name in Replaces: follows the shape of a slice name
// (1 to 64 lowercase letters, digits, and dashes, starting with a letter
// or a digit), but not the list of names Windows reserves.
func checkRules(in input) []Problem {
	reqs := requirements(in.doc)
	var problems []Problem
	for _, it := range in.doc.items {
		if it.kind != "SD" || it.removed {
			continue
		}
		if rules := labelsNamed(it, "Rule"); len(rules) > 0 {
			problems = append(problems, checkRule(in.doc, it, rules[0], reqs)...)
		}
		if list := labelsNamed(it, "Replaces"); len(list) > 0 {
			lb := list[0]
			if lb.value != "" && !hasPlaceholder(in.doc, it, lb) && !replacesRe.MatchString(lb.value) {
				problems = append(problems, Problem{lb.line, it.id + `'s Replaces: is not "<slice>/SD<n>"`,
					`write it as e.g. "Replaces: device-pairing/SD3"`})
			}
		}
	}
	return problems
}

// checkRule checks the value of one Rule: line (R21).
func checkRule(d doc, it item, lb label, reqs map[string]bool) []Problem {
	if lb.value == "" || lb.value == "none" || hasPlaceholder(d, it, lb) {
		return nil
	}
	ids, ok := requirementList(lb.value)
	if !ok {
		return []Problem{{lb.line, it.id + `'s Rule: is not "none" or a list of requirements`,
			`write it as e.g. "Rule: R11", or "Rule: none"`}}
	}
	var problems []Problem
	for _, id := range ids {
		removed, exists := reqs[id]
		switch {
		case !exists:
			problems = append(problems, Problem{lb.line, it.id + "'s Rule: names " + id + ", but there is no " + id,
				hintRuleExist})
		case removed:
			problems = append(problems, Problem{lb.line, it.id + "'s Rule: names " + id + ", but " + id + " is removed",
				hintRuleExist})
		}
	}
	return problems
}

// checkQuestions reports each question that is not removed (R22): a spec
// with an open question is not finished.
func checkQuestions(in input) []Problem {
	var problems []Problem
	for _, it := range in.doc.items {
		if it.kind == "Q" && !it.removed {
			problems = append(problems, Problem{it.line, it.id + " is not answered",
				"write the answer into the spec, and remove " + it.id + ". When no question is left, write None"})
		}
	}
	return problems
}
