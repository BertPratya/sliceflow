package spec

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// templateSections are the sections of the template, in order.
var templateSections = []struct {
	name     string
	optional bool
}{
	{"Summary", false},
	{"Why", false},
	{"Success criteria", true},
	{"User stories", false},
	{"Key concepts", true},
	{"Requirements", false},
	{"Out of scope", false},
	{"Decisions", false},
	{"Assumptions", false},
	{"Open questions", false},
}

// markerSections maps the letters of an ID to the section it belongs in.
var markerSections = map[string]string{
	"R":  "Requirements",
	"S":  "User stories",
	"SC": "Success criteria",
	"SD": "Decisions",
	"Q":  "Open questions",
}

var (
	checkKinds = map[string]bool{"unit": true, "api": true, "e2e": true, "manual": true}
	dateRe     = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}$`)
)

// isPlaceholder tells whether a value holds a placeholder. R18 skips it.
func isPlaceholder(v string) bool {
	return strings.Contains(v, "{{")
}

func sectionIndex(name string) int {
	for i, s := range templateSections {
		if s.name == name {
			return i
		}
	}
	return -1
}

func unquote(v string) string {
	if len(v) >= 2 && (v[0] == '"' || v[0] == '\'') && v[len(v)-1] == v[0] {
		return v[1 : len(v)-1]
	}
	return v
}

// frontErrors checks the front matter.
func frontErrors(d Doc, slice string) []Problem {
	if !d.HasFront {
		return []Problem{errorAt(1, "no front matter: the file must start with a line ---")}
	}
	var ps []Problem
	for _, key := range []string{"slice", "created", "ideas"} {
		var f *Field
		for i := range d.Front {
			if d.Front[i].Key == key {
				f = &d.Front[i]
				break
			}
		}
		if f == nil {
			ps = append(ps, errorAt(1, fmt.Sprintf("front matter has no \"%s\"", key)))
			continue
		}
		if isPlaceholder(f.Value) {
			continue
		}
		switch key {
		case "slice":
			if v := unquote(f.Value); v != slice {
				ps = append(ps, errorAt(f.Line, fmt.Sprintf("front matter slice is \"%s\", but the slice is %s", v, slice)))
			}
		case "created":
			v := unquote(f.Value)
			ok := dateRe.MatchString(v)
			if ok {
				_, err := time.Parse("2006-01-02", v)
				ok = err == nil
			}
			if !ok {
				ps = append(ps, errorAt(f.Line, fmt.Sprintf("front matter created is \"%s\", not a date like 2026-10-24", v)))
			}
		case "ideas":
			v := f.Value
			isList := (strings.HasPrefix(v, "[") && strings.HasSuffix(v, "]")) || (v == "" && len(f.ListItems) > 0)
			if !isList {
				ps = append(ps, errorAt(f.Line, "front matter ideas is not a list, e.g. ideas: []"))
			}
		}
	}
	return ps
}

// sectionErrors checks which sections are there, and their order.
func sectionErrors(d Doc) []Problem {
	var ps []Problem
	first := map[string]int{} // name -> line of its first heading
	maxIdx, maxName := -1, ""
	for _, s := range d.Sections {
		idx := sectionIndex(s.Name)
		if idx < 0 {
			ps = append(ps, errorAt(s.Line, fmt.Sprintf("unknown section \"%s\"", "## "+s.Name)))
			continue
		}
		if line, dup := first[s.Name]; dup {
			ps = append(ps, errorAt(s.Line, fmt.Sprintf("section \"%s\" appears twice (also line %d)", "## "+s.Name, line)))
			continue
		}
		first[s.Name] = s.Line
		if idx < maxIdx {
			ps = append(ps, errorAt(s.Line, fmt.Sprintf("section \"%s\" must come before \"%s\"", "## "+s.Name, "## "+maxName)))
			continue
		}
		maxIdx, maxName = idx, s.Name
	}
	last := d.LastLine
	if last < 1 {
		last = 1
	}
	for i, ts := range templateSections {
		if ts.optional {
			continue
		}
		if _, ok := first[ts.name]; ok {
			continue
		}
		line := last
		for _, later := range templateSections[i+1:] {
			if l, ok := first[later.name]; ok {
				line = l
				break
			}
		}
		ps = append(ps, errorAt(line, fmt.Sprintf("missing section \"%s\"", "## "+ts.name)))
	}
	return ps
}

// idLetters returns the letters at the start of an ID, e.g. "SD" for SD2.
func idLetters(id string) string {
	return strings.TrimRight(id, "0123456789")
}

func findLabels(it Item, name string) []Label {
	var out []Label
	for _, l := range it.Labels {
		if l.Name == name {
			out = append(out, l)
		}
	}
	return out
}

// splitList splits a list value like "R1, R2" at commas.
func splitList(v string) []string {
	var out []string
	for _, p := range strings.Split(v, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// itemErrors checks the markers and the labeled lines of the items.
func itemErrors(d Doc) []Problem {
	var ps []Problem
	reqs := map[string]Item{} // the requirements of this spec
	for _, s := range append([]Section{d.Preamble}, d.Sections...) {
		for _, it := range s.Items {
			if it.ID != "" && idLetters(it.ID) == "R" {
				if _, ok := reqs[it.ID]; !ok {
					reqs[it.ID] = it
				}
			}
		}
	}
	seen := map[string]int{}
	for _, s := range append([]Section{d.Preamble}, d.Sections...) {
		for _, m := range s.BadMarkers {
			text := strings.TrimSpace(strings.TrimPrefix(m.Text, "-"))
			ps = append(ps, errorAt(m.Line, fmt.Sprintf("\"%s\" is not a marker; write it like **R7.**", text)))
		}
		for _, it := range s.Items {
			if it.ID == "" {
				continue
			}
			if line, dup := seen[it.ID]; dup {
				ps = append(ps, errorAt(it.Line, fmt.Sprintf("%s is used twice (also line %d)", it.ID, line)))
			} else {
				seen[it.ID] = it.Line
			}
			letters := idLetters(it.ID)
			if want, ok := markerSections[letters]; ok && want != s.Name {
				ps = append(ps, errorAt(it.Line, fmt.Sprintf("%s is outside \"%s\"", it.ID, "## "+want)))
			}
			if it.Removed {
				continue
			}
			switch letters {
			case "R":
				ps = append(ps, requirementErrors(it)...)
			case "SC":
				if len(findLabels(it, "Measured")) == 0 {
					ps = append(ps, errorAt(it.Line, it.ID+" has no Measured: line"))
				}
			case "S":
				ps = append(ps, scenarioErrors(it, reqs)...)
			case "SD":
				ps = append(ps, decisionErrors(it, reqs)...)
			}
		}
	}
	return ps
}

func requirementErrors(it Item) []Problem {
	checks := findLabels(it, "Check")
	if len(checks) == 0 {
		return []Problem{errorAt(it.Line, it.ID+" has no Check: line")}
	}
	var ps []Problem
	for _, l := range checks {
		if isPlaceholder(l.Value) {
			continue
		}
		for _, kind := range splitList(l.Value) {
			if !checkKinds[kind] {
				ps = append(ps, errorAt(l.Line, fmt.Sprintf("%s: unknown check \"%s\"; use unit, api, e2e, or manual", it.ID, kind)))
			}
		}
	}
	return ps
}

func scenarioErrors(it Item, reqs map[string]Item) []Problem {
	shows := findLabels(it, "Shows")
	if len(shows) == 0 {
		return []Problem{errorAt(it.Line, it.ID+" has no Shows: line")}
	}
	var ps []Problem
	for _, l := range shows {
		if isPlaceholder(l.Value) {
			continue
		}
		for _, name := range splitList(l.Value) {
			r, ok := reqs[name]
			switch {
			case !ok:
				ps = append(ps, errorAt(l.Line, fmt.Sprintf("%s shows %s, which does not exist", it.ID, name)))
			case r.Removed:
				ps = append(ps, errorAt(l.Line, fmt.Sprintf("%s shows %s, which was removed", it.ID, name)))
			}
		}
	}
	return ps
}

func decisionErrors(it Item, reqs map[string]Item) []Problem {
	var ps []Problem
	for _, label := range []string{"Why", "Not chosen", "Rule"} {
		if len(findLabels(it, label)) == 0 {
			ps = append(ps, errorAt(it.Line, fmt.Sprintf("%s has no %s: line", it.ID, label)))
		}
	}
	for _, l := range findLabels(it, "Rule") {
		if isPlaceholder(l.Value) || strings.EqualFold(strings.TrimSpace(l.Value), "none") {
			continue
		}
		for _, name := range splitList(l.Value) {
			r, ok := reqs[name]
			switch {
			case !ok:
				ps = append(ps, errorAt(l.Line, fmt.Sprintf("%s: Rule: names %s, which does not exist", it.ID, name)))
			case r.Removed:
				ps = append(ps, errorAt(l.Line, fmt.Sprintf("%s: Rule: names %s, which was removed", it.ID, name)))
			}
		}
	}
	return ps
}

// outOfScopeErrors checks that each Out of scope item says where it goes.
func outOfScopeErrors(d Doc) []Problem {
	var ps []Problem
	for _, s := range d.Sections {
		if s.Name != "Out of scope" {
			continue
		}
		for _, it := range s.Items {
			if it.ID != "" {
				continue
			}
			const label = "Goes to:"
			i := strings.LastIndex(it.Tail, label)
			if i < 0 {
				ps = append(ps, errorAt(it.Line, "out-of-scope item has no Goes to:"))
				continue
			}
			v := strings.TrimSpace(it.Tail[i+len(label):])
			if isPlaceholder(v) {
				continue
			}
			w := strings.ToLower(strings.TrimRight(v, ". "))
			if w != "the backlog" && w != "not planned" && !strings.HasPrefix(w, "a later slice") {
				ps = append(ps, errorAt(it.Line, fmt.Sprintf("Goes to: \"%s\" is not a place; use the backlog, a later slice (<name>), or not planned", v)))
			}
		}
	}
	return ps
}
