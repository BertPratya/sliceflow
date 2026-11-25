package spec

import (
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strings"
)

// Problem is one thing wrong in a spec: the line it is on, counted from
// 1, what is wrong, and a hint on how to fix it.
type Problem struct {
	Line    int
	Message string
	Hint    string
}

// checks holds every check but checkPlaceholders, in the order their
// problems are listed when two are on the same line. A line with a
// placeholder keeps only its placeholder problems, so checkPlaceholders
// needs no place in this order.
var checks = []func(*Spec) []Problem{
	checkHasStory,
	checkStoryHeadings,
	checkStoryNumbers,
	checkStoryHasRule,
	checkStoryUses,
	checkUsesNames,
	checkSharedRulesUsed,
	checkRuleSections,
	checkSectionHeadings,
	checkRuleHeadings,
	checkRuleNameShape,
	checkRuleNameUnique,
	checkRuleHasOneCheck,
	checkCheckValues,
	checkRuleHasExample,
	checkExampleWhenThen,
	checkComments,
	checkOpenQuestions,
	checkDecisionNumbers,
	checkItemShape,
	checkDecisionRuleLine,
	checkDecisionRuleNames,
}

// Check returns every problem in s, sorted by line, then by the order of
// the checks. A line with a placeholder left gets only the placeholder
// problems.
func Check(s *Spec) []Problem {
	all := checkPlaceholders(s)
	placeholderLines := map[int]bool{}
	for _, p := range all {
		placeholderLines[p.Line] = true
	}
	for _, check := range checks {
		for _, p := range check(s) {
			if !placeholderLines[p.Line] {
				all = append(all, p)
			}
		}
	}
	sort.SliceStable(all, func(i, j int) bool { return all[i].Line < all[j].Line })
	return all
}

// storyNumbers returns the number each story counts as: its own number
// when its heading is in shape, else the number after the story before
// it.
func storyNumbers(s *Spec) map[*Story]int {
	numbers := map[*Story]int{}
	previous := 0
	for _, story := range s.Stories {
		n := previous + 1
		if story.ShapeOK {
			n = story.Number
		}
		numbers[story] = n
		previous = n
	}
	return numbers
}

func checkHasStory(s *Spec) []Problem {
	if len(s.Stories) > 0 {
		return nil
	}
	return []Problem{{1, "the spec has no story.",
		`Add one as "## Story 1: <title>", with at least one rule under it.`}}
}

func checkStoryHeadings(s *Spec) []Problem {
	var out []Problem
	for _, story := range s.Stories {
		if !story.ShapeOK {
			out = append(out, Problem{story.Line,
				fmt.Sprintf(`the heading "%s" is not in the shape "## Story <n>: <title>".`, strings.TrimRight(story.Heading, " \t")),
				`Write it as "## Story <n>: <title>", e.g. "## Story 2: Print the template".`})
		}
	}
	return out
}

func checkStoryNumbers(s *Spec) []Problem {
	var out []Problem
	previous := 0
	for _, story := range s.Stories {
		want := previous + 1
		if story.ShapeOK && story.Number != want {
			out = append(out, Problem{story.Line,
				fmt.Sprintf("Story %d should be Story %d: stories are numbered 1, 2, 3, ... in order.", story.Number, want),
				"Renumber the stories from 1, in the order they appear."})
		}
		if story.ShapeOK {
			want = story.Number
		}
		previous = want
	}
	return out
}

func checkStoryHasRule(s *Spec) []Problem {
	var out []Problem
	for _, story := range s.Stories {
		if story.ShapeOK && len(story.Rules) == 0 {
			out = append(out, Problem{story.Line,
				fmt.Sprintf("Story %d has no rules.", story.Number),
				"Add at least one rule under it, as \"### Rule `<name>`: <one line>\"."})
		}
	}
	return out
}

