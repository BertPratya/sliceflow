package spec

import (
	"reflect"
	"testing"
)

// buildText is a small spec with one of each part.
const buildText = "---\n" + // 1
	"slice: login\n" + // 2
	"---\n" + // 3
	"\n" + // 4
	"## Shared rules\n" + // 5
	"\n" + // 6
	"### Rule `unknown-event`: an unknown event\n" + // 7
	"\n" + // 8
	"Text.\n" + // 9
	"Check: unit, e2e,\n" + // 10
	"\n" + // 11
	"## Story 1: Log in\n" + // 12
	"\n" + // 13
	"Uses: unknown-event\n" + // 14
	"\n" + // 15
	"### Rule `logged-in`: a user logs in\n" + // 16
	"\n" + // 17
	"Check: e2e\n" + // 18
	"\n" + // 19
	"- **Example:** **Given** a user,\n" + // 20
	"  **When** they log in,\n" + // 21
	"  Check: e2e\n" + // 22
	"  ```\n" + // 23
	"  **Then** in code\n" + // 24
	"  ```\n" + // 25
	"\n" + // 26
	"  **Then** they see it.\n" + // 27
	"- **Example:** **When** `**Then**` in backticks\n" + // 28
	"Not in the example: **Then**\n" + // 29
	"\n" + // 30
	"## Story two: Bad\n" + // 31
	"Uses: none\n" + // 32
	"### Rule title-required: no backticks\n" + // 33
	"### Rules of thumb\n" + // 34
	"\n" + // 35
	"## Decisions\n" + // 36
	"\n" + // 37
	"- **SD1.** A choice.\n" + // 38
	"  Why: because.\n" + // 39
	"  Rule: logged-in, unknown-event\n" + // 40
	"   Rule: not-read\n" + // 41
	"- SD2. Wrong shape.\n" + // 42
	"  Rule: none\n" + // 43
	"\n" + // 44
	"## Open questions\n" + // 45
	"\n" + // 46
	"- **Q1.** First?\n" + // 47
	"  more of Q1\n" + // 48
	"- Q2. Wrong shape?\n" + // 49
	"None\n" // 50

// Read builds the sections, stories, rules, examples, decisions, and
// questions from the lines outside code blocks.
func TestReadBuildsTheParts(t *testing.T) {
	s := Read([]byte(buildText))

	var kinds []Kind
	for _, sec := range s.Sections {
		kinds = append(kinds, sec.Kind)
	}
	if want := []Kind{KindSharedRules, KindStory, KindStory, KindDecisions, KindOpenQuestions}; !reflect.DeepEqual(kinds, want) {
		t.Errorf("section kinds are %v, want %v", kinds, want)
	}

	if len(s.Stories) != 2 {
		t.Fatalf("got %d stories, want 2", len(s.Stories))
	}
	one, two := s.Stories[0], s.Stories[1]
	if !one.ShapeOK || one.Number != 1 || one.Line != 12 {
		t.Errorf("story 1 is %+v, want line 12, number 1, in shape", one)
	}
	if two.ShapeOK || two.Line != 31 {
		t.Errorf("story two is %+v, want line 31, not in shape", two)
	}
	if want := []NamesLine{{Line: 14, Names: []string{"unknown-event"}}}; !reflect.DeepEqual(one.Uses, want) {
		t.Errorf("story 1 uses %+v, want %+v", one.Uses, want)
	}
	if want := []NamesLine{{Line: 32}}; !reflect.DeepEqual(two.Uses, want) {
		t.Errorf("story two uses %+v, want %+v (none)", two.Uses, want)
	}

	var rules []string
	for _, r := range s.Rules {
		rules = append(rules, r.Name)
	}
	if want := []string{"unknown-event", "logged-in", ""}; !reflect.DeepEqual(rules, want) {
		t.Errorf("rule names are %q, want %q (### Rules of thumb is not a rule)", rules, want)
	}
	shared, loggedIn, bad := s.Rules[0], s.Rules[1], s.Rules[2]
	if shared.Section.Kind != KindSharedRules || shared.Story != nil {
		t.Errorf("unknown-event is in %v, story %v, want Shared rules and no story", shared.Section.Kind, shared.Story)
	}
	if want := []NamesLine{{Line: 10, Names: []string{"unit", "e2e"}}}; !reflect.DeepEqual(shared.Checks, want) {
		t.Errorf("unknown-event checks are %+v, want %+v", shared.Checks, want)
	}
	if loggedIn.Story != one || len(one.Rules) != 1 {
		t.Errorf("logged-in is not the one rule of story 1")
	}
	if want := []NamesLine{{Line: 18, Names: []string{"e2e"}}}; !reflect.DeepEqual(loggedIn.Checks, want) {
		t.Errorf("logged-in checks are %+v, want %+v (the indented Check: is text)", loggedIn.Checks, want)
	}
	wantExamples := []Example{{Line: 20, When: true, Then: true}, {Line: 28, When: true}}
	if !reflect.DeepEqual(loggedIn.Examples, wantExamples) {
		t.Errorf("logged-in examples are %+v, want %+v", loggedIn.Examples, wantExamples)
	}
	if bad.ShapeOK || bad.Story != two {
		t.Errorf("the rule on line 33 is %+v, want a rule of story two, not in shape", bad)
	}

	wantDecisions := []*Decision{{Line: 38, Number: 1,
		Rules: []NamesLine{{Line: 40, Names: []string{"logged-in", "unknown-event"}}}}}
	if !reflect.DeepEqual(s.Decisions, wantDecisions) {
		t.Errorf("decisions are %+v, want %+v", s.Decisions, wantDecisions)
	}
	wantBad := []BadItem{{Line: 42}, {Line: 49, Question: true}}
	if !reflect.DeepEqual(s.BadItems, wantBad) {
		t.Errorf("bad items are %+v, want %+v", s.BadItems, wantBad)
	}
	wantQuestions := &OpenQuestions{Line: 45, Questions: 1, Other: []string{"- Q2. Wrong shape?", "None"}}
	if !reflect.DeepEqual(s.OpenQuestion, wantQuestions) {
		t.Errorf("open questions are %+v, want %+v", s.OpenQuestion, wantQuestions)
	}
}

// A template section is found only by its exact heading, with spaces at
// the end ignored. A heading that matches one when capital letters are
// ignored is a lookalike, and a second copy points to the first.
func TestReadSections(t *testing.T) {
	s := Read([]byte("## Summary  \n## Open Questions\n## Notes\n## Summary\n## Storyboard\n"))

	want := []Section{
		{Line: 1, Heading: "Summary", Kind: KindSummary},
		{Line: 2, Heading: "Open Questions", Kind: KindOther, Lookalike: "Open questions"},
		{Line: 3, Heading: "Notes", Kind: KindOther},
		{Line: 4, Heading: "Summary", Kind: KindSummary, First: 1},
		{Line: 5, Heading: "Storyboard", Kind: KindOther},
	}
	var got []Section
	for _, sec := range s.Sections {
		got = append(got, *sec)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("sections are\n%+v\nwant\n%+v", got, want)
	}
	if s.OpenQuestion != nil {
		t.Errorf("OpenQuestion is %+v, want nil: the lookalike is not the section", s.OpenQuestion)
	}
}
