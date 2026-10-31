package spec

import (
	"cmp"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// Problem is one problem in a spec (PD9). Line is 0 when the problem
// belongs to no line. Hint has no period at the end: the caller adds it.
type Problem struct {
	Line    int
	Message string
	Hint    string
}

// input is what every check reads.
type input struct {
	doc   doc
	name  string    // the slice's name (R12)
	today time.Time // gives the date in R12's hint for a missing created
}

// checks are the checks of R12-R22, in that order (PD10). Each one adds
// its problems in the order it finds them.
var checks = []func(in input) []Problem{
	checkFrontMatter, // R12
	checkHeadings,    // R13
	checkLeftovers,   // R14
	checkNumbers,     // R15
	checkSections,    // R16
	checkStories,     // R17
}

// Check reads a spec and returns its problems (PD10). name is the slice's
// name, and today gives the date in a hint. The problems with no line come
// first, then the others by line; on one line they keep the order of
// R12-R22 (R2). It returns nil when there is no problem.
func Check(content []byte, name string, today time.Time) []Problem {
	in := input{doc: read(content), name: name, today: today}
	var problems []Problem
	for _, check := range checks {
		problems = append(problems, check(in)...)
	}
	slices.SortStableFunc(problems, func(a, b Problem) int {
		return cmp.Compare(a.Line, b.Line)
	})
	return problems
}

// The sections of a spec, in their order (R13).
const (
	sectionSummary       = "Summary"
	sectionStories       = "User stories"
	sectionRequirements  = "Requirements"
	sectionOutOfScope    = "Out of scope"
	sectionDecisions     = "Decisions"
	sectionOpenQuestions = "Open questions"
)

// sections are the sections of a spec, in their order (R13).
var sections = []string{
	sectionSummary, sectionStories, sectionRequirements,
	sectionOutOfScope, sectionDecisions, sectionOpenQuestions,
}

// mayBeNone are the sections that may be the one line "None." (R13).
var mayBeNone = []string{sectionOutOfScope, sectionDecisions, sectionOpenQuestions}

// The hints that more than one message shares.
const (
	hintFrontLine    = `write it as e.g. "created: 2026-10-29"`
	hintSectionNames = "the sections are: Summary, User stories, Requirements, Out of scope, Decisions, Open questions"
	hintSectionOrder = "put the sections in this order: Summary, User stories, Requirements, Out of scope, Decisions, Open questions"
	hintNumbers      = "number them in order from the top, and update the Shows: and Rule: lines"
)

var (
	// frontLineRe is a front matter line "<key>: <value>". The value is
	// checked to be not empty after it is trimmed.
	frontLineRe = regexp.MustCompile(`^([A-Za-z0-9_-]+):[ \t]+(.*)$`)
	// dateRe is a date as YYYY-MM-DD (R12).
	dateRe = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}$`)
)

// checkFrontMatter checks the front matter (R12).
//
// Assumed: a key is letters, digits, "-", or "_", and is followed by ":",
// one or more spaces, and a value that is not empty after it is trimmed.
// A blank line between the two "---" lines is also not "<key>: <value>".
// Assumed: a key that appears twice gets only the "twice" problem on its
// second line; its value is checked only on its first line.
// Assumed: created must also be a real day, e.g. 2026-02-30 is not a date.
// Assumed: a bad line that starts with "slice:" or "created:", e.g.
// "slice:demo" or "created:", counts as that key, so it gives no extra
// "front matter has no <key>".
func checkFrontMatter(in input) []Problem {
	d := in.doc
	if !d.front.found {
		return []Problem{{0, "no front matter", "start the file with the front matter of: sliceflow spec template"}}
	}
	if !d.front.closed {
		return []Problem{{1, "front matter is not closed", `add a line "---" after the last key`}}
	}
	var problems []Problem
	seen := map[string]bool{}
	for _, l := range d.lines[1 : d.front.end-1] {
		m := frontLineRe.FindStringSubmatch(l.text)
		if m == nil || strings.TrimSpace(m[2]) == "" {
			problems = append(problems, Problem{l.num, `front matter line is not "<key>: <value>"`, hintFrontLine})
			if key, _, ok := strings.Cut(l.text, ":"); ok && (key == "slice" || key == "created") {
				seen[key] = true
			}
			continue
		}
		key, value := m[1], strings.TrimSpace(m[2])
		if key != "slice" && key != "created" {
			problems = append(problems, Problem{l.num, "unknown key " + key + " in the front matter",
				"remove it. The keys are slice and created; the title lives in the event log"})
			continue
		}
		if seen[key] {
			problems = append(problems, Problem{l.num, "front matter has " + key + " twice", "keep one"})
			continue
		}
		seen[key] = true
		switch {
		case key == "slice" && value != in.name:
			problems = append(problems, Problem{l.num,
				"front matter says slice " + value + ", but this is the spec of " + in.name,
				`change it to "slice: ` + in.name + `"`})
		case key == "created" && !isDate(value):
			problems = append(problems, Problem{l.num, "created " + value + " is not a date as YYYY-MM-DD", hintFrontLine})
		}
	}
	for _, k := range []struct{ key, example string }{
		{"slice", in.name},
		{"created", in.today.Format("2006-01-02")},
	} {
		if !seen[k.key] {
			problems = append(problems, Problem{0, "front matter has no " + k.key, `add the line "` + k.key + ": " + k.example + `"`})
		}
	}
	return problems
}

// isDate says whether s is a real date as YYYY-MM-DD.
func isDate(s string) bool {
	if !dateRe.MatchString(s) {
		return false
	}
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}

// checkHeadings checks the first heading and the sections (R13).
//
// Assumed: with no heading at all, "the first heading is not "# Spec""
// has no line. Otherwise it gets the line of the first heading.
// Assumed: a section's problems (empty, cannot be None) get its ## line.
// Only the first of two sections with the same name is checked for
// being empty or None.
// Assumed: a line counts as text when it is code, or has text outside
// comments. So a section of only blank and comment lines is empty.
// Assumed: the sections, by their first heading in file order, keep the
// longest run that is already in the right order; each section not in
// that run is out of order. On a tie, the run that keeps the sections
// earliest in the file wins. So "## Open questions" moved to the top is
// one problem, on Open questions, and two neighbours swapped give one
// problem, on the lower one.
func checkHeadings(in input) []Problem {
	d := in.doc
	var problems []Problem
	if len(d.headings) == 0 {
		problems = append(problems, Problem{0, `the first heading is not "# Spec"`, `add the line "# Spec" after the front matter`})
	} else if h := d.headings[0]; h.level != 1 || h.text != "Spec" {
		problems = append(problems, Problem{h.line, `the first heading is not "# Spec"`, `add the line "# Spec" after the front matter`})
	}

	inOrder := sectionsInOrder(d.headings)
	found := map[string]int{} // section name to the line of its first heading
	for i, h := range d.headings {
		if h.level != 2 {
			continue
		}
		order := slices.Index(sections, h.text)
		switch {
		case order < 0:
			problems = append(problems, Problem{h.line, `unknown section "## ` + h.text + `"`, hintSectionNames})
			continue
		case found[h.text] > 0:
			problems = append(problems, Problem{h.line, `section "## ` + h.text + `" appears twice`, "merge them into one"})
			continue
		}
		found[h.text] = h.line
		if !inOrder[h.text] {
			problems = append(problems, Problem{h.line, `section "## ` + h.text + `" is out of order`, hintSectionOrder})
		}
		problems = append(problems, checkSectionText(d, i)...)
	}

	for i, name := range sections {
		if found[name] > 0 {
			continue
		}
		hint := `add it after "# Spec"`
		if i > 0 {
			hint = `add it after "## ` + sections[i-1] + `". With nothing to say, write None`
		}
		problems = append(problems, Problem{0, `missing section "## ` + name + `"`, hint})
	}
	return problems
}

// sectionsInOrder returns the sections that are in the longest run of
// sections already in the right order (R13). It reads the first heading
// of each known section, in file order. There are at most 6, so it tries
// every subset. On a tie, the subset whose sections come earliest in the
// file wins: the first place where two subsets differ is kept in it.
func sectionsInOrder(headings []heading) map[string]bool {
	var names []string // known sections, by first heading, in file order
	for _, h := range headings {
		if h.level == 2 && slices.Contains(sections, h.text) && !slices.Contains(names, h.text) {
			names = append(names, h.text)
		}
	}
	best := []int{} // indexes in names of the best run
	for mask := 1; mask < 1<<len(names); mask++ {
		var run []int
		for i := range names {
			if mask&(1<<i) != 0 {
				run = append(run, i)
			}
		}
		if !inSectionOrder(names, run) {
			continue
		}
		if len(run) > len(best) || (len(run) == len(best) && slices.Compare(run, best) < 0) {
			best = run
		}
	}
	inOrder := map[string]bool{}
	for _, i := range best {
		inOrder[names[i]] = true
	}
	return inOrder
}

// inSectionOrder says whether names[run[0]], names[run[1]], ... are in
// the order of sections.
func inSectionOrder(names []string, run []int) bool {
	for k := 1; k < len(run); k++ {
		if slices.Index(sections, names[run[k-1]]) > slices.Index(sections, names[run[k]]) {
			return false
		}
	}
	return true
}

// checkSectionText checks that the section of the heading d.headings[i]
// is not empty, and is "None." only where that is allowed (R13).
func checkSectionText(d doc, i int) []Problem {
	h := d.headings[i]
	end := len(d.lines) // number of the section's last line
	for _, next := range d.headings[i+1:] {
		if next.level == 2 {
			end = next.line - 1
			break
		}
	}
	text := textLines(d.lines[h.line:end])
	canBeNone := slices.Contains(mayBeNone, h.text)
	switch {
	case len(text) == 0 && canBeNone:
		return []Problem{{h.line, `section "## ` + h.text + `" is empty`, "write its text, or None"}}
	case len(text) == 0:
		return []Problem{{h.line, `section "## ` + h.text + `" is empty`, "write its text"}}
	case !canBeNone && len(text) == 1 && !text[0].inCode && strings.TrimSpace(text[0].clean) == "None.":
		return []Problem{{h.line, `section "## ` + h.text + `" cannot be None`,
			"a spec needs a summary, at least one story, and at least one requirement"}}
	}
	return nil
}

// textLines returns the lines that are text: code, or a line with text
// outside comments.
func textLines(lines []line) []line {
	var text []line
	for _, l := range lines {
		if l.inCode || strings.TrimSpace(l.clean) != "" {
			text = append(text, l)
		}
	}
	return text
}

// maxPlaceholder is the number of characters of a placeholder shown in a
// message (R14).
const maxPlaceholder = 40

// checkLeftovers finds the "{{" and "<!--" left from the template (R14).
// They are found in masked, where code is spaces, so they count also in
// the front matter and inside a comment.
//
// Assumed: each "{{" is a problem, also one inside another placeholder,
// e.g. "{{a {{b}}" gives two. A placeholder with no "}}" after it on its
// line is shown up to the end of the line.
// Assumed: characters are counted as Unicode code points (runes).
func checkLeftovers(in input) []Problem {
	var problems []Problem
	for _, l := range in.doc.lines {
		for i := 0; i < len(l.masked); {
			switch {
			case strings.HasPrefix(l.masked[i:], "{{"):
				problems = append(problems, Problem{l.num,
					"placeholder left from the template: " + placeholder(l.text, i),
					"replace it with real text"})
				i += len("{{")
			case strings.HasPrefix(l.masked[i:], "<!--"):
				problems = append(problems, Problem{l.num, "comment left from the template",
					"delete it. A finished spec has no comments"})
				i += len("<!--")
			default:
				i++
			}
		}
	}
	return problems
}

// placeholder returns the placeholder that starts at text[i], up to and
// including its "}}", cut after maxPlaceholder characters (R14).
func placeholder(text string, i int) string {
	p := text[i:]
	if j := strings.Index(p[len("{{"):], "}}"); j >= 0 {
		p = p[:len("{{")+j+len("}}")]
	}
	if utf8.RuneCountInString(p) <= maxPlaceholder {
		return p
	}
	runes := []rune(p)
	return string(runes[:maxPlaceholder]) + "…"
}

// numbered is one thing that R15 numbers: an item or a story.
type numbered struct {
	id   string // e.g. "R3" or "Story 2"
	num  int
	line int
}

// checkNumbers checks that requirements, scenarios, decisions, questions,
// and stories are each numbered 1, 2, 3, ... from the top (R15). A
// removed item keeps its number, so it counts too.
//
// Assumed: a story is a ### heading under "## User stories" whose number
// storyNumber can read, also when R17 finds its heading badly written, so
// one typo gives one problem. A ### heading with no number is skipped.
func checkNumbers(in input) []Problem {
	d := in.doc
	var problems []Problem
	for _, kind := range []string{"R", "S", "SD", "Q"} {
		var list []numbered
		for _, it := range d.items {
			if it.kind == kind {
				list = append(list, numbered{it.id, it.num, it.line})
			}
		}
		problems = append(problems, checkOrder(list, kind)...)
	}
	var stories []numbered
	for _, h := range storyHeadings(d) {
		if n, ok := storyNumber(h.text); ok {
			stories = append(stories, numbered{fmt.Sprintf("Story %d", n), n, h.line})
		}
	}
	return append(problems, checkOrder(stories, "Story ")...)
}

// checkOrder checks that the numbers of list, in file order, are 1, 2,
// 3, ... prefix makes an id from a number, e.g. "R" or "Story ".
//
// Assumed: the first one that is not 1 gives "<id> comes first: expected
// <prefix>1". After a number out of order, the next one expected is that
// number + 1, so one gap gives one problem, not a chain. A number seen
// before gives only "appears twice", and changes nothing that is expected.
// So two items swapped give three problems: R1, R3, R2, R4 gives "R3 comes
// after R1: expected R2", "R2 comes after R3: expected R4", and "R4 comes
// after R2: expected R3".
func checkOrder(list []numbered, prefix string) []Problem {
	var problems []Problem
	first := map[int]int{} // number to the line where it first appears
	expected, before := 1, ""
	for _, n := range list {
		if line, ok := first[n.num]; ok {
			problems = append(problems, Problem{n.line, fmt.Sprintf("%s appears twice (first on line %d)", n.id, line), hintNumbers})
			continue
		}
		first[n.num] = n.line
		if n.num != expected {
			want := fmt.Sprintf("%s%d", prefix, expected)
			msg := n.id + " comes first: expected " + want
			if before != "" {
				msg = n.id + " comes after " + before + ": expected " + want
			}
			problems = append(problems, Problem{n.line, msg, hintNumbers})
		}
		expected, before = n.num+1, n.id
	}
	return problems
}

// homeSection is the section each kind of item belongs in (R16).
var homeSection = map[string]string{
	"R":  sectionRequirements,
	"S":  sectionStories,
	"SD": sectionDecisions,
	"Q":  sectionOpenQuestions,
}

// checkSections checks that each item is in its section (R16). A removed
// item is checked only for its number (R11), so it is skipped.
//
// Assumed: when an item's section is missing from the spec, its items are
// not reported: R13's "missing section" covers them, e.g. a misspelled
// "## Requirement" gives no problem for each requirement under it.
func checkSections(in input) []Problem {
	present := map[string]bool{}
	for _, h := range in.doc.headings {
		if h.level == 2 {
			present[h.text] = true
		}
	}
	var problems []Problem
	for _, it := range in.doc.items {
		if it.removed {
			continue
		}
		if home := homeSection[it.kind]; it.section != home && present[home] {
			problems = append(problems, Problem{it.line, it.id + " is in the wrong section", `move it to "## ` + home + `"`})
		}
	}
	return problems
}

var (
	// storyTitleRe is a story heading as it must be written, with a
	// title (R17).
	storyTitleRe = regexp.MustCompile(`^Story [0-9]+:\s*\S`)
	// storyNumberRe reads the number of a story heading loosely, so a
	// typo in the heading gives only R17's heading problem (R15, R17).
	storyNumberRe = regexp.MustCompile(`(?i)^story\s+([0-9]{1,9})\b`)
)

// storyNumber reads n from a story heading's text, e.g. "Story 2: Go",
// also when it is badly written, e.g. "story 2 Go". ok is false when the
// text has no story number.
func storyNumber(text string) (n int, ok bool) {
	m := storyNumberRe.FindStringSubmatch(text)
	if m == nil {
		return 0, false
	}
	n, _ = strconv.Atoi(m[1]) // at most 9 digits: always fits
	return n, true
}

// checkStories checks the stories under "## User stories" (R17). Each
// story problem gets the line of its ### heading.
//
// Assumed: the "has no sentence" and "has no scenarios" checks run for
// every ### heading whose number storyNumber can read, also one that is
// badly written. A ### heading with no number gets only the heading
// problem. Either way, the scenarios under it are under a heading, so
// they are not "not under a story": one typo gives one problem.
// Assumed: a story's sentence is a text line (as in R13), not a heading,
// between the heading and its first item, or the end of the story when it
// has no item.
// Assumed: a removed scenario still counts as a scenario of its story, so
// a story whose scenarios were all removed does not lose them.
// Assumed: "<id> is not under a story" is only for a scenario in "## User
// stories": a scenario in another section gets R16's problem instead. The
// hint keeps "<n>" and "<title>" as written, since there is no number to
// give.
func checkStories(in input) []Problem {
	d := in.doc
	var problems []Problem
	for _, h := range storyHeadings(d) {
		if !storyTitleRe.MatchString(h.text) {
			problems = append(problems, Problem{h.line, `heading under "## User stories" is not "### Story <n>: <title>"`,
				`write it as e.g. "### Story 1: Start a new slice"`})
		}
		n, ok := storyNumber(h.text)
		if !ok {
			continue
		}
		problems = append(problems, checkStory(d, h, n)...)
	}
	for _, it := range d.items {
		if it.kind == "S" && !it.removed && it.section == sectionStories && it.storyLine == 0 {
			problems = append(problems, Problem{it.line, it.id + " is not under a story",
				`add a heading "### Story <n>: <title>" above it`})
		}
	}
	return problems
}

// checkStory checks that story n, with the heading h, has a sentence and
// a scenario (R17).
func checkStory(d doc, h heading, n int) []Problem {
	end := len(d.lines) // number of the story's last line
	for _, next := range d.headings {
		if next.line > h.line && next.level <= 3 {
			end = next.line - 1
			break
		}
	}
	firstItem, scenarios := 0, 0
	for _, it := range d.items {
		if it.storyLine != h.line {
			continue
		}
		if firstItem == 0 {
			firstItem = it.line
		}
		if it.kind == "S" {
			scenarios++
		}
	}
	if firstItem > 0 {
		end = firstItem - 1
	}
	var problems []Problem
	sentence := false
	for _, l := range textLines(d.lines[h.line:end]) {
		if _, isHeading := readHeading(l); !isHeading {
			sentence = true
		}
	}
	if !sentence {
		problems = append(problems, Problem{h.line, fmt.Sprintf("Story %d has no sentence", n),
			"add one sentence under the heading: who wants to do what"})
	}
	if scenarios == 0 {
		problems = append(problems, Problem{h.line, fmt.Sprintf("Story %d has no scenarios", n),
			`add a scenario under it, e.g. "- **S1.** **When** ..."`})
	}
	return problems
}

// storyHeadings returns the ### headings under "## User stories".
func storyHeadings(d doc) []heading {
	var list []heading
	section := ""
	for _, h := range d.headings {
		switch {
		case h.level == 2:
			section = h.text
		case h.level == 3 && section == sectionStories:
			list = append(list, h)
		}
	}
	return list
}