func checkStoryUses(s *Spec) []Problem {
	var out []Problem
	for _, story := range s.Stories {
		if !story.ShapeOK {
			continue
		}
		if len(story.Uses) == 0 {
			out = append(out, Problem{story.Line,
				fmt.Sprintf("Story %d has no Uses: line.", story.Number),
				`Add "Uses: none", or name the shared rules it needs, e.g. "Uses: unknown-event".`})
			continue
		}
		if len(story.Uses) > 1 {
			out = append(out, Problem{story.Uses[1].Line,
				fmt.Sprintf("Story %d has a second Uses: line.", story.Number),
				"Merge the two into one line."})
		}
	}
	return out
}

// sharedNames returns the names of the shared rules whose heading is in
// shape.
func sharedNames(s *Spec) []string {
	var names []string
	for _, rule := range s.Rules {
		if rule.ShapeOK && rule.Section != nil && rule.Section.Kind == KindSharedRules {
			names = append(names, rule.Name)
		}
	}
	return names
}

func checkUsesNames(s *Spec) []Problem {
	shared := sharedNames(s)
	var out []Problem
	for _, story := range s.Stories {
		if !story.ShapeOK || len(story.Uses) == 0 {
			continue
		}
		for _, name := range story.Uses[0].Names {
			if !slices.Contains(shared, name) {
				out = append(out, Problem{story.Uses[0].Line,
					fmt.Sprintf("Story %d uses %s, which is not a shared rule.", story.Number, name),
					"Fix the name, or move the rule to Shared rules."})
			}
		}
	}
	return out
}

func checkSharedRulesUsed(s *Spec) []Problem {
	used := map[string]bool{}
	for _, story := range s.Stories {
		if len(story.Uses) > 0 {
			for _, name := range story.Uses[0].Names {
				used[name] = true
			}
		}
	}
	var out []Problem
	for _, rule := range s.Rules {
		if rule.ShapeOK && rule.Section != nil && rule.Section.Kind == KindSharedRules && !used[rule.Name] {
			out = append(out, Problem{rule.Line,
				fmt.Sprintf("shared rule %s is not used by any story.", rule.Name),
				"Name it in a story's Uses: line, or delete it."})
		}
	}
	return out
}

func checkRuleSections(s *Spec) []Problem {
	var out []Problem
	for _, rule := range s.Rules {
		if !rule.ShapeOK {
			continue
		}
		switch {
		case rule.Section == nil:
			out = append(out, Problem{rule.Line,
				fmt.Sprintf("rule %s is not in any section.", rule.Name),
				"Move it into a story or Shared rules."})
		case rule.Section.Kind != KindStory && rule.Section.Kind != KindSharedRules:
			out = append(out, Problem{rule.Line,
				fmt.Sprintf(`rule %s is in the section "%s", which is not a story or Shared rules.`, rule.Name, rule.Section.Heading),
				`If the section is a story, write its heading as "## Story <n>: <title>". If not, move the rule into a story or Shared rules.`})
		}
	}
	return out
}

func checkSectionHeadings(s *Spec) []Problem {
	var out []Problem
	for _, section := range s.Sections {
		switch {
		case section.Lookalike != "":
			out = append(out, Problem{section.Line,
				fmt.Sprintf(`section "%s" looks like "%s".`, section.Heading, section.Lookalike),
				fmt.Sprintf(`Write the heading as "## %s".`, section.Lookalike)})
		case section.First != 0:
			out = append(out, Problem{section.Line,
				fmt.Sprintf(`section "%s" is here a second time. The first is on line %d.`, section.Heading, section.First),
				"Merge the two into one."})
		}
	}
	return out
}

func checkRuleHeadings(s *Spec) []Problem {
	var out []Problem
	for _, rule := range s.Rules {
		if !rule.ShapeOK {
			out = append(out, Problem{rule.Line,
				fmt.Sprintf("the heading \"%s\" is not in the shape \"### Rule `<name>`: <one line>\".", strings.TrimRight(rule.Heading, " \t")),
				"Write it as \"### Rule `<name>`: <one line>\", e.g. \"### Rule `title-required`: a slice needs a title\"."})
		}
	}
	return out
}

