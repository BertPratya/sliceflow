package spec

import (
	"regexp"
	"strconv"
	"strings"
)

// Kind says what a ## section is.
type Kind int

const (
	KindOther         Kind = iota // a section that is not in the template, or a lookalike
	KindStory                     // a heading whose first word is Story
	KindSummary                   // ## Summary
	KindSharedRules               // ## Shared rules
	KindOutOfScope                // ## Out of scope
	KindDecisions                 // ## Decisions
	KindOpenQuestions             // ## Open questions
)

// templateSections maps the exact heading of each template section that
// is not a story to its kind, in the template's order.
var templateSections = []struct {
	Heading string
	Kind    Kind
}{
	{"Summary", KindSummary},
	{"Shared rules", KindSharedRules},
	{"Out of scope", KindOutOfScope},
	{"Decisions", KindDecisions},
	{"Open questions", KindOpenQuestions},
}

// Spec is a spec as sliceflow reads it: its lines, and the parts built
// from the lines outside code blocks and comments.
type Spec struct {
	Lines []Line
	// Comments holds the line where each HTML comment outside backticks
	// starts, once per comment.
	Comments     []int
	Sections     []*Section // every ## heading, in order
	Stories      []*Story
	Rules        []*Rule // every ### Rule heading, in order, wherever it is
	Decisions    []*Decision
	BadItems     []BadItem
	OpenQuestion *OpenQuestions // nil when the spec has no Open questions section
}

// Section is one ## heading and what it is.
type Section struct {
	Line    int
	Heading string // the text after "## ", without spaces at the end
	Kind    Kind
	// Lookalike is the template heading this one matches when capital
	// letters are ignored, but not exactly; "" otherwise.
	Lookalike string
	// First is the line of the first section of the same template kind,
	// when this one is a second copy; 0 otherwise.
	First int
}

// Story is a ## heading whose first word is Story.
type Story struct {
	Line    int
	Heading string
	ShapeOK bool // the heading is ## Story <n>: <title>
	Number  int  // <n>, when ShapeOK
	Uses    []NamesLine
	Rules   []*Rule
}

// NamesLine is a Uses: or Rule: line: where it is, and its names. A line
// that says none has no names.
type NamesLine struct {
	Line  int
	Names []string
}

// Rule is a ### heading whose first word is Rule.
type Rule struct {
	Line    int
	Heading string
	ShapeOK bool   // the heading is ### Rule `<name>`: <one line>
	Name    string // <name>, when ShapeOK
	// Section is the ## section the rule is in, or nil when it comes
	// before the first one.
	Section  *Section
	Story    *Story // the story it is in, or nil
	Checks   []NamesLine
	Examples []Example
}

// Example is a list item that starts with - **Example:** under a rule.
type Example struct {
	Line       int
	When, Then bool
}

// Decision is an item - **SD<n>.** in the Decisions section.
type Decision struct {
	Line   int
	Number int
	Rules  []NamesLine
}

// BadItem is an item in Decisions or Open questions that looks like a
// decision or a question, but is not in the shape of one.
type BadItem struct {
	Line     int
	Question bool // a question; otherwise a decision
}

// OpenQuestions is the Open questions section.
type OpenQuestions struct {
	Line      int      // the line of its heading
	Questions int      // the items - **Q<n>.**
	Other     []string // its other lines that are not blank, trimmed
}

var (
	storyHeading    = regexp.MustCompile(`^Story ([0-9]+): (.*\S.*)$`)
	ruleHeading     = regexp.MustCompile("^Rule `([^`]*)`: (.*\\S.*)$")
	decisionItem    = regexp.MustCompile(`^- \*\*SD([0-9]+)\.\*\*`)
	questionItem    = regexp.MustCompile(`^- \*\*Q([0-9]+)\.\*\*`)
	decisionLookish = regexp.MustCompile(`^- (\*\*)?SD`)
	questionLookish = regexp.MustCompile(`^- (\*\*)?Q`)
)

// Read reads the spec in data. It never fails: whatever it cannot
// place, it leaves out, and the checks report what is missing.
func Read(data []byte) *Spec {
	lines, comments := readLines(data)
	s := &Spec{Lines: lines, Comments: comments}
	b := builder{spec: s}
	for _, line := range lines {
		if line.Code {
			continue
		}
		b.add(line)
	}
	return s
}

// builder holds where Read is while it goes through the lines.
type builder struct {
	spec     *Spec
	section  *Section
	story    *Story
	rule     *Rule
	example  *Example
	decision *Decision
	inItem   bool // inside a decision or question item, or a bad one
}

// add places one line that is not in a code block. A blank line changes
// nothing: an example or an item ends only at the next line that is not
// blank and starts at the first column.
func (b *builder) add(line Line) {
	text := line.Text
	if strings.TrimSpace(text) == "" {
		return
	}
	atStart := text[0] != ' ' && text[0] != '\t'
	if atStart {
		b.example = nil
		b.decision = nil
		b.inItem = false
	}

	switch {
	case strings.HasPrefix(text, "## "):
		b.startSection(line)
		return
	case strings.HasPrefix(text, "### "):
		b.startH3(line)
		return
	}

	switch {
	case strings.HasPrefix(text, "Check:"):
		if b.rule != nil {
			b.rule.Checks = append(b.rule.Checks, NamesLine{Line: line.Number, Names: values(text[len("Check:"):])})
		}
		return
	case strings.HasPrefix(text, "Uses:"):
		if b.story != nil {
			b.story.Uses = append(b.story.Uses, NamesLine{Line: line.Number, Names: names(text[len("Uses:"):])})
		}
		return
	case strings.HasPrefix(text, "- **Example:**"):
		if b.rule != nil {
			b.rule.Examples = append(b.rule.Examples, Example{Line: line.Number})
			b.example = &b.rule.Examples[len(b.rule.Examples)-1]
		}
	}

	if b.example != nil {
		plain := line.Plain()
		b.example.When = b.example.When || strings.Contains(plain, "**When**")
		b.example.Then = b.example.Then || strings.Contains(plain, "**Then**")
	}

	if b.section == nil || b.section.First != 0 {
		// The decisions and questions of a second copy of a template
		// section are read by no check: its heading gets the one
		// problem. Rules under a second Shared rules are still read, as
		// shared rules, so they do not also get a problem each.
		return
	}
	switch b.section.Kind {
	case KindDecisions:
		b.addDecisionLine(line, atStart)
	case KindOpenQuestions:
		b.addQuestionLine(line, atStart)
	}
}

