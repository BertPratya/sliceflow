package spec

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// spec joins lines into the text of a spec, each with its line end.
func spec(lines ...string) string {
	return strings.Join(lines, "\n") + "\n"
}

// story1 is a story with one rule and no problem, on lines 1 to 6.
var story1 = []string{
	"## Story 1: Log in",
	"Uses: none",
	"### Rule `log-in`: a user logs in",
	"Check: e2e",
	"- **Example:** **When** I log in,",
	"  **Then** I see the shop.",
}

// with returns story1's lines followed by more.
func with(more ...string) string {
	return spec(append(append([]string{}, story1...), more...)...)
}

// found returns each problem of text as "<line>: <message>".
func found(text string) []string {
	var out []string
	for _, p := range Check(Read([]byte(text))) {
		out = append(out, fmt.Sprintf("%d: %s", p.Line, p.Message))
	}
	return out
}

// Each case is a small spec, and the problems Check finds in it, in
// order. A case with no problems wants nil.
func TestCheck(t *testing.T) {
	tests := []struct {
		about string
		text  string
		want  []string
	}{
		{"spec-needs-a-story: only a summary",
			spec("## Summary", "", "A login page."),
			[]string{"1: the spec has no story."}},
		{"spec-needs-a-story: one story with one rule, and no other section",
			with(), nil},

		{"story-heading-shape: a story heading with no number",
			with("## Story two: Print", "Uses: none", "### Rule `print`: print", "Check: unit", "- **Example:** **When** a **Then** b",
				"## Story 3: More", "Uses: none", "### Rule `more`: more", "Check: unit", "- **Example:** **When** a **Then** b"),
			[]string{`7: the heading "## Story two: Print" is not in the shape "## Story <n>: <title>".`}},
		{"story-heading-shape: a heading whose first word is not Story",
			with("## Storyboard", "Some notes."), nil},

		{"stories-numbered-in-order: Story 3 after Story 1",
			with("## Story 3: Print", "Uses: none", "### Rule `print`: print", "Check: unit", "- **Example:** **When** a **Then** b"),
			[]string{"7: Story 3 should be Story 2: stories are numbered 1, 2, 3, ... in order."}},
		{"stories-numbered-in-order: Story 2 twice",
			with("## Story 2: Print", "Uses: none", "### Rule `print`: print", "Check: unit", "- **Example:** **When** a **Then** b",
				"## Story 2: More", "Uses: none", "### Rule `more`: more", "Check: unit", "- **Example:** **When** a **Then** b"),
			[]string{"12: Story 2 should be Story 3: stories are numbered 1, 2, 3, ... in order."}},

		{"story-needs-a-rule: a story with only a Uses: line",
			spec("## Shared rules", "### Rule `unknown-event`: x", "Check: unit",
				"## Story 1: Log in", "Uses: unknown-event", "### Rule `log-in`: x", "Check: unit", "- **Example:** **When** a **Then** b",
				"## Story 2: Print", "Uses: unknown-event"),
			[]string{"9: Story 2 has no rules."}},

		{"story-needs-uses: no Uses: line",
			spec("## Story 1: Log in", "### Rule `log-in`: x", "Check: unit", "- **Example:** **When** a **Then** b"),
			[]string{"1: Story 1 has no Uses: line."}},
		{"story-needs-uses: a second Uses: line, anywhere in the story",
			with("Uses: none"),
			[]string{"7: Story 1 has a second Uses: line."}},

		{"uses-names-shared-rules: a name that is not a shared rule",
			spec("## Shared rules", "### Rule `unknown-event`: x", "Check: unit",
				"## Story 1: Log in", "Uses: unknwon-event, unknown-event", "### Rule `log-in`: x", "Check: unit", "- **Example:** **When** a **Then** b"),
			[]string{"5: Story 1 uses unknwon-event, which is not a shared rule."}},

		{"shared-rule-used: a shared rule no story uses",
			spec("## Shared rules", "### Rule `unknown-event`: x", "Check: unit",
				"## Story 1: Log in", "Uses: none", "### Rule `log-in`: x", "Check: unit", "- **Example:** **When** a **Then** b"),
			[]string{"2: shared rule unknown-event is not used by any story."}},

		{"rule-in-unknown-section: a misspelled story",
			with("## Stroy 2: Print", "### Rule `template-printed`: x", "Check: unit"),
			[]string{`8: rule template-printed is in the section "Stroy 2: Print", which is not a story or Shared rules.`}},
		{"rule-in-unknown-section: a rule in Out of scope",
			with("## Out of scope", "### Rule `edit-needs-a-name`: x", "Check: unit"),
			[]string{`8: rule edit-needs-a-name is in the section "Out of scope", which is not a story or Shared rules.`}},
		{"rule-in-unknown-section: a rule before the first section",
			"### Rule `password-min-length`: x\nCheck: unit\n" + with(),
			[]string{"1: rule password-min-length is not in any section."}},
		{"rule-in-unknown-section: a section of its own with no rule",
			with("## Notes", "Some notes."), nil},

		{"section-heading-exact: capital letters differ",
			with("## Open Questions", "- **Q1.** x", "- **Q2.** y"),
			[]string{`7: section "Open Questions" looks like "Open questions".`}},
		{"section-heading-exact: a section twice",
			with("## Open questions", "None", "## Open questions", "None"),
			[]string{`9: section "Open questions" is here a second time. The first is on line 7.`}},
		{"section-heading-exact: spaces at the end are ignored",
			with("## Open questions  ", "None"), nil},

		{"rule-heading-shape: a name with no backticks",
			with("### Rule title-required: a slice needs a title", "Check: e2e"),
			[]string{"7: the heading \"### Rule title-required: a slice needs a title\" is not in the shape \"### Rule `<name>`: <one line>\"."}},
		{"rule-heading-shape: a heading whose first word is not Rule",
			with("### Rules of thumb", "Keep it short."), nil},

		{"rule-name-shape: capital letters and a space",
			with("### Rule `Title Required`: a slice needs a title", "Check: unit"),
			[]string{`7: rule name "Title Required" is not lowercase words joined by hyphens.`}},
		{"rule-name-shape: digits are allowed",
			with("### Rule `utf8-text`: text", "Check: unit"), nil},

		{"rule-name-unique: a shared rule and a story rule with one name",
			spec("## Shared rules", "### Rule `unknown-event`: x", "Check: unit",
				"## Story 1: Log in", "Uses: unknown-event", "### Rule `unknown-event`: x", "Check: unit", "- **Example:** **When** a **Then** b"),
			[]string{"6: rule unknown-event is already used on line 2."}},

		{"rule-has-one-check: no Check: line",
			with("### Rule `edit-empty-values`: x", "Text."),
			[]string{"7: rule edit-empty-values has no Check: line."}},
		{"rule-has-one-check: a Check: line with no value",
			with("### Rule `edit-empty-values`: x", "Check:"),
			[]string{"7: rule edit-empty-values has no Check: line."}},
		{"rule-has-one-check: two Check: lines",
			with("### Rule `edit-empty-values`: x", "Check: unit", "Text.", "Check: e2e", "- **Example:** **When** a **Then** b"),
			[]string{"10: rule edit-empty-values has more than one Check: line."}},

		{"check-values: a misspelled value",
			with("### Rule `nothing-to-change`: x", "Check: unit, manaul", "- **Example:** **When** a **Then** b"),
			[]string{"8: rule nothing-to-change has the Check: value manaul, which is not unit, api, e2e, or manual."}},
		{"check-values: capital letters are another value",
			with("### Rule `nothing-to-change`: x", "Check: Unit", "- **Example:** **When** a **Then** b"),
			[]string{"8: rule nothing-to-change has the Check: value Unit, which is not unit, api, e2e, or manual."}},
		{"check-values: two values, spaces, and a final comma",
			with("### Rule `nothing-to-change`: x", "Check: unit ,  e2e,", "- **Example:** **When** a **Then** b"), nil},

		{"rule-needs-an-example: e2e and no example",
			with("### Rule `nothing-to-change`: x", "Check: e2e"),
			[]string{"7: rule nothing-to-change has no example."}},
		{"rule-needs-an-example: unit, not the first rule, and no example",
			with("### Rule `state-replayed`: x", "Check: unit"), nil},
		{"rule-needs-an-example: the first rule of a story, with unit",
			with("## Story 2: State", "Uses: none", "### Rule `state-replayed`: x", "Check: unit"),
			[]string{"9: rule state-replayed is the first rule of Story 2, so it needs an example."}},

		{"example-has-when-then: no Then",
			with("### Rule `template-printed`: x", "Check: e2e", "- **Example:** **When** I run it,", "  it prints."),
			[]string{"9: an example of rule template-printed has no **Then**."}},
		{"example-has-when-then: neither, When first",
			with("### Rule `template-printed`: x", "Check: e2e", "- **Example:** it prints."),
			[]string{"9: an example of rule template-printed has no **When**.", "9: an example of rule template-printed has no **Then**."}},
		{"example-has-when-then: no Given is needed, and the example ends at the next line at the first column",
			with("### Rule `template-printed`: x", "Check: e2e", "- **Example:** **When** I run it,", "", "**Then** outside."),
			[]string{"9: an example of rule template-printed has no **Then**."}},

		{"placeholder-left: in a story heading",
			spec("## Story 1: {{short title}}", "Uses: none", "### Rule `log-in`: x", "Check: unit", "- **Example:** **When** a **Then** b"),
			[]string{"1: a placeholder from the template is still here: {{short title}}."}},
		{"placeholder-left: in backticks is not one",
			with("Text with `{{slice}}` and `{{date}}`."), nil},
		{"placeholder-left: only the placeholder problem on its line",
			with("### Rule `log-out`: x", "Check: {{unit, api, e2e, or manual}}", "- **Example:** **When** a **Then** b"),
			[]string{"8: a placeholder from the template is still here: {{unit, api, e2e, or manual}}."}},
		{"placeholder-left: no }} runs to the end of the line, and two are two",
			with("a {{one}} b {{two"),
			[]string{"7: a placeholder from the template is still here: {{one}}.", "7: a placeholder from the template is still here: {{two."}},

		{"comment-left: one problem for a comment with a placeholder inside",
			with("## Open questions", "<!-- Questions for the human.", "     {{x}}", "     more", "     The end. -->", "None"),
			[]string{"8: an HTML comment is still here. It is hidden when the spec is shown, but an agent reads it."}},
		{"comment-left: a code block inside a comment that closes on the comment's line",
			spec("## Shared rules", "<!-- Example:", "     ## Story 1: {{x}}", "     ```", "     text", "     ``` -->",
				"## Story 1: Start a new slice", "Uses: none", "### Rule `start`: x", "Check: e2e", "- **Example:** **When** a **Then** b"),
			[]string{"2: an HTML comment is still here. It is hidden when the spec is shown, but an agent reads it."}},

		{"open-questions-left: three questions",
			with("## Open questions", "- **Q1.** a?", "- **Q2.** b?", "  more of b", "- **Q3.** c?"),
			[]string{"7: 3 questions are still open."}},
		{"open-questions-left: one question",
			with("## Open questions", "- **Q1.** a?"),
			[]string{"7: 1 question is still open."}},
		{"open-questions-left: text that is not None",
			with("## Open questions", "Ask about the title length."),
			[]string{`7: Open questions must say None, or list questions as "- **Q1.** ...".`}},
		{"open-questions-left: an empty section",
			with("## Open questions"),
			[]string{`7: Open questions must say None, or list questions as "- **Q1.** ...".`}},
		{"open-questions-left: None, in any capitals, with a period",
			with("## Open questions", "none."), nil},

		{"decisions-numbered-in-order: SD4 after SD2",
			with("## Decisions", "- **SD1.** a", "  Rule: none", "- **SD2.** b", "  Rule: none", "- **SD4.** c", "  Rule: none"),
			[]string{"12: SD4 should be SD3: decisions are numbered SD1, SD2, SD3, ... in order."}},

		{"item-shape: a decision with no stars",
			with("## Decisions", "- SD5. Edits commit only the log.", "  Rule: none"),
			[]string{`8: this looks like a decision, but is not written as "- **SD<n>.** ...".`}},
		{"item-shape: a question with the period outside the stars",
			with("## Open questions", "- **Q1**. a?"),
			[]string{`7: Open questions must say None, or list questions as "- **Q1.** ...".`,
				`8: this looks like a question, but is not written as "- **Q<n>.** ...".`}},

		{"decision-needs-rule-line: no Rule: line",
			with("## Decisions", "- **SD1.** a", "  Why: b", "   Rule: three spaces is text"),
			[]string{"8: SD1 has no Rule: line."}},

		{"decision-rule-names-exist: a name that is not a rule",
			with("## Decisions", "- **SD1.** a", "  Rule: log-in, edit-on-base"),
			[]string{"9: SD1 names rule edit-on-base, which is not in this spec."}},

		{"code-blocks-skipped: a code block in an example",
			with("  ```", "  ## Story 1: {{short title}}", "  <!-- note -->", "  ```"), nil},
		{"code-blocks-skipped: three backticks in the middle of a line start no block",
			with("a ``` b", "## Story 2: {{short title}}", "Uses: none", "### Rule `b`: x", "Check: unit", "- **Example:** **When** a **Then** b"),
			[]string{"8: a placeholder from the template is still here: {{short title}}."}},

		{"spec-lines-read: an indented Check: in an example is text",
			with("### Rule `edit`: x", "- **Example:** **When** a", "  Check: e2e", "  **Then** b"),
			[]string{"7: rule edit has no Check: line."}},
		{"spec-lines-read: \\r\\n line ends and a byte order mark",
			"\uFEFF" + strings.ReplaceAll(with(), "\n", "\r\n"), nil},
	}
	for _, tt := range tests {
		t.Run(tt.about, func(t *testing.T) {
			if got := found(tt.text); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("problems:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(tt.want, "\n"))
			}
		})
	}
}

// Each problem has the hint of its rule, word for word.
func TestCheckHints(t *testing.T) {
	text := spec(
		"## Shared rules",                 // 1
		"### Rule `unknown-event`: x",     // 2
		"Check: unit",                     // 3
		"## Story 1: Log in",              // 4
		"### Rule `edit-empty-values`: x", // 5
		"Text.",                           // 6
	)
	want := []Problem{
		{2, "shared rule unknown-event is not used by any story.", "Name it in a story's Uses: line, or delete it."},
		{4, "Story 1 has no Uses: line.", `Add "Uses: none", or name the shared rules it needs, e.g. "Uses: unknown-event".`},
		{5, "rule edit-empty-values has no Check: line.", `Add a line like "Check: e2e" under the rule's heading.`},
		{5, "rule edit-empty-values is the first rule of Story 1, so it needs an example.",
			`Add one under its Check: line, as "- **Example:** **When** ... **Then** ...".`},
	}
	if got := Check(Read([]byte(text))); !reflect.DeepEqual(got, want) {
		t.Errorf("problems:\n%+v\nwant:\n%+v", got, want)
	}
}