// ruleName is the shape of a rule name: lowercase words joined by single
// hyphens.
var ruleName = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

func checkRuleNameShape(s *Spec) []Problem {
	var out []Problem
	for _, rule := range s.Rules {
		if rule.ShapeOK && !ruleName.MatchString(rule.Name) {
			out = append(out, Problem{rule.Line,
				fmt.Sprintf(`rule name "%s" is not lowercase words joined by hyphens.`, rule.Name),
				"Use only a to z, 0 to 9, and single hyphens, e.g. title-required."})
		}
	}
	return out
}

func checkRuleNameUnique(s *Spec) []Problem {
	first := map[string]int{}
	var out []Problem
	for _, rule := range s.Rules {
		if !rule.ShapeOK {
			continue
		}
		if line, ok := first[rule.Name]; ok {
			out = append(out, Problem{rule.Line,
				fmt.Sprintf("rule %s is already used on line %d.", rule.Name, line),
				"Give each rule its own name."})
			continue
		}
		first[rule.Name] = rule.Line
	}
	return out
}

// checkLines returns the Check: lines of rule that have a value: a line
// with no value counts as no Check: line.
func checkLines(rule *Rule) []NamesLine {
	var lines []NamesLine
	for _, c := range rule.Checks {
		if len(c.Names) > 0 {
			lines = append(lines, c)
		}
	}
	return lines
}

func checkRuleHasOneCheck(s *Spec) []Problem {
	var out []Problem
	for _, rule := range s.Rules {
		if !rule.ShapeOK {
			continue
		}
		lines := checkLines(rule)
		switch {
		case len(lines) == 0:
			out = append(out, Problem{rule.Line,
				fmt.Sprintf("rule %s has no Check: line.", rule.Name),
				`Add a line like "Check: e2e" under the rule's heading.`})
		case len(lines) > 1:
			out = append(out, Problem{lines[1].Line,
				fmt.Sprintf("rule %s has more than one Check: line.", rule.Name),
				`Keep one Check: line, with every value on it, e.g. "Check: unit, e2e".`})
		}
	}
	return out
}

// checkValues holds the values a Check: line may hold.
var checkValues = []string{"unit", "api", "e2e", "manual"}

func checkCheckValues(s *Spec) []Problem {
	var out []Problem
	for _, rule := range s.Rules {
		if !rule.ShapeOK {
			continue
		}
		for _, line := range checkLines(rule) {
			for _, value := range line.Names {
				if !slices.Contains(checkValues, value) {
					out = append(out, Problem{line.Line,
						fmt.Sprintf("rule %s has the Check: value %s, which is not unit, api, e2e, or manual.", rule.Name, value),
						"Fix the value."})
				}
			}
		}
	}
	return out
}

func checkRuleHasExample(s *Spec) []Problem {
	numbers := storyNumbers(s)
	const hint = `Add one under its Check: line, as "- **Example:** **When** ... **Then** ...".`
	var out []Problem
	for _, rule := range s.Rules {
		if !rule.ShapeOK || len(rule.Examples) > 0 {
			continue
		}
		if story := rule.Story; story != nil && story.Rules[0] == rule {
			out = append(out, Problem{rule.Line,
				fmt.Sprintf("rule %s is the first rule of Story %d, so it needs an example.", rule.Name, numbers[story]),
				hint})
			continue
		}
		lines := checkLines(rule)
		if len(lines) > 0 && slices.ContainsFunc(lines[0].Names, func(v string) bool { return v != "unit" }) {
			out = append(out, Problem{rule.Line, fmt.Sprintf("rule %s has no example.", rule.Name), hint})
		}
	}
	return out
}

func checkExampleWhenThen(s *Spec) []Problem {
	var out []Problem
	for _, rule := range s.Rules {
		if !rule.ShapeOK {
			continue
		}
		for _, example := range rule.Examples {
			if !example.When {
				out = append(out, Problem{example.Line,
					fmt.Sprintf("an example of rule %s has no **When**.", rule.Name),
					"Say what is done, starting with **When**."})
			}
			if !example.Then {
				out = append(out, Problem{example.Line,
					fmt.Sprintf("an example of rule %s has no **Then**.", rule.Name),
					"Say what happens, starting with **Then**."})
			}
		}
	}
	return out
}