// startSection starts the ## section of line. A story's heading also
// starts the story; any other heading ends the story before it.
func (b *builder) startSection(line Line) {
	heading := strings.TrimRight(strings.TrimPrefix(line.Text, "## "), " \t")
	section := &Section{Line: line.Number, Heading: heading, Kind: KindOther}
	b.spec.Sections = append(b.spec.Sections, section)
	b.section, b.story, b.rule = section, nil, nil

	if firstWord(heading) == "Story" {
		section.Kind = KindStory
		story := &Story{Line: line.Number, Heading: line.Text}
		if m := storyHeading.FindStringSubmatch(heading); m != nil {
			if n, err := strconv.Atoi(m[1]); err == nil {
				story.ShapeOK, story.Number = true, n
			}
		}
		b.spec.Stories = append(b.spec.Stories, story)
		b.story = story
		return
	}

	for _, t := range templateSections {
		switch {
		case heading == t.Heading:
			section.Kind = t.Kind
			for _, before := range b.spec.Sections[:len(b.spec.Sections)-1] {
				if before.Kind == t.Kind {
					section.First = before.Line
					break
				}
			}
			if t.Kind == KindOpenQuestions && b.spec.OpenQuestion == nil {
				b.spec.OpenQuestion = &OpenQuestions{Line: line.Number}
			}
			return
		case strings.EqualFold(heading, t.Heading):
			section.Lookalike = t.Heading
			return
		}
	}
}

// startH3 starts a rule when the first word of the ### heading is Rule.
// Any other ### heading ends the rule before it.
func (b *builder) startH3(line Line) {
	b.rule = nil
	heading := strings.TrimPrefix(line.Text, "### ")
	if firstWord(heading) != "Rule" {
		return
	}
	rule := &Rule{Line: line.Number, Heading: line.Text, Section: b.section, Story: b.story}
	if m := ruleHeading.FindStringSubmatch(strings.TrimRight(heading, " \t")); m != nil {
		rule.ShapeOK, rule.Name = true, m[1]
	}
	b.spec.Rules = append(b.spec.Rules, rule)
	if b.story != nil {
		b.story.Rules = append(b.story.Rules, rule)
	}
	b.rule = rule
}

// addDecisionLine reads a line of the Decisions section: the start of a
// decision, of an item in the wrong shape, or a Rule: line of the
// decision it is in.
func (b *builder) addDecisionLine(line Line, atStart bool) {
	text := line.Text
	if atStart {
		if m := decisionItem.FindStringSubmatch(text); m != nil {
			n, _ := strconv.Atoi(m[1])
			d := &Decision{Line: line.Number, Number: n}
			b.spec.Decisions = append(b.spec.Decisions, d)
			b.decision, b.inItem = d, true
		} else if decisionLookish.MatchString(text) {
			b.spec.BadItems = append(b.spec.BadItems, BadItem{Line: line.Number})
			b.inItem = true
		}
		return
	}
	if b.decision != nil && strings.HasPrefix(text, "  Rule:") {
		b.decision.Rules = append(b.decision.Rules, NamesLine{Line: line.Number, Names: names(text[len("  Rule:"):])})
	}
}

// addQuestionLine reads a line of the Open questions section: a
// question, an item in the wrong shape, or other text. The indented
// lines of an item belong to it.
func (b *builder) addQuestionLine(line Line, atStart bool) {
	q := b.spec.OpenQuestion
	if q == nil || b.section.Kind != KindOpenQuestions {
		return
	}
	text := line.Text
	switch {
	case atStart && questionItem.MatchString(text):
		q.Questions++
		b.inItem = true
	case atStart && questionLookish.MatchString(text):
		b.spec.BadItems = append(b.spec.BadItems, BadItem{Line: line.Number, Question: true})
		q.Other = append(q.Other, strings.TrimSpace(text))
		b.inItem = true
	case !b.inItem:
		q.Other = append(q.Other, strings.TrimSpace(text))
	}
}

// firstWord returns the first word of s, split at spaces.
func firstWord(s string) string {
	word, _, _ := strings.Cut(strings.TrimLeft(s, " "), " ")
	return word
}

// values returns the values of a Check: line: split at commas, each
// trimmed, with empty values left out.
func values(s string) []string {
	var out []string
	for _, v := range strings.Split(s, ",") {
		if v = strings.TrimSpace(v); v != "" {
			out = append(out, v)
		}
	}
	return out
}

// names returns the names of a Uses: or Rule: line, as values does,
// and none for a line that says none.
func names(s string) []string {
	if strings.TrimSpace(s) == "none" {
		return nil
	}
	return values(s)
}