func checkPlaceholders(s *Spec) []Problem {
	var out []Problem
	for _, line := range s.Lines {
		if line.Code {
			continue
		}
		plain := line.Plain()
		for i := 0; i < len(plain); {
			start := strings.Index(plain[i:], "{{")
			if start < 0 {
				break
			}
			start += i
			end := len(line.Text)
			if close := strings.Index(line.Text[start+2:], "}}"); close >= 0 {
				end = start + 2 + close + 2
			}
			out = append(out, Problem{line.Number,
				fmt.Sprintf("a placeholder from the template is still here: %s.", line.Text[start:end]),
				"Write the real text in its place. If you mean the characters themselves, put them in backticks."})
			i = end
		}
	}
	return out
}

func checkComments(s *Spec) []Problem {
	var out []Problem
	for _, line := range s.Comments {
		out = append(out, Problem{line,
			"an HTML comment is still here. It is hidden when the spec is shown, but an agent reads it.",
			"Delete it. Write what the reader needs as plain text."})
	}
	return out
}

func checkOpenQuestions(s *Spec) []Problem {
	q := s.OpenQuestion
	if q == nil {
		return nil
	}
	switch {
	case q.Questions == 1:
		return []Problem{{q.Line, "1 question is still open.", openQuestionsHint}}
	case q.Questions > 1:
		return []Problem{{q.Line, fmt.Sprintf("%d questions are still open.", q.Questions), openQuestionsHint}}
	}
	if len(q.Other) == 1 && strings.EqualFold(strings.TrimSuffix(q.Other[0], "."), "none") {
		return nil
	}
	return []Problem{{q.Line, `Open questions must say None, or list questions as "- **Q1.** ...".`,
		"Write each open question as a Q item, or write None."}}
}

const openQuestionsHint = "Answer each one in the spec, then remove it. When none are left, write None."

func checkDecisionNumbers(s *Spec) []Problem {
	var out []Problem
	previous := 0
	for _, d := range s.Decisions {
		if want := previous + 1; d.Number != want {
			out = append(out, Problem{d.Line,
				fmt.Sprintf("SD%d should be SD%d: decisions are numbered SD1, SD2, SD3, ... in order.", d.Number, want),
				"Renumber the decisions from SD1, in the order they appear."})
		}
		previous = d.Number
	}
	return out
}

func checkItemShape(s *Spec) []Problem {
	var out []Problem
	for _, item := range s.BadItems {
		word, label := "decision", "SD"
		if item.Question {
			word, label = "question", "Q"
		}
		out = append(out, Problem{item.Line,
			fmt.Sprintf(`this looks like a %s, but is not written as "- **%s<n>.** ...".`, word, label),
			fmt.Sprintf(`Write it as "- **%s<n>.** ...", with the number and the period inside the stars.`, label)})
	}
	return out
}

func checkDecisionRuleLine(s *Spec) []Problem {
	var out []Problem
	for _, d := range s.Decisions {
		if len(d.Rules) == 0 {
			out = append(out, Problem{d.Line,
				fmt.Sprintf("SD%d has no Rule: line.", d.Number),
				`Add "Rule: <rule names>", or "Rule: none".`})
		}
	}
	return out
}

func checkDecisionRuleNames(s *Spec) []Problem {
	var names []string
	for _, rule := range s.Rules {
		if rule.ShapeOK {
			names = append(names, rule.Name)
		}
	}
	var out []Problem
	for _, d := range s.Decisions {
		for _, line := range d.Rules {
			for _, name := range line.Names {
				if !slices.Contains(names, name) {
					out = append(out, Problem{line.Line,
						fmt.Sprintf("SD%d names rule %s, which is not in this spec.", d.Number, name),
						"Fix the name, or write none."})
				}
			}
		}
	}
	return out
}
